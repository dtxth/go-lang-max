# Тестовые скрипты

Эта папка содержит скрипты для запуска тестов и диагностики.

## Основные тестовые скрипты

- **run_tests.sh** - Запуск всех тестов с race detector
- **test_quick.sh** - Быстрая проверка тестов (без race detector)
- **test_api_handlers.sh** - Тесты API хендлеров всех сервисов
- **test_migrations.sh** - Тесты миграций БД

## Excel импорт тесты

- **test_excel_import.sh** - Unit-тесты Excel импорта
- **test_excel_upload.sh** - Тесты загрузки Excel файлов
- **test_excel_e2e.sh** - E2E тесты Excel импорта
- **test_migration_api.sh** - Тесты API миграций
- **diagnose_excel.sh** - Диагностика проблем с Excel

## Утилиты

- **create_test_excel.py** - Создание тестовых Excel файлов

## Использование

Все скрипты запускаются из корня проекта:

```bash
# Быстрая проверка
./tests/test_quick.sh

# Все тесты
./tests/run_tests.sh

# Тесты конкретного сервиса
./tests/test_quick.sh auth-service

# API тесты
./tests/test_api_handlers.sh
```

Или через Makefile:

```bash
make test          # Все тесты с race detector
make test-quick    # Быстрая проверка
make test-verbose  # С подробным выводом
```

## Интеграционные тесты

Интеграционные и E2E тесты находятся в папке `integration-tests/`:

```bash
cd integration-tests
./run_tests.sh
./run_e2e_tests.sh
```
