.PHONY: build run test clean docker-build docker-up docker-down migrate swagger

build:
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

