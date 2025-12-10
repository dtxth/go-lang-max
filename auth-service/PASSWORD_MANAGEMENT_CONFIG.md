# Password Management Configuration Guide

## Overview

This guide explains how to configure the password management features in the Auth Service.

## Table of Contents

1. [Environment Variables](#environment-variables)
2. [Notification Service Configuration](#notification-service-configuration)
3. [Password Policy Configuration](#password-policy-configuration)
4. [Token Configuration](#token-configuration)
5. [Database Configuration](#database-configuration)
6. [Deployment Examples](#deployment-examples)
7. [Configuration Validation](#configuration-validation)

---

## Environment Variables

### Required Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `DATABASE_URL` | PostgreSQL connection string | Yes | - |
| `ACCESS_SECRET` | JWT access token secret | Yes | - |
| `REFRESH_SECRET` | JWT refresh token secret | Yes | - |

### Password Management Variables

| Variable | Description | Required | Default | Valid Values |
|----------|-------------|----------|---------|--------------|
| `MIN_PASSWORD_LENGTH` | Minimum password length | No | 12 | 8-128 |
| `RESET_TOKEN_EXPIRATION` | Token expiration in minutes | No | 15 | 1-1440 |
| `TOKEN_CLEANUP_INTERVAL` | Cleanup interval in minutes | No | 60 | 1-1440 |
| `NOTIFICATION_SERVICE_TYPE` | Notification service type | No | mock | mock, max |
| `MAXBOT_SERVICE_ADDR` | MaxBot gRPC address | Conditional* | - | host:port |

\* Required when `NOTIFICATION_SERVICE_TYPE=max`

### Service Configuration Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `PORT` | HTTP server port | No | 8080 |
| `GRPC_PORT` | gRPC server port | No | 9090 |

---

## Notification Service Configuration

The Auth Service supports two notification service implementations:

### 1. Mock Notification Service (Development/Testing)

**Configuration:**
```bash
NOTIFICATION_SERVICE_TYPE=mock
```

**Behavior:**
- Logs notification attempts to console
- Does not send actual MAX Messenger notifications
- Sanitizes sensitive data in logs
- Always returns success

**Use Cases:**
- Local development
- Unit testing
- Integration testing without external dependencies
- Staging environments without MAX Messenger access

**Example Log Output:**
```
2024-01-15 10:30:00 INFO: MOCK: Would send password notification to phone ending in 4567
2024-01-15 10:30:05 INFO: MOCK: Would send reset token to phone ending in 4567
```

---

### 2. MAX Notification Service (Production)

**Configuration:**
```bash
NOTIFICATION_SERVICE_TYPE=max
MAXBOT_SERVICE_ADDR=maxbot-service:9095
```

**Behavior:**
- Sends actual MAX Messenger VIP notifications
- Requires MaxBot Service to be running
- Includes retry logic for transient failures
- Logs failures without exposing sensitive data

**Use Cases:**
- Production environments
- Staging environments with MAX Messenger access
- End-to-end testing

**Requirements:**
- MaxBot Service must be running and accessible
- MaxBot Service must be configured with MAX Messenger API credentials
- Network connectivity between Auth Service and MaxBot Service

**Health Check:**
The Auth Service periodically checks MaxBot Service health. Monitor the `/metrics` endpoint:
```json
{
  "maxbot_healthy": true,
  "last_health_check": "2024-01-15T10:30:00Z"
}
```

---

## Password Policy Configuration

### Minimum Password Length

**Variable:** `MIN_PASSWORD_LENGTH`

**Default:** 12 characters

**Valid Range:** 8-128 characters

**Recommendations:**
- **Development:** 8 (for easier testing)
- **Staging:** 12 (matches production)
- **Production:** 12-16 (balance security and usability)

**Example:**
```bash
# Require 16-character passwords
MIN_PASSWORD_LENGTH=16
```

**Impact:**
- Affects all password operations (creation, reset, change)
- Applies to both temporary and user-set passwords
- Validated at application startup

---

### Password Complexity

Password complexity requirements are **not configurable** and always include:
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one digit (0-9)
- At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

**Rationale:**
These requirements are based on security best practices and should not be weakened.

---

## Token Configuration

### Reset Token Expiration

**Variable:** `RESET_TOKEN_EXPIRATION`

**Default:** 15 minutes

**Valid Range:** 1-1440 minutes (1 minute to 24 hours)

**Recommendations:**
- **Production:** 15 minutes (balance security and usability)
- **Development:** 60 minutes (easier testing)
- **High-security environments:** 5 minutes

**Example:**
```bash
# Tokens expire after 5 minutes
RESET_TOKEN_EXPIRATION=5
```

**Considerations:**
- Shorter expiration = more secure, but users may not complete reset in time
- Longer expiration = more convenient, but increases security risk
- Users can always request a new token if theirs expires

---

### Token Cleanup Interval

**Variable:** `TOKEN_CLEANUP_INTERVAL`

**Default:** 60 minutes

**Valid Range:** 1-1440 minutes

**Recommendations:**
- **Production:** 60 minutes (hourly cleanup)
- **High-traffic systems:** 30 minutes (more frequent cleanup)
- **Low-traffic systems:** 120 minutes (less frequent cleanup)

**Example:**
```bash
# Clean up expired tokens every 30 minutes
TOKEN_CLEANUP_INTERVAL=30
```

**What Gets Cleaned Up:**
- Expired reset tokens
- Used reset tokens older than 24 hours

**Impact:**
- More frequent cleanup = less database bloat, more CPU usage
- Less frequent cleanup = more database bloat, less CPU usage

---

## Database Configuration

### Connection String

**Variable:** `DATABASE_URL`

**Format:** `postgres://user:password@host:port/database?sslmode=disable`

**Example:**
```bash
DATABASE_URL=postgres://auth_user:secure_password@localhost:5432/auth_db?sslmode=disable
```

**SSL Mode Options:**
- `disable` - No SSL (development only)
- `require` - Require SSL (production)
- `verify-ca` - Verify CA certificate
- `verify-full` - Verify CA and hostname

**Production Example:**
```bash
DATABASE_URL=postgres://auth_user:secure_password@db.example.com:5432/auth_db?sslmode=require
```

---

### Database Migrations

The password reset functionality requires migration `000005_add_password_reset_tokens.up.sql`.

**Apply migrations:**
```bash
cd auth-service
make migrate-up
```

**Verify migration:**
```bash
psql $DATABASE_URL -c "\d password_reset_tokens"
```

**Expected output:**
```
                                      Table "public.password_reset_tokens"
   Column    |            Type             | Collation | Nullable |                      Default
-------------+-----------------------------+-----------+----------+---------------------------------------------------
 id          | bigint                      |           | not null | nextval('password_reset_tokens_id_seq'::regclass)
 user_id     | bigint                      |           | not null |
 token       | character varying(64)       |           | not null |
 expires_at  | timestamp without time zone |           | not null |
 used_at     | timestamp without time zone |           |          |
 created_at  | timestamp without time zone |           |          | now()
Indexes:
    "password_reset_tokens_pkey" PRIMARY KEY, btree (id)
    "password_reset_tokens_token_key" UNIQUE CONSTRAINT, btree (token)
    "idx_password_reset_tokens_expires_at" btree (expires_at)
    "idx_password_reset_tokens_token" btree (token)
```

---

## Deployment Examples

### Development Environment

**docker-compose.yml:**
```yaml
services:
  auth-service:
    image: auth-service:latest
    environment:
      - DATABASE_URL=postgres://auth:password@postgres:5432/auth_dev?sslmode=disable
      - ACCESS_SECRET=dev_access_secret_change_in_production
      - REFRESH_SECRET=dev_refresh_secret_change_in_production
      - PORT=8080
      - GRPC_PORT=9090
      - NOTIFICATION_SERVICE_TYPE=mock
      - MIN_PASSWORD_LENGTH=8
      - RESET_TOKEN_EXPIRATION=60
      - TOKEN_CLEANUP_INTERVAL=60
    ports:
      - "8080:8080"
      - "9090:9090"
```

---

### Staging Environment

**docker-compose.yml:**
```yaml
services:
  auth-service:
    image: auth-service:latest
    environment:
      - DATABASE_URL=postgres://auth:${DB_PASSWORD}@postgres:5432/auth_staging?sslmode=require
      - ACCESS_SECRET=${ACCESS_SECRET}
      - REFRESH_SECRET=${REFRESH_SECRET}
      - PORT=8080
      - GRPC_PORT=9090
      - NOTIFICATION_SERVICE_TYPE=max
      - MAXBOT_SERVICE_ADDR=maxbot-service:9095
      - MIN_PASSWORD_LENGTH=12
      - RESET_TOKEN_EXPIRATION=15
      - TOKEN_CLEANUP_INTERVAL=60
    ports:
      - "8080:8080"
      - "9090:9090"
    depends_on:
      - postgres
      - maxbot-service
```

---

### Production Environment (Kubernetes)

**deployment.yaml:**
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: auth-service
        image: auth-service:v1.0.0
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: auth-secrets
              key: database-url
        - name: ACCESS_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-secrets
              key: access-secret
        - name: REFRESH_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-secrets
              key: refresh-secret
        - name: PORT
          value: "8080"
        - name: GRPC_PORT
          value: "9090"
        - name: NOTIFICATION_SERVICE_TYPE
          value: "max"
        - name: MAXBOT_SERVICE_ADDR
          value: "maxbot-service.default.svc.cluster.local:9095"
        - name: MIN_PASSWORD_LENGTH
          value: "12"
        - name: RESET_TOKEN_EXPIRATION
          value: "15"
        - name: TOKEN_CLEANUP_INTERVAL
          value: "60"
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 9090
          name: grpc
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

---

## Configuration Validation

The Auth Service validates all configuration at startup and fails fast if invalid.

### Validation Rules

1. **MIN_PASSWORD_LENGTH**
   - Must be >= 8
   - Must be <= 128
   - Error: `"MIN_PASSWORD_LENGTH must be at least 8, got X"`

2. **RESET_TOKEN_EXPIRATION**
   - Must be >= 1 minute
   - Must be <= 1440 minutes (24 hours)
   - Error: `"RESET_TOKEN_EXPIRATION must be at least 1 minute, got X"`

3. **TOKEN_CLEANUP_INTERVAL**
   - Must be >= 1 minute
   - Must be <= 1440 minutes (24 hours)
   - Error: `"TOKEN_CLEANUP_INTERVAL must be at least 1 minute, got X"`

4. **NOTIFICATION_SERVICE_TYPE**
   - Must be "mock" or "max"
   - Error: `"NOTIFICATION_SERVICE_TYPE must be 'mock' or 'max', got 'X'"`

5. **MAXBOT_SERVICE_ADDR**
   - Required when NOTIFICATION_SERVICE_TYPE=max
   - Error: `"MAXBOT_SERVICE_ADDR is required when NOTIFICATION_SERVICE_TYPE is 'max'"`

### Testing Configuration

**Test configuration validation:**
```bash
# Valid configuration
docker run --rm \
  -e DATABASE_URL=postgres://user:pass@host:5432/db \
  -e ACCESS_SECRET=secret1 \
  -e REFRESH_SECRET=secret2 \
  -e MIN_PASSWORD_LENGTH=12 \
  auth-service:latest

# Invalid configuration (should fail)
docker run --rm \
  -e DATABASE_URL=postgres://user:pass@host:5432/db \
  -e ACCESS_SECRET=secret1 \
  -e REFRESH_SECRET=secret2 \
  -e MIN_PASSWORD_LENGTH=5 \
  auth-service:latest
# Expected: "MIN_PASSWORD_LENGTH must be at least 8, got 5"
```

---

## Monitoring Configuration

### Metrics

Monitor password management metrics at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

**Key metrics to monitor:**
- `notification_success_rate` - Should be > 0.95
- `notification_failure_rate` - Should be < 0.05
- `maxbot_healthy` - Should be true
- `tokens_expired` - High rate may indicate UX issues

### Alerts

**Recommended alerts:**

1. **Notification Failure Rate > 10%**
   - Check MaxBot Service health
   - Check network connectivity
   - Check MAX Messenger API status

2. **MaxBot Unhealthy**
   - MaxBot Service is down or unreachable
   - Immediate action required

3. **High Token Expiration Rate**
   - Users may not be completing password reset in time
   - Consider increasing RESET_TOKEN_EXPIRATION

4. **Spike in Password Reset Requests**
   - May indicate attack or system issue
   - Review audit logs

---

## Security Considerations

### Secrets Management

**DO:**
- Store secrets in environment variables or secret management systems
- Use different secrets for each environment
- Rotate secrets regularly
- Use strong, random secrets (32+ characters)

**DON'T:**
- Commit secrets to version control
- Use the same secrets across environments
- Use weak or predictable secrets
- Share secrets via insecure channels

### Database Security

**DO:**
- Use SSL/TLS for database connections in production
- Use strong database passwords
- Limit database user permissions
- Enable database audit logging

**DON'T:**
- Use `sslmode=disable` in production
- Use default or weak database passwords
- Grant unnecessary database permissions

### Network Security

**DO:**
- Use TLS for all external communications
- Restrict network access to Auth Service
- Use internal DNS for service-to-service communication
- Implement network segmentation

**DON'T:**
- Expose Auth Service directly to the internet
- Use unencrypted connections
- Allow unrestricted network access

---

## Troubleshooting Configuration Issues

See [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) for detailed troubleshooting steps.

**Quick checks:**

1. **Service won't start:**
   ```bash
   # Check configuration validation
   docker logs auth-service | grep -i error
   ```

2. **Notifications not sending:**
   ```bash
   # Check notification service type
   echo $NOTIFICATION_SERVICE_TYPE
   
   # Check MaxBot address
   echo $MAXBOT_SERVICE_ADDR
   
   # Test MaxBot connectivity
   grpcurl -plaintext $MAXBOT_SERVICE_ADDR list
   ```

3. **Database connection issues:**
   ```bash
   # Test database connection
   psql $DATABASE_URL -c "SELECT 1"
   
   # Check migrations
   psql $DATABASE_URL -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 5"
   ```

---

## Related Documentation

- [API Documentation](./PASSWORD_MANAGEMENT_API.md) - API reference
- [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md) - End-user documentation
- [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) - Detailed troubleshooting
