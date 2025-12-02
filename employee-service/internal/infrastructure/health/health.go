package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"
)

// HealthStatus represents the health status of the service
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks"`
}

// Handler provides health check endpoints
type Handler struct {
	db *sql.DB
}

// NewHandler creates a new health check handler
func NewHandler(db *sql.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// HealthCheck performs a basic health check
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    make(map[string]string),
	}

	// Check database connectivity
	if h.db != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := h.db.PingContext(ctx); err != nil {
			status.Status = "unhealthy"
			status.Checks["database"] = "failed: " + err.Error()
		} else {
			status.Checks["database"] = "ok"
		}
	}

	// Set response status code
	statusCode := http.StatusOK
	if status.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(status)
}

// ReadinessCheck checks if the service is ready to accept traffic
func (h *Handler) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// For now, readiness is the same as health
	// In the future, this could check additional dependencies
	h.HealthCheck(w, r)
}

// LivenessCheck checks if the service is alive
func (h *Handler) LivenessCheck(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - just return OK
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":    "alive",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
