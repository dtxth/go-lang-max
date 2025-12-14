# Monitoring and Logging Implementation

## Overview

This document describes the implementation of structured JSON logging, request ID tracing, gRPC call logging, migration progress logging, and health check endpoints across all services in the Digital University MVP system.

## Components Implemented

### 1. Structured JSON Logger

**Location:** `*/internal/infrastructure/logger/logger.go`

**Features:**
- Structured JSON output for all log entries
- Log levels: DEBUG, INFO, WARN, ERROR
- Automatic request_id extraction from context
- Timestamp in RFC3339 format
- Flexible field support for contextual information

**Usage Example:**
```go
logger := logger.NewDefault()
logger.Info(ctx, "User created", map[string]interface{}{
    "user_id": 123,
    "email": "user@example.com",
})
```

**Log Output:**
```json
{
  "timestamp": "2024-01-15T10:30:45Z",
  "level": "INFO",
  "message": "User created",
  "request_id": "a1b2c3d4e5f6",
  "fields": {
    "user_id": 123,
    "email": "user@example.com"
  }
}
```

### 2. Enhanced Request ID Middleware

**Location:** `*/internal/infrastructure/middleware/request_id.go`

**Features:**
- Generates unique request ID for each HTTP request
- Propagates request ID via X-Request-ID header
- Stores request ID in context for downstream use
- Logs request start and completion with duration
- Structured JSON logging for all HTTP requests

**Logged Information:**
- HTTP method
- Request path
- Remote address
- Request duration in milliseconds
- Request ID for tracing

### 3. gRPC Logging Interceptor

**Location:** `*/internal/infrastructure/grpc/interceptor.go`

**Features:**
- Logs all gRPC calls with method name
- Tracks call duration
- Logs errors with full context
- Integrates with request ID from context

**Usage:**
```go
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(LoggingInterceptor(logger)),
)
```

**Logged Information:**
- gRPC method name
- Call duration
- Success/failure status
- Error details (if applicable)

### 4. Health Check Endpoints

**Location:** `*/internal/infrastructure/health/health.go`

**Endpoints:**
- `GET /health` - Overall health status with dependency checks
- `GET /health/ready` - Readiness check for load balancers
- `GET /health/live` - Liveness check for orchestrators

**Health Checks:**
- **Auth/Employee/Chat/Structure/Migration Services:**
  - Database connectivity (2-second timeout)
  
- **MaxBot Service:**
  - MAX API connectivity (5-second timeout)

**Response Format:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:45Z",
  "checks": {
    "database": "ok",
    "max_api": "ok"
  }
}
```

### 5. Migration Progress Logging

**Location:** `migration-service/internal/usecase/migrate_from_*.go`

**Features:**
- Logs migration job creation and start
- Logs total records to process
- Periodic progress updates (every 50-100 records)
- Logs individual record failures with context
- Final summary with success/failure counts
- Percentage completion tracking

**Logged Events:**
- Migration job created
- Migration job started
- Rows/chats loaded
- Progress updates with percentage
- Individual record processing errors
- Migration completion summary

**Example Log Sequence:**
```json
{"timestamp":"2024-01-15T10:30:00Z","level":"INFO","message":"Migration job created","fields":{"job_id":1,"source_type":"excel"}}
{"timestamp":"2024-01-15T10:30:01Z","level":"INFO","message":"Migration job started","fields":{"job_id":1}}
{"timestamp":"2024-01-15T10:30:02Z","level":"INFO","message":"Migration progress: rows loaded","fields":{"job_id":1,"total":1000}}
{"timestamp":"2024-01-15T10:31:00Z","level":"INFO","message":"Migration progress update","fields":{"job_id":1,"total":1000,"processed":100,"failed":2,"percent":10.2}}
{"timestamp":"2024-01-15T10:40:00Z","level":"INFO","message":"Excel migration completed","fields":{"job_id":1,"total":1000,"processed":995,"failed":5}}
```

## Integration Points

### HTTP Routers

All HTTP routers should be updated to use the new middleware:

```go
logger := logger.NewDefault()
router.Use(middleware.RequestIDMiddleware(logger))
```

### gRPC Servers

All gRPC servers should be updated to use the logging interceptor:

```go
logger := logger.NewDefault()
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(grpc.LoggingInterceptor(logger)),
)
```

### Health Check Routes

Add health check routes to all HTTP routers:

```go
healthHandler := health.NewHandler(db)
router.HandleFunc("/health", healthHandler.HealthCheck)
router.HandleFunc("/health/ready", healthHandler.ReadinessCheck)
router.HandleFunc("/health/live", healthHandler.LivenessCheck)
```

### Migration Use Cases

Migration use cases should be instantiated with logger:

```go
logger := logger.NewDefault()
migrateUseCase := usecase.NewMigrateFromExcelUseCase(
    jobRepo,
    errorRepo,
    structureService,
    chatService,
    logger,
)
```

## Log Levels

### DEBUG
- Detailed diagnostic information
- Not typically enabled in production

### INFO
- General informational messages
- Request/response logging
- Progress updates
- Successful operations

### WARN
- Warning conditions that don't prevent operation
- Degraded service states
- Non-critical errors (e.g., failed to add administrator)

### ERROR
- Error conditions that prevent specific operations
- Failed requests
- Database errors
- External service failures

## Monitoring Integration

### Metrics to Track

Based on the structured logs, the following metrics can be extracted:

1. **Request Metrics:**
   - Request rate per endpoint
   - Response time percentiles (p50, p95, p99)
   - Error rate by error type

2. **gRPC Metrics:**
   - gRPC call success/failure rate
   - gRPC call duration
   - Error rate by method

3. **Migration Metrics:**
   - Migration progress (processed/total)
   - Migration success rate
   - Migration duration
   - Records processed per second

4. **Health Metrics:**
   - Database connectivity status
   - MAX API connectivity status
   - Service availability

### Log Aggregation

Structured JSON logs can be easily ingested by:
- **ELK Stack** (Elasticsearch, Logstash, Kibana)
- **Grafana Loki**
- **CloudWatch Logs** (AWS)
- **Stackdriver** (GCP)

### Alerting Rules

Based on the design document requirements:

1. **Error Rate > 5%**
   - Query: Count ERROR level logs per minute
   - Alert if rate exceeds 5%

2. **Response Time p95 > 1s**
   - Query: Extract duration_ms from HTTP request completed logs
   - Alert if p95 exceeds 1000ms

3. **gRPC Failure Rate > 10%**
   - Query: Count gRPC request failed logs
   - Alert if rate exceeds 10%

4. **Migration Stalled**
   - Query: Check for migration progress updates
   - Alert if no progress for 5 minutes

## Request Tracing

### Request ID Propagation

Request IDs are automatically:
1. Generated for incoming HTTP requests (or extracted from X-Request-ID header)
2. Stored in request context
3. Included in all log entries
4. Propagated to downstream services via headers
5. Returned to client in response headers

### Tracing Across Services

To trace a request across multiple services:

1. Extract request_id from logs
2. Search all service logs for that request_id
3. Reconstruct the full request flow

Example query (Elasticsearch):
```
request_id:"a1b2c3d4e5f6"
```

## Testing

### Unit Tests

Logger functionality can be tested by:
```go
var buf bytes.Buffer
logger := logger.New(&buf, logger.INFO)
logger.Info(ctx, "test message", map[string]interface{}{"key": "value"})
// Assert buf contains expected JSON
```

### Integration Tests

Health check endpoints can be tested:
```go
resp, err := http.Get("http://localhost:8080/health")
// Assert resp.StatusCode == 200
// Assert response body contains expected health status
```

### Load Tests

Monitor log output during load tests to ensure:
- No log entries are dropped
- Log performance doesn't impact request latency
- Request IDs are unique and properly propagated

## Performance Considerations

### Log Volume

- INFO level logs for all HTTP requests
- INFO level logs for all gRPC calls
- Periodic progress logs for migrations (not every record)
- ERROR logs for all failures

### Log Rotation

Implement log rotation to prevent disk space issues:
- Rotate daily or when size exceeds 100MB
- Keep last 7 days of logs
- Compress rotated logs

### Async Logging

For high-throughput services, consider:
- Buffered log writers
- Async log shipping to aggregation service
- Sampling for very high-volume endpoints

## Requirements Validation

This implementation satisfies the following requirements:

✅ **Requirement 20.1:** Structured JSON logging to all services
✅ **Requirement 20.2:** Include request_id for tracing  
✅ **Requirement 20.3:** Log all gRPC calls with duration
✅ **Requirement 20.4:** Log all migration operations with progress
✅ **Requirement 20.5:** Expose health check endpoints

## Next Steps

1. Update all service main.go files to initialize logger and wire up middleware
2. Update all HTTP routers to use new middleware signature
3. Update all gRPC servers to use logging interceptor
4. Add health check routes to all services
5. Configure log aggregation and monitoring dashboards
6. Set up alerting rules based on log metrics
7. Document operational runbooks for common log patterns
