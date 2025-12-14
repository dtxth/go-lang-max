package domain

import "time"

// BatchUpdateJob represents a batch update operation
type BatchUpdateJob struct {
	ID          int64      `json:"id"`
	JobType     string     `json:"job_type"`     // 'max_id_update'
	Status      string     `json:"status"`       // 'running', 'completed', 'failed'
	Total       int        `json:"total"`        // Total records to process
	Processed   int        `json:"processed"`    // Successfully processed records
	Failed      int        `json:"failed"`       // Failed records
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// BatchUpdateResult represents the result of a batch update operation
type BatchUpdateResult struct {
	JobID     int64  `json:"job_id"`
	Total     int    `json:"total"`
	Success   int    `json:"success"`
	Failed    int    `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

// BatchUpdateJobRepository defines the interface for batch update job operations
type BatchUpdateJobRepository interface {
	// Create creates a new batch update job
	Create(job *BatchUpdateJob) error
	
	// GetByID retrieves a batch update job by ID
	GetByID(id int64) (*BatchUpdateJob, error)
	
	// Update updates a batch update job
	Update(job *BatchUpdateJob) error
	
	// GetAll retrieves all batch update jobs with pagination
	GetAll(limit, offset int) ([]*BatchUpdateJob, error)
}
