# Docker Cross-Service Dependencies Fix

## Проблема

После удаления `replace` директив из go.mod файлов, Docker не мог собрать сервисы, которые зависят от proto файлов других сервисов:

```
package auth-service/api/proto is not in std (/usr/local/go/src/auth-service/api/proto)
package maxbot-service/api/proto is not in std (/usr/local/go/src/maxbot-service/api/proto)
```

## Решение

### 1. Вернули `replace` директивы в go.mod

Для локальной разработки нужны `replace` директивы, которые указывают на локальные пути:

**chat-service/go.mod:**
```go
require (
    auth-service v0.0.0
    maxbot-service v0.0.0
    ...
)

replace (
    auth-service => ../auth-service
    maxbot-service => ../maxbot-service
)
```

**employee-service/go.mod:**
```go
require (
    auth-service v0.0.0
    maxbot-service v0.0.0
    ...
)

replace (
    auth-service => ../auth-service
    maxbot-service => ../maxbot-service
)
```

**structure-service/go.mod:**
```go
require (
    chat-service v0.0.0
    employee-service v0.0.0
    ...
)

replace (
    chat-service => ../chat-service
    employee-service => ../employee-service
)
```

### 2. Обновили Dockerfile для использования корневого контекста

Изменили контекст сборки в docker-compose.yml с `./service-name` на `.` (корневая директория):

```yaml
services:
  chat-service:
    build:
      context: .  # Было: ./chat-service
      dockerfile: ./chat-service/Dockerfile
```

### 3. Обновили Dockerfile для копирования зависимостей

Каждый Dockerfile теперь:
1. Копирует proto файлы и go.mod из зависимых сервисов
2. Копирует свой go.mod
3. Заменяет относительные пути в replace директивах на абсолютные для Docker
4. Скачивает зависимости
5. Копирует исходный код
6. Снова заменяет пути (т.к. go.mod был перезаписан)
7. Генерирует proto файлы для всех сервисов

**Пример для chat-service/Dockerfile:**
```dockerfile
# Копируем proto файлы и go.mod из других сервисов
COPY auth-service/api/proto /app/auth-service/api/proto
COPY auth-service/go.mod /app/auth-service/go.mod
COPY maxbot-service/api/proto /app/maxbot-service/api/proto
COPY maxbot-service/go.mod /app/maxbot-service/go.mod

# Копируем go mod файлы chat-service
COPY chat-service/go.mod chat-service/go.sum ./

# Исправляем replace директивы для Docker окружения
RUN sed -i 's|=> ../auth-service|=> /app/auth-service|g' go.mod && \
    sed -i 's|=> ../maxbot-service|=> /app/maxbot-service|g' go.mod

RUN go mod download

# Копируем исходный код chat-service
COPY chat-service/ .

# Снова исправляем replace директивы после копирования
RUN sed -i 's|=> ../auth-service|=> /app/auth-service|g' go.mod && \
    sed -i 's|=> ../maxbot-service|=> /app/maxbot-service|g' go.mod

# Генерируем proto файлы для всех сервисов
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/chat.proto

RUN protoc --proto_path=/app/auth-service --go_out=/app/auth-service --go_opt=paths=source_relative \
    --go-grpc_out=/app/auth-service --go-grpc_opt=paths=source_relative \
    api/proto/auth.proto

RUN protoc --proto_path=/app/maxbot-service --go_out=/app/maxbot-service --go_opt=paths=source_relative \
    --go-grpc_out=/app/maxbot-service --go-grpc_opt=paths=source_relative \
    api/proto/maxbot.proto
```

## Результат

✅ **Все сервисы теперь успешно собираются!**

```bash
docker images | grep "go-lang-max"
go-lang-max-migration-service   latest   eac3fbab9b82   44.5MB
go-lang-max-structure-service   latest   073edf790a7b   40.8MB
go-lang-max-chat-service        latest   bea015c029ff   37.9MB
go-lang-max-employee-service    latest   a98f3db7fad6   37.9MB
go-lang-max-auth-service        latest   8f7ace024232   36.9MB
go-lang-max-maxbot-service      latest   420b69c0952d   25.8MB
```

### Проверка сборки

```bash
# Все образы собрались успешно
docker-compose build --no-cache
# ✅ auth-service - собран
# ✅ chat-service - собран
# ✅ employee-service - собран
# ✅ structure-service - собран
# ✅ maxbot-service - собран
# ✅ migration-service - собран
```

## Команды для пересборки

```bash
# Пересобрать все сервисы
docker-compose build --no-cache

# Пересобрать конкретный сервис
docker-compose build --no-cache chat-service

# Запустить все сервисы
docker-compose up -d
```

## Зависимости между сервисами

- **chat-service** зависит от: auth-service, maxbot-service
- **employee-service** зависит от: auth-service, maxbot-service
- **structure-service** зависит от: chat-service, employee-service
- **migration-service** не имеет gRPC зависимостей от других сервисов

## Важные замечания

1. Для локальной разработки используйте `go mod tidy` после изменения зависимостей
2. Replace директивы в go.mod используют относительные пути для локальной разработки
3. В Docker эти пути автоматически заменяются на абсолютные через sed
4. Контекст сборки Docker должен быть корневой директорией проекта

## Известные проблемы (не связанные со сборкой)

После успешной сборки могут возникнуть runtime проблемы:

1. **Logger nil pointer** - в некоторых сервисах logger не инициализируется в middleware
2. **maxbot-service требует MAX_API_TOKEN** - сервис перезапускается без токена
3. **База данных employee-db** - возможны проблемы с именем базы данных

Эти проблемы не связаны с процессом сборки Docker и требуют отдельного исправления в коде приложений.
