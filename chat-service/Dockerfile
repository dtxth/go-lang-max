FROM golang:1.24-alpine AS builder

WORKDIR /app

# Устанавливаем protoc и плагины для генерации
RUN apk add --no-cache protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
ENV PATH=$PATH:/root/go/bin

# Копируем go mod файлы
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Генерируем proto файлы
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/chat.proto

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/chat-service ./cmd/chat

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник из builder
COPY --from=builder /app/chat-service .

EXPOSE 8082

CMD ["./chat-service"]

