# Руководство по E2E тестированию

## Подготовка к тестированию

### 1. Настройка окружения

```bash
# Создание .env файла
make setup

# Запуск всех сервисов
make up

# Проверка здоровья сервисов
make health
```

### 2. Проверка доступности сервисов

```bash
# Проверка всех сервисов
curl http://localhost:8080/health  # Auth Service
curl http://localhost:8081/health  # Employee Service  
curl http://localhost:8082/health  # Chat Service
curl http://localhost:8083/health  # Structure Service
curl http://localhost:8084/health  # Migration Service
curl http://localhost:8095/health  # MaxBot Service
```

## Запуск тестов

### Все E2E тесты
```bash
make test-e2e
```

### Тесты отдельных сервисов
```bash
make test-e2e-auth        # Auth Service
make test-e2e-structure   # Structure Service
make test-e2e-employee    # Employee Service
make test-e2e-chat        # Chat Service
make test-e2e-maxbot      # MaxBot Service
make test-e2e-migration   # Migration Service
```

### Интеграционные тесты
```bash
make test-e2e-integration
```

### Тесты производительности
```bash
make test-load     # Нагрузочные тесты
make benchmark     # Бенчмарки
```

## Структура тестов

### Auth Service Tests
- ✅ Health check
- ✅ Metrics endpoint
- ✅ User registration
- ✅ Login (email/phone)
- ✅ Token refresh
- ✅ Password management
- ✅ MAX authentication
- ✅ Bot info endpoint

### Structure Service Tests
- ✅ University CRUD operations
- ✅ Structure management
- ✅ Department managers
- ✅ Name updates (universities, branches, faculties, groups)
- ✅ Excel import
- ✅ Error handling

### Employee Service Tests
- ✅ Employee creation
- ✅ Batch operations
- ✅ MaxID updates
- ✅ Status checks
- ✅ Error handling

### Chat Service Tests
- ✅ Chat CRUD operations
- ✅ Message management
- ✅ Participant management
- ✅ Authorization checks
- ✅ Error handling

### MaxBot Service Tests
- ✅ Profile management
- ✅ Bot info
- ✅ Init data validation
- ✅ Webhook handling
- ✅ Cache operations
- ✅ Mock mode testing

### Migration Service Tests
- ✅ File upload
- ✅ Migration management
- ✅ Google Sheets integration
- ✅ Data validation
- ✅ Status tracking
- ✅ Error handling

### Integration Tests
- ✅ Cross-service workflows
- ✅ Data consistency
- ✅ Service communication
- ✅ Error propagation
- ✅ Performance monitoring

## Конфигурация тестов

### Переменные окружения для тестов

```bash
# Таймауты
export TEST_TIMEOUT=30s
export INTEGRATION_TIMEOUT=60s

# Базовые URL сервисов (по умолчанию localhost)
export AUTH_SERVICE_URL=http://localhost:8080
export EMPLOYEE_SERVICE_URL=http://localhost:8081
export CHAT_SERVICE_URL=http://localhost:8082
export STRUCTURE_SERVICE_URL=http://localhost:8083
export MIGRATION_SERVICE_URL=http://localhost:8084
export MAXBOT_SERVICE_URL=http://localhost:8095

# Тестовые данные
export TEST_USER_EMAIL=test@example.com
export TEST_USER_PHONE=+79001234567
export TEST_USER_PASSWORD=TestPassword123!
```

### Настройка для CI/CD

```yaml
# Пример для GitHub Actions
- name: Run E2E Tests
  run: |
    make up
    sleep 30  # Ждем запуска сервисов
    make test-e2e
  env:
    TEST_TIMEOUT: 60s
    INTEGRATION_TIMEOUT: 120s
```

## Отладка тестов

### Просмотр логов сервисов
```bash
make logs
```

### Запуск отдельного теста с подробным выводом
```bash
cd e2e-tests
go test -v -run TestAuthService_UserRegistration
```

### Запуск тестов с отладочной информацией
```bash
cd e2e-tests
go test -v -run TestIntegration -timeout 10m -args -debug
```

## Решение проблем

### Сервис недоступен
```bash
# Проверить статус контейнеров
docker-compose ps

# Перезапустить сервисы
make restart

# Проверить логи
make logs
```

### Тесты падают по таймауту
```bash
# Увеличить таймауты в go.mod или переменных окружения
export TEST_TIMEOUT=60s

# Или запустить с увеличенным таймаутом
go test -timeout 15m
```

### Проблемы с базой данных
```bash
# Сброс всех баз данных
make db-reset

# Перезапуск с чистыми данными
make clean && make up
```

### Проблемы с авторизацией
```bash
# Проверить, что auth-service запущен
curl http://localhost:8080/health

# Проверить токены в тестах
# Убедиться, что тестовые пользователи создаются корректно
```

## Метрики и мониторинг

### Просмотр метрик сервисов
```bash
curl http://localhost:8080/metrics  # Auth Service metrics
```

### Мониторинг ресурсов
```bash
make monitor
```

### Статистика тестов
```bash
# Запуск с подсчетом покрытия
go test -cover ./...

# Бенчмарки с детальной статистикой
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof
```

## Лучшие практики

1. **Изоляция тестов**: Каждый тест должен быть независимым
2. **Очистка данных**: Используйте уникальные тестовые данные
3. **Обработка ошибок**: Проверяйте как успешные, так и ошибочные сценарии
4. **Таймауты**: Устанавливайте разумные таймауты для тестов
5. **Логирование**: Используйте подробное логирование для отладки
6. **Параллелизм**: Тесты должны корректно работать параллельно

## Расширение тестов

### Добавление нового теста
1. Создайте функцию `TestNewFeature` в соответствующем файле
2. Используйте существующие утилиты из `utils/`
3. Следуйте паттерну AAA (Arrange, Act, Assert)
4. Добавьте тест в соответствующий Makefile target

### Добавление нового сервиса
1. Создайте новый файл `new_service_test.go`
2. Добавьте конфигурацию в `utils/client.go`
3. Создайте тестовые данные в `utils/test_data.go`
4. Добавьте Makefile target для нового сервиса