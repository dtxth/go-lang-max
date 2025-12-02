# Testing and Deployment Guide

Полное руководство по тестированию и развертыванию микросервисов.

## 📋 Содержание

- [Быстрый старт](#быстрый-старт)
- [Запуск тестов](#запуск-тестов)
- [Развертывание](#развертывание)
- [Makefile команды](#makefile-команды)
- [CI/CD интеграция](#cicd-интеграция)

## 🚀 Быстрый старт

### Полное развертывание (рекомендуется)

```bash
# Запустить тесты, собрать образы и запустить сервисы
make deploy

# Или напрямую через скрипт
./deploy.sh
```

### Быстрое развертывание (без тестов)

```bash
# Пропустить тесты и сразу запустить
make deploy-fast

# Или
./deploy.sh --skip-tests
```

## 🧪 Запуск тестов

### Все тесты

```bash
# Запустить все тесты с race detector
make test

# Или напрямую
./run_tests.sh
```

### Быстрая проверка

```bash
# Быстрая проверка без race detector (быстрее)
make test-quick

# Или
./test_quick.sh
```

### С подробным выводом

```bash
# Подробный вывод всех тестов
make test-verbose

# Или
./run_tests.sh --verbose
```

### С покрытием кода

```bash
# Генерация отчета о покрытии
make test-coverage

# Или
./run_tests.sh --coverage
```

### Тесты отдельного сервиса

```bash
# Через Makefile
make test-auth
make test-chat
make test-employee
make test-structure
make test-maxbot
make test-migration

# Или напрямую
cd auth-service && go test -v -race ./...
```

## 📦 Развертывание

### Скрипт deploy.sh

Основной скрипт для полного цикла развертывания.

#### Опции

```bash
./deploy.sh [опции]

Опции:
  --skip-tests     Пропустить запуск тестов
  --no-cache       Пересобрать Docker образы без кеша
  --verbose, -v    Подробный вывод
  --coverage, -c   Генерация отчета о покрытии кода
  --help, -h       Показать справку
```

#### Примеры

```bash
# Полное развертывание с тестами
./deploy.sh

# Быстрое развертывание без тестов
./deploy.sh --skip-tests

# Полная пересборка
./deploy.sh --no-cache

# С подробным выводом и покрытием
./deploy.sh --verbose --coverage
```

### Этапы развертывания

Скрипт `deploy.sh` выполняет следующие шаги:

1. **Остановка контейнеров** - Останавливает существующие контейнеры
2. **Запуск тестов** - Проверяет все unit-тесты (если не --skip-tests)
3. **Сборка образов** - Собирает Docker образы
4. **Запуск сервисов** - Запускает все сервисы через docker-compose
5. **Проверка здоровья** - Проверяет доступность Swagger endpoints

## 🛠 Makefile команды

### Основные команды

```bash
make help           # Показать все доступные команды
make test           # Запустить все тесты
make deploy         # Полное развертывание
make up             # Запустить сервисы
make down           # Остановить сервисы
make logs           # Просмотр логов
make ps             # Статус контейнеров
```

### Тестирование

```bash
make test              # Все тесты с race detector
make test-quick        # Быстрая проверка
make test-verbose      # С подробным выводом
make test-coverage     # С покрытием кода

# Тесты отдельных сервисов
make test-auth
make test-chat
make test-employee
make test-structure
make test-maxbot
make test-migration
```

### Развертывание

```bash
make deploy            # Полное развертывание (тесты + сборка + запуск)
make deploy-fast       # Без тестов
make deploy-rebuild    # Полная пересборка
make deploy-verbose    # С подробным выводом
```

### Docker операции

```bash
make build             # Собрать образы
make build-no-cache    # Пересобрать без кеша
make up                # Запустить сервисы
make down              # Остановить сервисы
make restart           # Перезапустить сервисы
make images            # Показать размеры образов
```

### Логи

```bash
make logs              # Все логи
make logs-auth         # Логи auth-service
make logs-chat         # Логи chat-service
make logs-employee     # Логи employee-service
make logs-structure    # Логи structure-service
make logs-maxbot       # Логи maxbot-service
make logs-migration    # Логи migration-service
```

### Проверка здоровья

```bash
make ps                # Статус контейнеров
make swagger           # Проверка Swagger endpoints
make health            # Полная проверка здоровья
```

### Очистка

```bash
make clean             # Удалить контейнеры и образы
make clean-volumes     # Удалить все включая volumes
```

### Утилиты

```bash
make fmt               # Форматировать код
make lint              # Запустить линтер
make mod-tidy          # Обновить go.mod
```

## 🔄 CI/CD интеграция

### GitHub Actions

Пример `.github/workflows/test-and-deploy.yml`:

```yaml
name: Test and Deploy

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'
      
      - name: Run tests
        run: ./run_tests.sh --coverage
      
      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./*/coverage.out

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Build Docker images
        run: docker-compose build
      
      - name: Push to registry
        run: |
          # Ваш код для push в registry
```

### GitLab CI

Пример `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - build
  - deploy

test:
  stage: test
  image: golang:1.24
  script:
    - chmod +x run_tests.sh
    - ./run_tests.sh --coverage
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

build:
  stage: build
  script:
    - docker-compose build
  only:
    - main
    - develop

deploy:
  stage: deploy
  script:
    - ./deploy.sh --skip-tests
  only:
    - main
  when: manual
```

## 📊 Структура тестов

### Unit тесты

Расположены в каждом сервисе:

```
auth-service/
├── internal/
│   ├── infrastructure/
│   │   ├── jwt/jwt_manager_test.go
│   │   ├── middleware/request_id_test.go
│   │   └── errors/errors_test.go
│   └── usecase/
│       └── validate_permission_test.go

chat-service/
├── internal/
│   └── usecase/
│       ├── add_administrator_with_permission_check_test.go
│       ├── remove_administrator_with_validation_test.go
│       ├── list_chats_with_role_filter_test.go
│       ├── list_chats_with_role_filter_pagination_test.go
│       └── search_chats_test.go

employee-service/
├── internal/
│   ├── infrastructure/
│   │   └── grpc/retry_test.go
│   └── usecase/
│       ├── employee_service_test.go
│       ├── create_employee_with_role_test.go
│       └── batch_update_max_id_test.go

structure-service/
├── internal/
│   └── usecase/
│       └── assign_operator_to_department_test.go

maxbot-service/
├── internal/
│   └── usecase/
│       ├── batch_get_users_by_phone_test.go
│       └── normalize_phone_test.go
```

### Integration тесты

```
integration-tests/
├── grpc_integration_test.go
├── migration_integration_test.go
├── structure_integration_test.go
├── employee_integration_test.go
├── chat_integration_test.go
└── helpers.go
```

## 🐛 Отладка

### Просмотр логов при ошибках

```bash
# Логи конкретного сервиса
docker-compose logs chat-service

# Последние 100 строк
docker-compose logs --tail=100 chat-service

# Следить за логами в реальном времени
docker-compose logs -f chat-service
```

### Проверка статуса

```bash
# Статус всех контейнеров
docker-compose ps

# Детальная информация
docker inspect chat-service
```

### Перезапуск проблемного сервиса

```bash
# Перезапустить один сервис
docker-compose restart chat-service

# Пересобрать и перезапустить
docker-compose up -d --build chat-service
```

## 📝 Лучшие практики

### Перед коммитом

```bash
# 1. Запустить тесты
make test

# 2. Форматировать код
make fmt

# 3. Проверить линтером (если установлен)
make lint

# 4. Обновить go.mod
make mod-tidy
```

### Перед деплоем

```bash
# 1. Полное тестирование с покрытием
make test-coverage

# 2. Полная пересборка
make deploy-rebuild

# 3. Проверка здоровья
make health
```

### Регулярное обслуживание

```bash
# Очистка неиспользуемых образов
docker image prune -a

# Очистка volumes
docker volume prune

# Полная очистка проекта
make clean-volumes
```

## 🔗 Связанные документы

- [Docker Cross-Service Dependencies](./DOCKER_CROSS_SERVICE_DEPENDENCIES.md)
- [Docker Build Quick Start](./DOCKER_BUILD_QUICK_START.md)
- [Integration Tests Guide](./integration-tests/INTEGRATION_TEST_GUIDE.md)
- [README](./README.md)

## 💡 Советы

1. **Используйте Makefile** - Все команды доступны через `make`, это удобнее
2. **Запускайте тесты локально** - Не полагайтесь только на CI/CD
3. **Проверяйте покрытие** - Стремитесь к >80% покрытия кода
4. **Следите за логами** - При проблемах сразу смотрите логи
5. **Используйте --verbose** - При отладке включайте подробный вывод
