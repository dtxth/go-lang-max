# API Reference - Цифровой Вуз

Полная документация API всех микросервисов системы "Цифровой Вуз".

## Содержание

1. [Auth Service API](#auth-service-api)
2. [Employee Service API](#employee-service-api)
3. [Chat Service API](#chat-service-api)
4. [Structure Service API](#structure-service-api)
5. [Migration Service API](#migration-service-api)
6. [Общие концепции](#общие-концепции)

## Общие концепции

### Аутентификация

Все защищенные endpoints требуют JWT токен в заголовке:

```
Authorization: Bearer <access_token>
```

### Формат ответов

Успешные ответы:
```json
{
  "data": { ... },
  "meta": { ... }
}
```

Ошибки:
```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { ... }
  }
}
```

### Пагинация

Параметры:
- `limit` - количество записей (по умолчанию 50, максимум 100)
- `offset` - смещение (по умолчанию 0)

Ответ включает метаданные:
```json
{
  "data": [...],
  "meta": {
    "total": 1000,
    "limit": 50,
    "offset": 0
  }
}
```

### Коды ошибок

- `400` - Validation Error
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `409` - Conflict
- `500` - Internal Server Error
- `502` - Bad Gateway (external service error)

## Auth Service API

Base URL: `http://localhost:8080`

### POST /auth/register
Регистрация нового пользователя


**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!",
  "first_name": "Иван",
  "last_name": "Иванов"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "user_id": 123,
    "email": "user@example.com",
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

### POST /auth/login
Вход в систему

**Request:**
```json
{
  "email": "user@example.com",
  "password": "SecurePassword123!"
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "user_id": 123,
    "role": "curator",
    "university_id": 1,
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

### POST /auth/refresh
Обновление access токена

**Request:**
```json
{
  "refresh_token": "eyJhbGc..."
}
```

**Response:** `200 OK`
```json
{
  "data": {
    "access_token": "eyJhbGc...",
    "refresh_token": "eyJhbGc..."
  }
}
```

### POST /auth/password-reset/request
Запрос сброса пароля (отправка токена на телефон через MAX Messenger)

**Request:**
```json
{
  "phone": "+79991234567"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password reset token sent to your phone"
}
```

### POST /auth/password-reset/confirm
Подтверждение сброса пароля с токеном

**Request:**
```json
{
  "token": "abc123def456",
  "new_password": "NewSecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password successfully reset"
}
```

**Notes:**
- Токен действителен 15 минут
- Токен одноразовый
- После сброса все refresh токены аннулируются

### POST /auth/password/change
Изменение пароля (требует аутентификации)

**Request:**
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewSecurePass123!"
}
```

**Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password successfully changed"
}
```

**Notes:**
- Требуется Bearer токен в заголовке Authorization
- После изменения все refresh токены аннулируются
- Пользователь будет разлогинен со всех устройств

### POST /roles/assign
Назначение роли пользователю (требует Superadmin или Curator)

**Request:**
```json
{
  "user_id": 456,
  "role": "operator",
  "university_id": 1,
  "branch_id": 10
}
```

**Response:** `200 OK`

### GET /users/{id}/permissions
Получение прав доступа пользователя

**Response:** `200 OK`
```json
{
  "data": {
    "user_id": 123,
    "roles": [
      {
        "role": "curator",
        "university_id": 1,
        "university_name": "МГУ"
      }
    ]
  }
}
```

## Employee Service API

Base URL: `http://localhost:8081`

### POST /employees
Создание сотрудника с автоматическим получением MAX_id

**Request:**
```json
{
  "first_name": "Иван",
  "last_name": "Иванов",
  "phone": "+79991234567",
  "role": "operator",
  "university": {
    "name": "МГУ",
    "inn": "7714357576",
    "kpp": "771401001"
  }
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "id": 789,
    "first_name": "Иван",
    "last_name": "Иванов",
    "phone": "+79991234567",
    "max_id": "max_user_123",
    "role": "operator",
    "user_id": 456,
    "university": {
      "id": 1,
      "name": "МГУ",
      "inn": "7714357576"
    }
  }
}
```

### GET /employees
Поиск сотрудников с фильтрацией по ролям

**Query Parameters:**
- `q` - поисковый запрос (имя, фамилия, название вуза)
- `limit` - количество записей
- `offset` - смещение

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": 789,
      "first_name": "Иван",
      "last_name": "Иванов",
      "phone": "+79991234567",
      "role": "operator",
      "university_name": "МГУ"
    }
  ],
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0
  }
}
```

### POST /employees/batch-update-maxid
Пакетное обновление MAX_id для сотрудников

**Response:** `202 Accepted`
```json
{
  "data": {
    "job_id": 1,
    "status": "running",
    "total": 1000
  }
}
```

### GET /employees/batch-status
Статус пакетного обновления

**Response:** `200 OK`
```json
{
  "data": {
    "job_id": 1,
    "status": "completed",
    "total": 1000,
    "processed": 1000,
    "failed": 23,
    "started_at": "2024-01-15T10:00:00Z",
    "completed_at": "2024-01-15T10:15:00Z"
  }
}
```

## Chat Service API

Base URL: `http://localhost:8082`

### GET /chats
Список чатов с фильтрацией по ролям и пагинацией

**Query Parameters:**
- `limit` - количество записей (по умолчанию 50, максимум 100)
- `offset` - смещение

**Response:** `200 OK`
```json
{
  "data": [
    {
      "id": 1,
      "name": "Математика 1 курс",
      "url": "https://max.ru/chat/123",
      "source": "academic_group",
      "university_id": 1,
      "university_name": "МГУ",
      "administrators": [
        {
          "id": 1,
          "phone": "+79991234567",
          "max_id": "max_user_123"
        }
      ]
    }
  ],
  "meta": {
    "total": 5000,
    "limit": 50,
    "offset": 0
  }
}
```

### GET /chats/search
Поиск чатов по названию

**Query Parameters:**
- `q` - поисковый запрос (обязательный)
- `limit` - количество записей
- `offset` - смещение

**Response:** `200 OK`

### POST /chats/{id}/admins
Добавление администратора чата

**Request:**
```json
{
  "phone": "+79991234567"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "id": 10,
    "chat_id": 1,
    "phone": "+79991234567",
    "max_id": "max_user_123"
  }
}
```

### DELETE /chats/{id}/admins/{admin_id}
Удаление администратора (защита последнего администратора)

**Response:** `204 No Content` или `409 Conflict` если последний администратор

## Structure Service API

Base URL: `http://localhost:8083`

### POST /import/excel
Импорт структуры из Excel файла

**Request:** `multipart/form-data`
- `file` - Excel файл

**Response:** `202 Accepted`
```json
{
  "data": {
    "job_id": "import_123",
    "status": "processing",
    "message": "Import started"
  }
}
```

### GET /universities/{id}/structure
Получение полной иерархии университета

**Response:** `200 OK`
```json
{
  "data": {
    "id": 1,
    "name": "МГУ",
    "inn": "7714357576",
    "branches": [
      {
        "id": 10,
        "name": "Главный корпус",
        "faculties": [
          {
            "id": 100,
            "name": "Механико-математический факультет",
            "groups": [
              {
                "id": 1000,
                "name": "ММ-101",
                "course": 1,
                "chat": {
                  "id": 5000,
                  "name": "Математика 1 курс",
                  "url": "https://max.ru/chat/123"
                }
              }
            ]
          }
        ]
      }
    ]
  }
}
```

### POST /departments/managers
Назначение оператора на подразделение

**Request:**
```json
{
  "employee_id": 789,
  "branch_id": 10
}
```

**Response:** `201 Created`

## Migration Service API

Base URL: `http://localhost:8084`

### POST /migration/database
Миграция из базы данных

**Request:**
```json
{
  "source_db_url": "postgres://user:pass@host:5432/old_db"
}
```

**Response:** `202 Accepted`
```json
{
  "data": {
    "job_id": 1,
    "status": "running"
  }
}
```

### POST /migration/excel
Миграция из Excel файла

**Request:** `multipart/form-data`
- `file` - Excel файл с чатами

**Response:** `202 Accepted`

### GET /migration/jobs/{id}
Статус миграции

**Response:** `200 OK`
```json
{
  "data": {
    "id": 1,
    "source_type": "excel",
    "status": "running",
    "total": 155000,
    "processed": 50000,
    "failed": 123,
    "started_at": "2024-01-15T10:00:00Z",
    "estimated_completion": "2024-01-15T13:00:00Z"
  }
}
```

### GET /migration/jobs/{id}/errors
Ошибки миграции

**Response:** `200 OK`
```json
{
  "data": [
    {
      "record_identifier": "row_1234",
      "error_message": "Invalid INN format",
      "created_at": "2024-01-15T10:05:00Z"
    }
  ]
}
```

## Swagger Documentation

Интерактивная документация доступна по адресам:

- Auth Service: http://localhost:8080/swagger/index.html
- Employee Service: http://localhost:8081/swagger/index.html
- Chat Service: http://localhost:8082/swagger/index.html
- Structure Service: http://localhost:8083/swagger/index.html

Для получения дополнительной информации см. [README.md](./README.md)
