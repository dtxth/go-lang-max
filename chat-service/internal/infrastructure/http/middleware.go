package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"chat-service/internal/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	authpb "auth-service/api/proto"
)

type contextKey string
type userIDContextKey string

const (
	tokenInfoKey contextKey = "tokenInfo"
	UserIDKey    userIDContextKey = "userID"
)

// ErrorResponse represents error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// AuthMiddleware проверяет JWT токен через gRPC auth-service
type AuthMiddleware struct{}

func NewAuthMiddleware() *AuthMiddleware {
	return &AuthMiddleware{}
}

// Authenticate проверяет токен и добавляет информацию о пользователе в контекст
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorizedError(w, "missing authorization header")
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeUnauthorizedError(w, "invalid authorization header format")
			return
		}

		token := parts[1]

		// Валидируем токен через gRPC auth-service
		userID, err := validateTokenWithAuthService(token)
		if err != nil {
			writeUnauthorizedError(w, "invalid or expired token")
			return
		}

		// Добавляем информацию о пользователе в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		next(w, r.WithContext(ctx))
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
func writeUnauthorizedError(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	
	errorResp := ErrorResponse{
		Error:   "UNAUTHORIZED",
		Code:    "UNAUTHORIZED",
		Message: message,
	}
	
	json.NewEncoder(w).Encode(errorResp)
}

// GetTokenInfo извлекает информацию о токене из контекста (для обратной совместимости)
func GetTokenInfo(r *http.Request) (*domain.TokenInfo, bool) {
	userID, ok := r.Context().Value(UserIDKey).(int64)
	if !ok {
		return nil, false
	}
	
	// Создаем TokenInfo для обратной совместимости
	tokenInfo := &domain.TokenInfo{
		Valid:  true,
		UserID: userID,
	}
	
	return tokenInfo, true
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}
