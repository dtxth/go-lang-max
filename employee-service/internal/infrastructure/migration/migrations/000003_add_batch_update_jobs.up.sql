-- Create batch_update_jobs table for tracking batch operations
CREATE TABLE IF NOT EXISTS batch_update_jobs (
  id SERIAL PRIMARY KEY,
  job_type TEXT NOT NULL, -- 'max_id_update'
  status TEXT NOT NULL, -- 'running', 'completed', 'failed'
  total INTEGER DEFAULT 0,
  processed INTEGER DEFAULT 0,
  failed INTEGER DEFAULT 0,
  started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  completed_at TIMESTAMP WITH TIME ZONE
);

-- Create index for querying jobs by status
CREATE INDEX IF NOT EXISTS idx_batch_update_jobs_status ON batch_update_jobs(status);
CREATE INDEX IF NOT EXISTS idx_batch_update_jobs_started_at ON batch_update_jobs(started_at DESC);

-- Add comments for documentation
COMMENT ON TABLE batch_update_jobs IS 'Tracks batch update operations for employees';
COMMENT ON COLUMN batch_update_jobs.job_type IS 'Type of batch job: max_id_update';
COMMENT ON COLUMN batch_update_jobs.status IS 'Job status: running, completed, failed';
COMMENT ON COLUMN batch_update_jobs.total IS 'Total number of records to process';
COMMENT ON COLUMN batch_update_jobs.processed IS 'Number of successfully processed records';
COMMENT ON COLUMN batch_update_jobs.failed IS 'Number of failed records';
