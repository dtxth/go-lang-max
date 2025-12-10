FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем необходимые инструменты
RUN apk add --no-cache git

# Copy the entire project (needed for any replace directives)
COPY . .

# Change to migration-service directory
WORKDIR /app/migration-service

# Download dependencies
RUN go mod download

# Генерируем Swagger документацию
RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g cmd/migration/main.go -o internal/infrastructure/http/docs

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /migration-service ./cmd/migration

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /migration-service .

EXPOSE 8084

CMD ["./migration-service"]
