-- Add role management columns to employees table
ALTER TABLE employees 
  ADD COLUMN role TEXT,
  ADD COLUMN user_id INTEGER,
  ADD COLUMN max_id_updated_at TIMESTAMP WITH TIME ZONE;

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_employees_role ON employees(role);
CREATE INDEX IF NOT EXISTS idx_employees_user_id ON employees(user_id);

-- Add comments for documentation
COMMENT ON COLUMN employees.role IS 'Employee role: curator, operator, or NULL for regular employee';
COMMENT ON COLUMN employees.user_id IS 'Reference to auth-service user ID';
COMMENT ON COLUMN employees.max_id_updated_at IS 'Timestamp when MAX_id was last updated';
