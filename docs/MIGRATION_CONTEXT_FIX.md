# Migration Service - Context Cancellation Fix

## Проблема

При запуске миграций (Excel, Database, Google Sheets) возникали ошибки:

```
Excel migration failed: failed to create migration job: context canceled
Excel migration failed: failed to create migration job: pq: canceling statement due to user request
```

## Причина

В коде использовался `r.Context()` (request context) для фоновых операций миграции:

```go
// ❌ НЕПРАВИЛЬНО
go func() {
    jobID, err := h.excelUseCase.Execute(r.Context(), filePath)
    // ...
}()
```

**Проблема:** Когда HTTP запрос завершается (клиент получает ответ 202 Accepted), request context отменяется. Это приводит к отмене всех операций, использующих этот context, включая:
- Создание записи в базе данных
- Чтение Excel файла
- Обработку данных

## Решение

Использовать `context.Background()` для долгих фоновых операций:

```go
// ✅ ПРАВИЛЬНО
go func() {
    ctx := context.Background()
    jobID, err := h.excelUseCase.Execute(ctx, filePath)
    // ...
}()
```

**Почему это работает:**
- `context.Background()` создает независимый context
- Он не связан с HTTP запросом
- Не отменяется когда запрос завершается
- Миграция продолжает работать в фоне

## Изменения

### Файл: `migration-service/internal/infrastructure/http/handler.go`

#### 1. Добавлен import

```go
import (
    "context"  // ← Добавлено
    "encoding/json"
    "io"
    // ...
)
```

#### 2. Исправлен StartDatabaseMigration

```go
// Start migration in background with independent context
// Use context.Background() instead of r.Context() to prevent cancellation
go func() {
    ctx := context.Background()  // ← Изменено
    jobID, err := h.databaseUseCase.Execute(ctx, req.SourceIdentifier)
    if err != nil {
        log.Printf("Database migration failed: %v", err)
    } else {
        log.Printf("Database migration completed with job ID: %d", jobID)
    }
}()
```

#### 3. Исправлен StartGoogleSheetsMigration

```go
// Start migration in background with independent context
// Use context.Background() instead of r.Context() to prevent cancellation
go func() {
    ctx := context.Background()  // ← Изменено
    jobID, err := h.googleSheetsUseCase.Execute(ctx, req.SpreadsheetID)
    if err != nil {
        log.Printf("Google Sheets migration failed: %v", err)
    } else {
        log.Printf("Google Sheets migration completed with job ID: %d", jobID)
    }
}()
```

#### 4. Исправлен StartExcelMigration

```go
// Start migration in background with independent context
// Use context.Background() instead of r.Context() to prevent cancellation
// when HTTP request completes
go func() {
    ctx := context.Background()  // ← Изменено
    jobID, err := h.excelUseCase.Execute(ctx, filePath)
    if err != nil {
        log.Printf("Excel migration failed: %v", err)
    } else {
        log.Printf("Excel migration completed with job ID: %d", jobID)
    }

    // Clean up file after migration
    os.Remove(filePath)
}()
```

## Когда использовать какой context

### Request Context (`r.Context()`)

Используйте для операций, которые должны завершиться вместе с HTTP запросом:

```go
// ✅ Правильно - синхронная операция
func (h *Handler) GetData(w http.ResponseWriter, r *http.Request) {
    data, err := h.service.FetchData(r.Context())
    // ...
}
```

**Когда использовать:**
- Синхронные операции
- Операции, которые должны прерваться если клиент отключился
- Короткие операции (< 30 секунд)

### Background Context (`context.Background()`)

Используйте для долгих фоновых операций:

```go
// ✅ Правильно - асинхронная операция
func (h *Handler) StartJob(w http.ResponseWriter, r *http.Request) {
    go func() {
        ctx := context.Background()
        h.service.LongRunningJob(ctx)
    }()
    
    w.WriteHeader(http.StatusAccepted)
}
```

**Когда использовать:**
- Асинхронные операции в goroutines
- Долгие операции (минуты, часы)
- Операции, которые должны завершиться независимо от клиента
- Batch обработка
- Миграции данных

### Context с таймаутом

Для фоновых операций с ограничением времени:

```go
// ✅ Правильно - фоновая операция с таймаутом
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
    defer cancel()
    
    h.service.LongRunningJob(ctx)
}()
```

## Проверка

### До исправления

```bash
# Запустить миграцию
curl -X POST http://localhost:8084/migration/excel \
  -F "file=@chats.xlsx"

# Ответ: 202 Accepted
# Логи: Excel migration failed: context canceled ❌
```

### После исправления

```bash
# Запустить миграцию
curl -X POST http://localhost:8084/migration/excel \
  -F "file=@chats.xlsx"

# Ответ: 202 Accepted
# Логи: Excel migration completed with job ID: 1 ✅
```

### Проверка статуса

```bash
# Получить статус миграции
curl http://localhost:8084/migration/jobs/1

# Ответ:
{
  "id": 1,
  "source_type": "excel",
  "status": "completed",  # ✅ Успешно завершена
  "total": 155000,
  "processed": 155000,
  "failed": 0
}
```

## Best Practices

### 1. Всегда используйте Background Context для goroutines

```go
// ❌ НЕПРАВИЛЬНО
go func() {
    h.service.DoWork(r.Context())
}()

// ✅ ПРАВИЛЬНО
go func() {
    ctx := context.Background()
    h.service.DoWork(ctx)
}()
```

### 2. Добавляйте таймауты для долгих операций

```go
// ✅ ПРАВИЛЬНО
go func() {
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
    defer cancel()
    
    h.service.DoWork(ctx)
}()
```

### 3. Логируйте начало и конец операций

```go
go func() {
    ctx := context.Background()
    log.Printf("Starting migration...")
    
    err := h.service.Migrate(ctx)
    if err != nil {
        log.Printf("Migration failed: %v", err)
    } else {
        log.Printf("Migration completed successfully")
    }
}()
```

### 4. Очищайте ресурсы после завершения

```go
go func() {
    ctx := context.Background()
    
    // Выполнить работу
    h.service.DoWork(ctx)
    
    // Очистить ресурсы
    os.Remove(tempFile)
    cleanup()
}()
```

## Связанные проблемы

Эта проблема может возникать в любых асинхронных операциях:

- ✅ Миграции данных
- ✅ Batch обработка
- ✅ Отправка email/уведомлений
- ✅ Генерация отчетов
- ✅ Экспорт данных
- ✅ Импорт файлов

## Тестирование

### Unit тесты

```go
func TestMigration_WithBackgroundContext(t *testing.T) {
    // Создать handler
    handler := NewHandler(...)
    
    // Запустить миграцию
    req := httptest.NewRequest("POST", "/migration/excel", body)
    w := httptest.NewRecorder()
    
    handler.StartExcelMigration(w, req)
    
    // Проверить что ответ 202
    if w.Code != http.StatusAccepted {
        t.Errorf("Expected 202, got %d", w.Code)
    }
    
    // Подождать завершения
    time.Sleep(5 * time.Second)
    
    // Проверить что миграция завершилась
    job, _ := handler.jobRepo.GetByID(context.Background(), 1)
    if job.Status != "completed" {
        t.Errorf("Expected completed, got %s", job.Status)
    }
}
```

## Статус

✅ **Исправлено** - Все миграции теперь используют `context.Background()` и работают корректно в фоне.

## Документация

- [Migration Service Implementation](./migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)
- [Migration Approach](./MIGRATION_APPROACH.md)
- [Migration Service Fix](./MIGRATION_SERVICE_FIX.md)
