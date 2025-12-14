-- Добавление полей для поддержки импорта из Excel

-- Добавляем external_chat_id для хранения chat_id из Excel (колонка 14)
ALTER TABLE chats ADD COLUMN IF NOT EXISTS external_chat_id TEXT;
CREATE INDEX IF NOT EXISTS idx_chats_external_chat_id ON chats(external_chat_id);

-- Добавляем флаги add_user и add_admin для администраторов (колонки 16-17)
ALTER TABLE administrators ADD COLUMN IF NOT EXISTS add_user BOOLEAN DEFAULT TRUE;
ALTER TABLE administrators ADD COLUMN IF NOT EXISTS add_admin BOOLEAN DEFAULT TRUE;

-- Комментарии для документации
COMMENT ON COLUMN chats.external_chat_id IS 'ID чата из внешней системы (например, из Excel импорта)';
COMMENT ON COLUMN administrators.add_user IS 'Флаг: может ли администратор добавлять пользователей';
COMMENT ON COLUMN administrators.add_admin IS 'Флаг: может ли администратор добавлять других администраторов';
