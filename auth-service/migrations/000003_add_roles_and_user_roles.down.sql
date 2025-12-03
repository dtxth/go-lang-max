-- Rollback for 003_add_roles_and_user_roles.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_user_roles_university_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_user_roles_user_id;

-- Drop user_roles table
DROP TABLE IF EXISTS user_roles;

-- Drop roles table
DROP TABLE IF EXISTS roles;
