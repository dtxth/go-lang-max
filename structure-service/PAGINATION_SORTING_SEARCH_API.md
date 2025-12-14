# API пагинации, сортировки и поиска университетов

## Обзор

Эндпоинт `/universities` теперь поддерживает расширенные возможности:
- **Пагинация** - разбиение результатов на страницы
- **Сортировка** - упорядочивание по любому полю в порядке возрастания или убывания
- **Поиск** - фильтрация по всем полям университета

## Эндпоинт

```
GET /universities
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

- `id` - ID университета
- `name` - Название университета
- `inn` - ИНН
- `kpp` - КПП
- `foiv` - ФОИВ (федеральный орган исполнительной власти)
- `created_at` - Дата создания
- `updated_at` - Дата обновления

## Формат ответа

```json
{
  "data": [
    {
      "id": 1,
      "name": "МГУ им. М.В. Ломоносова",
      "inn": "1234567890",
      "kpp": "123456789",
      "foiv": "Минобрнауки России",
      "created_at": "2023-12-10T09:00:00Z",
      "updated_at": "2023-12-10T09:00:00Z"
    }
  ],
  "total": 150,
  "limit": 50,
  "offset": 0,
  "total_pages": 3
}
```

### Поля ответа

- `data` - Массив университетов
- `total` - Общее количество записей (с учетом поиска)
- `limit` - Лимит записей на странице
- `offset` - Текущее смещение
- `total_pages` - Общее количество страниц

## Примеры использования

### Базовая пагинация

```bash
# Первая страница (50 записей)
GET /universities

# Вторая страница
GET /universities?limit=50&offset=50

# Третья страница с лимитом 20
GET /universities?limit=20&offset=40
```

### Сортировка

```bash
# Сортировка по названию (по умолчанию)
GET /universities?sort_by=name&sort_order=asc

# Сортировка по ИНН в убывающем порядке
GET /universities?sort_by=inn&sort_order=desc

# Сортировка по дате создания (новые первыми)
GET /universities?sort_by=created_at&sort_order=desc

# Сортировка по ФОИВ
GET /universities?sort_by=foiv&sort_order=asc
```

### Поиск

```bash
# Поиск по названию
GET /universities?search=МГУ

# Поиск по ИНН
GET /universities?search=1234567890

# Поиск по КПП
GET /universities?search=123456789

# Поиск по ФОИВ
GET /universities?search=Минобрнауки
```

### Комбинированные запросы

```bash
# Поиск с сортировкой
GET /universities?search=университет&sort_by=name&sort_order=asc

# Поиск с пагинацией
GET /universities?search=МГУ&limit=10&offset=0

# Полный запрос
GET /universities?search=Минобрнауки&sort_by=created_at&sort_order=desc&limit=25&offset=0
```

## JavaScript примеры

### Базовый запрос
```javascript
async function getAllUniversities(page = 0, limit = 50) {
  const response = await fetch(`/universities?limit=${limit}&offset=${page * limit}`);
  const data = await response.json();
  return data;
}
```

### С поиском и сортировкой
```javascript
async function searchUniversities(searchQuery, sortBy = 'name', sortOrder = 'asc', page = 0, limit = 50) {
  const params = new URLSearchParams({
    search: searchQuery,
    sort_by: sortBy,
    sort_order: sortOrder,
    limit: limit.toString(),
    offset: (page * limit).toString()
  });
  
  const response = await fetch(`/universities?${params}`);
  const data = await response.json();
  return data;
}

// Использование
const result = await searchUniversities('МГУ', 'name', 'asc', 0, 25);
console.log(`Найдено ${result.total} университетов`);
console.log(`Страница 1 из ${result.total_pages}`);
```

### Пагинация с состоянием
```javascript
class UniversityPaginator {
  constructor() {
    this.currentPage = 0;
    this.limit = 50;
    this.total = 0;
    this.totalPages = 0;
  }
  
  async loadPage(page, search = '', sortBy = 'name', sortOrder = 'asc') {
    const params = new URLSearchParams({
      limit: this.limit.toString(),
      offset: (page * this.limit).toString(),
      sort_by: sortBy,
      sort_order: sortOrder
    });
    
    if (search) {
      params.append('search', search);
    }
    
    const response = await fetch(`/universities?${params}`);
    const data = await response.json();
    
    this.currentPage = page;
    this.total = data.total;
    this.totalPages = data.total_pages;
    
    return data.data;
  }
  
  async nextPage(search, sortBy, sortOrder) {
    if (this.currentPage < this.totalPages - 1) {
      return await this.loadPage(this.currentPage + 1, search, sortBy, sortOrder);
    }
    return [];
  }
  
  async prevPage(search, sortBy, sortOrder) {
    if (this.currentPage > 0) {
      return await this.loadPage(this.currentPage - 1, search, sortBy, sortOrder);
    }
    return [];
  }
}
```

## Python примеры

### Базовый запрос
```python
import requests

def get_all_universities(base_url, page=0, limit=50, search='', sort_by='name', sort_order='asc'):
    params = {
        'limit': limit,
        'offset': page * limit,
        'sort_by': sort_by,
        'sort_order': sort_order
    }
    
    if search:
        params['search'] = search
    
    response = requests.get(f"{base_url}/universities", params=params)
    response.raise_for_status()
    return response.json()

# Использование
result = get_all_universities('http://localhost:8080', search='МГУ', sort_by='created_at', sort_order='desc')
print(f"Найдено {result['total']} университетов")
for university in result['data']:
    print(f"{university['name']} - ИНН: {university['inn']}")
```

### Класс для работы с API
```python
class UniversityAPI:
    def __init__(self, base_url):
        self.base_url = base_url
    
    def search_universities(self, search='', sort_by='name', sort_order='asc', page=0, limit=50):
        params = {
            'limit': limit,
            'offset': page * limit,
            'sort_by': sort_by,
            'sort_order': sort_order
        }
        
        if search:
            params['search'] = search
        
        response = requests.get(f"{self.base_url}/universities", params=params)
        response.raise_for_status()
        return response.json()
    
    def get_all_pages(self, search='', sort_by='name', sort_order='asc', limit=50):
        """Получить все университеты со всех страниц"""
        all_universities = []
        page = 0
        
        while True:
            result = self.search_universities(search, sort_by, sort_order, page, limit)
            all_universities.extend(result['data'])
            
            if len(result['data']) < limit or page >= result['total_pages'] - 1:
                break
            
            page += 1
        
        return all_universities
    
    def get_statistics(self):
        """Получить статистику по университетам"""
        all_universities = self.get_all_pages()
        
        stats = {
            'total_universities': len(all_universities),
            'by_foiv': {},
            'by_region': {},
            'creation_years': {}
        }
        
        for uni in all_universities:
            # Группировка по ФОИВ
            foiv = uni.get('foiv', 'Не указан')
            stats['by_foiv'][foiv] = stats['by_foiv'].get(foiv, 0) + 1
            
            # Группировка по годам создания
            if uni.get('created_at'):
                year = uni['created_at'][:4]
                stats['creation_years'][year] = stats['creation_years'].get(year, 0) + 1
        
        return stats

# Использование
api = UniversityAPI('http://localhost:8080')

# Получить все университеты Минобрнауки, отсортированные по названию
minobrnauki_unis = api.search_universities(search='Минобрнауки', sort_by='name', sort_order='asc')

# Получить статистику
stats = api.get_statistics()
print(f"Всего университетов: {stats['total_universities']}")
print(f"По ФОИВ: {stats['by_foiv']}")
```

## Особенности поиска

Поиск выполняется по следующим полям:
- Название университета
- ИНН
- КПП
- ФОИВ

Поиск **не чувствителен к регистру** и использует частичное совпадение (LIKE '%query%').

## Коды ответов

- `200 OK` - Успешный запрос
- `400 Bad Request` - Некорректные параметры запроса
- `500 Internal Server Error` - Внутренняя ошибка сервера

## Ограничения

- Максимальный лимит записей на странице: 100
- Минимальный лимит: 1 (по умолчанию 50)
- Минимальное смещение: 0
- Поиск работает только с непустыми строками

## Производительность

- Для больших объемов данных рекомендуется использовать пагинацию
- Поиск по индексированным полям (id, inn) выполняется быстрее
- Сортировка по полям с индексами также оптимизирована

## Интеграция с другими сервисами

Данный API может использоваться другими сервисами для:
- Получения списка университетов для выпадающих списков
- Поиска университетов по ИНН при создании сотрудников
- Аналитики и отчетности по университетам
- Синхронизации данных между сервисами