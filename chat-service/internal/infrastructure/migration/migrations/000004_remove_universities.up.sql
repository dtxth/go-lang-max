-- Удаляем внешний ключ на universities из таблицы chats
ALTER TABLE chats DROP CONSTRAINT IF EXISTS chats_university_id_fkey;

-- Удаляем индекс на university_id
DROP INDEX IF EXISTS idx_chats_university_id;

-- Удаляем триггер для universities
DROP TRIGGER IF EXISTS update_universities_updated_at ON universities;

-- Удаляем таблицу universities
DROP TABLE IF EXISTS universities;

-- Комментарий: university_id в таблице chats остается как обычное поле для ссылки на structure-service
COMMENT ON COLUMN chats.university_id IS 'ID университета из structure-service (без внешнего ключа)';