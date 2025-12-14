# Ролевая модель и управление доступом (ABAC)

Подробное руководство по системе управления доступом на основе атрибутов (Attribute-Based Access Control).

## Содержание

1. [Обзор](#обзор)
2. [Роли](#роли)
3. [Контекстные атрибуты](#контекстные-атрибуты)
4. [Иерархия прав](#иерархия-прав)
5. [Примеры использования](#примеры-использования)
6. [Реализация](#реализация)
7. [Тестирование](#тестирование)

## Обзор

Система "Цифровой Вуз" использует ABAC (Attribute-Based Access Control) для гибкого управления доступом к ресурсам. В отличие от простой ролевой модели (RBAC), ABAC учитывает не только роль пользователя, но и контекстные атрибуты (университет, филиал, факультет).

### Ключевые принципы

1. **Иерархия ролей**: Высшие роли имеют все права низших ролей
2. **Контекстная привязка**: Роли привязаны к конкретным организационным единицам
3. **Автоматическая фильтрация**: Данные автоматически фильтруются по правам доступа
4. **Валидация на уровне сервисов**: Каждый сервис проверяет права через Auth Service

## Роли

### Superadmin (Суперадминистратор)

**Описание:** Представитель VK с полными правами доступа ко всей системе.

**Права доступа:**
- Просмотр и управление всеми университетами
- Создание и управление кураторами
- Доступ ко всем чатам и сотрудникам
- Управление системными настройками

**Контекст:** Не привязан к конкретному университету

**Пример использования:**
```bash
# Superadmin видит все чаты всех университетов
GET /chats
→ Возвращает все чаты без фильтрации
```

### Curator (Куратор)

**Описание:** Ответственный представитель от конкретного университета.

**Права доступа:**
- Просмотр и управление данными своего университета
- Создание и управление операторами своего университета
- Управление сотрудниками своего университета
- Доступ к чатам своего университета

**Контекст:** Привязан к одному университету (university_id)

**Пример использования:**
```bash
# Curator видит только чаты своего университета
GET /chats
→ WHERE university_id = <curator_university_id>
```

### Operator (Оператор)

**Описание:** Представитель конкретного подразделения (филиала или факультета).

**Права доступа:**
- Просмотр данных своего подразделения
- Управление чатами своего подразделения
- Просмотр сотрудников своего подразделения

**Контекст:** Привязан к филиалу (branch_id) или факультету (faculty_id)

**Пример использования:**
```bash
# Operator видит только чаты своего филиала/факультета
GET /chats
→ WHERE branch_id = <operator_branch_id> OR faculty_id = <operator_faculty_id>
```

## Контекстные атрибуты

### Структура UserRole

```go
type UserRole struct {
    ID           int       `json:"id"`
    UserID       int       `json:"user_id"`
    RoleID       int       `json:"role_id"`
    Role         string    `json:"role"`         // "superadmin", "curator", "operator"
    UniversityID *int      `json:"university_id"` // NULL для superadmin
    BranchID     *int      `json:"branch_id"`     // NULL для curator и superadmin
    FacultyID    *int      `json:"faculty_id"`    // NULL для curator и superadmin
    AssignedBy   int       `json:"assigned_by"`
    AssignedAt   time.Time `json:"assigned_at"`
}
```

### Правила привязки

| Роль | university_id | branch_id | faculty_id |
|------|---------------|-----------|------------|
| Superadmin | NULL | NULL | NULL |
| Curator | Обязательно | NULL | NULL |
| Operator | Обязательно | Один из двух | Один из двух |

## Иерархия прав

```
Superadmin (все университеты)
    ↓ имеет все права
Curator (один университет)
    ↓ имеет все права в своем университете
Operator (один филиал или факультет)
    ↓ имеет права только в своем подразделении
```

### Кумулятивные права

Высшие роли автоматически получают все права низших ролей:

- **Superadmin** может делать все, что может Curator и Operator
- **Curator** может делать все, что может Operator в своем университете

## Примеры использования

### Назначение роли Curator

```bash
# Superadmin назначает куратора для МГУ
curl -X POST http://localhost:8080/roles/assign \
  -H "Authorization: Bearer <SUPERADMIN_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 123,
    "role": "curator",
    "university_id": 1
  }'
```

### Назначение роли Operator

```bash
# Curator назначает оператора для филиала
curl -X POST http://localhost:8080/roles/assign \
  -H "Authorization: Bearer <CURATOR_TOKEN>" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 456,
    "role": "operator",
    "university_id": 1,
    "branch_id": 10
  }'
```

### Фильтрация чатов по ролям

```go
// Пример реализации в Chat Service
func (s *ChatService) ListChats(ctx context.Context, userID int) ([]*Chat, error) {
    // Получение роли пользователя через Auth Service gRPC
    userRole, err := s.authClient.GetUserRole(ctx, userID)
    if err != nil {
        return nil, err
    }

    // Применение фильтрации на основе роли
    switch userRole.Role {
    case "superadmin":
        // Возвращаем все чаты
        return s.repo.FindAll(ctx)
    
    case "curator":
        // Фильтруем по университету
        return s.repo.FindByUniversity(ctx, *userRole.UniversityID)
    
    case "operator":
        // Фильтруем по филиалу или факультету
        if userRole.BranchID != nil {
            return s.repo.FindByBranch(ctx, *userRole.BranchID)
        }
        return s.repo.FindByFaculty(ctx, *userRole.FacultyID)
    
    default:
        return nil, ErrUnauthorized
    }
}
```

### Проверка прав доступа

```go
// Пример проверки прав перед добавлением администратора чата
func (s *ChatService) AddAdministrator(ctx context.Context, chatID int, phone string, userID int) error {
    // Получение чата
    chat, err := s.repo.FindByID(ctx, chatID)
    if err != nil {
        return err
    }

    // Получение роли пользователя
    userRole, err := s.authClient.GetUserRole(ctx, userID)
    if err != nil {
        return err
    }

    // Проверка прав доступа
    hasPermission := false
    switch userRole.Role {
    case "superadmin":
        hasPermission = true
    case "curator":
        hasPermission = chat.UniversityID == *userRole.UniversityID
    case "operator":
        if userRole.BranchID != nil {
            hasPermission = chat.BranchID != nil && *chat.BranchID == *userRole.BranchID
        } else if userRole.FacultyID != nil {
            hasPermission = chat.FacultyID != nil && *chat.FacultyID == *userRole.FacultyID
        }
    }

    if !hasPermission {
        return ErrForbidden
    }

    // Добавление администратора
    return s.repo.AddAdministrator(ctx, chatID, phone)
}
```

## Реализация

### Auth Service

**Файлы:**
- `internal/domain/role.go` - Доменная модель роли
- `internal/domain/user_role.go` - Доменная модель связи пользователь-роль
- `internal/domain/permission.go` - Value object для прав доступа
- `internal/usecase/validate_permission.go` - Use case валидации прав
- `internal/infrastructure/repository/role_postgres.go` - Репозиторий ролей

**gRPC API:**
```protobuf
service AuthService {
  rpc ValidateToken(ValidateTokenRequest) returns (ValidateTokenResponse);
  rpc GetUserPermissions(GetUserPermissionsRequest) returns (GetUserPermissionsResponse);
  rpc AssignRole(AssignRoleRequest) returns (AssignRoleResponse);
}

message ValidateTokenResponse {
  int32 user_id = 1;
  string role = 2;
  int32 university_id = 3;
  int32 branch_id = 4;
  int32 faculty_id = 5;
}
```

### Employee Service

**Интеграция с Auth Service:**
- Создание сотрудника → Назначение роли через Auth Service gRPC
- Обновление сотрудника → Синхронизация роли
- Удаление сотрудника → Отзыв всех ролей

**Файлы:**
- `internal/usecase/create_employee_with_role.go`
- `internal/usecase/update_employee_with_role_sync.go`
- `internal/usecase/delete_employee_with_permission_revocation.go`

### Chat Service

**Middleware для проверки прав:**
```go
// internal/infrastructure/http/middleware.go
func AuthMiddleware(authClient AuthServiceClient) gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        
        // Валидация токена через Auth Service
        userRole, err := authClient.ValidateToken(c.Request.Context(), token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        // Сохранение роли в контексте
        c.Set("user_role", userRole)
        c.Next()
    }
}
```

**Файлы:**
- `internal/usecase/list_chats_with_role_filter.go`
- `internal/usecase/add_administrator_with_permission_check.go`
- `internal/domain/chat_filter.go`

## Тестирование

### Unit тесты

```go
// auth-service/internal/usecase/validate_permission_test.go
func TestValidatePermission_Curator_CanAccessOwnUniversity(t *testing.T) {
    // Arrange
    userRole := &domain.UserRole{
        Role:         "curator",
        UniversityID: intPtr(1),
    }
    resource := &domain.Resource{
        Type:         "chat",
        UniversityID: intPtr(1),
    }
    
    // Act
    hasPermission := ValidatePermission(userRole, resource, "read")
    
    // Assert
    assert.True(t, hasPermission)
}

func TestValidatePermission_Curator_CannotAccessOtherUniversity(t *testing.T) {
    // Arrange
    userRole := &domain.UserRole{
        Role:         "curator",
        UniversityID: intPtr(1),
    }
    resource := &domain.Resource{
        Type:         "chat",
        UniversityID: intPtr(2),
    }
    
    // Act
    hasPermission := ValidatePermission(userRole, resource, "read")
    
    // Assert
    assert.False(t, hasPermission)
}
```

### Интеграционные тесты

```go
// integration-tests/chat_integration_test.go
func TestChatFiltering_ByRole(t *testing.T) {
    // Создание тестовых данных
    superadminToken := createUser(t, "superadmin", nil, nil, nil)
    curatorToken := createUser(t, "curator", intPtr(1), nil, nil)
    operatorToken := createUser(t, "operator", intPtr(1), intPtr(10), nil)
    
    // Создание чатов в разных университетах
    chat1 := createChat(t, 1, nil, nil) // Университет 1
    chat2 := createChat(t, 2, nil, nil) // Университет 2
    
    // Superadmin видит все чаты
    chats := listChats(t, superadminToken)
    assert.Len(t, chats, 2)
    
    // Curator видит только чаты своего университета
    chats = listChats(t, curatorToken)
    assert.Len(t, chats, 1)
    assert.Equal(t, chat1.ID, chats[0].ID)
    
    // Operator видит только чаты своего филиала
    chats = listChats(t, operatorToken)
    // ... проверка фильтрации по филиалу
}
```

## Документация

Дополнительная информация:
- [Тесты валидации прав](./auth-service/internal/usecase/validate_permission_test.go)
- [Реализация фильтрации чатов](./chat-service/ROLE_BASED_FILTERING_IMPLEMENTATION.md)
- [Требования ABAC](./.kiro/specs/digital-university-mvp-completion/requirements.md#requirement-1)
- [Дизайн ABAC](./.kiro/specs/digital-university-mvp-completion/design.md#role-based-access-control-flow)
