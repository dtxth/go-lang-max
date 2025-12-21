# Примеры использования E2E тестов

## Быстрый старт

```bash
# 1. Настройка окружения
make setup

# 2. Запуск всех сервисов
make up

# 3. Проверка здоровья сервисов
make health

# 4. Запуск всех E2E тестов
make test-e2e
```

## Примеры тестовых сценариев

### 1. Полный цикл пользователя

```bash
# Тест полного жизненного цикла пользователя
make test-e2e-integration
```

Этот тест выполняет:
- Регистрацию пользователя
- Авторизацию
- Создание университета
- Создание структуры
- Создание сотрудника
- Создание чата
- Проверку консистентности данных

### 2. Тестирование Auth Service

```bash
# Запуск всех тестов Auth Service
make test-e2e-auth
```

Покрывает:
- Регистрацию пользователей (email/phone)
- Авторизацию (email/phone)
- Обновление токенов
- Смену пароля
- Сброс пароля
- MAX авторизацию
- Получение информации о боте

### 3. Тестирование Structure Service

```bash
# Запуск тестов Structure Service
make test-e2e-structure
```

Покрывает:
- CRUD операции с университетами
- Создание структуры (филиалы, факультеты, кафедры, группы)
- Обновление названий
- Назначение операторов
- Импорт из Excel

### 4. Нагрузочное тестирование

```bash
# Запуск нагрузочных тестов
make test-load

# Запуск бенчмарков
make benchmark
```

## Примеры кастомных тестов

### Создание собственного теста

```go
// custom_test.go
package main

import (
    "e2e-tests/utils"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestCustomScenario(t *testing.T) {
    // Настройка
    configs := utils.DefaultServiceConfigs()
    authClient := utils.NewTestClient(configs["auth"])
    
    // Проверка доступности
    err := utils.WaitForService(configs["auth"].BaseURL, 10)
    require.NoError(t, err)
    
    // Ваш тест
    resp, err := authClient.GetClient().R().Get("/health")
    require.NoError(t, err)
    assert.Equal(t, 200, resp.StatusCode())
}
```

### Запуск кастомного теста

```bash
cd e2e-tests
go test -v -run TestCustomScenario
```

## Отладка и диагностика

### 1. Просмотр логов во время тестов

```bash
# В одном терминале запустите логи
make logs

# В другом терминале запустите тесты
make test-e2e-auth
```

### 2. Запуск одного конкретного теста

```bash
cd e2e-tests
go test -v -run TestAuthService/User_Registration
```

### 3. Запуск с увеличенным таймаутом

```bash
cd e2e-tests
go test -v -timeout 15m -run TestIntegration
```

### 4. Запуск с отладочной информацией

```bash
cd e2e-tests
go test -v -run TestAuthService -args -test.v
```

## Примеры для CI/CD

### GitHub Actions

```yaml
name: E2E Tests

on: [push, pull_request]

jobs:
  e2e-tests:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Setup environment
      run: make setup
    
    - name: Start services
      run: |
        make up
        sleep 60  # Ждем запуска всех сервисов
    
    - name: Check service health
      run: make health
    
    - name: Run E2E tests
      run: make test-e2e
      timeout-minutes: 20
    
    - name: Run load tests
      run: make test-load
      timeout-minutes: 10
    
    - name: Cleanup
      if: always()
      run: make clean
```

### GitLab CI

```yaml
stages:
  - test

e2e-tests:
  stage: test
  image: golang:1.21
  services:
    - docker:dind
  variables:
    DOCKER_DRIVER: overlay2
  before_script:
    - apt-get update -qq && apt-get install -y -qq docker-compose
    - make setup
  script:
    - make up
    - sleep 60
    - make health
    - make test-e2e
    - make test-load
  after_script:
    - make clean
  timeout: 30m
```

## Мониторинг и метрики

### 1. Сбор метрик во время тестов

```bash
# Запуск тестов с мониторингом
make monitor &
make test-e2e
```

### 2. Анализ производительности

```bash
# Бенчмарки с профилированием
cd e2e-tests
go test -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

# Анализ профилей
go tool pprof cpu.prof
go tool pprof mem.prof
```

### 3. Проверка метрик сервисов

```bash
# Метрики Auth Service
curl http://localhost:8080/metrics

# Метрики MaxBot Service  
curl http://localhost:8095/metrics
```

## Расширенные сценарии

### 1. Тест миграции данных

```bash
# Создание тестовых данных
make test-e2e-structure

# Тест миграции
make test-e2e-migration

# Проверка консистентности после миграции
make test-e2e-integration
```

### 2. Тест отказоустойчивости

```bash
# Запуск тестов
make test-e2e &

# Остановка одного сервиса
docker-compose stop auth-service

# Проверка обработки ошибок
# Перезапуск сервиса
docker-compose start auth-service
```

### 3. Тест масштабирования

```bash
# Масштабирование сервиса
docker-compose up -d --scale auth-service=3

# Запуск нагрузочных тестов
make test-load

# Проверка балансировки нагрузки
```

## Полезные команды

### Быстрая диагностика

```bash
# Статус всех сервисов
docker-compose ps

# Использование ресурсов
docker stats --no-stream

# Проверка сетевого взаимодействия
docker-compose exec auth-service ping chat-service
```

### Очистка и перезапуск

```bash
# Полная очистка
make clean

# Сброс баз данных
make db-reset

# Перезапуск с чистого листа
make clean && make setup && make up
```

### Работа с отдельными сервисами

```bash
# Запуск только баз данных для разработки
make dev-up

# Остановка dev окружения
make dev-down

# Перезапуск конкретного сервиса
docker-compose restart auth-service
```