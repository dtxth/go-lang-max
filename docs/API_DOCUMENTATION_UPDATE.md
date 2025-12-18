# API Documentation Update: Profile Source Integration

## Overview

This document describes the API changes and new endpoints introduced with the MAX webhook profile integration system. All existing APIs remain backward compatible while gaining enhanced profile functionality.

## Employee Service API Updates

### Enhanced Employee Creation

#### POST /employees

The employee creation endpoint now automatically integrates with the profile cache system.

**Enhanced Request Body** (all fields optional except required ones):
```json
{
  "phone": "+79001234567",
  "first_name": "Иван",          // Optional: If empty, will use cached profile
  "last_name": "Петров",         // Optional: If empty, will use cached profile  
  "middle_name": "Иванович",     // Optional
  "max_id": "12345",             // Optional: If provided, used for profile lookup
  "university_id": 1,            // Required
  "inn": "1234567890",           // Optional
  "kpp": "123456789"             // Optional
}
```

**Enhanced Response** (includes profile source information):
```json
{
  "id": 123,
  "phone": "+79001234567",
  "first_name": "Иван",
  "last_name": "Петров",
  "middle_name": "Иванович",
  "max_id": "12345",
  "university_id": 1,
  "university_name": "МГУ",
  "inn": "1234567890",
  "kpp": "123456789",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  
  // NEW: Profile source tracking
  "profile_source": "webhook",                    // Source of name data
  "profile_last_updated": "2024-01-15T10:30:00Z" // When profile was last updated
}
```

**Profile Source Values**:
- `"manual"` - Names provided explicitly in the request
- `"webhook"` - Names retrieved from MAX Messenger webhook events
- `"user_input"` - Names provided by user through bot commands
- `"default"` - Default values used when no profile available

**Behavior Changes**:
1. **When names are provided**: Uses provided names, sets `profile_source` to `"manual"`
2. **When names are empty**: Attempts to retrieve from profile cache using `max_id`
3. **When profile found**: Uses cached names, sets `profile_source` to `"webhook"` or `"user_input"`
4. **When profile not found**: Uses default values, sets `profile_source` to `"default"`

### Enhanced Employee Retrieval

#### GET /employees/{id}

**Enhanced Response** (includes profile source information):
```json
{
  "id": 123,
  "phone": "+79001234567",
  "first_name": "Иван",
  "last_name": "Петров",
  "middle_name": "Иванович",
  "max_id": "12345",
  "university_id": 1,
  "university_name": "МГУ",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z",
  
  // NEW: Profile source information
  "profile_source": "webhook",
  "profile_last_updated": "2024-01-15T10:30:00Z"
}
```

#### GET /employees (Search)

**Enhanced Response** (each employee includes profile source):
```json
{
  "employees": [
    {
      "id": 123,
      "first_name": "Иван",
      "last_name": "Петров",
      "phone": "+79001234567",
      "max_id": "12345",
      "university_name": "МГУ",
      "profile_source": "webhook",
      "profile_last_updated": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "limit": 50,
  "offset": 0
}
```

## MaxBot Service API - New Endpoints

### Webhook Processing

#### POST /webhook/max

Processes incoming webhook events from MAX Messenger to collect user profile data.

**Request Body** (MAX Messenger webhook format):
```json
{
  "type": "message_new",
  "message": {
    "from": {
      "user_id": "12345",
      "first_name": "Иван",
      "last_name": "Петров"
    },
    "text": "Hello",
    "chat": {
      "chat_id": 67890,
      "type": "private"
    }
  }
}
```

**Alternative Event Type**:
```json
{
  "type": "callback_query",
  "callback_query": {
    "user": {
      "user_id": "12345",
      "first_name": "Иван",
      "last_name": "Петров"
    },
    "data": "button_data",
    "chat": {
      "chat_id": 67890,
      "type": "private"
    }
  }
}
```

**Response**: Always returns `200 OK` (as required by MAX Messenger)
```json
{
  "status": "ok"
}
```

**Behavior**:
- Extracts user profile from `message.from` or `callback_query.user`
- Stores/updates profile in Redis cache with 30-day TTL
- Logs processing statistics for monitoring
- Handles malformed events gracefully

### Profile Management API

#### GET /profiles/{user_id}

Retrieve cached user profile information.

**Response**:
```json
{
  "user_id": "12345",
  "max_first_name": "Иван",
  "max_last_name": "Петров",
  "user_provided_name": "",
  "display_name": "Иван Петров",
  "source": "webhook",
  "last_updated": "2024-01-15T10:30:00Z",
  "has_full_name": true
}
```

**Error Responses**:
- `404 Not Found` - Profile not found in cache
- `400 Bad Request` - Invalid user ID format
- `500 Internal Server Error` - Cache unavailable

#### PUT /profiles/{user_id}

Update user profile information (admin use).

**Request Body**:
```json
{
  "max_first_name": "Иван",
  "max_last_name": "Петров",
  "user_provided_name": "Иван П."
}
```

**Response**: Updated profile (same format as GET)

#### POST /profiles/{user_id}/name

Set user-provided name (typically called by bot commands).

**Request Body**:
```json
{
  "name": "Иван Петрович"
}
```

**Response**: Updated profile with `user_provided_name` set and `source` changed to `"user_input"`

#### GET /profiles/stats

Get profile statistics and coverage metrics.

**Response**:
```json
{
  "total_profiles": 1000,
  "profiles_with_full_name": 750,
  "profiles_by_source": {
    "webhook": 600,
    "user_input": 150,
    "default": 250
  }
}
```

### Monitoring API

#### GET /monitoring/profiles/coverage

Get detailed profile coverage metrics.

**Response**:
```json
{
  "total_users": 10000,
  "users_with_profiles": 8000,
  "users_with_full_names": 6000,
  "coverage_percentage": 80.0,
  "full_name_percentage": 60.0,
  "profiles_by_source": {
    "webhook": 7000,
    "user_input": 1000,
    "default": 2000
  },
  "last_updated": "2024-01-15T10:30:00Z"
}
```

#### GET /monitoring/profiles/quality

Get detailed profile quality report.

**Response**:
```json
{
  "total_profiles": 10000,
  "quality_metrics": {
    "complete_profiles": 7000,
    "partial_profiles": 2000,
    "empty_profiles": 1000,
    "stale_profiles": 500,
    "quality_score": 85.5,
    "completeness_score": 70.0,
    "freshness_score": 90.0
  },
  "source_breakdown": {
    "webhook": {
      "count": 7000,
      "complete_profiles": 6000,
      "quality_score": 85.7,
      "average_age_days": 7.5
    },
    "user_input": {
      "count": 1000,
      "complete_profiles": 1000,
      "quality_score": 100.0,
      "average_age_days": 2.1
    }
  },
  "data_issues": [
    {
      "type": "incomplete_profiles",
      "description": "Profiles missing last names",
      "count": 2000,
      "severity": "medium"
    }
  ],
  "recommended_actions": [
    "Encourage users to provide full names via bot commands",
    "Review webhook event processing for completeness"
  ],
  "generated_at": "2024-01-15T10:30:00Z"
}
```

#### GET /monitoring/webhook/stats

Get webhook processing statistics.

**Query Parameters**:
- `period` (optional): Time period for statistics (`hour`, `day`, `week`, `month`)

**Response**:
```json
{
  "total_events": 5000,
  "successful_events": 4800,
  "failed_events": 200,
  "events_by_type": {
    "message_new": 4000,
    "callback_query": 1000
  },
  "errors_by_type": {
    "invalid_json": 100,
    "missing_user_info": 50,
    "cache_error": 50
  },
  "profiles_extracted": 3000,
  "profiles_stored": 2900,
  "average_processing_time_ms": 150.5,
  "period": {
    "from": "2024-01-15T00:00:00Z",
    "to": "2024-01-16T00:00:00Z"
  }
}
```

## Error Handling

### Common Error Responses

All APIs use consistent error response format:

```json
{
  "error": "error_code",
  "message": "Human-readable error message",
  "details": {
    "field": "specific_field",
    "code": "validation_error"
  }
}
```

### Profile-Related Error Codes

- `profile_not_found` - User profile not found in cache
- `profile_cache_unavailable` - Profile cache service unavailable
- `invalid_profile_data` - Invalid profile data format
- `webhook_processing_error` - Error processing webhook event
- `profile_update_failed` - Failed to update profile information

### Graceful Degradation

The system is designed to continue operating even when profile services are unavailable:

1. **Employee Creation**: Falls back to default names when profiles unavailable
2. **Profile Retrieval**: Returns empty profile instead of failing
3. **Webhook Processing**: Logs errors but continues processing other events
4. **Cache Failures**: Uses in-memory fallback or skips caching

## Backward Compatibility

### Existing API Behavior

All existing APIs maintain their original behavior:

1. **Employee Creation**: If `first_name`/`last_name` are provided, they are used as before
2. **Employee Retrieval**: Returns all original fields plus new profile source fields
3. **Search**: Works exactly as before, with additional profile source information
4. **Updates**: Existing update logic unchanged

### Migration Considerations

1. **New Fields**: New `profile_source` and `profile_last_updated` fields are added but optional
2. **Default Values**: Existing employees get `profile_source = "manual"` by default
3. **API Clients**: Existing clients continue to work without changes
4. **Response Format**: New fields are added to responses but don't break existing parsing

## Authentication and Authorization

### Webhook Endpoint

- **Authentication**: Optional webhook secret validation
- **Rate Limiting**: Consider implementing rate limiting for production
- **HTTPS**: Required for production deployments

### Profile Management API

- **Authentication**: Requires valid JWT token
- **Authorization**: 
  - GET operations: Any authenticated user
  - PUT/POST operations: Admin users only
  - Stats endpoints: Admin users only

### Monitoring API

- **Authentication**: Requires valid JWT token
- **Authorization**: Admin users or monitoring systems only
- **Rate Limiting**: Reasonable limits for monitoring queries

## Usage Examples

### Creating Employee with Automatic Profile

```bash
# Create employee - system will automatically get profile from cache
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "phone": "+79001234567",
    "max_id": "12345",
    "university_id": 1
  }'

# Response includes profile source information
{
  "id": 123,
  "first_name": "Иван",
  "last_name": "Петров",
  "phone": "+79001234567",
  "max_id": "12345",
  "university_id": 1,
  "profile_source": "webhook",
  "profile_last_updated": "2024-01-15T10:30:00Z"
}
```

### Managing User Profiles

```bash
# Get user profile
curl http://localhost:8095/profiles/12345 \
  -H "Authorization: Bearer $JWT_TOKEN"

# Set user-provided name
curl -X POST http://localhost:8095/profiles/12345/name \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{"name": "Иван Петрович"}'

# Get profile statistics
curl http://localhost:8095/profiles/stats \
  -H "Authorization: Bearer $JWT_TOKEN"
```

### Monitoring Profile Quality

```bash
# Get coverage metrics
curl http://localhost:8095/monitoring/profiles/coverage \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get detailed quality report
curl http://localhost:8095/monitoring/profiles/quality \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get webhook statistics
curl http://localhost:8095/monitoring/webhook/stats?period=day \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## Integration Testing

### Testing Profile Integration

```bash
# Test webhook processing
curl -X POST http://localhost:8095/webhook/max \
  -H "Content-Type: application/json" \
  -d '{
    "type": "message_new",
    "message": {
      "from": {
        "user_id": "test123",
        "first_name": "Тест",
        "last_name": "Пользователь"
      },
      "text": "Hello"
    }
  }'

# Verify profile was cached
curl http://localhost:8095/profiles/test123 \
  -H "Authorization: Bearer $JWT_TOKEN"

# Test employee creation with cached profile
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "phone": "+1234567890",
    "max_id": "test123",
    "university_id": 1
  }'
```

## Performance Considerations

### Caching Strategy

- **TTL**: 30 days default (configurable)
- **Memory Usage**: Monitor Redis memory usage
- **Hit Rate**: Target >90% cache hit rate for active users

### API Performance

- **Profile Lookup**: <100ms for cached profiles
- **Webhook Processing**: <50ms per event
- **Employee Creation**: <200ms including profile lookup

### Monitoring Recommendations

- Monitor profile cache hit rates
- Track webhook processing latency
- Alert on profile quality degradation
- Monitor Redis memory usage and performance

## Security Considerations

### Data Privacy

- Phone numbers are partially masked in logs
- Profile data has appropriate TTL
- Webhook secrets for authentication

### Access Control

- Profile management requires authentication
- Admin operations require elevated permissions
- Monitoring endpoints have appropriate access controls

### Input Validation

- All webhook data is validated and sanitized
- Profile updates validate data formats
- SQL injection protection in database queries

For additional technical details, refer to the main integration documentation and service-specific README files.