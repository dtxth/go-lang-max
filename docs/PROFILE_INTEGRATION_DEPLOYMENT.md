# Profile Integration Deployment Guide

## Overview

This guide provides step-by-step instructions for deploying the MAX webhook profile integration system.

## Prerequisites

1. **Docker and Docker Compose** installed
2. **MAX Messenger Bot** configured and accessible
3. **Redis** service available (included in docker-compose.yml)
4. **Domain/IP address** accessible from MAX servers

## Deployment Steps

### 1. Environment Configuration

Copy the example environment file and configure it:

```bash
cp .env.example .env
```

Edit `.env` file with your specific configuration:

```bash
# MAX API Configuration
MAX_API_TOKEN=your-actual-max-api-token
MOCK_MODE=false

# Profile Cache Configuration
REDIS_ADDR=redis:6379
REDIS_PASSWORD=your-redis-password-if-needed
REDIS_DB=1
PROFILE_TTL=720h

# Webhook Configuration
WEBHOOK_SECRET=your-secure-webhook-secret

# Monitoring Configuration
MONITORING_ENABLED=true
PROFILE_QUALITY_ALERT_THRESHOLD=0.8
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05

# Service Ports (adjust if needed)
MAXBOT_HTTP_PORT=8095
EMPLOYEE_SERVICE_PORT=8081

# Profile Cache Integration
PROFILE_CACHE_ENABLED=true
PROFILE_CACHE_TIMEOUT=3s
```

### 2. Build and Deploy Services

Deploy the complete system:

```bash
# Full deployment with tests
make deploy

# Or quick deployment without tests
make deploy-fast
```

Alternatively, use Docker Compose directly:

```bash
# Build and start all services
docker-compose up -d --build

# Check service status
docker-compose ps
```

### 3. Verify Service Health

Check that all services are running correctly:

```bash
# Check service health
make health

# Or manually check each service
curl http://localhost:8095/health  # MaxBot service
curl http://localhost:8081/health  # Employee service
curl http://localhost:6379         # Redis (should connect)
```

### 4. Configure MAX Bot Webhook

#### 4.1 Determine Webhook URL

For production deployment:
```
https://your-domain.com/webhook/max
```

For local development with ngrok:
```bash
# Install and run ngrok
npm install -g ngrok
ngrok http 8095

# Use the generated URL
https://abc123.ngrok.io/webhook/max
```

#### 4.2 Set Webhook in MAX Bot Settings

1. Access your MAX Messenger bot administration panel
2. Navigate to webhook configuration
3. Set the webhook URL: `https://your-domain.com/webhook/max`
4. Enable event types:
   - `message_new`
   - `callback_query`
5. Set webhook secret (if supported): Use the value from `WEBHOOK_SECRET`

### 5. Test Webhook Integration

#### 5.1 Manual Webhook Test

Test the webhook endpoint directly:

```bash
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
```

Expected response: `200 OK`

#### 5.2 Verify Profile Caching

Check that profiles are being cached:

```bash
# Check Redis for cached profile
docker exec -it redis redis-cli
> GET "profile:user:test123"
```

#### 5.3 Test Employee Creation with Profile

Create an employee and verify profile integration:

```bash
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-jwt-token" \
  -d '{
    "phone": "+1234567890",
    "max_id": "test123",
    "university_id": 1
  }'
```

The employee should be created with the cached profile name.

### 6. Monitor System Health

#### 6.1 Check Monitoring Endpoints

```bash
# Profile statistics
curl http://localhost:8095/monitoring/profile-stats

# Webhook statistics
curl http://localhost:8095/monitoring/webhook-stats

# Cache health
curl http://localhost:8095/monitoring/cache-health
```

#### 6.2 View Service Logs

```bash
# View all service logs
make logs

# View specific service logs
docker logs maxbot-service
docker logs employee-service
docker logs redis
```

#### 6.3 Monitor Profile Quality

Check profile completeness rate:

```bash
# Get current profile statistics
curl http://localhost:8095/monitoring/profile-stats | jq '.completeness_rate'
```

### 7. Production Considerations

#### 7.1 Security Configuration

```bash
# Use strong secrets in production
WEBHOOK_SECRET=$(openssl rand -base64 32)
REDIS_PASSWORD=$(openssl rand -base64 32)
ACCESS_SECRET=$(openssl rand -base64 32)
REFRESH_SECRET=$(openssl rand -base64 32)
```

#### 7.2 SSL/TLS Configuration

Ensure HTTPS is configured for webhook endpoints:

```nginx
# Nginx configuration example
server {
    listen 443 ssl;
    server_name your-domain.com;
    
    ssl_certificate /path/to/certificate.crt;
    ssl_certificate_key /path/to/private.key;
    
    location /webhook/max {
        proxy_pass http://localhost:8095;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

#### 7.3 Resource Limits

Configure appropriate resource limits:

```yaml
# docker-compose.override.yml for production
version: '3.8'
services:
  maxbot-service:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: '0.5'
        reservations:
          memory: 256M
          cpus: '0.25'
  
  redis:
    deploy:
      resources:
        limits:
          memory: 256M
          cpus: '0.25'
```

#### 7.4 Backup Configuration

Set up Redis data persistence:

```bash
# Configure Redis persistence in production
REDIS_SAVE_INTERVAL=900 1  # Save if at least 1 key changed in 900 seconds
REDIS_APPENDONLY=yes       # Enable append-only file
```

### 8. Troubleshooting

#### 8.1 Common Issues

**Webhook not receiving events:**
```bash
# Check service accessibility
curl -I http://your-domain:8095/webhook/max

# Check firewall rules
sudo ufw status

# Check service logs
docker logs maxbot-service | grep webhook
```

**Profile data not being cached:**
```bash
# Check Redis connectivity
docker exec -it redis redis-cli ping

# Check profile cache logs
docker logs maxbot-service | grep profile_cache

# Verify Redis configuration
docker logs maxbot-service | grep "Redis connected"
```

**Employee creation not using profiles:**
```bash
# Check profile cache integration
docker logs employee-service | grep profile

# Test profile retrieval
curl http://localhost:8095/profiles/test123
```

#### 8.2 Performance Issues

**High webhook processing latency:**
```bash
# Check webhook processing metrics
curl http://localhost:8095/monitoring/webhook-stats

# Monitor Redis performance
docker exec -it redis redis-cli --latency

# Check service resource usage
docker stats maxbot-service
```

**Redis memory usage:**
```bash
# Check Redis memory usage
docker exec -it redis redis-cli info memory

# Monitor cache hit rate
curl http://localhost:8095/monitoring/cache-health
```

### 9. Monitoring and Alerting Setup

#### 9.1 Configure External Monitoring

Set up external monitoring systems:

```bash
# Prometheus metrics endpoint
curl http://localhost:8095/metrics

# Configure Grafana dashboard
# Import dashboard configuration from monitoring documentation
```

#### 9.2 Set Up Alerts

Configure alert notifications:

```bash
# Slack notifications
ALERT_SLACK_WEBHOOK=https://hooks.slack.com/your-webhook
ALERT_SLACK_CHANNEL=#monitoring

# Email alerts
ALERT_EMAIL_ENABLED=true
ALERT_EMAIL_TO=devops@company.com
```

### 10. Maintenance Procedures

#### 10.1 Regular Maintenance Tasks

```bash
# Weekly: Check profile quality metrics
curl http://localhost:8095/monitoring/profile-stats

# Monthly: Review and clean old profiles
# (Profiles automatically expire based on PROFILE_TTL)

# Quarterly: Review webhook processing patterns
curl http://localhost:8095/monitoring/webhook-stats
```

#### 10.2 Backup Procedures

```bash
# Backup Redis data
docker exec redis redis-cli BGSAVE

# Backup configuration
cp .env .env.backup.$(date +%Y%m%d)
```

#### 10.3 Update Procedures

```bash
# Update services
git pull
make deploy

# Verify health after update
make health
curl http://localhost:8095/monitoring/profile-stats
```

## Rollback Procedures

If issues occur after deployment:

```bash
# Stop services
docker-compose down

# Restore previous configuration
cp .env.backup.YYYYMMDD .env

# Restart with previous version
git checkout previous-tag
make deploy

# Verify rollback success
make health
```

## Support and Documentation

- **Webhook Configuration**: See `maxbot-service/WEBHOOK_CONFIGURATION.md`
- **Monitoring Setup**: See `maxbot-service/MONITORING_ALERTS.md`
- **API Documentation**: Access Swagger UI at `http://localhost:8095/swagger/`
- **Service Logs**: Use `make logs` or `docker logs <service-name>`

For additional support, check the service-specific documentation in each service directory.