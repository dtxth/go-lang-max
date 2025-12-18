-- Создание таблицы universities (вузы) - если не существует
CREATE TABLE IF NOT EXISTS universities (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  inn TEXT NOT NULL,
  kpp TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(inn, kpp)
);

-- Создание индексов для universities
CREATE INDEX IF NOT EXISTS idx_universities_inn ON universities(inn);
CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);

-- Создание таблицы chats (чаты)
CREATE TABLE IF NOT EXISTS chats (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  max_chat_id TEXT,
  participants_count INTEGER DEFAULT 0,
  university_id INTEGER REFERENCES universities(id) ON DELETE SET NULL,
  department TEXT,
  source TEXT NOT NULL CHECK (source IN ('admin_panel', 'bot_registrar', 'academic_group')),
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- Создание индексов для chats
CREATE INDEX IF NOT EXISTS idx_chats_name ON chats(name);
CREATE INDEX IF NOT EXISTS idx_chats_max_chat_id ON chats(max_chat_id);
CREATE INDEX IF NOT EXISTS idx_chats_university_id ON chats(university_id);
CREATE INDEX IF NOT EXISTS idx_chats_source ON chats(source);
CREATE INDEX IF NOT EXISTS idx_chats_name_search ON chats USING gin(to_tsvector('russian', name));

-- Создание таблицы administrators (администраторы чатов)
CREATE TABLE IF NOT EXISTS administrators (
  id SERIAL PRIMARY KEY,
  chat_id INTEGER NOT NULL REFERENCES chats(id) ON DELETE CASCADE,
  phone TEXT NOT NULL,
  max_id TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(chat_id, phone)
);

-- Создание индексов для administrators
CREATE INDEX IF NOT EXISTS idx_administrators_chat_id ON administrators(chat_id);
CREATE INDEX IF NOT EXISTS idx_administrators_phone ON administrators(phone);
CREATE INDEX IF NOT EXISTS idx_administrators_max_id ON administrators(max_id);

-- Триггер для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_universities_updated_at BEFORE UPDATE ON universities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_chats_updated_at BEFORE UPDATE ON chats
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_administrators_updated_at BEFORE UPDATE ON administrators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

