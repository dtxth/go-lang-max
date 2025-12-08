# API для работы с администраторами чатов

## Обзор

Этот документ описывает API endpoints для управления администраторами чатов в chat-service.

## Endpoints

### 1. Получить администратора по ID

**GET** `/administrators/{admin_id}`

Возвращает информацию об администраторе по его ID.

#### Параметры пути
- `admin_id` (int64, обязательный) - ID администратора

#### Пример запроса
```bash
curl -X GET "http://localhost:8082/administrators/1"
```

#### Пример ответа (200 OK)
```json
{
  "id": 1,
  "chat_id": 10,
  "phone": "+79991234567",
  "max_id": "496728250",
  "add_user": true,
  "add_admin": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Коды ответов
- `200 OK` - Администратор найден
- `400 Bad Request` - Неверный формат ID
- `404 Not Found` - Администратор не найден

---

### 2. Получить всех администраторов

**GET** `/administrators`

Возвращает список всех администраторов с поддержкой пагинации и поиска.

#### Query параметры
- `query` (string, опциональный) - Поисковый запрос (поиск по телефону, MAX ID или названию чата)
- `limit` (int, опциональный) - Лимит результатов (по умолчанию 50, максимум 100)
- `offset` (int, опциональный) - Смещение для пагинации (по умолчанию 0)

#### Пример запроса без фильтров
```bash
curl -X GET "http://localhost:8082/administrators?limit=10&offset=0"
```

#### Пример запроса с поиском
```bash
curl -X GET "http://localhost:8082/administrators?query=%2B79991234567&limit=10&offset=0"
```

#### Пример ответа (200 OK)
```json
{
  "administrators": [
    {
      "id": 1,
      "chat_id": 10,
      "phone": "+79991234567",
      "max_id": "496728250",
      "add_user": true,
      "add_admin": true,
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    },
    {
      "id": 2,
      "chat_id": 11,
      "phone": "+79997654321",
      "max_id": "496728251",
      "add_user": true,
      "add_admin": false,
      "created_at": "2024-01-15T11:00:00Z",
      "updated_at": "2024-01-15T11:00:00Z"
    }
  ],
  "total_count": 2,
  "limit": 10,
  "offset": 0
}
```

#### Коды ответов
- `200 OK` - Успешный запрос
- `400 Bad Request` - Неверные параметры запроса
- `500 Internal Server Error` - Внутренняя ошибка сервера

---

### 3. Добавить администратора к чату

**POST** `/chats/{chat_id}/administrators`

Добавляет нового администратора к чату.

#### Параметры пути
- `chat_id` (int64, обязательный) - ID чата

#### Тело запроса
```json
{
  "phone": "+79991234567",
  "max_id": "496728250",
  "add_user": true,
  "add_admin": true,
  "skip_phone_validation": false
}
```

#### Поля запроса
- `phone` (string, обязательный) - Номер телефона администратора
- `max_id` (string, опциональный) - MAX ID администратора (если не указан, будет получен автоматически)
- `add_user` (bool, опциональный) - Может ли администратор добавлять пользователей (по умолчанию true)
- `add_admin` (bool, опциональный) - Может ли администратор добавлять других администраторов (по умолчанию true)
- `skip_phone_validation` (bool, опциональный) - Пропустить валидацию телефона (для миграции, по умолчанию false)

#### Пример запроса
```bash
curl -X POST "http://localhost:8082/chats/10/administrators" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79991234567",
    "add_user": true,
    "add_admin": true
  }'
```

#### Пример ответа (201 Created)
```json
{
  "id": 1,
  "chat_id": 10,
  "phone": "+79991234567",
  "max_id": "496728250",
  "add_user": true,
  "add_admin": true,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

#### Коды ответов
- `201 Created` - Администратор успешно добавлен
- `400 Bad Request` - Неверные данные запроса или невалидный телефон
- `404 Not Found` - Чат не найден
- `409 Conflict` - Администратор с таким телефоном уже существует в этом чате

---

### 4. Удалить администратора

**DELETE** `/administrators/{admin_id}`

Удаляет администратора из чата. Нельзя удалить последнего администратора (должно быть минимум 2).

#### Параметры пути
- `admin_id` (int64, обязательный) - ID администратора

#### Пример запроса
```bash
curl -X DELETE "http://localhost:8082/administrators/1"
```

#### Пример ответа (200 OK)
```json
{
  "status": "deleted"
}
```

#### Коды ответов
- `200 OK` - Администратор успешно удален
- `400 Bad Request` - Неверный формат ID
- `404 Not Found` - Администратор не найден
- `409 Conflict` - Нельзя удалить последнего администратора (должно быть минимум 2)

---

## Примеры использования

### Получить список всех администраторов с пагинацией
```bash
# Первая страница (10 записей)
curl -X GET "http://localhost:8082/administrators?limit=10&offset=0"

# Вторая страница (10 записей)
curl -X GET "http://localhost:8082/administrators?limit=10&offset=10"
```

### Поиск администратора по телефону
```bash
curl -X GET "http://localhost:8082/administrators?query=%2B79991234567"
```

### Поиск администратора по MAX ID
```bash
curl -X GET "http://localhost:8082/administrators?query=496728250"
```

### Поиск администратора по названию чата
```bash
curl -X GET "http://localhost:8082/administrators?query=Test%20Chat"
```

### Получить конкретного администратора
```bash
curl -X GET "http://localhost:8082/administrators/1"
```

### Добавить администратора к чату
```bash
curl -X POST "http://localhost:8082/chats/10/administrators" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79991234567",
    "add_user": true,
    "add_admin": true
  }'
```

### Удалить администратора
```bash
curl -X DELETE "http://localhost:8082/administrators/1"
```

## Особенности

### Пагинация
- По умолчанию возвращается 50 записей
- Максимальный лимит - 100 записей
- Используйте параметры `limit` и `offset` для навигации по страницам

### Поиск
- Поиск выполняется по трем полям: телефон, MAX ID и название чата
- Поиск регистронезависимый (ILIKE)
- Поддерживается частичное совпадение

### Валидация
- Телефон должен быть в международном формате (например, +79991234567)
- При добавлении администратора проверяется существование чата
- Нельзя добавить дубликат администратора (одинаковый телефон в одном чате)
- Нельзя удалить последнего администратора чата

### Сортировка
- Администраторы сортируются по дате создания (от новых к старым)

## Ошибки

### 400 Bad Request
```json
{
  "error": "invalid administrator id"
}
```

### 404 Not Found
```json
{
  "error": "administrator not found"
}
```

### 409 Conflict
```json
{
  "error": "cannot delete last administrator, chat must have at least 2 administrators"
}
```

### 500 Internal Server Error
```json
{
  "error": "internal server error"
}
```
