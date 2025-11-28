-- Rollback for 001_init.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_migration_errors_job_id;
DROP INDEX IF EXISTS idx_migration_jobs_source_type;
DROP INDEX IF EXISTS idx_migration_jobs_status;

-- Drop tables
DROP TABLE IF EXISTS migration_errors;
DROP TABLE IF EXISTS migration_jobs;
