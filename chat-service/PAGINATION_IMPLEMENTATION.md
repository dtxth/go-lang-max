# Pagination Implementation for Chat Lists

## Overview

This document describes the pagination implementation for chat list endpoints in the Chat Service, fulfilling Requirements 16.1-16.5.

## Implementation Details

### Use Case Layer

The `ListChatsWithRoleFilterUseCase` handles pagination logic:

**Location:** `internal/usecase/list_chats_with_role_filter.go`

**Key Features:**
1. **Default Limit (Requirement 16.2):** When `limit <= 0`, defaults to 50
2. **Limit Cap (Requirement 16.3):** When `limit > 100`, caps at 100
3. **Offset Handling (Requirement 16.5):** Negative offsets are normalized to 0
4. **Total Count (Requirement 16.4):** Returns total count from repository

```go
// Apply pagination defaults
if limit <= 0 {
    limit = 50
}
if limit > 100 {
    limit = 100
}
if offset < 0 {
    offset = 0
}
```

### Repository Layer

The `ChatPostgres` repository implements pagination at the database level:

**Location:** `internal/infrastructure/repository/chat_postgres.go`

**Key Features:**
1. Counts total records matching the filter
2. Applies `LIMIT` and `OFFSET` to SQL query
3. Returns empty array when `offset >= total`

```sql
SELECT COUNT(*) FROM chats c WHERE ...
SELECT ... FROM chats c WHERE ... LIMIT $n OFFSET $m
```

### HTTP Handler Layer

The HTTP handler returns pagination metadata in the response:

**Location:** `internal/infrastructure/http/handler.go`

**Response Format:**
```json
{
  "chats": [...],
  "total_count": 150,
  "limit": 50,
  "offset": 0
}
```

**Endpoints:**
- `GET /chats?query=...&limit=50&offset=0` - Search chats with pagination
- `GET /chats/all?limit=50&offset=0` - Get all chats with pagination

## Requirements Validation

### ✅ Requirement 16.1: Accept limit and offset parameters
Both endpoints accept `limit` and `offset` query parameters.

### ✅ Requirement 16.2: Use default limit of 50
When `limit` is not provided or is 0, the system defaults to 50.

### ✅ Requirement 16.3: Cap limit at 100
When `limit` exceeds 100, it is automatically capped at 100.

### ✅ Requirement 16.4: Include total count in response metadata
The response includes `total_count` field with the total number of records matching the filter.

### ✅ Requirement 16.5: Return empty array for offset > total
When `offset` exceeds the total count, an empty array is returned with the correct `total_count`.

## Testing

### Unit Tests

**Location:** `internal/usecase/list_chats_with_role_filter_pagination_test.go`

Tests cover:
- `TestPaginationTotalCount` - Verifies total count is returned correctly across pages
- `TestPaginationOffsetExceedsTotal` - Verifies empty array when offset > total
- `TestPaginationWithRoleFiltering` - Verifies pagination works with role-based filtering

**Location:** `internal/usecase/list_chats_with_role_filter_test.go`

Existing tests cover:
- `TestListChatsWithRoleFilterUseCase_Execute_PaginationDefaults` - Verifies default limit of 50
- `TestListChatsWithRoleFilterUseCase_Execute_PaginationLimitCap` - Verifies limit cap at 100

### Running Tests

```bash
cd chat-service
go test -v ./internal/usecase/ -run TestPagination
go test -v ./internal/usecase/ -run "TestListChatsWithRoleFilterUseCase_Execute"
```

## Example Usage

### Request with default pagination
```bash
curl -H "Authorization: Bearer <token>" \
  http://localhost:8082/chats
```

Response:
```json
{
  "chats": [...50 chats...],
  "total_count": 150,
  "limit": 50,
  "offset": 0
}
```

### Request with custom pagination
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8082/chats?limit=25&offset=50"
```

Response:
```json
{
  "chats": [...25 chats...],
  "total_count": 150,
  "limit": 25,
  "offset": 50
}
```

### Request with offset exceeding total
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8082/chats?limit=50&offset=200"
```

Response:
```json
{
  "chats": [],
  "total_count": 150,
  "limit": 50,
  "offset": 200
}
```

### Request with limit exceeding cap
```bash
curl -H "Authorization: Bearer <token>" \
  "http://localhost:8082/chats?limit=200&offset=0"
```

Response (limit automatically capped at 100):
```json
{
  "chats": [...100 chats...],
  "total_count": 150,
  "limit": 100,
  "offset": 0
}
```

## Integration with Role-Based Filtering

Pagination works seamlessly with role-based filtering:

- **Superadmin:** Paginated results include all chats from all universities
- **Curator:** Paginated results include only chats from their university
- **Operator:** Paginated results include only chats from their branch/faculty

The `total_count` reflects the filtered total, not the absolute total of all chats in the system.

## Performance Considerations

1. **Database Indexing:** Ensure proper indexes on `university_id`, `name` for efficient filtering and sorting
2. **Count Query:** The count query runs separately from the data query for accuracy
3. **Batch Loading:** Administrators are loaded in batch for all chats in the result set to minimize N+1 queries

## Future Enhancements

Potential improvements for pagination:
1. Cursor-based pagination for better performance with large datasets
2. Caching of total counts for frequently accessed filters
3. Streaming responses for very large result sets
