# Migration Service - Source Column Fix

## Проблема

При запуске Excel миграции все записи падали с ошибками:

```
failed to create chat: chat service error: pq: column "source" of relation "chats" does not exist
failed to create chat: chat service error: pq: column "max_id" does not exist
```

**Статус миграции:**
```json
{
  "id": 1,
  "source_type": "excel",
  "status": "running",
  "total": 0,
  "processed": 0,
  "failed": 18300,
  "started_at": "2025-12-04T15:38:35Z"
}
```

## Причина

В таблицах отсутствовали необходимые колонки:
1. В таблице `chats` отсутствовала колонка `source`
2. В таблице `administrators` отсутствовала колонка `max_id`

Хотя обе колонки были определены в миграции `000001_init.up.sql`:

```sql
CREATE TABLE IF NOT EXISTS chats (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  -- ...
  source TEXT NOT NULL CHECK (source IN ('admin_panel', 'bot_registrar', 'academic_group')),
  -- ...
);
```

**Проблема:** Миграция использовала `CREATE TABLE IF NOT EXISTS`, поэтому если таблица уже существовала (без колонки `source`), новые колонки не добавлялись.

## Решение

### 1. Применение миграций

Сначала применили все миграции через скрипт:

```bash
bash bin/apply_migrations.sh
```

Результат:
```
=== Chat Service ===
✓ Applied: 000001_init.up.sql
✓ Applied: 000002_add_excel_fields.up.sql
✓ Chat Service migrations completed
```

### 2. Добавление недостающих колонок вручную

Так как таблицы уже существовали, добавили колонки вручную:

```sql
-- Добавить колонку source в chats
ALTER TABLE chats 
ADD COLUMN IF NOT EXISTS source TEXT 
NOT NULL DEFAULT 'admin_panel' 
CHECK (source IN ('admin_panel', 'bot_registrar', 'academic_group'));

-- Создать индекс для source
CREATE INDEX IF NOT EXISTS idx_chats_source ON chats(source);

-- Добавить колонку max_id в administrators
ALTER TABLE administrators 
ADD COLUMN IF NOT EXISTS max_id TEXT;

-- Создать индекс для max_id
CREATE INDEX IF NOT EXISTS idx_administrators_max_id ON administrators(max_id);
```

### 3. Создание новой миграции

Создали миграцию `000003_add_source_column.up.sql` для будущих развертываний:

```sql
-- Добавление колонки source для отслеживания источника чата
ALTER TABLE chats 
ADD COLUMN IF NOT EXISTS source TEXT 
NOT NULL DEFAULT 'admin_panel' 
CHECK (source IN ('admin_panel', 'bot_registrar', 'academic_group'));

-- Создание индекса для source
CREATE INDEX IF NOT EXISTS idx_chats_source ON chats(source);

-- Добавление колонки max_id для администраторов (если отсутствует)
ALTER TABLE administrators 
ADD COLUMN IF NOT EXISTS max_id TEXT;

-- Создание индекса для max_id
CREATE INDEX IF NOT EXISTS idx_administrators_max_id ON administrators(max_id);

-- Комментарии для документации
COMMENT ON COLUMN chats.source IS 'Источник создания чата: admin_panel, bot_registrar, или academic_group';
COMMENT ON COLUMN administrators.max_id IS 'ID пользователя в системе MAX';
```

## Проверка

### До исправления

```bash
# Проверить структуру таблицы
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\d chats"

# Результат: колонка source отсутствует ❌
```

### После исправления

```bash
# Проверить структуру таблицы chats
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\d chats" | grep source

# Результат:
# source             | text                     |           | not null | 'admin_panel'::text
# "idx_chats_source" btree (source)
# "chats_source_check" CHECK (source = ANY (ARRAY['admin_panel'::text, 'bot_registrar'::text, 'academic_group'::text]))
# ✅ Колонка source добавлена

# Проверить структуру таблицы administrators
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\d administrators" | grep max_id

# Результат:
# max_id     | text                     |           |          | 
# "idx_administrators_max_id" btree (max_id)
# ✅ Колонка max_id добавлена
```

### Проверка миграции

```bash
# Запустить новую миграцию
curl -X POST http://localhost:8084/migration/excel \
  -F "file=@chats.xlsx"

# Проверить статус
curl http://localhost:8084/migration/jobs/2

# Результат:
{
  "id": 2,
  "source_type": "excel",
  "status": "completed",  # ✅ Успешно
  "total": 18300,
  "processed": 18300,
  "failed": 0
}
```

## Значения колонки source

Колонка `source` используется для отслеживания источника создания чата:

| Значение | Описание | Используется в |
|----------|----------|----------------|
| `admin_panel` | Чат создан через админ-панель | Database Migration |
| `bot_registrar` | Чат создан через бот-регистратор | Google Sheets Migration |
| `academic_group` | Чат создан для учебной группы | Excel Migration |

## Изменения в коде

### Migration Service

В `migration-service/internal/infrastructure/chat/chat_client.go` передается `source`:

```go
reqBody := CreateChatRequest{
    Name:              chat.Name,
    URL:               chat.URL,
    ExternalChatID:    externalChatID,
    Source:            chat.Source,  // ← Передается source
    UniversityID:      &chat.UniversityID,
    BranchID:          chat.BranchID,
    FacultyID:         chat.FacultyID,
    ParticipantsCount: 0,
}
```

В `migration-service/internal/usecase/migrate_from_excel.go`:

```go
chatData := &domain.ChatData{
    Name:           row.ChatName,
    URL:            row.ChatURL,
    ExternalChatID: row.ChatID,
    UniversityID:   0,
    Source:         "academic_group",  // ← Устанавливается source
}
```

## Best Practices для миграций

### 1. Используйте ALTER TABLE для существующих таблиц

```sql
-- ❌ НЕПРАВИЛЬНО - не добавит колонку если таблица существует
CREATE TABLE IF NOT EXISTS chats (
  id SERIAL PRIMARY KEY,
  new_column TEXT
);

-- ✅ ПРАВИЛЬНО - добавит колонку если её нет
ALTER TABLE chats ADD COLUMN IF NOT EXISTS new_column TEXT;
```

### 2. Создавайте отдельные миграции для изменений схемы

```
migrations/
  000001_init.up.sql              # Создание таблиц
  000002_add_excel_fields.up.sql  # Добавление полей для Excel
  000003_add_source_column.up.sql # Добавление колонки source
```

### 3. Всегда добавляйте down миграции

```sql
-- 000003_add_source_column.down.sql
DROP INDEX IF EXISTS idx_chats_source;
ALTER TABLE chats DROP COLUMN IF EXISTS source;
```

### 4. Используйте значения по умолчанию

```sql
-- ✅ ПРАВИЛЬНО - можно добавить к существующим данным
ALTER TABLE chats 
ADD COLUMN IF NOT EXISTS source TEXT 
NOT NULL DEFAULT 'admin_panel';

-- ❌ НЕПРАВИЛЬНО - упадет если есть данные
ALTER TABLE chats 
ADD COLUMN IF NOT EXISTS source TEXT NOT NULL;
```

### 5. Добавляйте комментарии

```sql
COMMENT ON COLUMN chats.source IS 'Источник создания чата: admin_panel, bot_registrar, или academic_group';
```

## Проверка миграций

### Скрипт для проверки

```bash
#!/bin/bash

# Проверить что все колонки существуют
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "
SELECT column_name, data_type, is_nullable, column_default
FROM information_schema.columns
WHERE table_name = 'chats'
ORDER BY ordinal_position;
"
```

### Ожидаемый результат

```
 column_name        | data_type                   | is_nullable | column_default
--------------------+-----------------------------+-------------+-------------------
 id                 | integer                     | NO          | nextval('chats_id_seq'::regclass)
 name               | text                        | NO          | 
 url                | text                        | NO          | 
 max_chat_id        | text                        | YES         | 
 participants_count | integer                     | YES         | 0
 university_id      | integer                     | YES         | 
 department         | text                        | YES         | 
 source             | text                        | NO          | 'admin_panel'::text  ✅
 created_at         | timestamp with time zone    | YES         | now()
 updated_at         | timestamp with time zone    | YES         | now()
 external_chat_id   | text                        | YES         | 
```

## Статус

✅ **Исправлено** - Колонка `source` добавлена в таблицу `chats`, создана миграция для будущих развертываний.

## Связанные документы

- [Migration Context Fix](./MIGRATION_CONTEXT_FIX.md)
- [Migration Service Implementation](../migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)
- [Migration Approach](./MIGRATION_APPROACH.md)
- [Migrations Status](./MIGRATIONS_STATUS.md)
