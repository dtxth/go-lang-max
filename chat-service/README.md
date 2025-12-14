# Chat Service

Сервис управления групповыми чатами для мини-приложения "Цифровой вуз".

## Описание

Сервис обеспечивает централизованное управление групповыми чатами: создание, взаимодействие, фильтрация и сортировка. Поддерживает работу с чатами из разных источников:
- Чаты из действующей админки групповых чатов
- Чаты, добавленные через бота-регистратора
- Чаты академических групп

## Функционал

- Поиск чатов по названию
- Пагинация списка чатов
- Фильтрация чатов по роли пользователя и вузу
- Управление администраторами чатов:
  - Получение списка всех администраторов с пагинацией и поиском
  - Получение администратора по ID
  - Добавление администратора
  - Удаление администратора (нельзя удалить последнего)

## Структура проекта

```
chat-service/
├── cmd/
│   └── chat/
│       └── main.go
├── internal/
│   ├── app/
│   │   └── server.go
│   ├── config/
│   │   └── config.go
│   ├── domain/
│   │   ├── chat.go
│   │   ├── administrator.go
│   │   ├── errors.go
│   │   ├── chat_repository.go
│   │   ├── administrator_repository.go
│   │   ├── university_repository.go
│   │   └── max_service.go
│   ├── infrastructure/
│   │   ├── http/
│   │   │   ├── handler.go
│   │   │   └── router.go
│   │   ├── repository/
│   │   │   ├── chat_postgres.go
│   │   │   ├── administrator_postgres.go
│   │   │   └── university_postgres.go
│   │   └── max/
│   │       └── max_client.go
│   └── usecase/
│       └── chat_service.go
├── migrations/
│   └── 001_init.sql
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── go.mod
```

## API Endpoints

### Чаты

- `GET /chats` - Поиск чатов по названию
- `GET /chats/all` - Получить все чаты с пагинацией
- `GET /chats/{id}` - Получить чат по ID

### Администраторы

- `GET /administrators` - Получить всех администраторов с пагинацией и поиском
- `GET /administrators/{admin_id}` - Получить администратора по ID
- `POST /chats/{chat_id}/administrators` - Добавить администратора к чату
- `DELETE /administrators/{admin_id}` - Удалить администратора из чата

### Параметры запросов

- `query` - Поисковый запрос (название чата)
- `limit` - Лимит результатов (по умолчанию 50, максимум 100)
- `offset` - Смещение для пагинации
- `user_role` - Роль пользователя (superadmin, admin, user)
- `university_id` - ID вуза (для фильтрации, если не superadmin)

## Запуск

### Локально

```bash
# Установить зависимости
go mod download

# Запустить миграции
make migrate

# Запустить сервис
make run
```

### Docker

```bash
# Собрать и запустить
make docker-build
make docker-up

# Остановить
make docker-down
```

## Конфигурация

Сервис использует переменные окружения:

- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт сервера (по умолчанию 8082)
- `MAX_API_URL` - URL для MAX API (опционально)

## База данных

Сервис использует PostgreSQL. Миграции находятся в директории `migrations/`.

### Таблицы

- `universities` - Вузы
- `chats` - Чаты
- `administrators` - Администраторы чатов

## Swagger

После генерации документации Swagger доступен по адресу:
```
http://localhost:8082/swagger/index.html
```

Для генерации документации:
```bash
make swagger
```

## Дополнительная документация

- [ADMINISTRATORS_API.md](./ADMINISTRATORS_API.md) - Полная документация API администраторов
- [ADMINISTRATORS_QUICK_START.md](./ADMINISTRATORS_QUICK_START.md) - Быстрый старт с примерами
- [ADMINISTRATORS_IMPLEMENTATION_SUMMARY.md](./ADMINISTRATORS_IMPLEMENTATION_SUMMARY.md) - Сводка реализации
- [DEPLOYMENT_SUCCESS.md](./DEPLOYMENT_SUCCESS.md) - Результаты развертывания и тестирования

