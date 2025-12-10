# Token Cleanup Job Implementation

## Overview

This document describes the implementation of the automated token cleanup job for expired password reset tokens, as specified in task 14 of the secure password management feature.

## Implementation Details

### 1. Token Cleanup Job

**Location**: `auth-service/internal/infrastructure/cleanup/token_cleanup.go`

The `TokenCleanupJob` is a background service that periodically removes expired password reset tokens from the database.

**Key Features**:
- Runs on a configurable interval (default: 60 minutes)
- Executes cleanup immediately on startup
- Gracefully handles errors without crashing
- Can be stopped via context cancellation or explicit Stop() call
- Logs all cleanup operations

**Configuration**:
- `TOKEN_CLEANUP_INTERVAL`: Environment variable to set cleanup interval in minutes (default: 60)

### 2. Integration

The cleanup job is integrated into the main application startup in `cmd/auth/main.go`:

```go
// Token cleanup job
cleanupInterval := time.Duration(cfg.TokenCleanupInterval) * time.Minute
cleanupLogger := log.New(os.Stdout, "[CLEANUP] ", log.LstdFlags)
cleanupJob := cleanup.NewTokenCleanupJob(passwordResetRepo, cleanupInterval, cleanupLogger)
ctx := context.Background()

// Start cleanup job in background
go cleanupJob.Start(ctx)
```

### 3. DeleteExpired Method

The `DeleteExpired()` method was already implemented in the `PasswordResetRepository` interface and `PasswordResetPostgres` implementation:

**Location**: `auth-service/internal/infrastructure/repository/password_reset_postgres.go`

```go
func (r *PasswordResetPostgres) DeleteExpired() error {
    query := `
        DELETE FROM password_reset_tokens
        WHERE expires_at < $1
    `
    _, err := r.db.Exec(query, time.Now())
    return err
}
```

This method efficiently removes all tokens where the expiration time is in the past.

### 4. Unit Tests

**Location**: `auth-service/internal/infrastructure/cleanup/token_cleanup_test.go`

Comprehensive unit tests cover:
- ✅ Periodic execution of cleanup
- ✅ Context cancellation handling
- ✅ Explicit Stop() call handling
- ✅ Successful cleanup operations
- ✅ Error handling (logs errors but doesn't crash)
- ✅ Proper initialization

**Test Results**: All tests passing

### 5. Configuration Validation

Added validation in `config.go` to ensure:
- `TOKEN_CLEANUP_INTERVAL` must be at least 1 minute
- Invalid values cause the application to fail fast at startup

## Usage

### Environment Variables

```bash
# Set cleanup interval to 30 minutes
TOKEN_CLEANUP_INTERVAL=30

# Use default (60 minutes)
# TOKEN_CLEANUP_INTERVAL not set
```

### Logging

The cleanup job logs all operations:

```
[CLEANUP] 2025/12/09 22:52:51 Token cleanup job started (interval: 1h0m0s)
[CLEANUP] 2025/12/09 22:52:51 Running token cleanup...
[CLEANUP] 2025/12/09 22:52:51 Token cleanup completed successfully
```

In case of errors:
```
[CLEANUP] 2025/12/09 22:52:51 ERROR: Token cleanup failed: <error details>
```

## Testing

### Run Unit Tests

```bash
cd auth-service
go test ./internal/infrastructure/cleanup/... -v
```

### Run Repository Tests (requires database)

```bash
cd auth-service
export TEST_DATABASE_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
go test ./internal/infrastructure/repository/... -run TestPasswordResetPostgres_DeleteExpired -v
```

## Requirements Validation

This implementation satisfies **Requirement 3.5**:
> WHEN a reset token is used or expires THEN the system SHALL invalidate the token to prevent reuse

The cleanup job ensures that expired tokens are automatically removed from the database, preventing:
- Database bloat from accumulating expired tokens
- Potential security issues from old tokens remaining in the system
- Performance degradation from large token tables

## Performance Considerations

- The cleanup operation uses an indexed query on `expires_at` for efficient deletion
- Cleanup runs in a separate goroutine to avoid blocking the main application
- Configurable interval allows tuning based on token creation rate
- Error handling ensures one failed cleanup doesn't prevent future cleanups

## Future Enhancements

Potential improvements:
- Add metrics for number of tokens deleted per cleanup
- Add alerting if cleanup consistently fails
- Consider batch deletion for very large token tables
- Add manual cleanup endpoint for administrative purposes
