-- Добавляем тестового пользователя для MAX аутентификации
-- Пароль: TestPassword123! (хеш сгенерирован с помощью bcrypt.DefaultCost)
INSERT INTO users (phone, email, password_hash, role, max_id, username, name, created_at) VALUES
  ('+79999999998', 'maxtest@digitaluniversity.ru', '$2a$10$faLCoMlD6CjDbsdfmEky.etk8dijbIv.FsXH.iHRzuSKSPZ4Lmj5W', 'operator', 79999999999, 'user79999999999', 'Test User', now())
ON CONFLICT (phone) DO UPDATE SET
  max_id = EXCLUDED.max_id,
  username = EXCLUDED.username,
  name = EXCLUDED.name;

-- Обновляем существующего superadmin пользователя (+79999999999), добавляя ему MAX поля для тестирования
UPDATE users SET 
  max_id = 79999999997,
  username = 'superadmin_max',
  name = 'Super Admin'
WHERE phone = '+79999999999' AND max_id IS NULL;