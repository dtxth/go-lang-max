# Подход к миграциям базы данных

## Обзор

В проекте используется стандартный механизм PostgreSQL для применения миграций через `/docker-entrypoint-initdb.d`.

## Как это работает

### 1. Файлы миграций

Каждый сервис имеет папку `migrations/` с SQL файлами:

```
auth-service/migrations/001_init.sql
chat-service/migrations/001_init.sql
employee-service/migrations/001_init.sql
structure-service/migrations/001_init.sql
migration-service/migrations/001_init.sql
```

### 2. Docker Compose конфигурация

В `docker-compose.yml` каждая база данных монтирует папку с миграциями:

```yaml
migration-db:
  image: postgres:15-alpine
  environment:
    POSTGRES_DB: migration_db
    POSTGRES_USER: postgres
    POSTGRES_PASSWORD: postgres
  volumes:
    - migration_db_data:/var/lib/postgresql/data
    - ./migration-service/migrations:/docker-entrypoint-initdb.d  # ← Миграции
```

### 3. Автоматическое применение

PostgreSQL автоматически выполняет все `.sql` файлы из `/docker-entrypoint-initdb.d` при **первом** создании базы данных.

**Важно:** Миграции применяются только при первом запуске, когда база данных еще не существует!

## Проблема с существующими базами

Если база данных уже существует (volume уже создан), миграции не применятся автоматически.

### Решение 1: Пересоздать базу данных

```bash
# Остановить все сервисы
docker-compose down

# Удалить volume базы данных
docker volume rm go-lang-max_migration_db_data

# Запустить заново
docker-compose up -d
```

### Решение 2: Fallback в коде (текущий подход)

В `migration-service/internal/app/server.go` добавлена функция `ensureTablesExist()`:

```go
func ensureTablesExist(db *sql.DB) error {
	// Проверяем существуют ли таблицы
	var exists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'migration_jobs'
		);
	`).Scan(&exists)
	
	if exists {
		log.Println("Database tables already exist")
		return nil
	}

	// Создаем таблицы если их нет
	log.Println("Creating database tables...")
	// ... SQL для создания таблиц ...
}
```

Эта функция:
1. Проверяет существуют ли таблицы
2. Если да - ничего не делает
3. Если нет - создает их (fallback)

## Преимущества текущего подхода

✅ **Миграции в отдельных файлах** - легко версионировать и управлять  
✅ **Автоматическое применение** - PostgreSQL делает это сам при первом запуске  
✅ **Fallback в коде** - если миграции не применились, код создаст таблицы  
✅ **Идемпотентность** - `CREATE TABLE IF NOT EXISTS` безопасно запускать многократно  
✅ **Нет дублирования** - SQL код в одном месте (`migrations/001_init.sql`)  

## Логи

### При первом запуске (база не существует)

PostgreSQL применяет миграции автоматически:
```
migration-db | /usr/local/bin/docker-entrypoint.sh: running /docker-entrypoint-initdb.d/001_init.sql
```

Сервис видит что таблицы уже есть:
```
migration-service | Connected to database
migration-service | Database tables already exist
migration-service | Starting migration service on port 8084
```

### При запуске с существующей базой (миграции не применялись)

Сервис создает таблицы через fallback:
```
migration-service | Connected to database
migration-service | Creating database tables (migrations not applied yet)...
migration-service | Database tables created successfully
migration-service | Starting migration service on port 8084
```

## Рекомендации

### Для разработки

Если нужно применить новые миграции:

```bash
# Остановить сервисы
docker-compose down

# Удалить volume конкретной базы
docker volume rm go-lang-max_migration_db_data

# Запустить заново
docker-compose up -d migration-db migration-service
```

### Для продакшена

Используйте инструменты миграций:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)
- [sql-migrate](https://github.com/rubenv/sql-migrate)

Пример с golang-migrate:

```bash
# Установка
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Применение миграций
migrate -path ./migration-service/migrations \
        -database "postgres://postgres:postgres@localhost:5433/migration_db?sslmode=disable" \
        up
```

## Структура миграций

### Именование файлов

```
001_init.sql          - Основная миграция
001_init_down.sql     - Откат миграции (опционально)
002_add_column.sql    - Следующая миграция
002_add_column_down.sql
```

### Пример миграции

**001_init.sql:**
```sql
-- Migration jobs table
CREATE TABLE migration_jobs (
  id SERIAL PRIMARY KEY,
  source_type TEXT NOT NULL,
  status TEXT NOT NULL,
  ...
);

CREATE INDEX idx_migration_jobs_status ON migration_jobs(status);
```

**001_init_down.sql:**
```sql
DROP TABLE IF EXISTS migration_errors;
DROP TABLE IF EXISTS migration_jobs;
```

## Заключение

Текущий подход использует:
1. **Стандартный механизм PostgreSQL** для автоматического применения миграций
2. **Fallback в коде** для случаев когда база уже существует
3. **Идемпотентные операции** (`CREATE TABLE IF NOT EXISTS`)

Это обеспечивает надежность и простоту развертывания без дублирования SQL кода.
