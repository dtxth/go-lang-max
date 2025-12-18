-- Rollback for 001_init.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_employees_updated_at ON employees;
DROP TRIGGER IF EXISTS update_universities_updated_at ON universities;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_employees_inn;
DROP INDEX IF EXISTS idx_employees_name;
DROP INDEX IF EXISTS idx_employees_university_id;
DROP INDEX IF EXISTS idx_employees_max_id;
DROP INDEX IF EXISTS idx_employees_phone;
DROP INDEX IF EXISTS idx_universities_name;
DROP INDEX IF EXISTS idx_universities_inn;

-- Drop tables
DROP TABLE IF EXISTS employees;
DROP TABLE IF EXISTS universities;
