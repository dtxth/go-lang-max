# Password Management User Guide

## Overview

This guide explains how to use the password management features in the Digital University system.

## Table of Contents

1. [First Login with Temporary Password](#first-login-with-temporary-password)
2. [Changing Your Password](#changing-your-password)
3. [Resetting a Forgotten Password](#resetting-a-forgotten-password)
4. [Password Requirements](#password-requirements)
5. [Troubleshooting](#troubleshooting)

---

## First Login with Temporary Password

When your account is created by an administrator, you will receive a temporary password via MAX Messenger.

### Step 1: Receive Your Temporary Password

You will receive a MAX Messenger notification that looks like this:

```
Ваш временный пароль для входа в систему: TempPass123!

Рекомендуем сменить пароль после первого входа.
```

**Important Notes:**
- The temporary password is unique to your account
- Keep this password secure and don't share it
- You should change this password after your first login

### Step 2: Log In

1. Open the Digital University application
2. Enter your phone number (e.g., +79991234567)
3. Enter the temporary password you received
4. Click "Login"

### Step 3: Change Your Password (Recommended)

After logging in with your temporary password, you should immediately change it to a password you can remember:

1. Go to your profile settings
2. Click "Change Password"
3. Enter your current (temporary) password
4. Enter your new password
5. Confirm your new password
6. Click "Save"

You will be logged out and need to log in again with your new password.

---

## Changing Your Password

You can change your password at any time while logged in.

### Steps

1. **Log in** to your account
2. **Navigate** to your profile or settings page
3. **Click** "Change Password"
4. **Enter** your current password
5. **Enter** your new password (must meet requirements)
6. **Confirm** your new password
7. **Click** "Save" or "Change Password"

### What Happens After Changing Your Password

- You will be logged out of all devices
- All active sessions will be terminated
- You will need to log in again with your new password

### Example API Request

If you're using the API directly:

```bash
curl -X POST http://localhost:8080/auth/password/change \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "current_password": "OldPassword123!",
    "new_password": "NewSecurePass456!"
  }'
```

---

## Resetting a Forgotten Password

If you forget your password, you can reset it using your phone number.

### Step 1: Request Password Reset

1. Go to the login page
2. Click "Forgot Password?"
3. Enter your phone number
4. Click "Send Reset Code"

**API Example:**
```bash
curl -X POST http://localhost:8080/auth/password-reset/request \
  -H "Content-Type: application/json" \
  -d '{"phone": "+79991234567"}'
```

### Step 2: Receive Reset Token

You will receive a MAX Messenger notification with a reset token:

```
Ваш код для сброса пароля: abc123def456

Код действителен в течение 15 минут.
```

**Important:**
- The reset token expires after 15 minutes
- The token can only be used once
- If you don't receive the token, check your MAX Messenger or request a new one

### Step 3: Reset Your Password

1. Enter the reset token you received
2. Enter your new password (must meet requirements)
3. Confirm your new password
4. Click "Reset Password"

**API Example:**
```bash
curl -X POST http://localhost:8080/auth/password-reset/confirm \
  -H "Content-Type: application/json" \
  -d '{
    "token": "abc123def456",
    "new_password": "NewSecurePass456!"
  }'
```

### Step 4: Log In with New Password

After successfully resetting your password:
1. Go to the login page
2. Enter your phone number
3. Enter your new password
4. Click "Login"

---

## Password Requirements

All passwords must meet the following security requirements:

### Length
- **Minimum:** 12 characters
- **Recommended:** 16+ characters for better security

### Complexity
Your password must contain at least one of each:
- ✅ Uppercase letter (A-Z)
- ✅ Lowercase letter (a-z)
- ✅ Digit (0-9)
- ✅ Special character (!@#$%^&*()_+-=[]{}|;:,.<>?)

### Examples

**✅ Good Passwords:**
- `MySecure#Pass2024`
- `University@Student123`
- `Digital!Vuze456`
- `Temp#Password789`

**❌ Bad Passwords:**
- `password` (too short, no uppercase, no digits, no special chars)
- `Password123` (no special characters)
- `PASSWORD123!` (no lowercase)
- `MyPassword!` (no digits)
- `12345678!@#$` (no letters)

### Password Tips

**DO:**
- Use a unique password for this system
- Use a password manager to generate and store passwords
- Change your password if you suspect it's been compromised
- Use a passphrase (e.g., "Coffee@Morning2024!")

**DON'T:**
- Use personal information (name, birthday, phone number)
- Reuse passwords from other websites
- Share your password with anyone
- Write your password down in an insecure location
- Use common passwords (Password123!, Admin123!, etc.)

---

## Troubleshooting

### I didn't receive my temporary password

**Possible causes:**
- The notification is still being sent (wait 1-2 minutes)
- You haven't started a conversation with the MAX Messenger bot
- Your phone number is incorrect in the system

**Solutions:**
1. Check your MAX Messenger for notifications
2. Wait a few minutes and check again
3. Contact your administrator to verify your phone number
4. Ask your administrator to resend the password

---

### I didn't receive my password reset token

**Possible causes:**
- The token is still being sent (wait 1-2 minutes)
- Your phone number is not registered in the system
- MAX Messenger service is temporarily unavailable

**Solutions:**
1. Wait 1-2 minutes and check MAX Messenger again
2. Verify you entered the correct phone number
3. Try requesting a new reset token
4. Contact your administrator if the problem persists

---

### My reset token expired

**What happened:**
Reset tokens expire after 15 minutes for security reasons.

**Solution:**
1. Go back to the "Forgot Password?" page
2. Request a new reset token
3. Complete the password reset within 15 minutes

---

### My password doesn't meet requirements

**Error message:**
"Password must contain uppercase, lowercase, digit, and special character"

**Solution:**
Make sure your password includes:
- At least one uppercase letter (A-Z)
- At least one lowercase letter (a-z)
- At least one digit (0-9)
- At least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)
- At least 12 characters total

**Example:** `MySecure#Pass2024`

---

### I entered the wrong current password

**Error message:**
"Invalid current password"

**Solutions:**
1. Double-check your current password (check for typos)
2. If you forgot your current password, use the "Forgot Password?" flow instead
3. Make sure Caps Lock is not on

---

### I was logged out after changing my password

**This is expected behavior.**

For security reasons, changing your password logs you out of all devices and terminates all active sessions.

**Solution:**
Simply log in again with your new password.

---

### The system says my password is too weak

Even if your password meets the minimum requirements, the system may reject passwords that are:
- Too common (e.g., "Password123!")
- Too simple (e.g., "Abcd1234!")
- Similar to your personal information

**Solution:**
Use a more complex password or a passphrase:
- `Coffee@Morning2024!`
- `Digital#University456`
- `Secure!Student789`

---

## Security Best Practices

### Keep Your Password Secure

1. **Never share your password** with anyone, including administrators
2. **Don't write it down** in an insecure location
3. **Use a password manager** to store it securely
4. **Change it immediately** if you suspect it's been compromised

### Recognize Phishing Attempts

**Legitimate password reset:**
- You initiated the reset request
- Token is sent via MAX Messenger
- You enter the token on the official website

**Phishing attempt:**
- Unsolicited password reset emails or messages
- Links to unfamiliar websites
- Requests to send your password or token to someone

**If you suspect phishing:**
1. Don't click any links
2. Don't provide your password or token
3. Report it to your administrator
4. Change your password immediately

---

## Getting Help

If you continue to experience issues:

1. **Check the troubleshooting section** above
2. **Contact your system administrator** with:
   - Your phone number
   - Description of the problem
   - Any error messages you received
3. **Check system status** - there may be a known outage

---

## Frequently Asked Questions

### How often should I change my password?

We recommend changing your password:
- Immediately after receiving a temporary password
- Every 90 days for security
- Immediately if you suspect it's been compromised

### Can I reuse an old password?

While the system doesn't currently prevent this, we strongly recommend using a new, unique password each time.

### What if I forget my password multiple times?

You can request password resets as many times as needed. However, frequent password resets may indicate:
- You need a more memorable password
- You should use a password manager
- Your account may be compromised (contact your administrator)

### Can administrators see my password?

No. Passwords are encrypted using bcrypt hashing. Even administrators cannot see your password. They can only reset it for you.

### Why do I need such a complex password?

Complex passwords protect:
- Your personal information
- University data you have access to
- The entire system from unauthorized access

Strong passwords are essential for maintaining security in an educational environment.

---

## Related Documentation

- [API Documentation](./PASSWORD_MANAGEMENT_API.md) - For developers
- [Configuration Guide](./PASSWORD_MANAGEMENT_CONFIG.md) - For administrators
- [Troubleshooting Guide](./PASSWORD_MANAGEMENT_TROUBLESHOOTING.md) - Detailed troubleshooting
