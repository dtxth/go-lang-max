package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"maxbot-service/internal/infrastructure/middleware"
	"github.com/gorilla/mux"
)

// Server represents the HTTP server
type Server struct {
	handler *MaxBotHTTPHandler
	port    string
	server  *http.Server
}

// NewServer creates a new HTTP server
func NewServer(handler *MaxBotHTTPHandler, port string) *Server {
	log.Printf("=== CREATING HTTP SERVER ON PORT %s ===", port)
	log.Printf("Handler: %+v", handler)
	server := &Server{
		handler: handler,
		port:    port,
	}
	log.Printf("=== HTTP SERVER CREATED SUCCESSFULLY ===")
	return server
}

// Run starts the HTTP server
func (s *Server) Run() error {
	log.Printf("=== HTTP Server Run() called ===")
	
	// Создаем максимально простой HTTP сервер для диагностики
	simpleMux := http.NewServeMux()
	
	// Простейший endpoint
	simpleMux.HandleFunc("/simple", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Simple endpoint called!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"simple works"}`)
	})
	
	// Health check
	simpleMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Health endpoint called!")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"status":"ok","service":"maxbot-service"}`)
	})
	
	log.Printf("=== Simple mux created with /simple and /health endpoints ===")
	
	s.server = &http.Server{
		Addr:         ":" + s.port,
		Handler:      simpleMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("=== HTTP server starting on port %s ===", s.port)
	log.Printf("Server configuration: Addr=%s", s.server.Addr)
	
	err := s.server.ListenAndServe()
	log.Printf("=== HTTP server ListenAndServe returned with error: %v ===", err)
	return err
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
	log.Printf("=== SETTING UP HTTP ROUTES ===")
	
	if s.handler == nil {
		log.Printf("❌ CRITICAL ERROR: s.handler is nil!")
		return nil
	}
	
	router := mux.NewRouter()
	log.Printf("✅ Created mux.Router")

	// Health check - самый простой endpoint (без авторизации)
	router.HandleFunc("/health", s.healthCheck).Methods("GET")
	log.Printf("✅ Registered /health endpoint")

	// Простой тестовый endpoint без middleware (без авторизации)
	router.HandleFunc("/test-simple", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Simple test endpoint called")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"simple test works"}`)
	}).Methods("GET")
	log.Printf("✅ Registered /test-simple endpoint")

	// Add middleware
	router.Use(s.loggingMiddleware)
	// router.Use(s.corsMiddleware) // CORS отключен
	router.Use(s.requestIDMiddleware)
	log.Printf("✅ Middleware added")

	// API routes with authentication
	api := router.PathPrefix("/api/v1").Subrouter()
	log.Printf("✅ Created API subrouter with prefix /api/v1")
	
	// Auth middleware for API routes
	authMiddleware := middleware.AuthMiddleware()
	
	// Bot endpoints (с авторизацией)
	api.Handle("/me", authMiddleware(http.HandlerFunc(s.handler.GetMe))).Methods("GET")
	log.Printf("✅ Registered /api/v1/me endpoint with auth")
	
	// Chat endpoints (с авторизацией)
	api.Handle("/chats/{chat_id}", authMiddleware(http.HandlerFunc(s.handler.GetChatInfo))).Methods("GET")
	log.Printf("✅ Registered /api/v1/chats/{chat_id} endpoint with auth")
	
	// Profile endpoints (с авторизацией)
	api.Handle("/profiles/{user_id}", authMiddleware(http.HandlerFunc(s.handler.GetProfile))).Methods("GET")
	api.Handle("/profiles/{user_id}", authMiddleware(http.HandlerFunc(s.handler.UpdateProfile))).Methods("PUT")
	api.Handle("/profiles/{user_id}/name", authMiddleware(http.HandlerFunc(s.handler.SetUserProvidedName))).Methods("POST")
	api.Handle("/profiles/stats", authMiddleware(http.HandlerFunc(s.handler.GetProfileStats))).Methods("GET")
	log.Printf("✅ Registered profile endpoints with auth")
	
	// Monitoring endpoints (с авторизацией)
	api.Handle("/monitoring/webhook/stats", authMiddleware(http.HandlerFunc(s.handler.GetWebhookStats))).Methods("GET")
	api.Handle("/monitoring/profiles/coverage", authMiddleware(http.HandlerFunc(s.handler.GetProfileCoverage))).Methods("GET")
	api.Handle("/monitoring/profiles/quality", authMiddleware(http.HandlerFunc(s.handler.GetProfileQualityReport))).Methods("GET")
	log.Printf("✅ Registered monitoring endpoints with auth")
	
	// Webhook endpoint (без авторизации - для внешних систем)
	api.HandleFunc("/webhook/max", s.handler.HandleMaxWebhook).Methods("POST")
	log.Printf("✅ Registered /api/v1/webhook/max endpoint without auth")
	
	// Test endpoint (с авторизацией)
	api.Handle("/test", authMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"message":"test endpoint works"}`)
	}))).Methods("GET")
	log.Printf("✅ Registered /api/v1/test endpoint with auth")

	log.Printf("=== HTTP ROUTES SETUP COMPLETED ===")
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
		
		// Log incoming request details
		log.Printf("[DEBUG] Incoming request: %s %s", r.Method, r.URL.Path)
		
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