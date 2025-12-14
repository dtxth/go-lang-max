# Реализация API для редактирования названий элементов структуры

## Обзор

Успешно реализована функциональность для редактирования названий всех элементов иерархической структуры университета через REST API.

## Реализованные endpoints

### 1. Обновление названия университета
- **Endpoint**: `PUT /universities/{id}/name`
- **Описание**: Обновляет название университета по ID
- **Валидация**: ID должен быть корректным, название не пустое

### 2. Обновление названия филиала  
- **Endpoint**: `PUT /branches/{id}/name`
- **Описание**: Обновляет название филиала по ID
- **Валидация**: ID должен быть корректным, название не пустое

### 3. Обновление названия факультета
- **Endpoint**: `PUT /faculties/{id}/name`  
- **Описание**: Обновляет название факультета по ID
- **Валидация**: ID должен быть корректным, название не пустое

### 4. Обновление номера группы
- **Endpoint**: `PUT /groups/{id}/name`
- **Описание**: Обновляет номер группы по ID  
- **Валидация**: ID должен быть корректным, номер не пустой

## Архитектурные изменения

### Domain Layer
- Добавлен `UpdateNameRequest` в `domain/structure.go`
- Расширен интерфейс `StructureServiceInterface` новыми методами:
  - `UpdateUniversityName(id int64, name string) error`
  - `UpdateBranchName(id int64, name string) error`
  - `UpdateFacultyName(id int64, name string) error`
  - `UpdateGroupName(id int64, name string) error`
  - `GetBranchByID(id int64) (*Branch, error)`
  - `GetFacultyByID(id int64) (*Faculty, error)`

### Use Case Layer
- Реализованы новые методы в `usecase/structure_service.go`:
  - Получение существующей записи по ID
  - Обновление поля name/number
  - Сохранение изменений через repository

### Infrastructure Layer
- **HTTP Handlers**: Добавлены 4 новых handler'а в `handler.go`
- **Routing**: Обновлен `router.go` с новыми маршрутами
- **Repository**: Используются существующие методы PostgreSQL репозитория

## Тестирование

### Unit Tests
- **Файл**: `handler_name_update_test.go`
- **Покрытие**: 8 тестов для всех сценариев
- **Тестируемые случаи**:
  - Успешное обновление названий
  - Некорректный ID
  - Пустое название
  - Несуществующая запись
  - Правильность маршрутизации

### Integration Tests
- **Файл**: `structure_name_update_integration_test.go`
- **Тестируемые сценарии**:
  - Полный цикл обновления с проверкой в `/universities/{id}/structure`
  - Обработка ошибок (404, 400)
  - Восстановление исходных данных

## Swagger Documentation

- Автоматически сгенерирована документация для всех новых endpoints
- Доступна по адресу: `http://localhost:8083/swagger/`
- Включает описания параметров, тел запросов и ответов

## Валидация и обработка ошибок

### Валидация входных данных
- Проверка корректности ID (должен быть числом)
- Проверка названия (не может быть пустым или содержать только пробелы)

### HTTP Status Codes
- **200 OK** - Успешное обновление
- **400 Bad Request** - Некорректные данные запроса
- **404 Not Found** - Элемент не найден
- **405 Method Not Allowed** - Неподдерживаемый HTTP метод
- **500 Internal Server Error** - Внутренняя ошибка сервера

## Интеграция с существующей системой

### Совместимость
- Все изменения обратно совместимы
- Существующие endpoints работают без изменений
- Новые методы не влияют на производительность существующих

### Обновление структуры
- После изменения названия обновляется поле `updated_at`
- Изменения сразу отражаются в API `/universities/{id}/structure`
- Связанные элементы остаются без изменений

## Примеры использования

### JavaScript/Frontend
```javascript
// Обновление названия университета
const response = await fetch('/universities/1/name', {
  method: 'PUT',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ name: 'Новое название' })
});
```

### cURL
```bash
curl -X PUT "http://localhost:8083/groups/482/name" \
  -H "Content-Type: application/json" \
  -d '{"name": "ВБ(асп)-12"}'
```

## Файлы изменений

### Новые файлы
- `structure-service/NAME_UPDATE_API.md` - Полная документация API
- `structure-service/QUICK_NAME_UPDATE_GUIDE.md` - Быстрое руководство
- `structure-service/internal/infrastructure/http/handler_name_update_test.go` - Unit тесты
- `integration-tests/structure_name_update_integration_test.go` - Интеграционные тесты

### Измененные файлы
- `structure-service/internal/domain/structure_service.go` - Расширен интерфейс
- `structure-service/internal/domain/structure.go` - Добавлен UpdateNameRequest
- `structure-service/internal/usecase/structure_service.go` - Реализованы новые методы
- `structure-service/internal/infrastructure/http/handler.go` - Добавлены handlers
- `structure-service/internal/infrastructure/http/router.go` - Обновлена маршрутизация
- `structure-service/internal/infrastructure/http/handler_pagination_test.go` - Обновлен mock

## Результаты тестирования

```
=== RUN   TestUpdateUniversityName_Success
--- PASS: TestUpdateUniversityName_Success (0.00s)
=== RUN   TestUpdateBranchName_Success  
--- PASS: TestUpdateBranchName_Success (0.00s)
=== RUN   TestUpdateFacultyName_Success
--- PASS: TestUpdateFacultyName_Success (0.00s)
=== RUN   TestUpdateGroupName_Success
--- PASS: TestUpdateGroupName_Success (0.00s)
PASS
ok      structure-service/internal/infrastructure/http  0.572s
```

## Готовность к продакшену

✅ **Реализация завершена**
✅ **Unit тесты покрывают все сценарии**  
✅ **Интеграционные тесты проверяют полный workflow**
✅ **Swagger документация сгенерирована**
✅ **Валидация и обработка ошибок реализованы**
✅ **Обратная совместимость сохранена**

Функциональность готова к использованию в продакшене и полностью интегрирована с существующей архитектурой системы.