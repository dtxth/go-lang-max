# Chat Search Implementation

## Overview

This document describes the implementation of full-text search functionality for chats with Russian language support, multi-word search, and role-based filtering.

## Requirements

- **Requirement 17.1**: Use full-text search on chat name
- **Requirement 17.2**: Apply role-based filtering before returning results
- **Requirement 17.3**: Configure Russian language text search
- **Requirement 17.4**: Support multi-word search (all words must match)
- **Requirement 17.5**: Return empty array for no matches

## Implementation Details

### Database Schema

The database already has a GIN index for full-text search with Russian language support:

```sql
CREATE INDEX IF NOT EXISTS idx_chats_name_search 
ON chats USING gin(to_tsvector('russian', name));
```

This index enables efficient full-text search on chat names using PostgreSQL's built-in text search capabilities.

### Full-Text Search Query

The implementation uses PostgreSQL's `to_tsvector` and `to_tsquery` functions with Russian language configuration:

```sql
WHERE to_tsvector('russian', c.name) @@ to_tsquery('russian', $1)
```

#### Multi-Word Search

For multi-word queries, the implementation:
1. Splits the query into individual words using `strings.Fields()`
2. Creates a tsquery with AND operator (`&`) between words
3. Adds prefix matching (`:*`) to each word for partial matches

Example:
- Query: "группа математика"
- Generated tsquery: "группа:* & математика:*"
- Matches: "Группа математика 101", "Математика группа А"
- Does not match: "Группа физики" (missing "математика")

### Role-Based Filtering

The search applies role-based filtering before returning results:

- **Superadmin**: Sees all chats from all universities
- **Curator**: Sees only chats from their assigned university
- **Operator**: Sees only chats from their assigned university (future: branch/faculty filtering)

The filtering is applied in the WHERE clause:

```sql
WHERE to_tsvector('russian', c.name) @@ to_tsquery('russian', $1)
  AND c.university_id = $2  -- For curator/operator
```

### Result Ordering

Results are ordered by relevance using `ts_rank`:

```sql
ORDER BY ts_rank(to_tsvector('russian', c.name), to_tsquery('russian', $1)) DESC, c.name
```

This ensures:
1. Most relevant results appear first
2. Results with the same relevance are sorted alphabetically

### Empty Query Handling

When the query is empty or contains only whitespace:
- The implementation calls `GetAll()` instead
- Returns all chats with role-based filtering applied
- Uses alphabetical ordering

### No Matches

When no chats match the search criteria:
- Returns an empty array `[]`
- Returns total count of 0
- Returns HTTP 200 status code (not 404)

## API Endpoint

### GET /chats

**Query Parameters:**
- `query` (optional): Search query for chat name
- `limit` (optional): Maximum number of results (default: 50, max: 100)
- `offset` (optional): Pagination offset

**Headers:**
- `Authorization`: Bearer token (required)

**Response:**
```json
{
  "chats": [
    {
      "id": 1,
      "name": "Группа математика 101",
      "url": "https://example.com/chat1",
      "max_chat_id": "chat123",
      "participants_count": 25,
      "university_id": 1,
      "department": "Математический факультет",
      "source": "academic_group",
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z",
      "administrators": []
    }
  ],
  "total_count": 1,
  "limit": 50,
  "offset": 0
}
```

## Examples

### Single Word Search

**Request:**
```
GET /chats?query=математика
Authorization: Bearer <token>
```

**Result:** Returns all chats containing "математика" in the name

### Multi-Word Search

**Request:**
```
GET /chats?query=группа математика
Authorization: Bearer <token>
```

**Result:** Returns only chats containing BOTH "группа" AND "математика"

### Search with Pagination

**Request:**
```
GET /chats?query=физика&limit=20&offset=0
Authorization: Bearer <token>
```

**Result:** Returns first 20 chats matching "физика"

### Empty Query (List All)

**Request:**
```
GET /chats?limit=50&offset=0
Authorization: Bearer <token>
```

**Result:** Returns all chats (with role-based filtering)

### No Matches

**Request:**
```
GET /chats?query=несуществующий
Authorization: Bearer <token>
```

**Response:**
```json
{
  "chats": [],
  "total_count": 0,
  "limit": 50,
  "offset": 0
}
```

## Testing

### Unit Tests

The implementation includes comprehensive unit tests:

1. **TestSearchChats_EmptyQuery**: Verifies empty query returns all chats
2. **TestSearchChats_WithQuery**: Verifies single-word search
3. **TestSearchChats_MultiWordQuery**: Verifies multi-word search (AND logic)
4. **TestSearchChats_NoMatches**: Verifies empty array for no matches
5. **TestSearchChats_WithRoleFiltering**: Verifies role-based filtering is applied
6. **TestSearchChats_RepositoryError**: Verifies error handling

### Integration Testing

To test with real database:

```bash
# Start PostgreSQL
docker-compose up -d chat-db

# Run integration tests
go test ./internal/infrastructure/repository/... -v
```

## Performance Considerations

### Index Usage

The GIN index on `to_tsvector('russian', name)` provides:
- Fast full-text search (O(log n) lookup)
- Efficient multi-word queries
- Support for Russian morphology (stemming)

### Query Optimization

- Uses prepared statements with parameterized queries
- Batch loads administrators for multiple chats
- Applies filtering in database (not in application)
- Uses LIMIT/OFFSET for pagination

### Scalability

For large datasets (150,000+ chats):
- GIN index handles millions of documents efficiently
- Pagination prevents memory issues
- Role-based filtering reduces result set size
- Consider adding caching for frequently searched terms

## Russian Language Support

PostgreSQL's Russian text search configuration provides:

1. **Stemming**: "группа", "группы", "группе" all match "группа"
2. **Stop words**: Common words like "и", "в", "на" are ignored
3. **Morphology**: Handles Russian word forms and cases
4. **Accent insensitivity**: "е" and "ё" are treated as equivalent

## Future Enhancements

1. **Operator Branch/Faculty Filtering**: Add filtering by branch_id and faculty_id for operators
2. **Search Highlighting**: Return matched text with highlights
3. **Fuzzy Search**: Add support for typo tolerance
4. **Search Analytics**: Track popular search terms
5. **Autocomplete**: Suggest chat names as user types
6. **Advanced Filters**: Filter by source, department, participant count
