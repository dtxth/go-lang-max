# Password Logging Implementation

## Overview

Добавлено логирование паролей для новых сотрудников и токенов сброса пароля в консоль для администраторов.

## Changes Made

### 1. Employee Service - Create Employee with Role

**File:** `employee-service/internal/usecase/create_employee_with_role.go`

**Change:** Добавлено логирование сгенерированного пароля при создании нового сотрудника с ролью.

```go
// Логируем сгенерированный пароль для администраторов
log.Printf("Generated password for new employee with phone ending in %s: %s", 
    sanitizePhone(phone), password)
```

**When triggered:** При создании нового сотрудника через Employee Service с назначением роли.

### 2. Auth Service - Create User

**File:** `auth-service/internal/usecase/auth_service.go`

**Change:** Добавлено логирование пароля при создании нового пользователя.

```go
// Логируем сгенерированный пароль для администраторов
fmt.Printf("Generated password for new user with phone ending in %s: %s\n", 
    sanitizePhone(phone), password)
```

**When triggered:** При создании нового пользователя через Auth Service.

### 3. Auth Service - Password Reset Token

**File:** `auth-service/internal/usecase/auth_service.go`

**Change:** Добавлено логирование токена сброса пароля.

```go
// Логируем токен сброса пароля для администраторов
fmt.Printf("Generated password reset token for user with phone ending in %s: %s\n", 
    sanitizePhone(phone), token)
```

**When triggered:** При запросе сброса пароля пользователем.

## Security Considerations

### Phone Number Sanitization

Все логи используют функцию `sanitizePhone()` которая показывает только последние 4 цифры номера телефона:

```go
func sanitizePhone(phone string) string {
    if len(phone) <= 4 {
        return "****"
    }
    return "****" + phone[len(phone)-4:]
}
```

**Example:** `+79161234567` → `****4567`

### Password Visibility

⚠️ **ВАЖНО:** Пароли и токены сброса логируются в открытом виде в консоль.

**Рекомендации по безопасности:**
1. Логи должны быть доступны только администраторам
2. Настройте ротацию логов для удаления старых записей
3. Рассмотрите использование отдельного защищенного лог-файла для паролей
4. В production окружении рассмотрите отключение этого логирования

## Usage Examples

### Creating New Employee

```bash
# В логах employee-service появится:
2024-01-15 10:30:00 Generated password for new employee with phone ending in ****4567: Abc123!@#Def
```

### Creating New User

```bash
# В логах auth-service появится:
Generated password for new user with phone ending in ****4567: Xyz789$%^Ghi
```

### Password Reset Request

```bash
# В логах auth-service появится:
Generated password reset token for user with phone ending in ****4567: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6A7B8C9D0E1F2G3H4
```

## Monitoring and Alerts

### Log Monitoring

Настройте мониторинг логов для отслеживания:
- Частоты создания новых пользователей
- Частоты запросов сброса пароля
- Подозрительной активности

### Example Log Queries

**Grep для поиска паролей:**
```bash
grep "Generated password" /var/log/employee-service.log
grep "Generated password" /var/log/auth-service.log
grep "Generated password reset token" /var/log/auth-service.log
```

**Подсчет создания пользователей за день:**
```bash
grep "Generated password for new" /var/log/*.log | grep "$(date +%Y-%m-%d)" | wc -l
```

## Configuration

### Environment Variables

Для отключения логирования паролей в production (если потребуется):

```bash
# Добавить в будущем
LOG_PASSWORDS=false
```

### Log Rotation

Настройте ротацию логов для безопасности:

```bash
# /etc/logrotate.d/digital-university
/var/log/employee-service.log /var/log/auth-service.log {
    daily
    rotate 7
    compress
    delaycompress
    missingok
    notifempty
    create 0644 app app
}
```

## Testing

### Manual Testing

1. **Create Employee:**
   ```bash
   curl -X POST http://localhost:8081/employees \
     -H "Content-Type: application/json" \
     -d '{
       "phone": "+79161234567",
       "first_name": "Test",
       "last_name": "User",
       "inn": "1234567890",
       "university_name": "Test University",
       "role": "operator"
     }'
   ```

2. **Request Password Reset:**
   ```bash
   curl -X POST http://localhost:8080/auth/password-reset/request \
     -H "Content-Type: application/json" \
     -d '{"phone": "+79161234567"}'
   ```

3. **Check Logs:**
   ```bash
   docker logs employee-service | grep "Generated password"
   docker logs auth-service | grep "Generated password"
   ```

## Related Documentation

- [Password Management Configuration](../auth-service/PASSWORD_MANAGEMENT_CONFIG.md)
- [Password Management API](../auth-service/PASSWORD_MANAGEMENT_API.md)
- [Security Best Practices](./SECURITY_BEST_PRACTICES.md)