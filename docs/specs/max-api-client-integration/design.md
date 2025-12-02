# Design Document

## Overview

This design document describes the integration of the official max-bot-api-client-go library into the maxbot-service. The integration replaces the current stub implementation with a real Max Messenger API client, enabling actual bot operations while maintaining backward compatibility with existing service interfaces.

The design follows the existing clean architecture pattern with domain, usecase, and infrastructure layers. The Max API client will be implemented as an infrastructure component that satisfies the domain interface contract.

## Architecture

### Current Architecture

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
│              Stub Client Implementation                      │
│              (infrastructure/maxapi)                         │
│         (only phone normalization, no API calls)             │
└─────────────────────────────────────────────────────────────┘
```

### Target Architecture

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
│                                                              │
│  ┌────────────────────────────────────────────────┐         │
│  │     max-bot-api-client-go Library              │         │
│  │  (github.com/max-messenger/                    │         │
│  │   max-bot-api-client-go)                       │         │
│  └────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

### Key Design Decisions

1. **Maintain Domain Interface**: The existing `MaxAPIClient` interface remains unchanged to preserve compatibility with the usecase layer
2. **Wrapper Pattern**: The infrastructure implementation wraps the official client library rather than exposing it directly to maintain clean architecture boundaries
3. **Phone Validation**: Keep the existing phone normalization logic in the infrastructure layer before making API calls
4. **Error Mapping**: Map Max API errors to domain errors for consistent error handling across the service
5. **Configuration**: Use environment variables for all client configuration (token, URL, timeout)

## Components and Interfaces

### Domain Layer

#### MaxAPIClient Interface (unchanged)

```go
package domain

import "context"

type MaxAPIClient interface {
    GetMaxIDByPhone(ctx context.Context, phone string) (string, error)
    ValidatePhone(phone string) (bool, string, error)
}
```

#### Domain Errors (unchanged)

```go
package domain

import "errors"

var (
    ErrInvalidPhone  = errors.New("invalid phone number")
    ErrMaxIDNotFound = errors.New("max id not found")
)
```

### Infrastructure Layer

#### Max API Client Implementation

```go
package maxapi

import (
    "context"
    "regexp"
    "strings"
    
    maxapi "github.com/max-messenger/max-bot-api-client-go"
    "maxbot-service/internal/domain"
)

type Client struct {
    client *maxapi.Client
}

func NewClient(token string, baseURL string) (*Client, error) {
    // Initialize the official Max API client
    // Handle configuration and validation
}

func (c *Client) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
    // 1. Validate and normalize phone
    // 2. Call Max API to get user by phone
    // 3. Extract Max ID from response
    // 4. Map errors to domain errors
}

func (c *Client) ValidatePhone(phone string) (bool, string, error) {
    // Keep existing phone normalization logic
}

// Private helper methods
func (c *Client) normalizePhone(phone string) string
func (c *Client) mapAPIError(err error) error
```

### Configuration

```go
package config

import (
    "os"
    "time"
)

type Config struct {
    GRPCPort       string
    MaxAPIURL      string        // Base URL for Max API
    MaxAPIToken    string        // Bot authentication token
    RequestTimeout time.Duration // Timeout for API requests
}

func Load() *Config {
    return &Config{
        GRPCPort:       getEnv("GRPC_PORT", "9095"),
        MaxAPIURL:      getEnv("MAX_API_URL", "https://api.max.ru"),
        MaxAPIToken:    getEnv("MAX_API_TOKEN", ""),
        RequestTimeout: getDurationEnv("MAX_API_TIMEOUT", 5*time.Second),
    }
}
```

## Data Models

### Max API Client Library Models

The max-bot-api-client-go library provides the following key types (based on typical Max Messenger API structure):

```go
// User represents a Max Messenger user
type User struct {
    ID          string // Max ID
    Phone       string
    FirstName   string
    LastName    string
    // ... other fields
}

// Error response from Max API
type APIError struct {
    Code    int
    Message string
}
```

### Internal Data Flow

1. **Request Flow**: gRPC Request → Handler → UseCase → Domain Interface → Infrastructure Client → Max API
2. **Response Flow**: Max API → Infrastructure Client (map to domain types) → UseCase → Handler → gRPC Response
3. **Error Flow**: Max API Error → Infrastructure Client (map to domain error) → UseCase → Handler (map to proto error code) → gRPC Response

## Correctne
ss Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

Property 1: Client initialization with valid tokens
*For any* valid bot token string, initializing the Max API client should succeed without errors
**Validates: Requirements 1.1**

Property 2: API calls use real client
*For any* valid phone number, calling GetMaxIDByPhone should result in an actual API request to Max Messenger (not stub behavior)
**Validates: Requirements 2.1**

Property 3: Max ID extraction from API response
*For any* user returned by the Max API, the service should extract and return the user's Max ID
**Validates: Requirements 2.2**

Property 4: Invalid phone rejection
*For any* phone number that fails validation, GetMaxIDByPhone should return ERROR_CODE_INVALID_PHONE without making an API call
**Validates: Requirements 2.4**

Property 5: Phone normalization consistency
*For any* phone number input, the normalization function should produce a consistent result when called multiple times with the same input
**Validates: Requirements 3.1**

Property 6: Eight-to-seven conversion
*For any* 11-digit phone number starting with "8", normalization should convert it to start with "7" while preserving the remaining 10 digits
**Validates: Requirements 3.2**

Property 7: Ten-digit prepending
*For any* 10-digit phone number, normalization should prepend "7" to create an 11-digit number
**Validates: Requirements 3.3**

Property 8: Length validation
*For any* phone number with fewer than 10 or more than 15 digits (after removing non-digits), validation should reject it as invalid
**Validates: Requirements 3.4**

Property 9: Non-digit removal
*For any* phone number containing non-digit characters, the validation process should remove all non-digits before applying normalization rules
**Validates: Requirements 3.5**

Property 10: Error messages presence
*For any* error returned by the service, the response should include a non-empty descriptive error message
**Validates: Requirements 5.5**

## Error Handling

### Error Mapping Strategy

The infrastructure layer maps Max API errors to domain errors:

```go
func (c *Client) mapAPIError(err error) error {
    if err == nil {
        return nil
    }
    
    // Check for specific Max API error types
    switch {
    case isNotFoundError(err):
        return domain.ErrMaxIDNotFound
    case isAuthError(err):
        // Log authentication error
        return fmt.Errorf("max api authentication failed: %w", err)
    case isRateLimitError(err):
        // Log rate limit error
        return fmt.Errorf("max api rate limit exceeded: %w", err)
    case isTimeoutError(err):
        return fmt.Errorf("max api request timeout: %w", err)
    default:
        // Log unexpected error with full details
        return fmt.Errorf("max api error: %w", err)
    }
}
```

### Error Categories

1. **Validation Errors** (domain.ErrInvalidPhone)
   - Detected before API calls
   - Returned immediately to caller
   - No API request made

2. **Not Found Errors** (domain.ErrMaxIDNotFound)
   - User not found in Max Messenger
   - Mapped from Max API 404 responses
   - Expected error case

3. **Internal Errors** (wrapped errors)
   - Authentication failures
   - Rate limiting
   - Network timeouts
   - Unexpected API errors
   - All logged with full details

### Logging Strategy

- **Authentication Errors**: Log at ERROR level with sanitized token info
- **Rate Limit Errors**: Log at WARN level with retry information
- **Timeout Errors**: Log at WARN level with request duration
- **Unexpected Errors**: Log at ERROR level with full error details and stack trace
- **Successful API Calls**: Log at DEBUG level with phone number (last 4 digits only)

## Testing Strategy

### Unit Testing

Unit tests will verify specific behaviors and edge cases:

1. **Phone Normalization Tests**
   - Test specific examples: "89991234567" → "79991234567"
   - Test 10-digit numbers: "9991234567" → "79991234567"
   - Test non-digit removal: "+7 (999) 123-45-67" → "79991234567"
   - Test invalid lengths: "123" → invalid, "12345678901234567" → invalid

2. **Error Mapping Tests**
   - Test Max API not found → domain.ErrMaxIDNotFound
   - Test Max API auth error → wrapped internal error
   - Test Max API timeout → wrapped timeout error

3. **Configuration Tests**
   - Test environment variable loading
   - Test default values
   - Test missing required configuration

### Property-Based Testing

Property-based tests will verify universal properties across many inputs using the `gopter` library for Go:

**Testing Framework**: gopter (https://github.com/leanovate/gopter)
- Minimum 100 iterations per property test
- Each property test tagged with: `**Feature: max-api-client-integration, Property {number}: {property_text}**`

1. **Property 1: Client initialization with valid tokens**
   - Generate random valid token strings
   - Verify client initializes successfully
   - Tag: `**Feature: max-api-client-integration, Property 1: Client initialization with valid tokens**`

2. **Property 4: Invalid phone rejection**
   - Generate random invalid phone numbers
   - Verify all return ERROR_CODE_INVALID_PHONE
   - Verify no API calls are made
   - Tag: `**Feature: max-api-client-integration, Property 4: Invalid phone rejection**`

3. **Property 5: Phone normalization consistency**
   - Generate random phone numbers
   - Call normalization twice
   - Verify results are identical
   - Tag: `**Feature: max-api-client-integration, Property 5: Phone normalization consistency**`

4. **Property 6: Eight-to-seven conversion**
   - Generate random 11-digit numbers starting with "8"
   - Verify conversion to "7" prefix
   - Verify last 10 digits unchanged
   - Tag: `**Feature: max-api-client-integration, Property 6: Eight-to-seven conversion**`

5. **Property 7: Ten-digit prepending**
   - Generate random 10-digit numbers
   - Verify "7" is prepended
   - Verify result is 11 digits
   - Tag: `**Feature: max-api-client-integration, Property 7: Ten-digit prepending**`

6. **Property 8: Length validation**
   - Generate random phone numbers with invalid lengths
   - Verify all are rejected
   - Tag: `**Feature: max-api-client-integration, Property 8: Length validation**`

7. **Property 9: Non-digit removal**
   - Generate random phone numbers with non-digits
   - Verify all non-digits are removed
   - Verify validation works on cleaned number
   - Tag: `**Feature: max-api-client-integration, Property 9: Non-digit removal**`

8. **Property 10: Error messages presence**
   - Generate various error conditions
   - Verify all error responses contain non-empty messages
   - Tag: `**Feature: max-api-client-integration, Property 10: Error messages presence**`

### Integration Testing

Integration tests will verify the complete flow with a test Max API instance or mock server:

1. **Successful User Lookup**
   - Call GetMaxIDByPhone with known test phone
   - Verify correct Max ID returned

2. **User Not Found**
   - Call GetMaxIDByPhone with non-existent phone
   - Verify ERROR_CODE_MAX_ID_NOT_FOUND returned

3. **Invalid Phone Handling**
   - Call GetMaxIDByPhone with invalid phone
   - Verify ERROR_CODE_INVALID_PHONE returned
   - Verify no API call made

4. **Usecase Layer Compatibility**
   - Verify usecase layer works unchanged with new client
   - Test all existing usecase methods

## Implementation Notes

### Dependency Management

Add to go.mod:
```
require github.com/max-messenger/max-bot-api-client-go v1.x.x
```

Use `go get github.com/max-messenger/max-bot-api-client-go@latest` to fetch the latest stable version.

### Client Initialization

The client should be initialized once at service startup and reused for all requests:

```go
func main() {
    cfg := config.Load()
    
    // Validate required configuration
    if cfg.MaxAPIToken == "" {
        log.Fatal("MAX_API_TOKEN is required")
    }
    
    // Initialize Max API client
    apiClient, err := maxapi.NewClient(cfg.MaxAPIToken, cfg.MaxAPIURL)
    if err != nil {
        log.Fatalf("failed to initialize Max API client: %v", err)
    }
    
    // Continue with service initialization...
}
```

### Phone Normalization Logic

Keep the existing normalization logic as it's already well-tested:

1. Remove all non-digit characters
2. Check length (must be 10-15 digits)
3. Apply Russian phone number rules:
   - 11 digits starting with "8" → replace "8" with "7"
   - 10 digits → prepend "7"
   - Other valid lengths → keep as-is
4. Return normalized phone or empty string if invalid

### Max API Client Usage

The max-bot-api-client-go library likely provides methods like:

```go
// Get user by phone number
user, err := client.GetUserByPhone(ctx, phone)

// Send message
err := client.SendMessage(ctx, chatID, message)

// Other bot operations...
```

Our wrapper will use these methods to implement the domain interface.

### Future Extensions

The Max API client library likely supports additional operations:

- Sending messages to users/chats
- Creating and managing chats
- Uploading files
- Managing bot commands
- Webhook handling

These can be added to the domain interface and implemented in the infrastructure layer as needed, without breaking existing functionality.

## Security Considerations

1. **Token Storage**: Bot token should never be logged or exposed in error messages
2. **Phone Number Privacy**: Log only last 4 digits of phone numbers
3. **Error Messages**: Sanitize error messages before returning to clients
4. **Rate Limiting**: Respect Max API rate limits to avoid service disruption
5. **Timeout Configuration**: Set reasonable timeouts to prevent resource exhaustion

## Performance Considerations

1. **Connection Pooling**: The Max API client should reuse HTTP connections
2. **Timeout Configuration**: Default 5-second timeout, configurable via environment
3. **Context Propagation**: Pass context through all layers for proper cancellation
4. **Logging Overhead**: Use appropriate log levels (DEBUG for success, ERROR for failures)

## Deployment Considerations

### Environment Variables

Required:
- `MAX_API_TOKEN`: Bot authentication token (no default, must be provided)

Optional:
- `MAX_API_URL`: API base URL (default: "https://api.max.ru")
- `MAX_API_TIMEOUT`: Request timeout (default: "5s")
- `GRPC_PORT`: gRPC server port (default: "9095")

### Migration Path

1. Deploy new version with max-bot-api-client-go integration
2. Configure MAX_API_TOKEN in environment
3. Monitor logs for API errors
4. Verify GetMaxIDByPhone returns real Max IDs
5. Remove old stub implementation code

### Rollback Plan

If issues occur:
1. Revert to previous version with stub implementation
2. Investigate API connectivity issues
3. Verify token configuration
4. Check Max API service status
