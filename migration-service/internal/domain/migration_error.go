package domain

import "time"

// MigrationError represents an error that occurred during migration
type MigrationError struct {
	ID               int
	JobID            int
	RecordIdentifier string
	ErrorMessage     string
	CreatedAt        time.Time
}
