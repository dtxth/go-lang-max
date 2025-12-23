package http

import (
	"auth-service/internal/infrastructure/middleware"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/login-phone", h.LoginByPhone)
	mux.HandleFunc("/refresh", h.Refresh)
	mux.HandleFunc("/logout", h.Logout)
	mux.HandleFunc("/auth/max", h.AuthenticateMAX)
	
	// Password management endpoints
	mux.HandleFunc("/auth/password-reset/request", h.RequestPasswordReset)
	mux.HandleFunc("/auth/password-reset/confirm", h.ResetPassword)
	
	// Protected password change endpoint (requires authentication)
	changePasswordHandler := middleware.AuthMiddleware(h.auth)(http.HandlerFunc(h.ChangePassword))
	mux.Handle("/auth/password/change", changePasswordHandler)
	
	// Health check and metrics
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/metrics", h.GetMetrics)
	
	// Bot endpoints
	mux.HandleFunc("/bot/me", h.GetBotMe)
	
	// Token validation endpoint for other services
	mux.HandleFunc("/validate-token", h.ValidateToken)
	
	// Swagger UI
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap with CORS middleware (отключен) и request ID middleware
	return middleware.RequestIDMiddleware(nil)(middleware.CORSMiddleware(mux))
}
