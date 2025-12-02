# gRPC Retry Logic Implementation

## Overview

This document describes the implementation of gRPC retry logic across all services in the Digital University MVP system. The retry logic ensures resilience against transient network failures and service unavailability.

## Requirements

- **Requirement 18.4**: Retry gRPC calls up to 3 times with exponential backoff (1s, 2s, 4s)
- **Requirement 18.5**: Log each retry attempt and return error after final failure with appropriate HTTP status

## Implementation

### Retry Package

Each service has a `retry.go` file in its `internal/infrastructure/grpc` package that provides:

1. **RetryConfig**: Configuration for retry behavior
   - `MaxRetries`: Number of retry attempts (default: 3)
   - `Backoff`: Exponential backoff durations (default: 1s, 2s, 4s)

2. **WithRetry**: Main retry wrapper function
   - Wraps any gRPC call with retry logic
   - Logs each retry attempt with operation name
   - Supports context cancellation
   - Returns error after all retries exhausted

3. **IsRetryableError**: Determines if an error should be retried
   - Retries on transient gRPC errors:
     - `Unavailable`: Service temporarily unavailable
     - `DeadlineExceeded`: Request timeout
     - `ResourceExhausted`: Rate limiting or resource constraints
     - `Aborted`: Transaction conflicts
     - `Internal`: Internal server errors
     - `Unknown`: Unknown errors
   - Does NOT retry on client errors:
     - `InvalidArgument`: Bad request data
     - `NotFound`: Resource not found
     - `PermissionDenied`: Authorization failure
     - `Unauthenticated`: Authentication failure

### Services Updated

#### 1. Employee Service
- **MaxClient**: `internal/infrastructure/max/max_client.go`
  - `GetMaxIDByPhone()`: Retry MAX_id lookup
  - `ValidatePhone()`: Retry phone validation

#### 2. Chat Service
- **AuthClient**: `internal/infrastructure/auth/auth_client.go`
  - `ValidateToken()`: Retry token validation
- **MaxClient**: `internal/infrastructure/max/max_client.go`
  - `GetMaxIDByPhone()`: Retry MAX_id lookup
  - `ValidatePhone()`: Retry phone validation

#### 3. Structure Service
- **EmployeeClient**: `internal/infrastructure/employee/employee_client.go`
  - `GetEmployeeByID()`: Retry employee lookup
- **ChatClient**: `internal/infrastructure/grpc/chat_client.go`
  - `GetChatByID()`: Retry chat lookup
  - `CreateChat()`: Retry chat creation

## Usage Example

```go
import (
    grpcretry "employee-service/internal/infrastructure/grpc"
)

func (c *MaxClient) GetMaxIDByPhone(phone string) (string, error) {
    ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
    defer cancel()

    var resp *maxbotproto.GetMaxIDByPhoneResponse
    err := grpcretry.WithRetry(ctx, "MaxBot.GetMaxIDByPhone", func() error {
        var callErr error
        resp, callErr = c.client.GetMaxIDByPhone(ctx, &maxbotproto.GetMaxIDByPhoneRequest{Phone: phone})
        return callErr
    })
    
    if err != nil {
        return "", err
    }

    if resp.Error != "" {
        return "", mapError(resp.ErrorCode, resp.Error)
    }

    return resp.MaxId, nil
}
```

## Retry Behavior

### Success Case
```
[Request] -> [Success] -> Return result
```

### Retry Case
```
[Request] -> [Unavailable Error]
  -> Log: "Retryable error for MaxBot.GetMaxIDByPhone (attempt 1/4)"
  -> Wait 1s
  -> Log: "Attempt 2/4 for MaxBot.GetMaxIDByPhone after 1s backoff"
  -> [Request] -> [Unavailable Error]
  -> Log: "Retryable error for MaxBot.GetMaxIDByPhone (attempt 2/4)"
  -> Wait 2s
  -> Log: "Attempt 3/4 for MaxBot.GetMaxIDByPhone after 2s backoff"
  -> [Request] -> [Success]
  -> Log: "Success for MaxBot.GetMaxIDByPhone after 2 retries"
  -> Return result
```

### Exhausted Retries
```
[Request] -> [Unavailable Error]
  -> Retry 1 (wait 1s) -> [Unavailable Error]
  -> Retry 2 (wait 2s) -> [Unavailable Error]
  -> Retry 3 (wait 4s) -> [Unavailable Error]
  -> Log: "All retries exhausted for MaxBot.GetMaxIDByPhone"
  -> Return error: "gRPC call failed after 3 retries: <original error>"
```

### Non-Retryable Error
```
[Request] -> [InvalidArgument Error]
  -> Log: "Non-retryable error for MaxBot.GetMaxIDByPhone"
  -> Return error immediately (no retries)
```

## Logging

All retry attempts are logged with the following format:

```
[gRPC Retry] Retryable error for <operation> (attempt <N>/<total>): <error>
[gRPC Retry] Attempt <N>/<total> for <operation> after <duration> backoff
[gRPC Retry] Success for <operation> after <N> retries
[gRPC Retry] All retries exhausted for <operation>: <error>
[gRPC Retry] Non-retryable error for <operation>: <error>
```

Example logs:
```
2025/11/28 14:37:58 [gRPC Retry] Retryable error for MaxBot.GetMaxIDByPhone (attempt 1/4): rpc error: code = Unavailable desc = service unavailable
2025/11/28 14:37:58 [gRPC Retry] Attempt 2/4 for MaxBot.GetMaxIDByPhone after 1s backoff
2025/11/28 14:37:59 [gRPC Retry] Success for MaxBot.GetMaxIDByPhone after 1 retries
```

## Testing

### Unit Tests

Each service includes comprehensive unit tests in `internal/infrastructure/grpc/retry_test.go`:

1. **TestWithRetry_Success**: Verifies successful call without retries
2. **TestWithRetry_SuccessAfterRetries**: Verifies success after transient failures
3. **TestWithRetry_NonRetryableError**: Verifies non-retryable errors fail immediately
4. **TestWithRetry_AllRetriesExhausted**: Verifies behavior when all retries fail
5. **TestWithRetry_ContextCancelled**: Verifies context cancellation handling
6. **TestIsRetryableError**: Verifies error classification logic

Run tests:
```bash
cd employee-service
go test ./internal/infrastructure/grpc/... -v
```

### Integration Tests

The retry logic is automatically tested through existing integration tests:
- Employee creation with MAX_id lookup
- Chat filtering with token validation
- Structure hierarchy with chat details

## Error Handling

### HTTP Status Mapping

When gRPC calls fail after all retries, the error is propagated to the HTTP layer with appropriate status codes:

- `Unavailable` → 502 Bad Gateway
- `DeadlineExceeded` → 504 Gateway Timeout
- `ResourceExhausted` → 429 Too Many Requests
- `Internal` → 502 Bad Gateway
- `Unknown` → 502 Bad Gateway

### Error Messages

Error messages include context about the retry attempts:
```
"gRPC call failed after 3 retries: rpc error: code = Unavailable desc = connection refused"
```

## Configuration

### Default Configuration

```go
RetryConfig{
    MaxRetries: 3,
    Backoff:    []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second},
}
```

### Custom Configuration

For specific use cases, you can use `WithRetryConfig`:

```go
config := grpcretry.RetryConfig{
    MaxRetries: 5,
    Backoff:    []time.Duration{500*time.Millisecond, 1*time.Second, 2*time.Second, 4*time.Second, 8*time.Second},
}

err := grpcretry.WithRetryConfig(ctx, "operation", func() error {
    // Your gRPC call
}, config)
```

## Performance Considerations

### Total Retry Time

With default configuration, the maximum retry time is:
- Initial attempt: 0s
- Retry 1: +1s = 1s
- Retry 2: +2s = 3s
- Retry 3: +4s = 7s
- **Total: 7 seconds maximum**

### Context Timeouts

Ensure context timeouts are longer than the maximum retry time:
```go
// Good: 10s timeout allows for retries
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

// Bad: 2s timeout will cancel during retries
ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
```

## Best Practices

1. **Use descriptive operation names**: Helps with debugging logs
   ```go
   WithRetry(ctx, "MaxBot.GetMaxIDByPhone", ...)  // Good
   WithRetry(ctx, "operation", ...)                // Bad
   ```

2. **Set appropriate timeouts**: Context timeout should exceed max retry time
   ```go
   ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
   ```

3. **Handle non-retryable errors**: Don't retry on client errors
   ```go
   // InvalidArgument errors are not retried automatically
   ```

4. **Monitor retry logs**: Track retry frequency to identify service issues
   ```bash
   grep "gRPC Retry" logs/*.log | grep "Retryable error"
   ```

## Future Enhancements

Potential improvements for future iterations:

1. **Circuit Breaker**: Stop retrying if service is consistently down
2. **Jitter**: Add randomness to backoff to prevent thundering herd
3. **Metrics**: Export retry metrics to monitoring system
4. **Configurable per-service**: Different retry configs for different services
5. **Interceptor-based**: Use gRPC interceptors for automatic retry

## Related Documentation

- [GRPC_SETUP.md](GRPC_SETUP.md): gRPC service setup guide
- [Requirements](/.kiro/specs/digital-university-mvp-completion/requirements.md): Requirement 18
- [Design](/.kiro/specs/digital-university-mvp-completion/design.md): Error handling strategy
