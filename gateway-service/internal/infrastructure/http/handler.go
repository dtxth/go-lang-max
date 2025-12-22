package http

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gateway-service/internal/config"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	"gateway-service/internal/infrastructure/errors"
	"gateway-service/internal/infrastructure/middleware"
	"gateway-service/internal/infrastructure/tracing"
)

// Handler manages HTTP requests and routes them to gRPC services
type Handler struct {
	config        *config.Config
	clientManager *grpcClient.ClientManager
	errorHandler  *errors.ErrorHandler
	logger        *middleware.Logger
	tracer        *tracing.Tracer
}

// NewHandler creates a new HTTP handler
func NewHandler(cfg *config.Config, clientManager *grpcClient.ClientManager) *Handler {
	// Initialize logger based on config
	logLevel := middleware.LogLevelInfo
	switch cfg.Logging.Level {
	case "debug":
		logLevel = middleware.LogLevelDebug
	case "warn":
		logLevel = middleware.LogLevelWarn
	case "error":
		logLevel = middleware.LogLevelError
	}

	logger := middleware.NewLogger(logLevel)
	tracer := tracing.NewTracer("gateway-service")

	// Set the logger on the client manager for structured error logging
	clientManager.SetLogger(logger)

	return &Handler{
		config:        cfg,
		clientManager: clientManager,
		errorHandler:  clientManager.GetErrorHandler(),
		logger:        logger,
		tracer:        tracer,
	}
}

// HandleGRPCError converts gRPC errors to HTTP responses with enhanced error handling (public for testing)
func (h *Handler) HandleGRPCError(w http.ResponseWriter, err error, requestID, serviceName, methodName string) {
	h.errorHandler.HandleGRPCError(w, err, requestID, serviceName, methodName)
}

// handleGRPCError converts gRPC errors to HTTP responses with enhanced error handling (internal)
func (h *Handler) handleGRPCError(w http.ResponseWriter, err error, requestID, serviceName, methodName string) {
	h.errorHandler.HandleGRPCError(w, err, requestID, serviceName, methodName)
}

// handleGRPCErrorLegacy provides backward compatibility for existing calls
func (h *Handler) handleGRPCErrorLegacy(w http.ResponseWriter, err error, requestID string) {
	h.errorHandler.HandleGRPCError(w, err, requestID, "", "")
}

// handleConnectionError handles gRPC connection errors gracefully
func (h *Handler) handleConnectionError(w http.ResponseWriter, err error, requestID, serviceName string) {
	h.errorHandler.HandleConnectionError(w, err, requestID, serviceName)
}

// executeWithRetryAndCircuitBreaker executes a gRPC call with retry and circuit breaker protection
func (h *Handler) executeWithRetryAndCircuitBreaker(ctx context.Context, serviceName string, fn func(ctx context.Context) error) error {
	// Start tracing span for the gRPC call
	span, spanCtx := h.tracer.StartGRPCSpan(ctx, serviceName, "execute")
	defer h.tracer.FinishSpan(span, nil)

	// Propagate context to gRPC
	grpcCtx := middleware.PropagateContextToGRPC(spanCtx)

	// Log the gRPC call attempt
	h.logger.Debug(ctx, "Starting gRPC call", map[string]interface{}{
		"service": serviceName,
	})

	start := time.Now()
	var err error

	// Get the appropriate circuit breaker and retrier based on service name
	switch serviceName {
	case "auth":
		circuitBreaker := h.clientManager.GetAuthCircuitBreaker()
		retrier := h.clientManager.GetAuthRetrier()
		err = retrier.ExecuteWithCircuitBreaker(grpcCtx, circuitBreaker, fn)
	case "chat":
		circuitBreaker := h.clientManager.GetChatCircuitBreaker()
		retrier := h.clientManager.GetChatRetrier()
		err = retrier.ExecuteWithCircuitBreaker(grpcCtx, circuitBreaker, fn)
	case "employee":
		circuitBreaker := h.clientManager.GetEmployeeCircuitBreaker()
		retrier := h.clientManager.GetEmployeeRetrier()
		err = retrier.ExecuteWithCircuitBreaker(grpcCtx, circuitBreaker, fn)
	case "structure":
		circuitBreaker := h.clientManager.GetStructureCircuitBreaker()
		retrier := h.clientManager.GetStructureRetrier()
		err = retrier.ExecuteWithCircuitBreaker(grpcCtx, circuitBreaker, fn)
	default:
		err = fmt.Errorf("unknown service: %s", serviceName)
	}

	duration := time.Since(start)

	// Log the gRPC call completion
	h.logger.LogGRPCCall(ctx, serviceName, "execute", duration, err)

	// Update span with error if present
	if err != nil {
		span.SetTag("error", true)
		span.SetTag("error.message", err.Error())
		h.tracer.FinishSpan(span, err)
	}

	return err
}

// writeErrorResponse writes an error response in JSON format (legacy method for compatibility)
func (h *Handler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorType, message, requestID string) {
	h.errorHandler.HandleGRPCError(w, fmt.Errorf("%s: %s", errorType, message), requestID, "", "")
}

// writeJSONResponse writes a successful JSON response
func (h *Handler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode JSON response: %v", err)
		h.handleGRPCError(w, fmt.Errorf("encoding_error: %v", err), h.getRequestID(nil), "gateway", "writeJSONResponse")
	}
}

// getRequestID extracts or generates a request ID
func (h *Handler) getRequestID(r *http.Request) string {
	if r != nil {
		if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
			return requestID
		}
	}
	return fmt.Sprintf("req_%d", time.Now().UnixNano())
}

// checkServiceAvailability checks if a service client is available and writes error response if not
func (h *Handler) checkServiceAvailability(w http.ResponseWriter, client interface{}, serviceName, requestID string) bool {
	if client == nil {
		// Create a simple error response without using the complex error handler
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		
		errorResp := map[string]interface{}{
			"error":      "service_unavailable",
			"message":    fmt.Sprintf("%s service is not available", serviceName),
			"request_id": requestID,
			"timestamp":  time.Now().UTC().Format(time.RFC3339),
		}
		
		if err := json.NewEncoder(w).Encode(errorResp); err != nil {
			log.Printf("Failed to encode error response: %v", err)
		}
		return false
	}
	return true
}

// createContextWithTimeout creates a context with timeout for gRPC calls
func (h *Handler) createContextWithTimeout(r *http.Request, timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := r.Context()
	
	// The context already has request ID, trace ID, etc. from middleware
	// Just add timeout
	return context.WithTimeout(ctx, timeout)
}

// parseIntParam parses an integer parameter from URL path
func (h *Handler) parseIntParam(param string) (int64, error) {
	if param == "" {
		return 0, fmt.Errorf("parameter is empty")
	}
	return strconv.ParseInt(param, 10, 64)
}

// parseQueryParams parses common query parameters for pagination and sorting
func (h *Handler) parseQueryParams(r *http.Request) (page, limit int32, sortBy, sortOrder string) {
	// Parse page
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = int32(p)
		}
	}
	if page == 0 {
		page = 1
	}

	// Parse limit
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}
	if limit == 0 {
		limit = 10
	}

	// Parse sort_by
	sortBy = r.URL.Query().Get("sort_by")
	if sortBy == "" {
		sortBy = "created_at"
	}

	// Parse sort_order
	sortOrder = r.URL.Query().Get("sort_order")
	if sortOrder == "" {
		sortOrder = "desc"
	}

	return page, limit, sortBy, sortOrder
}