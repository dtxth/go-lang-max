# go-lang-max

Микросервисная архитектура для мини-приложения "Цифровой вуз" на базе Go.

## Описание

Проект представляет собой систему микросервисов, обеспечивающих функциональность для управления университетами, сотрудниками, чатами и структурными подразделениями. Все сервисы взаимодействуют через HTTP REST API и gRPC.

## Архитектура

Проект состоит из 4 основных микросервисов:

### 1. Auth Service (порт 8080, gRPC: 9090)
Сервис аутентификации и авторизации пользователей.
- Управление пользователями
- JWT токены (access и refresh)
- Валидация токенов через gRPC
- Роли пользователей (superadmin, admin, user)

### 2. Chat Service (порт 8082, gRPC: 9092)
Сервис управления групповыми чатами.
- Создание и управление чатами
- Управление администраторами чатов
- Поиск и фильтрация чатов
- Интеграция с MAX API

### 3. Employee Service (порт 8081, gRPC: 9091)
Сервис управления сотрудниками.
- Управление сотрудниками вузов
- Работа с данными университетов
- Интеграция с MAX API

### 4. Structure Service (порт 8083, gRPC: 9093)
Сервис управления структурными подразделениями.
- Импорт данных из Excel
- Управление структурой университетов
- Интеграция с Chat Service через gRPC

## Технологический стек

- **Язык**: Go
- **База данных**: PostgreSQL 15
- **API**: REST (HTTP) + gRPC
- **Контейнеризация**: Docker, Docker Compose
- **Документация API**: Swagger/OpenAPI

## Быстрый старт

### Требования

- Go 1.21+
- Docker и Docker Compose
- protoc (для генерации gRPC кода)
- PostgreSQL 15 (или используйте Docker Compose)

### Установка зависимостей

```bash
# Установка protoc (macOS)
brew install protobuf

# Установка Go плагинов для gRPC
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Генерация gRPC кода

```bash
# Из корневой директории проекта
chmod +x generate_proto.sh
./generate_proto.sh
```

### Запуск через Docker Compose

```bash
# Запуск всех сервисов и баз данных
docker-compose up -d

# Просмотр логов
docker-compose logs -f

# Остановка всех сервисов
docker-compose down

# Остановка с удалением volumes
docker-compose down -v
```

Все сервисы будут доступны по следующим адресам:
- Auth Service: http://localhost:8080
- Chat Service: http://localhost:8082
- Employee Service: http://localhost:8081
- Structure Service: http://localhost:8083

### Запуск локально (без Docker)

Для каждого сервиса:

```bash
# Перейти в директорию сервиса
cd auth-service  # или chat-service, employee-service, structure-service

# Установить зависимости
go mod download

# Запустить миграции (если необходимо)
# Убедитесь, что PostgreSQL запущен и настроен

# Запустить сервис
go run cmd/auth/main.go  # или соответствующий путь для других сервисов
```

## Структура проекта

```
go-lang-max/
├── auth-service/          # Сервис аутентификации
│   ├── api/proto/         # gRPC proto файлы
│   ├── cmd/               # Точка входа
│   ├── internal/          # Внутренняя логика
│   │   ├── app/          # Инициализация приложения
│   │   ├── config/       # Конфигурация
│   │   ├── domain/       # Доменные модели
│   │   ├── infrastructure/ # Реализации (HTTP, gRPC, репозитории)
│   │   └── usecase/      # Бизнес-логика
│   ├── migrations/       # Миграции БД
│   └── Dockerfile
├── chat-service/          # Сервис чатов
├── employee-service/     # Сервис сотрудников
├── structure-service/    # Сервис структуры
├── docker-compose.yml    # Оркестрация всех сервисов
├── generate_proto.sh     # Скрипт генерации gRPC кода
└── GRPC_SETUP.md         # Документация по настройке gRPC
```

## gRPC взаимодействие

Сервисы взаимодействуют через gRPC для внутренней коммуникации:

- **Auth Service** предоставляет методы валидации токенов
- **Chat Service** предоставляет методы для работы с чатами
- **Employee Service** предоставляет методы для работы с университетами
- **Structure Service** использует Chat Service для получения информации о чатах

Подробная информация о gRPC настройке и использовании описана в [GRPC_SETUP.md](./GRPC_SETUP.md).

## Базы данных

Каждый сервис использует свою собственную базу данных PostgreSQL:

- **auth-db**: порт 5432
- **employee-db**: порт 5433
- **chat-db**: порт 5434
- **structure-db**: порт 5435

Миграции автоматически применяются при первом запуске через Docker Compose.

## API Документация

После запуска сервисов, Swagger документация доступна по адресам:

- Auth Service: http://localhost:8080/swagger/index.html
- Chat Service: http://localhost:8082/swagger/index.html
- Employee Service: http://localhost:8081/swagger/index.html
- Structure Service: http://localhost:8083/swagger/index.html

## Переменные окружения

### Auth Service
- `DATABASE_URL` - URL подключения к PostgreSQL
- `HTTP_ADDR` - Адрес HTTP сервера (по умолчанию :8080)
- `GRPC_PORT` - Порт gRPC сервера (по умолчанию 9090)
- `JWT_ACCESS_SECRET` - Секретный ключ для access токенов
- `JWT_REFRESH_SECRET` - Секретный ключ для refresh токенов
- `ACCESS_MINUTES` - Время жизни access токена в минутах (по умолчанию 15)
- `REFRESH_HOURS` - Время жизни refresh токена в часах (по умолчанию 168)

### Chat Service
- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт HTTP сервера (по умолчанию 8082)
- `GRPC_PORT` - Порт gRPC сервера (по умолчанию 9092)
- `MAX_API_URL` - URL для MAX API (опционально)

### Employee Service
- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт HTTP сервера (по умолчанию 8081)
- `GRPC_PORT` - Порт gRPC сервера (по умолчанию 9091)
- `MAX_API_URL` - URL для MAX API (опционально)

### Structure Service
- `DATABASE_URL` - URL подключения к PostgreSQL
- `PORT` - Порт HTTP сервера (по умолчанию 8080)
- `GRPC_PORT` - Порт gRPC сервера (по умолчанию 9093)
- `CHAT_SERVICE_GRPC` - Адрес chat-service gRPC (по умолчанию localhost:9092)

## Разработка

### Добавление нового сервиса

1. Создайте директорию для сервиса
2. Инициализируйте Go модуль: `go mod init <service-name>`
3. Добавьте структуру проекта (cmd, internal, migrations)
4. Добавьте Dockerfile и docker-compose.yml для сервиса
5. Обновите корневой docker-compose.yml
6. Добавьте proto файлы в `api/proto/`
7. Обновите `generate_proto.sh` для генерации кода

### Генерация Swagger документации

Для каждого сервиса:

```bash
cd <service-name>
make swagger  # или используйте swag init
```

## Тестирование

```bash
# Запуск тестов для конкретного сервиса
cd auth-service
go test ./...

# Запуск всех тестов
go test ./...
```
