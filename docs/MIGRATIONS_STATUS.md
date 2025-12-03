# Статус миграций базы данных

## Обзор

Все миграции успешно применены во всех сервисах. Используется инструмент `golang-migrate` для управления версиями схемы базы данных.

## Статус по сервисам

### ✅ Auth Service
- **База данных**: `auth-db` (postgres)
- **Версия миграций**: 3
- **Таблицы**:
  - `users` - пользователи системы
  - `refresh_tokens` - токены обновления
  - `roles` - роли пользователей
  - `user_roles` - связь пользователей и ролей
  - `schema_migrations` - версии миграций

### ✅ Employee Service
- **База данных**: `employee-db` (employee_db)
- **Версия миграций**: 3
- **Таблицы**:
  - `employees` - сотрудники
  - `universities` - университеты
  - `batch_update_jobs` - задачи пакетного обновления
  - `schema_migrations` - версии миграций

### ✅ Chat Service
- **База данных**: `chat-db` (chat_db)
- **Версия миграций**: 2
- **Таблицы**:
  - `chats` - чаты (с полями `external_chat_id`, `source`)
  - `administrators` - администраторы (с полями `max_id`, `add_user`, `add_admin`)
  - `universities` - университеты
  - `schema_migrations` - версии миграций

### ✅ Structure Service
- **База данных**: `structure-db` (structure_db)
- **Версия миграций**: 3
- **Таблицы**:
  - `universities` - университеты
  - `branches` - филиалы
  - `faculties` - факультеты
  - `groups` - группы (с полями `chat_url`, `chat_name`)
  - `department_managers` - менеджеры подразделений
  - `schema_migrations` - версии миграций

### ⚠️ Migration Service
- **База данных**: `migration-db` (migration_db)
- **Статус**: База данных не запущена
- **Примечание**: Миграции будут применены при первом запуске сервиса

## Формат файлов миграций

Все миграции переименованы в формат, совместимый с `golang-migrate`:
- Up миграции: `000001_name.up.sql`
- Down миграции: `000001_name.down.sql`

## Скрипты для управления миграциями

### check_migrations.sh
Проверяет статус миграций во всех сервисах:
```bash
./check_migrations.sh
```

### force_migrations.sh
Принудительно устанавливает версию миграций (используется для синхронизации):
```bash
./force_migrations.sh
```

### apply_migrations.sh
Применяет миграции через docker-compose (использует initdb):
```bash
./apply_migrations.sh
```

## Ключевые изменения

1. **Переименование файлов миграций** - все файлы приведены к формату `000001_name.up.sql`
2. **Добавлены down-миграции** - для всех up-миграций созданы соответствующие down-миграции
3. **Исправлена конфигурация structure-db** - изменено имя базы с `postgres` на `structure_db`
4. **Применены все миграции** - все таблицы и поля созданы корректно

## Проверка миграций

Для проверки статуса миграций в конкретном сервисе:

```bash
# Auth Service
docker exec auth-db psql -U postgres -d postgres -c "SELECT version, dirty FROM schema_migrations;"

# Employee Service
docker exec employee-db psql -U employee_user -d employee_db -c "SELECT version, dirty FROM schema_migrations;"

# Chat Service
docker exec chat-db psql -U chat_user -d chat_db -c "SELECT version, dirty FROM schema_migrations;"

# Structure Service
docker exec structure-db psql -U postgres -d structure_db -c "SELECT version, dirty FROM schema_migrations;"
```

## Следующие шаги

1. При необходимости добавления новых миграций используйте формат `00000X_description.up.sql` и `00000X_description.down.sql`
2. Для применения новых миграций используйте `golang-migrate` или docker-compose restart
3. Всегда создавайте down-миграции для возможности отката изменений

## Дата последней проверки

3 декабря 2025 г.
