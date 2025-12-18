-- Add profile source tracking columns to employees table
ALTER TABLE employees 
  ADD COLUMN profile_source TEXT DEFAULT 'default',
  ADD COLUMN profile_last_updated TIMESTAMP WITH TIME ZONE;

-- Create index for profile_source
CREATE INDEX IF NOT EXISTS idx_employees_profile_source ON employees(profile_source);

-- Add comments for documentation
COMMENT ON COLUMN employees.profile_source IS 'Source of profile information: webhook, user_input, or default';
COMMENT ON COLUMN employees.profile_last_updated IS 'Timestamp when profile information was last updated';

-- Update existing records to have default profile source
UPDATE employees SET profile_source = 'default' WHERE profile_source IS NULL;