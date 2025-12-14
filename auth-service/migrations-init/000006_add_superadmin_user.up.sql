-- Добавляем супер админа с паролем tEst123!
-- Хеш пароля сгенерирован с помощью bcrypt.DefaultCost (10)
INSERT INTO users (phone, email, password_hash, role, created_at) VALUES
  ('+79999999999', 'superadmin@digitaluniversity.ru', '$2a$10$faLCoMlD6CjDbsdfmEky.etk8dijbIv.FsXH.iHRzuSKSPZ4Lmj5W', 'superadmin', now())
ON CONFLICT (phone) DO NOTHING;

-- Получаем ID созданного пользователя и назначаем роль superadmin
INSERT INTO user_roles (user_id, role_id, university_id, branch_id, faculty_id, assigned_at)
SELECT 
  u.id,
  r.id,
  NULL, -- superadmin имеет доступ ко всем университетам
  NULL,
  NULL,
  now()
FROM users u
CROSS JOIN roles r
WHERE u.phone = '+79999999999' 
  AND r.name = 'superadmin'
ON CONFLICT (user_id, role_id, university_id, branch_id, faculty_id) DO NOTHING;