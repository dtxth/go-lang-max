# Примеры использования API сотрудников

## Базовые запросы

### Получить всех сотрудников (первая страница)
```bash
curl -X GET "http://localhost:8080/employees/all"
```

### Получить сотрудников с пагинацией
```bash
# Вторая страница по 25 записей
curl -X GET "http://localhost:8080/employees/all?limit=25&offset=25"

# Третья страница по 10 записей  
curl -X GET "http://localhost:8080/employees/all?limit=10&offset=20"
```

## Сортировка

### По фамилии (по умолчанию)
```bash
curl -X GET "http://localhost:8080/employees/all?sort_by=last_name&sort_order=asc"
```

### По имени в убывающем порядке
```bash
curl -X GET "http://localhost:8080/employees/all?sort_by=first_name&sort_order=desc"
```

### По дате создания (новые первыми)
```bash
curl -X GET "http://localhost:8080/employees/all?sort_by=created_at&sort_order=desc"
```

### По университету
```bash
curl -X GET "http://localhost:8080/employees/all?sort_by=university&sort_order=asc"
```

### По роли
```bash
curl -X GET "http://localhost:8080/employees/all?sort_by=role&sort_order=asc"
```

## Поиск

### Поиск по имени
```bash
curl -X GET "http://localhost:8080/employees/all?search=Иван"
```

### Поиск по фамилии
```bash
curl -X GET "http://localhost:8080/employees/all?search=Петров"
```

### Поиск по телефону
```bash
curl -X GET "http://localhost:8080/employees/all?search=%2B7900"
# Примечание: + нужно кодировать как %2B в URL
```

### Поиск по университету
```bash
curl -X GET "http://localhost:8080/employees/all?search=МГУ"
```

### Поиск по ИНН
```bash
curl -X GET "http://localhost:8080/employees/all?search=1234567890"
```

### Поиск по роли
```bash
curl -X GET "http://localhost:8080/employees/all?search=curator"
```

## Комбинированные запросы

### Поиск с сортировкой
```bash
# Найти всех Ивановых и отсортировать по имени
curl -X GET "http://localhost:8080/employees/all?search=Иванов&sort_by=first_name&sort_order=asc"
```

### Поиск с пагинацией
```bash
# Найти сотрудников МГУ, первые 10 записей
curl -X GET "http://localhost:8080/employees/all?search=МГУ&limit=10&offset=0"
```

### Полный запрос
```bash
# Найти кураторов, отсортировать по дате создания (новые первыми), взять первые 20
curl -X GET "http://localhost:8080/employees/all?search=curator&sort_by=created_at&sort_order=desc&limit=20&offset=0"
```

## JavaScript примеры

### Базовый запрос
```javascript
async function getAllEmployees(page = 0, limit = 50) {
  const response = await fetch(`/employees/all?limit=${limit}&offset=${page * limit}`);
  const data = await response.json();
  return data;
}
```

### С поиском и сортировкой
```javascript
async function searchEmployees(searchQuery, sortBy = 'last_name', sortOrder = 'asc', page = 0, limit = 50) {
  const params = new URLSearchParams({
    search: searchQuery,
    sort_by: sortBy,
    sort_order: sortOrder,
    limit: limit.toString(),
    offset: (page * limit).toString()
  });
  
  const response = await fetch(`/employees/all?${params}`);
  const data = await response.json();
  return data;
}

// Использование
const result = await searchEmployees('Иван', 'first_name', 'desc', 0, 25);
console.log(`Найдено ${result.total} сотрудников`);
console.log(`Страница 1 из ${result.total_pages}`);
```

### Пагинация с состоянием
```javascript
class EmployeePaginator {
  constructor() {
    this.currentPage = 0;
    this.limit = 50;
    this.total = 0;
    this.totalPages = 0;
  }
  
  async loadPage(page, search = '', sortBy = 'last_name', sortOrder = 'asc') {
    const params = new URLSearchParams({
      limit: this.limit.toString(),
      offset: (page * this.limit).toString(),
      sort_by: sortBy,
      sort_order: sortOrder
    });
    
    if (search) {
      params.append('search', search);
    }
    
    const response = await fetch(`/employees/all?${params}`);
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

def get_all_employees(base_url, page=0, limit=50, search='', sort_by='last_name', sort_order='asc'):
    params = {
        'limit': limit,
        'offset': page * limit,
        'sort_by': sort_by,
        'sort_order': sort_order
    }
    
    if search:
        params['search'] = search
    
    response = requests.get(f"{base_url}/employees/all", params=params)
    response.raise_for_status()
    return response.json()

# Использование
result = get_all_employees('http://localhost:8080', search='МГУ', sort_by='created_at', sort_order='desc')
print(f"Найдено {result['total']} сотрудников")
for employee in result['data']:
    print(f"{employee['last_name']} {employee['first_name']} - {employee['university']['name']}")
```

### Класс для работы с API
```python
class EmployeeAPI:
    def __init__(self, base_url):
        self.base_url = base_url
    
    def search_employees(self, search='', sort_by='last_name', sort_order='asc', page=0, limit=50):
        params = {
            'limit': limit,
            'offset': page * limit,
            'sort_by': sort_by,
            'sort_order': sort_order
        }
        
        if search:
            params['search'] = search
        
        response = requests.get(f"{self.base_url}/employees/all", params=params)
        response.raise_for_status()
        return response.json()
    
    def get_all_pages(self, search='', sort_by='last_name', sort_order='asc', limit=50):
        """Получить всех сотрудников со всех страниц"""
        all_employees = []
        page = 0
        
        while True:
            result = self.search_employees(search, sort_by, sort_order, page, limit)
            all_employees.extend(result['data'])
            
            if len(result['data']) < limit or page >= result['total_pages'] - 1:
                break
            
            page += 1
        
        return all_employees

# Использование
api = EmployeeAPI('http://localhost:8080')
all_curators = api.get_all_pages(search='curator', sort_by='university')
```

## Обработка ошибок

### JavaScript
```javascript
async function safeGetEmployees(search, sortBy, sortOrder, page, limit) {
  try {
    const response = await fetch(`/employees/all?search=${encodeURIComponent(search)}&sort_by=${sortBy}&sort_order=${sortOrder}&limit=${limit}&offset=${page * limit}`);
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    console.error('Ошибка при получении сотрудников:', error);
    return { success: false, error: error.message };
  }
}
```

### Python
```python
def safe_get_employees(base_url, **kwargs):
    try:
        result = get_all_employees(base_url, **kwargs)
        return {'success': True, 'data': result}
    except requests.exceptions.RequestException as e:
        return {'success': False, 'error': str(e)}
    except Exception as e:
        return {'success': False, 'error': f'Unexpected error: {str(e)}'}
```