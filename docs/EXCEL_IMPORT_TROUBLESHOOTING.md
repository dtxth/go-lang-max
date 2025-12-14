# 🔧 Troubleshooting: Excel Import

## ✅ Что исправлено

### Проблема 1: Nil pointer dereference в logger

**Симптом:**
```
panic: runtime error: invalid memory address or nil pointer dereference
migration-service/internal/infrastructure/logger.(*Logger).Warn(...)
```

**Причина:**  
Logger не был инициализирован в некоторых путях кода.

**Решение:**  
Добавлены проверки `if uc.logger != nil` перед вызовами logger.

**Файл:** `migration-service/internal/usecase/migrate_from_excel.go`

```go
// До:
uc.logger.Warn(ctx, "Skipping row: insufficient columns", ...)

// После:
if uc.logger != nil {
    uc.logger.Warn(ctx, "Skipping row: insufficient columns", ...)
}
```

### Проблема 2: Миграции БД не применены

**Симптом:**  
Поля `external_chat_id`, `add_user`, `add_admin` отсутствуют в БД.

**Решение:**  
Применена миграция `chat-service/migrations/002_add_excel_fields.sql`

**Проверка:**
```bash
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\d chats"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "\d administrators"
```

**Результат:** ✅ Все поля добавлены

## 🔍 Текущее состояние

### Сервисы

```bash
docker-compose ps
```

- ✅ migration-service: Running
- ✅ chat-service: Running  
- ✅ structure-service: Running
- ✅ БД: Все healthy

### Миграции

```bash
# Chat DB
✅ chats.external_chat_id - добавлено
✅ administrators.add_user - добавлено
✅ administrators.add_admin - добавлено
```

### API Endpoints

```bash
# Migration Service
✅ GET  /health - работает
✅ GET  /migration/jobs - работает
✅ POST /migration/excel - работает (но нужен тестовый файл)

# Chat Service  
⚠️  GET  /chats - возвращает ошибку парсинга (нужна проверка)

# Structure Service
✅ GET  /universities - работает (пустой массив)
```

## 📊 Данные в БД

### Chat DB
```sql
SELECT COUNT(*) FROM chats;          -- 0
SELECT COUNT(*) FROM administrators; -- 0
```

### Structure DB
```sql
SELECT COUNT(*) FROM universities; -- 0
SELECT COUNT(*) FROM branches;     -- 0
SELECT COUNT(*) FROM faculties;    -- 0
SELECT COUNT(*) FROM groups;       -- 0
```

**Вывод:** БД пустые, данные не импортировались из-за panic.

## 🧪 Тестирование

### Unit-тесты

```bash
cd migration-service
go test -v ./internal/usecase -run TestExcelFileParsing
```

**Результат:** ✅ Все 6 тестов пройдены

```bash
go test -v ./internal/usecase -run TestNormalizePhone
```

**Результат:** ✅ Все 10 тестов пройдены

```bash
go test -v ./internal/infrastructure/http -run TestUpload
```

**Результат:** ✅ Все 5 тестов пройдены

### Интеграционное тестирование

**Проблема:** Нет тестового Excel файла для загрузки.

**Решение:** Создать тестовый файл вручную или использовать Python/Go скрипт.

## 🚀 Следующие шаги

### 1. Создать тестовый Excel файл

Вариант A: Вручную в Excel/LibreOffice
- Создать файл с 18 колонками
- Добавить заголовок и 1-2 строки данных
- Сохранить как test_import.xlsx

Вариант B: Использовать Python
```bash
pip install openpyxl
python3 create_test_excel.py
```

Вариант C: Использовать существующий файл
- Взять реальный файл "Наименование чатов групп МГТУ в МАХ 17.11.25_ИТОГ.xlsx"

### 2. Загрузить файл

```bash
curl -X POST http://localhost:8084/migration/excel \
  -F "file=@test_import.xlsx"
```

### 3. Проверить результат

```bash
# Получить job_id из ответа
JOB_ID=1

# Проверить статус
curl http://localhost:8084/migration/jobs/$JOB_ID | jq '.'

# Проверить данные в БД
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT * FROM chats;"
docker-compose exec -T chat-db psql -U chat_user -d chat_db -c "SELECT * FROM administrators;"
```

### 4. Проверить логи при ошибках

```bash
# Логи migration-service
docker-compose logs migration-service --tail=100

# Логи chat-service
docker-compose logs chat-service --tail=50

# Логи structure-service
docker-compose logs structure-service --tail=50
```

## 🐛 Известные проблемы

### 1. Chat Service API возвращает ошибку парсинга

**Симптом:**
```bash
curl http://localhost:8082/chats
# jq: parse error: Invalid numeric literal
```

**Возможная причина:**  
- Сервис возвращает не JSON
- Ошибка в handler

**Проверка:**
```bash
curl -v http://localhost:8082/chats
```

**Решение:** Проверить логи chat-service

### 2. Старые jobs в статусе "running"

**Симптом:**  
Jobs с id 2, 3, 4 остались в статусе "running" после перезапуска.

**Причина:**  
Сервис упал с panic, jobs не были завершены.

**Решение:**  
Обновить статус вручную или игнорировать старые jobs.

```sql
-- Обновить статус старых jobs
docker-compose exec -T migration-db psql -U postgres -d migration_db -c \
  "UPDATE migration_jobs SET status='failed' WHERE status='running' AND id < 5;"
```

## ✅ Чек-лист готовности

- [x] Миграции БД применены
- [x] Сервисы запущены
- [x] Unit-тесты пройдены
- [x] Код исправлен (nil pointer fix)
- [x] Сервис пересобран
- [ ] Тестовый Excel файл создан
- [ ] Файл успешно загружен
- [ ] Данные попали в БД
- [ ] Chat Service API работает корректно

## 📝 Команды для быстрой диагностики

```bash
# Проверка сервисов
docker-compose ps

# Проверка логов
docker-compose logs migration-service --tail=50
docker-compose logs chat-service --tail=50

# Проверка БД
./test_migration_api.sh

# Запуск тестов
cd migration-service && go test -v ./internal/usecase -run TestExcelFileParsing

# Пересборка после изменений
docker-compose build migration-service
docker-compose up -d migration-service
```

## 🎯 Итог

**Статус:** Система готова к импорту, но требует тестового файла для проверки.

**Что работает:**
- ✅ Парсинг Excel (тесты пройдены)
- ✅ Нормализация телефонов
- ✅ Загрузка файлов через HTTP
- ✅ Миграции БД применены
- ✅ Nil pointer исправлен

**Что нужно:**
- 📝 Создать тестовый Excel файл
- 🧪 Протестировать полный цикл импорта
- 🔍 Проверить Chat Service API

**Следующий шаг:** Создать тестовый Excel файл и загрузить его.
