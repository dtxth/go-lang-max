-- Rollback for 002_add_role_to_users.sql

-- Drop index
DROP INDEX IF EXISTS idx_users_role;

-- Remove role column from users table
ALTER TABLE users DROP COLUMN IF EXISTS role;
