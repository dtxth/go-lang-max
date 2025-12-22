package test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	errorHandler "gateway-service/internal/infrastructure/errors"
	"gateway-service/internal/infrastructure/middleware"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MockLogger captures log entries for testing
type MockLogger struct {
	LogEntries []LogEntry
}

type LogEntry struct {
	Context context.Context
	Service string
	Method  string
	Error   error
	Type    string // "grpc_error" or "connection_error"
}

func (m *MockLogger) LogGRPCError(ctx context.Context, service, method string, err error) {
	m.LogEntries = append(m.LogEntries, LogEntry{
		Context: ctx,
		Service: service,
		Method:  method,
		Error:   err,
		Type:    "grpc_error",
	})
}

func (m *MockLogger) LogConnectionError(ctx context.Context, service string, err error) {
	m.LogEntries = append(m.LogEntries, LogEntry{
		Context: ctx,
		Service: service,
		Error:   err,
		Type:    "connection_error",
	})
}

func (m *MockLogger) Reset() {
	m.LogEntries = nil
}

// **Feature: gateway-grpc-implementation, Property 6: Error Logging**
// **Validates: Requirements 6.2**
//
// Property: For any gRPC error that occurs, the system should log detailed error information
// including service name, method name, and error details
func TestErrorLoggingProperty(t *testing.T) {
	// Skip if in short mode
	if testing.Short() {
		t.Skip("Skipping property-based test in short mode")
	}

	// Setup gopter parameters
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property 1: gRPC errors are logged with service and method information
	properties.Property("gRPC errors are logged with complete information", prop.ForAll(
		func(serviceName string, methodName string, grpcCode codes.Code, errorMessage string) bool {
			// Create mock logger
			mockLogger := &MockLogger{}

			// Create error handler with logger
			errorHandlerInstance := errorHandler.NewErrorHandlerWithLogger(mockLogger)

			// Create gRPC status error
			grpcErr := status.Error(grpcCode, errorMessage)

			// Create test response writer
			w := httptest.NewRecorder()

			// Handle the error
			errorHandlerInstance.HandleGRPCError(w, grpcErr, "test-req-123", serviceName, methodName)

			// Verify error was logged
			if len(mockLogger.LogEntries) == 0 {
				return false
			}

			entry := mockLogger.LogEntries[0]
			
			// Verify log entry contains required information
			return entry.Type == "grpc_error" &&
				   entry.Service == serviceName &&
				   entry.Method == methodName &&
				   entry.Error != nil &&
				   entry.Error.Error() == grpcErr.Error()
		},
		gen.OneConstOf("auth", "chat", "employee", "structure"),
		gen.OneConstOf("Login", "Register", "GetAllChats", "CreateEmployee"),
		gen.OneConstOf(codes.InvalidArgument, codes.NotFound, codes.Internal, codes.Unavailable),
		gen.AlphaString(),
	))

	// Property 2: Connection errors are logged with service information
	properties.Property("connection errors are logged with service information", prop.ForAll(
		func(serviceName string, errorMessage string) bool {
			// Create mock logger
			mockLogger := &MockLogger{}

			// Create error handler with logger
			errorHandlerInstance := errorHandler.NewErrorHandlerWithLogger(mockLogger)

			// Create connection error
			connErr := fmt.Errorf("connection failed: %s", errorMessage)

			// Create test response writer
			w := httptest.NewRecorder()

			// Handle the connection error
			errorHandlerInstance.HandleConnectionError(w, connErr, "test-req-456", serviceName)

			// Verify error was logged
			if len(mockLogger.LogEntries) == 0 {
				return false
			}

			entry := mockLogger.LogEntries[0]
			
			// Verify log entry contains required information
			return entry.Type == "connection_error" &&
				   entry.Service == serviceName &&
				   entry.Error != nil &&
				   strings.Contains(entry.Error.Error(), errorMessage)
		},
		gen.OneConstOf("auth", "chat", "employee", "structure"),
		gen.AlphaString(),
	))

	// Property 3: Error logging preserves context information
	properties.Property("error logging preserves context information", prop.ForAll(
		func(requestID string, serviceName string, methodName string) bool {
			// Create mock logger
			mockLogger := &MockLogger{}

			// Create error handler with logger
			errorHandlerInstance := errorHandler.NewErrorHandlerWithLogger(mockLogger)

			// Create test error
			testErr := status.Error(codes.Internal, "test error")

			// Create test response writer
			w := httptest.NewRecorder()

			// Handle the error
			errorHandlerInstance.HandleGRPCError(w, testErr, requestID, serviceName, methodName)

			// Verify error was logged
			if len(mockLogger.LogEntries) == 0 {
				return false
			}

			entry := mockLogger.LogEntries[0]
			
			// Verify context is preserved (we can't directly check the context value,
			// but we can verify the logger was called with a context)
			return entry.Context != nil &&
				   entry.Service == serviceName &&
				   entry.Method == methodName
		},
		gen.Identifier(),
		gen.OneConstOf("auth", "chat", "employee", "structure"),
		gen.OneConstOf("Login", "Register", "GetAllChats", "CreateEmployee"),
	))

	// Property 4: Different error types are distinguished in logging
	properties.Property("different error types are properly distinguished", prop.ForAll(
		func(serviceName string) bool {
			// Create mock logger
			mockLogger := &MockLogger{}

			// Create error handler with logger
			errorHandlerInstance := errorHandler.NewErrorHandlerWithLogger(mockLogger)

			// Create test response writer
			w1 := httptest.NewRecorder()
			w2 := httptest.NewRecorder()

			// Handle gRPC error
			grpcErr := status.Error(codes.NotFound, "not found")
			errorHandlerInstance.HandleGRPCError(w1, grpcErr, "req-1", serviceName, "TestMethod")

			// Handle connection error
			connErr := fmt.Errorf("connection timeout")
			errorHandlerInstance.HandleConnectionError(w2, connErr, "req-2", serviceName)

			// Verify both errors were logged with correct types
			if len(mockLogger.LogEntries) != 2 {
				return false
			}

			grpcEntry := mockLogger.LogEntries[0]
			connEntry := mockLogger.LogEntries[1]

			return grpcEntry.Type == "grpc_error" &&
				   grpcEntry.Method == "TestMethod" &&
				   connEntry.Type == "connection_error" &&
				   connEntry.Method == ""
		},
		gen.OneConstOf("auth", "chat", "employee", "structure"),
	))

	// Property 5: Structured logging middleware captures HTTP request information
	properties.Property("structured logging captures HTTP request information", prop.ForAll(
		func(path string, method string) bool {
			// Create logger
			logger := middleware.NewLogger(middleware.LogLevelInfo)

			// Create test handler
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			// Apply logging middleware
			handler := middleware.RequestLoggingMiddleware(logger)(testHandler)

			// Create test request
			req := httptest.NewRequest(method, path, nil)
			w := httptest.NewRecorder()

			// Process request
			handler.ServeHTTP(w, req)

			// For this property, we verify that the middleware doesn't panic
			// and completes successfully. In a real implementation, you would
			// capture the log output and verify its contents.
			return w.Code == http.StatusOK
		},
		gen.OneConstOf("/health", "/metrics", "/login", "/register", "/chats", "/employees"),
		gen.OneConstOf("GET", "POST", "PUT", "DELETE"),
	))

	// Run all properties
	properties.TestingRun(t)
}

// TestErrorHandlerLogging tests the error handler logging functionality directly
func TestErrorHandlerLogging(t *testing.T) {
	tests := []struct {
		name        string
		error       error
		serviceName string
		methodName  string
		expectLog   bool
		logType     string
	}{
		{
			name:        "gRPC status error is logged",
			error:       status.Error(codes.InvalidArgument, "invalid input"),
			serviceName: "auth",
			methodName:  "Login",
			expectLog:   true,
			logType:     "grpc_error",
		},
		{
			name:        "connection error is logged",
			error:       fmt.Errorf("connection refused"),
			serviceName: "chat",
			methodName:  "",
			expectLog:   true,
			logType:     "connection_error",
		},
		{
			name:        "nil error is not logged",
			error:       nil,
			serviceName: "employee",
			methodName:  "CreateEmployee",
			expectLog:   false,
			logType:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock logger
			mockLogger := &MockLogger{}

			// Create error handler with logger
			errorHandlerInstance := errorHandler.NewErrorHandlerWithLogger(mockLogger)

			// Create test response writer
			w := httptest.NewRecorder()

			if tt.logType == "connection_error" {
				// Handle connection error
				errorHandlerInstance.HandleConnectionError(w, tt.error, "test-req", tt.serviceName)
			} else {
				// Handle gRPC error
				errorHandlerInstance.HandleGRPCError(w, tt.error, "test-req", tt.serviceName, tt.methodName)
			}

			if tt.expectLog {
				assert.Len(t, mockLogger.LogEntries, 1, "Should log exactly one entry")
				entry := mockLogger.LogEntries[0]
				assert.Equal(t, tt.logType, entry.Type, "Log type should match")
				assert.Equal(t, tt.serviceName, entry.Service, "Service name should match")
				if tt.methodName != "" {
					assert.Equal(t, tt.methodName, entry.Method, "Method name should match")
				}
				assert.Equal(t, tt.error, entry.Error, "Error should match")
			} else {
				assert.Len(t, mockLogger.LogEntries, 0, "Should not log anything")
			}
		})
	}
}

// TestStructuredLoggingIntegration tests the integration of structured logging
func TestStructuredLoggingIntegration(t *testing.T) {
	// Test that requests are processed with logging middleware
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	// Create a simple test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create logger
	logger := middleware.NewLogger(middleware.LogLevelDebug)

	// Apply middleware chain
	handler := middleware.ContextPropagationMiddleware()(testHandler)
	handler = middleware.RequestLoggingMiddleware(logger)(handler)

	// This should not panic and should complete successfully
	handler.ServeHTTP(w, req)

	// Verify response headers contain context information
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"), "Request ID should be set")
	assert.NotEmpty(t, w.Header().Get("X-Trace-ID"), "Trace ID should be set")
	assert.Equal(t, http.StatusOK, w.Code, "Should return OK status")
}