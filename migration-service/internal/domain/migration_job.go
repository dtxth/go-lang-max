package domain

import "time"

// MigrationJob represents a migration job
type MigrationJob struct {
	ID               int
	SourceType       string // 'database', 'google_sheets', 'excel'
	SourceIdentifier string // file path or sheet ID
	Status           string // 'pending', 'running', 'completed', 'failed'
	Total            int
	Processed        int
	Failed           int
	StartedAt        time.Time
	CompletedAt      *time.Time
}

// MigrationJobStatus constants
const (
	MigrationJobStatusPending   = "pending"
	MigrationJobStatusRunning   = "running"
	MigrationJobStatusCompleted = "completed"
	MigrationJobStatusFailed    = "failed"
)

// MigrationSourceType constants
const (
	MigrationSourceDatabase     = "database"
	MigrationSourceGoogleSheets = "google_sheets"
	MigrationSourceExcel        = "excel"
)
