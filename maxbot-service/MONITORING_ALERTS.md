# Profile Data Quality Monitoring & Alerts

## Overview

This document describes the monitoring and alerting system for profile data quality in the MAX webhook integration.

## Monitoring Metrics

### Profile Collection Metrics

1. **Profile Completeness Rate**
   - Percentage of users with complete profile data (first_name + last_name)
   - Target: > 80% (configurable via `PROFILE_QUALITY_ALERT_THRESHOLD`)

2. **Webhook Processing Success Rate**
   - Percentage of successfully processed webhook events
   - Target: > 95% (configurable via `WEBHOOK_ERROR_ALERT_THRESHOLD`)

3. **Profile Source Distribution**
   - Breakdown of profile data sources:
     - `webhook` - Data from MAX webhook events
     - `user_input` - Data provided by users explicitly
     - `default` - Default/empty profile data

4. **Cache Performance Metrics**
   - Cache hit rate for profile lookups
   - Cache storage success rate
   - Redis connection health

### Real-time Monitoring Endpoints

```bash
# Profile statistics
GET /monitoring/profile-stats
{
  "total_profiles": 1500,
  "complete_profiles": 1200,
  "completeness_rate": 0.80,
  "source_breakdown": {
    "webhook": 800,
    "user_input": 400,
    "default": 300
  },
  "last_updated": "2024-01-15T10:30:00Z"
}

# Webhook processing statistics
GET /monitoring/webhook-stats
{
  "total_events": 5000,
  "successful_events": 4950,
  "failed_events": 50,
  "success_rate": 0.99,
  "event_types": {
    "message_new": 3000,
    "callback_query": 2000
  },
  "last_24h": {
    "events": 500,
    "success_rate": 0.98
  }
}

# Cache health status
GET /monitoring/cache-health
{
  "redis_connected": true,
  "cache_hit_rate": 0.85,
  "total_operations": 10000,
  "failed_operations": 15,
  "last_error": null
}
```

## Alert Configuration

### Environment Variables

```bash
# Enable monitoring and alerts
MONITORING_ENABLED=true

# Profile quality alert threshold (0.0-1.0)
# Alert when profile completeness drops below this value
PROFILE_QUALITY_ALERT_THRESHOLD=0.8

# Webhook error alert threshold (0.0-1.0)
# Alert when webhook error rate exceeds this value
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05

# Alert check interval
MONITORING_CHECK_INTERVAL=5m

# Alert notification settings
ALERT_WEBHOOK_URL=https://your-monitoring-system/alerts
ALERT_EMAIL_ENABLED=false
ALERT_SLACK_WEBHOOK=https://hooks.slack.com/your-webhook
```

### Alert Types

1. **Profile Quality Alert**
   ```json
   {
     "alert_type": "profile_quality_low",
     "severity": "warning",
     "message": "Profile completeness rate dropped to 75% (threshold: 80%)",
     "current_rate": 0.75,
     "threshold": 0.80,
     "total_profiles": 1000,
     "complete_profiles": 750,
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

2. **Webhook Error Rate Alert**
   ```json
   {
     "alert_type": "webhook_error_rate_high",
     "severity": "critical",
     "message": "Webhook error rate exceeded 5% (current: 8%)",
     "current_rate": 0.08,
     "threshold": 0.05,
     "failed_events": 40,
     "total_events": 500,
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

3. **Cache Connection Alert**
   ```json
   {
     "alert_type": "cache_connection_failed",
     "severity": "critical",
     "message": "Redis cache connection failed",
     "error": "connection refused",
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

## Monitoring Dashboard Setup

### Grafana Dashboard Configuration

Create a Grafana dashboard with the following panels:

1. **Profile Completeness Rate (Time Series)**
   ```promql
   # Query for profile completeness rate
   rate(profile_completeness_total[5m])
   ```

2. **Webhook Success Rate (Gauge)**
   ```promql
   # Query for webhook success rate
   rate(webhook_success_total[5m]) / rate(webhook_total[5m])
   ```

3. **Profile Source Distribution (Pie Chart)**
   ```promql
   # Query for profile sources
   sum by (source) (profile_source_total)
   ```

4. **Cache Performance (Time Series)**
   ```promql
   # Cache hit rate
   rate(cache_hits_total[5m]) / rate(cache_operations_total[5m])
   ```

### Prometheus Metrics

The service exposes the following Prometheus metrics:

```bash
# Profile metrics
profile_total{source="webhook|user_input|default"} - Total profiles by source
profile_completeness_rate - Current profile completeness rate

# Webhook metrics
webhook_events_total{type="message_new|callback_query",status="success|error"} - Webhook events
webhook_processing_duration_seconds - Webhook processing time

# Cache metrics
cache_operations_total{operation="get|set|delete",status="success|error"} - Cache operations
cache_connection_status - Redis connection status (1=connected, 0=disconnected)
```

## Log-based Monitoring

### Structured Logging

The service uses structured JSON logging for monitoring:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "component": "webhook_handler",
  "event": "profile_extracted",
  "user_id": "12345",
  "has_first_name": true,
  "has_last_name": true,
  "source": "webhook",
  "processing_time_ms": 15
}
```

### Log Aggregation Queries

Use these queries in your log aggregation system (ELK, Splunk, etc.):

```bash
# Profile extraction success rate
component:"webhook_handler" AND event:"profile_extracted" 
| stats count by has_first_name, has_last_name

# Webhook processing errors
level:"error" AND component:"webhook_handler"
| timechart span=5m count

# Cache operation failures
component:"profile_cache" AND level:"error"
| stats count by error_type
```

## Alert Response Procedures

### Profile Quality Degradation

1. **Immediate Actions**
   - Check webhook endpoint accessibility
   - Verify MAX bot configuration
   - Review recent webhook events for patterns

2. **Investigation Steps**
   - Analyze profile source distribution
   - Check for changes in MAX API behavior
   - Review webhook event payload structure

3. **Remediation**
   - Update webhook event parsing if needed
   - Implement fallback profile collection methods
   - Notify users about profile completion

### Webhook Processing Failures

1. **Immediate Actions**
   - Check service health and logs
   - Verify Redis connectivity
   - Test webhook endpoint manually

2. **Investigation Steps**
   - Analyze failed webhook payloads
   - Check for rate limiting or throttling
   - Review error patterns and frequencies

3. **Remediation**
   - Fix parsing errors or validation issues
   - Implement retry mechanisms
   - Scale service if needed

### Cache Connection Issues

1. **Immediate Actions**
   - Check Redis service status
   - Verify network connectivity
   - Enable fallback mode (continue without cache)

2. **Investigation Steps**
   - Review Redis logs and metrics
   - Check connection pool configuration
   - Analyze network issues

3. **Remediation**
   - Restart Redis service if needed
   - Adjust connection parameters
   - Implement circuit breaker pattern

## Monitoring Best Practices

1. **Set Appropriate Thresholds**
   - Start with conservative thresholds (80% profile quality, 5% error rate)
   - Adjust based on historical data and business requirements

2. **Monitor Trends, Not Just Absolute Values**
   - Track rate of change in metrics
   - Set up alerts for sudden drops or spikes

3. **Implement Gradual Alerting**
   - Warning alerts at 75% profile quality
   - Critical alerts at 70% profile quality

4. **Regular Review and Tuning**
   - Weekly review of alert frequency and accuracy
   - Monthly adjustment of thresholds based on data

5. **Documentation and Runbooks**
   - Maintain up-to-date runbooks for each alert type
   - Document common issues and solutions

## Integration with External Systems

### Slack Notifications

```bash
# Configure Slack webhook for alerts
ALERT_SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
ALERT_SLACK_CHANNEL=#monitoring
```

### Email Alerts

```bash
# Configure email alerts
ALERT_EMAIL_ENABLED=true
ALERT_EMAIL_SMTP_HOST=smtp.company.com
ALERT_EMAIL_FROM=monitoring@company.com
ALERT_EMAIL_TO=devops@company.com,product@company.com
```

### PagerDuty Integration

```bash
# Configure PagerDuty for critical alerts
ALERT_PAGERDUTY_ENABLED=true
ALERT_PAGERDUTY_INTEGRATION_KEY=your-integration-key
ALERT_PAGERDUTY_SEVERITY_THRESHOLD=critical
```