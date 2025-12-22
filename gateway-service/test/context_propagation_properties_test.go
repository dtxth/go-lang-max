package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"gateway-service/internal/infrastructure/middleware"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// **Feature: gateway-grpc-implementation, Property 4: Request Context Propagation**
// **Validates: Requirements 1.5, 6.3**
//
// Property: For any HTTP request processed by the Gateway, the system should propagate
// request context (including trace IDs) through all gRPC calls
func TestContextPropagationProperty(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	// Setup gopter parameters
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1: Request ID propagation
	properties.Property("request ID is propagated through context", prop.ForAll(
		func(path string, method string) bool {
			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Create a simple test handler that checks context
			contextValid := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				requestID := middleware.GetRequestID(ctx)
				if requestID != "" && requestID != "unknown" {
					contextValid = true
				}
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)
			handler.ServeHTTP(w, req)

			// Verify request ID is set in response header and context
			requestID := w.Header().Get("X-Request-ID")
			return requestID != "" && contextValid
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property 2: Trace ID propagation
	properties.Property("trace ID is propagated through context", prop.ForAll(
		func(path string, method string) bool {
			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Create a simple test handler that checks context
			contextValid := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				traceID := middleware.GetTraceID(ctx)
				if traceID != "" && traceID != "unknown" {
					contextValid = true
				}
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)
			handler.ServeHTTP(w, req)

			// Verify trace ID is set in response header and context
			traceID := w.Header().Get("X-Trace-ID")
			return traceID != "" && contextValid
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Property 3: Existing request ID is preserved
	properties.Property("existing request ID is preserved", prop.ForAll(
		func(path string, method string, requestID string) bool {
			// Create test request with existing request ID
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("X-Request-ID", requestID)
			w := httptest.NewRecorder()

			// Create a simple test handler that checks context
			contextValid := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				ctxRequestID := middleware.GetRequestID(ctx)
				if ctxRequestID == requestID {
					contextValid = true
				}
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)
			handler.ServeHTTP(w, req)

			// Verify the same request ID is returned
			returnedRequestID := w.Header().Get("X-Request-ID")
			return returnedRequestID == requestID && contextValid
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
		gen.Identifier(),
	))

	// Property 4: Existing trace ID is preserved
	properties.Property("existing trace ID is preserved", prop.ForAll(
		func(path string, method string, traceID string) bool {
			// Create test request with existing trace ID
			req := httptest.NewRequest(method, path, nil)
			req.Header.Set("X-Trace-ID", traceID)
			w := httptest.NewRecorder()

			// Create a simple test handler that checks context
			contextValid := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				ctxTraceID := middleware.GetTraceID(ctx)
				if ctxTraceID == traceID {
					contextValid = true
				}
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)
			handler.ServeHTTP(w, req)

			// Verify the same trace ID is returned
			returnedTraceID := w.Header().Get("X-Trace-ID")
			return returnedTraceID == traceID && contextValid
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
		gen.Identifier(),
	))

	// Property 5: Context values are accessible in handlers
	properties.Property("context values are accessible throughout request lifecycle", prop.ForAll(
		func(path string) bool {
			// Create test request
			req := httptest.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			// Track if context was properly set
			contextValid := false

			// Create a test handler that checks context
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				
				// Check if request ID is in context
				requestID := middleware.GetRequestID(ctx)
				if requestID == "" || requestID == "unknown" {
					return
				}

				// Check if trace ID is in context
				traceID := middleware.GetTraceID(ctx)
				if traceID == "" || traceID == "unknown" {
					return
				}

				contextValid = true
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)
			handler.ServeHTTP(w, req)

			return contextValid
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register"),
	))

	// Property 6: Context propagation to gRPC metadata
	properties.Property("context is properly formatted for gRPC propagation", prop.ForAll(
		func(requestID string, traceID string, userID string) bool {
			// Create context with values
			ctx := context.Background()
			ctx = context.WithValue(ctx, middleware.RequestIDKey, requestID)
			ctx = context.WithValue(ctx, middleware.TraceIDKey, traceID)
			ctx = context.WithValue(ctx, middleware.UserIDKey, userID)

			// Propagate to gRPC
			grpcCtx := middleware.PropagateContextToGRPC(ctx)

			// Verify context is not nil
			if grpcCtx == nil {
				return false
			}

			// Verify original values are still accessible
			if middleware.GetRequestID(grpcCtx) != requestID {
				return false
			}
			if middleware.GetTraceID(grpcCtx) != traceID {
				return false
			}
			if middleware.GetUserID(grpcCtx) != userID {
				return false
			}

			return true
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.OneConstOf("anonymous", "authenticated_user", "admin"),
	))

	// Run all properties
	properties.TestingRun(t)
}

// TestContextPropagationMiddleware tests the context propagation middleware directly
func TestContextPropagationMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		existingReqID  string
		existingTraceID string
		expectNewReqID bool
		expectNewTraceID bool
	}{
		{
			name:             "generates new IDs when none provided",
			existingReqID:    "",
			existingTraceID:  "",
			expectNewReqID:   true,
			expectNewTraceID: true,
		},
		{
			name:             "preserves existing request ID",
			existingReqID:    "existing-req-123",
			existingTraceID:  "",
			expectNewReqID:   false,
			expectNewTraceID: true,
		},
		{
			name:             "preserves existing trace ID",
			existingReqID:    "",
			existingTraceID:  "existing-trace-456",
			expectNewReqID:   true,
			expectNewTraceID: false,
		},
		{
			name:             "preserves both existing IDs",
			existingReqID:    "existing-req-123",
			existingTraceID:  "existing-trace-456",
			expectNewReqID:   false,
			expectNewTraceID: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test handler
			var capturedReqID, capturedTraceID string
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx := r.Context()
				capturedReqID = middleware.GetRequestID(ctx)
				capturedTraceID = middleware.GetTraceID(ctx)
				w.WriteHeader(http.StatusOK)
			})

			// Apply middleware
			handler := middleware.ContextPropagationMiddleware()(testHandler)

			// Create request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingReqID != "" {
				req.Header.Set("X-Request-ID", tt.existingReqID)
			}
			if tt.existingTraceID != "" {
				req.Header.Set("X-Trace-ID", tt.existingTraceID)
			}

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			// Verify request ID
			if tt.expectNewReqID {
				assert.NotEmpty(t, capturedReqID, "Request ID should be generated")
				assert.NotEqual(t, tt.existingReqID, capturedReqID, "Should generate new request ID")
			} else {
				assert.Equal(t, tt.existingReqID, capturedReqID, "Should preserve existing request ID")
			}

			// Verify trace ID
			if tt.expectNewTraceID {
				assert.NotEmpty(t, capturedTraceID, "Trace ID should be generated")
				assert.NotEqual(t, tt.existingTraceID, capturedTraceID, "Should generate new trace ID")
			} else {
				assert.Equal(t, tt.existingTraceID, capturedTraceID, "Should preserve existing trace ID")
			}

			// Verify response headers
			assert.Equal(t, capturedReqID, w.Header().Get("X-Request-ID"), "Request ID should be in response header")
			assert.Equal(t, capturedTraceID, w.Header().Get("X-Trace-ID"), "Trace ID should be in response header")
		})
	}
}
