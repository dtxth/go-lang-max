-- Удаление полей для Excel импорта из таблицы chats
ALTER TABLE chats DROP COLUMN IF EXISTS external_chat_id;
ALTER TABLE chats DROP COLUMN IF EXISTS source;

-- Удаление полей для Excel импорта из таблицы administrators
ALTER TABLE administrators DROP COLUMN IF EXISTS max_id;
ALTER TABLE administrators DROP COLUMN IF EXISTS can_add_users;
ALTER TABLE administrators DROP COLUMN IF EXISTS can_add_admins;

-- Удаление индексов
DROP INDEX IF EXISTS idx_chats_external_chat_id;
DROP INDEX IF EXISTS idx_chats_source;
DROP INDEX IF EXISTS idx_administrators_max_id;
