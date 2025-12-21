-- Удаляем тестового пользователя для MAX аутентификации
DELETE FROM users WHERE phone = '+79999999998' AND max_id = 79999999999;

-- Сбрасываем MAX поля для superadmin пользователя +79999999999
UPDATE users SET max_id = NULL, username = NULL, name = NULL WHERE phone = '+79999999999';