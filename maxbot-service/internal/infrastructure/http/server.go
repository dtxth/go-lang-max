package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

// Server represents the HTTP server
type Server struct {
	handler *MaxBotHTTPHandler
	port    string
	server  *http.Server
}

// NewServer creates a new HTTP server
func NewServer(handler *MaxBotHTTPHandler, port string) *Server {
	return &Server{
		handler: handler,
		port:    port,
	}
}

// Run starts the HTTP server
func (s *Server) Run() error {
	router := s.setupRoutes()
	
	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("HTTP server starting on port %s", s.port)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server
func (s *Server) Shutdown(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// setupRoutes configures the HTTP routes
func (s *Server) setupRoutes() *mux.Router {
	router := mux.NewRouter()

	// Add middleware
	router.Use(s.loggingMiddleware)
	router.Use(s.corsMiddleware)
	router.Use(s.requestIDMiddleware)

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()
	
	// Bot endpoints
	api.HandleFunc("/me", s.handler.GetMe).Methods("GET")
	
	// Webhook endpoints
	api.HandleFunc("/webhook/max", s.handler.HandleMaxWebhook).Methods("POST")
	
	// Profile management endpoints (Requirements 5.4, 5.5)
	api.HandleFunc("/profiles/stats", s.handler.GetProfileStats).Methods("GET")
	api.HandleFunc("/profiles/{user_id}", s.handler.GetProfile).Methods("GET")
	api.HandleFunc("/profiles/{user_id}", s.handler.UpdateProfile).Methods("PUT")
	api.HandleFunc("/profiles/{user_id}/name", s.handler.SetUserProvidedName).Methods("POST")

	// Monitoring and analytics endpoints (Requirements 6.1, 6.3, 6.4)
	api.HandleFunc("/monitoring/webhook/stats", s.handler.GetWebhookStats).Methods("GET")
	api.HandleFunc("/monitoring/profiles/coverage", s.handler.GetProfileCoverage).Methods("GET")
	api.HandleFunc("/monitoring/profiles/quality", s.handler.GetProfileQualityReport).Methods("GET")

	// Health check
	router.HandleFunc("/health", s.healthCheck).Methods("GET")

	// Swagger documentation
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return router
}

// healthCheck provides a simple health check endpoint
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok","service":"maxbot-service"}`)
}

// loggingMiddleware logs HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Create a response writer wrapper to capture status code
		wrapper := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(wrapper, r)
		
		duration := time.Since(start)
		log.Printf("HTTP %s %s %d %v", r.Method, r.URL.Path, wrapper.statusCode, duration)
	})
}

// corsMiddleware adds CORS headers
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// requestIDMiddleware adds a request ID to the context
func (s *Server) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), "request_id", requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID generates a simple request ID
func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}