-- Откатываем изменения: делаем INN обязательным
-- Сначала удаляем записи с NULL INN (если есть)
DELETE FROM universities WHERE inn IS NULL;

-- Восстанавливаем NOT NULL ограничение
ALTER TABLE universities ALTER COLUMN inn SET NOT NULL;

-- Восстанавливаем старые индексы
DROP INDEX IF EXISTS idx_universities_inn_unique;
DROP INDEX IF EXISTS idx_universities_inn_kpp_unique;

CREATE INDEX idx_universities_inn ON universities(inn);
ALTER TABLE universities ADD CONSTRAINT universities_inn_kpp_key UNIQUE(inn, kpp);