# Password Management Troubleshooting Guide

## Overview

This guide helps diagnose and resolve common issues with the password management system.

## Table of Contents

1. [Service Startup Issues](#service-startup-issues)
2. [Notification Delivery Issues](#notification-delivery-issues)
3. [Password Reset Issues](#password-reset-issues)
4. [Password Change Issues](#password-change-issues)
5. [Database Issues](#database-issues)
6. [Performance Issues](#performance-issues)
7. [Security Issues](#security-issues)

---

## Service Startup Issues

### Issue: Service fails to start with configuration error

**Symptoms:**
```
Error: MIN_PASSWORD_LENGTH must be at least 8, got 5
```

**Cause:**
Invalid configuration value.

**Solution:**
1. Check environment variables:
   ```bash
   env | grep -E "(MIN_PASSWORD_LENGTH|RESET_TOKEN_EXPIRATION|NOTIFICATION_SERVICE_TYPE|MAXBOT_SERVICE_ADDR)"
   ```

2. Verify values meet requirements:
   - `MIN_PASSWORD_LENGTH` >= 8
   - `RESET_TOKEN_EXPIRATION` >= 1
   - `TOKEN_CLEANUP_INTERVAL` >= 1
   - `NOTIFICATION_SERVICE_TYPE` = "mock" or "max"
   - `MAXBOT_SERVICE_ADDR` set when type is "max"

3. Fix invalid values and restart service

---

### Issue: Service fails to start with database error

**Symptoms:**
```
Error: failed to connect to database: connection refused
```

**Cause:**
Database is not accessible or connection string is incorrect.

**Solution:**
1. Verify database is running:
   ```bash
   docker ps | grep postgres
   ```

2. Test database connection:
   ```bash
   psql $DATABASE_URL -c "SELECT 1"
   ```

3. Check connection string format:
   ```
   postgres://user:password@host:port/database?sslmode=disable
   ```

4. Verify network connectivity:
   ```bash
   nc -zv database-host 5432
   ```

5. Check database logs:
   ```bash
   docker logs postgres-container
   ```

---

### Issue: Service starts but migrations not applied

**Symptoms:**
```
Error: relation "password_reset_tokens" does not exist
```

**Cause:**
Database migrations have not been applied.

**Solution:**
1. Check current migration version:
   ```bash
   psql $DATABASE_URL -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 1"
   ```

2. Apply migrations:
   ```bash
   cd auth-service
   make migrate-up
   ```

3. Verify password_reset_tokens table exists:
   ```bash
   psql $DATABASE_URL -c "\d password_reset_tokens"
   ```

4. Restart service

---

## Notification Delivery Issues

### Issue: Notifications not being sent (Mock Service)

**Symptoms:**
- Users not receiving passwords or reset tokens
- Logs show "MOCK: Would send..." messages

**Cause:**
Service is configured to use mock notification service.

**Solution:**
1. Check notification service type:
   ```bash
   echo $NOTIFICATION_SERVICE_TYPE
   ```

2. If in production, change to "max":
   ```bash
   export NOTIFICATION_SERVICE_TYPE=max
   export MAXBOT_SERVICE_ADDR=maxbot-service:9095
   ```

3. Restart service

**Note:** Mock service is intended for development/testing only.

---

### Issue: Notifications not being sent (MAX Service)

**Symptoms:**
- Users not receiving passwords or reset tokens
- Logs show "Failed to send notification" errors
- Metrics show high notification failure rate

**Diagnosis:**
1. Check MaxBot Service health:
   ```bash
   curl http://localhost:8080/metrics | jq '.maxbot_healthy'
   ```

2. Check notification metrics:
   ```bash
   curl http://localhost:8080/metrics | jq '{
     sent: .notifications_sent,
     failed: .notifications_failed,
     success_rate: .notification_success_rate
   }'
   ```

3. Test MaxBot Service connectivity:
   ```bash
   grpcurl -plaintext $MAXBOT_SERVICE_ADDR list
   ```

**Solutions:**

**If MaxBot Service is down:**
```bash
# Check if MaxBot Service is running
docker ps | grep maxbot

# Check MaxBot Service logs
docker logs maxbot-service

# Restart MaxBot Service
docker restart maxbot-service
```

**If network connectivity issue:**
```bash
# Test connectivity
nc -zv maxbot-service-host 9095

# Check DNS resolution
nslookup maxbot-service-host

# Check firewall rules
iptables -L | grep 9095
```

**If MaxBot Service is healthy but notifications failing:**
1. Check MaxBot Service logs for MAX Messenger API errors
2. Verify MAX Messenger API credentials
3. Check MAX Messenger API status
4. Verify phone numbers are in correct format (+79991234567)

---

### Issue: Notifications delayed

**Symptoms:**
- Users receive notifications several minutes late
- Metrics show increasing notification queue

**Cause:**
- MaxBot Service overloaded
- MAX Messenger API rate limiting
- Network latency

**Solutions:**

1. Check MaxBot Service performance:
   ```bash
   docker stats maxbot-service
   ```

2. Scale MaxBot Service:
   ```bash
   docker-compose up -d --scale maxbot-service=3
   ```

3. Check for rate limiting in MaxBot logs:
   ```bash
   docker logs maxbot-service | grep -i "rate limit"
   ```

4. Implement exponential backoff in MaxBot Service

---

## Password Reset Issues

### Issue: Reset token not found or expired

**Symptoms:**
```
Error: Invalid or expired reset token
```

**Diagnosis:**
1. Check if token exists:
   ```sql
   SELECT id, user_id, expires_at, used_at, created_at 
   FROM password_reset_tokens 
   WHERE token = 'abc123def456';
   ```

2. Check if token is expired:
   ```sql
   SELECT id, expires_at, NOW() as current_time,
          expires_at < NOW() as is_expired
   FROM password_reset_tokens 
   WHERE token = 'abc123def456';
   ```

3. Check if token was already used:
   ```sql
   SELECT id, used_at 
   FROM password_reset_tokens 
   WHERE token = 'abc123def456';
   ```

**Solutions:**

**If token expired:**
- User needs to request a new reset token
- Consider increasing `RESET_TOKEN_EXPIRATION` if users consistently can't complete reset in time

**If token already used:**
- User needs to request a new reset token
- This is expected behavior (tokens are single-use)

**If token not found:**
- User may have entered wrong token
- Token may have been cleaned up
- User needs to request a new reset token

---

### Issue: User not found when requesting reset

**Symptoms:**
```
Error: User not found
```

**Cause:**
Phone number not registered in system.

**Solution:**
1. Verify phone number format:
   ```sql
   SELECT id, phone FROM users WHERE phone = '+79991234567';
   ```

2. Check for phone number variations:
   ```sql
   SELECT id, phone FROM users WHERE phone LIKE '%9991234567%';
   ```

3. If user doesn't exist, they need to be created by an administrator

---

### Issue: Reset token generation fails

**Symptoms:**
```
Error: Failed to generate reset token
```

**Cause:**
- Database connection issue
- Insufficient entropy for random generation
- Database constraint violation

**Diagnosis:**
1. Check database connectivity:
   ```bash
   psql $DATABASE_URL -c "SELECT 1"
   ```

2. Check for duplicate tokens (very rare):
   ```sql
   SELECT token, COUNT(*) 
   FROM password_reset_tokens 
   GROUP BY token 
   HAVING COUNT(*) > 1;
   ```

3. Check service logs for detailed error:
   ```bash
   docker logs auth-service | grep -i "failed to generate"
   ```

**Solutions:**
- Restart service if database connection issue
- Check database constraints are correct
- Verify crypto/rand is working (should never fail on modern systems)

---

## Password Change Issues

### Issue: Current password incorrect

**Symptoms:**
```
Error: Invalid current password
```

**Cause:**
User entered wrong current password.

**Solution:**
1. Verify user is entering correct password
2. Check for common issues:
   - Caps Lock enabled
   - Keyboard layout changed
   - Copy/paste including extra spaces
3. If user forgot password, use password reset flow instead

---

### Issue: New password doesn't meet requirements

**Symptoms:**
```
Error: Password must be at least 12 characters
Error: Password must contain uppercase, lowercase, digit, and special character
```

**Cause:**
Password doesn't meet security requirements.

**Solution:**
1. Check minimum length requirement:
   ```bash
   echo $MIN_PASSWORD_LENGTH
   ```

2. Verify password contains:
   - At least one uppercase letter (A-Z)
   - At least one lowercase letter (a-z)
   - At least one digit (0-9)
   - At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

3. Provide user with password requirements and examples

---

### Issue: User logged out after password change

**Symptoms:**
User complains they were logged out after changing password.

**Cause:**
This is expected behavior for security.

**Explanation:**
When a user changes their password:
1. All refresh tokens are invalidated
2. User is logged out of all devices
3. User must log in again with new password

**This is intentional** to prevent:
- Unauthorized access if password was compromised
- Old sessions remaining active with old password

**Solution:**
Inform user this is expected and they should log in with their new password.

---

## Database Issues

### Issue: Password reset tokens table growing too large

**Symptoms:**
- Slow queries on password_reset_tokens table
- High disk usage
- Database performance degradation

**Diagnosis:**
1. Check table size:
   ```sql
   SELECT 
     pg_size_pretty(pg_total_relation_size('password_reset_tokens')) as total_size,
     COUNT(*) as row_count
   FROM password_reset_tokens;
   ```

2. Check expired tokens:
   ```sql
   SELECT COUNT(*) as expired_count
   FROM password_reset_tokens
   WHERE expires_at < NOW();
   ```

3. Check old used tokens:
   ```sql
   SELECT COUNT(*) as old_used_count
   FROM password_reset_tokens
   WHERE used_at IS NOT NULL 
     AND used_at < NOW() - INTERVAL '24 hours';
   ```

**Solutions:**

1. **Verify cleanup job is running:**
   ```bash
   docker logs auth-service | grep -i "cleanup"
   ```

2. **Check cleanup interval:**
   ```bash
   echo $TOKEN_CLEANUP_INTERVAL
   ```

3. **Manually clean up expired tokens:**
   ```sql
   DELETE FROM password_reset_tokens 
   WHERE expires_at < NOW();
   ```

4. **Manually clean up old used tokens:**
   ```sql
   DELETE FROM password_reset_tokens 
   WHERE used_at IS NOT NULL 
     AND used_at < NOW() - INTERVAL '24 hours';
   ```

5. **Increase cleanup frequency:**
   ```bash
   export TOKEN_CLEANUP_INTERVAL=30  # Run every 30 minutes
   ```

6. **Add database maintenance:**
   ```sql
   VACUUM ANALYZE password_reset_tokens;
   ```

---

### Issue: Database connection pool exhausted

**Symptoms:**
```
Error: could not obtain connection from pool
```

**Cause:**
Too many concurrent requests or connections not being released.

**Diagnosis:**
1. Check active connections:
   ```sql
   SELECT COUNT(*) FROM pg_stat_activity 
   WHERE datname = 'auth_db';
   ```

2. Check connection pool settings:
   ```bash
   docker logs auth-service | grep -i "pool"
   ```

**Solutions:**
1. Increase connection pool size in database configuration
2. Check for connection leaks in application code
3. Scale Auth Service horizontally
4. Optimize slow queries

---

## Performance Issues

### Issue: Slow password hashing

**Symptoms:**
- Login/register/password change operations are slow
- High CPU usage during password operations

**Cause:**
Bcrypt cost factor may be too high for your hardware.

**Diagnosis:**
1. Measure password hashing time:
   ```bash
   time curl -X POST http://localhost:8080/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"Test123!"}'
   ```

2. Check CPU usage:
   ```bash
   docker stats auth-service
   ```

**Solutions:**
1. **Adjust bcrypt cost factor** (requires code change):
   - Current: 10 (default)
   - Lower for faster hashing: 8-9
   - Higher for more security: 11-12

2. **Scale horizontally:**
   ```bash
   docker-compose up -d --scale auth-service=3
   ```

3. **Use faster hardware:**
   - More CPU cores
   - Faster CPU

**Note:** Bcrypt is intentionally slow for security. Don't reduce cost factor below 8.

---

### Issue: High memory usage

**Symptoms:**
- Auth Service using excessive memory
- Out of memory errors

**Diagnosis:**
```bash
docker stats auth-service
```

**Solutions:**
1. Check for memory leaks in logs
2. Restart service periodically
3. Increase memory limits:
   ```yaml
   services:
     auth-service:
       deploy:
         resources:
           limits:
             memory: 512M
   ```

---

## Security Issues

### Issue: Suspicious password reset activity

**Symptoms:**
- Spike in password reset requests
- Multiple reset requests for same user
- Reset requests for many users in short time

**Diagnosis:**
1. Check metrics:
   ```bash
   curl http://localhost:8080/metrics | jq '.password_resets'
   ```

2. Check audit logs:
   ```sql
   SELECT user_id, COUNT(*) as reset_count
   FROM password_reset_tokens
   WHERE created_at > NOW() - INTERVAL '1 hour'
   GROUP BY user_id
   ORDER BY reset_count DESC
   LIMIT 10;
   ```

3. Check for patterns:
   ```sql
   SELECT DATE_TRUNC('minute', created_at) as minute,
          COUNT(*) as reset_count
   FROM password_reset_tokens
   WHERE created_at > NOW() - INTERVAL '1 hour'
   GROUP BY minute
   ORDER BY minute DESC;
   ```

**Solutions:**

1. **Implement rate limiting:**
   - Limit reset requests per phone number
   - Limit reset requests per IP address

2. **Block suspicious IPs:**
   ```bash
   iptables -A INPUT -s suspicious-ip -j DROP
   ```

3. **Investigate affected users:**
   ```sql
   SELECT u.id, u.phone, COUNT(prt.id) as reset_count
   FROM users u
   JOIN password_reset_tokens prt ON u.id = prt.user_id
   WHERE prt.created_at > NOW() - INTERVAL '1 hour'
   GROUP BY u.id, u.phone
   HAVING COUNT(prt.id) > 5;
   ```

4. **Contact affected users** to verify legitimate activity

---

### Issue: Passwords appearing in logs

**Symptoms:**
Plaintext passwords visible in application logs.

**Cause:**
**CRITICAL SECURITY ISSUE** - This should never happen.

**Immediate Actions:**
1. **Stop the service immediately**
2. **Rotate all affected passwords**
3. **Review code for logging statements**
4. **Check for:**
   - Debug logging enabled in production
   - Logging request bodies
   - Logging error details with sensitive data

**Prevention:**
1. Never log request bodies containing passwords
2. Sanitize all logs
3. Use structured logging
4. Review logs regularly for sensitive data

---

### Issue: Weak passwords being accepted

**Symptoms:**
Users able to set passwords that don't meet requirements.

**Cause:**
Validation not working correctly.

**Diagnosis:**
1. Test password validation:
   ```bash
   # Should fail
   curl -X POST http://localhost:8080/auth/password-reset/confirm \
     -H "Content-Type: application/json" \
     -d '{"token":"test","new_password":"weak"}'
   ```

2. Check MIN_PASSWORD_LENGTH:
   ```bash
   echo $MIN_PASSWORD_LENGTH
   ```

**Solutions:**
1. Verify validation logic in code
2. Check configuration is correct
3. Add additional validation if needed
4. Force password reset for users with weak passwords

---

## Diagnostic Commands

### Check Service Health

```bash
# HTTP health check
curl http://localhost:8080/health

# Check metrics
curl http://localhost:8080/metrics | jq '.'

# Check logs
docker logs auth-service --tail 100

# Check resource usage
docker stats auth-service
```

### Check Database

```bash
# Test connection
psql $DATABASE_URL -c "SELECT 1"

# Check migrations
psql $DATABASE_URL -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 5"

# Check password reset tokens
psql $DATABASE_URL -c "SELECT COUNT(*) as total, 
  COUNT(*) FILTER (WHERE expires_at < NOW()) as expired,
  COUNT(*) FILTER (WHERE used_at IS NOT NULL) as used
FROM password_reset_tokens"

# Check recent activity
psql $DATABASE_URL -c "SELECT 
  DATE_TRUNC('hour', created_at) as hour,
  COUNT(*) as tokens_created
FROM password_reset_tokens
WHERE created_at > NOW() - INTERVAL '24 hours'
GROUP BY hour
ORDER BY hour DESC"
```

### Check MaxBot Service

```bash
# Check if running
docker ps | grep maxbot

# Check logs
docker logs maxbot-service --tail 100

# Test gRPC connectivity
grpcurl -plaintext $MAXBOT_SERVICE_ADDR list

# Check health from Auth Service
curl http://localhost:8080/metrics | jq '{
  maxbot_healthy: .maxbot_healthy,
  last_health_check: .last_health_check
}'
```

---

## Getting Help

If you can't resolve the issue:

1. **Gather diagnostic information:**
   - Service logs
   - Database logs
   - Configuration
   - Metrics snapshot
   - Steps to reproduce

2. **Check documentation:**
   - [API Documentation](./PASSWORD_MANAGEMENT_API.md)
   - [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md)
   - [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md)

3. **Contact support** with:
   - Description of the issue
   - Error messages
   - Diagnostic information
   - What you've tried

---

## Common Error Messages Reference

| Error Message | Cause | Solution |
|---------------|-------|----------|
| `User not found` | Phone number not registered | Create user account |
| `Invalid or expired token` | Token expired or already used | Request new token |
| `Invalid current password` | Wrong password entered | Verify password or use reset |
| `Password must be at least 12 characters` | Password too short | Use longer password |
| `Password must contain uppercase, lowercase, digit, and special character` | Password doesn't meet complexity | Add required character types |
| `Failed to send notification` | MaxBot Service unavailable | Check MaxBot Service |
| `Authentication required` | Missing or invalid access token | Log in again |
| `MIN_PASSWORD_LENGTH must be at least 8` | Invalid configuration | Fix configuration |
| `MAXBOT_SERVICE_ADDR is required when NOTIFICATION_SERVICE_TYPE is 'max'` | Missing configuration | Set MAXBOT_SERVICE_ADDR |

---

## Related Documentation

- [API Documentation](./PASSWORD_MANAGEMENT_API.md) - API reference
- [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md) - End-user documentation
- [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md) - Configuration options
