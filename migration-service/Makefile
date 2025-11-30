.PHONY: swagger swagger-fmt run test

# Generate Swagger documentation
swagger:
	swag init -g cmd/migration/main.go -o internal/infrastructure/http/docs

# Format Swagger comments
swagger-fmt:
	swag fmt

# Run the service
run:
	go run cmd/migration/main.go

# Run tests
test:
	go test -v ./...

# Install dependencies
deps:
	go mod download
	go mod tidy

# Build the service
build:
	go build -o bin/migration-service cmd/migration/main.go
