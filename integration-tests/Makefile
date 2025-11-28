.PHONY: test test-employee test-chat test-structure test-migration test-grpc clean setup

# Run all integration tests
test:
	@echo "Running all integration tests..."
	go test -v -timeout 10m ./...

# Run specific test suites
test-employee:
	@echo "Running employee integration tests..."
	go test -v -timeout 5m -run TestEmployee

test-chat:
	@echo "Running chat integration tests..."
	go test -v -timeout 5m -run TestChat

test-structure:
	@echo "Running structure integration tests..."
	go test -v -timeout 5m -run TestStructure

test-migration:
	@echo "Running migration integration tests..."
	go test -v -timeout 5m -run TestMigration

test-grpc:
	@echo "Running gRPC integration tests..."
	go test -v -timeout 5m -run TestGRPC

# Setup test environment
setup:
	@echo "Setting up test environment..."
	cd .. && docker-compose up -d
	@echo "Waiting for services to be ready..."
	sleep 30

# Clean up test data
clean:
	@echo "Cleaning up test data..."
	@echo "Note: Run this manually if needed to clean databases"

# Stop all services
stop:
	@echo "Stopping all services..."
	cd .. && docker-compose down

# View logs
logs:
	cd .. && docker-compose logs -f

# Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	go test -v -timeout 10m -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
