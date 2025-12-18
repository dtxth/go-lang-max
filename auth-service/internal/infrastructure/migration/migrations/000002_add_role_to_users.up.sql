-- Добавляем колонку role в таблицу users
ALTER TABLE users ADD COLUMN IF NOT EXISTS role TEXT NOT NULL DEFAULT 'operator';

-- Создаём индекс для быстрого поиска по роли
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);

-- Комментарий к колонке
COMMENT ON COLUMN users.role IS 'Роль пользователя: super_admin, curator, operator';

