# MaxBot Service

MaxBot Service is a gRPC microservice that provides an interface for interacting with the Max Messenger Bot API. It enables user lookup by phone number and other bot operations through integration with the official [max-bot-api-client-go](https://github.com/max-messenger/max-bot-api-client-go) library.

## Features

- **User Lookup**: Retrieve Max Messenger user IDs by phone number
- **Phone Validation**: Automatic normalization and validation of Russian phone numbers
- **Max API Integration**: Full integration with Max Messenger Bot API using the official client library
- **gRPC Interface**: Clean gRPC API for service-to-service communication
- **Error Handling**: Comprehensive error mapping and descriptive error messages

## Architecture

The service follows clean architecture principles with three main layers:

```
┌─────────────────────────────────────────────────────────────┐
│                        gRPC Handler                          │
│                  (infrastructure/grpc)                       │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                     MaxBotService                            │
│                      (usecase)                               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│                  MaxAPIClient Interface                      │
│                      (domain)                                │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│          Real Max API Client Implementation                  │
│              (infrastructure/maxapi)                         │
│         (wraps max-bot-api-client-go library)                │
└─────────────────────────────────────────────────────────────┘
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Access to Max Messenger Bot API
- Max Bot authentication token

### Installation

1. Clone the repository and navigate to the maxbot-service directory:

```bash
cd maxbot-service
```

2. Install dependencies:

```bash
go mod download
```

3. Generate gRPC code from proto files:

```bash
# From the project root
./generate_proto.sh
```

## Configuration

The service is configured through environment variables. All configuration options are listed below:

### Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `MAX_API_TOKEN` | Bot authentication token for Max Messenger API | `your-bot-token-here` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MAX_API_URL` | Base URL for Max Messenger API | `https://api.max.ru` |
| `MAX_API_TIMEOUT` | Timeout for API requests | `5s` |
| `GRPC_PORT` | Port for gRPC server | `9095` |

### Configuration Examples

#### Development Environment

```bash
export MAX_API_TOKEN="your-dev-bot-token"
export MAX_API_URL="https://api-dev.max.ru"
export MAX_API_TIMEOUT="10s"
export GRPC_PORT="9095"
```

#### Production Environment

```bash
export MAX_API_TOKEN="your-prod-bot-token"
export MAX_API_URL="https://api.max.ru"
export MAX_API_TIMEOUT="5s"
export GRPC_PORT="9095"
```

#### Docker Compose

```yaml
services:
  maxbot-service:
    image: maxbot-service:latest
    environment:
      - MAX_API_TOKEN=your-bot-token
      - MAX_API_URL=https://api.max.ru
      - MAX_API_TIMEOUT=5s
      - GRPC_PORT=9095
    ports:
      - "9095:9095"
```

#### Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: maxbot-config
data:
  MAX_API_URL: "https://api.max.ru"
  MAX_API_TIMEOUT: "5s"
  GRPC_PORT: "9095"
---
apiVersion: v1
kind: Secret
metadata:
  name: maxbot-secrets
type: Opaque
stringData:
  MAX_API_TOKEN: "your-bot-token"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: maxbot-service
spec:
  template:
    spec:
      containers:
      - name: maxbot
        image: maxbot-service:latest
        envFrom:
        - configMapRef:
            name: maxbot-config
        - secretRef:
            name: maxbot-secrets
```

## Running the Service

### Local Development

```bash
# Set required environment variables
export MAX_API_TOKEN="your-bot-token"

# Run the service
go run cmd/maxbot/main.go
```

### Docker

```bash
# Build the Docker image
docker build -t maxbot-service:latest .

# Run the container
docker run -p 9095:9095 \
  -e MAX_API_TOKEN="your-bot-token" \
  -e MAX_API_URL="https://api.max.ru" \
  maxbot-service:latest
```

### Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f maxbot-service

# Stop the service
docker-compose down
```

## API Reference

### gRPC Methods

The service exposes the following gRPC methods defined in `api/proto/maxbot.proto`:

#### GetMaxIDByPhone

Retrieves a Max Messenger user ID by phone number.

**Request:**
```protobuf
message GetMaxIDByPhoneRequest {
  string phone = 1;  // Phone number in any format
}
```

**Response:**
```protobuf
message GetMaxIDByPhoneResponse {
  string max_id = 1;  // Max Messenger user ID
}
```

**Error Codes:**
- `ERROR_CODE_INVALID_PHONE`: Phone number format is invalid
- `ERROR_CODE_MAX_ID_NOT_FOUND`: User not found in Max Messenger
- `ERROR_CODE_INTERNAL`: Internal error (authentication, timeout, etc.)

**Example Usage (Go):**

```go
import (
    pb "path/to/maxbot/proto"
    "google.golang.org/grpc"
)

conn, err := grpc.Dial("localhost:9095", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := pb.NewMaxBotServiceClient(conn)

resp, err := client.GetMaxIDByPhone(context.Background(), &pb.GetMaxIDByPhoneRequest{
    Phone: "+7 (999) 123-45-67",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Max ID: %s\n", resp.MaxId)
```

## Phone Number Validation

The service automatically normalizes and validates phone numbers according to Russian phone number format rules:

### Normalization Rules

1. **Non-digit removal**: All non-digit characters are removed (spaces, dashes, parentheses, etc.)
2. **Eight-to-seven conversion**: 11-digit numbers starting with "8" are converted to start with "7"
3. **Ten-digit prepending**: 10-digit numbers are prepended with "7"
4. **Length validation**: Numbers must be between 10 and 15 digits after normalization

### Examples

| Input | Normalized Output | Valid |
|-------|------------------|-------|
| `+7 (999) 123-45-67` | `79991234567` | ✓ |
| `8 999 123 45 67` | `79991234567` | ✓ |
| `9991234567` | `79991234567` | ✓ |
| `123` | - | ✗ (too short) |
| `12345678901234567` | - | ✗ (too long) |

## Max API Features

### Currently Implemented

- ✅ **User Lookup by Phone**: Get Max Messenger user ID from phone number
- ✅ **Phone Validation**: Normalize and validate phone numbers

### Available for Future Extension

The [max-bot-api-client-go](https://github.com/max-messenger/max-bot-api-client-go) library provides additional capabilities that can be integrated:

- 📨 **Message Sending**: Send text messages to users or chats
- 💬 **Chat Management**: Create and manage group chats
- 📎 **File Operations**: Upload and send files/media
- 🤖 **Bot Commands**: Register and handle bot commands
- 🔔 **Webhooks**: Receive real-time updates from Max Messenger
- 👥 **User Management**: Get user profiles and information
- 📊 **Chat Information**: Retrieve chat details and participant lists

### Extending the Service

To add new Max API features:

1. **Update the domain interface** (`internal/domain/max_api_client.go`):
```go
type MaxAPIClient interface {
    GetMaxIDByPhone(ctx context.Context, phone string) (string, error)
    ValidatePhone(phone string) (bool, string, error)
    // Add new methods here
    SendMessage(ctx context.Context, chatID, message string) error
}
```

2. **Implement in infrastructure layer** (`internal/infrastructure/maxapi/client.go`):
```go
func (c *Client) SendMessage(ctx context.Context, chatID, message string) error {
    // Use max-bot-api-client-go to send message
    return c.client.SendMessage(ctx, chatID, message)
}
```

3. **Add usecase methods** (`internal/usecase/maxbot_service.go`):
```go
func (s *MaxBotService) SendMessage(ctx context.Context, chatID, message string) error {
    return s.apiClient.SendMessage(ctx, chatID, message)
}
```

4. **Update gRPC handler** (`internal/infrastructure/grpc/handler.go`):
```go
func (h *MaxBotHandler) SendMessage(ctx context.Context, req *pb.SendMessageRequest) (*pb.SendMessageResponse, error) {
    err := h.service.SendMessage(ctx, req.ChatId, req.Message)
    // Handle response and errors
}
```

5. **Update proto file** (`api/proto/maxbot.proto`) and regenerate code

## Error Handling

The service provides comprehensive error handling with descriptive messages:

### Error Categories

1. **Validation Errors** (`ERROR_CODE_INVALID_PHONE`)
   - Invalid phone number format
   - Phone number too short or too long
   - Detected before API calls

2. **Not Found Errors** (`ERROR_CODE_MAX_ID_NOT_FOUND`)
   - User not found in Max Messenger
   - Expected error case for non-existent users

3. **Internal Errors** (`ERROR_CODE_INTERNAL`)
   - Authentication failures
   - Rate limiting
   - Network timeouts
   - Unexpected API errors

### Error Logging

The service logs errors with appropriate levels:

- **ERROR**: Authentication failures, unexpected API errors
- **WARN**: Rate limiting, timeouts
- **DEBUG**: Successful API calls (with privacy considerations)

Phone numbers in logs are sanitized to show only the last 4 digits for privacy.

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./internal/infrastructure/maxapi/...
```

### Test Types

The service includes:

- **Unit Tests**: Test individual functions and components
- **Property-Based Tests**: Verify universal properties using gopter
- **Integration Tests**: Test complete flows with Max API

## Security Considerations

1. **Token Security**: Never log or expose the `MAX_API_TOKEN` in error messages or logs
2. **Phone Privacy**: Only last 4 digits of phone numbers are logged
3. **Error Sanitization**: Error messages are sanitized before returning to clients
4. **Rate Limiting**: Respects Max API rate limits to prevent service disruption
5. **Timeout Configuration**: Reasonable timeouts prevent resource exhaustion

## Performance

- **Connection Pooling**: HTTP connections to Max API are reused
- **Default Timeout**: 5 seconds (configurable)
- **Context Propagation**: Proper context handling for cancellation and timeouts
- **Efficient Logging**: Appropriate log levels minimize overhead

## Troubleshooting

### Service Won't Start

**Problem**: Service fails to start with "MAX_API_TOKEN is required"

**Solution**: Ensure the `MAX_API_TOKEN` environment variable is set:
```bash
export MAX_API_TOKEN="your-bot-token"
```

### Authentication Errors

**Problem**: Getting authentication errors from Max API

**Solution**: 
1. Verify your bot token is correct
2. Check that the token has not expired
3. Ensure the token has appropriate permissions

### Timeout Errors

**Problem**: Requests timing out

**Solution**:
1. Increase the timeout: `export MAX_API_TIMEOUT="10s"`
2. Check network connectivity to Max API
3. Verify Max API service status

### User Not Found

**Problem**: Getting `ERROR_CODE_MAX_ID_NOT_FOUND` for valid users

**Solution**:
1. Verify the phone number is registered in Max Messenger
2. Check phone number format is correct
3. Ensure the user has interacted with your bot

## Development

### Project Structure

```
maxbot-service/
├── api/
│   └── proto/              # gRPC protocol definitions
│       ├── maxbot.proto
│       ├── maxbot.pb.go
│       └── maxbot_grpc.pb.go
├── cmd/
│   └── maxbot/
│       └── main.go         # Service entry point
├── internal/
│   ├── config/             # Configuration management
│   │   └── config.go
│   ├── domain/             # Domain models and interfaces
│   │   ├── errors.go
│   │   └── max_api_client.go
│   ├── infrastructure/     # External integrations
│   │   ├── grpc/          # gRPC server and handlers
│   │   │   ├── handler.go
│   │   │   └── server.go
│   │   └── maxapi/        # Max API client implementation
│   │       └── client.go
│   └── usecase/           # Business logic
│       └── maxbot_service.go
├── Dockerfile
├── go.mod
├── go.sum
└── README.md
```

### Adding New Features

See the [Max API Features](#max-api-features) section for guidance on extending the service with additional Max Messenger capabilities.

## License

[Your License Here]

## Support

For issues related to:
- **This service**: Open an issue in this repository
- **Max API**: Refer to [Max Messenger Bot API documentation](https://github.com/max-messenger/max-bot-api-client-go)
- **gRPC**: See [gRPC documentation](https://grpc.io/docs/)

## Contributing

[Your contribution guidelines here]
