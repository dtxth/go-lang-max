.PHONY: build run test clean docker-build docker-up docker-down migrate swagger proto-gen

proto-gen:
	@echo "Generating proto files for employee-service..."
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/proto/employee.proto

build: proto-gen
	go build -o bin/employee-service ./cmd/employee

run:
	go run ./cmd/employee/main.go

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
	swag init -g cmd/employee/main.go -o internal/infrastructure/http/docs

