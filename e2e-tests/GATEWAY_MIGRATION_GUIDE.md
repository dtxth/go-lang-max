# E2E Tests Gateway Migration Guide

## Overview

The E2E tests have been updated to use the Gateway Service as the single entry point for all microservice communication. This document explains the changes made and how to run the tests.

## Changes Made

### 1. Client Configuration Updates

**File: `e2e-tests/utils/client.go`**

- Updated `DefaultServiceConfigs()` to route all microservice requests through the Gateway Service
- All auth, employee, chat, and structure service requests now go to `http://localhost:8080` (Gateway Service)
- Migration and MaxBot services remain direct connections as they are not part of the Gateway routing

### 2. Test Initialization Updates

**File: `e2e-tests/main_test.go`**

- Modified `TestMain()` to check Gateway Service availability first
- Simplified service availability checks to focus on Gateway Service
- Updated benchmark tests to use Gateway Service endpoints
- Updated load tests to use Gateway Service

### 3. Individual Service Test Updates

**Files: `auth_service_test.go`, `employee_service_test.go`, `chat_service_test.go`, `structure_service_test.go`**

- Updated all service tests to wait for Gateway Service availability instead of individual services
- Maintained all existing test logic and assertions
- All HTTP endpoints remain the same (API contract preserved)

### 4. Gateway Compatibility Tests

**File: `e2e-tests/gateway_compatibility_test.go`**

- Added comprehensive compatibility tests to verify API contract preservation
- Tests request/response format preservation
- Tests status code mapping accuracy
- Validates that Gateway Service maintains backward compatibility

## Service Routing

### Through Gateway Service (Port 8080)
- Auth Service endpoints: `/register`, `/login`, `/health`, `/metrics`, etc.
- Employee Service endpoints: `/employees/all`, `/simple-employee`, etc.
- Chat Service endpoints: `/chats`, `/administrators`, etc.
- Structure Service endpoints: `/universities`, `/structure`, etc.

### Direct Connections (Not through Gateway)
- Migration Service: `http://localhost:8084`
- MaxBot Service: `http://localhost:8095`

## Running the Tests

### Prerequisites

1. **Gateway Service must be running** on port 8080
2. All backend microservices must be running and accessible via gRPC
3. Database services must be available

### Starting Services

```bash
# Start all services using Docker Compose
docker-compose up -d

# Or start Gateway Service specifically
cd gateway-service
go run cmd/gateway/main.go
```

### Running E2E Tests

```bash
# Run all E2E tests
cd e2e-tests
go test -v

# Run specific test suites
go test -v -run TestAuthService
go test -v -run TestEmployeeService
go test -v -run TestChatService
go test -v -run TestStructureService

# Run Gateway compatibility tests
go test -v -run TestGatewayCompatibility

# Run benchmark tests
go test -bench=.

# Run load tests
go test -v -run TestLoadTest
```

## Expected Behavior

### Success Scenario
- Gateway Service is available at `http://localhost:8080`
- All existing E2E tests pass without modification to test logic
- Response formats and status codes remain unchanged
- Performance is maintained or improved due to gRPC backend communication

### Failure Scenarios
- If Gateway Service is not running, tests will fail immediately with clear error message
- If backend microservices are not available, Gateway will return appropriate error responses
- Circuit breaker and retry logic in Gateway will handle transient failures

## Verification

The migration is successful if:

1. ✅ All existing E2E tests pass without changes to test assertions
2. ✅ Response formats match original microservice responses
3. ✅ HTTP status codes are correctly mapped from gRPC status codes
4. ✅ Error responses maintain expected structure
5. ✅ Performance is acceptable (similar or better than direct calls)

## Troubleshooting

### Gateway Service Not Available
```
ERROR: Gateway service should be available
```
**Solution:** Start the Gateway Service on port 8080

### Backend Services Not Available
```
ERROR: gRPC connection failed
```
**Solution:** Ensure all microservices are running and accessible via gRPC

### Test Failures After Migration
1. Check Gateway Service logs for errors
2. Verify gRPC service configurations
3. Test individual endpoints manually using curl or Postman
4. Run Gateway compatibility tests to identify specific issues

## Benefits of Gateway Migration

1. **Single Entry Point:** All HTTP traffic goes through one service
2. **Better Performance:** gRPC communication between services
3. **Centralized Logging:** All HTTP requests logged in one place
4. **Circuit Breaker Protection:** Automatic failure handling
5. **Retry Logic:** Automatic retry for transient failures
6. **Type Safety:** Protocol Buffer validation
7. **Easier Monitoring:** Single service to monitor for HTTP traffic

## Rollback Plan

If issues arise, you can temporarily rollback by:

1. Reverting the changes in `e2e-tests/utils/client.go`
2. Starting individual microservices on their original ports
3. Running tests against direct service endpoints

However, this should only be a temporary measure while fixing Gateway issues.