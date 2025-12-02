# Employee Service

Сервис управления сотрудниками вузов для мини-приложения.

## Функциональность

- **Поиск сотрудников** по имени, фамилии и названию вуза
- **Добавление сотрудников** по номеру телефона
- **Автоматическое получение MAX_id** по номеру телефона
- **Управление вузами** (автоматическое создание при добавлении сотрудника)
- **CRUD операции** для сотрудников

## API Endpoints

### Сотрудники

- `GET /employees?query=...&limit=50&offset=0` - Поиск сотрудников
- `GET /employees/all?limit=50&offset=0` - Получить всех сотрудников
- `GET /employees/{id}` - Получить сотрудника по ID
- `POST /employees` - Добавить сотрудника
- `PUT /employees/{id}` - Обновить сотрудника
- `DELETE /employees/{id}` - Удалить сотрудника

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

### Добавление сотрудника

```bash
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "+79001234567",
    "first_name": "Иван",
    "last_name": "Иванов",
    "middle_name": "Иванович",
    "inn": "1234567890",
    "kpp": "123456789",
    "university_name": "МГУ"
  }'
```

### Поиск сотрудников

```bash
curl "http://localhost:8081/employees?query=Иванов&limit=10"
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

- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт сервера (по умолчанию 8081)
- `MAX_API_URL` - URL для MAX API (опционально)

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

