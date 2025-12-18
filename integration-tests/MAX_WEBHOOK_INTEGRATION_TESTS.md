# MAX Webhook Profile Integration Tests

This document describes the integration tests for the MAX webhook profile integration feature.

## Overview

The integration tests validate the complete end-to-end flow from webhook events to employee creation, including:

1. **End-to-end webhook to employee creation flow** - Tests the complete pipeline from receiving a webhook event to creating an employee with cached profile data
2. **Profile persistence across service restarts** - Verifies that profiles stored in Redis persist with proper TTL settings
3. **Concurrent webhook processing** - Tests that multiple webhook events can be processed simultaneously without data corruption
4. **Different webhook event types** - Validates handling of `message_new`, `callback_query`, and unknown event types

## Test Structure

### Core Test Files

- `max_webhook_profile_integration_test.go` - Main integration test file containing all webhook profile integration tests

### Test Categories

#### 1. Setup Tests
- `TestIntegrationSetup` - Validates test structure and JSON marshaling/unmarshaling

#### 2. Full Integration Tests (require running services)
- `TestEndToEndWebhookToEmployeeCreationWithServices` - Complete webhook to employee creation flow
- `TestConcurrentWebhookProcessingWithServices` - Concurrent webhook processing
- `TestProfilePersistenceAcrossServiceRestartsWithServices` - Profile persistence validation
- `TestWebhookEventTypesWithServices` - Different webhook event type handling

## Prerequisites

### For Basic Tests
- Go 1.24+
- Test dependencies (automatically installed via `go mod download`)

### For Full Integration Tests
- All microservices running (auth-service, employee-service, maxbot-service)
- Redis instance running on localhost:6379
- PostgreSQL databases for each service

## Running Tests

### Run All Tests
```bash
cd integration-tests
go test -v
```

### Run Specific Test Categories

#### Basic Setup Tests (no services required)
```bash
go test -v -run TestIntegrationSetup
```

#### Full Integration Tests (requires services)
```bash
# Start services first
docker-compose up -d

# Wait for services to be ready
sleep 30

# Run integration tests
go test -v -run "WithServices"
```

#### Individual Test Cases
```bash
# Test webhook to employee creation flow
go test -v -run TestEndToEndWebhookToEmployeeCreationWithServices

# Test concurrent processing
go test -v -run TestConcurrentWebhookProcessingWithServices

# Test profile persistence
go test -v -run TestProfilePersistenceAcrossServiceRestartsWithServices

# Test different event types
go test -v -run TestWebhookEventTypesWithServices
```

## Test Behavior

### Service Availability Detection
Tests automatically detect if required services are available by checking health endpoints:
- MaxBot Service: `http://localhost:8095/health`
- Employee Service: `http://localhost:8081/health`
- Redis: Connection test to `localhost:6379`

If services are not available, tests are automatically skipped with appropriate messages.

### Test Data Cleanup
All tests include automatic cleanup of:
- Redis profile cache keys (`profile:user:*`)
- Test employee records from PostgreSQL
- Test university records from PostgreSQL

### Concurrent Testing
Concurrent tests use a reduced number of simultaneous operations (5) for faster execution while still validating concurrent processing capabilities.

## Expected Test Results

### When Services Are Available
```
=== RUN   TestEndToEndWebhookToEmployeeCreationWithServices
=== RUN   TestEndToEndWebhookToEmployeeCreationWithServices/Step1_SendWebhookEvent
=== RUN   TestEndToEndWebhookToEmployeeCreationWithServices/Step2_VerifyProfileInCache
=== RUN   TestEndToEndWebhookToEmployeeCreationWithServices/Step3_CreateEmployeeUsingCache
--- PASS: TestEndToEndWebhookToEmployeeCreationWithServices (2.15s)
```

### When Services Are Not Available
```
=== RUN   TestEndToEndWebhookToEmployeeCreationWithServices
    max_webhook_profile_integration_test.go:687: MaxBot service not available, skipping integration test
--- SKIP: TestEndToEndWebhookToEmployeeCreationWithServices (0.00s)
```

## Troubleshooting

### Common Issues

1. **Services Not Starting**
   - Check Docker Compose logs: `docker-compose logs`
   - Verify port availability: `netstat -tulpn | grep :8095`

2. **Redis Connection Issues**
   - Verify Redis is running: `redis-cli ping`
   - Check Redis configuration in docker-compose.yml

3. **Database Connection Issues**
   - Check PostgreSQL containers: `docker ps | grep postgres`
   - Verify database migrations have run

4. **Test Timeouts**
   - Increase timeout values in test configuration
   - Check system resources and Docker performance

### Debug Mode
To run tests with verbose output and debug information:
```bash
go test -v -run "WithServices" -timeout 5m
```

## Integration with CI/CD

These tests are designed to work in CI/CD environments:

1. **Local Development**: Tests skip when services aren't available
2. **CI Pipeline**: Tests run against Docker Compose environment
3. **Production Validation**: Tests can validate deployed services

### Example CI Configuration
```yaml
- name: Run Integration Tests
  run: |
    docker-compose up -d
    sleep 30
    cd integration-tests
    go test -v -run "WithServices" -timeout 10m
    docker-compose down
```

## Requirements Validation

These integration tests validate the following requirements from the specification:

- **Requirements 1.1-1.5**: Webhook event processing and profile extraction
- **Requirements 2.1-2.4**: User input processing and name prioritization  
- **Requirements 3.1-3.5**: Profile caching with TTL and error handling
- **Requirements 4.1-4.5**: Webhook endpoint functionality and event handling
- **Requirements 5.1-5.5**: Profile source tracking and data integrity
- **Requirements 7.1-7.5**: Backward compatibility and graceful degradation

Each test includes specific assertions that map back to these requirements, ensuring comprehensive validation of the MAX webhook profile integration feature.