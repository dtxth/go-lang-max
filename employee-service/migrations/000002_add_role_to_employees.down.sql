-- Rollback for 002_add_role_to_employees.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_employees_user_id;
DROP INDEX IF EXISTS idx_employees_role;

-- Remove columns from employees table
ALTER TABLE employees 
  DROP COLUMN IF EXISTS max_id_updated_at,
  DROP COLUMN IF EXISTS user_id,
  DROP COLUMN IF EXISTS role;
