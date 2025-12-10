# Примеры использования API чатов

## Базовые запросы

### Получить все чаты (первая страница)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all"
```

### Получить чаты с пагинацией
```bash
# Вторая страница по 25 записей
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?limit=25&offset=25"

# Третья страница по 10 записей  
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?limit=10&offset=20"
```

## Сортировка

### По названию (по умолчанию)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=name&sort_order=asc"
```

### По количеству участников (убывание)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=participants_count&sort_order=desc"
```

### По дате создания (новые первыми)
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=created_at&sort_order=desc"
```

### По университету
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=university&sort_order=asc"
```

### По источнику
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=source&sort_order=asc"
```

## Поиск

### Поиск по названию чата
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=студенты"
```

### Поиск по университету
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=МГУ"
```

### Поиск по подразделению
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=математика"
```

### Поиск по MAX Chat ID
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=123456789"
```

### Поиск по источнику
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=admin_panel"
```

## Комбинированные запросы

### Поиск с сортировкой
```bash
# Найти чаты МГУ и отсортировать по количеству участников
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=МГУ&sort_by=participants_count&sort_order=desc"
```

### Поиск с пагинацией
```bash
# Найти чаты со словом "студенты", первые 10 записей
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=студенты&limit=10&offset=0"
```

### Полный запрос
```bash
# Найти чаты математических факультетов, отсортировать по дате создания (новые первыми), взять первые 20
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=математика&sort_by=created_at&sort_order=desc&limit=20&offset=0"
```

## JavaScript примеры

### Базовый запрос
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

### Пагинация с состоянием
```javascript
class ChatPaginator {
  constructor(token) {
    this.token = token;
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
    
    const response = await fetch(`/chats/all?${params}`, {
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json'
      }
    });
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
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

### Класс для работы с API
```python
class ChatAPI:
    def __init__(self, base_url, token):
        self.base_url = base_url
        self.token = token
        self.headers = {
            'Authorization': f'Bearer {token}',
            'Content-Type': 'application/json'
        }
    
    def search_chats(self, search='', sort_by='name', sort_order='asc', page=0, limit=50):
        params = {
            'limit': limit,
            'offset': page * limit,
            'sort_by': sort_by,
            'sort_order': sort_order
        }
        
        if search:
            params['search'] = search
        
        response = requests.get(f"{self.base_url}/chats/all", params=params, headers=self.headers)
        response.raise_for_status()
        return response.json()
    
    def get_all_pages(self, search='', sort_by='name', sort_order='asc', limit=50):
        """Получить все чаты со всех страниц"""
        all_chats = []
        page = 0
        
        while True:
            result = self.search_chats(search, sort_by, sort_order, page, limit)
            all_chats.extend(result['data'])
            
            if len(result['data']) < limit or page >= result['total_pages'] - 1:
                break
            
            page += 1
        
        return all_chats
    
    def get_chat_statistics(self):
        """Получить статистику по чатам"""
        all_chats = self.get_all_pages()
        
        stats = {
            'total_chats': len(all_chats),
            'total_participants': sum(chat['participants_count'] for chat in all_chats),
            'by_source': {},
            'by_university': {},
            'avg_participants': 0
        }
        
        if all_chats:
            stats['avg_participants'] = stats['total_participants'] / len(all_chats)
            
            # Группировка по источникам
            for chat in all_chats:
                source = chat['source']
                stats['by_source'][source] = stats['by_source'].get(source, 0) + 1
            
            # Группировка по университетам
            for chat in all_chats:
                if chat['university']:
                    uni_name = chat['university']['name']
                    stats['by_university'][uni_name] = stats['by_university'].get(uni_name, 0) + 1
        
        return stats

# Использование
api = ChatAPI('http://localhost:8080', token)

# Получить все чаты МГУ, отсортированные по количеству участников
mgu_chats = api.search_chats(search='МГУ', sort_by='participants_count', sort_order='desc')

# Получить статистику
stats = api.get_chat_statistics()
print(f"Всего чатов: {stats['total_chats']}")
print(f"Всего участников: {stats['total_participants']}")
print(f"Среднее количество участников: {stats['avg_participants']:.1f}")
```

## Обработка ошибок

### JavaScript
```javascript
async function safeGetChats(token, search, sortBy, sortOrder, page, limit) {
  try {
    const params = new URLSearchParams({
      search: encodeURIComponent(search),
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
      if (response.status === 401) {
        throw new Error('Токен авторизации недействителен или истек');
      } else if (response.status === 403) {
        throw new Error('Недостаточно прав доступа');
      } else {
        throw new Error(`HTTP error! status: ${response.status}`);
      }
    }
    
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    console.error('Ошибка при получении чатов:', error);
    return { success: false, error: error.message };
  }
}
```

### Python
```python
def safe_get_chats(base_url, token, **kwargs):
    try:
        result = get_all_chats(base_url, token, **kwargs)
        return {'success': True, 'data': result}
    except requests.exceptions.HTTPError as e:
        if e.response.status_code == 401:
            return {'success': False, 'error': 'Токен авторизации недействителен или истек'}
        elif e.response.status_code == 403:
            return {'success': False, 'error': 'Недостаточно прав доступа'}
        else:
            return {'success': False, 'error': f'HTTP error: {e.response.status_code}'}
    except requests.exceptions.RequestException as e:
        return {'success': False, 'error': f'Network error: {str(e)}'}
    except Exception as e:
        return {'success': False, 'error': f'Unexpected error: {str(e)}'}
```

## Примеры ролевой фильтрации

### Суперадмин
```bash
# Суперадмин видит все чаты
curl -H "Authorization: Bearer SUPERADMIN_TOKEN" \
     -X GET "http://localhost:8080/chats/all"
```

### Куратор университета
```bash
# Куратор видит только чаты своего университета
curl -H "Authorization: Bearer CURATOR_TOKEN" \
     -X GET "http://localhost:8080/chats/all"
```

### Оператор
```bash
# Оператор видит только чаты своего университета
curl -H "Authorization: Bearer OPERATOR_TOKEN" \
     -X GET "http://localhost:8080/chats/all"
```

## Мониторинг и аналитика

### Получение топ чатов по участникам
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=participants_count&sort_order=desc&limit=10"
```

### Получение новых чатов
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?sort_by=created_at&sort_order=desc&limit=20"
```

### Анализ по источникам
```bash
# Чаты из админ-панели
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=admin_panel"

# Чаты от бот-регистратора
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=bot_registrar"

# Академические группы
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -X GET "http://localhost:8080/chats/all?search=academic_group"
```