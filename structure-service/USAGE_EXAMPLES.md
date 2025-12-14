# Примеры использования API университетов

## Базовые запросы

### Получить все университеты (первая страница)
```bash
curl -X GET "http://localhost:8080/universities"
```

### Получить университеты с пагинацией
```bash
# Вторая страница по 25 записей
curl -X GET "http://localhost:8080/universities?limit=25&offset=25"

# Третья страница по 10 записей  
curl -X GET "http://localhost:8080/universities?limit=10&offset=20"
```

## Сортировка

### По названию (по умолчанию)
```bash
curl -X GET "http://localhost:8080/universities?sort_by=name&sort_order=asc"
```

### По ИНН в убывающем порядке
```bash
curl -X GET "http://localhost:8080/universities?sort_by=inn&sort_order=desc"
```

### По дате создания (новые первыми)
```bash
curl -X GET "http://localhost:8080/universities?sort_by=created_at&sort_order=desc"
```

### По ФОИВ
```bash
curl -X GET "http://localhost:8080/universities?sort_by=foiv&sort_order=asc"
```

### По КПП
```bash
curl -X GET "http://localhost:8080/universities?sort_by=kpp&sort_order=asc"
```

## Поиск

### Поиск по названию университета
```bash
curl -X GET "http://localhost:8080/universities?search=МГУ"
```

### Поиск по ИНН
```bash
curl -X GET "http://localhost:8080/universities?search=1234567890"
```

### Поиск по КПП
```bash
curl -X GET "http://localhost:8080/universities?search=123456789"
```

### Поиск по ФОИВ
```bash
curl -X GET "http://localhost:8080/universities?search=Минобрнауки"
```

### Поиск по части названия
```bash
curl -X GET "http://localhost:8080/universities?search=университет"
```

## Комбинированные запросы

### Поиск с сортировкой
```bash
# Найти все университеты с "МГУ" в названии и отсортировать по названию
curl -X GET "http://localhost:8080/universities?search=МГУ&sort_by=name&sort_order=asc"
```

### Поиск с пагинацией
```bash
# Найти университеты Минобрнауки, первые 10 записей
curl -X GET "http://localhost:8080/universities?search=Минобрнауки&limit=10&offset=0"
```

### Полный запрос
```bash
# Найти университеты Минобрнауки, отсортировать по дате создания (новые первыми), взять первые 20
curl -X GET "http://localhost:8080/universities?search=Минобрнауки&sort_by=created_at&sort_order=desc&limit=20&offset=0"
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

### Автокомплит для поиска университетов
```javascript
class UniversityAutocomplete {
  constructor(inputElement, resultsElement) {
    this.input = inputElement;
    this.results = resultsElement;
    this.debounceTimer = null;
    
    this.input.addEventListener('input', (e) => {
      this.handleInput(e.target.value);
    });
  }
  
  handleInput(query) {
    clearTimeout(this.debounceTimer);
    
    if (query.length < 2) {
      this.clearResults();
      return;
    }
    
    this.debounceTimer = setTimeout(() => {
      this.search(query);
    }, 300);
  }
  
  async search(query) {
    try {
      const result = await searchUniversities(query, 'name', 'asc', 0, 10);
      this.displayResults(result.data);
    } catch (error) {
      console.error('Ошибка поиска:', error);
    }
  }
  
  displayResults(universities) {
    this.results.innerHTML = '';
    
    universities.forEach(uni => {
      const item = document.createElement('div');
      item.className = 'autocomplete-item';
      item.innerHTML = `
        <div class="uni-name">${uni.name}</div>
        <div class="uni-details">ИНН: ${uni.inn} | ${uni.foiv}</div>
      `;
      item.addEventListener('click', () => {
        this.selectUniversity(uni);
      });
      this.results.appendChild(item);
    });
  }
  
  selectUniversity(university) {
    this.input.value = university.name;
    this.input.dataset.universityId = university.id;
    this.clearResults();
  }
  
  clearResults() {
    this.results.innerHTML = '';
  }
}

// Использование
const autocomplete = new UniversityAutocomplete(
  document.getElementById('university-search'),
  document.getElementById('search-results')
);
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
    
    def find_by_inn(self, inn):
        """Найти университет по ИНН"""
        result = self.search_universities(search=inn, limit=1)
        return result['data'][0] if result['data'] else None
    
    def get_statistics(self):
        """Получить статистику по университетам"""
        all_universities = self.get_all_pages()
        
        stats = {
            'total_universities': len(all_universities),
            'by_foiv': {},
            'creation_years': {},
            'with_kpp': 0,
            'without_kpp': 0
        }
        
        for uni in all_universities:
            # Группировка по ФОИВ
            foiv = uni.get('foiv', 'Не указан')
            stats['by_foiv'][foiv] = stats['by_foiv'].get(foiv, 0) + 1
            
            # Группировка по годам создания
            if uni.get('created_at'):
                year = uni['created_at'][:4]
                stats['creation_years'][year] = stats['creation_years'].get(year, 0) + 1
            
            # Статистика по КПП
            if uni.get('kpp'):
                stats['with_kpp'] += 1
            else:
                stats['without_kpp'] += 1
        
        return stats
    
    def export_to_csv(self, filename='universities.csv', search=''):
        """Экспорт университетов в CSV"""
        import csv
        
        universities = self.get_all_pages(search=search)
        
        with open(filename, 'w', newline='', encoding='utf-8') as csvfile:
            fieldnames = ['id', 'name', 'inn', 'kpp', 'foiv', 'created_at', 'updated_at']
            writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
            
            writer.writeheader()
            for uni in universities:
                writer.writerow(uni)
        
        print(f"Экспортировано {len(universities)} университетов в {filename}")

# Использование
api = UniversityAPI('http://localhost:8080')

# Получить все университеты Минобрнауки, отсортированные по названию
minobrnauki_unis = api.search_universities(search='Минобрнауки', sort_by='name', sort_order='asc')

# Найти университет по ИНН
mgu = api.find_by_inn('1234567890')
if mgu:
    print(f"Найден: {mgu['name']}")

# Получить статистику
stats = api.get_statistics()
print(f"Всего университетов: {stats['total_universities']}")
print(f"По ФОИВ: {stats['by_foiv']}")

# Экспорт в CSV
api.export_to_csv('all_universities.csv')
```

## Обработка ошибок

### JavaScript
```javascript
async function safeGetUniversities(search, sortBy, sortOrder, page, limit) {
  try {
    const params = new URLSearchParams({
      search: encodeURIComponent(search),
      sort_by: sortBy,
      sort_order: sortOrder,
      limit: limit.toString(),
      offset: (page * limit).toString()
    });
    
    const response = await fetch(`/universities?${params}`);
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }
    
    const data = await response.json();
    return { success: true, data };
  } catch (error) {
    console.error('Ошибка при получении университетов:', error);
    return { success: false, error: error.message };
  }
}
```

### Python
```python
def safe_get_universities(base_url, **kwargs):
    try:
        result = get_all_universities(base_url, **kwargs)
        return {'success': True, 'data': result}
    except requests.exceptions.RequestException as e:
        return {'success': False, 'error': str(e)}
    except Exception as e:
        return {'success': False, 'error': f'Unexpected error: {str(e)}'}
```

## Интеграция с формами

### HTML форма поиска
```html
<form id="university-search-form">
  <div class="search-controls">
    <input type="text" id="search-input" placeholder="Поиск университетов...">
    
    <select id="sort-field">
      <option value="name">По названию</option>
      <option value="inn">По ИНН</option>
      <option value="foiv">По ФОИВ</option>
      <option value="created_at">По дате создания</option>
    </select>
    
    <select id="sort-order">
      <option value="asc">По возрастанию</option>
      <option value="desc">По убыванию</option>
    </select>
    
    <button type="submit">Поиск</button>
  </div>
</form>

<div id="results"></div>
<div id="pagination"></div>
```

### JavaScript для формы
```javascript
document.getElementById('university-search-form').addEventListener('submit', async (e) => {
  e.preventDefault();
  
  const search = document.getElementById('search-input').value;
  const sortBy = document.getElementById('sort-field').value;
  const sortOrder = document.getElementById('sort-order').value;
  
  const result = await searchUniversities(search, sortBy, sortOrder, 0, 20);
  
  displayResults(result.data);
  displayPagination(result);
});

function displayResults(universities) {
  const resultsDiv = document.getElementById('results');
  resultsDiv.innerHTML = universities.map(uni => `
    <div class="university-card">
      <h3>${uni.name}</h3>
      <p>ИНН: ${uni.inn} | КПП: ${uni.kpp || 'Не указан'}</p>
      <p>ФОИВ: ${uni.foiv}</p>
      <p>Создан: ${new Date(uni.created_at).toLocaleDateString()}</p>
    </div>
  `).join('');
}
```

## Мониторинг и аналитика

### Получение топ университетов по дате создания
```bash
curl -X GET "http://localhost:8080/universities?sort_by=created_at&sort_order=desc&limit=10"
```

### Анализ по ФОИВ
```bash
# Университеты Минобрнауки
curl -X GET "http://localhost:8080/universities?search=Минобрнауки"

# Университеты Минздрава
curl -X GET "http://localhost:8080/universities?search=Минздрав"
```

### Поиск университетов без КПП
```python
# В Python можно получить все университеты и отфильтровать
api = UniversityAPI('http://localhost:8080')
all_unis = api.get_all_pages()
without_kpp = [uni for uni in all_unis if not uni.get('kpp')]
print(f"Университетов без КПП: {len(without_kpp)}")
```