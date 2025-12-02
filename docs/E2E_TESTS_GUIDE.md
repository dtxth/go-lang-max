# End-to-End Tests Guide

## Обзор

End-to-End (E2E) тесты проверяют полный flow работы системы от начала до конца, включая взаимодействие всех микросервисов.

## Структура E2E тестов

### Файлы

```
integration-tests/
├── e2e_full_flow_test.go        - Полные пользовательские сценарии
├── e2e_error_handling_test.go   - Обработка ошибок и edge cases
├── run_e2e_tests.sh             - Скрипт для запуска всех E2E тестов
└── helpers.go                   - Вспомогательные функции
```

## Покрытие тестами

### 1. Complete User Journey (e2e_full_flow_test.go)

Тестирует полный путь пользователя:

1. ✅ Регистрация superadmin
2. ✅ Создание университета
3. ✅ Регистрация curator
4. ✅ Создание сотрудника (operator)
5. ✅ Создание чата
6. ✅ Добавление администратора чата
7. ✅ Поиск чатов (superadmin видит все)
8. ✅ Поиск чатов (curator видит только свой университет)
9. ✅ Получение информации о сотруднике
10. ✅ Получение структуры университета

**Проверяет:**
- Взаимодействие всех сервисов
- Передачу данных между сервисами
- Корректность ролевой модели

### 2. Role-Based Access Control

Тестирует ролевую модель доступа:

- ✅ Superadmin видит все чаты
- ✅ Curator видит только чаты своего университета
- ✅ Operator видит только чаты своего подразделения
- ✅ Неавторизованный доступ блокируется

**Проверяет:**
- ABAC (Attribute-Based Access Control)
- Фильтрацию по ролям
- Защиту endpoints

### 3. Chat Administrator Management

Тестирует управление администраторами чатов:

- ✅ Добавление первого администратора
- ✅ Добавление второго администратора
- ✅ Предотвращение дубликатов (409 Conflict)
- ✅ Удаление администратора
- ✅ Защита последнего администратора (нельзя удалить)

**Проверяет:**
- Бизнес-логику управления администраторами
- Валидацию правил (минимум 1 администратор)
- Обработку конфликтов

### 4. Pagination and Search

Тестирует пагинацию и поиск:

- ✅ Создание множества записей
- ✅ Пагинация (limit/offset)
- ✅ Поиск по названию
- ✅ Корректность результатов

**Проверяет:**
- Работу пагинации
- Полнотекстовый поиск
- Производительность с большим количеством данных

### 5. Error Handling (e2e_error_handling_test.go)

Тестирует обработку ошибок:

- ✅ Некорректный JSON (400 Bad Request)
- ✅ Отсутствующие обязательные поля (400 Bad Request)
- ✅ Невалидные ID (400 Bad Request)
- ✅ Несуществующие ресурсы (404 Not Found)
- ✅ Дубликаты (409 Conflict)
- ✅ Отсутствие авторизации (401 Unauthorized)
- ✅ Невалидный токен (401 Unauthorized)

**Проверяет:**
- Корректные HTTP статус-коды
- Валидацию входных данных
- Обработку edge cases

### 6. Concurrent Operations

Тестирует конкурентные операции:

- ✅ Одновременное создание множества чатов
- ✅ Отсутствие race conditions
- ✅ Целостность данных при конкурентном доступе

**Проверяет:**
- Thread-safety
- Транзакционность
- Производительность под нагрузкой

### 7. Data Consistency

Тестирует консистентность данных:

- ✅ Создание связанных данных (университет → сотрудник → чат → администратор)
- ✅ Проверка связей между сущностями
- ✅ Целостность данных после операций

**Проверяет:**
- Referential integrity
- Консистентность между сервисами
- Корректность связей

## Запуск тестов

### Предварительные требования

1. Все сервисы должны быть запущены:

```bash
docker-compose up -d
```

2. Проверьте что сервисы работают:

```bash
curl http://localhost:8080/health  # Auth Service
curl http://localhost:8081/employees/all  # Employee Service
curl http://localhost:8082/chats/all  # Chat Service
curl http://localhost:8083/universities  # Structure Service
```

### Запуск всех E2E тестов

```bash
cd integration-tests
./run_e2e_tests.sh
```

### Запуск конкретного теста

```bash
cd integration-tests

# Complete User Journey
go test -v -run TestE2E_CompleteUserJourney -timeout 5m

# Role-Based Access Control
go test -v -run TestE2E_RoleBasedAccessControl -timeout 5m

# Chat Administrator Management
go test -v -run TestE2E_ChatAdministratorManagement -timeout 5m

# Pagination and Search
go test -v -run TestE2E_PaginationAndSearch -timeout 5m

# Error Handling
go test -v -run TestE2E_ErrorHandling -timeout 5m

# Concurrent Operations
go test -v -run TestE2E_ConcurrentOperations -timeout 5m

# Data Consistency
go test -v -run TestE2E_DataConsistency -timeout 5m
```

### Запуск с подробным выводом

```bash
go test -v -run TestE2E_ -timeout 10m
```

## Результаты

### Успешный запуск

```
=========================================
Running End-to-End Tests
=========================================

✓ All services are running

=== Test 1: Complete User Journey ===
✓ Complete User Journey test passed

=== Test 2: Role-Based Access Control ===
✓ Role-Based Access Control test passed

=== Test 3: Chat Administrator Management ===
✓ Chat Administrator Management test passed

=== Test 4: Pagination and Search ===
✓ Pagination and Search test passed

=== Test 5: Error Handling ===
✓ Error Handling test passed

=== Test 6: Concurrent Operations ===
✓ Concurrent Operations test passed

=== Test 7: Data Consistency ===
✓ Data Consistency test passed

=========================================
All E2E tests passed!
=========================================
```

## Архитектура тестов

### Структура теста

```go
func TestE2E_FeatureName(t *testing.T) {
    t.Log("=== Starting E2E Feature Test ===")
    
    // 1. Wait for services
    WaitForService(t, AuthServiceURL, 30)
    WaitForService(t, EmployeeServiceURL, 30)
    
    // 2. Setup test data
    token := registerUser(t, "test@test.com", "Pass123!", "superadmin")
    
    // 3. Execute test scenario
    // ... test steps ...
    
    // 4. Verify results
    if result != expected {
        t.Errorf("Expected %v, got %v", expected, result)
    }
    
    t.Log("=== E2E Feature Test PASSED ===")
}
```

### Helper функции

```go
// Регистрация пользователя
registerUser(t, email, password, role) string

// Создание университета
createUniversity(t, token, name, inn, kpp) int64

// Создание сотрудника
createEmployee(t, token, phone, firstName, lastName, universityName) int64

// Создание чата
createChat(t, token, name, url, universityID) int64

// Добавление администратора
addChatAdministrator(t, token, chatID, phone) int64

// Поиск чатов
searchChats(t, token, query) []map[string]interface{}

// Список чатов с пагинацией
listChats(t, token, limit, offset) []map[string]interface{}

// Получение сотрудника
getEmployee(t, token, employeeID) map[string]interface{}

// Получение структуры университета
getUniversityStructure(t, token, universityID) map[string]interface{}

// Выполнение HTTP запроса
makeRequest(t, method, url, token, body) *http.Response
```

## Отладка

### Просмотр логов сервисов

```bash
# Все сервисы
docker-compose logs -f

# Конкретный сервис
docker-compose logs -f auth-service
docker-compose logs -f employee-service
docker-compose logs -f chat-service
docker-compose logs -f structure-service
```

### Проверка состояния базы данных

```bash
# Auth DB
docker-compose exec auth-db psql -U postgres -d auth_db -c "SELECT * FROM users;"

# Employee DB
docker-compose exec employee-db psql -U employee_user -d employee_db -c "SELECT * FROM employees;"

# Chat DB
docker-compose exec chat-db psql -U chat_user -d chat_db -c "SELECT * FROM chats;"

# Structure DB
docker-compose exec structure-db psql -U postgres -d structure_db -c "SELECT * FROM universities;"
```

### Очистка тестовых данных

```bash
# Пересоздать базы данных
docker-compose down -v
docker-compose up -d
```

## Best Practices

### 1. Изоляция тестов

Каждый тест должен быть независимым и не зависеть от других тестов.

```go
// ✅ Good - создаем свои данные
func TestE2E_Feature(t *testing.T) {
    token := registerUser(t, "unique@test.com", "Pass123!", "superadmin")
    // ... test logic ...
}

// ❌ Bad - зависим от данных других тестов
func TestE2E_Feature(t *testing.T) {
    // Предполагаем что пользователь уже существует
}
```

### 2. Таймауты

Всегда устанавливайте таймауты для HTTP запросов:

```go
client := &http.Client{Timeout: 10 * time.Second}
```

### 3. Обработка ошибок

Используйте `t.Logf` для предупреждений и `t.Fatalf` для критических ошибок:

```go
// Предупреждение - тест продолжается
if resp.StatusCode != http.StatusOK {
    t.Logf("Warning: Unexpected status %d", resp.StatusCode)
}

// Критическая ошибка - тест останавливается
if token == "" {
    t.Fatal("Failed to get authentication token")
}
```

### 4. Очистка ресурсов

Используйте `defer` для очистки:

```go
resp, err := client.Do(req)
if err != nil {
    t.Fatalf("Request failed: %v", err)
}
defer resp.Body.Close()
```

## Интеграция с CI/CD

### GitHub Actions

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Start services
        run: docker-compose up -d
      
      - name: Wait for services
        run: sleep 30
      
      - name: Run E2E tests
        run: cd integration-tests && ./run_e2e_tests.sh
      
      - name: Stop services
        run: docker-compose down -v
```

## Метрики

### Покрытие

- **7 E2E тестов** покрывают основные сценарии
- **4 сервиса** тестируются в интеграции
- **10+ API endpoints** проверяются в каждом тесте

### Производительность

- Среднее время выполнения одного теста: **30-60 секунд**
- Общее время выполнения всех E2E тестов: **5-10 минут**

## Troubleshooting

### Проблема: Тесты падают с timeout

**Решение:**
1. Проверьте что все сервисы запущены
2. Увеличьте timeout: `-timeout 10m`
3. Проверьте логи сервисов

### Проблема: 401 Unauthorized

**Решение:**
1. Проверьте что auth-service работает
2. Проверьте что токен генерируется корректно
3. Проверьте формат Authorization header

### Проблема: Данные не создаются

**Решение:**
1. Проверьте логи соответствующего сервиса
2. Проверьте состояние базы данных
3. Проверьте что миграции применились

## Дополнительные ресурсы

- [Integration Tests Guide](./integration-tests/INTEGRATION_TEST_GUIDE.md)
- [API Tests Coverage](./API_TESTS_COVERAGE.md)
- [Testing and Deployment](./TESTING_AND_DEPLOYMENT.md)
- [README](./README.md)
