# Использование актуального количества участников

## Обзор

Система поддерживает получение актуального количества участников чатов из MAX API с использованием кэширования и фонового обновления.

## Быстрый старт

### 1. Настройка Redis

```bash
# Запуск Redis через Docker
docker run -d --name chat-redis -p 6379:6379 redis:7-alpine

# Или через Docker Compose
docker-compose -f docker-compose.participants.yml up -d chat-redis
```

### 2. Конфигурация переменных окружения

```bash
# Redis подключение
export REDIS_URL="redis://localhost:6379/0"

# Настройки кэширования
export PARTICIPANTS_CACHE_TTL="1h"
export PARTICIPANTS_UPDATE_INTERVAL="15m"
export PARTICIPANTS_FULL_UPDATE_HOUR="3"
export PARTICIPANTS_BATCH_SIZE="50"

# Включение функций
export PARTICIPANTS_ENABLE_BACKGROUND_SYNC="true"
export PARTICIPANTS_ENABLE_LAZY_UPDATE="true"
```

### 3. Запуск сервиса

```bash
# С поддержкой участников
docker-compose -f docker-compose.yml -f docker-compose.participants.yml up -d

# Или локально
cd chat-service
go run cmd/chat/main.go
```

## API Endpoints

### Получение чатов с актуальными участниками

```bash
# Обычный запрос - возвращает кэшированные или обновленные данные
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8082/chats/all?limit=10"
```

### Принудительное обновление участников

```bash
# Обновить количество участников для конкретного чата
curl -X POST \
     -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8082/chats/123/refresh-participants"
```

Ответ:
```json
{
  "status": "updated",
  "chat_id": 123,
  "participants_count": 42,
  "updated_at": "2025-01-15T10:30:00Z",
  "source": "api"
}
```

## Мониторинг

### Проверка состояния Redis

```bash
# Подключение к Redis
redis-cli -h localhost -p 6379

# Проверка ключей участников
KEYS chat_participants:*

# Просмотр данных
GET chat_participants:123
```

### Логи сервиса

```bash
# Просмотр логов обновления участников
docker logs chat-service | grep participants

# Примеры логов:
# INFO  Updated stale participants data updated_count=15 duration=2.3s
# DEBUG Successfully updated participants count chat_id=123 count=42
# ERROR Failed to get chat info from MAX API chat_id=456 error="timeout"
```

## Производительность

### Ожидаемые показатели

- **Время ответа**: 50-100ms (из кэша), 1-3s (с обновлением)
- **Hit rate кэша**: 85-95%
- **Нагрузка на MAX API**: ~100-200 запросов/час
- **Актуальность**: 90% данных свежее 1 часа

### Оптимизация

```bash
# Увеличить TTL кэша для снижения нагрузки
export PARTICIPANTS_CACHE_TTL="2h"

# Уменьшить частоту обновления
export PARTICIPANTS_UPDATE_INTERVAL="30m"

# Увеличить размер батча
export PARTICIPANTS_BATCH_SIZE="100"
```

## Отключение функциональности

```bash
# Отключить фоновое обновление
export PARTICIPANTS_ENABLE_BACKGROUND_SYNC="false"

# Отключить ленивое обновление
export PARTICIPANTS_ENABLE_LAZY_UPDATE="false"

# Полное отключение интеграции
export PARTICIPANTS_INTEGRATION_DISABLED="true"
```

## Troubleshooting

### Проблема: Медленные ответы API

**Причина**: Частые обращения к MAX API

**Решение**:
```bash
# Увеличить TTL кэша
export PARTICIPANTS_CACHE_TTL="4h"

# Проверить hit rate
redis-cli info stats | grep keyspace_hits
```

### Проблема: Устаревшие данные

**Причина**: Отключено фоновое обновление

**Решение**:
```bash
# Включить фоновое обновление
export PARTICIPANTS_ENABLE_BACKGROUND_SYNC="true"

# Уменьшить интервал обновления
export PARTICIPANTS_UPDATE_INTERVAL="10m"
```

### Проблема: Ошибки MAX API

**Причина**: Rate limiting или недоступность API

**Решение**:
```bash
# Увеличить интервал между батчами
export PARTICIPANTS_BATCH_SIZE="25"

# Проверить логи MAX API
docker logs maxbot-service | grep "rate limit"
```

### Проблема: Высокое потребление памяти Redis

**Причина**: Большое количество кэшированных данных

**Решение**:
```bash
# Настроить eviction policy
redis-cli CONFIG SET maxmemory-policy allkeys-lru
redis-cli CONFIG SET maxmemory 512mb

# Уменьшить TTL
export PARTICIPANTS_CACHE_TTL="30m"
```

## Метрики для мониторинга

### Ключевые метрики

1. **participants_cache_hit_rate** - процент попаданий в кэш
2. **participants_update_duration** - время обновления данных
3. **participants_max_api_errors** - ошибки MAX API
4. **participants_stale_data_count** - количество устаревших данных

### Настройка алертов

```yaml
# Пример для Prometheus
- alert: LowParticipantsCacheHitRate
  expr: participants_cache_hit_rate < 0.8
  for: 5m
  annotations:
    summary: "Низкий hit rate кэша участников"

- alert: HighParticipantsAPIErrors
  expr: rate(participants_max_api_errors[5m]) > 0.05
  for: 2m
  annotations:
    summary: "Высокая частота ошибок MAX API"
```

## Миграция существующих данных

### Первоначальное заполнение кэша

```bash
# Запуск полного обновления вручную
curl -X POST \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     "http://localhost:8082/admin/participants/full-update"
```

### Проверка качества данных

```bash
# Сравнение данных БД и MAX API для выборки чатов
curl -H "Authorization: Bearer $TOKEN" \
     "http://localhost:8082/admin/participants/validate?sample_size=100"
```