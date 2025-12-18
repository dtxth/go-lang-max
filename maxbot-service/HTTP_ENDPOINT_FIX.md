# Исправление HTTP Endpoint /api/v1/chats/{chat_id}

## Проблема

HTTP endpoint `/api/v1/chats/{chat_id}` в maxbot-service возвращает 404 вместо обработки запроса.

## Диагностика

1. **gRPC работает корректно** - GetChatInfo через gRPC возвращает ожидаемые ошибки 404 для несуществующих чатов
2. **HTTP сервер запускается** - логи показывают успешный запуск на порту 8095
3. **Маршруты не регистрируются** - логи настройки маршрутов не выводятся
4. **Проблема в HTTP сервере** - даже простые endpoints возвращают 404

## Причина

HTTP сервер не доходит до вызова `setupRoutes()` или падает при настройке маршрутов.

## Исправления (применены)

### 1. Исправлена маршрутизация в server.go

```go
// Убрано дублирование маршрутов
api.HandleFunc("/chats/{chat_id}", s.handler.GetChatInfo).Methods("GET")

// Добавлены отладочные логи
fmt.Println("=== SETTING UP HTTP ROUTES ===")
```

### 2. Улучшена функция извлечения chat_id

```go
func extractChatIDFromPath(r *http.Request) string {
    vars := mux.Vars(r)
    chatID := vars["chat_id"]
    log.Printf("[DEBUG] mux.Vars: %+v, chat_id: '%s'", vars, chatID)
    return chatID
}
```

### 3. Добавлены отладочные логи в main.go

```go
log.Println("=== CREATING HTTP HANDLER ===")
log.Println("=== CREATING HTTP SERVER ===")
log.Printf("=== STARTING HTTP SERVER GOROUTINE ===")
```

## Текущий статус

- ✅ gRPC endpoint работает корректно
- ✅ HTTP сервер запускается и базовые endpoints работают (/health, /api/v1/me)
- ❌ HTTP endpoint /api/v1/chats/{chat_id} возвращает 404
- ❌ Логи настройки маршрутов не выводятся
- ❌ API subrouter не регистрирует все маршруты корректно

## Обходное решение

### 1. Используйте gRPC endpoint (рекомендуется)

```bash
# Через существующий тестовый скрипт
./test_chat_info.sh

# Результат для несуществующего чата:
# [ERROR] Failed to get chat info for chat 123456789: HTTP 404: Not Found
```

### 2. Проверка работоспособности других endpoints

```bash
# Health check - работает
curl http://localhost:8095/health
# {"status":"ok","service":"maxbot-service"}

# Bot info - работает  
curl http://localhost:8095/api/v1/me
# {"name":"Digital University Support Bot","add_link":"..."}
```

## Рекомендации

1. **Проверить Docker build context** - возможно, изменения не попадают в контейнер
2. **Исследовать HTTP сервер** - добавить больше отладочной информации
3. **Проверить зависимости** - возможно, проблема в импортах или middleware
4. **Рассмотреть альтернативный HTTP роутер** - если проблема в gorilla/mux

## Тестирование

```bash
# Тест gRPC (работает)
./test_chat_info.sh

# Тест HTTP (не работает)
curl http://localhost:8095/api/v1/chats/123456789
# Ожидается: {"error": "chat_not_found", ...}
# Фактически: 404 page not found

# Тест health check
curl http://localhost:8095/health
# Работает: {"status":"ok","service":"maxbot-service"}
```

## Следующие шаги

1. Исследовать, почему HTTP сервер не вызывает setupRoutes()
2. Проверить, есть ли ошибки при запуске HTTP сервера
3. Рассмотреть создание минимального HTTP сервера для тестирования
4. Проверить совместимость версий gorilla/mux и Go 1.24