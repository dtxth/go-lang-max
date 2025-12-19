# End-to-End Tests

Комплексные end-to-end тесты для всех микросервисов системы Digital University MVP.

## Структура тестов

- `auth_service_test.go` - Тесты для Auth Service
- `employee_service_test.go` - Тесты для Employee Service  
- `chat_service_test.go` - Тесты для Chat Service
- `structure_service_test.go` - Тесты для Structure Service
- `maxbot_service_test.go` - Тесты для MaxBot Service
- `migration_service_test.go` - Тесты для Migration Service
- `integration_test.go` - Интеграционные тесты между сервисами
- `utils/` - Вспомогательные утилиты для тестов

## Запуск тестов

```bash
# Запуск всех e2e тестов
make test-e2e

# Запуск тестов конкретного сервиса
go test -v ./e2e-tests -run TestAuthService
go test -v ./e2e-tests -run TestEmployeeService
go test -v ./e2e-tests -run TestChatService
go test -v ./e2e-tests -run TestStructureService
go test -v ./e2e-tests -run TestMaxBotService
go test -v ./e2e-tests -run TestMigrationService

# Запуск интеграционных тестов
go test -v ./e2e-tests -run TestIntegration
```

## Требования

- Docker и docker-compose
- Все сервисы должны быть запущены
- Тестовые данные в базах данных