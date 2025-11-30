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
	
	// Health check
	mux.HandleFunc("/health", h.Health)
	
	// Swagger UI
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

	// Wrap with request ID middleware
	return middleware.RequestIDMiddleware(nil)(mux)
}
