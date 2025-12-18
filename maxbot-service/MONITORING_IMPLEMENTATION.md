# Monitoring and Analytics Implementation

## Overview

The MaxBot service now includes comprehensive monitoring and analytics capabilities for tracking webhook processing, profile coverage, and data quality. This implementation satisfies Requirements 6.1, 6.3, and 6.4 from the MAX Webhook Profile Integration specification.

## Features Implemented

### 1. Profile Coverage Metrics (Requirement 6.1)

Track the percentage of users with complete profile information:

- **Total Users**: Count of all users in the system
- **Users with Profiles**: Users who have any profile data
- **Users with Full Names**: Users with complete first and last names
- **Coverage Percentage**: Percentage of users with profiles
- **Full Name Percentage**: Percentage of users with complete names
- **Profiles by Source**: Breakdown by source (webhook, user_input, default)

**Endpoint**: `GET /api/v1/monitoring/profiles/coverage`

**Example Response**:
```json
{
  "total_users": 10000,
  "users_with_profiles": 8000,
  "users_with_full_names": 6000,
  "coverage_percentage": 80.0,
  "full_name_percentage": 60.0,
  "profiles_by_source": {
    "webhook": 5000,
    "user_input": 2000,
    "default": 1000
  },
  "last_updated": "2024-01-15T10:30:00Z"
}
```

### 2. Webhook Processing Statistics (Requirement 6.3)

Monitor webhook event processing with detailed metrics:

- **Total Events**: Count of all webhook events processed
- **Successful/Failed Events**: Success rate tracking
- **Events by Type**: Breakdown by event type (message_new, callback_query)
- **Profiles Extracted**: Number of profiles found in events
- **Profiles Stored**: Number of profiles successfully cached
- **Average Processing Time**: Performance metrics in milliseconds
- **Errors by Type**: Error categorization for troubleshooting

**Endpoint**: `GET /api/v1/monitoring/webhook/stats?period={hour|day|week|month}`

**Example Response**:
```json
{
  "period": {
    "from": "2024-01-15T00:00:00Z",
    "to": "2024-01-16T00:00:00Z"
  },
  "total_events": 5000,
  "successful_events": 4800,
  "failed_events": 200,
  "events_by_type": {
    "message_new": 3500,
    "callback_query": 1500
  },
  "profiles_extracted": 3000,
  "profiles_stored": 2900,
  "average_processing_time_ms": 150.5,
  "errors_by_type": {
    "validation_error": 150,
    "cache_error": 50
  }
}
```

### 3. Profile Quality Reporting (Requirement 6.4)

Comprehensive data quality analysis with actionable insights:

- **Quality Metrics**: Overall quality score (0-100)
- **Completeness Score**: Percentage of complete profiles
- **Freshness Score**: Recency of profile data
- **Source Breakdown**: Quality metrics per data source
- **Recommended Actions**: Automated suggestions for improvement
- **Data Issues**: Identified problems with severity levels

**Endpoint**: `GET /api/v1/monitoring/profiles/quality`

**Example Response**:
```json
{
  "generated_at": "2024-01-15T10:30:00Z",
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
      "count": 5000,
      "complete_profiles": 4000,
      "average_age_days": 7.5,
      "quality_score": 85.0
    },
    "user_input": {
      "count": 2000,
      "complete_profiles": 1800,
      "average_age_days": 3.0,
      "quality_score": 95.0
    },
    "default": {
      "count": 1000,
      "complete_profiles": 200,
      "average_age_days": 30.0,
      "quality_score": 30.0
    }
  },
  "recommended_actions": [
    "Увеличить покрытие полных имен пользователей через webhook события",
    "Добавить больше возможностей для пользователей указывать свои имена"
  ],
  "data_issues": [
    {
      "type": "incomplete_profiles",
      "description": "Более 50% профилей не содержат полных имен",
      "count": 3000,
      "severity": "high"
    }
  ]
}
```

## Architecture

### Components

1. **MonitoringService Interface** (`internal/domain/monitoring_service.go`)
   - Defines the contract for monitoring operations
   - Supports multiple implementations (Redis, Mock)

2. **RedisMonitoringService** (`internal/infrastructure/monitoring/redis_monitoring.go`)
   - Production implementation using Redis for persistence
   - Stores webhook events with 7-day TTL
   - Aggregates daily statistics with 30-day retention
   - Calculates quality metrics in real-time

3. **MockMonitoringService** (`internal/infrastructure/monitoring/mock_monitoring.go`)
   - In-memory implementation for testing and development
   - Provides realistic mock data for demonstrations

4. **HTTP Handlers** (`internal/infrastructure/http/handler.go`)
   - RESTful endpoints for accessing monitoring data
   - Swagger documentation for all endpoints
   - Query parameter support for time periods

### Data Flow

```
Webhook Event → WebhookHandler → MonitoringService.RecordWebhookEvent()
                                          ↓
                                    Redis Storage
                                          ↓
                              Daily Aggregation
                                          ↓
                        Statistics/Coverage/Quality APIs
```

## Integration with Webhook Processing

The webhook handler automatically records metrics for every processed event:

```go
// Webhook handler records metrics after processing
metric := domain.WebhookEventMetric{
    EventType:      eventType,
    UserID:         userInfo.UserID,
    ProcessedAt:    startTime,
    Success:        processingError == nil,
    ProcessingTime: time.Since(startTime).Milliseconds(),
    ProfileFound:   profileFound,
    ProfileStored:  profileStored,
}

monitoring.RecordWebhookEvent(ctx, metric)
```

## Configuration

Monitoring uses the same Redis configuration as the profile cache:

```bash
# Redis connection
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# Mock mode (for testing without Redis)
MOCK_MODE=true
```

## Testing

Comprehensive test coverage includes:

1. **Unit Tests**
   - Mock monitoring service functionality
   - Webhook event metric validation
   - Time period filtering

2. **Integration Tests**
   - Complete monitoring flow
   - Error handling scenarios
   - Different time period queries

3. **HTTP Tests**
   - Endpoint response structure
   - Query parameter validation
   - Error responses

Run tests:
```bash
# All monitoring tests
go test -v ./internal/infrastructure/monitoring/

# HTTP endpoint tests
go test -v ./internal/infrastructure/http/ -run TestMonitoring

# All tests
go test -v ./...
```

## Usage Examples

### Using the Demo Script

A demonstration script is provided to test the monitoring functionality:

```bash
./test_monitoring_demo.sh
```

This script:
1. Checks service health
2. Retrieves initial statistics
3. Simulates webhook events
4. Shows updated metrics
5. Demonstrates profile retrieval

### Manual API Testing

```bash
# Get webhook statistics for the last day
curl http://localhost:8095/api/v1/monitoring/webhook/stats?period=day | jq

# Get profile coverage metrics
curl http://localhost:8095/api/v1/monitoring/profiles/coverage | jq

# Get profile quality report
curl http://localhost:8095/api/v1/monitoring/profiles/quality | jq

# Get profile statistics
curl http://localhost:8095/api/v1/profiles/stats | jq
```

### Swagger Documentation

Interactive API documentation is available at:
```
http://localhost:8095/swagger/index.html
```

## Performance Considerations

### Redis Storage Strategy

1. **Webhook Events**: Stored with 7-day TTL to limit storage growth
2. **Daily Statistics**: Aggregated counters with 30-day retention
3. **Processing Times**: Limited to last 1000 values per day
4. **Pipeline Operations**: Atomic updates using Redis pipelines

### Query Optimization

1. **Time Period Queries**: Pre-aggregated daily statistics for fast retrieval
2. **Profile Coverage**: Calculated from profile cache statistics
3. **Quality Reports**: Generated on-demand with caching potential

### Timeouts

All monitoring operations have appropriate timeouts:
- Recording events: 2 seconds
- Retrieving statistics: 10 seconds
- Coverage metrics: 15 seconds
- Quality reports: 20 seconds

## Monitoring Best Practices

### 1. Regular Review

- Check webhook statistics daily to ensure events are being processed
- Monitor profile coverage trends to track data collection progress
- Review quality reports weekly to identify data issues

### 2. Alert Thresholds

Consider setting up alerts for:
- Failed webhook events > 5%
- Profile coverage < 70%
- Quality score < 60
- Processing time > 500ms average

### 3. Data Quality Improvement

Use recommended actions from quality reports:
- Increase webhook event coverage
- Encourage user-provided names
- Reduce reliance on default values

## Troubleshooting

### No Statistics Showing

1. Check if Redis is running: `redis-cli ping`
2. Verify webhook events are being received
3. Check service logs for errors
4. Ensure MOCK_MODE is disabled for production

### Incorrect Metrics

1. Verify time period parameters are correct
2. Check Redis data: `redis-cli KEYS "webhook:*"`
3. Review webhook handler logs for processing errors
4. Ensure system time is synchronized

### Performance Issues

1. Monitor Redis memory usage
2. Check for slow queries in logs
3. Consider increasing Redis resources
4. Review timeout configurations

## Future Enhancements

Potential improvements for monitoring:

1. **Real-time Dashboards**: WebSocket-based live metrics
2. **Historical Trends**: Long-term data analysis and visualization
3. **Alerting System**: Automated notifications for issues
4. **Export Capabilities**: CSV/JSON export for external analysis
5. **Custom Metrics**: User-defined monitoring points
6. **Performance Profiling**: Detailed timing breakdowns

## Related Documentation

- [Profile Cache Service](./PROFILE_CACHE_SERVICE.md)
- [Webhook Integration](./README.md)
- [API Reference](./docs/swagger.yaml)
- [Requirements](../.kiro/specs/max-webhook-profile-integration/requirements.md)
- [Design Document](../.kiro/specs/max-webhook-profile-integration/design.md)

## Support

For issues or questions about monitoring:
1. Check the troubleshooting section above
2. Review service logs for error messages
3. Consult the Swagger documentation
4. Run the demo script to verify functionality