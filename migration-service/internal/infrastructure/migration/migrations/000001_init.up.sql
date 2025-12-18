-- Migration jobs table
CREATE TABLE migration_jobs (
  id SERIAL PRIMARY KEY,
  source_type TEXT NOT NULL, -- 'database', 'google_sheets', 'excel'
  source_identifier TEXT, -- file path or sheet ID
  status TEXT NOT NULL, -- 'pending', 'running', 'completed', 'failed'
  total INTEGER DEFAULT 0,
  processed INTEGER DEFAULT 0,
  failed INTEGER DEFAULT 0,
  started_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_migration_jobs_status ON migration_jobs(status);
CREATE INDEX idx_migration_jobs_source_type ON migration_jobs(source_type);

-- Migration errors table
CREATE TABLE migration_errors (
  id SERIAL PRIMARY KEY,
  job_id INTEGER NOT NULL REFERENCES migration_jobs(id) ON DELETE CASCADE,
  record_identifier TEXT NOT NULL,
  error_message TEXT NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_migration_errors_job_id ON migration_errors(job_id);
