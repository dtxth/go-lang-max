-- Remove MAX-specific columns from users table
DROP INDEX IF EXISTS idx_users_max_id;
ALTER TABLE users DROP COLUMN IF EXISTS max_id;
ALTER TABLE users DROP COLUMN IF EXISTS username;
ALTER TABLE users DROP COLUMN IF EXISTS name;