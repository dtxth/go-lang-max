# build stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
# install git for go modules if private modules are used
RUN apk add --no-cache git protobuf protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest
ENV PATH=$PATH:/root/go/bin
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Генерируем proto файлы
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/auth.proto
# Генерируем Swagger документацию
RUN swag init -g cmd/auth/main.go -o internal/infrastructure/http/docs
RUN CGO_ENABLED=0 GOOS=linux go build -o /auth-service ./cmd/auth

# final
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /auth-service .
EXPOSE 8080
ENV HTTP_ADDR=:8080
CMD ["./auth-service"]