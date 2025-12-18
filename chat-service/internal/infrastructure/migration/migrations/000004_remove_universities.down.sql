-- Восстанавливаем таблицу universities
CREATE TABLE IF NOT EXISTS universities (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  inn TEXT NOT NULL,
  kpp TEXT,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
  UNIQUE(inn, kpp)
);

-- Восстанавливаем индексы для universities
CREATE INDEX IF NOT EXISTS idx_universities_inn ON universities(inn);
CREATE INDEX IF NOT EXISTS idx_universities_name ON universities(name);

-- Восстанавливаем индекс на university_id в chats
CREATE INDEX IF NOT EXISTS idx_chats_university_id ON chats(university_id);

-- Восстанавливаем внешний ключ (осторожно - может не работать если есть несуществующие ID)
-- ALTER TABLE chats ADD CONSTRAINT chats_university_id_fkey FOREIGN KEY (university_id) REFERENCES universities(id) ON DELETE SET NULL;

-- Восстанавливаем триггер для universities
CREATE TRIGGER update_universities_updated_at BEFORE UPDATE ON universities
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Убираем комментарий
COMMENT ON COLUMN chats.university_id IS NULL;