package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	authpb "structure-service/api/proto/authproto"
)

// UserIDKey is the context key for user ID
type userIDContextKey string

const UserIDKey userIDContextKey = "userID"

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AuthMiddleware validates JWT token by calling auth-service via gRPC
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeUnauthorizedError(w, "missing authorization header", requestID)
				return
			}

			// Check for Bearer token format
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				writeUnauthorizedError(w, "invalid authorization header format", requestID)
				return
			}

			token := parts[1]

			// Validate token with auth-service via gRPC
			userID, err := validateTokenWithAuthService(token)
			if err != nil {
				writeUnauthorizedError(w, "invalid or expired token", requestID)
				return
			}

			// Add user ID to context
			ctx := context.WithValue(r.Context(), UserIDKey, userID)

			// Call next handler
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// validateTokenWithAuthService validates token by calling auth-service via gRPC
func validateTokenWithAuthService(token string) (int64, error) {
	authServiceAddr := os.Getenv("AUTH_SERVICE_GRPC_ADDR")
	if authServiceAddr == "" {
		authServiceAddr = "auth-service:9090"
	}

	// Create gRPC connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, authServiceAddr, 
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to connect to auth service: %w", err)
	}
	defer conn.Close()

	// Create auth service client
	client := authpb.NewAuthServiceClient(conn)

	// Make ValidateToken request
	req := &authpb.ValidateTokenRequest{
		Token: token,
	}

	resp, err := client.ValidateToken(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to validate token: %w", err)
	}

	if !resp.Valid {
		return 0, fmt.Errorf("token is invalid")
	}

	return resp.UserId, nil
}

// writeUnauthorizedError writes unauthorized error response
func writeUnauthorizedError(w http.ResponseWriter, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	errorResp := ErrorResponse{
		Error:   "UNAUTHORIZED",
		Code:    "UNAUTHORIZED",
		Message: message,
	}
	
	json.NewEncoder(w).Encode(errorResp)
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}