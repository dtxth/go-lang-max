# Swagger Documentation Guide - Migration Service

## Доступ к Swagger UI

После запуска сервиса, откройте в браузере:

```
http://localhost:8084/swagger/index.html
```

## Доступные Endpoints

### 1. POST /migration/database
Запуск миграции из существующей базы данных (admin_panel source)

**Request Body:**
```json
{
  "source_identifier": "postgres://user:pass@host:5432/old_db"
}
```

**Response:** `202 Accepted`
```json
{
  "message": "Database migration started"
}
```

### 2. POST /migration/google-sheets
Запуск миграции из Google Sheets (bot_registrar source)

**Request Body:**
```json
{
  "spreadsheet_id": "1abc...xyz"
}
```

**Response:** `202 Accepted`

### 3. POST /migration/excel
Загрузка и миграция из Excel файла (academic_group source)

**Content-Type:** `multipart/form-data`

**Form Data:**
- `file`: Excel файл (.xlsx)

**Response:** `202 Accepted`

### 4. GET /migration/jobs/{id}
Получение статуса конкретной миграции

**Response:** `200 OK`
```json
{
  "id": 1,
  "source_type": "excel",
  "source_identifier": "academic_groups.xlsx",
  "status": "running",
  "total": 155000,
  "processed": 50000,
  "failed": 123,
  "started_at": "2024-01-15T10:00:00Z",
  "completed_at": null
}
```

### 5. GET /migration/jobs
Получение списка всех миграций

**Response:** `200 OK` - массив объектов MigrationJobResponse

## Обновление Swagger документации

### Автоматическая генерация

```bash
# Из директории migration-service
make swagger
```

### Ручная генерация

```bash
swag init -g cmd/migration/main.go -o internal/infrastructure/http/docs
```

## Добавление новых endpoints

1. Добавьте Swagger аннотации к handler методу:

```go
// @Summary      Your endpoint summary
// @Description  Detailed description
// @Tags         migration
// @Accept       json
// @Produce      json
// @Param        id path int true "ID parameter"
// @Success      200 {object} YourResponseType
// @Failure      400 {object} ErrorResponse
// @Security     Bearer
// @Router       /your/endpoint [get]
func (h *Handler) YourEndpoint(w http.ResponseWriter, r *http.Request) {
    // implementation
}
```

2. Регенерируйте Swagger документацию:

```bash
make swagger
```

3. Перезапустите сервис и проверьте Swagger UI

## Swagger Annotations Reference

### Основные теги

- `@Summary` - Краткое описание endpoint
- `@Description` - Подробное описание
- `@Tags` - Группировка endpoints (например, "migration")
- `@Accept` - Тип принимаемых данных (json, multipart/form-data)
- `@Produce` - Тип возвращаемых данных (обычно json)
- `@Param` - Параметры запроса
- `@Success` - Успешный ответ с кодом и типом
- `@Failure` - Ошибка с кодом и типом
- `@Security` - Требования безопасности (Bearer token)
- `@Router` - Путь и HTTP метод

### Типы параметров

- `path` - Параметр в URL пути (/jobs/{id})
- `query` - Query параметр (?limit=10)
- `body` - Тело запроса (JSON)
- `formData` - Данные формы (для multipart/form-data)
- `header` - HTTP заголовок

### Примеры

**Path параметр:**
```go
// @Param id path int true "Migration Job ID"
```

**Query параметр:**
```go
// @Param limit query int false "Limit" default(50)
```

**Body параметр:**
```go
// @Param request body StartDatabaseMigrationRequest true "Migration request"
```

**Form Data:**
```go
// @Param file formData file true "Excel file"
```

## Troubleshooting

### Swagger UI не открывается

1. Проверьте, что сервис запущен:
```bash
curl http://localhost:8084/health
```

2. Проверьте, что Swagger файлы сгенерированы:
```bash
ls -la internal/infrastructure/http/docs/
```

3. Проверьте импорт в router.go:
```go
_ "migration-service/internal/infrastructure/http/docs"
```

### Изменения не отображаются

1. Регенерируйте Swagger:
```bash
make swagger
```

2. Перезапустите сервис

3. Очистите кэш браузера (Ctrl+Shift+R)

### Ошибки при генерации

Убедитесь, что:
- Установлен swag: `go install github.com/swaggo/swag/cmd/swag@latest`
- Все аннотации корректны
- Типы данных существуют и экспортированы

## Дополнительная информация

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [Swagger Specification](https://swagger.io/specification/)
- [OpenAPI 3.0](https://spec.openapis.org/oas/v3.0.0)
