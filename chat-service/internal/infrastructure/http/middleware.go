package http

import (
	"chat-service/internal/domain"
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	tokenInfoKey contextKey = "tokenInfo"
)

// AuthMiddleware проверяет JWT токен и извлекает информацию о пользователе
type AuthMiddleware struct {
	authService domain.AuthService
}

func NewAuthMiddleware(authService domain.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate проверяет токен и добавляет информацию о пользователе в контекст
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем токен из заголовка Authorization
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		// Проверяем формат "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Валидируем токен через Auth Service
		tokenInfo, err := m.authService.ValidateToken(token)
		if err != nil {
			if err == domain.ErrInvalidToken {
				http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			} else {
				http.Error(w, "authentication service error", http.StatusInternalServerError)
			}
			return
		}

		if !tokenInfo.Valid {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		// Добавляем информацию о токене в контекст
		ctx := context.WithValue(r.Context(), tokenInfoKey, tokenInfo)
		next(w, r.WithContext(ctx))
	}
}

// GetTokenInfo извлекает информацию о токене из контекста
func GetTokenInfo(r *http.Request) (*domain.TokenInfo, bool) {
	tokenInfo, ok := r.Context().Value(tokenInfoKey).(*domain.TokenInfo)
	return tokenInfo, ok
}
