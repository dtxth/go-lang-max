# Анализ поведения при недоступности кэша и MAX API

## Сценарии отказов и поведение системы

### 1. 🔴 Redis кэш недоступен

**Что происходит:**
```go
// В enrichChatsWithParticipants
cachedData, err := s.participantsCache.GetMultiple(ctx, chatIDs)
if err != nil {
    return chats, err  // ❌ ПРОБЛЕМА: возвращаем ошибку
}
```

**Текущее поведение:** ❌ **API возвращает ошибку**
- Пользователь получает HTTP 500
- Данные из БД не возвращаются вообще

**Ожидаемое поведение:** ✅ **Fallback на данные из БД**

---

### 2. 🔴 MAX API недоступен

**Что происходит:**
```go
// В UpdateSingle
chatInfo, err := s.maxService.GetChatInfo(ctx, maxChatIDInt)
if err != nil {
    s.logger.Error(ctx, "Failed to get chat info from MAX API", ...)
    return s.getFallbackInfo(ctx, chatID)  // ✅ Правильно - fallback на БД
}
```

**Текущее поведение:** ✅ **Fallback на данные из БД**
- Система логирует ошибку
- Возвращает статические данные из PostgreSQL
- API продолжает работать

---

### 3. 🔴 И кэш, И MAX API недоступны

**Что происходит:**
1. Кэш недоступен → API падает с ошибкой (не доходит до MAX API)
2. Данные из БД не возвращаются

**Текущее поведение:** ❌ **Полный отказ API**

---

### 4. 🟡 Кэш доступен, но данные устарели + MAX API недоступен

**Что происходит:**
```go
// Данные в кэше старше staleThreshold (1 час)
if !exists || cachedInfo.UpdatedAt.Before(staleThreshold) {
    // Пытаемся обновить через MAX API асинхронно
    // Если MAX API недоступен - логируем ошибку, но НЕ обновляем данные
}
```

**Текущее поведение:** ✅ **Возвращает устаревшие данные из кэша**
- Пользователь получает данные (пусть и устаревшие)
- Система пытается обновить в фоне
- При неудаче - просто логирует ошибку

---

## 🚨 Критические проблемы

### Проблема 1: Отказ при недоступности Redis

**Код:**
```go
cachedData, err := s.participantsCache.GetMultiple(ctx, chatIDs)
if err != nil {
    return chats, err  // ❌ Возвращаем ошибку вместо fallback
}
```

**Последствия:**
- При падении Redis весь API `/chats/all` становится недоступен
- Пользователи не получают даже базовые данные из БД
- Нарушается принцип graceful degradation

### Проблема 2: Нет индикации источника данных

**Проблема:**
- API не показывает, откуда пришли данные (кэш/БД/API)
- Сложно диагностировать проблемы
- Пользователь не знает, актуальны ли данные

## ✅ Исправления

### 1. Исправить fallback при недоступности кэша

```go
// enrichChatsWithParticipants - ИСПРАВЛЕННАЯ версия
func (s *ChatService) enrichChatsWithParticipants(ctx context.Context, chats []*domain.Chat) ([]*domain.Chat, error) {
    if len(chats) == 0 || s.participantsCache == nil || s.participantsConfig == nil || !s.participantsConfig.EnableLazyUpdate {
        return chats, nil
    }
    
    // Собираем ID чатов для батчевого запроса
    chatIDs := make([]int64, len(chats))
    chatMap := make(map[int64]*domain.Chat)
    for i, chat := range chats {
        chatIDs[i] = chat.ID
        chatMap[chat.ID] = chat
    }
    
    // Получаем данные из кэша с fallback
    cachedData, err := s.participantsCache.GetMultiple(ctx, chatIDs)
    if err != nil {
        // ✅ ИСПРАВЛЕНИЕ: логируем ошибку, но продолжаем работу
        s.logger.Error(ctx, "Failed to get data from cache, using database fallback", map[string]interface{}{
            "error": err.Error(),
        })
        // Возвращаем данные из БД без обогащения
        return chats, nil
    }
    
    // Остальная логика без изменений...
}
```

### 2. Добавить поле источника данных в ответ API

```go
// В domain/chat.go
type Chat struct {
    ID                int64           `json:"id"`
    Name              string          `json:"name"`
    // ... другие поля
    ParticipantsCount int             `json:"participants_count"`
    ParticipantsSource string         `json:"participants_source,omitempty"` // ✅ НОВОЕ ПОЛЕ
    ParticipantsUpdatedAt *time.Time  `json:"participants_updated_at,omitempty"` // ✅ НОВОЕ ПОЛЕ
}
```

### 3. Добавить health check для зависимостей

```go
// В handler.go
// HealthCheck godoc
// @Summary      Проверка состояния сервиса и зависимостей
// @Description  Возвращает статус сервиса, Redis и MAX API
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health/detailed [get]
func (h *Handler) DetailedHealthCheck(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "service": "healthy",
        "timestamp": time.Now(),
        "dependencies": map[string]string{
            "database": "healthy",
            "redis": checkRedisHealth(h.participantsCache),
            "max_api": checkMaxAPIHealth(h.maxService),
        },
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}
```

## 📊 Матрица поведения системы

| Кэш Redis | MAX API | Поведение | Источник данных | HTTP Status |
|-----------|---------|-----------|-----------------|-------------|
| ✅ Доступен | ✅ Доступен | Оптимально | cache/api | 200 |
| ✅ Доступен | ❌ Недоступен | Устаревшие данные | cache | 200 |
| ❌ Недоступен | ✅ Доступен | Данные из БД | database | 200 |
| ❌ Недоступен | ❌ Недоступен | Данные из БД | database | 200 |

## 🔧 Рекомендации по мониторингу

### Метрики для алертов
```yaml
# Prometheus метрики
- participants_cache_errors_total
- participants_max_api_errors_total  
- participants_fallback_usage_total
- participants_data_freshness_seconds

# Алерты
- alert: ParticipantsCacheDown
  expr: participants_cache_errors_total > 0
  for: 1m
  
- alert: ParticipantsMaxAPIDown  
  expr: participants_max_api_errors_total > 0
  for: 2m
  
- alert: ParticipantsDataStale
  expr: participants_data_freshness_seconds > 7200  # 2 часа
  for: 5m
```

### Логирование
```go
// Структурированные логи для мониторинга
s.logger.Warn(ctx, "Using fallback data source", map[string]interface{}{
    "reason": "cache_unavailable",
    "chat_count": len(chats),
    "fallback_source": "database",
})
```

## 🎯 Итоговые рекомендации

1. **✅ Исправить критический баг** - добавить fallback при недоступности Redis
2. **✅ Добавить индикацию источника** - поля `participants_source` и `participants_updated_at`
3. **✅ Улучшить мониторинг** - метрики и алерты для каждого компонента
4. **✅ Добавить health checks** - детальная проверка зависимостей
5. **✅ Документировать поведение** - четкие SLA для каждого сценария

**Приоритет:** 🔥 **КРИТИЧЕСКИЙ** - исправление fallback логики должно быть сделано до production deployment!