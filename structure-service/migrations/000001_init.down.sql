-- Rollback for 001_init.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_groups_updated_at ON groups;
DROP TRIGGER IF EXISTS update_faculties_updated_at ON faculties;
DROP TRIGGER IF EXISTS update_branches_updated_at ON branches;
DROP TRIGGER IF EXISTS update_universities_updated_at ON universities;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_groups_course;
DROP INDEX IF EXISTS idx_groups_chat_id;
DROP INDEX IF EXISTS idx_groups_faculty_id;
DROP INDEX IF EXISTS idx_faculties_name;
DROP INDEX IF EXISTS idx_faculties_branch_id;
DROP INDEX IF EXISTS idx_branches_name;
DROP INDEX IF EXISTS idx_branches_university_id;
DROP INDEX IF EXISTS idx_universities_name;
DROP INDEX IF EXISTS idx_universities_inn;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS faculties;
DROP TABLE IF EXISTS branches;
DROP TABLE IF EXISTS universities;
