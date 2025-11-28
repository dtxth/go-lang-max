package http

import (
	"migration-service/internal/infrastructure/middleware"
	"net/http"
)

// SetupRoutes sets up HTTP routes
func SetupRoutes(handler *Handler) http.Handler {
	mux := http.NewServeMux()

	// Migration endpoints
	mux.HandleFunc("/migration/database", handler.StartDatabaseMigration)
	mux.HandleFunc("/migration/google-sheets", handler.StartGoogleSheetsMigration)
	mux.HandleFunc("/migration/excel", handler.StartExcelMigration)
	mux.HandleFunc("/migration/jobs/", handler.GetMigrationJob)
	mux.HandleFunc("/migration/jobs", handler.ListMigrationJobs)

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with request ID middleware
	return middleware.RequestIDMiddleware(mux)
}
