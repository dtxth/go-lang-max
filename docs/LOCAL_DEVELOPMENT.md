# Локальная разработка без Docker

Руководство по запуску сервисов локально для разработки и тестирования Swagger.

## Проблема с Docker

Docker build не поддерживает `replace` директивы в go.mod с относительными путями. Для разработки и тестирования Swagger рекомендуется запускать сервисы локально.

## Требования

- Go 1.21+
- PostgreSQL 15
- protoc (для генерации gRPC кода)

## Подготовка

### 1. Установка PostgreSQL

```bash
# macOS
brew install postgresql@15
brew services start postgresql@15

# Создание баз данных
createdb auth_db
createdb employee_db
createdb chat_db
createdb structure_db
createdb migration_db
```

### 2. Применение миграций

```bash
# Auth Service
cd auth-service
export DATABASE_URL="postgres://localhost:5432/auth_db?sslmode=disable"
# Примените миграции вручную или через golang-migrate

# Повторите для других сервисов
```

## Запуск сервисов

### Auth Service

```bash
cd auth-service

# Установите зависимости
go mod download

# Настройте переменные окружения
export DATABASE_URL="postgres://localhost:5432/auth_db?sslmode=disable"
export HTTP_ADDR=":8080"
export GRPC_PORT="9090"
export JWT_ACCESS_SECRET="your-secret-key-min-32-chars-long"
export JWT_REFRESH_SECRET="your-refresh-secret-key-min-32"

# Запустите сервис
go run cmd/auth/main.go
```

**Swagger UI:** http://localhost:8080/swagger/index.html

### Employee Service

```bash
cd employee-service

export DATABASE_URL="postgres://localhost:5433/employee_db?sslmode=disable"
export PORT="8081"
export GRPC_PORT="9091"
export AUTH_SERVICE_GRPC="localhost:9090"
export MAXBOT_SERVICE_GRPC="localhost:9095"

go run cmd/employee/main.go
```

**Swagger UI:** http://localhost:8081/swagger/index.html

### Chat Service

```bash
cd chat-service

export DATABASE_URL="postgres://localhost:5434/chat_db?sslmode=disable"
export PORT="8082"
export GRPC_PORT="9092"
export AUTH_SERVICE_GRPC="localhost:9090"
export MAXBOT_SERVICE_GRPC="localhost:9095"

go run cmd/chat/main.go
```

**Swagger UI:** http://localhost:8082/swagger/index.html

### Structure Service

```bash
cd structure-service

export DATABASE_URL="postgres://localhost:5435/structure_db?sslmode=disable"
export PORT="8083"
export GRPC_PORT="9093"
export CHAT_SERVICE_GRPC="localhost:9092"
export EMPLOYEE_SERVICE_GRPC="localhost:9091"

go run cmd/structure/main.go
```

**Swagger UI:** http://localhost:8083/swagger/index.html

### Migration Service

```bash
cd migration-service

export DATABASE_URL="postgres://localhost:5436/migration_db?sslmode=disable"
export PORT="8084"
export CHAT_SERVICE_GRPC="localhost:9092"
export STRUCTURE_SERVICE_GRPC="localhost:9093"

go run cmd/migration/main.go
```

**Swagger UI:** http://localhost:8084/swagger/index.html

### MaxBot Service

```bash
cd maxbot-service

export MAX_API_TOKEN="your-max-api-token"
export MAX_API_URL="https://api.max.ru"
export GRPC_PORT="9095"

go run cmd/maxbot/main.go
```

## Workflow разработки

### 1. Обновление Swagger аннотаций

```bash
# Отредактируйте handler.go, добавьте/измените аннотации
vim auth-service/internal/infrastructure/http/handler.go

# Регенерируйте Swagger
cd auth-service
make swagger

# Перезапустите сервис (Ctrl+C и go run снова)
go run cmd/auth/main.go

# Обновите браузер (Ctrl+Shift+R)
# Откройте http://localhost:8080/swagger/index.html
```

### 2. Быстрое тестирование изменений

```bash
# Обновите код
vim auth-service/internal/infrastructure/http/handler.go

# Регенерируйте Swagger
make swagger

# Перезапустите (hot reload не поддерживается)
# Ctrl+C в терминале с go run
go run cmd/auth/main.go
```

### 3. Автоматическая перезагрузка (опционально)

Установите `air` для hot reload:

```bash
go install github.com/air-verse/air@latest

# Создайте .air.toml в директории сервиса
cd auth-service
air init

# Запустите с hot reload
air
```

## Преимущества локального запуска

✅ Быстрая итерация (нет пересборки Docker образов)  
✅ Swagger обновляется мгновенно  
✅ Легче отлаживать (можно использовать debugger)  
✅ Видны все логи в реальном времени  
✅ Не нужно ждать сборку Docker образа  

## Недостатки

❌ Нужно запускать каждый сервис отдельно  
❌ Нужно настраивать PostgreSQL локально  
❌ Нужно следить за портами  

## Решение проблемы с Docker

Для production и полного тестирования используйте Docker, но:

### Вариант 1: Удалите replace директивы (текущее решение)

Уже сделано для employee-service, chat-service, structure-service.

### Вариант 2: Используйте монорепозиторий build context

Измените docker-compose.yml:

```yaml
services:
  chat-service:
    build:
      context: .  # Корневая директория, не ./chat-service
      dockerfile: chat-service/Dockerfile
```

И обновите Dockerfile:

```dockerfile
WORKDIR /app/chat-service
COPY chat-service/go.mod chat-service/go.sum ./
# ... и т.д.
```

### Вариант 3: Скопируйте proto файлы

Скопируйте необходимые proto файлы в каждый сервис и измените импорты.

## Рекомендация

Для разработки и тестирования Swagger:
- ✅ Используйте **локальный запуск** (go run)
- ✅ Swagger обновляется мгновенно через `make swagger`

Для production и интеграционного тестирования:
- ✅ Используйте **Docker** с правильно настроенными зависимостями

## Быстрый старт для Swagger разработки

```bash
# 1. Запустите PostgreSQL
brew services start postgresql@15

# 2. Создайте базы данных
createdb auth_db

# 3. Запустите auth-service
cd auth-service
export DATABASE_URL="postgres://localhost:5432/auth_db?sslmode=disable"
export JWT_ACCESS_SECRET="test-secret-key-min-32-chars-long-12345"
export JWT_REFRESH_SECRET="test-refresh-secret-key-min-32-chars"
go run cmd/auth/main.go

# 4. Откройте Swagger UI
open http://localhost:8080/swagger/index.html

# 5. Обновите аннотации и регенерируйте
make swagger
# Перезапустите сервис (Ctrl+C и go run снова)
```

Готово! Теперь вы можете быстро итерировать по Swagger документации без Docker! 🚀
