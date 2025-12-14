# Исправление архитектуры: Удаление University из chat-service

## Проблема

Сущность `University` была неправильно дублирована в `chat-service`, что нарушало принципы микросервисной архитектуры и DDD (Domain-Driven Design).

### Что было неправильно:
- `University` определена в двух сервисах: `structure-service` (правильно) и `chat-service` (неправильно)
- `chat-service` создавал и управлял университетами через `CreateOrGetUniversity`
- `migration-service` использовал `chat-service` для создания университетов
- Нарушение границ bounded context'ов

## Решение

### 1. Удалено из chat-service:
- `internal/domain/university_repository.go`
- `internal/infrastructure/repository/university_postgres.go`
- Поле `University` из доменной модели `Chat`
- Метод `CreateOrGetUniversity` из usecase
- HTTP endpoint `/universities`
- Все ссылки на таблицу `universities` в SQL запросах

### 2. Добавлено в structure-service:
- gRPC метод `CreateOrGetUniversity` в proto файле
- Реализация метода в gRPC handler
- Метод в usecase для создания/получения университета

### 3. Обновлен migration-service:
- Использует `structure-service` вместо `chat-service` для создания университетов
- Обновлены gRPC клиенты
- Исправлены типы данных (`UniversityID` теперь `*int`)

### 4. Миграция БД:
- Создана миграция `000004_remove_universities.up.sql` для удаления таблицы `universities` из chat-service
- Поле `university_id` в таблице `chats` остается как ссылка на `structure-service`

## Архитектура после исправления

```
┌─────────────────┐    gRPC     ┌─────────────────┐
│ migration-      │────────────▶│ structure-      │
│ service         │             │ service         │
└─────────────────┘             └─────────────────┘
         │                               │
         │ gRPC/HTTP                     │ Manages
         ▼                               ▼
┌─────────────────┐             ┌─────────────────┐
│ chat-service    │             │ universities    │
│                 │             │ table           │
│ university_id   │─────────────▶│                 │
│ (reference only)│             │                 │
└─────────────────┘             └─────────────────┘
```

## Преимущества

1. **Правильное разделение ответственности**: Только `structure-service` управляет университетами
2. **Соблюдение DDD**: Каждый сервис управляет только своими доменными сущностями  
3. **Упрощение архитектуры**: Убрана дублированная логика
4. **Масштабируемость**: Легче добавлять новую функциональность для университетов
5. **Консистентность данных**: Единый источник истины для университетов

## Обратная совместимость

- Поле `university_id` в чатах сохранено для ссылки на университеты
- API chat-service не изменился (кроме удаления `/universities` endpoint)
- Существующие чаты продолжат работать после применения миграции

## Применение изменений

1. Применить миграцию БД: `000004_remove_universities.up.sql`
2. Перезапустить все сервисы
3. Убедиться, что `structure-service` доступен по gRPC для `migration-service`

## Статус

✅ **Завершено**
- Все изменения реализованы
- Все сервисы успешно компилируются
- Все тесты проходят
- Готово к развертыванию