# Реализация HTTP метода /bot/me в auth-service

## Что было реализовано

✅ **HTTP endpoint `/bot/me`** в auth-service для получения информации о боте (имя и ссылка для добавления)

## Архитектура решения

### 1. Domain Layer
- **`internal/domain/maxbot_client.go`** - интерфейс `MaxBotClient` и структура `BotInfo`
- **`internal/domain/errors.go`** - добавлена ошибка `ErrMaxBotUnavailable`

### 2. Use Case Layer  
- **`internal/usecase/auth_service.go`** - добавлен метод `GetBotInfo()` и поле `maxBotClient`

### 3. Infrastructure Layer
- **`internal/infrastructure/maxbot/client.go`** - gRPC клиент для MaxBot service
- **`internal/infrastructure/maxbot/mock_client.go`** - mock клиент для тестирования
- **`internal/infrastructure/http/handler.go`** - HTTP handler для `/bot/me`

### 4. HTTP Router
- **`internal/infrastructure/http/router.go`** - добавлен маршрут `/bot/me`

### 5. Configuration
- **`internal/config/config.go`** - уже поддерживает `MAXBOT_SERVICE_ADDR`
- **`cmd/auth/main.go`** - инициализация MaxBot клиента

## API Specification

### Endpoint
```
GET /bot/me
```

### Response (Success)
```json
{
  "name": "MAX Bot",
  "add_link": "https://max.ru/add-bot"
}
```

### Response (Error)
```json
{
  "error": "EXTERNAL_SERVICE_ERROR", 
  "message": "MaxBot service error"
}
```

## Конфигурация

### Environment Variables
- `MAXBOT_SERVICE_ADDR` - адрес MaxBot gRPC сервиса (например: `localhost:9095`)

### Fallback поведение
- Если `MAXBOT_SERVICE_ADDR` не настроен → используется mock клиент
- Если MaxBot сервис недоступен → используется mock клиент как fallback

## Swagger Documentation

✅ Endpoint автоматически добавлен в Swagger документацию:
- Доступен по адресу: `http://localhost:8080/swagger/index.html`
- Включает описание, параметры и примеры ответов

## Тестирование

### Unit Tests
```bash
go test ./internal/infrastructure/http/ -v -run TestHandler_GetBotMe
```

### Integration Test Script
```bash
./test_bot_endpoint.sh
```

## Зависимости

### Go Modules
- Добавлена локальная зависимость на `maxbot-service` в `go.mod`
- Используются protobuf определения из `maxbot-service/api/proto`

### gRPC Communication
- Клиент подключается к MaxBot service через gRPC
- Timeout: 5 секунд на запрос
- Автоматический fallback на mock при ошибках

## Файлы, которые были изменены/добавлены

### Новые файлы:
- `internal/domain/maxbot_client.go`
- `internal/infrastructure/maxbot/client.go` 
- `internal/infrastructure/maxbot/mock_client.go`
- `internal/infrastructure/http/bot_handler_test.go`
- `BOT_ENDPOINT_DOCUMENTATION.md`
- `test_bot_endpoint.sh`

### Измененные файлы:
- `internal/domain/errors.go` - добавлена ошибка
- `internal/usecase/auth_service.go` - добавлен MaxBot клиент
- `internal/infrastructure/http/handler.go` - добавлен endpoint
- `internal/infrastructure/http/router.go` - добавлен маршрут  
- `cmd/auth/main.go` - инициализация клиента
- `go.mod` - добавлена зависимость
- `README.md` - обновлена документация
- Swagger документация (автогенерация)

## Использование

### cURL
```bash
curl -X GET http://localhost:8080/bot/me
```

### JavaScript
```javascript
fetch('/bot/me').then(r => r.json()).then(console.log)
```

### Результат

**С реальным MaxBot service:**
```json
{
  "name": "Ваш реальный бот",
  "add_link": "https://max.ru/bot/your_bot_username"
}
```

**С mock клиентом (fallback):**
```json
{
  "name": "Digital University Bot",
  "add_link": "https://max.ru/bot/digital_university_bot"
}
```

## Следующие шаги

1. **Запуск MaxBot service** для получения реальных данных вместо mock
2. **Настройка переменной окружения** `MAXBOT_SERVICE_ADDR=localhost:9095`
3. **Интеграционное тестирование** с реальным MaxBot service
4. **Мониторинг** добавить метрики для вызовов MaxBot service

## Совместимость

- ✅ Обратная совместимость сохранена
- ✅ Graceful degradation при недоступности MaxBot service  
- ✅ Следует архитектурным принципам проекта
- ✅ Соответствует стандартам кодирования