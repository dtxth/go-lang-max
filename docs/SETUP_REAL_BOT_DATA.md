# 🤖 Настройка для получения реальных данных бота

## Проблема
Получаете mock данные вместо реальных, даже с настоящим токеном.

## Решение: Пошаговая настройка

### Шаг 1: Подготовка токена
```bash
# Замените на ваш реальный токен MAX API
export MAX_BOT_TOKEN="your-real-max-bot-token-here"

# Убедитесь, что MOCK_MODE отключен
unset MOCK_MODE

# Проверьте переменные
echo "Token: ${MAX_BOT_TOKEN:0:10}..."
echo "Mock mode: ${MOCK_MODE:-disabled}"
```

### Шаг 2: Запуск MaxBot service
```bash
# Терминал 1: MaxBot service
cd maxbot-service

# Проверка конфигурации
./debug_config.sh

# Или запуск напрямую:
# export MAX_BOT_TOKEN="your-token"
# go run cmd/maxbot/main.go
```

**Ожидаемый вывод:**
```
✅ MAX_BOT_TOKEN is set
✅ MOCK_MODE is disabled - will use real client
Starting MaxBot Service...
Configuration loaded - GRPC Port: 9095, HTTP Port: 8095
MAX_BOT_TOKEN validated (length: XX characters)
Initializing Max API client...
Max API client initialized successfully
Starting gRPC server on port 9095
Starting HTTP server on port 8095
```

### Шаг 3: Запуск auth-service
```bash
# Терминал 2: Auth service
cd auth-service

# Настройка подключения к MaxBot
export MAXBOT_SERVICE_ADDR="localhost:9095"

# Запуск
go run cmd/auth/main.go
```

**Ожидаемый вывод:**
```
Initialized MaxBot client (addr: localhost:9095)
Starting HTTP server on port 8080
```

### Шаг 4: Тестирование
```bash
# Терминал 3: Тестирование
curl http://localhost:8080/bot/me
```

**Ожидаемый результат (реальные данные):**
```json
{
  "name": "Your Real Bot Name",
  "add_link": "https://max.ru/bot/your_bot_username"
}
```

**Если получаете mock данные:**
```json
{
  "name": "Digital University Bot",
  "add_link": "https://max.ru/bot/digital_university_bot"
}
```

## Диагностика проблем

### Проблема 1: MaxBot service использует mock
**Симптомы:**
- В логах MaxBot: `"Running in MOCK MODE"`
- Или: `"MAX_BOT_TOKEN environment variable is required"`

**Решение:**
```bash
# Проверьте переменные
env | grep -E "(MAX_BOT_TOKEN|MOCK_MODE)"

# Установите токен
export MAX_BOT_TOKEN="your-real-token"

# Убедитесь, что MOCK_MODE не установлен
unset MOCK_MODE
```

### Проблема 2: Auth-service не подключается к MaxBot
**Симптомы:**
- В логах auth-service: `"Using mock MaxBot client"`
- Или: `"Failed to initialize MaxBot client"`

**Решение:**
```bash
# Проверьте, что MaxBot service запущен
curl -s http://localhost:9095 || echo "MaxBot service not running"

# Установите адрес
export MAXBOT_SERVICE_ADDR="localhost:9095"

# Перезапустите auth-service
```

### Проблема 3: MAX API возвращает ошибку
**Симптомы:**
- В логах MaxBot: `"Failed to get bot info from MAX API"`
- HTTP 500 ошибка

**Решение:**
```bash
# Проверьте токен
curl -H "Authorization: Bearer $MAX_BOT_TOKEN" https://api.max.ru/bot/getMe

# Или проверьте в коде MAX API клиента
```

## Быстрая проверка всей цепочки

```bash
# Запустите полную диагностику
./debug_full_chain.sh
```

## Автоматический запуск

Создайте скрипт `start_real_bot.sh`:
```bash
#!/bin/bash
export MAX_BOT_TOKEN="your-real-token"
export MAXBOT_SERVICE_ADDR="localhost:9095"
unset MOCK_MODE

# Запуск MaxBot service в фоне
cd maxbot-service
go run cmd/maxbot/main.go &
MAXBOT_PID=$!

# Ждем запуска
sleep 3

# Запуск auth-service
cd ../auth-service
go run cmd/auth/main.go &
AUTH_PID=$!

echo "Services started!"
echo "Test: curl http://localhost:8080/bot/me"
echo "Stop: kill $MAXBOT_PID $AUTH_PID"
```

## Проверка результата

✅ **Реальные данные** - имя и username от MAX API
❌ **Mock данные** - "Digital University Bot"

Если все настроено правильно, вы получите реальное имя вашего бота и корректную ссылку!