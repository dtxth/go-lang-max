package http

import (
	"migration-service/internal/infrastructure/middleware"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
	_ "migration-service/internal/infrastructure/http/docs" // swagger docs
)

// SetupRoutes sets up HTTP routes
func SetupRoutes(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	// Auth middleware
	authMiddleware := middleware.AuthMiddleware()

	// Migration endpoints (с авторизацией)
	mux.Handle("/migration/database", authMiddleware(http.HandlerFunc(handler.StartDatabaseMigration)))
	mux.Handle("/migration/google-sheets", authMiddleware(http.HandlerFunc(handler.StartGoogleSheetsMigration)))
	mux.Handle("/migration/excel", authMiddleware(http.HandlerFunc(handler.StartExcelMigration)))
	mux.Handle("/migration/jobs/", authMiddleware(http.HandlerFunc(handler.HandleJobsRoute)))
	mux.Handle("/migration/jobs", authMiddleware(http.HandlerFunc(handler.ListMigrationJobs)))

	// Health check (без авторизации)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Swagger UI (без авторизации)
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Wrap with CORS middleware (отключен) и request ID middleware
	return middleware.RequestIDMiddleware(nil)(middleware.CORSMiddleware(mux))
}
