package repository

import (
	"database/sql"
	"employee-service/internal/domain"
)

type BatchUpdateJobPostgres struct {
	db *sql.DB
}

func NewBatchUpdateJobPostgres(db *sql.DB) *BatchUpdateJobPostgres {
	return &BatchUpdateJobPostgres{db: db}
}

func (r *BatchUpdateJobPostgres) Create(job *domain.BatchUpdateJob) error {
	err := r.db.QueryRow(
		`INSERT INTO batch_update_jobs (job_type, status, total, processed, failed) 
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, started_at`,
		job.JobType, job.Status, job.Total, job.Processed, job.Failed,
	).Scan(&job.ID, &job.StartedAt)
	return err
}

func (r *BatchUpdateJobPostgres) GetByID(id int64) (*domain.BatchUpdateJob, error) {
	job := &domain.BatchUpdateJob{}
	
	err := r.db.QueryRow(
		`SELECT id, job_type, status, total, processed, failed, started_at, completed_at
		 FROM batch_update_jobs
		 WHERE id = $1`,
		id,
	).Scan(
		&job.ID, &job.JobType, &job.Status, &job.Total, &job.Processed, 
		&job.Failed, &job.StartedAt, &job.CompletedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return job, nil
}

func (r *BatchUpdateJobPostgres) Update(job *domain.BatchUpdateJob) error {
	_, err := r.db.Exec(
		`UPDATE batch_update_jobs 
		 SET status = $1, total = $2, processed = $3, failed = $4, completed_at = $5
		 WHERE id = $6`,
		job.Status, job.Total, job.Processed, job.Failed, job.CompletedAt, job.ID,
	)
	return err
}

func (r *BatchUpdateJobPostgres) GetAll(limit, offset int) ([]*domain.BatchUpdateJob, error) {
	rows, err := r.db.Query(
		`SELECT id, job_type, status, total, processed, failed, started_at, completed_at
		 FROM batch_update_jobs
		 ORDER BY started_at DESC
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var jobs []*domain.BatchUpdateJob
	for rows.Next() {
		job := &domain.BatchUpdateJob{}
		
		err := rows.Scan(
			&job.ID, &job.JobType, &job.Status, &job.Total, &job.Processed,
			&job.Failed, &job.StartedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		
		jobs = append(jobs, job)
	}
	
	return jobs, rows.Err()
}
