# Быстрый старт: API администраторов

## Новые endpoints

В chat-service добавлены два новых метода для работы с администраторами:

1. **GET /administrators/{admin_id}** - Получить администратора по ID
2. **GET /administrators** - Получить всех администраторов с пагинацией и поиском

## Примеры использования

### 1. Получить администратора по ID

```bash
curl -X GET "http://localhost:8082/administrators/1"
```

**Ответ:**
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

### 2. Получить всех администраторов (с пагинацией)

```bash
# Первая страница
curl -X GET "http://localhost:8082/administrators?limit=10&offset=0"

# Вторая страница
curl -X GET "http://localhost:8082/administrators?limit=10&offset=10"
```

**Ответ:**
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
    }
  ],
  "total_count": 25,
  "limit": 10,
  "offset": 0
}
```

### 3. Поиск администраторов

#### По телефону
```bash
curl -X GET "http://localhost:8082/administrators?query=%2B79991234567"
```

#### По MAX ID
```bash
curl -X GET "http://localhost:8082/administrators?query=496728250"
```

#### По названию чата
```bash
curl -X GET "http://localhost:8082/administrators?query=Test%20Chat"
```

## Параметры пагинации

- `limit` - количество записей на странице (по умолчанию 50, максимум 100)
- `offset` - смещение (по умолчанию 0)
- `query` - поисковый запрос (опционально)

## Особенности поиска

Поиск выполняется по трем полям:
- Телефон администратора
- MAX ID администратора
- Название чата

Поиск регистронезависимый и поддерживает частичное совпадение.

## Swagger UI

Полная документация доступна по адресу:
```
http://localhost:8082/swagger/index.html
```

## Тестирование

Запустить тесты:
```bash
cd chat-service
go test ./internal/usecase -run "TestGetAdministratorByID|TestGetAllAdministrators" -v
```

## Связанные endpoints

- **POST /chats/{chat_id}/administrators** - Добавить администратора
- **DELETE /administrators/{admin_id}** - Удалить администратора

Подробная документация: [ADMINISTRATORS_API.md](./ADMINISTRATORS_API.md)
