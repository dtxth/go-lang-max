package domain

import "context"

// MigrationErrorRepository defines the interface for migration error persistence
type MigrationErrorRepository interface {
	// Create creates a new migration error
	Create(ctx context.Context, err *MigrationError) error

	// ListByJobID retrieves all errors for a specific job
	ListByJobID(ctx context.Context, jobID int) ([]*MigrationError, error)
}
