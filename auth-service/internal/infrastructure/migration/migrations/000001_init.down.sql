-- Rollback for 001_init.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

-- Drop tables
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
