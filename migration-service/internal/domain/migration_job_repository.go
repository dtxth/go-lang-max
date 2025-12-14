package domain

import "context"

// MigrationJobRepository defines the interface for migration job persistence
type MigrationJobRepository interface {
	// Create creates a new migration job
	Create(ctx context.Context, job *MigrationJob) error

	// GetByID retrieves a migration job by ID
	GetByID(ctx context.Context, id int) (*MigrationJob, error)

	// List retrieves all migration jobs
	List(ctx context.Context) ([]*MigrationJob, error)

	// Update updates a migration job
	Update(ctx context.Context, job *MigrationJob) error

	// UpdateProgress updates the progress of a migration job
	UpdateProgress(ctx context.Context, id int, processed, failed int) error

	// UpdateStatus updates the status of a migration job
	UpdateStatus(ctx context.Context, id int, status string) error
}
