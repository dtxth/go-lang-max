-- Add MAX-specific columns to users table
ALTER TABLE users ADD COLUMN max_id BIGINT UNIQUE;
ALTER TABLE users ADD COLUMN username VARCHAR(255);
ALTER TABLE users ADD COLUMN name VARCHAR(255);

-- Create index on max_id for efficient lookups
CREATE INDEX idx_users_max_id ON users(max_id) WHERE max_id IS NOT NULL;