-- Удаляем роль superadmin для пользователя
DELETE FROM user_roles 
WHERE user_id IN (
  SELECT id FROM users WHERE phone = '+79999999999'
);

-- Удаляем пользователя superadmin
DELETE FROM users WHERE phone = '+79999999999';