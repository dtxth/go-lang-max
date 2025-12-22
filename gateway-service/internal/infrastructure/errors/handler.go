package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ErrorResponse represents an HTTP error response
type ErrorResponse struct {
	Error     string `json:"error"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
	Service   string `json:"service,omitempty"`
	Method    string `json:"method,omitempty"`
	Timestamp string `json:"timestamp"`
}

// ErrorHandler handles gRPC errors and converts them to HTTP responses
type ErrorHandler struct {
	// grpcToHTTPStatusMap maps gRPC status codes to HTTP status codes
	grpcToHTTPStatusMap map[codes.Code]int
	// logger for structured logging (optional, can be nil for backward compatibility)
	logger Logger
}

// Logger interface for structured logging
type Logger interface {
	LogGRPCError(ctx context.Context, service, method string, err error)
	LogConnectionError(ctx context.Context, service string, err error)
}

// NewErrorHandler creates a new error handler
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		grpcToHTTPStatusMap: map[codes.Code]int{
			codes.OK:                 http.StatusOK,
			codes.Canceled:           http.StatusRequestTimeout,
			codes.Unknown:            http.StatusInternalServerError,
			codes.InvalidArgument:    http.StatusBadRequest,
			codes.DeadlineExceeded:   http.StatusRequestTimeout,
			codes.NotFound:           http.StatusNotFound,
			codes.AlreadyExists:      http.StatusConflict,
			codes.PermissionDenied:   http.StatusForbidden,
			codes.ResourceExhausted:  http.StatusTooManyRequests,
			codes.FailedPrecondition: http.StatusBadRequest,
			codes.Aborted:            http.StatusConflict,
			codes.OutOfRange:         http.StatusBadRequest,
			codes.Unimplemented:      http.StatusNotImplemented,
			codes.Internal:           http.StatusInternalServerError,
			codes.Unavailable:        http.StatusServiceUnavailable,
			codes.DataLoss:           http.StatusInternalServerError,
			codes.Unauthenticated:    http.StatusUnauthorized,
		},
	}
}

// NewErrorHandlerWithLogger creates a new error handler with structured logging
func NewErrorHandlerWithLogger(logger Logger) *ErrorHandler {
	eh := NewErrorHandler()
	eh.logger = logger
	return eh
}

// HandleGRPCError converts gRPC errors to HTTP responses with detailed logging
func (eh *ErrorHandler) HandleGRPCError(w http.ResponseWriter, err error, requestID, serviceName, methodName string) {
	if err == nil {
		return
	}

	// Create context with request ID for logging
	ctx := context.WithValue(context.Background(), "request_id", requestID)

	// Log the error with structured logging if available, fallback to standard logging
	if eh.logger != nil {
		eh.logger.LogGRPCError(ctx, serviceName, methodName, err)
	} else {
		eh.logError(err, requestID, serviceName, methodName)
	}

	// Handle circuit breaker errors
	if eh.isCircuitBreakerError(err) {
		eh.writeErrorResponse(w, http.StatusServiceUnavailable, "circuit_breaker_open", 
			"Service temporarily unavailable", requestID, serviceName, methodName)
		return
	}

	// Handle connection errors
	if eh.isConnectionError(err) {
		eh.writeErrorResponse(w, http.StatusServiceUnavailable, "connection_error", 
			"Service connection failed", requestID, serviceName, methodName)
		return
	}

	// Handle gRPC status errors
	st, ok := status.FromError(err)
	if !ok {
		// Not a gRPC error, treat as internal server error
		eh.writeErrorResponse(w, http.StatusInternalServerError, "internal_error", 
			err.Error(), requestID, serviceName, methodName)
		return
	}

	httpStatus, exists := eh.grpcToHTTPStatusMap[st.Code()]
	if !exists {
		httpStatus = http.StatusInternalServerError
	}

	eh.writeErrorResponse(w, httpStatus, st.Code().String(), st.Message(), requestID, serviceName, methodName)
}

// HandleConnectionError handles gRPC connection errors gracefully
func (eh *ErrorHandler) HandleConnectionError(w http.ResponseWriter, err error, requestID, serviceName string) {
	// Create context with request ID for logging
	ctx := context.WithValue(context.Background(), "request_id", requestID)

	// Log the connection error with structured logging if available
	if eh.logger != nil {
		eh.logger.LogConnectionError(ctx, serviceName, err)
	} else {
		eh.logConnectionError(err, requestID, serviceName)
	}

	eh.writeErrorResponse(w, http.StatusServiceUnavailable, "service_unavailable", 
		fmt.Sprintf("%s service is currently unavailable", serviceName), requestID, serviceName, "")
}

// writeErrorResponse writes an error response in JSON format
func (eh *ErrorHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorType, message, requestID, serviceName, methodName string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:     errorType,
		Message:   message,
		RequestID: requestID,
		Service:   serviceName,
		Method:    methodName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}

// logError logs detailed error information
func (eh *ErrorHandler) logError(err error, requestID, serviceName, methodName string) {
	st, ok := status.FromError(err)
	if ok {
		log.Printf("gRPC Error - RequestID: %s, Service: %s, Method: %s, Code: %s, Message: %s", 
			requestID, serviceName, methodName, st.Code().String(), st.Message())
	} else {
		log.Printf("Error - RequestID: %s, Service: %s, Method: %s, Error: %v", 
			requestID, serviceName, methodName, err)
	}
}

// logConnectionError logs connection-specific errors
func (eh *ErrorHandler) logConnectionError(err error, requestID, serviceName string) {
	log.Printf("Connection Error - RequestID: %s, Service: %s, Error: %v", 
		requestID, serviceName, err)
}

// isCircuitBreakerError checks if the error is from a circuit breaker
func (eh *ErrorHandler) isCircuitBreakerError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "circuit breaker is open") ||
		   strings.Contains(err.Error(), "circuit breaker") ||
		   strings.Contains(err.Error(), "max requests exceeded")
}

// isConnectionError checks if the error is a connection-related error
func (eh *ErrorHandler) isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection") ||
		   strings.Contains(errStr, "dial") ||
		   strings.Contains(errStr, "network") ||
		   strings.Contains(errStr, "timeout") ||
		   strings.Contains(errStr, "unavailable")
}

// GetHTTPStatusFromGRPCCode returns the HTTP status code for a gRPC code
func (eh *ErrorHandler) GetHTTPStatusFromGRPCCode(code codes.Code) int {
	if status, exists := eh.grpcToHTTPStatusMap[code]; exists {
		return status
	}
	return http.StatusInternalServerError
}

// IsRetryableError determines if an error should trigger a retry
func (eh *ErrorHandler) IsRetryableError(err error) bool {
	if eh.isCircuitBreakerError(err) {
		return false // Don't retry if circuit breaker is open
	}

	st, ok := status.FromError(err)
	if !ok {
		// Non-gRPC errors might be connection issues, so retry
		return eh.isConnectionError(err)
	}

	// Retry on specific gRPC codes
	switch st.Code() {
	case codes.Unavailable, codes.DeadlineExceeded, codes.ResourceExhausted, codes.Aborted:
		return true
	default:
		return false
	}
}