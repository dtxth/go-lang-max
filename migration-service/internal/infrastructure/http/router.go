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

	// Migration endpoints
	mux.HandleFunc("/migration/database", handler.StartDatabaseMigration)
	mux.HandleFunc("/migration/google-sheets", handler.StartGoogleSheetsMigration)
	mux.HandleFunc("/migration/excel", handler.StartExcelMigration)
	mux.HandleFunc("/migration/jobs/", handler.HandleJobsRoute)
	mux.HandleFunc("/migration/jobs", handler.ListMigrationJobs)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy"}`))
	})

	// Swagger UI
	mux.HandleFunc("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// Wrap with CORS middleware (отключен) и request ID middleware
	return middleware.RequestIDMiddleware(nil)(middleware.CORSMiddleware(mux))
}
