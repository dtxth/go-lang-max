# Employee Service

Сервис управления сотрудниками вузов для мини-приложения.

## Функциональность

- **Поиск сотрудников** по имени, фамилии и названию вуза
- **Добавление сотрудников** по номеру телефона
- **Автоматическое получение MAX_id** по номеру телефона
- **Автоматическое получение профилей** из MAX Messenger через webhook интеграцию
- **Приоритизация источников имен** (пользователь > webhook > по умолчанию)
- **Отслеживание источников профилей** для аудита и качества данных
- **Управление вузами** (автоматическое создание при добавлении сотрудника)
- **CRUD операции** для сотрудников

## API Endpoints

### Сотрудники

- `GET /employees?query=...&limit=50&offset=0` - Поиск сотрудников
- `GET /employees/all?limit=50&offset=0` - Получить всех сотрудников
- `GET /employees/{id}` - Получить сотрудника по ID
- `POST /employees` - Добавить сотрудника (с автоматическим получением профиля)
- `PUT /employees/{id}` - Обновить сотрудника
- `DELETE /employees/{id}` - Удалить сотрудника

### Профили пользователей (NEW)

Сервис автоматически интегрируется с системой профилей MAX Messenger:

- **Автоматическое получение имен**: При создании сотрудника система автоматически получает имя и фамилию из кэша профилей
- **Приоритизация источников**: 
  1. Имена, предоставленные пользователем (наивысший приоритет)
  2. Имена из профиля MAX Messenger (webhook)
  3. Значения по умолчанию ("Неизвестно")
- **Отслеживание источников**: Каждый сотрудник имеет поле `profile_source` для отслеживания источника данных
- **Graceful degradation**: Система продолжает работать даже при недоступности кэша профилей

### Другие

- `GET /health` - Health check
- `GET /swagger/` - Swagger UI документация

## Swagger документация

Swagger документация доступна по адресу `http://localhost:8081/swagger/` после запуска сервиса.

Для генерации/обновления Swagger документации используйте:

```bash
make swagger
```

или

```bash
swag init -g cmd/employee/main.go -o internal/infrastructure/http/docs
```

## Примеры запросов

### Добавление сотрудника с автоматическим получением профиля

```bash
# Создание сотрудника с автоматическим получением имени из профиля
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "phone": "+79001234567",
    "max_id": "12345",
    "university_id": 1
  }'

# Ответ включает информацию об источнике профиля:
{
  "id": 123,
  "first_name": "Иван",
  "last_name": "Петров",
  "phone": "+79001234567",
  "max_id": "12345",
  "university_id": 1,
  "profile_source": "webhook",
  "profile_last_updated": "2024-01-15T10:30:00Z"
}
```

### Добавление сотрудника с явно указанными именами

```bash
# Создание сотрудника с предоставленными именами (приоритет над профилем)
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "phone": "+79001234567",
    "first_name": "Иван",
    "last_name": "Иванов",
    "middle_name": "Иванович",
    "max_id": "12345",
    "university_id": 1
  }'

# Ответ использует предоставленные имена:
{
  "id": 124,
  "first_name": "Иван",
  "last_name": "Иванов",
  "middle_name": "Иванович",
  "phone": "+79001234567",
  "max_id": "12345",
  "university_id": 1,
  "profile_source": "manual",
  "profile_last_updated": null
}
```

### Поиск сотрудников

```bash
curl "http://localhost:8081/employees?query=Иванов&limit=10"
```

### Получение сотрудника с информацией о профиле

```bash
curl "http://localhost:8081/employees/123" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Ответ включает информацию об источнике имени:
{
  "id": 123,
  "first_name": "Иван",
  "last_name": "Петров",
  "phone": "+79001234567",
  "max_id": "12345",
  "university_id": 1,
  "profile_source": "webhook",
  "profile_last_updated": "2024-01-15T10:30:00Z"
}
```

## Запуск

### Локально

```bash
# Установить зависимости
go mod download

# Запустить миграции
psql $DATABASE_URL < migrations/001_init.sql

# Запустить сервис
go run cmd/employee/main.go
```

### Docker Compose

```bash
docker-compose up -d
```

## Переменные окружения

### Основные настройки
- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт сервера (по умолчанию 8081)
- `GRPC_PORT` - Порт gRPC сервера (по умолчанию 9091)

### Интеграция с MAX API
- `MAXBOT_GRPC_ADDR` - Адрес MaxBot gRPC сервиса (по умолчанию maxbot-service:9095)
- `MAX_API_URL` - URL для MAX API (опционально)

### Интеграция профилей (NEW)
- `PROFILE_CACHE_ENABLED` - Включить интеграцию с кэшем профилей (по умолчанию true)
- `PROFILE_CACHE_TIMEOUT` - Таймаут запросов к кэшу профилей (по умолчанию 3s)

### Аутентификация
- `JWT_ACCESS_SECRET` - Секрет для JWT токенов доступа
- `JWT_REFRESH_SECRET` - Секрет для JWT токенов обновления
- `AUTH_GRPC_ADDR` - Адрес Auth gRPC сервиса (по умолчанию auth-service:9090)

## Структура проекта

```
employee-service/
├── cmd/
│   └── employee/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── server.go
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── employee.go
│   │   ├── employee_repository.go
│   │   ├── university.go
│   │   ├── university_repository.go
│   │   ├── max_service.go
│   │   └── errors.go
│   ├── infrastructure/
│   │   ├── http/
│   │   │   ├── handler.go
│   │   │   └── router.go
│   │   ├── repository/
│   │   │   ├── employee_postgres.go
│   │   │   └── university_postgres.go
│   │   └── max/
│   │       └── max_client.go
│   └── usecase/
│       └── employee_service.go
├── migrations/
│   └── 001_init.sql
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── README.md
```

