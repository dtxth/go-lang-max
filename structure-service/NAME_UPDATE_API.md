# API для редактирования названий элементов структуры

Данный документ описывает новые endpoints для редактирования названий элементов иерархической структуры университета.

## Endpoints

### 1. Обновление названия университета

**PUT** `/universities/{id}/name`

Обновляет название университета по ID.

#### Параметры запроса:
- `id` (path, int64) - ID университета

#### Тело запроса:
```json
{
  "name": "Новое название университета"
}
```

#### Ответы:
- **200 OK** - Название успешно обновлено
- **400 Bad Request** - Некорректный ID или пустое название
- **404 Not Found** - Университет не найден
- **500 Internal Server Error** - Внутренняя ошибка сервера

#### Пример запроса:
```bash
curl -X PUT "http://localhost:8083/universities/1/name" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ БЮДЖЕТНОЕ ОБРАЗОВАТЕЛЬНОЕ УЧРЕЖДЕНИЕ ВЫСШЕГО ОБРАЗОВАНИЯ \"НОВЫЙ МАЙКОПСКИЙ ГОСУДАРСТВЕННЫЙ ТЕХНОЛОГИЧЕСКИЙ УНИВЕРСИТЕТ\""
  }'
```

#### Пример ответа:
```json
{
  "message": "university name updated successfully"
}
```

### 2. Обновление названия филиала

**PUT** `/branches/{id}/name`

Обновляет название филиала по ID.

#### Параметры запроса:
- `id` (path, int64) - ID филиала

#### Тело запроса:
```json
{
  "name": "Новое название филиала"
}
```

#### Ответы:
- **200 OK** - Название успешно обновлено
- **400 Bad Request** - Некорректный ID или пустое название
- **404 Not Found** - Филиал не найден
- **500 Internal Server Error** - Внутренняя ошибка сервера

#### Пример запроса:
```bash
curl -X PUT "http://localhost:8083/branches/1/name" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Главный корпус (обновленное название)"
  }'
```

### 3. Обновление названия факультета

**PUT** `/faculties/{id}/name`

Обновляет название факультета по ID.

#### Параметры запроса:
- `id` (path, int64) - ID факультета

#### Тело запроса:
```json
{
  "name": "Новое название факультета"
}
```

#### Ответы:
- **200 OK** - Название успешно обновлено
- **400 Bad Request** - Некорректный ID или пустое название
- **404 Not Found** - Факультет не найден
- **500 Internal Server Error** - Внутренняя ошибка сервера

#### Пример запроса:
```bash
curl -X PUT "http://localhost:8083/faculties/3/name" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Аспирантура и докторантура"
  }'
```

### 4. Обновление номера группы

**PUT** `/groups/{id}/name`

Обновляет номер группы по ID.

#### Параметры запроса:
- `id` (path, int64) - ID группы

#### Тело запроса:
```json
{
  "name": "Новый номер группы"
}
```

#### Ответы:
- **200 OK** - Номер группы успешно обновлен
- **400 Bad Request** - Некорректный ID или пустой номер
- **404 Not Found** - Группа не найдена
- **500 Internal Server Error** - Внутренняя ошибка сервера

#### Пример запроса:
```bash
curl -X PUT "http://localhost:8083/groups/482/name" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "ВБ(асп)-12"
  }'
```

## Валидация

Все endpoints выполняют следующую валидацию:

1. **ID параметр** - должен быть корректным числом
2. **Название** - не может быть пустым или содержать только пробелы
3. **Существование записи** - проверяется существование записи с указанным ID

## Безопасность

В будущих версиях планируется добавить:
- JWT аутентификацию
- Проверку прав доступа (RBAC)
- Аудит изменений

## Интеграция с существующей структурой

После обновления названия элемента:
1. Изменения сразу отражаются в API `/universities/{id}/structure`
2. Обновляется поле `updated_at` для соответствующей записи
3. Все связанные элементы остаются без изменений

## Примеры использования в JavaScript

```javascript
// Обновление названия университета
async function updateUniversityName(universityId, newName) {
  const response = await fetch(`/universities/${universityId}/name`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name: newName })
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  return await response.json();
}

// Обновление названия группы
async function updateGroupName(groupId, newName) {
  const response = await fetch(`/groups/${groupId}/name`, {
    method: 'PUT',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ name: newName })
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  return await response.json();
}
```

## Ошибки и их обработка

### Типичные ошибки:

1. **400 Bad Request**
   ```json
   {
     "error": "invalid university id"
   }
   ```

2. **400 Bad Request** (пустое название)
   ```json
   {
     "error": "name cannot be empty"
   }
   ```

3. **404 Not Found**
   ```json
   {
     "error": "university not found"
   }
   ```

4. **405 Method Not Allowed**
   ```json
   {
     "error": "method not allowed"
   }
   ```

## Тестирование

Для тестирования endpoints можно использовать:

1. **curl** (примеры выше)
2. **Postman** коллекцию
3. **Swagger UI** по адресу `http://localhost:8083/swagger/`

### Запуск тестов:

```bash
cd structure-service
go test -v ./internal/infrastructure/http/
```