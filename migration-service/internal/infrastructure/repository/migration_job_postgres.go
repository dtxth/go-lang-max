package repository

import (
	"context"
	"database/sql"
	"migration-service/internal/domain"
	"time"
)

// MigrationJobPostgresRepository implements MigrationJobRepository using PostgreSQL
type MigrationJobPostgresRepository struct {
	db *sql.DB
}

// NewMigrationJobPostgresRepository creates a new MigrationJobPostgresRepository
func NewMigrationJobPostgresRepository(db *sql.DB) *MigrationJobPostgresRepository {
	return &MigrationJobPostgresRepository{db: db}
}

// Create creates a new migration job
func (r *MigrationJobPostgresRepository) Create(ctx context.Context, job *domain.MigrationJob) error {
	query := `
		INSERT INTO migration_jobs (source_type, source_identifier, status, total, processed, failed, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, started_at
	`

	err := r.db.QueryRowContext(
		ctx,
		query,
		job.SourceType,
		job.SourceIdentifier,
		job.Status,
		job.Total,
		job.Processed,
		job.Failed,
		time.Now(),
	).Scan(&job.ID, &job.StartedAt)

	return err
}

// GetByID retrieves a migration job by ID
func (r *MigrationJobPostgresRepository) GetByID(ctx context.Context, id int) (*domain.MigrationJob, error) {
	query := `
		SELECT id, source_type, source_identifier, status, total, processed, failed, started_at, completed_at
		FROM migration_jobs
		WHERE id = $1
	`

	job := &domain.MigrationJob{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID,
		&job.SourceType,
		&job.SourceIdentifier,
		&job.Status,
		&job.Total,
		&job.Processed,
		&job.Failed,
		&job.StartedAt,
		&job.CompletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrMigrationJobNotFound
	}

	return job, err
}

// List retrieves all migration jobs
func (r *MigrationJobPostgresRepository) List(ctx context.Context) ([]*domain.MigrationJob, error) {
	query := `
		SELECT id, source_type, source_identifier, status, total, processed, failed, started_at, completed_at
		FROM migration_jobs
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*domain.MigrationJob
	for rows.Next() {
		job := &domain.MigrationJob{}
		err := rows.Scan(
			&job.ID,
			&job.SourceType,
			&job.SourceIdentifier,
			&job.Status,
			&job.Total,
			&job.Processed,
			&job.Failed,
			&job.StartedAt,
			&job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// Update updates a migration job
func (r *MigrationJobPostgresRepository) Update(ctx context.Context, job *domain.MigrationJob) error {
	query := `
		UPDATE migration_jobs
		SET source_type = $1, source_identifier = $2, status = $3, total = $4, 
		    processed = $5, failed = $6, completed_at = $7
		WHERE id = $8
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		job.SourceType,
		job.SourceIdentifier,
		job.Status,
		job.Total,
		job.Processed,
		job.Failed,
		job.CompletedAt,
		job.ID,
	)

	return err
}

// UpdateProgress updates the progress of a migration job
func (r *MigrationJobPostgresRepository) UpdateProgress(ctx context.Context, id int, processed, failed int) error {
	query := `
		UPDATE migration_jobs
		SET processed = $1, failed = $2
		WHERE id = $3
	`

	_, err := r.db.ExecContext(ctx, query, processed, failed, id)
	return err
}

// UpdateStatus updates the status of a migration job
func (r *MigrationJobPostgresRepository) UpdateStatus(ctx context.Context, id int, status string) error {
	query := `
		UPDATE migration_jobs
		SET status = $1, completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN now() ELSE completed_at END
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}
