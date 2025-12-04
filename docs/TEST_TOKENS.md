# Тестовые JWT токены для API

Эти токены сгенерированы для тестирования API endpoints, требующих аутентификации.

**Срок действия:** 24 часа с момента генерации  
**Дата генерации:** 04.12.2024

---

## 1. Superadmin Token

**Роль:** `superadmin`  
**Доступ:** Все чаты всех университетов  
**User ID:** 1  
**Email:** superadmin@example.com

### Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJzdXBlcmFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyYWRtaW4iLCJleHAiOjE3NjQ5NDk4MDAsImlhdCI6MTc2NDg2MzQwMH0.8BxeMMZV_II5Ffa-JsSpLCdJDL46sts0PkqHMn-ZNZQ
```

### Примеры использования:

```bash
# Получить все чаты
curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJzdXBlcmFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyYWRtaW4iLCJleHAiOjE3NjQ5NDk4MDAsImlhdCI6MTc2NDg2MzQwMH0.8BxeMMZV_II5Ffa-JsSpLCdJDL46sts0PkqHMn-ZNZQ' \
  'http://localhost:8082/chats/all?limit=10&offset=0'

# Поиск чатов
curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJzdXBlcmFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyYWRtaW4iLCJleHAiOjE3NjQ5NDk4MDAsImlhdCI6MTc2NDg2MzQwMH0.8BxeMMZV_II5Ffa-JsSpLCdJDL46sts0PkqHMn-ZNZQ' \
  'http://localhost:8082/chats?query=курс&limit=10'

# Получить чат по ID
curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJzdXBlcmFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyYWRtaW4iLCJleHAiOjE3NjQ5NDk4MDAsImlhdCI6MTc2NDg2MzQwMH0.8BxeMMZV_II5Ffa-JsSpLCdJDL46sts0PkqHMn-ZNZQ' \
  'http://localhost:8082/chats/1'
```

---

## 2. Curator Token (University 1)

**Роль:** `curator`  
**Доступ:** Только чаты университета с ID = 1  
**User ID:** 2  
**Email:** curator@example.com  
**University ID:** 1

### Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyIiwiZW1haWwiOiJjdXJhdG9yQGV4YW1wbGUuY29tIiwicm9sZSI6ImN1cmF0b3IiLCJleHAiOjE3NjQ5NDk4NTMsImlhdCI6MTc2NDg2MzQ1MywidW5pdmVyc2l0eV9pZCI6MX0.yKNjyVHGsOYEw0h3GyDox3406lKZFHvGqvm-qhIHi9M
```

### Примеры использования:

```bash
# Получить чаты своего университета
curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIyIiwiZW1haWwiOiJjdXJhdG9yQGV4YW1wbGUuY29tIiwicm9sZSI6ImN1cmF0b3IiLCJleHAiOjE3NjQ5NDk4NTMsImlhdCI6MTc2NDg2MzQ1MywidW5pdmVyc2l0eV9pZCI6MX0.yKNjyVHGsOYEw0h3GyDox3406lKZFHvGqvm-qhIHi9M' \
  'http://localhost:8082/chats/all?limit=10'
```

---

## 3. Operator Token (University 1)

**Роль:** `operator`  
**Доступ:** Только чаты университета с ID = 1  
**User ID:** 3  
**Email:** operator@example.com  
**University ID:** 1

### Token:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIzIiwiZW1haWwiOiJvcGVyYXRvckBleGFtcGxlLmNvbSIsInJvbGUiOiJvcGVyYXRvciIsImV4cCI6MTc2NDk0OTg1MywiaWF0IjoxNzY0ODYzNDUzLCJ1bml2ZXJzaXR5X2lkIjoxfQ.IpZTpqsFdvN_NlGunf5xjKoyH4OCYtdfTGNVUA5EXl8
```

### Примеры использования:

```bash
# Получить чаты своего университета
curl -H 'Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIzIiwiZW1haWwiOiJvcGVyYXRvckBleGFtcGxlLmNvbSIsInJvbGUiOiJvcGVyYXRvciIsImV4cCI6MTc2NDk0OTg1MywiaWF0IjoxNzY0ODYzNDUzLCJ1bml2ZXJzaXR5X2lkIjoxfQ.IpZTpqsFdvN_NlGunf5xjKoyH4OCYtdfTGNVUA5EXl8' \
  'http://localhost:8082/chats/all?limit=10'
```

---

## Генерация новых токенов

Если токены истекли, используйте скрипт `generate_token.py`:

```bash
python3 generate_token.py
```

Или создайте токен с кастомными параметрами:

```python
from generate_token import generate_token

# Superadmin
token, payload = generate_token(role="superadmin", user_id=1, email="admin@test.com")

# Curator для университета 5
token, payload = generate_token(role="curator", user_id=10, email="curator@test.com", university_id=5)

# Operator для университета 3
token, payload = generate_token(role="operator", user_id=20, email="operator@test.com", university_id=3)

print(token)
```

---

## Роли и права доступа

### Superadmin
- ✅ Видит все чаты всех университетов
- ✅ Может управлять всеми чатами
- ✅ Полный доступ ко всем операциям

### Curator
- ✅ Видит только чаты своего университета (university_id)
- ✅ Может управлять чатами своего университета
- ❌ Не видит чаты других университетов

### Operator
- ✅ Видит только чаты своего университета (university_id)
- ✅ Может управлять чатами своего университета
- ❌ Не видит чаты других университетов

---

## Проверка токена

Вы можете декодировать токен на [jwt.io](https://jwt.io) для просмотра содержимого.

Или использовать команду:

```bash
# Декодировать токен (требует jq)
echo "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxIiwiZW1haWwiOiJzdXBlcmFkbWluQGV4YW1wbGUuY29tIiwicm9sZSI6InN1cGVyYWRtaW4iLCJleHAiOjE3NjQ5NDk4MDAsImlhdCI6MTc2NDg2MzQwMH0.8BxeMMZV_II5Ffa-JsSpLCdJDL46sts0PkqHMn-ZNZQ" | \
  cut -d'.' -f2 | base64 -d 2>/dev/null | jq '.'
```

---

## Troubleshooting

### "invalid or expired token"
- Токен истек (срок действия 24 часа)
- Неправильный секрет (должен быть `super-secret-access`)
- Сгенерируйте новый токен

### "unauthorized"
- Токен не передан в заголовке
- Неправильный формат заголовка (должен быть `Authorization: Bearer <token>`)

### "forbidden"
- У пользователя нет прав на доступ к этому ресурсу
- Curator/Operator пытается получить доступ к чатам другого университета
