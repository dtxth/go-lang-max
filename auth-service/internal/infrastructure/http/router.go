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
	mux.HandleFunc("/refresh", h.Refresh)
	mux.HandleFunc("/logout", h.Logout)
	
	// Password management endpoints
	mux.HandleFunc("/auth/password-reset/request", h.RequestPasswordReset)
	mux.HandleFunc("/auth/password-reset/confirm", h.ResetPassword)
	
	// Protected password change endpoint (requires authentication)
	changePasswordHandler := middleware.AuthMiddleware(h.auth)(http.HandlerFunc(h.ChangePassword))
	mux.Handle("/auth/password/change", changePasswordHandler)
	
	// Health check and metrics
	mux.HandleFunc("/health", h.Health)
	mux.HandleFunc("/metrics", h.GetMetrics)
	
	// Swagger UI
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap with request ID middleware
	return middleware.RequestIDMiddleware(nil)(mux)
}
