FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем Swagger документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/migration/main.go -o internal/infrastructure/http/docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/migration-service ./cmd/migration

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/migration-service .

EXPOSE 8084

CMD ["./migration-service"]
