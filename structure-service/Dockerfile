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

# Модифицируем go_package в proto файле для structure-service
RUN sed -i 's|option go_package = "chat-service/api/proto;proto";|option go_package = "structure-service/api/proto/chat;chatproto";|' api/proto/chat/chat.proto || \
    sed -i '' 's|option go_package = "chat-service/api/proto;proto";|option go_package = "structure-service/api/proto/chat;chatproto";|' api/proto/chat/chat.proto

# Генерируем proto файлы
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/structure.proto \
    api/proto/chat/chat.proto

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

