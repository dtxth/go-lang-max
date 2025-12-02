# Comprehensive Error Handling Implementation

## Overview

This document describes the comprehensive error handling system implemented across all microservices in the Digital University MVP project. The implementation provides consistent error responses, detailed error codes, request tracking, and structured logging.

## Architecture

### Error Package Structure

Each service now has an `internal/infrastructure/errors` package that provides:

1. **Structured Error Types** - `AppError` with error codes, messages, details, and HTTP status codes
2. **Error Response Format** - Consistent JSON error responses across all services
3. **Error Constructors** - Helper functions for common error types
4. **Error Logging** - Structured logging with request context

### Request ID Middleware

Each service has a `internal/infrastructure/middleware` package that provides:

1. **Request ID Generation** - Unique ID for each request
2. **Request Logging** - Log request start and completion with duration
3. **Context Propagation** - Request ID available throughout request lifecycle

## Error Categories and Codes

### Validation Errors (400 Bad Request)

- `VALIDATION_ERROR` - General validation error
- `INVALID_PHONE` - Invalid phone number format
- `MISSING_FIELD` - Required field is missing
- `INVALID_FORMAT` - Invalid data format
- `INVALID_RANGE` - Value out of valid range

### Authentication Errors (401 Unauthorized)

- `UNAUTHORIZED` - General unauthorized access
- `INVALID_TOKEN` - Invalid or malformed token
- `EXPIRED_TOKEN` - Token has expired
- `MISSING_TOKEN` - Authorization token is missing
- `INVALID_CREDENTIALS` - Invalid email or password

### Authorization Errors (403 Forbidden)

- `FORBIDDEN` - General forbidden access
- `INSUFFICIENT_PERMISSIONS` - User lacks required permissions
- `INVALID_ROLE` - Invalid role for operation

### Not Found Errors (404 Not Found)

- `NOT_FOUND` - General resource not found
- `USER_NOT_FOUND` - User not found
- `RESOURCE_NOT_FOUND` - Specific resource not found

### Conflict Errors (409 Conflict)

- `CONFLICT` - General conflict
- `ALREADY_EXISTS` - Resource already exists
- `CANNOT_DELETE` - Resource cannot be deleted

### External Service Errors (502 Bad Gateway)

- `EXTERNAL_SERVICE_ERROR` - External service error
- `SERVICE_UNAVAILABLE` - Service is unavailable
- `GRPC_ERROR` - gRPC call failed

### Internal Errors (500 Internal Server Error)

- `INTERNAL_ERROR` - General internal error
- `DATABASE_ERROR` - Database operation failed
- `TRANSACTION_ERROR` - Transaction failed

## Error Response Format

All errors follow this consistent JSON structure:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid phone format",
    "details": {
      "field": "phone",
      "value": "123",
      "expected": "E.164 format (+7XXXXXXXXXX)"
    }
  }
}
```

## Usage Examples

### In HTTP Handlers

```go
func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
    requestID := middleware.GetRequestID(r.Context())
    
    var req CreateEmployeeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        errors.WriteError(w, errors.ValidationError("invalid request body").WithError(err), requestID)
        return
    }
    
    if req.Phone == "" {
        errors.WriteError(w, errors.MissingFieldError("phone"), requestID)
        return
    }
    
    employee, err := h.employeeService.CreateEmployee(ctx, req)
    if err != nil {
        errors.WriteError(w, err, requestID)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(employee)
}
```

### In Domain Layer

```go
// In domain/errors.go
var (
    ErrEmployeeNotFound = errors.NotFoundError("employee")
    ErrInvalidPhone     = errors.InvalidPhoneError("")
)

// In use case
func (uc *EmployeeService) GetEmployee(id int64) (*Employee, error) {
    employee, err := uc.repo.GetByID(id)
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, domain.ErrEmployeeNotFound
        }
        return nil, errors.DatabaseError("GetByID", err)
    }
    return employee, nil
}
```

### Creating Custom Errors

```go
// Simple error
err := errors.ValidationError("invalid input")

// Error with details
err := errors.InvalidPhoneError(phone).
    WithDetails("expected", "E.164 format")

// Error with underlying error
err := errors.DatabaseError("insert", dbErr).
    WithDetails("table", "employees")

// gRPC error
err := errors.GRPCError("AuthService", "ValidateToken", grpcErr)
```

## Request ID Tracking

Every request gets a unique request ID that is:

1. Generated automatically or extracted from `X-Request-ID` header
2. Added to response headers as `X-Request-ID`
3. Available in request context via `middleware.GetRequestID(ctx)`
4. Logged with every error and request log

### Log Format

```
[INFO] [abc123def456] GET /employees - Started
[ERROR] [abc123def456] Code: VALIDATION_ERROR, Message: Missing required field: phone, Details: {"field":"phone"}
[INFO] [abc123def456] GET /employees - Completed in 45ms
```

## Implementation Checklist

### ✅ Completed

1. **Error Package** - Created `internal/infrastructure/errors` for all services
2. **Middleware Package** - Created `internal/infrastructure/middleware` for all services
3. **Domain Errors** - Updated all domain error definitions to use structured errors
4. **HTTP Handlers** - Updated key handlers in all services to use new error handling
5. **Routers** - Added request ID middleware to all service routers
6. **Error Codes** - Defined comprehensive error code taxonomy
7. **Error Constructors** - Created helper functions for common error types
8. **Logging** - Implemented structured error logging with request context

### Services Updated

- ✅ Auth Service
- ✅ Employee Service
- ✅ Chat Service
- ✅ Structure Service
- ✅ MaxBot Service
- ✅ Migration Service

## Benefits

1. **Consistency** - All services return errors in the same format
2. **Debuggability** - Request IDs enable tracing requests across services
3. **Client-Friendly** - Error codes allow clients to handle errors programmatically
4. **Detailed Context** - Error details provide specific information about failures
5. **Structured Logging** - JSON-formatted logs with request context
6. **Maintainability** - Centralized error handling logic

## Testing Error Handling

### Example Test

```go
func TestInvalidPhoneError(t *testing.T) {
    w := httptest.NewRecorder()
    req := httptest.NewRequest("POST", "/employees", strings.NewReader(`{"phone":"invalid"}`))
    
    handler.CreateEmployee(w, req)
    
    assert.Equal(t, http.StatusBadRequest, w.Code)
    
    var response errors.ErrorResponse
    json.Unmarshal(w.Body.Bytes(), &response)
    
    assert.Equal(t, errors.ErrCodeInvalidPhone, response.Error.Code)
    assert.Contains(t, response.Error.Message, "Invalid phone format")
}
```

## Migration Notes

### For Existing Code

When updating existing handlers:

1. Add imports for `errors` and `middleware` packages
2. Get request ID at start of handler: `requestID := middleware.GetRequestID(r.Context())`
3. Replace `http.Error()` calls with `errors.WriteError(w, err, requestID)`
4. Use appropriate error constructors instead of `errors.New()`
5. Add field validation with `MissingFieldError()` for required fields

### For New Code

1. Always use structured errors from the errors package
2. Always get request ID from context
3. Add meaningful details to errors when possible
4. Wrap underlying errors with `WithError()`
5. Use appropriate HTTP status codes via error constructors

## Future Enhancements

1. **Error Metrics** - Track error rates by code and endpoint
2. **Error Alerting** - Alert on high error rates or critical errors
3. **Error Recovery** - Automatic retry for transient errors
4. **Error Translation** - Multi-language error messages
5. **Error Documentation** - Auto-generate error code documentation

## Related Requirements

This implementation satisfies:

- **Requirement 19.5** - Consistent error response format with error codes
- **Requirement 20.1-20.5** - Structured logging with request context
- **Design Document** - Error handling section specifications

## References

- Design Document: Error Handling section
- Requirements Document: Requirement 19.5
- Error Package: `internal/infrastructure/errors/errors.go`
- Middleware Package: `internal/infrastructure/middleware/request_id.go`
