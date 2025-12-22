package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	TraceIDKey   contextKey = "trace_id"
	UserIDKey    contextKey = "user_id"
	StartTimeKey contextKey = "start_time"
)

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateTraceID generates a unique trace ID for distributed tracing
func GenerateTraceID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return "unknown"
}

// GetTraceID retrieves the trace ID from context
func GetTraceID(ctx context.Context) string {
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok {
		return traceID
	}
	return "unknown"
}

// GetUserID retrieves the user ID from context
func GetUserID(ctx context.Context) string {
	if userID, ok := ctx.Value(UserIDKey).(string); ok {
		return userID
	}
	return "anonymous"
}

// GetStartTime retrieves the request start time from context
func GetStartTime(ctx context.Context) time.Time {
	if startTime, ok := ctx.Value(StartTimeKey).(time.Time); ok {
		return startTime
	}
	return time.Now()
}

// ContextPropagationMiddleware adds request context and trace IDs to each request
func ContextPropagationMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get or generate request ID
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Get or generate trace ID
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = GenerateTraceID()
			}

			// Get user ID from authorization header if present
			userID := extractUserIDFromAuth(r)

			// Record start time
			startTime := time.Now()

			// Add IDs to response headers
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("X-Trace-ID", traceID)

			// Create context with all values
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
			ctx = context.WithValue(ctx, TraceIDKey, traceID)
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, StartTimeKey, startTime)

			// Call next handler with enriched context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PropagateContextToGRPC creates gRPC metadata from HTTP request context
func PropagateContextToGRPC(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{
		"request-id": GetRequestID(ctx),
		"trace-id":   GetTraceID(ctx),
		"user-id":    GetUserID(ctx),
		"timestamp":  fmt.Sprintf("%d", time.Now().Unix()),
	})

	return metadata.NewOutgoingContext(ctx, md)
}

// extractUserIDFromAuth extracts user ID from authorization header
func extractUserIDFromAuth(r *http.Request) string {
	// This is a placeholder - in a real implementation, you would
	// decode the JWT token or session to get the user ID
	auth := r.Header.Get("Authorization")
	if auth != "" {
		// For now, just return a placeholder indicating auth is present
		return "authenticated_user"
	}
	return "anonymous"
}