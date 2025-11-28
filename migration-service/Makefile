.PHONY: build run test clean migrate-up migrate-down

build:
	go build -o migration-service ./cmd/migration

run:
	go run ./cmd/migration/main.go

test:
	go test -v ./...

clean:
	rm -f migration-service

migrate-up:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5436/migration_db?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgresql://postgres:postgres@localhost:5436/migration_db?sslmode=disable" down

docker-build:
	docker build -t migration-service:latest .

docker-run:
	docker-compose up -d
