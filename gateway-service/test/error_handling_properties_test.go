package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gateway-service/internal/config"
	"gateway-service/internal/infrastructure/errors"
	grpcClient "gateway-service/internal/infrastructure/grpc"
	httpHandler "gateway-service/internal/infrastructure/http"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// **Feature: gateway-grpc-implementation, Property 3: Error Handling and Status Code Mapping**
// **Validates: Requirements 1.4, 6.4**
func TestErrorHandlingAndStatusCodeMapping(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Create test configuration
	cfg := &config.Config{
		Services: config.ServicesConfig{
			Auth: config.ServiceConfig{
				Address:           "localhost:50051",
				Timeout:           5 * time.Second,
				MaxRetries:        3,
				RetryDelay:        100 * time.Millisecond,
				MaxRetryDelay:     5 * time.Second,
				BackoffMultiplier: 2.0,
				CircuitBreaker: config.CircuitBreakerConfig{
					MaxRequests: 10,
					Interval:    60 * time.Second,
					Timeout:     60 * time.Second,
				},
			},
		},
	}

	// Create client manager and handler
	clientManager := grpcClient.NewClientManager(cfg)
	handler := httpHandler.NewHandler(cfg, clientManager)
	errorHandler := errors.NewErrorHandler()

	// Property: For any gRPC error, the system should return the appropriate HTTP status code
	properties.Property("gRPC errors map to correct HTTP status codes", prop.ForAll(
		func(grpcCode codes.Code, errorMessage string) bool {
			// Skip OK status as it's not an error
			if grpcCode == codes.OK {
				return true
			}
			
			// Create a gRPC error
			grpcErr := status.Error(grpcCode, errorMessage)
			
			// Create a test HTTP response recorder
			w := httptest.NewRecorder()
			
			// Handle the error
			requestID := "test-request-123"
			serviceName := "test-service"
			methodName := "test-method"
			
			handler.HandleGRPCError(w, grpcErr, requestID, serviceName, methodName)
			
			// Verify the HTTP status code is appropriate for the gRPC code
			expectedHTTPStatus := errorHandler.GetHTTPStatusFromGRPCCode(grpcCode)
			actualHTTPStatus := w.Code
			
			if actualHTTPStatus != expectedHTTPStatus {
				t.Logf("Expected HTTP status %d for gRPC code %s, got %d", 
					expectedHTTPStatus, grpcCode.String(), actualHTTPStatus)
				return false
			}
			
			// Verify response is JSON
			contentType := w.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Logf("Expected JSON content type, got %s", contentType)
				return false
			}
			
			// Verify response body contains error information
			body := w.Body.String()
			if body == "" {
				t.Logf("Expected non-empty response body")
				return false
			}
			
			// Verify response contains expected fields
			if !strings.Contains(body, "error") || !strings.Contains(body, "message") {
				t.Logf("Response body missing required fields: %s", body)
				return false
			}
			
			return true
		},
		genGRPCCode(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property: Circuit breaker errors should return 503 Service Unavailable
	properties.Property("circuit breaker errors return 503", prop.ForAll(
		func(errorMessage string) bool {
			// Create a circuit breaker error
			cbErr := fmt.Errorf("circuit breaker is open: %s", errorMessage)
			
			// Create a test HTTP response recorder
			w := httptest.NewRecorder()
			
			// Handle the error
			requestID := "test-request-456"
			serviceName := "test-service"
			methodName := "test-method"
			
			handler.HandleGRPCError(w, cbErr, requestID, serviceName, methodName)
			
			// Verify 503 status code
			if w.Code != http.StatusServiceUnavailable {
				t.Logf("Expected 503 for circuit breaker error, got %d", w.Code)
				return false
			}
			
			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property: Connection errors should return 503 Service Unavailable
	properties.Property("connection errors return 503", prop.ForAll(
		func(errorMessage string) bool {
			// Create a connection error
			connErr := fmt.Errorf("connection failed: %s", errorMessage)
			
			// Create a test HTTP response recorder
			w := httptest.NewRecorder()
			
			// Handle the connection error
			requestID := "test-request-789"
			serviceName := "test-service"
			
			errorHandler.HandleConnectionError(w, connErr, requestID, serviceName)
			
			// Verify 503 status code
			if w.Code != http.StatusServiceUnavailable {
				t.Logf("Expected 503 for connection error, got %d", w.Code)
				return false
			}
			
			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	// Property: Error responses should contain detailed logging information
	properties.Property("error responses contain detailed information", prop.ForAll(
		func(grpcCode codes.Code, errorMessage string) bool {
			// Create a gRPC error
			grpcErr := status.Error(grpcCode, errorMessage)
			
			// Create a test HTTP response recorder
			w := httptest.NewRecorder()
			
			// Handle the error with fixed service and method names
			requestID := "test-request-123"
			serviceName := "test-service"
			methodName := "test-method"
			
			handler.HandleGRPCError(w, grpcErr, requestID, serviceName, methodName)
			
			// Verify response contains all required fields
			body := w.Body.String()
			
			// Check for required JSON fields
			requiredFields := []string{"error", "message", "timestamp"}
			for _, field := range requiredFields {
				if !strings.Contains(body, field) {
					t.Logf("Response missing required field '%s': %s", field, body)
					return false
				}
			}
			
			// Service and method should be in the response
			if !strings.Contains(body, serviceName) {
				t.Logf("Response missing service name '%s': %s", serviceName, body)
				return false
			}
			
			if !strings.Contains(body, methodName) {
				t.Logf("Response missing method name '%s': %s", methodName, body)
				return false
			}
			
			return true
		},
		genGRPCErrorCode(), // Use error codes only
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 100 }),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genGRPCErrorCode generates valid gRPC error status codes (excluding OK)
func genGRPCErrorCode() gopter.Gen {
	errorCodes := []codes.Code{
		codes.Canceled,
		codes.Unknown,
		codes.InvalidArgument,
		codes.DeadlineExceeded,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unimplemented,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss,
		codes.Unauthenticated,
	}
	
	return gen.IntRange(0, len(errorCodes)-1).Map(func(i int) codes.Code {
		return errorCodes[i]
	})
}

// genGRPCCode generates valid gRPC status codes (including OK)
func genGRPCCode() gopter.Gen {
	validCodes := []codes.Code{
		codes.OK,
		codes.Canceled,
		codes.Unknown,
		codes.InvalidArgument,
		codes.DeadlineExceeded,
		codes.NotFound,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.ResourceExhausted,
		codes.FailedPrecondition,
		codes.Aborted,
		codes.OutOfRange,
		codes.Unimplemented,
		codes.Internal,
		codes.Unavailable,
		codes.DataLoss,
		codes.Unauthenticated,
	}
	
	return gen.IntRange(0, len(validCodes)-1).Map(func(i int) codes.Code {
		return validCodes[i]
	})
}