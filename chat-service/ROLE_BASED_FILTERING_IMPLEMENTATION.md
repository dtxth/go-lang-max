# Role-Based Filtering Implementation

## Overview

This document describes the implementation of role-based filtering in the Chat Service, which enables different users to see different sets of chats based on their role and organizational context.

## Implementation Summary

### 1. Domain Layer

#### New Domain Entities

**`domain.AuthService`** - Interface for Auth Service integration
- `ValidateToken(token string) (*TokenInfo, error)` - Validates JWT tokens via gRPC

**`domain.TokenInfo`** - Contains user information from JWT token
- Fields: `Valid`, `UserID`, `Email`, `Role`, `UniversityID`, `BranchID`, `FacultyID`

**`domain.ChatFilter`** - Value object for role-based filtering
- Fields: `Role`, `UniversityID`, `BranchID`, `FacultyID`
- Methods: `IsSuperadmin()`, `IsCurator()`, `IsOperator()`
- Factory: `NewChatFilter(tokenInfo *TokenInfo)`

#### New Domain Errors
- `ErrInvalidToken` - Invalid or expired JWT token
- `ErrUnauthorized` - Missing authentication
- `ErrForbidden` - Insufficient permissions
- `ErrInvalidRole` - Invalid role specified

### 2. Infrastructure Layer

#### Auth gRPC Client (`infrastructure/auth/auth_client.go`)
- Connects to Auth Service via gRPC
- Implements `domain.AuthService` interface
- Validates JWT tokens and extracts user information
- Configuration: `AUTH_GRPC_ADDR` (default: localhost:9090)

#### Auth Middleware (`infrastructure/http/middleware.go`)
- `AuthMiddleware.Authenticate()` - HTTP middleware for JWT validation
- Extracts Bearer token from Authorization header
- Calls Auth Service to validate token
- Stores `TokenInfo` in request context
- Returns 401 for invalid/missing tokens

#### Updated Repository (`infrastructure/repository/chat_postgres.go`)
- Updated `Search()` and `GetAll()` to accept `*domain.ChatFilter`
- Filtering logic:
  - **Superadmin**: No filtering, sees all chats
  - **Curator**: Filters by `university_id`
  - **Operator**: Filters by `university_id` (TODO: add branch/faculty filtering)

### 3. Use Case Layer

#### ListChatsWithRoleFilterUseCase (`usecase/list_chats_with_role_filter.go`)
- Implements role-based chat filtering logic
- Validates that curators and operators have `university_id`
- Applies pagination defaults (limit: 50, max: 100)
- Returns `ErrForbidden` if curator/operator lacks context

#### Updated ChatService (`usecase/chat_service.go`)
- Integrated `ListChatsWithRoleFilterUseCase`
- Updated `SearchChats()` and `GetAllChats()` to use `ChatFilter`

### 4. HTTP Layer

#### Updated Handlers (`infrastructure/http/handler.go`)
- `SearchChats()` - Now requires authentication, extracts filter from token
- `GetAllChats()` - Now requires authentication, extracts filter from token
- Returns 401 for missing/invalid tokens
- Returns 403 for insufficient permissions

#### Updated Router (`infrastructure/http/router.go`)
- Applied `authMiddleware.Authenticate()` to `/chats` and `/chats/all` endpoints
- All chat list endpoints now require valid JWT token

### 5. Configuration

#### Updated Config (`internal/config/config.go`)
- Added `AuthAddress` - Auth Service gRPC address
- Added `AuthTimeout` - Auth Service call timeout
- Environment variables:
  - `AUTH_GRPC_ADDR` (default: localhost:9090)
  - `AUTH_TIMEOUT` (default: 5s)

#### Updated go.mod
- Added `auth-service` dependency
- Added replace directive: `replace auth-service => ../auth-service`

### 6. Main Application (`cmd/chat/main.go`)
- Initialize Auth gRPC client
- Create AuthMiddleware with Auth client
- Pass AuthMiddleware to HTTP handler

## Filtering Rules

### Superadmin
- **Access**: All chats from all universities
- **Filter**: No filtering applied
- **Context Required**: None

### Curator
- **Access**: Only chats from their assigned university
- **Filter**: `WHERE c.university_id = $university_id`
- **Context Required**: `university_id` must be present in token

### Operator
- **Access**: Only chats from their assigned branch/faculty
- **Filter**: Currently filters by `university_id` (branch/faculty filtering TODO)
- **Context Required**: `university_id` must be present in token
- **Note**: Full operator filtering requires additional schema changes or integration with Structure Service

## API Changes

### Before
```
GET /chats?query=test&limit=50&offset=0&user_role=curator&university_id=1
GET /chats/all?limit=50&offset=0&user_role=curator&university_id=1
```

### After
```
GET /chats?query=test&limit=50&offset=0
Authorization: Bearer <jwt_token>

GET /chats/all?limit=50&offset=0
Authorization: Bearer <jwt_token>
```

Role and context are now extracted from the JWT token automatically.

## Testing

### Unit Tests (`usecase/list_chats_with_role_filter_test.go`)
- ✅ Superadmin sees all chats
- ✅ Curator sees only university chats
- ✅ Curator without university_id returns forbidden
- ✅ Nil filter returns invalid role error
- ✅ Pagination defaults applied correctly
- ✅ Pagination limit capped at 100
- ✅ Repository errors propagated correctly

All tests pass successfully.

## Requirements Validation

### Requirement 5.4: Token Validation
✅ Implemented - Auth middleware validates JWT via Auth Service gRPC

### Requirement 5.1: Superadmin Access
✅ Implemented - Superadmin filter returns all chats without restrictions

### Requirement 5.2: Curator Filtering
✅ Implemented - Curator filter restricts to assigned university

### Requirement 5.3: Operator Filtering
⚠️ Partially Implemented - Currently filters by university, branch/faculty filtering requires schema changes

### Requirement 5.5: Invalid Role Handling
✅ Implemented - Returns 403 for invalid roles or missing context

## Future Enhancements

1. **Operator Branch/Faculty Filtering**
   - Add `branch_id` and `faculty_id` columns to `chats` table
   - Or integrate with Structure Service to resolve group → chat relationships
   - Update repository filtering logic

2. **Caching**
   - Cache token validation results to reduce Auth Service calls
   - Implement token cache with TTL matching token expiry

3. **Audit Logging**
   - Log all filtered queries with user context
   - Track access patterns by role

4. **Performance Optimization**
   - Add database indexes for role-based filtering
   - Optimize queries for large datasets

## Dependencies

- **Auth Service**: Must be running on configured gRPC address
- **Auth Service Proto**: `auth-service/api/proto` must be available
- **JWT Tokens**: Must include role and context fields (university_id, branch_id, faculty_id)

## Environment Variables

```bash
# Auth Service Configuration
AUTH_GRPC_ADDR=localhost:9090
AUTH_TIMEOUT=5s

# Existing Configuration
DATABASE_URL=postgres://...
PORT=8082
GRPC_PORT=9092
MAXBOT_GRPC_ADDR=localhost:9095
MAXBOT_TIMEOUT=5s
```

## Error Responses

### 401 Unauthorized
```json
{
  "error": "missing authorization header"
}
```

### 401 Unauthorized (Invalid Token)
```json
{
  "error": "invalid or expired token"
}
```

### 403 Forbidden
```json
{
  "error": "forbidden: insufficient permissions"
}
```

### 500 Internal Server Error
```json
{
  "error": "authentication service error"
}
```
