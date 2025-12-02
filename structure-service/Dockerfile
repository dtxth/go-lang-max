FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем protoc и плагины для генерации
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH=$PATH:/root/go/bin

# Копируем proto файлы и go.mod из других сервисов сначала
COPY chat-service/api/proto /app/chat-service/api/proto
COPY chat-service/go.mod /app/chat-service/go.mod
COPY employee-service/api/proto /app/employee-service/api/proto
COPY employee-service/go.mod /app/employee-service/go.mod

# Копируем go mod файлы structure-service
COPY structure-service/go.mod structure-service/go.sum ./

# Исправляем replace директивы для Docker окружения
RUN sed -i 's|=> ../chat-service|=> /app/chat-service|g' go.mod && \
    sed -i 's|=> ../employee-service|=> /app/employee-service|g' go.mod

RUN go mod download

# Копируем исходный код structure-service
COPY structure-service/ .

# Снова исправляем replace директивы после копирования
RUN sed -i 's|=> ../chat-service|=> /app/chat-service|g' go.mod && \
    sed -i 's|=> ../employee-service|=> /app/employee-service|g' go.mod

# Генерируем proto файлы для всех сервисов
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/structure.proto

RUN protoc --proto_path=/app/chat-service --go_out=/app/chat-service --go_opt=paths=source_relative \
    --go-grpc_out=/app/chat-service --go-grpc_opt=paths=source_relative \
    api/proto/chat.proto

RUN protoc --proto_path=/app/employee-service --go_out=/app/employee-service --go_opt=paths=source_relative \
    --go-grpc_out=/app/employee-service --go-grpc_opt=paths=source_relative \
    api/proto/employee.proto

# Генерируем Swagger документацию
RUN go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/structure/main.go -o internal/infrastructure/http/docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/structure-service ./cmd/structure

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/structure-service .

EXPOSE 8082

CMD ["./structure-service"]

