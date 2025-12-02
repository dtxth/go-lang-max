# Migration Service - Исправление проблемы с базой данных

## Проблема

При запросе к `/migration/jobs` возвращалась ошибка 500:
```json
{
  "error": "Failed to list migration jobs"
}
```

## Причина

Таблицы `migration_jobs` и `migration_errors` не создавались автоматически при старте сервиса. Миграции существовали в папке `migrations/`, но не применялись.

## Решение

Добавлена функция `ensureTablesExist()` в `migration-service/internal/app/server.go`, которая проверяет наличие таблиц и создает их при необходимости.

**Важно:** Миграции должны применяться автоматически через `/docker-entrypoint-initdb.d` в PostgreSQL контейнере (см. `docker-compose.yml`). Функция `ensureTablesExist()` - это fallback для случаев, когда база данных уже существует и миграции не были применены автоматически.

### Изменения

**Файл:** `migration-service/internal/app/server.go`

Добавлена функция `ensureTablesExist()` которая проверяет наличие таблиц и создает их при необходимости:

```go
// ensureTablesExist creates tables if they don't exist
// This is a fallback for when migrations in docker-entrypoint-initdb.d didn't run
func ensureTablesExist(db *sql.DB) error {
	// Check if migration_jobs table exists
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'migration_jobs'
		);
	`).Scan(&exists)
	
	if err != nil {
		return fmt.Errorf("failed to check if tables exist: %w", err)
	}

	if exists {
		log.Println("Database tables already exist")
		return nil
	}

	log.Println("Creating database tables (migrations not applied yet)...")
	// ... создание таблиц ...
}
```

**Файл:** `migration-service/migrations/001_init.sql`

Основные миграции находятся здесь и применяются автоматически через PostgreSQL `/docker-entrypoint-initdb.d` при первом создании базы данных.

**Файл:** `docker-compose.yml`

```yaml
migration-db:
  volumes:
    - migration_db_data:/var/lib/postgresql/data
    - ./migration-service/migrations:/docker-entrypoint-initdb.d  # Автоматическое применение миграций
```

## Проверка

### 1. Проверка таблиц в базе данных

```bash
docker-compose exec -T migration-db psql -U postgres -d migration_db -c "\dt"
```

Результат:
```
              List of relations
 Schema |       Name       | Type  |  Owner   
--------+------------------+-------+----------
 public | migration_errors | table | postgres
 public | migration_jobs   | table | postgres
(2 rows)
```

### 2. Проверка API

```bash
curl -s http://localhost:8084/migration/jobs -H "accept: application/json"
```

Результат:
```json
null
```

Пустой список (null) - это правильный ответ, так как миграций еще не было запущено.

### 3. Проверка логов

```bash
docker-compose logs migration-service --tail=10
```

Результат:
```
migration-service  | 2025/12/02 16:34:55 Connected to database
migration-service  | 2025/12/02 16:34:55 Running database migrations...
migration-service  | 2025/12/02 16:34:56 Database migrations completed successfully
migration-service  | 2025/12/02 16:34:56 Starting migration service on port 8084
```

## Пересборка и перезапуск

Для применения исправления:

```bash
# Остановить сервис
docker-compose stop migration-service

# Пересобрать образ
docker-compose build migration-service

# Пересоздать и запустить контейнер
docker-compose rm -sf migration-service
docker-compose create migration-service
docker-compose start migration-service

# Проверить логи
docker-compose logs migration-service --tail=20
```

## Статус

✅ **Исправлено** - Migration Service теперь автоматически создает необходимые таблицы при старте и API работает корректно.

## Связанные файлы

- `migration-service/internal/app/server.go` - Добавлена функция runMigrations()
- `migration-service/migrations/001_init.sql` - SQL миграция (уже существовала)
- `migration-service/internal/infrastructure/repository/migration_job_postgres.go` - Репозиторий (без изменений)

## Тесты

Все тесты API хендлеров проходят успешно:

```bash
cd migration-service && go test -v ./internal/infrastructure/http/
```

Результат: **10/10 тестов PASSED**
