package middleware

import (
	"context"
	"net/http"
	"strings"

	"auth-service/internal/infrastructure/errors"
	"auth-service/internal/usecase"
)

// AuthMiddleware validates JWT token and adds user ID to context
func AuthMiddleware(authService *usecase.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteError(w, errors.UnauthorizedError("missing authorization header"), requestID)
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				errors.WriteError(w, errors.UnauthorizedError("invalid authorization header format"), requestID)
				return
			}

			token := parts[1]

			// Validate token
			userID, _, _, _, err := authService.ValidateTokenWithContext(token)
			if err != nil {
				errors.WriteError(w, errors.UnauthorizedError("invalid or expired token"), requestID)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
