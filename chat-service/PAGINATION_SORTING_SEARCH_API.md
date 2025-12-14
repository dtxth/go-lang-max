# API пагинации, сортировки и поиска чатов

## Обзор

Эндпоинт `/chats/all` теперь поддерживает расширенные возможности:
- **Пагинация** - разбиение результатов на страницы
- **Сортировка** - упорядочивание по любому полю в порядке возрастания или убывания
- **Поиск** - фильтрация по всем полям чата
- **Ролевая фильтрация** - автоматическая фильтрация по роли пользователя

## Эндпоинт

```
GET /chats/all
```

## Параметры запроса

| Параметр | Тип | Обязательный | По умолчанию | Описание |
|----------|-----|--------------|--------------|----------|
| `limit` | int | Нет | 50 | Количество записей на странице (максимум 100) |
| `offset` | int | Нет | 0 | Смещение от начала результатов |
| `sort_by` | string | Нет | name | Поле для сортировки |
| `sort_order` | string | Нет | asc | Порядок сортировки (asc/desc) |
| `search` | string | Нет | - | Поисковый запрос |

### Доступные поля для сортировки

- `id` - ID чата
- `name` - Название чата
- `url` - URL чата
- `max_chat_id` - MAX Chat ID
- `participants_count` - Количество участников
- `department` - Подразделение
- `source` - Источник чата
- `university` - Название университета
- `created_at` - Дата создания
- `updated_at` - Дата обновления

## Авторизация

Эндпоинт требует авторизации через Bearer token в заголовке `Authorization`.

### Ролевая фильтрация

- **Суперадмин** - видит все чаты
- **Куратор** - видит только чаты своего университета
- **Оператор** - видит только чаты своего университета

## Формат ответа

```json
{
  "data": [
    {
      "id": 1,
      "name": "Чат студентов МГУ",
      "url": "https://t.me/mgu_students",
      "max_chat_id": "123456789",
      "external_chat_id": "ext_123",
      "participants_count": 150,
      "university_id": 1,
      "university": {
        "id": 1,
        "name": "МГУ им. М.В. Ломоносова",
        "inn": "1234567890",
        "kpp": "123456789",
        "created_at": "2023-12-10T09:00:00Z",
        "updated_at": "2023-12-10T09:00:00Z"
      },
      "department": "Факультет математики",
      "source": "admin_panel",
      "administrators": [
        {
          "id": 1,
          "chat_id": 1,
          "phone": "+79001234567",
          "max_id": "admin_123",
          "add_user": true,
          "add_admin": true,
          "created_at": "2023-12-10T09:30:00Z",
          "updated_at": "2023-12-10T09:30:00Z"
        }
      ],
      "created_at": "2023-12-10T09:30:00Z",
      "updated_at": "2023-12-10T10:00:00Z"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0,
  "total_pages": 3
}
```

### Поля ответа

- `data` - Массив чатов
- `total` - Общее количество записей (с учетом поиска и ролевой фильтрации)
- `limit` - Лимит записей на странице
- `offset` - Текущее смещение
- `total_pages` - Общее количество страниц

## Примеры использования

### Базовая пагинация

```bash
# Первая страница (50 записей)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all"

# Вторая страница
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?limit=50&offset=50"

# Третья страница с лимитом 20
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?limit=20&offset=40"
```

### Сортировка

```bash
# Сортировка по названию (по умолчанию)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=name&sort_order=asc"

# Сортировка по количеству участников (убывание)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=participants_count&sort_order=desc"

# Сортировка по дате создания (новые первыми)
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=created_at&sort_order=desc"

# Сортировка по университету
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=university&sort_order=asc"
```

### Поиск

```bash
# Поиск по названию чата
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=студенты"

# Поиск по университету
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=МГУ"

# Поиск по подразделению
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=математика"

# Поиск по MAX Chat ID
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=123456"

# Поиск по источнику
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=admin_panel"
```

### Комбинированные запросы

```bash
# Поиск с сортировкой
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=МГУ&sort_by=participants_count&sort_order=desc"

# Поиск с пагинацией
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=студенты&limit=10&offset=0"

# Полный запрос
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=математика&sort_by=created_at&sort_order=desc&limit=25&offset=0"
```

## JavaScript примеры

### Базовый запрос с авторизацией
```javascript
async function getAllChats(token, page = 0, limit = 50) {
  const response = await fetch(`/chats/all?limit=${limit}&offset=${page * limit}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const data = await response.json();
  return data;
}
```

### С поиском и сортировкой
```javascript
async function searchChats(token, searchQuery, sortBy = 'name', sortOrder = 'asc', page = 0, limit = 50) {
  const params = new URLSearchParams({
    search: searchQuery,
    sort_by: sortBy,
    sort_order: sortOrder,
    limit: limit.toString(),
    offset: (page * limit).toString()
  });
  
  const response = await fetch(`/chats/all?${params}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    }
  });
  
  if (!response.ok) {
    throw new Error(`HTTP error! status: ${response.status}`);
  }
  
  const data = await response.json();
  return data;
}

// Использование
const result = await searchChats(token, 'МГУ', 'participants_count', 'desc', 0, 25);
console.log(`Найдено ${result.total} чатов`);
console.log(`Страница 1 из ${result.total_pages}`);
```

## Python примеры

### Базовый запрос
```python
import requests

def get_all_chats(base_url, token, page=0, limit=50, search='', sort_by='name', sort_order='asc'):
    params = {
        'limit': limit,
        'offset': page * limit,
        'sort_by': sort_by,
        'sort_order': sort_order
    }
    
    if search:
        params['search'] = search
    
    headers = {
        'Authorization': f'Bearer {token}',
        'Content-Type': 'application/json'
    }
    
    response = requests.get(f"{base_url}/chats/all", params=params, headers=headers)
    response.raise_for_status()
    return response.json()

# Использование
result = get_all_chats('http://localhost:8080', token, search='МГУ', sort_by='created_at', sort_order='desc')
print(f"Найдено {result['total']} чатов")
for chat in result['data']:
    print(f"{chat['name']} - {chat['participants_count']} участников")
```

## Особенности поиска

Поиск выполняется по следующим полям:
- Название чата
- Подразделение (department)
- Название университета
- MAX Chat ID
- Источник (source)

Поиск **не чувствителен к регистру** и использует частичное совпадение (ILIKE '%query%').

Поддерживается поиск по нескольким словам - каждое слово должно встречаться в любом из полей.

## Коды ответов

- `200 OK` - Успешный запрос
- `400 Bad Request` - Некорректные параметры запроса
- `401 Unauthorized` - Отсутствует или недействительный токен авторизации
- `403 Forbidden` - Недостаточно прав доступа
- `500 Internal Server Error` - Внутренняя ошибка сервера

## Ограничения

- Максимальный лимит записей на странице: 100
- Минимальный лимит: 1 (по умолчанию 50)
- Минимальное смещение: 0
- Поиск работает только с непустыми строками
- Требуется валидный JWT токен для авторизации

## Производительность

- Для больших объемов данных рекомендуется использовать пагинацию
- Поиск по индексированным полям (id, max_chat_id) выполняется быстрее
- Сортировка по полям с индексами также оптимизирована
- Ролевая фильтрация применяется на уровне базы данных для оптимальной производительности