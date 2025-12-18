# Migration Guide: From Direct API to Webhook Profile Integration

## Overview

This guide provides step-by-step instructions for migrating from the current direct API approach to the new webhook-based profile integration system. The migration ensures backward compatibility while enabling enhanced profile data collection through MAX Messenger webhooks.

## Migration Strategy

The migration follows a **zero-downtime, backward-compatible** approach:

1. **Phase 1**: Deploy new webhook infrastructure alongside existing system
2. **Phase 2**: Enable webhook collection while maintaining API fallback
3. **Phase 3**: Gradually transition to webhook-primary approach
4. **Phase 4**: Monitor and optimize the new system

## Pre-Migration Assessment

### Current System Analysis

Before starting the migration, assess your current implementation:

#### 1. Current API Usage Patterns

```bash
# Analyze current MAX API usage
grep -r "GetMaxIDByPhone" employee-service/
grep -r "GetUserProfileByPhone" employee-service/

# Check current employee creation patterns
grep -r "POST /employees" integration-tests/
```

#### 2. Data Volume Assessment

```sql
-- Check current employee data
SELECT 
    COUNT(*) as total_employees,
    COUNT(CASE WHEN first_name != 'Неизвестно' THEN 1 END) as employees_with_names,
    COUNT(CASE WHEN max_id IS NOT NULL THEN 1 END) as employees_with_max_id
FROM employees;

-- Check name source distribution (if profile_source column exists)
SELECT profile_source, COUNT(*) 
FROM employees 
WHERE profile_source IS NOT NULL 
GROUP BY profile_source;
```

#### 3. Current Configuration Review

```bash
# Review current environment variables
grep -E "(MAX_API|MAXBOT)" .env

# Check current Docker Compose setup
grep -A 10 -B 5 "maxbot-service" docker-compose.yml
```

## Migration Steps

### Phase 1: Infrastructure Preparation

#### 1.1 Update Environment Configuration

Add new environment variables to `.env`:

```bash
# Add to existing .env file
echo "
# Profile Cache Configuration (NEW)
REDIS_DB=1
PROFILE_TTL=720h
PROFILE_CACHE_ENABLED=true
PROFILE_CACHE_TIMEOUT=3s

# Webhook Configuration (NEW)
MAXBOT_HTTP_PORT=8095
WEBHOOK_SECRET=webhook-secret-change-in-production

# Monitoring Configuration (NEW)
MONITORING_ENABLED=true
PROFILE_QUALITY_ALERT_THRESHOLD=0.8
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05
" >> .env
```

#### 1.2 Update Docker Compose Configuration

**Backup current configuration**:
```bash
cp docker-compose.yml docker-compose.yml.backup
```

**Add Redis dependency and new environment variables**:
```yaml
# Add to maxbot-service in docker-compose.yml
services:
  maxbot-service:
    environment:
      # Existing variables...
      
      # NEW: Profile Cache Configuration
      - REDIS_ADDR=redis:6379
      - REDIS_DB=${REDIS_DB:-1}
      - PROFILE_TTL=${PROFILE_TTL:-720h}
      
      # NEW: Webhook Configuration
      - WEBHOOK_SECRET=${WEBHOOK_SECRET:-}
      
      # NEW: Monitoring Configuration
      - MONITORING_ENABLED=${MONITORING_ENABLED:-true}
      - PROFILE_QUALITY_ALERT_THRESHOLD=${PROFILE_QUALITY_ALERT_THRESHOLD:-0.8}
    
    # NEW: Add Redis dependency
    depends_on:
      redis:
        condition: service_healthy
    
    # NEW: Expose HTTP port for webhooks
    ports:
      - "${MAXBOT_HTTP_PORT:-8095}:8095"

  # Add to employee-service
  employee-service:
    environment:
      # Existing variables...
      
      # NEW: Profile Cache Integration
      - PROFILE_CACHE_ENABLED=${PROFILE_CACHE_ENABLED:-true}
      - PROFILE_CACHE_TIMEOUT=${PROFILE_CACHE_TIMEOUT:-3s}
```

#### 1.3 Database Schema Migration

**Create migration script** (`employee-service/migrations/000005_add_profile_source_tracking.up.sql`):
```sql
-- Add profile source tracking columns
ALTER TABLE employees 
ADD COLUMN IF NOT EXISTS profile_source VARCHAR(20) DEFAULT 'manual',
ADD COLUMN IF NOT EXISTS profile_last_updated TIMESTAMP;

-- Create index for profile source queries
CREATE INDEX IF NOT EXISTS idx_employees_profile_source ON employees(profile_source);

-- Update existing records to have 'manual' source
UPDATE employees 
SET profile_source = 'manual' 
WHERE profile_source IS NULL;
```

**Create rollback script** (`employee-service/migrations/000005_add_profile_source_tracking.down.sql`):
```sql
-- Remove profile source tracking columns
ALTER TABLE employees 
DROP COLUMN IF EXISTS profile_source,
DROP COLUMN IF EXISTS profile_last_updated;

-- Drop index
DROP INDEX IF EXISTS idx_employees_profile_source;
```

#### 1.4 Deploy Infrastructure Changes

```bash
# Apply database migrations
make migrate-up

# Deploy updated services (without webhook enabled yet)
WEBHOOK_ENABLED=false make deploy

# Verify services are running
make health
```

### Phase 2: Enable Webhook Collection

#### 2.1 Configure MAX Bot Webhook

**For Development**:
```bash
# Use ngrok for local development
ngrok http 8095

# Configure webhook URL in MAX bot settings:
# https://abc123.ngrok.io/webhook/max
```

**For Production**:
```bash
# Configure webhook URL in MAX bot settings:
# https://your-domain.com/webhook/max

# Enable event types:
# - message_new
# - callback_query
```

#### 2.2 Enable Webhook Processing

```bash
# Update environment to enable webhook
echo "WEBHOOK_ENABLED=true" >> .env

# Redeploy services
make deploy

# Verify webhook endpoint is accessible
curl -X POST http://localhost:8095/webhook/max \
  -H "Content-Type: application/json" \
  -d '{"type":"message_new","message":{"from":{"user_id":"test","first_name":"Test"}}}'
```

#### 2.3 Monitor Initial Webhook Activity

```bash
# Monitor webhook processing
docker logs -f maxbot-service | grep webhook

# Check profile cache activity
docker logs -f maxbot-service | grep profile_cache

# Monitor Redis for cached profiles
docker exec -it redis redis-cli
> KEYS "profile:user:*"
> GET "profile:user:test"
```

### Phase 3: Gradual Transition

#### 3.1 Enable Profile Cache Integration

**Update employee service configuration**:
```bash
# Enable profile cache integration
echo "PROFILE_CACHE_ENABLED=true" >> .env

# Redeploy employee service
docker-compose up -d employee-service

# Test employee creation with profile integration
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $JWT_TOKEN" \
  -d '{
    "phone": "+1234567890",
    "max_id": "test_user",
    "university_id": 1
  }'
```

#### 3.2 Monitor Profile Usage

```bash
# Check profile statistics
curl http://localhost:8095/monitoring/profiles/coverage

# Monitor employee creation with profiles
docker logs -f employee-service | grep profile

# Check database for profile source tracking
psql $DATABASE_URL -c "
SELECT profile_source, COUNT(*) 
FROM employees 
GROUP BY profile_source;
"
```

#### 3.3 Gradual Rollout Strategy

**Option A: Feature Flag Approach**
```bash
# Enable for specific users/environments first
PROFILE_INTEGRATION_ROLLOUT_PERCENTAGE=25

# Gradually increase percentage
PROFILE_INTEGRATION_ROLLOUT_PERCENTAGE=50
PROFILE_INTEGRATION_ROLLOUT_PERCENTAGE=75
PROFILE_INTEGRATION_ROLLOUT_PERCENTAGE=100
```

**Option B: Service-by-Service Rollout**
```bash
# Enable for specific services first
EMPLOYEE_SERVICE_PROFILE_ENABLED=true
CHAT_SERVICE_PROFILE_ENABLED=false

# Then enable for all services
CHAT_SERVICE_PROFILE_ENABLED=true
```

### Phase 4: Optimization and Monitoring

#### 4.1 Performance Monitoring

```bash
# Monitor webhook processing performance
curl http://localhost:8095/monitoring/webhook/stats

# Monitor cache performance
curl http://localhost:8095/monitoring/cache-health

# Monitor profile quality
curl http://localhost:8095/monitoring/profiles/quality
```

#### 4.2 Alert Configuration

```bash
# Configure monitoring alerts
PROFILE_QUALITY_ALERT_THRESHOLD=0.8
WEBHOOK_ERROR_ALERT_THRESHOLD=0.05

# Set up external monitoring (optional)
PROMETHEUS_ENABLED=true
GRAFANA_DASHBOARD_ENABLED=true
```

#### 4.3 Performance Optimization

**Redis Configuration Tuning**:
```bash
# Optimize Redis for profile caching
REDIS_MAXMEMORY=256mb
REDIS_MAXMEMORY_POLICY=allkeys-lru
REDIS_SAVE_INTERVAL="900 1"
```

**Cache TTL Optimization**:
```bash
# Adjust TTL based on usage patterns
PROFILE_TTL=720h  # 30 days (default)
# or
PROFILE_TTL=168h  # 7 days (more aggressive)
```

## Data Migration Procedures

### Existing Employee Data

**No data migration required** - the system is designed to be backward compatible:

1. **Existing employees** keep their current names and data
2. **Profile source** is set to `'manual'` for existing records
3. **New employees** will use webhook profiles when available
4. **API compatibility** is maintained for all existing integrations

### Profile Data Backfill (Optional)

If you want to backfill profile data for existing employees:

```sql
-- Create a script to backfill profile sources
UPDATE employees 
SET profile_source = 'webhook',
    profile_last_updated = NOW()
WHERE max_id IS NOT NULL 
  AND max_id IN (
    SELECT DISTINCT user_id 
    FROM profile_cache 
    WHERE first_name IS NOT NULL
  );
```

## Rollback Procedures

### Emergency Rollback

If issues occur during migration:

```bash
# 1. Disable webhook processing
echo "WEBHOOK_ENABLED=false" >> .env

# 2. Disable profile cache integration
echo "PROFILE_CACHE_ENABLED=false" >> .env

# 3. Redeploy services
make deploy

# 4. Restore previous configuration
cp docker-compose.yml.backup docker-compose.yml

# 5. Rollback database changes (if needed)
make migrate-down
```

### Partial Rollback

To rollback specific components:

```bash
# Rollback webhook only (keep cache integration)
WEBHOOK_ENABLED=false

# Rollback cache integration only (keep webhook)
PROFILE_CACHE_ENABLED=false

# Rollback monitoring only
MONITORING_ENABLED=false
```

## Testing and Validation

### Pre-Migration Testing

```bash
# Test current system functionality
make test

# Run integration tests
make test-integration

# Test employee creation
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{"phone":"+1234567890","first_name":"Test","last_name":"User","university_id":1}'
```

### Post-Migration Validation

```bash
# Test webhook endpoint
make test-webhook

# Test profile integration
make test-profile-integration

# Validate monitoring endpoints
curl http://localhost:8095/monitoring/profiles/coverage
curl http://localhost:8095/monitoring/webhook/stats

# Test employee creation with profile integration
curl -X POST http://localhost:8081/employees \
  -H "Content-Type: application/json" \
  -d '{"phone":"+1234567890","max_id":"test_user","university_id":1}'
```

### Regression Testing

```bash
# Run full test suite
make test-all

# Test backward compatibility
make test-backward-compatibility

# Performance testing
make test-performance
```

## Migration Checklist

### Pre-Migration
- [ ] Backup current configuration files
- [ ] Backup database
- [ ] Review current API usage patterns
- [ ] Assess data volume and performance requirements
- [ ] Plan rollback procedures

### Infrastructure Setup
- [ ] Update environment variables
- [ ] Update Docker Compose configuration
- [ ] Create database migration scripts
- [ ] Deploy infrastructure changes
- [ ] Verify service health

### Webhook Configuration
- [ ] Configure MAX bot webhook URL
- [ ] Enable webhook event types
- [ ] Test webhook endpoint
- [ ] Monitor initial webhook activity
- [ ] Verify profile caching

### Integration Enablement
- [ ] Enable profile cache integration
- [ ] Test employee creation with profiles
- [ ] Monitor profile usage statistics
- [ ] Validate profile source tracking
- [ ] Check database updates

### Monitoring and Optimization
- [ ] Configure monitoring alerts
- [ ] Set up performance monitoring
- [ ] Optimize cache configuration
- [ ] Validate profile quality metrics
- [ ] Document operational procedures

### Post-Migration
- [ ] Run regression tests
- [ ] Validate backward compatibility
- [ ] Monitor system performance
- [ ] Update documentation
- [ ] Train operations team

## Common Migration Issues

### Issue 1: Webhook Not Receiving Events

**Symptoms**:
- No profile data being cached
- Webhook statistics show zero events

**Solutions**:
```bash
# Check webhook URL accessibility
curl -I http://your-domain:8095/webhook/max

# Verify MAX bot configuration
# Check firewall and network settings
# Review webhook endpoint logs
docker logs maxbot-service | grep webhook
```

### Issue 2: Profile Cache Not Working

**Symptoms**:
- Employee creation not using cached profiles
- Cache statistics show zero hits

**Solutions**:
```bash
# Check Redis connectivity
docker exec -it redis redis-cli ping

# Verify Redis configuration
docker logs maxbot-service | grep redis

# Check profile cache integration
docker logs employee-service | grep profile
```

### Issue 3: Performance Degradation

**Symptoms**:
- Slow employee creation
- High Redis memory usage
- Webhook processing delays

**Solutions**:
```bash
# Monitor Redis performance
docker exec -it redis redis-cli --latency

# Check cache hit rates
curl http://localhost:8095/monitoring/cache-health

# Optimize Redis configuration
REDIS_MAXMEMORY=512mb
REDIS_MAXMEMORY_POLICY=allkeys-lru
```

### Issue 4: Profile Quality Issues

**Symptoms**:
- Low profile completeness rates
- Missing profile data

**Solutions**:
```bash
# Check webhook event processing
curl http://localhost:8095/monitoring/webhook/stats

# Review profile extraction logic
docker logs maxbot-service | grep "profile extracted"

# Validate MAX bot event configuration
```

## Post-Migration Operations

### Daily Operations

```bash
# Check system health
make health

# Monitor profile quality
curl http://localhost:8095/monitoring/profiles/coverage

# Review webhook statistics
curl http://localhost:8095/monitoring/webhook/stats
```

### Weekly Operations

```bash
# Review profile quality trends
curl http://localhost:8095/monitoring/profiles/quality

# Check cache performance
curl http://localhost:8095/monitoring/cache-health

# Review error logs
docker logs maxbot-service | grep ERROR
```

### Monthly Operations

```bash
# Analyze profile source distribution
psql $DATABASE_URL -c "
SELECT profile_source, COUNT(*), 
       ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
FROM employees 
GROUP BY profile_source;
"

# Review and optimize cache TTL
# Update monitoring thresholds based on operational data
# Plan capacity adjustments if needed
```

## Support and Resources

### Documentation References

- **Main Integration Guide**: `MAX_WEBHOOK_PROFILE_INTEGRATION.md`
- **Webhook Configuration**: `maxbot-service/WEBHOOK_CONFIGURATION.md`
- **Monitoring Setup**: `maxbot-service/MONITORING_ALERTS.md`
- **Deployment Guide**: `PROFILE_INTEGRATION_DEPLOYMENT.md`

### Troubleshooting Resources

- **Service Logs**: `make logs`
- **Health Checks**: `make health`
- **Monitoring Endpoints**: See monitoring API documentation
- **Test Commands**: `make test-webhook`, `make test-profile-integration`

### Emergency Contacts

- **Technical Lead**: [Contact Information]
- **Operations Team**: [Contact Information]
- **MAX API Support**: [Contact Information]

For additional support during migration, refer to the troubleshooting sections in the main integration documentation.