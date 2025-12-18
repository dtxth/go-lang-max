package repository

import (
	"context"
	"database/sql"
	"fmt"
	"migration-service/internal/domain"
	"time"
)

// MigrationJobPostgresRepository implements MigrationJobRepository using PostgreSQL
type MigrationJobPostgresRepository struct {
	db  *sql.DB
	dsn string
}

// NewMigrationJobPostgresRepository creates a new MigrationJobPostgresRepository
func NewMigrationJobPostgresRepository(db *sql.DB) *MigrationJobPostgresRepository {
	return &MigrationJobPostgresRepository{db: db}
}

// NewMigrationJobPostgresRepositoryWithDSN creates a new MigrationJobPostgresRepository with DSN for reconnection
func NewMigrationJobPostgresRepositoryWithDSN(db *sql.DB, dsn string) *MigrationJobPostgresRepository {
	return &MigrationJobPostgresRepository{db: db, dsn: dsn}
}

// getDB returns a working database connection, reconnecting if necessary
func (r *MigrationJobPostgresRepository) getDB(ctx context.Context) (*sql.DB, error) {
	// Try to ping the existing connection
	if r.db != nil {
		if err := r.db.PingContext(ctx); err == nil {
			return r.db, nil
		}
	}

	// If we have a DSN, try to reconnect
	if r.dsn != "" {
		db, err := sql.Open("postgres", r.dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to reconnect to database: %w", err)
		}
		
		// Configure connection pool
		db.SetMaxOpenConns(25)
		db.SetMaxIdleConns(5)
		db.SetConnMaxLifetime(0)
		
		if err := db.PingContext(ctx); err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to ping reconnected database: %w", err)
		}
		
		r.db = db
		return db, nil
	}

	return r.db, nil
}

// Create creates a new migration job
func (r *MigrationJobPostgresRepository) Create(ctx context.Context, job *domain.MigrationJob) error {
	db, err := r.getDB(ctx)
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		INSERT INTO migration_jobs (source_type, source_identifier, status, total, processed, failed, started_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, started_at
	`

	err = db.QueryRowContext(
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

	if err != nil {
		return fmt.Errorf("failed to insert migration job: %w", err)
	}

	return nil
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
	db, err := r.getDB(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	query := `
		SELECT id, source_type, source_identifier, status, total, processed, failed, started_at, completed_at
		FROM migration_jobs
		ORDER BY started_at DESC
	`

	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query migration jobs: %w", err)
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
			return nil, fmt.Errorf("failed to scan migration job: %w", err)
		}
		jobs = append(jobs, job)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating migration jobs: %w", err)
	}

	return jobs, nil
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
