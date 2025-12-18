# Bot Information Endpoint

## Overview

The auth-service now includes a `/bot/me` endpoint that provides information about the MaxBot, including the bot name and a link to add the bot.

## Endpoint Details

### GET /bot/me

Retrieves bot information from the MaxBot service.

**Request:**
```http
GET /bot/me
Content-Type: application/json
```

**Response (Success - 200 OK):**
```json
{
  "name": "MAX Bot",
  "add_link": "https://max.ru/add-bot"
}
```

**Response (Error - 500 Internal Server Error):**
```json
{
  "error": "EXTERNAL_SERVICE_ERROR",
  "message": "MaxBot service error"
}
```

## Configuration

The endpoint requires the following environment variables:

- `MAXBOT_SERVICE_ADDR`: Address of the MaxBot gRPC service (e.g., `localhost:9095`)

If `MAXBOT_SERVICE_ADDR` is not configured, the service will use a mock client that returns:
```json
{
  "name": "Digital University Bot",
  "add_link": "https://max.ru/bot/digital_university_bot"
}
```

When connected to a real MaxBot service, it will return actual bot information from MAX API:
```json
{
  "name": "Your Real Bot Name",
  "add_link": "https://max.ru/bot/your_bot_username"
}
```

## Architecture

The endpoint follows the clean architecture pattern:

1. **HTTP Handler** (`internal/infrastructure/http/handler.go`): Handles HTTP requests
2. **Use Case** (`internal/usecase/auth_service.go`): Business logic layer
3. **Domain Interface** (`internal/domain/maxbot_client.go`): Defines the contract
4. **Infrastructure** (`internal/infrastructure/maxbot/client.go`): gRPC client implementation

## Testing

Run the tests with:
```bash
go test ./internal/infrastructure/http/ -v -run TestHandler_GetBotMe
```

## Swagger Documentation

The endpoint is automatically documented in Swagger and available at:
```
http://localhost:8080/swagger/index.html
```

## Example Usage

### cURL
```bash
curl -X GET http://localhost:8080/bot/me \
  -H "Content-Type: application/json"
```

### JavaScript (fetch)
```javascript
fetch('/bot/me')
  .then(response => response.json())
  .then(data => {
    console.log('Bot Name:', data.name);
    console.log('Add Bot Link:', data.add_link);
  });
```

### Go Client
```go
resp, err := http.Get("http://localhost:8080/bot/me")
if err != nil {
    log.Fatal(err)
}
defer resp.Body.Close()

var botInfo struct {
    Name    string `json:"name"`
    AddLink string `json:"add_link"`
}

json.NewDecoder(resp.Body).Decode(&botInfo)
fmt.Printf("Bot: %s, Link: %s\n", botInfo.Name, botInfo.AddLink)
```

## Error Handling

The endpoint handles the following error scenarios:

1. **MaxBot Service Unavailable**: Returns 500 with appropriate error message
2. **Invalid Response**: Returns 500 if MaxBot service returns invalid data
3. **Network Timeout**: Returns 500 if MaxBot service doesn't respond within timeout

## Dependencies

- MaxBot service must be running and accessible via gRPC
- Protobuf definitions from `maxbot-service/api/proto`
- gRPC client libraries