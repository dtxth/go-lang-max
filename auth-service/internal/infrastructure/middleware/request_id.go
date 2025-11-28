package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"auth-service/internal/infrastructure/logger"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
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

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if request ID already exists in header
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = GenerateRequestID()
			}

			// Add request ID to response header
			w.Header().Set("X-Request-ID", requestID)

			// Add request ID to context
			ctx := context.WithValue(r.Context(), RequestIDKey, requestID)

			// Log request start
			start := time.Now()
			log.Info(ctx, "HTTP request started", map[string]interface{}{
				"method": r.Method,
				"path":   r.URL.Path,
				"remote": r.RemoteAddr,
			})

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))

			// Log request completion
			duration := time.Since(start)
			log.Info(ctx, "HTTP request completed", map[string]interface{}{
				"method":       r.Method,
				"path":         r.URL.Path,
				"duration_ms":  duration.Milliseconds(),
				"duration":     duration.String(),
			})
		})
	}
}
