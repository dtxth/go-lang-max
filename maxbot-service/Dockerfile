FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY maxbot-service/go.mod maxbot-service/go.sum ./
RUN go mod download

# Copy the source code
COPY maxbot-service/ .

# Install protobuf tools
RUN apk add --no-cache protobuf-dev
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf files
RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/maxbotproto/maxbot.proto

RUN protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/proto/authproto/auth.proto

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/maxbot-service ./cmd/maxbot

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/maxbot-service .

EXPOSE 9095

CMD ["./maxbot-service"]
