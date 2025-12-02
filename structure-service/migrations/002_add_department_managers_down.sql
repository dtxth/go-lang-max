-- Rollback for 002_add_department_managers.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_department_managers_faculty_id;
DROP INDEX IF EXISTS idx_department_managers_branch_id;
DROP INDEX IF EXISTS idx_department_managers_employee_id;

-- Drop department_managers table
DROP TABLE IF EXISTS department_managers;
