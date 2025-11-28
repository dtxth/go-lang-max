-- Add chat_url and chat_name columns to groups table
ALTER TABLE groups 
  ADD COLUMN IF NOT EXISTS chat_url TEXT,
  ADD COLUMN IF NOT EXISTS chat_name TEXT;

-- Create index for chat_url for faster lookups
CREATE INDEX IF NOT EXISTS idx_groups_chat_url ON groups(chat_url);
