# MAX Test Users Setup

## Автоматическое создание тестовых пользователей

Миграция `000008_add_max_test_user.sql` автоматически создает тестовых пользователей для MAX аутентификации при запуске сервиса.

### Создаваемые пользователи:

1. **Operator пользователь**
   - Телефон: `+79999999998`
   - Email: `maxtest@digitaluniversity.ru`
   - Роль: `operator`
   - MAX ID: `79999999999`
   - Username: `user79999999999`
   - Name: `Test User`
   - Пароль: `tEst123!`

2. **Superadmin пользователь** (обновляется существующий)
   - Телефон: `+79999999999`
   - Роль: `superadmin`
   - MAX ID: `79999999997`
   - Username: `superadmin_max`
   - Name: `Super Admin`
   - Пароль: `tEst123!`

### Готовые команды для тестирования:

#### Operator пользователь (max_id=79999999999):
```bash
# Минимальная версия
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "max_id=79999999999&first_name=TestUser&hash=097f0f5c738a684d92eb7f798fd464ac61e17653ea69ac2bbc28f6c00af81a4f"
  }'

# Полная версия
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "max_id=79999999999&username=user79999999999&first_name=Test&last_name=User&hash=59de91763780ef7e6c8b72657750c2686872f6622f23194e9958e4fc87b97e13"
  }'
```

#### Superadmin пользователь (max_id=79999999997):
```bash
curl -X 'POST' \
  'http://localhost:8080/auth/max' \
  -H 'Content-Type: application/json' \
  -d '{
    "init_data": "max_id=79999999997&username=superadmin_max&first_name=Super&last_name=Admin&hash=1c74b1fd37d24cf4cf1f091cc7c78b9935fdbc5386ac4f1d3b74712a51b5cf0b"
  }'
```

### Ручное применение миграции (если автоматическая не сработала):

```sql
-- Добавляем тестового пользователя для MAX аутентификации
INSERT INTO users (phone, email, password_hash, role, max_id, username, name, created_at) VALUES
  ('+79999999998', 'maxtest@digitaluniversity.ru', '$2a$10$faLCoMlD6CjDbsdfmEky.etk8dijbIv.FsXH.iHRzuSKSPZ4Lmj5W', 'operator', 79999999999, 'user79999999999', 'Test User', now())
ON CONFLICT (phone) DO UPDATE SET
  max_id = EXCLUDED.max_id,
  username = EXCLUDED.username,
  name = EXCLUDED.name;

-- Обновляем существующего superadmin пользователя
UPDATE users SET 
  max_id = 79999999997,
  username = 'superadmin_max',
  name = 'Super Admin'
WHERE phone = '+79999999999' AND max_id IS NULL;

-- Добавляем запись в schema_migrations
INSERT INTO schema_migrations (version, dirty) VALUES (8, false) ON CONFLICT (version) DO NOTHING;
```

### Проверка создания пользователей:

```bash
docker exec auth-db psql -U postgres -d postgres -c "SELECT id, phone, max_id, username, name, role FROM users WHERE max_id IS NOT NULL ORDER BY max_id;"
```

### Примечания:

- Используется тестовый токен бота: `test_bot_token_12345`
- Хеши init_data сгенерированы с правильной HMAC-SHA256 подписью
- При успешной аутентификации система обновляет username и name пользователя
- Пароли зашифрованы с помощью bcrypt (cost=10)