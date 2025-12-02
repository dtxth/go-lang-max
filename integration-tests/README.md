# Integration Tests

This directory contains end-to-end integration tests for the Digital University microservices system.

## Prerequisites

- Docker and Docker Compose installed
- All services running via `docker-compose up`
- Databases initialized with migrations

## Running Tests

```bash
# Start all services
docker-compose up -d

# Wait for services to be healthy
sleep 30

# Run integration tests
cd integration-tests
go test -v ./...
```

## Test Coverage

1. **Employee Creation Flow** - Tests end-to-end employee creation with role assignment and MAX_id lookup
2. **Chat Filtering** - Tests role-based access control for chat lists (Superadmin, Curator, Operator)
3. **Excel Import** - Tests structure import from Excel files with full hierarchy creation
4. **Migration Sources** - Tests migration from database, Google Sheets, and Excel sources
5. **gRPC Communication** - Tests inter-service communication via gRPC

## Test Structure

- `employee_integration_test.go` - Employee service integration tests
- `chat_integration_test.go` - Chat service integration tests
- `structure_integration_test.go` - Structure service integration tests
- `migration_integration_test.go` - Migration service integration tests
- `grpc_integration_test.go` - gRPC communication tests
- `helpers.go` - Test utilities and helpers

## Configuration

Tests use the following service endpoints:
- Auth Service: http://localhost:8080 (gRPC: localhost:9090)
- Employee Service: http://localhost:8081 (gRPC: localhost:9091)
- Chat Service: http://localhost:8082 (gRPC: localhost:9092)
- Structure Service: http://localhost:8083 (gRPC: localhost:9093)
- MaxBot Service: gRPC localhost:9095
- Migration Service: http://localhost:8084

Database connections:
- Auth DB: localhost:5432
- Employee DB: localhost:5433
- Chat DB: localhost:5434
- Structure DB: localhost:5435
- Migration DB: localhost:5436
