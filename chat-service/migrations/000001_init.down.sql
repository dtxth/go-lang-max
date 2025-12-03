-- Rollback for 001_init.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_administrators_updated_at ON administrators;
DROP TRIGGER IF EXISTS update_chats_updated_at ON chats;
DROP TRIGGER IF EXISTS update_universities_updated_at ON universities;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_administrators_max_id;
DROP INDEX IF EXISTS idx_administrators_phone;
DROP INDEX IF EXISTS idx_administrators_chat_id;
DROP INDEX IF EXISTS idx_chats_name_search;
DROP INDEX IF EXISTS idx_chats_source;
DROP INDEX IF EXISTS idx_chats_university_id;
DROP INDEX IF EXISTS idx_chats_max_chat_id;
DROP INDEX IF EXISTS idx_chats_name;
DROP INDEX IF EXISTS idx_universities_name;
DROP INDEX IF EXISTS idx_universities_inn;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS administrators;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS universities;
