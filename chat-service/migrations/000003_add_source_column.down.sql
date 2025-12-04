-- Удаление индексов
DROP INDEX IF EXISTS idx_chats_source;
DROP INDEX IF EXISTS idx_administrators_max_id;

-- Удаление колонок
ALTER TABLE chats DROP COLUMN IF EXISTS source;
ALTER TABLE administrators DROP COLUMN IF EXISTS max_id;
