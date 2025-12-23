# Известные проблемы и их решения

Документ описывает известные проблемы в проекте и способы их решения.

## 🐛 Текущие проблемы

### 1. Logger Nil Pointer в Middleware (КРИТИЧНО)

**Статус:** 🔴 Требует исправления

**Затронутые сервисы:**
- employee-service
- chat-service  
- structure-service

**Описание:**
При первом HTTP запросе к сервису происходит panic из-за неинициализированного logger в middleware.

**Ошибка:**
```
panic: runtime error: invalid memory address or nil pointer dereference
employee-service/internal/infrastructure/logger.(*Logger).shouldLog(...)
    /app/internal/infrastructure/logger/logger.go:94
employee-service/internal/infrastructure/logger.(*Logger).log(0x0, ...)
    /app/internal/infrastructure/logger/logger.go:56 +0x110
```

**Причина:**
Logger не передается в middleware при инициализации HTTP handler.

**Решение:**
Необходимо убедиться, что logger инициализируется и передается в middleware:

```go
// В main.go или при создании handler
logger := logger.NewLogger(logger.InfoLevel)

// При создании handler
handler := http.NewHandler(
    chatService,
    authClient,
    maxClient,
    logger,  // Передать logger
)
```

**Временное решение:**
Можно добавить проверку на nil в middleware:

```go
// В request_id.go
func (h *Handler) RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := uuid.New().String()
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        
        // Проверка на nil
        if h.logger != nil {
            h.logger.Info(ctx, "Request started", "request_id", requestID)
        }
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

### 2. MaxBot Service требует MAX_BOT_TOKEN

**Статус:** 🟡 Ожидаемое поведение

**Затронутый сервис:**
- maxbot-service

**Описание:**
MaxBot service постоянно перезапускается, требуя MAX_BOT_TOKEN.

**Ошибка:**
```
MAX_BOT_TOKEN environment variable is required but not set. 
Please configure the bot token.
```

**Причина:**
Сервис требует токен для работы с MAX API, но он не настроен в docker-compose.yml.

**Решение:**
Добавить токен в docker-compose.yml:

```yaml
maxbot-service:
  environment:
    MAX_BOT_TOKEN: "your-token-here"
    MAX_API_URL: "https://api.max.com"
```

**Временное решение:**
Для локальной разработки можно сделать токен опциональным или использовать mock.

---

### 3. Employee DB - неправильное имя базы данных

**Статус:** 🟡 Требует проверки

**Затронутый сервис:**
- employee-db

**Описание:**
В логах иногда появляется ошибка о несуществующей базе данных "employee_user".

**Ошибка:**
```
FATAL: database "employee_user" does not exist
```

**Причина:**
Возможно, в коде используется неправильное имя базы данных. В docker-compose.yml база называется "employee_db", а не "employee_user".

**Решение:**
Проверить строку подключения в docker-compose.yml:

```yaml
employee-service:
  environment:
    DATABASE_URL: postgres://employee_user:employee_pass@employee-db:5432/employee_db?sslmode=disable
    #                                                                          ^^^^^^^^^^^
    # Должно быть employee_db, а не employee_user
```

---

## ✅ Решенные проблемы

### Docker Build - Cross-Service Dependencies

**Статус:** ✅ Решено

**Описание:**
Сервисы не могли собраться из-за отсутствия proto файлов из других сервисов.

**Решение:**
- Добавлены replace директивы в go.mod
- Изменен контекст сборки Docker на корневую директорию
- Добавлено копирование proto файлов в Dockerfile
- Автоматическая замена путей через sed

**Документация:**
- [DOCKER_CROSS_SERVICE_DEPENDENCIES.md](./DOCKER_CROSS_SERVICE_DEPENDENCIES.md)

---

## 🔧 Рекомендации по исправлению

### Приоритет 1: Logger Nil Pointer

Это критическая проблема, которая блокирует работу 3 из 5 сервисов.

**Шаги для исправления:**

1. Проверить инициализацию logger в main.go каждого сервиса
2. Убедиться, что logger передается в handler
3. Добавить проверки на nil в middleware
4. Протестировать локально перед деплоем

**Файлы для проверки:**
```
employee-service/cmd/employee/main.go
employee-service/internal/infrastructure/http/handler.go
employee-service/internal/infrastructure/middleware/request_id.go

chat-service/cmd/chat/main.go
chat-service/internal/infrastructure/http/handler.go
chat-service/internal/infrastructure/middleware/request_id.go

structure-service/cmd/structure/main.go
structure-service/internal/infrastructure/http/handler.go
structure-service/internal/infrastructure/middleware/request_id.go
```

### Приоритет 2: MaxBot Token

Это ожидаемое поведение, но можно улучшить:

1. Добавить пример токена в .env.example
2. Сделать токен опциональным для dev окружения
3. Добавить mock режим для тестирования

### Приоритет 3: Database Names

Проверить консистентность имен баз данных во всех конфигурациях.

---

## 🧪 Тестирование после исправлений

После исправления проблем запустите:

```bash
# 1. Запустить тесты
make test

# 2. Пересобрать образы
make build-no-cache

# 3. Запустить сервисы
make up

# 4. Проверить здоровье
make health

# 5. Проверить логи
make logs
```

---

## 📊 Статус сервисов

| Сервис | Сборка | Запуск | Swagger | Проблема |
|--------|--------|--------|---------|----------|
| auth-service | ✅ | ✅ | ✅ | Нет |
| employee-service | ✅ | ❌ | ❌ | Logger nil |
| chat-service | ✅ | ❌ | ❌ | Logger nil |
| structure-service | ✅ | ❌ | ❌ | Logger nil |
| maxbot-service | ✅ | 🔄 | N/A | Требует токен |
| migration-service | ✅ | ✅ | ✅ | Нет |

**Легенда:**
- ✅ Работает
- ❌ Не работает
- 🔄 Перезапускается
- N/A Не применимо

---

## 💡 Полезные команды для отладки

```bash
# Проверить логи конкретного сервиса
docker-compose logs employee-service | grep -i error

# Проверить статус всех контейнеров
docker-compose ps

# Перезапустить проблемный сервис
docker-compose restart employee-service

# Войти в контейнер для отладки
docker exec -it employee-service sh

# Проверить переменные окружения
docker exec employee-service env

# Проверить процессы в контейнере
docker exec employee-service ps aux
```

---

## 📚 Связанные документы

- [Testing and Deployment Guide](./TESTING_AND_DEPLOYMENT.md)
- [Docker Cross-Service Dependencies](./DOCKER_CROSS_SERVICE_DEPENDENCIES.md)
- [Cheatsheet](./CHEATSHEET.md)
- [README](./README.md)

---

## 🔄 История изменений

### 2025-11-30
- ✅ Решена проблема со сборкой Docker (cross-service dependencies)
- 🔴 Обнаружена проблема с logger nil pointer в middleware
- 🟡 Обнаружена проблема с MAX_BOT_TOKEN в maxbot-service

---

## 📝 Заметки

- Проблемы с logger не связаны с изменениями в Docker
- Все образы собираются успешно
- Тесты проходят успешно
- Проблемы возникают только при runtime
