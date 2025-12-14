# Быстрое руководство по миграциям

## 🚀 Самый простой способ (рекомендуется)

Пересоздать все базы данных с автоматическим применением миграций:

```bash
# 1. Остановить и удалить все (включая volumes)
docker-compose down -v

# 2. Запустить заново
docker-compose up -d

# 3. Подождать пока сервисы запустятся
sleep 10

# 4. Проверить статус
./check_migrations.sh
```

**Что происходит:**
- PostgreSQL автоматически выполняет все `.sql` файлы из `/docker-entrypoint-initdb.d` при первом создании БД
- Миграции из папок `*/migrations/` монтируются в эту директорию через docker-compose.yml

## 📋 Доступные команды

### Проверить статус миграций

```bash
./check_migrations.sh
```

### Применить миграции вручную (если база уже существует)

```bash
./apply_migrations.sh
```

### Проверить конкретную базу данных

```bash
# Auth Service
docker-compose exec auth-db psql -U postgres -d auth_db -c "\dt"

# Employee Service  
docker-compose exec employee-db psql -U employee_user -d employee_db -c "\dt"

# Chat Service
docker-compose exec chat-db psql -U chat_user -d chat_db -c "\dt"

# Structure Service
docker-compose exec structure-db psql -U postgres -d structure_db -c "\dt"

# Migration Service
docker-compose exec migration-db psql -U postgres -d migration_db -c "\dt"
```

## 🔧 Если миграции не применились

### Вариант 1: Пересоздать конкретную базу

```bash
# Остановить сервис
docker-compose stop auth-service auth-db

# Удалить volume
docker volume rm go-lang-max_auth_pgdata

# Запустить заново
docker-compose up -d auth-db auth-service
```

### Вариант 2: Применить миграции вручную

```bash
./apply_migrations.sh
```

### Вариант 3: Применить для конкретного сервиса

```bash
# Auth Service
docker-compose exec -T auth-db psql -U postgres -d auth_db < auth-service/migrations/001_init.sql

# Employee Service
docker-compose exec -T employee-db psql -U employee_user -d employee_db < employee-service/migrations/001_init.sql

# Chat Service
docker-compose exec -T chat-db psql -U chat_user -d chat_db < chat-service/migrations/001_init.sql

# Structure Service
docker-compose exec -T structure-db psql -U postgres -d structure_db < structure-service/migrations/001_init.sql

# Migration Service
docker-compose exec -T migration-db psql -U postgres -d migration_db < migration-service/migrations/001_init.sql
```

## ✅ Проверка успешности

После применения миграций вы должны увидеть:

```bash
./check_migrations.sh
```

**Ожидаемый результат:**

```
=== Auth Service Database ===
Tables:
 public | users          | table | postgres
 public | refresh_tokens | table | postgres
 public | user_roles     | table | postgres

=== Employee Service Database ===
Tables:
 public | employees          | table | employee_user
 public | universities       | table | employee_user
 public | batch_update_jobs  | table | employee_user

=== Chat Service Database ===
Tables:
 public | chats          | table | chat_user
 public | administrators | table | chat_user

=== Structure Service Database ===
Tables:
 public | universities        | table | postgres
 public | branches            | table | postgres
 public | faculties           | table | postgres
 public | groups              | table | postgres
 public | department_managers | table | postgres

=== Migration Service Database ===
Tables:
 public | migration_jobs   | table | postgres
 public | migration_errors | table | postgres
```

## 📁 Где находятся миграции

```
auth-service/migrations/001_init.sql
employee-service/migrations/001_init.sql
chat-service/migrations/001_init.sql
structure-service/migrations/001_init.sql
migration-service/migrations/001_init.sql
```

## 🔄 Откат миграций

Если есть `*_down.sql` файлы:

```bash
# Откатить конкретную миграцию
docker-compose exec -T auth-db psql -U postgres -d auth_db < auth-service/migrations/001_init_down.sql
```

Или просто пересоздать базу:

```bash
docker-compose down -v
docker-compose up -d
```

## 📚 Дополнительная документация

- [MIGRATION_APPROACH.md](./MIGRATION_APPROACH.md) - Подробное описание подхода
- [MIGRATIONS.md](./MIGRATIONS.md) - Общая документация
- [verify_migrations.sh](./verify_migrations.sh) - Скрипт проверки
