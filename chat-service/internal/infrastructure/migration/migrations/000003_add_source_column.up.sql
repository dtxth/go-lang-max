-- Добавление колонки source для отслеживания источника чата
ALTER TABLE chats ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'admin_panel' CHECK (source IN ('admin_panel', 'bot_registrar', 'academic_group'));

-- Создание индекса для source
CREATE INDEX IF NOT EXISTS idx_chats_source ON chats(source);

-- Добавление колонки max_id для администраторов (если отсутствует)
ALTER TABLE administrators ADD COLUMN IF NOT EXISTS max_id TEXT;

-- Создание индекса для max_id
CREATE INDEX IF NOT EXISTS idx_administrators_max_id ON administrators(max_id);

-- Комментарии для документации
COMMENT ON COLUMN chats.source IS 'Источник создания чата: admin_panel, bot_registrar, или academic_group';
COMMENT ON COLUMN administrators.max_id IS 'ID пользователя в системе MAX';
