# MAX Webhook Configuration Guide

## Overview

This guide explains how to configure the MAX Messenger webhook integration for profile data collection.

## Webhook URL Configuration

### 1. Webhook Endpoint

The MaxBot service exposes a webhook endpoint at:
```
POST http://your-domain:8095/webhook/max
```

### 2. MAX Bot Settings Configuration

To configure the webhook URL in MAX Messenger bot settings:

1. **Access MAX Bot Admin Panel**
   - Log into your MAX Messenger bot administration interface
   - Navigate to bot settings or webhook configuration section

2. **Set Webhook URL**
   ```
   Webhook URL: http://your-domain:8095/webhook/max
   ```
   
   Replace `your-domain` with your actual server domain or IP address.

3. **Configure Webhook Events**
   Enable the following event types:
   - `message_new` - For capturing user profile data from messages
   - `callback_query` - For capturing user profile data from button interactions

4. **Optional: Webhook Secret**
   If your MAX bot supports webhook secrets for security:
   ```bash
   # Set in environment variables
   WEBHOOK_SECRET=your-secure-webhook-secret
   ```

### 3. Local Development Setup

For local development, you can use ngrok or similar tools to expose your local service:

```bash
# Install ngrok
npm install -g ngrok

# Expose local maxbot-service
ngrok http 8095

# Use the generated URL in MAX bot settings
# Example: https://abc123.ngrok.io/webhook/max
```

### 4. Production Deployment

For production deployment:

1. **Use HTTPS**
   ```
   Webhook URL: https://your-domain.com/webhook/max
   ```

2. **Configure Load Balancer/Reverse Proxy**
   Ensure your load balancer or reverse proxy forwards requests to the maxbot-service on port 8095.

3. **Set Environment Variables**
   ```bash
   # Production environment
   MAXBOT_HTTP_PORT=8095
   WEBHOOK_SECRET=your-production-webhook-secret
   MONITORING_ENABLED=true
   ```

## Webhook Event Processing

### Supported Event Types

1. **message_new Event**
   ```json
   {
     "type": "message_new",
     "message": {
       "from": {
         "user_id": "12345",
         "first_name": "Иван",
         "last_name": "Петров"
       },
       "text": "Hello"
     }
   }
   ```

2. **callback_query Event**
   ```json
   {
     "type": "callback_query",
     "callback_query": {
       "user": {
         "user_id": "12345",
         "first_name": "Иван",
         "last_name": "Петров"
       },
       "data": "button_data"
     }
   }
   ```

### Profile Data Extraction

The webhook handler extracts the following profile information:
- `user_id` - Unique MAX user identifier
- `first_name` - User's first name (if available)
- `last_name` - User's last name (if available)

## Monitoring and Alerts

### Profile Quality Monitoring

The system monitors profile data quality and can trigger alerts:

```bash
# Alert threshold for profile completeness (0.0-1.0)
PROFILE_QUALITY_ALERT_THRESHOLD=0.8

# Alert threshold for webhook processing errors (0.0-1.0)
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05
```

### Monitoring Endpoints

Access monitoring data via:
- `GET /monitoring/profile-stats` - Profile collection statistics
- `GET /monitoring/webhook-stats` - Webhook processing statistics
- `GET /monitoring/health` - Service health status

### Log Monitoring

Monitor webhook processing in logs:
```bash
# View webhook processing logs
docker logs maxbot-service | grep "webhook"

# Monitor profile cache operations
docker logs maxbot-service | grep "profile_cache"
```

## Troubleshooting

### Common Issues

1. **Webhook Not Receiving Events**
   - Verify webhook URL is accessible from MAX servers
   - Check firewall and network configuration
   - Ensure service is running on correct port (8095)

2. **Profile Data Not Being Cached**
   - Verify Redis connection: `REDIS_ADDR=redis:6379`
   - Check Redis service health: `docker logs redis`
   - Monitor profile cache logs

3. **Authentication Issues**
   - Verify webhook secret configuration if used
   - Check MAX bot token configuration

### Health Checks

```bash
# Check service health
curl http://localhost:8095/health

# Check Redis connectivity
curl http://localhost:8095/monitoring/cache-health

# Test webhook endpoint
curl -X POST http://localhost:8095/webhook/max \
  -H "Content-Type: application/json" \
  -d '{"type":"message_new","message":{"from":{"user_id":"test","first_name":"Test"}}}'
```

## Security Considerations

1. **Use HTTPS in Production**
   Always use HTTPS for webhook URLs in production environments.

2. **Webhook Secret Validation**
   Configure webhook secrets to verify request authenticity.

3. **Rate Limiting**
   Consider implementing rate limiting for webhook endpoints.

4. **Input Validation**
   The service validates all incoming webhook data and handles malformed requests gracefully.

## Configuration Examples

### Development Environment
```bash
# .env file for development
MAXBOT_HTTP_PORT=8095
REDIS_ADDR=localhost:6379
WEBHOOK_SECRET=dev-secret
MONITORING_ENABLED=true
PROFILE_QUALITY_ALERT_THRESHOLD=0.7
```

### Production Environment
```bash
# Production environment variables
MAXBOT_HTTP_PORT=8095
REDIS_ADDR=redis-cluster:6379
REDIS_PASSWORD=secure-redis-password
WEBHOOK_SECRET=production-webhook-secret
MONITORING_ENABLED=true
PROFILE_QUALITY_ALERT_THRESHOLD=0.8
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05
```