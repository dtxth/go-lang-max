package http

import (
	"encoding/json"
	"io"
	"log"
	"migration-service/internal/domain"
	"migration-service/internal/usecase"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Handler handles HTTP requests for migration service
type Handler struct {
	databaseUseCase     *usecase.MigrateFromDatabaseUseCase
	googleSheetsUseCase *usecase.MigrateFromGoogleSheetsUseCase
	excelUseCase        *usecase.MigrateFromExcelUseCase
	jobRepo             domain.MigrationJobRepository
	errorRepo           domain.MigrationErrorRepository
}

// NewHandler creates a new HTTP handler
func NewHandler(
	databaseUseCase *usecase.MigrateFromDatabaseUseCase,
	googleSheetsUseCase *usecase.MigrateFromGoogleSheetsUseCase,
	excelUseCase *usecase.MigrateFromExcelUseCase,
	jobRepo domain.MigrationJobRepository,
	errorRepo domain.MigrationErrorRepository,
) *Handler {
	return &Handler{
		databaseUseCase:     databaseUseCase,
		googleSheetsUseCase: googleSheetsUseCase,
		excelUseCase:        excelUseCase,
		jobRepo:             jobRepo,
		errorRepo:           errorRepo,
	}
}

// StartDatabaseMigrationRequest represents the request to start database migration
type StartDatabaseMigrationRequest struct {
	SourceIdentifier string `json:"source_identifier"`
}

// StartGoogleSheetsMigrationRequest represents the request to start Google Sheets migration
type StartGoogleSheetsMigrationRequest struct {
	SpreadsheetID string `json:"spreadsheet_id"`
}

// MigrationJobResponse represents the response for a migration job
type MigrationJobResponse struct {
	ID               int     `json:"id"`
	SourceType       string  `json:"source_type"`
	SourceIdentifier string  `json:"source_identifier"`
	Status           string  `json:"status"`
	Total            int     `json:"total"`
	Processed        int     `json:"processed"`
	Failed           int     `json:"failed"`
	StartedAt        string  `json:"started_at"`
	CompletedAt      *string `json:"completed_at,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// StartDatabaseMigration handles POST /migration/database
func (h *Handler) StartDatabaseMigration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartDatabaseMigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Start migration in background
	go func() {
		jobID, err := h.databaseUseCase.Execute(r.Context(), req.SourceIdentifier)
		if err != nil {
			log.Printf("Database migration failed: %v", err)
		} else {
			log.Printf("Database migration completed with job ID: %d", jobID)
		}
	}()

	respondJSON(w, map[string]string{"message": "Database migration started"}, http.StatusAccepted)
}

// StartGoogleSheetsMigration handles POST /migration/google-sheets
func (h *Handler) StartGoogleSheetsMigration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req StartGoogleSheetsMigrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SpreadsheetID == "" {
		respondError(w, "spreadsheet_id is required", http.StatusBadRequest)
		return
	}

	// Start migration in background
	go func() {
		jobID, err := h.googleSheetsUseCase.Execute(r.Context(), req.SpreadsheetID)
		if err != nil {
			log.Printf("Google Sheets migration failed: %v", err)
		} else {
			log.Printf("Google Sheets migration completed with job ID: %d", jobID)
		}
	}()

	respondJSON(w, map[string]string{"message": "Google Sheets migration started"}, http.StatusAccepted)
}

// StartExcelMigration handles POST /migration/excel
func (h *Handler) StartExcelMigration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		respondError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get the file from the form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondError(w, "Failed to get file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".xlsx") {
		respondError(w, "Only .xlsx files are supported", http.StatusBadRequest)
		return
	}

	// Create temporary directory if it doesn't exist
	tmpDir := "/tmp/migration-uploads"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		respondError(w, "Failed to create upload directory", http.StatusInternalServerError)
		return
	}

	// Save file to temporary location
	filePath := filepath.Join(tmpDir, header.Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		respondError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		respondError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Start migration in background
	go func() {
		jobID, err := h.excelUseCase.Execute(r.Context(), filePath)
		if err != nil {
			log.Printf("Excel migration failed: %v", err)
		} else {
			log.Printf("Excel migration completed with job ID: %d", jobID)
		}

		// Clean up file after migration
		os.Remove(filePath)
	}()

	respondJSON(w, map[string]string{"message": "Excel migration started"}, http.StatusAccepted)
}

// GetMigrationJob handles GET /migration/jobs/{id}
func (h *Handler) GetMigrationJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract job ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 3 {
		respondError(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	jobID, err := strconv.Atoi(pathParts[2])
	if err != nil {
		respondError(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	// Get job from repository
	job, err := h.jobRepo.GetByID(r.Context(), jobID)
	if err != nil {
		if err == domain.ErrMigrationJobNotFound {
			respondError(w, "Migration job not found", http.StatusNotFound)
			return
		}
		respondError(w, "Failed to get migration job", http.StatusInternalServerError)
		return
	}

	// Convert to response
	response := jobToResponse(job)
	respondJSON(w, response, http.StatusOK)
}

// ListMigrationJobs handles GET /migration/jobs
func (h *Handler) ListMigrationJobs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get all jobs from repository
	jobs, err := h.jobRepo.List(r.Context())
	if err != nil {
		respondError(w, "Failed to list migration jobs", http.StatusInternalServerError)
		return
	}

	// Convert to response
	var responses []MigrationJobResponse
	for _, job := range jobs {
		responses = append(responses, jobToResponse(job))
	}

	respondJSON(w, responses, http.StatusOK)
}

// jobToResponse converts a domain.MigrationJob to MigrationJobResponse
func jobToResponse(job *domain.MigrationJob) MigrationJobResponse {
	response := MigrationJobResponse{
		ID:               job.ID,
		SourceType:       job.SourceType,
		SourceIdentifier: job.SourceIdentifier,
		Status:           job.Status,
		Total:            job.Total,
		Processed:        job.Processed,
		Failed:           job.Failed,
		StartedAt:        job.StartedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if job.CompletedAt != nil {
		completedAt := job.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		response.CompletedAt = &completedAt
	}

	return response
}

// respondJSON sends a JSON response
func respondJSON(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// respondError sends an error response
func respondError(w http.ResponseWriter, message string, statusCode int) {
	respondJSON(w, ErrorResponse{Error: message}, statusCode)
}
