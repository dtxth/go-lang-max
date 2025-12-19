# MAX Mini App Authentication API

## Endpoint: POST /auth/max

### Описание
Этот endpoint предназначен для аутентификации пользователей MAX Mini App. Он валидирует `initData`, полученные от MAX платформы, и возвращает JWT токены для дальнейшего использования в приложении.

### URL
```
POST /auth/max
```

### Headers
```
Content-Type: application/json
```

### Request Body
```json
{
  "init_data": "max_id=123456789&first_name=John&last_name=Doe&username=johndoe&hash=a1b2c3d4e5f6..."
}
```

#### Параметры:
- `init_data` (string, required): Строка с данными инициализации от MAX платформы, включающая информацию о пользователе и криптографическую подпись

### Response

#### Успешный ответ (200 OK)
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

#### Ошибки

**400 Bad Request** - Неверный формат запроса
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Missing required field: init_data"
}
```

**401 Unauthorized** - Ошибка аутентификации
```json
{
  "error": "UNAUTHORIZED", 
  "message": "Invalid authentication data"
}
```

**500 Internal Server Error** - Внутренняя ошибка сервера
```json
{
  "error": "INTERNAL_ERROR",
  "message": "Internal server error"
}
```

### Пример использования

#### cURL
```bash
curl -X POST http://localhost:8080/auth/max \
  -H "Content-Type: application/json" \
  -d '{
    "init_data": "max_id=123456789&first_name=John&last_name=Doe&username=johndoe&hash=a1b2c3d4e5f6..."
  }'
```

#### JavaScript (fetch)
```javascript
const response = await fetch('/auth/max', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    init_data: 'max_id=123456789&first_name=John&last_name=Doe&username=johndoe&hash=a1b2c3d4e5f6...'
  })
});

const data = await response.json();
console.log('Access token:', data.access_token);
console.log('Refresh token:', data.refresh_token);
```

### Безопасность

1. **Криптографическая проверка**: Все `initData` проверяются с использованием HMAC-SHA256 подписи
2. **Валидация данных**: Проверяется формат и целостность всех параметров
3. **Защита от подделки**: Невозможно создать валидные `initData` без знания секретного ключа бота

### Интеграция с MAX Mini App

Для получения `initData` в MAX Mini App используйте:

```javascript
// В MAX Mini App
const initData = window.Telegram?.WebApp?.initData || '';

// Отправка на сервер аутентификации
const authResponse = await fetch('/auth/max', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ init_data: initData })
});
```

### Использование полученных токенов

После успешной аутентификации используйте `access_token` для авторизованных запросов:

```javascript
const response = await fetch('/api/protected-endpoint', {
  headers: {
    'Authorization': `Bearer ${access_token}`
  }
});
```

Для обновления токенов используйте `refresh_token` с endpoint `/refresh`.