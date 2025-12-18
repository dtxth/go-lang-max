# MaxBot Service

MaxBot Service is a gRPC microservice that provides an interface for interacting with the Max Messenger Bot API. It enables user lookup by phone number and other bot operations through integration with the official [max-bot-api-client-go](https://github.com/max-messenger/max-bot-api-client-go) library.

## Features

- **User Lookup**: Retrieve Max Messenger user IDs by phone number
- **Phone Validation**: Automatic normalization and validation of Russian phone numbers
- **Max API Integration**: Full integration with Max Messenger Bot API using the official client library
- **Webhook Integration**: Process MAX Messenger webhook events for profile collection
- **Profile Caching**: Redis-based caching of user profile information with TTL
- **Profile Management**: REST API for managing user profiles and names
- **Monitoring & Analytics**: Comprehensive monitoring of profile quality and webhook processing
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

### Profile Integration Variables (NEW)

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `REDIS_ADDR` | Redis server address for profile cache | `redis:6379` | `localhost:6379` |
| `REDIS_PASSWORD` | Redis password (if required) | _(empty)_ | `secure-password` |
| `REDIS_DB` | Redis database number for profiles | `1` | `1` |
| `PROFILE_TTL` | Profile cache TTL | `720h` | `168h` |
| `WEBHOOK_SECRET` | Webhook authentication secret | _(empty)_ | `secure-webhook-secret` |
| `MONITORING_ENABLED` | Enable monitoring endpoints | `true` | `false` |
| `PROFILE_QUALITY_ALERT_THRESHOLD` | Profile quality alert threshold | `0.8` | `0.9` |
| `WEBHOOK_ERROR_ALERT_THRESHOLD` | Webhook error rate alert threshold | `0.05` | `0.1` |

### Optional Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `MAX_API_URL` | Base URL for Max Messenger API | `https://api.max.ru` |
| `MAX_API_TIMEOUT` | Timeout for API requests | `5s` |
| `GRPC_PORT` | Port for gRPC server | `9095` |
| `MAXBOT_HTTP_PORT` | Port for HTTP server (webhooks, API) | `8095` |

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

### HTTP Endpoints (NEW)

The service now exposes HTTP endpoints for webhook processing and profile management:

#### Webhook Processing

- `POST /webhook/max` - Process MAX Messenger webhook events
- `GET /health` - Service health check

#### Profile Management

- `GET /profiles/{user_id}` - Get user profile information
- `PUT /profiles/{user_id}` - Update user profile (admin)
- `POST /profiles/{user_id}/name` - Set user-provided name
- `GET /profiles/stats` - Get profile statistics

#### Monitoring

- `GET /monitoring/profiles/coverage` - Profile coverage metrics
- `GET /monitoring/profiles/quality` - Profile quality report
- `GET /monitoring/webhook/stats` - Webhook processing statistics

#### Documentation

- `GET /swagger/` - Swagger UI for HTTP API documentation

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
resp, err := client.GetMaxIDByPhone(context.Background(), &pb.GetMaxIDByPhoneRequest{
    Phone: "+7 (999) 123-45-67",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Max ID: %s\n", resp.MaxId)
```

#### SendMessage

Sends a text message to a user or chat.

**Request:**
```protobuf
message SendMessageRequest {
  oneof recipient {
    int64 chat_id = 1;  // Chat ID (for group chats)
    int64 user_id = 2;  // User ID (for direct messages)
  }
  string text = 3;      // Message text
}
```

**Response:**
```protobuf
message SendMessageResponse {
  string message_id = 1;  // Sent message ID
}
```

**Example Usage (Go):**

```go
// Send to chat
resp, err := client.SendMessage(context.Background(), &pb.SendMessageRequest{
    Recipient: &pb.SendMessageRequest_ChatId{ChatId: 12345},
    Text: "Hello, chat!",
})

// Send to user
resp, err := client.SendMessage(context.Background(), &pb.SendMessageRequest{
    Recipient: &pb.SendMessageRequest_UserId{UserId: 67890},
    Text: "Hello, user!",
})
```

#### SendNotification

Sends a VIP notification to a user by phone number. Notifications are delivered even if the user hasn't started a conversation with the bot.

**Request:**
```protobuf
message SendNotificationRequest {
  string phone = 1;  // Phone number
  string text = 2;   // Notification text
}
```

**Response:**
```protobuf
message SendNotificationResponse {
  bool success = 1;  // Whether notification was sent
}
```

**Example Usage (Go):**

```go
resp, err := client.SendNotification(context.Background(), &pb.SendNotificationRequest{
    Phone: "+79991234567",
    Text: "Important notification!",
})
```

#### GetChatInfo

Retrieves detailed information about a chat.

**Request:**
```protobuf
message GetChatInfoRequest {
  int64 chat_id = 1;
}
```

**Response:**
```protobuf
message GetChatInfoResponse {
  ChatInfo chat = 1;
}

message ChatInfo {
  int64 chat_id = 1;
  string title = 2;
  string type = 3;
  int32 participants_count = 4;
  string description = 5;
}
```

**Example Usage (Go):**

```go
resp, err := client.GetChatInfo(context.Background(), &pb.GetChatInfoRequest{
    ChatId: 12345,
})
fmt.Printf("Chat: %s (%d members)\n", resp.Chat.Title, resp.Chat.ParticipantsCount)
```

#### GetChatMembers

Retrieves a paginated list of chat members.

**Request:**
```protobuf
message GetChatMembersRequest {
  int64 chat_id = 1;
  int32 limit = 2;    // Max 100, default 50
  int64 marker = 3;   // Pagination marker
}
```

**Response:**
```protobuf
message GetChatMembersResponse {
  repeated ChatMember members = 1;
  int64 marker = 2;  // Next page marker
}

message ChatMember {
  int64 user_id = 1;
  string name = 2;
  bool is_admin = 3;
  bool is_owner = 4;
}
```

**Example Usage (Go):**

```go
resp, err := client.GetChatMembers(context.Background(), &pb.GetChatMembersRequest{
    ChatId: 12345,
    Limit: 50,
})
for _, member := range resp.Members {
    fmt.Printf("Member: %s (ID: %d)\n", member.Name, member.UserId)
}
```

#### GetChatAdmins

Retrieves all administrators of a chat.

**Request:**
```protobuf
message GetChatAdminsRequest {
  int64 chat_id = 1;
}
```

**Response:**
```protobuf
message GetChatAdminsResponse {
  repeated ChatMember admins = 1;
}
```

**Example Usage (Go):**

```go
resp, err := client.GetChatAdmins(context.Background(), &pb.GetChatAdminsRequest{
    ChatId: 12345,
})
for _, admin := range resp.Admins {
    fmt.Printf("Admin: %s\n", admin.Name)
}
```

#### CheckPhoneNumbers

Checks which phone numbers exist in Max Messenger (bulk operation).

**Request:**
```protobuf
message CheckPhoneNumbersRequest {
  repeated string phones = 1;
}
```

**Response:**
```protobuf
message CheckPhoneNumbersResponse {
  repeated string existing_phones = 1;  // Phones that exist
}
```

**Example Usage (Go):**

```go
resp, err := client.CheckPhoneNumbers(context.Background(), &pb.CheckPhoneNumbersRequest{
    Phones: []string{"+79991234567", "+79997654321", "+79995555555"},
})
fmt.Printf("Found %d existing phones\n", len(resp.ExistingPhones))
```

#### GetMe

Retrieves information about the bot (name and add bot link).

**Request:**
```protobuf
message GetMeRequest {
  // Empty request
}
```

**Response:**
```protobuf
message GetMeResponse {
  BotInfo bot = 1;
  ErrorCode error_code = 2;
  string error = 3;
}

message BotInfo {
  string name = 1;      // Bot name
  string add_link = 2;  // Link to add the bot
}
```

**Example Usage (Go):**

```go
resp, err := client.GetMe(context.Background(), &pb.GetMeRequest{})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Bot Name: %s\n", resp.Bot.Name)
fmt.Printf("Add Bot Link: %s\n", resp.Bot.AddLink)
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
- ✅ **Message Sending**: Send text messages to users or chats
- ✅ **Notifications**: Send VIP notifications to users by phone number
- ✅ **Chat Information**: Get detailed information about chats
- ✅ **Chat Members**: Retrieve list of chat participants with pagination
- ✅ **Chat Admins**: Get list of chat administrators
- ✅ **Bulk Phone Check**: Check multiple phone numbers for existence in Max Messenger
- ✅ **Bot Information**: Get bot name and add bot link via GetMe method

### Available for Future Extension

The [max-bot-api-client-go](https://github.com/max-messenger/max-bot-api-client-go) library provides additional capabilities that can be integrated:

- �  **File Operations**: Upload and send files/media
- 🤖 **Bot Commands**: Register and handle bot commands
- 🔔 **Webhooks**: Receive real-time updates from Max Messenger
- 👥 **User Management**: Get detailed user profiles and information
- ✏️ **Message Editing**: Edit and delete sent messages
- 🔧 **Chat Management**: Create chats, add/remove members, edit chat settings

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
