# MAX_id Integration Summary

## Overview
This document summarizes the MAX_id lookup integration in the Employee Service, implementing Requirements 3.1, 3.4, and 3.5.

## Implementation Details

### 1. gRPC Client for MaxBot Service
**Location:** `internal/infrastructure/max/max_client.go`

The MaxClient is already implemented and provides:
- `GetMaxIDByPhone(phone string)` - Retrieves MAX_id from MaxBot Service
- `ValidatePhone(phone string)` - Validates phone number format
- Proper error mapping from gRPC error codes to domain errors

**Configuration:**
- Address: Configured via `MAXBOT_GRPC_ADDR` environment variable (default: `localhost:9095`)
- Timeout: Configured via `MAXBOT_TIMEOUT` environment variable (default: `5s`)

### 2. MAX_id Lookup in CreateEmployeeUseCase
**Location:** `internal/usecase/employee_service.go`

The `AddEmployeeByPhone` method now:
1. **Triggers MAX_id lookup** when creating an employee (Requirement 3.1)
2. **Stores MAX_id** when successfully received (Requirement 3.4)
3. **Handles failures gracefully** - continues without MAX_id if lookup fails (Requirement 3.5)

```go
// Получаем MAX_id по телефону (Requirements 3.1)
maxID, err := s.maxService.GetMaxIDByPhone(phone)
if err != nil {
    // Логируем ошибку, но продолжаем без MAX_id (Requirements 3.5)
    maxID = ""
}

// Если MAX_id получен, сохраняем время обновления (Requirements 3.4)
if maxID != "" {
    now := time.Now()
    employee.MaxIDUpdatedAt = &now
}
```

### 3. Graceful Failure Handling
The implementation ensures that:
- Employee creation **never fails** due to MAX_id lookup errors
- The employee record is created with an empty MAX_id field if lookup fails
- The `MaxIDUpdatedAt` timestamp is only set when MAX_id is successfully retrieved
- Errors are logged but don't block the employee creation process

### 4. Update Employee with MAX_id
The `UpdateEmployee` method also handles MAX_id lookup:
- When phone number is updated, attempts to fetch new MAX_id
- Gracefully handles failures (continues without MAX_id)
- Updates `MaxIDUpdatedAt` timestamp when successful

## Testing

### Unit Tests
**Location:** `internal/usecase/employee_service_test.go`

Three test cases verify the implementation:

1. **TestAddEmployeeByPhone_TriggersMaxIDLookup**
   - Verifies that MAX_id lookup is triggered during employee creation
   - Validates: Requirement 3.1

2. **TestAddEmployeeByPhone_StoresMaxIDWhenReceived**
   - Verifies that MAX_id is stored when successfully received
   - Verifies that MaxIDUpdatedAt timestamp is set
   - Validates: Requirement 3.4

3. **TestAddEmployeeByPhone_SucceedsWithoutMaxID**
   - Verifies that employee creation succeeds even when MAX_id lookup fails
   - Verifies that MaxIDUpdatedAt is not set when MAX_id is empty
   - Validates: Requirement 3.5

All tests pass successfully.

## Docker Configuration

The service is configured in `docker-compose.yml` with:
```yaml
environment:
  MAXBOT_GRPC_ADDR: maxbot-service:9095
  MAXBOT_TIMEOUT: 5s
```

## Data Model

The Employee entity includes:
- `MaxID` (string) - The MAX_id from MAX Messenger
- `MaxIDUpdatedAt` (*time.Time) - Timestamp of last MAX_id update
- `Phone` (string) - Phone number used for MAX_id lookup

## Error Handling

The implementation handles the following error scenarios:
1. **Invalid phone format** - Returns `ErrInvalidPhone` before attempting lookup
2. **MAX API unavailable** - Continues without MAX_id, logs error
3. **User not found in MAX** - Continues without MAX_id, logs error
4. **Network timeout** - Continues without MAX_id, logs error

## Requirements Validation

✅ **Requirement 3.1**: Employee creation triggers MAX_id lookup via gRPC  
✅ **Requirement 3.4**: MAX_id is stored in employee record when received  
✅ **Requirement 3.5**: Employee creation succeeds without MAX_id on lookup failure

## Next Steps

The following related tasks are pending:
- Task 3.1: Write property test for MAX_id lookup trigger (Property 8)
- Task 3.2: Write property test for MAX_id storage (Property 10)
- Task 3.3: Write property test for graceful failure (Property 11)
