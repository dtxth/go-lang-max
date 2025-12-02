# API Tests Coverage - Цифровой Вуз

Документ описывает покрытие тестами всех API методов системы "Цифровой Вуз".

## Обзор

Все микросервисы системы теперь имеют unit-тесты для HTTP хендлеров, которые проверяют:
- Валидацию входных данных
- Обработку некорректных запросов
- Правильные HTTP статус-коды
- Обработку ошибок

## Запуск тестов

### Запуск всех тестов API хендлеров

```bash
./test_api_handlers.sh
```

### Запуск тестов для конкретного сервиса

```bash
# Auth Service
cd auth-service && go test -v ./internal/infrastructure/http/

# Employee Service
cd employee-service && go test -v ./internal/infrastructure/http/

# Chat Service
cd chat-service && go test -v ./internal/infrastructure/http/

# Structure Service
cd structure-service && go test -v ./internal/infrastructure/http/

# Migration Service
cd migration-service && go test -v ./internal/infrastructure/http/
```

## Покрытие по сервисам

### 1. Auth Service

**Файл тестов:** `auth-service/internal/infrastructure/http/handler_test.go`

**Покрытые endpoints:**

| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/register` | POST | ✅ Invalid JSON, Missing email, Missing password |
| `/login` | POST | ✅ Invalid JSON, Missing email, Missing password |
| `/refresh` | POST | ✅ Invalid JSON, Missing token |
| `/logout` | POST | ✅ Invalid JSON, Missing token |
| `/health` | GET | ✅ Success response |

**Количество тестов:** 11

**Что тестируется:**
- Валидация обязательных полей (email, password, refresh_token)
- Обработка некорректного JSON
- Health check endpoint

---

### 2. Employee Service

**Файл тестов:** `employee-service/internal/infrastructure/http/handler_test.go`

**Покрытые endpoints:**

| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/employees` | GET | ✅ Missing auth, Invalid auth format |
| `/employees` | POST | ✅ Missing phone, Missing name, Invalid JSON |
| `/employees/{id}` | GET | ✅ Invalid ID |
| `/employees/{id}` | PUT | ✅ Invalid ID, Invalid JSON |
| `/employees/{id}` | DELETE | ✅ Invalid ID |
| `/employees/batch-update-maxid` | POST | ✅ Service not available |
| `/employees/batch-status` | GET | ✅ Service not available |
| `/employees/batch-status/{id}` | GET | ✅ Service not available, Invalid ID |

**Количество тестов:** 13

**Что тестируется:**
- Валидация ID параметров
- Валидация обязательных полей (phone, first_name, last_name)
- Проверка авторизации (Bearer token)
- Обработка некорректного JSON
- Проверка доступности batch сервиса

---

### 3. Chat Service

**Файл тестов:** `chat-service/internal/infrastructure/http/handler_test.go`

**Покрытые endpoints:**

| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/chats` | GET | ✅ Unauthorized |
| `/chats/all` | GET | ✅ Unauthorized |
| `/chats/{id}` | GET | ✅ Invalid ID |
| `/chats/{chat_id}/administrators` | POST | ✅ Missing phone, Invalid chat ID, Invalid JSON, Invalid path |
| `/administrators/{admin_id}` | DELETE | ✅ Invalid ID |

**Количество тестов:** 8

**Что тестируется:**
- Проверка авторизации для защищенных endpoints
- Валидация ID параметров
- Валидация обязательных полей (phone)
- Обработка некорректного JSON
- Валидация URL путей

---

### 4. Structure Service

**Файл тестов:** `structure-service/internal/infrastructure/http/handler_test.go`

**Покрытые endpoints:**

| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/universities` | POST | ✅ Invalid JSON |
| `/universities/{id}` | GET | ✅ Invalid ID |
| `/universities/{university_id}/structure` | GET | ✅ Invalid ID |
| `/departments/managers` | POST | ✅ Invalid JSON |
| `/departments/managers/{id}` | DELETE | ✅ Invalid ID |
| `/import/excel` | POST | ✅ Invalid method, Missing file |

**Количество тестов:** 7

**Что тестируется:**
- Валидация ID параметров
- Обработка некорректного JSON
- Проверка HTTP методов
- Валидация multipart/form-data для загрузки файлов

---

### 5. Migration Service

**Файл тестов:** `migration-service/internal/infrastructure/http/handler_test.go`

**Покрытые endpoints:**

| Endpoint | Метод | Тесты |
|----------|-------|-------|
| `/migration/database` | POST | ✅ Invalid method, Invalid JSON |
| `/migration/google-sheets` | POST | ✅ Invalid method, Invalid JSON, Missing spreadsheet_id |
| `/migration/excel` | POST | ✅ Invalid method, Missing file |
| `/migration/jobs/{id}` | GET | ✅ Invalid method, Invalid ID |
| `/migration/jobs` | GET | ✅ Invalid method |

**Количество тестов:** 10

**Что тестируется:**
- Проверка HTTP методов
- Обработка некорректного JSON
- Валидация ID параметров
- Валидация обязательных полей (spreadsheet_id)
- Валидация multipart/form-data для загрузки файлов

---

## Общая статистика

| Сервис | Endpoints | Тестов | Статус |
|--------|-----------|--------|--------|
| Auth Service | 5 | 11 | ✅ |
| Employee Service | 8 | 13 | ✅ |
| Chat Service | 5 | 8 | ✅ |
| Structure Service | 6 | 7 | ✅ |
| Migration Service | 5 | 10 | ✅ |
| **ИТОГО** | **29** | **49** | ✅ |

## Типы тестов

### 1. Валидация входных данных
- Проверка обязательных полей
- Проверка формата данных
- Проверка типов данных

### 2. Обработка ошибок
- Некорректный JSON
- Невалидные ID
- Отсутствующие параметры

### 3. Авторизация и аутентификация
- Проверка наличия токена
- Проверка формата токена
- Проверка прав доступа

### 4. HTTP методы
- Проверка правильности HTTP метода
- Обработка неподдерживаемых методов

### 5. Загрузка файлов
- Проверка наличия файла
- Проверка формата файла
- Проверка размера файла

## Интеграционные тесты

Помимо unit-тестов для HTTP хендлеров, система также имеет интеграционные тесты:

**Файл:** `integration-tests/`

- `chat_integration_test.go` - Тесты фильтрации чатов по ролям, пагинации, поиска
- `employee_integration_test.go` - Тесты создания и управления сотрудниками
- `migration_integration_test.go` - Тесты миграции данных
- `structure_integration_test.go` - Тесты импорта структуры
- `grpc_integration_test.go` - Тесты gRPC взаимодействия

**Запуск интеграционных тестов:**

```bash
cd integration-tests
go test -v
```

## Use Case тесты

Каждый сервис также имеет тесты для бизнес-логики (use cases):

### Auth Service
- `internal/usecase/validate_permission_test.go` - Тесты проверки прав доступа

### Employee Service
- `internal/usecase/create_employee_with_role_test.go` - Тесты создания сотрудника с ролью
- `internal/usecase/batch_update_max_id_test.go` - Тесты пакетного обновления MAX_id

### Chat Service
- `internal/usecase/list_chats_with_role_filter_test.go` - Тесты фильтрации чатов
- `internal/usecase/list_chats_with_role_filter_pagination_test.go` - Тесты пагинации
- `internal/usecase/search_chats_test.go` - Тесты поиска чатов
- `internal/usecase/add_administrator_with_permission_check_test.go` - Тесты добавления администратора
- `internal/usecase/remove_administrator_with_validation_test.go` - Тесты удаления администратора

### Structure Service
- `internal/usecase/assign_operator_to_department_test.go` - Тесты назначения оператора

### Maxbot Service
- `internal/usecase/batch_get_users_by_phone_test.go` - Тесты пакетного получения пользователей
- `internal/usecase/normalize_phone_test.go` - Тесты нормализации телефонов

## Рекомендации по расширению тестов

### 1. Добавить тесты с моками
Для более полного покрытия рекомендуется добавить тесты с моками сервисов:
- Успешные сценарии создания/обновления/удаления
- Обработка ошибок от зависимых сервисов
- Проверка корректности возвращаемых данных

### 2. Добавить тесты производительности
- Нагрузочное тестирование endpoints
- Тестирование пагинации с большими объемами данных
- Тестирование batch операций

### 3. Добавить тесты безопасности
- SQL injection
- XSS атаки
- CSRF защита
- Rate limiting

### 4. Добавить E2E тесты
- Полные сценарии использования системы
- Тестирование взаимодействия между сервисами
- Тестирование UI (если есть)

## Continuous Integration

Рекомендуется настроить CI/CD pipeline для автоматического запуска тестов:

```yaml
# Пример для GitHub Actions
name: API Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - name: Run API Handler Tests
        run: ./test_api_handlers.sh
      - name: Run Integration Tests
        run: cd integration-tests && go test -v
```

## Заключение

Все основные API методы системы "Цифровой Вуз" покрыты базовыми unit-тестами, которые проверяют валидацию входных данных и обработку ошибок. Это обеспечивает базовый уровень качества кода и помогает предотвратить регрессии при внесении изменений.

Для запуска всех тестов используйте:

```bash
# API Handler тесты
./test_api_handlers.sh

# Интеграционные тесты
cd integration-tests && go test -v

# Все тесты проекта
./run_tests.sh
```
