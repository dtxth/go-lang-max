package repository

import (
	"context"
	"database/sql"
	"migration-service/internal/domain"
	"time"
)

// MigrationErrorPostgresRepository implements MigrationErrorRepository using PostgreSQL
type MigrationErrorPostgresRepository struct {
	db *sql.DB
}

// NewMigrationErrorPostgresRepository creates a new MigrationErrorPostgresRepository
func NewMigrationErrorPostgresRepository(db *sql.DB) *MigrationErrorPostgresRepository {
	return &MigrationErrorPostgresRepository{db: db}
}

// Create creates a new migration error
func (r *MigrationErrorPostgresRepository) Create(ctx context.Context, migrationErr *domain.MigrationError) error {
	query := `
		INSERT INTO migration_errors (job_id, record_identifier, error_message, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		migrationErr.JobID,
		migrationErr.RecordIdentifier,
		migrationErr.ErrorMessage,
		time.Now(),
	).Scan(&migrationErr.ID, &migrationErr.CreatedAt)

	return err
}

// ListByJobID retrieves all errors for a specific job
func (r *MigrationErrorPostgresRepository) ListByJobID(ctx context.Context, jobID int) ([]*domain.MigrationError, error) {
	query := `
		SELECT id, job_id, record_identifier, error_message, created_at
		FROM migration_errors
		WHERE job_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var errors []*domain.MigrationError
	for rows.Next() {
		migrationErr := &domain.MigrationError{}
		err := rows.Scan(
			&migrationErr.ID,
			&migrationErr.JobID,
			&migrationErr.RecordIdentifier,
			&migrationErr.ErrorMessage,
			&migrationErr.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		errors = append(errors, migrationErr)
	}

	return errors, rows.Err()
}
