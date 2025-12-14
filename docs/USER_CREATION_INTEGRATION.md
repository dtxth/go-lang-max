# Интеграция создания пользователей Employee ↔ Auth Service

## Проблема

Employee Service создавал сотрудников с ролями, но не создавал соответствующих пользователей в Auth Service. Поле `user_id` в таблице `employees` оставалось пустым, что делало невозможным:
- Вход в систему под учетной записью сотрудника
- Получение JWT токенов
- Использование ролевой модели ABAC

## Решение

### 1. Добавлен метод CreateUser в Auth Service

**Proto файл** (`auth-service/api/proto/auth.proto`):
```protobuf
rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);

message CreateUserRequest {
  string email = 1;
  string password = 2;
}

message CreateUserResponse {
  int64 user_id = 1;
  string error = 2;
}
```

**UseCase** (`auth-service/internal/usecase/auth_service.go`):
- Метод `CreateUser(email, password string) (int64, error)`
- Проверяет существование пользователя
- Хеширует пароль
- Создает запись в таблице `users`
- Возвращает `user_id`

**gRPC Handler** (`auth-service/internal/infrastructure/grpc/auth_handler.go`):
- Метод `CreateUser` для обработки gRPC запросов

### 2. Обновлен Employee Service

**Domain интерфейс** (`employee-service/internal/domain/auth_service.go`):
```go
type AuthService interface {
    CreateUser(ctx context.Context, email, password string) (int64, error)
    AssignRole(ctx context.Context, userID int64, role string, universityID, branchID, facultyID *int64) error
    RevokeUserRoles(ctx context.Context, userID int64) error
}
```

**AuthClient** (`employee-service/internal/infrastructure/auth/auth_client.go`):
- Добавлен метод `CreateUser` для вызова Auth Service через gRPC

**CreateEmployeeWithRoleUseCase** (`employee-service/internal/usecase/create_employee_with_role.go`):
- Обновлен для создания пользователя ПЕРЕД созданием сотрудника
- Генерирует email: `{phone}@employee.local`
- Генерирует временный пароль (TODO: улучшить безопасность)
- Сохраняет `user_id` в `employee.UserID`
- Назначает роль через `AssignRole`
- Откатывает изменения при ошибках

**HTTP Handler** (`employee-service/internal/infrastructure/http/handler.go`):
- Добавлено поле `Role` в `AddEmployeeRequest`
- Обновлен метод `AddEmployee` для вызова `CreateEmployeeWithRole` при наличии роли

**EmployeeService** (`employee-service/internal/usecase/employee_service.go`):
- Добавлена зависимость `authService`
- Добавлен метод `CreateEmployeeWithRole`

### 3. Правильный flow создания сотрудника с ролью

```
1. HTTP POST /employees с role="curator"
   ↓
2. Employee Service: AddEmployee handler
   ↓
3. Employee Service: CreateEmployeeWithRole usecase
   ↓
4. Auth Service: CreateUser(email, password) → user_id
   ↓
5. Employee Service: Создать employee с user_id
   ↓
6. Auth Service: AssignRole(user_id, role, university_id)
   ↓
7. Вернуть employee с заполненным user_id
```

## Текущий статус

### ✅ Реализовано:
- Метод CreateUser в Auth Service (proto, usecase, handler)
- Интеграция CreateUser в Employee Service
- Обновлен CreateEmployeeWithRoleUseCase
- Добавлено поле role в HTTP API

### ⚠️ Требует отладки:
- Роль не передается из HTTP handler в usecase
- `user_id` не сохраняется в базе данных
- Нужно добавить логирование для отладки

### 📋 TODO:
1. Отладить передачу роли из HTTP handler
2. Проверить, вызывается ли CreateUser
3. Добавить логирование в критических точках
4. Улучшить генерацию паролей (использовать crypto/rand)
5. Добавить механизм отправки пароля пользователю
6. Добавить тесты для CreateUser
7. Обновить Swagger документацию

## Тестирование

### Создание сотрудника с ролью:
```bash
curl -X POST 'http://localhost:8081/employees' \
  -H 'Content-Type: application/json' \
  -d '{
  "first_name": "Иван",
  "last_name": "Иванов",
  "phone": "+79001234567",
  "inn": "1234567890",
  "kpp": "123456789",
  "university_name": "МГУ",
  "role": "curator"
}'
```

### Проверка создания пользователя:
```bash
# В employee-db
docker exec -it employee-db psql -U employee_user -d employee_db \
  -c "SELECT id, phone, user_id, role FROM employees WHERE phone = '+79001234567';"

# В auth-db
docker exec -it auth-db psql -U postgres -d postgres \
  -c "SELECT id, email FROM users WHERE email LIKE '%79001234567%';"
```

## Архитектурные решения

### Почему email = phone@employee.local?
- Телефон уникален для каждого сотрудника
- Email требуется для Auth Service
- В будущем можно заменить на реальный email

### Почему временный пароль?
- Пользователь должен сменить пароль при первом входе
- В продакшене: генерировать случайный пароль и отправлять по SMS/email

### Почему CreateUser перед Create(employee)?
- Если создание пользователя не удалось, не создаем сотрудника
- Проще откатить изменения
- Гарантирует консистентность данных

## Связанные файлы

- `auth-service/api/proto/auth.proto`
- `auth-service/internal/usecase/auth_service.go`
- `auth-service/internal/infrastructure/grpc/auth_handler.go`
- `employee-service/internal/domain/auth_service.go`
- `employee-service/internal/infrastructure/auth/auth_client.go`
- `employee-service/internal/usecase/create_employee_with_role.go`
- `employee-service/internal/usecase/employee_service.go`
- `employee-service/internal/infrastructure/http/handler.go`
- `employee-service/cmd/employee/main.go`


## Методы создания пользователей в Auth Service

### CreateUser (gRPC) - для Employee Service
- Используется Employee Service при создании сотрудников с ролями
- Создает пользователя БЕЗ роли (роль назначается отдельно через AssignRole)
- Возвращает user_id для связи с employee
- **Основной метод для создания пользователей в системе**

### Register (HTTP) - для административных задач
- Используется для создания суперадминов и служебных аккаунтов
- Создает пользователя С ролью сразу
- Не создает запись в employee-service
- **Используется редко, только для административных целей**

### Когда использовать какой метод:

| Сценарий | Метод | Сервис |
|----------|-------|--------|
| Создание сотрудника вуза (curator/operator) | CreateUser + AssignRole | Employee Service |
| Создание первого суперадмина | Register | Auth Service (HTTP) |
| Создание служебного аккаунта | Register | Auth Service (HTTP) |
| Создание тестового пользователя | Register | Auth Service (HTTP) |

## Важные замечания

1. **Register vs CreateUser:**
   - `Register` (HTTP) - создает пользователя с ролью, используется для административных задач
   - `CreateUser` (gRPC) - создает пользователя без роли, используется Employee Service

2. **Не все пользователи - сотрудники:**
   - Суперадмины создаются через Register и НЕ имеют записи в employee-service
   - Сотрудники вузов создаются через Employee Service и имеют связь user_id

3. **Безопасность:**
   - Register должен быть защищен и доступен только суперадминам
   - CreateUser доступен только через gRPC (внутренняя коммуникация)
