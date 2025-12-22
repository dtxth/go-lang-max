# Gateway Service

The Gateway Service acts as a unified API gateway that exposes HTTP endpoints to external clients while communicating with backend microservices via gRPC. This service centralizes all HTTP endpoints and provides features like circuit breaking, retry logic, and health monitoring.

## Features

- **Unified HTTP API**: Single entry point for all microservice endpoints
- **gRPC Communication**: High-performance communication with backend services
- **Circuit Breaker**: Fault tolerance for service failures
- **Retry Logic**: Exponential backoff for transient failures
- **Health Monitoring**: Comprehensive health checks for all services
- **Request Tracing**: Distributed tracing support
- **Error Handling**: Consistent error responses and logging

## Configuration

The Gateway Service is configured via environment variables:

### Server Configuration
- `GATEWAY_PORT`: HTTP server port (default: 8080)
- `GATEWAY_READ_TIMEOUT`: HTTP read timeout (default: 30s)
- `GATEWAY_WRITE_TIMEOUT`: HTTP write timeout (default: 30s)

### Service Addresses
- `AUTH_SERVICE_ADDRESS`: Auth Service gRPC address (default: auth-service:9090)
- `CHAT_SERVICE_ADDRESS`: Chat Service gRPC address (default: chat-service:9092)
- `EMPLOYEE_SERVICE_ADDRESS`: Employee Service gRPC address (default: employee-service:9091)
- `STRUCTURE_SERVICE_ADDRESS`: Structure Service gRPC address (default: structure-service:9093)

### Service Timeouts
- `AUTH_SERVICE_TIMEOUT`: Auth Service timeout (default: 10s)
- `CHAT_SERVICE_TIMEOUT`: Chat Service timeout (default: 10s)
- `EMPLOYEE_SERVICE_TIMEOUT`: Employee Service timeout (default: 10s)
- `STRUCTURE_SERVICE_TIMEOUT`: Structure Service timeout (default: 10s)

### Retry Configuration
- `{SERVICE}_MAX_RETRIES`: Maximum retry attempts (default: 3)
- `{SERVICE}_RETRY_DELAY`: Initial retry delay (default: 100ms)
- `{SERVICE}_MAX_RETRY_DELAY`: Maximum retry delay (default: 5s)
- `{SERVICE}_BACKOFF_MULTIPLIER`: Backoff multiplier (default: 2.0)

### Circuit Breaker Configuration
- `{SERVICE}_CB_MAX_REQUESTS`: Max requests in half-open state (default: 10)
- `{SERVICE}_CB_INTERVAL`: Circuit breaker interval (default: 60s)
- `{SERVICE}_CB_TIMEOUT`: Circuit breaker timeout (default: 60s)

### Logging Configuration
- `LOG_LEVEL`: Logging level (default: info)
- `LOG_FORMAT`: Log format (default: json)

## Health Check

The Gateway Service provides a health check endpoint at `/health` that returns the status of all backend services:

```bash
curl http://localhost:8080/health
```

Response format:
```json
{
  "status": "ok",
  "services": {
    "auth": "healthy",
    "chat": "healthy",
    "employee": "healthy",
    "structure": "healthy"
  }
}
```

If any service is unhealthy, the endpoint returns HTTP 503 with status "degraded".

## Docker Deployment

The Gateway Service is included in the main docker-compose.yml configuration:

```bash
# Start all services including Gateway
docker-compose up -d

# Check Gateway Service logs
docker-compose logs gateway-service

# Check Gateway Service health
curl http://localhost:8080/health
```

## Development

### Local Development

```bash
# Install dependencies
go mod download

# Run the service
go run cmd/gateway/main.go
```

### Building

```bash
# Build binary
go build -o gateway ./cmd/gateway

# Build Docker image
docker build -t gateway-service -f Dockerfile .
```

## API Documentation

The Gateway Service provides comprehensive OpenAPI 3.0 documentation for all endpoints.

### Viewing Documentation

1. **Local Swagger UI**:
   ```bash
   # Start documentation server
   make docs-serve
   
   # Open http://localhost:8082 in your browser
   ```

2. **Online Swagger Editor**:
   - Visit https://editor.swagger.io/
   - Copy content from `docs/swagger.yaml`

3. **Documentation Files**:
   - `docs/swagger.yaml` - OpenAPI 3.0 specification
   - `docs/index.html` - Swagger UI interface
   - `docs/README.md` - Documentation guide

### Validation

```bash
# Validate OpenAPI specification
make docs-validate
```

## API Endpoints

The Gateway Service exposes all endpoints from the backend microservices:

### Auth Service Endpoints
- `POST /register` - User registration
- `POST /login` - User login
- `POST /login-phone` - Phone-based login
- `POST /refresh` - Token refresh
- `POST /logout` - User logout
- `POST /auth/max` - MAX authentication
- `POST /auth/password-reset/request` - Request password reset
- `POST /auth/password-reset/reset` - Reset password
- `POST /auth/password/change` - Change password
- `GET /bot/me` - Get bot information
- `GET /metrics` - Service metrics

### Chat Service Endpoints
- `GET /chats` - List all chats
- `POST /chats` - Create chat
- `GET /chats/{id}` - Get chat by ID
- `GET /chats/search` - Search chats
- `GET /administrators` - List administrators
- `POST /administrators` - Add administrator
- `GET /administrators/{id}` - Get administrator by ID
- `DELETE /administrators/{id}` - Remove administrator
- `POST /chats/refresh-participants` - Refresh participants count

### Employee Service Endpoints
- `GET /employees/all` - List all employees
- `GET /employees/search` - Search employees
- `POST /employees/{id}` - Create employee
- `GET /employees/{id}` - Get employee by ID
- `PUT /employees/{id}` - Update employee
- `DELETE /employees/{id}` - Delete employee
- `POST /employees/batch-update-maxid` - Batch update MAX IDs
- `GET /employees/batch-status` - Get batch status
- `GET /employees/batch-status/{id}` - Get batch status by ID
- `POST /simple-employee` - Create simple employee

### Structure Service Endpoints
- `GET /universities` - List universities
- `POST /universities` - Create university
- `GET /universities/{id}` - Get university by ID
- `GET /universities/{id}/structure` - Get university structure
- `PUT /universities/{id}/name` - Update university name
- `POST /structure` - Create structure
- `POST /import/excel` - Import Excel data
- `GET /departments/managers` - List department managers
- `POST /departments/managers` - Create department manager
- `DELETE /departments/managers/{id}` - Remove department manager

### System Endpoints
- `GET /health` - Health check

## Architecture

The Gateway Service follows a layered architecture:

```
HTTP Layer (Router/Handlers)
    ↓
gRPC Client Layer (Client Manager)
    ↓
Infrastructure Layer (Circuit Breaker, Retry, Error Handling)
    ↓
Backend Services (Auth, Chat, Employee, Structure)
```

### Key Components

- **Router**: HTTP request routing and middleware
- **Handlers**: HTTP request/response processing
- **Client Manager**: gRPC client connection management
- **Circuit Breaker**: Fault tolerance implementation
- **Retry Logic**: Exponential backoff retry mechanism
- **Error Handler**: Consistent error response formatting
- **Service Registry**: Service discovery and health monitoring

## Monitoring

The Gateway Service provides comprehensive monitoring:

- **Health Checks**: Continuous monitoring of backend services
- **Request Tracing**: Distributed tracing with request IDs
- **Error Logging**: Structured error logging with context
- **Metrics**: Service performance and availability metrics
- **Circuit Breaker Status**: Real-time circuit breaker state

## Error Handling

The Gateway Service implements robust error handling:

- **gRPC to HTTP Status Mapping**: Consistent status code translation
- **Circuit Breaker**: Fast failure for unhealthy services
- **Retry Logic**: Automatic retry for transient failures
- **Error Logging**: Detailed error information for debugging
- **Graceful Degradation**: Partial functionality during service outages