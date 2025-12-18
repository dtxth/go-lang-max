# Profile Cache Service

Сервис кэширования профилей пользователей MAX Messenger для получения имен через webhook события.

## Обзор

Profile Cache Service предоставляет кэширование профильной информации пользователей MAX Messenger с использованием Redis. Сервис поддерживает:

- Хранение профилей с TTL (по умолчанию 30 дней)
- Отслеживание источника данных (webhook, пользовательский ввод, по умолчанию)
- Приоритизацию имен (пользовательское > MAX профиль > по умолчанию)
- Статистику покрытия профилей

## Архитектура

```
Domain Layer:
├── ProfileCacheService (interface)
├── UserProfileCache (struct)
├── ProfileSource (enum)
└── ProfileUpdates (struct)

Infrastructure Layer:
├── ProfileRedisCache (Redis implementation)
├── RedisClient (Redis connection)
└── Configuration (Redis settings)
```

## Конфигурация

Добавьте следующие переменные окружения:

```bash
# Redis configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
PROFILE_TTL=720h  # 30 days
```

## Использование

### 1. Создание сервиса

```go
import (
    "maxbot-service/internal/config"
    "maxbot-service/internal/infrastructure/cache"
)

cfg := config.Load()
profileCache, err := cache.NewProfileCacheService(cfg)
if err != nil {
    log.Fatal("Failed to create profile cache:", err)
}
```

### 2. Сохранение профиля из webhook

```go
// Из webhook события MAX Messenger
profile := domain.UserProfileCache{
    UserID:       "max_user_123",
    MaxFirstName: "Иван",
    MaxLastName:  "Петров",
    Source:       domain.SourceWebhook,
}

err := profileCache.StoreProfile(ctx, profile.UserID, profile)
```

### 3. Получение профиля

```go
profile, err := profileCache.GetProfile(ctx, "max_user_123")
if err != nil {
    return err
}

if profile != nil {
    displayName := profile.GetDisplayName()
    hasFullName := profile.HasFullName()
}
```

### 4. Обновление профиля

```go
// Обновление пользовательским вводом
userProvidedName := "Иван Петрович Петров"
source := domain.SourceUserInput

updates := domain.ProfileUpdates{
    UserProvidedName: &userProvidedName,
    Source:           &source,
}

err := profileCache.UpdateProfile(ctx, userID, updates)
```

### 5. Получение статистики

```go
stats, err := profileCache.GetProfileStats(ctx)
if err != nil {
    return err
}

fmt.Printf("Total profiles: %d\n", stats.TotalProfiles)
fmt.Printf("Profiles with full name: %d\n", stats.ProfilesWithFullName)
```

## Приоритет имен

Система использует следующий приоритет для отображения имен:

1. **user_provided_name** - имя, предоставленное пользователем явно
2. **max_first_name + max_last_name** - полное имя из MAX профиля
3. **max_first_name** - только имя из MAX профиля
4. **""** - пустая строка, если нет данных

```go
displayName := profile.GetDisplayName()
```

## Источники данных

- `SourceWebhook` - данные получены из webhook событий MAX
- `SourceUserInput` - данные предоставлены пользователем явно
- `SourceDefault` - значения по умолчанию

## Обработка ошибок

Сервис спроектирован для graceful degradation:

- Если Redis недоступен, методы возвращают ошибки
- Если профиль не найден, `GetProfile` возвращает `nil` без ошибки
- Некорректные данные в кэше игнорируются при получении статистики

## Тестирование

Запуск тестов:

```bash
cd maxbot-service
go test -v ./internal/infrastructure/cache/
```

Тесты требуют доступного Redis сервера на `localhost:6379`.

## Интеграция с Employee Service

Пример интеграции с сервисом сотрудников:

```go
// При создании сотрудника
func (s *EmployeeService) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) error {
    // Если имя не предоставлено, пытаемся получить из кэша профилей
    if req.FirstName == "" && req.MaxID != "" {
        profile, err := s.profileCache.GetProfile(ctx, req.MaxID)
        if err == nil && profile != nil {
            displayName := profile.GetDisplayName()
            if displayName != "" {
                req.FirstName = displayName
            }
        }
    }
    
    // Продолжаем создание сотрудника...
}
```

## Мониторинг

Рекомендуемые метрики для мониторинга:

- Количество профилей в кэше
- Процент профилей с полными именами
- Распределение по источникам данных
- Частота обновлений профилей
- Ошибки Redis соединения