-- Rollback for 003_add_batch_update_jobs.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_batch_update_jobs_started_at;
DROP INDEX IF EXISTS idx_batch_update_jobs_status;

-- Drop batch_update_jobs table
DROP TABLE IF EXISTS batch_update_jobs;
