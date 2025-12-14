# Рефакторинг: Телефон как основной идентификатор

## Проблема

В текущей реализации Auth Service использует `email` как основной идентификатор пользователя, но:
- Employee Service использует `phone` как основной идентификатор
- Приходилось генерировать искусственные email: `+79001234567@employee.local`
- MAX Messenger также использует телефон
- Пользователи должны входить по телефону, а не по email

## Решение

Заменить `email` на `phone` как основной идентификатор в Auth Service.

## Изменения

### 1. База данных (Auth Service)

**Миграция:** `000004_replace_email_with_phone.up.sql`

```sql
-- Добавляем phone как основной идентификатор
ALTER TABLE users ADD COLUMN phone TEXT;
UPDATE users SET phone = email WHERE phone IS NULL;
ALTER TABLE users ALTER COLUMN phone SET NOT NULL;
CREATE UNIQUE INDEX idx_users_phone ON users(phone);

-- Email становится опциональным
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;
```

### 2. Domain модель User

```go
type User struct {
    ID       int64  `json:"id"`
    Phone    string `json:"phone"`              // Основной идентификатор
    Email    string `json:"email,omitempty"`    // Опциональный
    Password string `json:"-"`
    Role     string `json:"role"`
}
```

### 3. Proto API

**CreateUserRequest:**
```protobuf
message CreateUserRequest {
  string phone = 1;      // Основной идентификатор
  string password = 2;
  string email = 3;      // Опциональный
}
```

**ValidateTokenResponse:**
```protobuf
message ValidateTokenResponse {
  bool valid = 1;
  int64 user_id = 2;
  string phone = 3;      // Вместо email
  string role = 4;
  // ...
  string email = 9;      // Опциональный
}
```

### 4. Repository методы

```go
type UserRepository interface {
    Create(user *User) error
    GetByPhone(phone string) (*User, error)  // Основной метод
    GetByEmail(email string) (*User, error)  // Для обратной совместимости
    GetByID(id int64) (*User, error)
}
```

### 5. AuthService usecase

```go
func (s *AuthService) CreateUser(phone, password string) (int64, error) {
    // Проверяем по телефону
    existingUser, err := s.repo.GetByPhone(phone)
    if err == nil && existingUser != nil && existingUser.ID > 0 {
        return existingUser.ID, nil
    }
    
    user := &domain.User{
        Phone:    phone,
        Email:    "",  // Опциональный
        Password: hashed,
        Role:     "",
    }
    
    s.repo.Create(user)
    return user.ID, nil
}
```

### 6. Employee Service

**CreateEmployeeWithRoleUseCase:**
```go
// Больше не нужно генерировать искусственный email
userID, err := uc.authService.CreateUser(ctx, phone, tempPassword)
```

## Преимущества

1. ✅ **Единый идентификатор** - телефон используется везде
2. ✅ **Нет искусственных данных** - не нужно генерировать fake email
3. ✅ **Проще для пользователей** - вход по телефону
4. ✅ **Соответствие MAX Messenger** - та же логика идентификации
5. ✅ **Меньше путаницы** - один источник правды

## Обратная совместимость

- Email остается в базе данных как опциональное поле
- Метод `GetByEmail` сохранен для старых пользователей
- Существующие пользователи с email продолжат работать

## Миграция существующих данных

Для существующих пользователей с email:
```sql
-- Если у пользователя есть только email, можно оставить его в phone
UPDATE users SET phone = email WHERE phone IS NULL;

-- Или вручную обновить для конкретных пользователей
UPDATE users SET phone = '+79001234567' WHERE email = 'user@example.com';
```

## Вход в систему

Теперь пользователи входят по телефону:

**Старый способ:**
```json
{
  "email": "user@example.com",
  "password": "password"
}
```

**Новый способ:**
```json
{
  "phone": "+79001234567",
  "password": "password"
}
```

## Связанные файлы

- `auth-service/migrations/000004_replace_email_with_phone.up.sql`
- `auth-service/internal/domain/user.go`
- `auth-service/api/proto/auth.proto`
- `auth-service/internal/infrastructure/repository/user_postgres.go`
- `auth-service/internal/usecase/auth_service.go`
- `employee-service/internal/domain/auth_service.go`
- `employee-service/internal/infrastructure/auth/auth_client.go`
- `employee-service/internal/usecase/create_employee_with_role.go`

## Тестирование

После применения изменений:

1. Применить миграцию:
```bash
docker-compose restart auth-db
```

2. Пересобрать сервисы:
```bash
docker-compose build auth-service employee-service
docker-compose restart auth-service employee-service
```

3. Создать сотрудника с ролью:
```bash
curl -X POST 'http://localhost:8081/employees' \
  -H 'Content-Type: application/json' \
  -d '{
  "first_name": "Тест",
  "last_name": "Телефон",
  "phone": "+79007777777",
  "inn": "7777777777",
  "kpp": "777777777",
  "university_name": "Тест",
  "role": "curator"
}'
```

4. Проверить в базе:
```sql
SELECT id, phone, email FROM users WHERE phone = '+79007777777';
```

## Статус

✅ Миграция создана
✅ Domain модели обновлены
✅ Proto файлы обновлены
✅ Repository обновлен
✅ UseCase обновлены
✅ gRPC handlers обновлены
✅ Employee Service обновлен

⏳ Требуется: пересборка и тестирование
