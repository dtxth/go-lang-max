# Digital University MVP - Makefile

.PHONY: help build up down logs test test-e2e clean clean-volumes restart setup health urls monitor deploy-rebuild

# Default target
help:
	@echo "Available commands:"
	@echo ""
	@echo "🚀 Service Management:"
	@echo "  build      - Build all services"
	@echo "  up         - Start all services"
	@echo "  down       - Stop all services"
	@echo "  restart    - Restart all services"
	@echo "  setup      - Setup development environment"
	@echo ""
	@echo "📊 Monitoring:"
	@echo "  logs       - Show logs from all services"
	@echo "  health     - Check health of all services"
	@echo "  urls       - Show service URLs"
	@echo "  monitor    - Show service status and resource usage"
	@echo ""
	@echo "🧪 Testing:"
	@echo "  test       - Run unit tests for all services"
	@echo "  test-e2e   - Run all end-to-end tests"
	@echo "  test-e2e-auth        - Test Auth Service"
	@echo "  test-e2e-structure   - Test Structure Service"
	@echo "  test-e2e-employee    - Test Employee Service"
	@echo "  test-e2e-chat        - Test Chat Service"
	@echo "  test-e2e-maxbot      - Test MaxBot Service"
	@echo "  test-e2e-migration   - Test Migration Service"
	@echo "  test-e2e-integration - Test service integration"
	@echo "  test-load            - Run load tests"
	@echo "  benchmark            - Run benchmark tests"
	@echo "  quick-test           - Quick health check tests"
	@echo ""
	@echo "🧹 Maintenance:"
	@echo "  clean         - Clean up containers and volumes"
	@echo "  clean-volumes - Clean up all Docker volumes"
	@echo "  db-reset      - Reset all databases"
	@echo "  deploy-rebuild - Full rebuild and deploy"
	@echo ""
	@echo "👨‍💻 Development:"
	@echo "  dev-up     - Start only databases for development"
	@echo "  dev-down   - Stop development services"

# Build all services
build:
	@echo "Building all services..."
	docker-compose build

# Start all services
up:
	@echo "Starting all services..."
	docker-compose up -d
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo "Services are starting up. Check logs with 'make logs'"

# Stop all services
down:
	@echo "Stopping all services..."
	docker-compose down

# Show logs
logs:
	docker-compose logs -f

# Run unit tests for all services
test:
	@echo "Running unit tests..."
	@echo "Testing auth-service..."
	cd auth-service && go test ./... -v
	@echo "Testing employee-service..."
	cd employee-service && go test ./... -v
	@echo "Testing chat-service..."
	cd chat-service && go test ./... -v
	@echo "Testing structure-service..."
	cd structure-service && go test ./... -v
	@echo "Testing maxbot-service..."
	cd maxbot-service && go test ./... -v
	@echo "Testing migration-service..."
	cd migration-service && go test ./... -v

# Run end-to-end tests
test-e2e:
	@echo "Running end-to-end tests..."
	@echo "Checking if services are running..."
	@docker-compose ps
	@echo "Starting E2E tests..."
	cd e2e-tests && go mod tidy && go test -v ./... -timeout 10m

# Run specific E2E test
test-e2e-auth:
	cd e2e-tests && go test -v -run TestAuthService -timeout 5m

test-e2e-structure:
	cd e2e-tests && go test -v -run TestStructureService -timeout 5m

test-e2e-employee:
	cd e2e-tests && go test -v -run TestEmployeeService -timeout 5m

test-e2e-chat:
	cd e2e-tests && go test -v -run TestChatService -timeout 5m

test-e2e-maxbot:
	cd e2e-tests && go test -v -run TestMaxBotService -timeout 5m

test-e2e-migration:
	cd e2e-tests && go test -v -run TestMigrationService -timeout 5m

test-e2e-integration:
	cd e2e-tests && go test -v -run TestIntegration -timeout 10m

# Run load tests
test-load:
	cd e2e-tests && go test -v -run TestLoadTest -timeout 15m

# Run benchmark tests
benchmark:
	cd e2e-tests && go test -bench=. -benchmem -timeout 10m

# Clean up
clean:
	@echo "Cleaning up containers and volumes..."
	docker-compose down -v
	docker system prune -f

# Clean up all Docker volumes
clean-volumes:
	@echo "Cleaning up all Docker volumes..."
	docker-compose down -v
	docker volume rm auth-service_pgdata || true
	docker volume rm go-microservices_chat_db_data || true
	docker volume rm go-microservices_employee_db_data || true
	docker volume rm go-microservices_structure_pgdata || true
	docker volume rm go-microservices_migration_db_data || true
	docker volume rm go-microservices_redis_data || true
	docker system prune -f --volumes
	@echo "All volumes cleaned up!"

# Full rebuild and deploy
deploy-rebuild:
	@echo "Starting full rebuild and deploy..."
	@echo "Step 1: Cleaning up..."
	make clean-volumes
	@echo "Step 2: Building services..."
	docker-compose build --no-cache
	@echo "Step 3: Setting up environment..."
	make setup
	@echo "Step 4: Starting services..."
	make up
	@echo "Step 5: Waiting for services to be ready..."
	@sleep 30
	@echo "Step 6: Running health checks..."
	make health
	@echo "Deploy rebuild complete!"

# Restart all services
restart: down up

# Setup development environment
setup:
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env file with your configuration"; \
	fi
	@echo "Creating required volumes..."
	docker volume create auth-service_pgdata || true
	docker volume create go-microservices_chat_db_data || true
	docker volume create go-microservices_employee_db_data || true
	docker volume create go-microservices_structure_pgdata || true
	@echo "Setup complete!"

# Health check for all services
health:
	@echo "Checking service health..."
	@curl -f http://localhost:8080/health && echo " ✓ Auth Service" || echo " ✗ Auth Service"
	@curl -f http://localhost:8081/health && echo " ✓ Employee Service" || echo " ✗ Employee Service"
	@curl -f http://localhost:8082/health && echo " ✓ Chat Service" || echo " ✗ Chat Service"
	@curl -f http://localhost:8083/health && echo " ✓ Structure Service" || echo " ✗ Structure Service"
	@curl -f http://localhost:8084/health && echo " ✓ Migration Service" || echo " ✗ Migration Service"
	@curl -f http://localhost:8095/health && echo " ✓ MaxBot Service" || echo " ✗ MaxBot Service"

# Show service URLs
urls:
	@echo "Service URLs:"
	@echo "  Auth Service:      http://localhost:8080"
	@echo "  Employee Service:  http://localhost:8081"
	@echo "  Chat Service:      http://localhost:8082"
	@echo "  Structure Service: http://localhost:8083"
	@echo "  Migration Service: http://localhost:8084"
	@echo "  MaxBot Service:    http://localhost:8095"
	@echo ""
	@echo "Swagger Documentation:"
	@echo "  Auth Service:      http://localhost:8080/swagger/"
	@echo "  Structure Service: http://localhost:8083/swagger/"

# Development helpers
dev-up:
	@echo "Starting services for development..."
	docker-compose up -d auth-db employee-db chat-db structure-db migration-db redis
	@echo "Databases and Redis are running. Start individual services manually for development."

dev-down:
	docker-compose down

# Database operations
db-reset:
	@echo "Resetting all databases..."
	docker-compose down -v
	docker volume rm auth-service_pgdata go-microservices_chat_db_data go-microservices_employee_db_data go-microservices_structure_pgdata || true
	docker volume create auth-service_pgdata
	docker volume create go-microservices_chat_db_data
	docker volume create go-microservices_employee_db_data
	docker volume create go-microservices_structure_pgdata
	@echo "Databases reset complete"

# Monitoring
monitor:
	@echo "Service status:"
	@docker-compose ps
	@echo ""
	@echo "Resource usage:"
	@docker stats --no-stream

# Quick test - just health checks
quick-test:
	@echo "Running quick health check tests..."
	cd e2e-tests && go test -v -run "Health" -timeout 2m