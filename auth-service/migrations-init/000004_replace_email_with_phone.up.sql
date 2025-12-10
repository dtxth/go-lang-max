-- Добавляем колонку phone
ALTER TABLE users ADD COLUMN phone TEXT;

-- Копируем данные из email в phone (для существующих пользователей)
UPDATE users SET phone = email WHERE phone IS NULL;

-- Делаем phone обязательным и уникальным
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
CREATE UNIQUE INDEX idx_users_phone ON users(phone);

-- Делаем email опциональным (для обратной совместимости)
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;

-- Добавляем индекс на email для поиска (если нужно)
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email) WHERE email IS NOT NULL;
