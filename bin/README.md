# Утилиты и скрипты сборки

Эта папка содержит утилиты для сборки, развертывания и управления проектом.

## Скрипты развертывания

- **deploy.sh** - Полное развертывание (тесты → сборка → запуск)
- **generate_proto.sh** - Генерация Go кода из proto файлов
- **update_swagger.sh** - Обновление Swagger документации для всех сервисов

## Скрипты миграций

- **apply_migrations.sh** - Применение миграций БД
- **apply_excel_migration.sh** - Применение миграций для Excel импорта
- **check_migrations.sh** - Проверка статуса миграций
- **force_migrations.sh** - Принудительная установка версии миграций
- **rename_migrations.sh** - Переименование файлов миграций
- **verify_migrations.sh** - Проверка целостности миграций
- **validate_sql_syntax.sh** - Валидация SQL синтаксиса

## Использование

Все скрипты запускаются из корня проекта:

```bash
# Развертывание
./bin/deploy.sh

# Генерация proto
./bin/generate_proto.sh

# Проверка миграций
./bin/verify_migrations.sh
```

Или через Makefile:

```bash
make deploy
make help
```
