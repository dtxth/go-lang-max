-- Делаем INN опциональным в таблице universities
ALTER TABLE universities ALTER COLUMN inn DROP NOT NULL;

-- Добавляем уникальный индекс для случаев когда INN не NULL
DROP INDEX IF EXISTS idx_universities_inn;
CREATE UNIQUE INDEX idx_universities_inn_unique ON universities(inn) WHERE inn IS NOT NULL;

-- Удаляем старое ограничение уникальности и создаем новое
ALTER TABLE universities DROP CONSTRAINT IF EXISTS universities_inn_kpp_key;
CREATE UNIQUE INDEX idx_universities_inn_kpp_unique ON universities(inn, kpp) WHERE inn IS NOT NULL;