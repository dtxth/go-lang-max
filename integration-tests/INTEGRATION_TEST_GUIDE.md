# Integration Test Guide

## Overview

This directory contains comprehensive end-to-end integration tests for the Digital University microservices system. These tests validate the complete functionality of the system including inter-service communication, data persistence, and business logic.

## Test Coverage

### 1. Employee Integration Tests (`employee_integration_test.go`)

**TestEmployeeCreationWithRoleAndMaxID**
- Validates end-to-end employee creation flow
- Tests role assignment via Auth Service gRPC
- Tests MAX_id lookup via MaxBot Service gRPC
- Verifies data persistence across services
- **Requirements**: 2.1, 2.3, 3.1, 3.4

**TestEmployeeRoleUpdate**
- Tests role synchronization between Employee and Auth services
- Validates role update propagation
- **Requirements**: 2.4

**TestEmployeeDeletion**
- Tests permission revocation on employee deletion
- Validates cascade deletion and token invalidation
- **Requirements**: 2.5

**TestBatchMaxIDUpdate**
- Tests batch MAX_id update functionality
- Validates batch processing and reporting
- **Requirements**: 4.1, 4.2, 4.4, 4.5

### 2. Chat Integration Tests (`chat_integration_test.go`)

**TestChatFilteringBySuperadmin**
- Validates that superadmin sees all chats across universities
- **Requirements**: 5.1

**TestChatFilteringByCurator**
- Validates that curator sees only their university's chats
- **Requirements**: 5.2

**TestChatFilteringByOperator**
- Validates that operator sees only their department's chats
- **Requirements**: 5.3

**TestChatAdministratorManagement**
- Tests adding and removing chat administrators
- Validates last administrator protection
- Tests MAX_id lookup for administrators
- **Requirements**: 6.1, 6.2, 6.3, 6.4

**TestChatPagination**
- Tests pagination with limit and offset
- Validates limit capping at 100
- Verifies total count in metadata
- **Requirements**: 16.1, 16.2, 16.3, 16.4

**TestChatSearch**
- Tests full-text search functionality
- Validates Russian language support
- Tests multi-word search
- **Requirements**: 17.1, 17.2, 17.4

### 3. Structure Integration Tests (`structure_integration_test.go`)

**TestExcelImportWithFullStructure**
- Tests Excel file import
- Validates full hierarchy creation (University → Branch → Faculty → Group)
- **Requirements**: 9.1, 9.2, 9.3, 12.1, 12.2

**TestUniversityStructureRetrieval**
- Tests retrieving complete university hierarchy
- Validates nested JSON structure
- Tests chat information inclusion
- **Requirements**: 10.2, 10.5, 13.1, 13.2

**TestDepartmentManagerAssignment**
- Tests operator assignment to departments
- Validates employee verification via Employee Service
- **Requirements**: 11.1, 11.2, 11.3

**TestStructureAlphabeticalOrdering**
- Validates alphabetical ordering of entities at each level
- **Requirements**: 13.5

### 4. Migration Integration Tests (`migration_integration_test.go`)

**TestDatabaseMigration**
- Tests migration from existing database (admin_panel source)
- Validates chat creation with correct source
- Tests migration job tracking
- **Requirements**: 7.1, 7.2, 7.3, 7.4, 7.5

**TestGoogleSheetsMigration**
- Tests migration from Google Sheets (bot_registrar source)
- Validates Google Sheets API integration
- **Requirements**: 8.1, 8.2, 8.3, 8.4, 8.5

**TestExcelMigration**
- Tests migration from Excel files (academic_group source)
- Validates structure and chat creation
- Tests group-chat linking
- **Requirements**: 9.1, 9.2, 9.3, 9.4, 9.5

**TestMigrationJobListing**
- Tests listing all migration jobs
- Validates job status tracking
- **Requirements**: 20.5

**TestMigrationErrorTracking**
- Tests error tracking during migration
- Validates error reporting
- **Requirements**: 20.3, 20.4

### 5. gRPC Integration Tests (`grpc_integration_test.go`)

**TestAuthServiceGRPC**
- Tests Auth Service gRPC connectivity
- Validates connection state

**TestMaxBotServiceGRPC**
- Tests MaxBot Service gRPC connectivity
- Validates phone normalization service

**TestChatServiceGRPC**
- Tests Chat Service gRPC connectivity

**TestEmployeeServiceGRPC**
- Tests Employee Service gRPC connectivity

**TestStructureServiceGRPC**
- Tests Structure Service gRPC connectivity

**TestGRPCRetryMechanism**
- Tests retry logic for failed gRPC calls
- Validates exponential backoff
- **Requirements**: 18.4, 18.5

**TestInterServiceCommunication**
- Tests end-to-end communication between services
- Validates Employee → Auth → MaxBot flow
- Validates Chat → Auth → MaxBot flow
- **Requirements**: 18.1, 18.2, 18.3

**TestGRPCConnectionPooling**
- Tests connection reuse and pooling
- Validates performance under load

## Prerequisites

### Required Services
- Docker and Docker Compose
- Go 1.21 or higher
- PostgreSQL 15 (via Docker)

### Environment Setup
All services must be running via Docker Compose:
```bash
docker-compose up -d
```

### Database Migrations
Ensure all database migrations have been applied:
```bash
# Auth Service
docker exec -it auth-db psql -U postgres -d postgres -f /docker-entrypoint-initdb.d/001_init.sql

# Employee Service
docker exec -it employee-db psql -U employee_user -d employee_db -f /docker-entrypoint-initdb.d/001_init.sql

# Chat Service
docker exec -it chat-db psql -U chat_user -d chat_db -f /docker-entrypoint-initdb.d/001_init.sql

# Structure Service
docker exec -it structure-db psql -U postgres -d postgres -f /docker-entrypoint-initdb.d/001_init.sql

# Migration Service
docker exec -it migration-db psql -U postgres -d migration_db -f /docker-entrypoint-initdb.d/001_init.sql
```

## Running Tests

### Run All Tests
```bash
# Using the test runner script
./run_tests.sh

# Or using Make
make test

# Or using go test directly
go test -v -timeout 10m ./...
```

### Run Specific Test Suites
```bash
# Employee tests only
make test-employee

# Chat tests only
make test-chat

# Structure tests only
make test-structure

# Migration tests only
make test-migration

# gRPC tests only
make test-grpc
```

### Run Individual Tests
```bash
# Run a specific test
go test -v -run TestEmployeeCreationWithRoleAndMaxID

# Run tests matching a pattern
go test -v -run TestChat
```

### Run with Coverage
```bash
make coverage
# Opens coverage.html in browser
```

## Test Configuration

### Service Endpoints
Tests connect to services on localhost:
- Auth Service: http://localhost:8080 (gRPC: 9090)
- Employee Service: http://localhost:8081 (gRPC: 9091)
- Chat Service: http://localhost:8082 (gRPC: 9092)
- Structure Service: http://localhost:8083 (gRPC: 9093)
- MaxBot Service: gRPC localhost:9095
- Migration Service: http://localhost:8084

### Database Connections
Tests connect directly to databases for setup and cleanup:
- Auth DB: localhost:5432
- Employee DB: localhost:5433
- Chat DB: localhost:5434
- Structure DB: localhost:5435
- Migration DB: localhost:5436

## Test Data Management

### Setup
Each test creates its own test data using:
- Direct database inserts for complex scenarios
- HTTP API calls for realistic flows
- Test helper functions for common operations

### Cleanup
Tests clean up after themselves:
- Database tables are truncated after each test
- Test users are removed
- Temporary files are deleted

### Isolation
Tests are designed to be independent:
- Each test uses unique identifiers (timestamps, random values)
- Tests can run in parallel (with caution)
- No shared state between tests

## Troubleshooting

### Services Not Ready
If tests fail with connection errors:
```bash
# Check service health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
curl http://localhost:8084/health

# Check service logs
docker-compose logs auth-service
docker-compose logs employee-service
docker-compose logs chat-service
docker-compose logs structure-service
docker-compose logs migration-service
```

### Database Connection Issues
```bash
# Check database containers
docker ps | grep db

# Test database connectivity
docker exec -it auth-db psql -U postgres -c "SELECT 1"
docker exec -it employee-db psql -U employee_user -d employee_db -c "SELECT 1"
docker exec -it chat-db psql -U chat_user -d chat_db -c "SELECT 1"
```

### gRPC Connection Failures
```bash
# Check gRPC ports are exposed
docker ps | grep service

# Test gRPC connectivity
grpcurl -plaintext localhost:9090 list
grpcurl -plaintext localhost:9091 list
grpcurl -plaintext localhost:9092 list
```

### Test Timeouts
If tests timeout:
- Increase timeout: `go test -timeout 20m`
- Check service performance: `docker stats`
- Review service logs for slow queries

### Cleanup Issues
If cleanup fails:
```bash
# Manually clean databases
docker exec -it auth-db psql -U postgres -c "TRUNCATE users, user_roles CASCADE"
docker exec -it employee-db psql -U employee_user -d employee_db -c "TRUNCATE employees, universities CASCADE"
docker exec -it chat-db psql -U chat_user -d chat_db -c "TRUNCATE chats, administrators CASCADE"
```

## Best Practices

### Writing New Tests
1. Follow existing test patterns
2. Use descriptive test names
3. Add comments explaining what's being tested
4. Clean up test data
5. Use assertions from testify
6. Handle errors properly

### Test Organization
- Group related tests in the same file
- Use subtests for variations: `t.Run("subtest", func(t *testing.T) {...})`
- Keep tests focused and single-purpose

### Performance
- Minimize database operations
- Reuse connections where possible
- Use parallel tests cautiously: `t.Parallel()`
- Clean up efficiently

## CI/CD Integration

### GitHub Actions Example
```yaml
name: Integration Tests

on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Start services
        run: docker-compose up -d
      - name: Wait for services
        run: sleep 30
      - name: Run tests
        run: cd integration-tests && go test -v ./...
      - name: Stop services
        run: docker-compose down
```

## Maintenance

### Updating Tests
When adding new features:
1. Add corresponding integration tests
2. Update this documentation
3. Ensure tests pass locally
4. Update CI/CD pipeline if needed

### Deprecating Tests
When removing features:
1. Remove or skip obsolete tests
2. Update documentation
3. Clean up test data and helpers

## Support

For issues or questions:
- Check service logs: `docker-compose logs -f`
- Review test output: `test_results.log`
- Consult service-specific documentation
- Check requirements and design documents

## References

- Requirements: `.kiro/specs/digital-university-mvp-completion/requirements.md`
- Design: `.kiro/specs/digital-university-mvp-completion/design.md`
- Tasks: `.kiro/specs/digital-university-mvp-completion/tasks.md`
- Docker Compose: `../docker-compose.yml`
