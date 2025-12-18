# Environment Variables Documentation

This document describes all environment variables used by the chat-service, including participants background sync configuration.

## Required Variables

### DATABASE_URL
- **Description**: PostgreSQL database connection string
- **Format**: `postgres://user:password@host:port/database?sslmode=disable`
- **Example**: `postgres://chat_user:password@localhost:5434/chat_db?sslmode=disable`
- **Required**: Yes

## Core Service Configuration

### PORT
- **Description**: HTTP server port
- **Default**: `8082`
- **Valid Range**: 1-65535
- **Example**: `PORT=8082`

### GRPC_PORT
- **Description**: gRPC server port
- **Default**: `9092`
- **Valid Range**: 1-65535
- **Example**: `GRPC_PORT=9092`

### AUTH_GRPC_ADDR
- **Description**: Auth service gRPC address
- **Default**: `localhost:9090`
- **Format**: `host:port`
- **Example**: `AUTH_GRPC_ADDR=auth-service:9090`

### AUTH_TIMEOUT
- **Description**: Timeout for auth service calls
- **Default**: `5s`
- **Valid Range**: 1s-60s
- **Format**: Go duration string (e.g., "5s", "1m", "500ms")
- **Example**: `AUTH_TIMEOUT=10s`

### MAXBOT_GRPC_ADDR
- **Description**: MaxBot service gRPC address
- **Default**: `localhost:9095`
- **Format**: `host:port`
- **Example**: `MAXBOT_GRPC_ADDR=maxbot-service:9095`

### MAXBOT_TIMEOUT
- **Description**: Timeout for MaxBot service calls
- **Default**: `5s`
- **Valid Range**: 1s-60s
- **Format**: Go duration string
- **Example**: `MAXBOT_TIMEOUT=30s`

### MAX_API_URL
- **Description**: MAX API base URL (optional)
- **Default**: Empty (disabled)
- **Format**: Valid HTTP/HTTPS URL
- **Example**: `MAX_API_URL=https://api.max.com`

### REDIS_URL
- **Description**: Redis connection URL for participants caching
- **Default**: `redis://localhost:6379`
- **Format**: `redis://[user:password@]host:port[/database]` or `rediss://` for TLS
- **Example**: `REDIS_URL=redis://redis:6379/0`

### REDIS_MAX_RETRIES
- **Description**: Maximum number of automatic reconnection attempts for Redis
- **Default**: `5`
- **Valid Range**: 1-20
- **Example**: `REDIS_MAX_RETRIES=10`

### REDIS_RETRY_DELAY
- **Description**: Initial delay between Redis reconnection attempts (with exponential backoff)
- **Default**: `1s`
- **Valid Range**: 100ms-30s
- **Format**: Go duration string
- **Example**: `REDIS_RETRY_DELAY=2s`

### REDIS_HEALTH_CHECK_INTERVAL
- **Description**: Interval for Redis health checks and reconnection monitoring
- **Default**: `30s`
- **Valid Range**: 10s-5m
- **Format**: Go duration string
- **Example**: `REDIS_HEALTH_CHECK_INTERVAL=1m`

## Participants Background Sync Configuration

### PARTICIPANTS_CACHE_TTL
- **Description**: Time-to-live for cached participants data
- **Default**: `1h`
- **Valid Range**: 1m-24h
- **Format**: Go duration string
- **Example**: `PARTICIPANTS_CACHE_TTL=2h`

### PARTICIPANTS_UPDATE_INTERVAL
- **Description**: Interval for background stale data updates
- **Default**: `15m`
- **Valid Range**: 1m-24h
- **Format**: Go duration string
- **Example**: `PARTICIPANTS_UPDATE_INTERVAL=30m`

### PARTICIPANTS_FULL_UPDATE_HOUR
- **Description**: Hour of day (0-23) for full participants update
- **Default**: `3`
- **Valid Range**: 0-23
- **Example**: `PARTICIPANTS_FULL_UPDATE_HOUR=2`

### PARTICIPANTS_BATCH_SIZE
- **Description**: Number of chats to process in each batch
- **Default**: `50`
- **Valid Range**: 1-1000
- **Example**: `PARTICIPANTS_BATCH_SIZE=100`

### PARTICIPANTS_MAX_API_TIMEOUT
- **Description**: Timeout for MAX API calls during participants updates
- **Default**: `30s`
- **Valid Range**: 1s-5m
- **Format**: Go duration string
- **Example**: `PARTICIPANTS_MAX_API_TIMEOUT=45s`

### PARTICIPANTS_STALE_THRESHOLD
- **Description**: Age threshold for considering cached data stale
- **Default**: `1h`
- **Valid Range**: 1m-24h
- **Format**: Go duration string
- **Example**: `PARTICIPANTS_STALE_THRESHOLD=30m`

### PARTICIPANTS_ENABLE_BACKGROUND_SYNC
- **Description**: Enable/disable background participants synchronization
- **Default**: `true`
- **Valid Values**: `true`, `false`, `1`, `0`, `yes`, `no`
- **Example**: `PARTICIPANTS_ENABLE_BACKGROUND_SYNC=false`

### PARTICIPANTS_ENABLE_LAZY_UPDATE
- **Description**: Enable/disable lazy updates during API requests
- **Default**: `true`
- **Valid Values**: `true`, `false`, `1`, `0`, `yes`, `no`
- **Example**: `PARTICIPANTS_ENABLE_LAZY_UPDATE=true`

### PARTICIPANTS_MAX_RETRIES
- **Description**: Maximum number of retries for failed API calls
- **Default**: `3`
- **Valid Range**: 0-10
- **Example**: `PARTICIPANTS_MAX_RETRIES=5`

### PARTICIPANTS_DISABLED
- **Description**: Completely disable participants integration (overrides all other participants settings)
- **Default**: `false`
- **Valid Values**: `true`, `false`, `1`, `0`, `yes`, `no`
- **Example**: `PARTICIPANTS_DISABLED=true`
- **Note**: When set to `true`, all participants-related initialization is skipped

## Configuration Validation

The service performs comprehensive validation of all configuration values with enhanced validation features:

1. **Type Validation**: Ensures values can be parsed as the expected type (duration, integer, boolean, URL)
2. **Range Validation**: Checks that numeric values are within acceptable ranges
3. **Format Validation**: Validates URL formats, address formats, etc.
4. **Consistency Validation**: Warns about potentially problematic configuration combinations
5. **Integration Validation**: Validates Redis connectivity and participants integration requirements
6. **Comprehensive Logging**: Logs configuration summaries and validation results for monitoring

### Enhanced Validation Features

- **Redis URL Validation**: Comprehensive validation of Redis connection strings with scheme and host validation
- **Participants Integration Control**: Ability to completely disable participants integration via `PARTICIPANTS_DISABLED`
- **Configuration Summary Logging**: Detailed logging of loaded configuration for monitoring and debugging
- **Batch Validation**: Ability to validate all configuration parameters at once during startup
- **Graceful Degradation**: Service continues to operate with default values when invalid configuration is provided

### Invalid Configuration Handling

When invalid configuration is detected:
- A warning is logged with details about the invalid value
- The service falls back to the documented default value
- The service continues to start (graceful degradation)
- For required variables (like DATABASE_URL), the service will panic and exit

### Configuration Warnings

The service may log warnings for:
- Invalid format or out-of-range values
- Potentially problematic configuration combinations (e.g., stale threshold smaller than cache TTL)
- Missing optional configuration that may affect functionality

## Docker Compose Example

```yaml
services:
  chat-service:
    environment:
      - DATABASE_URL=postgres://chat_user:password@chat-db:5432/chat_db?sslmode=disable
      - REDIS_URL=redis://redis:6379/0
      - AUTH_GRPC_ADDR=auth-service:9090
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
      - PARTICIPANTS_CACHE_TTL=2h
      - PARTICIPANTS_UPDATE_INTERVAL=30m
      - PARTICIPANTS_BATCH_SIZE=100
      - PARTICIPANTS_ENABLE_BACKGROUND_SYNC=true
```

## Development vs Production

### Development Settings
```bash
export DATABASE_URL="postgres://chat_user:password@localhost:5434/chat_db?sslmode=disable"
export REDIS_URL="redis://localhost:6379/0"
export PARTICIPANTS_UPDATE_INTERVAL="5m"
export PARTICIPANTS_BATCH_SIZE="10"
```

### Production Settings
```bash
export DATABASE_URL="postgres://chat_user:secure_password@db.example.com:5432/chat_db?sslmode=require"
export REDIS_URL="rediss://redis.example.com:6380/0"
export PARTICIPANTS_UPDATE_INTERVAL="15m"
export PARTICIPANTS_BATCH_SIZE="100"
export PARTICIPANTS_MAX_RETRIES="5"
```