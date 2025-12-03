# Реорганизация структуры проекта

## Изменения

Для улучшения организации проекта все скрипты из корневой директории были перемещены в соответствующие папки.

### Новая структура

```
.
├── bin/                    # Утилиты сборки и развертывания
│   ├── deploy.sh
│   ├── generate_proto.sh
│   ├── update_swagger.sh
│   ├── apply_migrations.sh
│   ├── check_migrations.sh
│   ├── verify_migrations.sh
│   └── ...
│
├── tests/                  # Тестовые скрипты
│   ├── run_tests.sh
│   ├── test_quick.sh
│   ├── test_api_handlers.sh
│   ├── test_migrations.sh
│   ├── test_excel_*.sh
│   └── ...
│
└── integration-tests/      # Интеграционные тесты (без изменений)
    ├── run_tests.sh
    └── run_e2e_tests.sh
```

### Перемещенные файлы

#### bin/ (утилиты и сборка)
- `deploy.sh` - развертывание
- `generate_proto.sh` - генерация proto
- `update_swagger.sh` - обновление Swagger
- `apply_migrations.sh` - применение миграций
- `apply_excel_migration.sh` - миграции Excel
- `check_migrations.sh` - проверка миграций
- `force_migrations.sh` - принудительные миграции
- `rename_migrations.sh` - переименование миграций
- `verify_migrations.sh` - верификация миграций
- `validate_sql_syntax.sh` - валидация SQL

#### tests/ (тестовые скрипты)
- `run_tests.sh` - все тесты
- `test_quick.sh` - быстрая проверка
- `test_api_handlers.sh` - API тесты
- `test_migrations.sh` - тесты миграций
- `test_excel_import.sh` - тесты Excel импорта
- `test_excel_upload.sh` - тесты загрузки Excel
- `test_excel_e2e.sh` - E2E тесты Excel
- `test_migration_api.sh` - тесты API миграций
- `diagnose_excel.sh` - диагностика Excel
- `create_test_excel.py` - создание тестовых файлов
- `test_excel_read.go` - утилита чтения Excel

## Обновленные пути

### Makefile
Все команды в Makefile обновлены для использования новых путей:
```bash
make test          # ./tests/run_tests.sh
make test-quick    # ./tests/test_quick.sh
make deploy        # ./bin/deploy.sh
```

### README.md
Все ссылки на скрипты обновлены:
```bash
./bin/deploy.sh
./tests/run_tests.sh
./tests/test_quick.sh
```

### Скрипты
Внутренние ссылки в скриптах также обновлены.

## Использование

### Через Makefile (рекомендуется)
```bash
make help          # Показать все команды
make test          # Запустить тесты
make deploy        # Развертывание
```

### Напрямую
```bash
# Утилиты
./bin/deploy.sh
./bin/generate_proto.sh
./bin/verify_migrations.sh

# Тесты
./tests/run_tests.sh
./tests/test_quick.sh
./tests/test_api_handlers.sh
```

## Преимущества

1. **Чистый корень проекта** - только основные конфигурационные файлы
2. **Логическая группировка** - скрипты сгруппированы по назначению
3. **Легче навигация** - понятно где искать нужный скрипт
4. **Масштабируемость** - проще добавлять новые скрипты

## Обратная совместимость

Все команды через `make` работают без изменений. Если вы использовали прямые вызовы скриптов, обновите пути:

```bash
# Старый способ
./deploy.sh

# Новый способ
./bin/deploy.sh

# Или через make (рекомендуется)
make deploy
```
