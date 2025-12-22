.PHONY: build run test clean docker-build docker-up docker-down migrate swagger proto-gen

proto-gen:
	@echo "Generating proto files for chat-service..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/chat.proto

build: proto-gen
	go build -o bin/chat-service ./cmd/chat

run:
	go run ./cmd/chat/main.go

test:
	go test ./...

clean:
	rm -rf bin/

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

migrate:
	psql $(DATABASE_URL) < migrations/001_init.sql

swagger:
	@which swag > /dev/null || (echo "Installing swag..." && go install github.com/swaggo/swag/cmd/swag@latest)
	@swag init -g cmd/chat/main.go -o internal/infrastructure/http/docs || ~/go/bin/swag init -g cmd/chat/main.go -o internal/infrastructure/http/docs

