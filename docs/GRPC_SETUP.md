# Настройка gRPC для микросервисов

Этот документ описывает настройку gRPC связи между микросервисами.

## Структура gRPC сервисов

### Auth Service (порт 9090)
- `ValidateToken` - проверка валидности токена
- `GetUser` - получение информации о пользователе по ID

### Chat Service (порт 9092)
- `GetChatByID` - получение чата по ID
- `CreateChat` - создание нового чата

### Employee Service (порт 9091)
- `GetUniversityByID` - получение вуза по ID
- `GetUniversityByINN` - получение вуза по ИНН
- `GetUniversityByINNAndKPP` - получение вуза по ИНН и КПП

### Structure Service (порт 9093)
- `GetUniversityByID` - получение вуза по ID
- `GetUniversityByINN` - получение вуза по ИНН

## Зависимости между сервисами

- **structure-service** использует **chat-service** для получения информации о чатах при импорте из Excel

## Установка зависимостей

### 1. Установите protoc и плагины

```bash
# Установка protoc (пример для macOS)
brew install protobuf

# Установка Go плагинов
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### 2. Добавьте gRPC зависимости в каждый сервис

```bash
# Auth Service
cd auth-service
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy

# Chat Service
cd ../chat-service
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy

# Employee Service
cd ../employee-service
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy

# Structure Service
cd ../structure-service
go get google.golang.org/grpc
go get google.golang.org/protobuf
go mod tidy
```

### 3. Сгенерируйте proto код

```bash
# Из корневой директории проекта
./generate_proto.sh
```

Или вручную для каждого сервиса:

```bash
# Auth Service
cd auth-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/auth.proto

# Chat Service
cd ../chat-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/chat.proto

# Employee Service
cd ../employee-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/employee.proto

# Structure Service
cd ../structure-service
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/structure.proto
```

## Переменные окружения

### Auth Service
- `GRPC_PORT` - порт для gRPC сервера (по умолчанию 9090)

### Chat Service
- `GRPC_PORT` - порт для gRPC сервера (по умолчанию 9092)

### Employee Service
- `GRPC_PORT` - порт для gRPC сервера (по умолчанию 9091)

### Structure Service
- `GRPC_PORT` - порт для gRPC сервера (по умолчанию 9093)
- `CHAT_SERVICE_GRPC` - адрес chat-service gRPC (по умолчанию localhost:9092)

## Запуск сервисов

Все сервисы теперь запускают одновременно HTTP и gRPC серверы:

- HTTP сервер на порту, указанном в `PORT`
- gRPC сервер на порту, указанном в `GRPC_PORT`

## Пример использования

### Вызов ValidateToken из другого сервиса

```go
import (
    "auth-service/api/proto"
    "google.golang.org/grpc"
)

conn, err := grpc.NewClient("localhost:9090", grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := proto.NewAuthServiceClient(conn)
resp, err := client.ValidateToken(ctx, &proto.ValidateTokenRequest{
    Token: "your-token-here",
})
```

## Примечания

- Все gRPC серверы используют insecure credentials (без TLS). Для production рекомендуется настроить TLS.
- Proto файлы должны быть сгенерированы перед компиляцией сервисов.
- При изменении proto файлов необходимо перегенерировать код и перезапустить сервисы.

