# MAX Chat Info Testing Guide

## Overview

This document describes how to test the MAX Chat Info functionality in the maxbot-service.

## Current Status

✅ **HTTP API /api/v1/chats/{chat_id}**: **FIXED AND WORKING** - полностью функционален  
✅ **HTTP сервер**: Запускается корректно, все endpoints работают  
✅ **MAX API Integration**: Working correctly  
✅ **Error Handling**: Корректная обработка ошибок (404 для несуществующих чатов, 400 для невалидных ID)  

## Testing gRPC GetChatInfo

### Prerequisites

1. Ensure maxbot-service is running:
   ```bash
   docker-compose up maxbot-service
   ```

2. Install grpcurl:
   ```bash
   brew install grpcurl  # macOS
   ```

### Test Commands

#### 1. Test with positive chat ID
```bash
grpcurl -plaintext -import-path maxbot-service/api/proto -proto maxbot.proto \
  -d '{"chat_id": 123456789}' \
  localhost:9095 maxbot.MaxBotService/GetChatInfo
```

**Expected Response** (for non-existent chat):
```json
{
  "errorCode": "ERROR_CODE_INTERNAL",
  "error": "max api error: HTTP 404: Not Found"
}
```

#### 2. Test with negative chat ID
```bash
grpcurl -plaintext -import-path maxbot-service/api/proto -proto maxbot.proto \
  -d '{"chat_id": -69257108032233}' \
  localhost:9095 maxbot.MaxBotService/GetChatInfo
```

#### 3. Test with real MAX chat ID (if available)
```bash
grpcurl -plaintext -import-path maxbot-service/api/proto -proto maxbot.proto \
  -d '{"chat_id": YOUR_REAL_CHAT_ID}' \
  localhost:9095 maxbot.MaxBotService/GetChatInfo
```

**Expected Response** (for existing chat):
```json
{
  "chat": {
    "chatId": "123456789",
    "title": "Test Chat",
    "type": "group",
    "participantsCount": 25,
    "description": "Chat description"
  }
}
```

## Testing HTTP API

### Working Endpoints
```bash
# Health check - работает
curl http://localhost:8095/health
# Response: {"status":"ok","service":"maxbot-service"}

# Bot info - работает  
curl http://localhost:8095/api/v1/me
# Response: {"name":"Digital University Support Bot","add_link":"..."}
```

### Chat Endpoint (✅ WORKING)
```bash
curl -X 'GET' "http://localhost:8095/api/v1/chats/123456789" -H 'accept: application/json'
```

**Response for non-existent chat**: `{"error":"chat not found","message":"max api error: HTTP 404: Not Found"}`  
**Response for invalid chat_id**: `{"error":"invalid chat_id"}`  
**Status**: ✅ **FIXED AND WORKING** - HTTP endpoint полностью функционален

## Implementation Details

### gRPC Service Definition
```protobuf
service MaxBotService {
  rpc GetChatInfo(GetChatInfoRequest) returns (GetChatInfoResponse);
}

message GetChatInfoRequest {
  int64 chat_id = 1;
}

message GetChatInfoResponse {
  ChatInfo chat = 1;
  ErrorCode error_code = 2;
  string error = 3;
}
```

### MAX API Integration
- Uses official MAX Bot API client
- Handles authentication with MAX_API_TOKEN
- Maps MAX API errors to domain errors
- Supports both positive and negative chat IDs

### Error Handling
- `ERROR_CODE_INTERNAL`: MAX API returned 404 (chat not found)
- `ERROR_CODE_INVALID_PHONE`: Invalid phone number format
- `ERROR_CODE_MAX_ID_NOT_FOUND`: MAX ID not found

## Troubleshooting

### Common Issues

1. **gRPC reflection not enabled**: Use proto files directly
2. **Chat not found (404)**: Normal for test chat IDs
3. **Authentication errors**: Check MAX_API_TOKEN
4. **HTTP routing issues**: Use gRPC as workaround

### Debug Commands

Check service status:
```bash
docker-compose logs maxbot-service --tail=20
```

Test basic connectivity:
```bash
curl http://localhost:8095/health
curl http://localhost:8095/api/v1/me
```

## Next Steps

1. Fix HTTP routing issue in `maxbot-service/internal/infrastructure/http/server.go`
2. Add Swagger documentation for chat endpoints
3. Add integration tests for chat info functionality
4. Test with real MAX chat IDs when available