-- Remove profile source tracking columns from employees table
ALTER TABLE employees 
  DROP COLUMN IF EXISTS profile_source,
  DROP COLUMN IF EXISTS profile_last_updated;