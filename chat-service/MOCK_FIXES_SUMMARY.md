# Исправление моков после добавления нового метода

## Проблема
После добавления нового метода `GetAllWithSortingAndSearch` в интерфейс `ChatRepository`, все существующие моки перестали компилироваться, так как не реализовывали новый метод.

## Исправленные файлы

### 1. `internal/usecase/add_administrator_with_permission_check_test.go`
- Добавлен метод `GetAllWithSortingAndSearch` в мок `mockChatRepoForAdd`

### 2. `internal/usecase/list_chats_with_role_filter_test.go`
- Добавлен метод `GetAllWithSortingAndSearch` в мок `MockChatRepository`
- Метод делегирует вызов к существующему методу `Search`

### 3. `internal/usecase/search_chats_test.go`
- Добавлен метод `GetAllWithSortingAndSearch` в мок `MockChatRepositoryForSearch`
- Метод делегирует вызов к существующему методу `Search`

### 4. `internal/usecase/remove_administrator_with_validation_test.go`
- Добавлен метод `GetAllWithSortingAndSearch` в мок `mockChatRepoForRemove`

## Реализация в моках

Все моки реализуют новый метод одним из способов:

### Простая заглушка (для тестов, не использующих новый функционал):
```go
func (m *mockChatRepo) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
    return nil, 0, nil
}
```

### Делегирование к существующему методу (для тестов поиска):
```go
func (m *MockChatRepository) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string, filter *domain.ChatFilter) ([]*domain.Chat, int, error) {
    return m.Search(search, limit, offset, filter)
}
```

## Результат
- Все тесты успешно компилируются и проходят
- Обратная совместимость сохранена
- Новый функционал готов к использованию

## Статус тестов
```
✅ chat-service/internal/infrastructure/http - PASS
✅ chat-service/internal/usecase - PASS
✅ Общая компиляция проекта - SUCCESS
```