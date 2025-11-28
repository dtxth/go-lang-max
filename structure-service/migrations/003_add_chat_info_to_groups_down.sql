-- Rollback for 003_add_chat_info_to_groups.sql

-- Drop index
DROP INDEX IF EXISTS idx_groups_chat_url;

-- Remove columns from groups table
ALTER TABLE groups 
  DROP COLUMN IF EXISTS chat_name,
  DROP COLUMN IF EXISTS chat_url;
