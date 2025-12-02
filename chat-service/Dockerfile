FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем protoc и плагины для генерации
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH=$PATH:/root/go/bin

# Копируем proto файлы и go.mod из других сервисов сначала
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

# Снова исправляем replace директивы после копирования (go.mod был перезаписан)
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

# Генерируем Swagger документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/chat/main.go -o internal/infrastructure/http/docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chat-service ./cmd/chat

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/chat-service .

EXPOSE 8082

CMD ["./chat-service"]

