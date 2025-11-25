package http

import (
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

func (h *Handler) Router() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/register", h.Register)
	mux.HandleFunc("/login", h.Login)
	mux.HandleFunc("/refresh", h.Refresh)
	mux.HandleFunc("/logout", h.Logout)
	
	// Swagger UI
    mux.Handle("/swagger/", httpSwagger.WrapHandler)

	return mux
}
