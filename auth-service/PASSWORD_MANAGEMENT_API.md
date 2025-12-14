# Password Management API Documentation

## Overview

The Auth Service provides secure password management functionality including:
- Cryptographically secure random password generation
- Password delivery via MAX Messenger
- Self-service password reset with time-limited tokens
- User-initiated password changes
- Comprehensive audit logging

## Table of Contents

1. [REST API Endpoints](#rest-api-endpoints)
2. [gRPC API](#grpc-api)
3. [Password Requirements](#password-requirements)
4. [Security Features](#security-features)
5. [Error Codes](#error-codes)

---

## REST API Endpoints

### 1. Request Password Reset

**Endpoint:** `POST /auth/password-reset/request`

**Description:** Generates a time-limited reset token and sends it to the user's phone via MAX Messenger.

**Request Body:**
```json
{
  "phone": "+79991234567"
}
```

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password reset token sent to your phone"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid phone number or missing field
- `404 Not Found` - User not found
- `500 Internal Server Error` - Failed to generate token or send notification

**Example:**
```bash
curl -X POST http://localhost:8080/auth/password-reset/request \
  -H "Content-Type: application/json" \
  -d '{"phone": "+79991234567"}'
```

**Notes:**
- Reset tokens expire after 15 minutes (configurable)
- Tokens are single-use only
- User will receive a MAX Messenger notification with the token

---

### 2. Reset Password with Token

**Endpoint:** `POST /auth/password-reset/confirm`

**Description:** Validates the reset token and updates the user's password.

**Request Body:**
```json
{
  "token": "abc123def456",
  "new_password": "NewSecurePass123!"
}
```

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password successfully reset"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid token format or password doesn't meet requirements
- `401 Unauthorized` - Invalid or expired token
- `500 Internal Server Error` - Failed to update password

**Example:**
```bash
curl -X POST http://localhost:8080/auth/password-reset/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123def456",
    "new_password": "NewSecurePass123!"
  }'
```

**Notes:**
- Token is invalidated after successful use
- All existing refresh tokens are revoked
- Password must meet minimum security requirements

---

### 3. Change Password (Authenticated)

**Endpoint:** `POST /auth/password/change`

**Description:** Allows an authenticated user to change their password.

**Authentication:** Required (Bearer token)

**Request Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "current_password": "OldPassword123!",
  "new_password": "NewSecurePass123!"
}
```

**Success Response:** `200 OK`
```json
{
  "success": true,
  "message": "Password successfully changed"
}
```

**Error Responses:**
- `400 Bad Request` - Missing fields or password doesn't meet requirements
- `401 Unauthorized` - Invalid current password or missing authentication
- `500 Internal Server Error` - Failed to update password

**Example:**
```bash
curl -X POST http://localhost:8080/auth/password/change \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGc..." \
  -d '{
    "current_password": "OldPassword123!",
    "new_password": "NewSecurePass123!"
  }'
```

**Notes:**
- Current password must be correct
- All existing refresh tokens are revoked after password change
- User must re-authenticate after password change

---

## gRPC API

### 1. RequestPasswordReset

**Service:** `AuthService`

**Method:** `RequestPasswordReset`

**Request:**
```protobuf
message RequestPasswordResetRequest {
  string phone = 1;
}
```

**Response:**
```protobuf
message RequestPasswordResetResponse {
  bool success = 1;
  string error = 2;
}
```

**Example (Go):**
```go
resp, err := client.RequestPasswordReset(ctx, &proto.RequestPasswordResetRequest{
    Phone: "+79991234567",
})
if err != nil {
    log.Fatal(err)
}
if !resp.Success {
    log.Printf("Error: %s", resp.Error)
}
```

---

### 2. ResetPassword

**Service:** `AuthService`

**Method:** `ResetPassword`

**Request:**
```protobuf
message ResetPasswordRequest {
  string token = 1;
  string new_password = 2;
}
```

**Response:**
```protobuf
message ResetPasswordResponse {
  bool success = 1;
  string error = 2;
}
```

**Example (Go):**
```go
resp, err := client.ResetPassword(ctx, &proto.ResetPasswordRequest{
    Token:       "abc123def456",
    NewPassword: "NewSecurePass123!",
})
if err != nil {
    log.Fatal(err)
}
if !resp.Success {
    log.Printf("Error: %s", resp.Error)
}
```

---

### 3. ChangePassword

**Service:** `AuthService`

**Method:** `ChangePassword`

**Request:**
```protobuf
message ChangePasswordRequest {
  int64 user_id = 1;
  string current_password = 2;
  string new_password = 3;
}
```

**Response:**
```protobuf
message ChangePasswordResponse {
  bool success = 1;
  string error = 2;
}
```

**Example (Go):**
```go
resp, err := client.ChangePassword(ctx, &proto.ChangePasswordRequest{
    UserId:          123,
    CurrentPassword: "OldPassword123!",
    NewPassword:     "NewSecurePass123!",
})
if err != nil {
    log.Fatal(err)
}
if !resp.Success {
    log.Printf("Error: %s", resp.Error)
}
```

---

## Password Requirements

All passwords (temporary and user-set) must meet the following requirements:

### Minimum Length
- **Default:** 12 characters
- **Configurable via:** `MIN_PASSWORD_LENGTH` environment variable
- **Minimum allowed:** 8 characters

### Complexity Requirements
Passwords must contain at least one of each:
- Uppercase letter (A-Z)
- Lowercase letter (a-z)
- Digit (0-9)
- Special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Examples

**Valid passwords:**
- `SecurePass123!`
- `MyP@ssw0rd2024`
- `Temp#Pass456`

**Invalid passwords:**
- `short` (too short)
- `alllowercase123!` (no uppercase)
- `ALLUPPERCASE123!` (no lowercase)
- `NoDigitsHere!` (no digits)
- `NoSpecialChars123` (no special characters)

---

## Security Features

### 1. Password Generation
- Uses `crypto/rand` for cryptographically secure randomness
- Generates unique passwords for each user
- Automatically meets all complexity requirements

### 2. Password Storage
- All passwords are hashed using bcrypt (cost factor: 10)
- Plaintext passwords are never stored
- Password hashes are never logged

### 3. Reset Token Security
- Tokens are cryptographically random (32 bytes, hex-encoded)
- Tokens expire after 15 minutes (configurable)
- Tokens are single-use only
- Tokens are invalidated after use or expiration

### 4. Session Management
- All refresh tokens are invalidated on password change
- All refresh tokens are invalidated on password reset
- Users must re-authenticate after password operations

### 5. Audit Logging
All password operations are logged with:
- Operation type (create, reset, change)
- User ID
- Timestamp
- Success/failure status
- Sanitized phone number (last 4 digits only)

**Never logged:**
- Plaintext passwords
- Password hashes
- Reset tokens
- Full phone numbers

### 6. Notification Security
- Passwords and tokens are sent via MAX Messenger VIP notifications
- Notification failures are logged without exposing sensitive data
- Phone numbers are sanitized in logs

---

## Error Codes

### HTTP Status Codes

| Code | Description | Common Causes |
|------|-------------|---------------|
| 200 | Success | Operation completed successfully |
| 400 | Bad Request | Invalid input, missing fields, password doesn't meet requirements |
| 401 | Unauthorized | Invalid credentials, expired token, missing authentication |
| 404 | Not Found | User not found |
| 500 | Internal Server Error | Database error, notification service unavailable |
| 502 | Bad Gateway | MaxBot Service unavailable |

### Error Response Format

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Password must be at least 12 characters",
    "request_id": "req_abc123"
  }
}
```

### Common Error Messages

**Password Reset:**
- `"User not found"` - Phone number not registered
- `"Invalid or expired token"` - Token is invalid, expired, or already used
- `"Password must be at least 12 characters"` - Password too short
- `"Password must contain uppercase, lowercase, digit, and special character"` - Password doesn't meet complexity requirements
- `"Failed to send notification"` - MaxBot Service unavailable

**Password Change:**
- `"Invalid current password"` - Current password is incorrect
- `"Authentication required"` - Missing or invalid access token
- `"Password must be at least 12 characters"` - New password too short
- `"Password must contain uppercase, lowercase, digit, and special character"` - New password doesn't meet complexity requirements

---

## Rate Limiting

Currently, there is no rate limiting implemented. Consider implementing rate limiting for password reset requests to prevent abuse:

**Recommended limits:**
- Password reset requests: 3 per hour per phone number
- Password change attempts: 5 per hour per user

---

## Monitoring

### Metrics Endpoint

**Endpoint:** `GET /metrics`

**Response:**
```json
{
  "user_creations": 1234,
  "password_resets": 567,
  "password_changes": 890,
  "notifications_sent": 1801,
  "notifications_failed": 23,
  "tokens_generated": 567,
  "tokens_used": 543,
  "tokens_expired": 24,
  "tokens_invalidated": 0,
  "maxbot_healthy": true,
  "last_health_check": "2024-01-15T10:30:00Z",
  "notification_success_rate": 0.987,
  "notification_failure_rate": 0.013
}
```

### Key Metrics to Monitor

1. **Notification Success Rate** - Should be > 95%
2. **Token Expiration Rate** - High rate may indicate UX issues
3. **Password Reset Frequency** - Spike may indicate attack
4. **MaxBot Health** - Should always be true

---

## See Also

- [User Guide](./PASSWORD_MANAGEMENT_USER_GUIDE.md) - End-user documentation
- [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md) - Configuration options
- [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) - Common issues and solutions
