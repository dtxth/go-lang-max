# Excel Import - Маппинг колонок

## Структура входного Excel файла

### Колонки (по порядку)

1. **Нормализованный номер телефона администратора** - Телефон в формате 79884753064
2. **max_id** - ID пользователя в MAX (может быть пустым)
3. **ИНН_Справочник** - ИНН из справочника (105014177)
4. **ФОИВ_Справочник** - Название ФОИВ (Минобрнауки России)
5. **Наименование организации_Справочник** - Полное название организации
6. **Наименование головного подразделения/филиала** - Название филиала
7. **ИНН юридического лица** - ИНН (105014177)
8. **КПП головного подразделения/филиала** - КПП (10501001)
9. **Факультет/институт/иная структурная классификация** - Название факультета
10. **Курс обучения** - Номер курса (2)
11. **Номер группы** - Номер группы (Колледж ИП-22 (2024)
12. **Название чата** - Название чата (Колледж ИП-22 (2024 ОФО МГТУ)
13. **Мобильный номер телефона администратора чата** - Телефон (79884753064)
14. **Наименование файла** - Путь к файлу источнику
15. **chat_id** - ID чата в MAX (-69257108032233)
16. **link** - Ссылка на чат (https://max.ru/join/...)
17. **add_user** - Флаг (ИСТИНА/ЛОЖЬ)
18. **add_admin** - Флаг (ИСТИНА/ЛОЖЬ)

## Маппинг на структуру базы данных

### 1. University (Университет)

```
Источник → Таблица universities
─────────────────────────────────
Колонка 7 (ИНН юридического лица) → inn
Колонка 5 (Наименование организации) → name
Колонка 8 (КПП) → kpp
```

**Пример:**
```sql
INSERT INTO universities (name, inn, kpp)
VALUES (
  'ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ БЮДЖЕТНОЕ ОБРАЗОВАТЕЛЬНОЕ УЧРЕЖДЕНИЕ ВЫСШЕГО ОБРАЗОВАНИЯ "МАЙКОПСКИЙ ГОСУДАРСТВЕННЫЙ ТЕХНОЛОГИЧЕСКИЙ УНИВЕРСИТЕТ"',
  '105014177',
  '10501001'
);
```

### 2. Branch (Филиал)

```
Источник → Таблица branches
─────────────────────────────────
Колонка 6 (Наименование головного подразделения) → name
university_id (из шага 1) → university_id
```

**Пример:**
```sql
INSERT INTO branches (name, university_id)
VALUES (
  'Федеральное государственное бюджетное образовательное учреждение высшего образования «Майкопский государственный технологический университет»',
  1
);
```

### 3. Faculty (Факультет)

```
Источник → Таблица faculties
─────────────────────────────────
Колонка 9 (Факультет/институт) → name
branch_id (из шага 2) → branch_id
```

**Пример:**
```sql
INSERT INTO faculties (name, branch_id)
VALUES (
  'Политехнический колледж МГТУ',
  1
);
```

### 4. Group (Группа)

```
Источник → Таблица groups
─────────────────────────────────
Колонка 11 (Номер группы) → name
Колонка 10 (Курс обучения) → course
faculty_id (из шага 3) → faculty_id
```

**Пример:**
```sql
INSERT INTO groups (name, course, faculty_id)
VALUES (
  'Колледж ИП-22 (2024',
  2,
  1
);
```

### 5. Chat (Чат)

```
Источник → Таблица chats
─────────────────────────────────
Колонка 12 (Название чата) → name
Колонка 16 (link) → url
Колонка 15 (chat_id) → external_id (опционально)
university_id (из шага 1) → university_id
group_id (из шага 4) → group_id (опционально)
'academic_group' → source
```

**Пример:**
```sql
INSERT INTO chats (name, url, external_id, university_id, group_id, source)
VALUES (
  'Колледж ИП-22 (2024 ОФО МГТУ',
  'https://max.ru/join/fqQlVkO6LU-RAw5HkshlQy6giI9kJiFU_a0OoJ75TTQ',
  '-69257108032233',
  1,
  1,
  'academic_group'
);
```

### 6. Administrator (Администратор)

```
Источник → Таблица administrators
─────────────────────────────────
Колонка 1 или 13 (Телефон) → phone
Колонка 2 (max_id) → max_id (если есть)
chat_id (из шага 5) → chat_id
```

**Пример:**
```sql
INSERT INTO administrators (chat_id, phone, max_id)
VALUES (
  1,
  '+79884753064',
  '496728250'
);
```

## Логика обработки

### 1. Нормализация телефона

```go
// Из: 79884753064
// В: +79884753064
phone := normalizePhone(row[0]) // или row[12]
```

### 2. Обработка дубликатов

- **University**: Проверка по INN
- **Branch**: Проверка по name + university_id
- **Faculty**: Проверка по name + branch_id
- **Group**: Проверка по name + faculty_id
- **Chat**: Проверка по url или external_id
- **Administrator**: Проверка по chat_id + phone

### 3. Обработка пустых значений

```go
// Если max_id пустой - получить из MAX API
if row[1] == "" {
    maxID, err := maxClient.GetUserByPhone(phone)
    if err != nil {
        // Логировать ошибку, продолжить без max_id
    }
}

// Если КПП пустой - использовать NULL
kpp := sql.NullString{
    String: row[7],
    Valid:  row[7] != "",
}
```

### 4. Обработка флагов

```go
// Колонка 17: add_user (ИСТИНА/ЛОЖЬ)
addUser := strings.ToUpper(row[16]) == "ИСТИНА" || 
           strings.ToUpper(row[16]) == "TRUE"

// Колонка 18: add_admin (ИСТИНА/ЛОЖЬ)
addAdmin := strings.ToUpper(row[17]) == "ИСТИНА" || 
            strings.ToUpper(row[17]) == "TRUE"

// Если add_admin = ИСТИНА, добавить администратора
if addAdmin {
    // Создать administrator
}
```

## Пример полной обработки строки

```go
type ExcelRow struct {
    Phone1              string // 0
    MaxID               string // 1
    INNReference        string // 2
    FOIVReference       string // 3
    OrgNameReference    string // 4
    BranchName          string // 5
    INN                 string // 6
    KPP                 string // 7
    FacultyName         string // 8
    Course              string // 9
    GroupNumber         string // 10
    ChatName            string // 11
    Phone2              string // 12
    FileName            string // 13
    ChatID              string // 14
    Link                string // 15
    AddUser             string // 16
    AddAdmin            string // 17
}

func ProcessRow(row ExcelRow) error {
    // 1. Создать/найти университет
    university, err := findOrCreateUniversity(row.INN, row.OrgNameReference, row.KPP)
    
    // 2. Создать/найти филиал
    branch, err := findOrCreateBranch(row.BranchName, university.ID)
    
    // 3. Создать/найти факультет
    faculty, err := findOrCreateFaculty(row.FacultyName, branch.ID)
    
    // 4. Создать/найти группу
    course, _ := strconv.Atoi(row.Course)
    group, err := findOrCreateGroup(row.GroupNumber, course, faculty.ID)
    
    // 5. Создать/найти чат
    chat, err := findOrCreateChat(row.ChatName, row.Link, row.ChatID, university.ID, group.ID)
    
    // 6. Добавить администратора (если add_admin = ИСТИНА)
    if strings.ToUpper(row.AddAdmin) == "ИСТИНА" {
        phone := normalizePhone(row.Phone1) // или row.Phone2
        maxID := row.MaxID
        
        // Если max_id пустой, получить из MAX API
        if maxID == "" {
            maxID, _ = getMaxIDByPhone(phone)
        }
        
        err := addAdministrator(chat.ID, phone, maxID)
    }
    
    return nil
}
```

## Валидация данных

### Обязательные поля

- ✅ INN (колонка 6)
- ✅ Название организации (колонка 5)
- ✅ Название чата (колонка 12)
- ✅ Ссылка на чат (колонка 16)
- ✅ Телефон администратора (колонка 1 или 13)

### Опциональные поля

- КПП (колонка 8)
- Название филиала (колонка 6)
- Название факультета (колонка 9)
- Курс (колонка 10)
- Номер группы (колонка 11)
- max_id (колонка 2)
- chat_id (колонка 15)

### Правила валидации

```go
// INN: 10 или 12 цифр
if !regexp.MustCompile(`^\d{10}$|^\d{12}$`).MatchString(inn) {
    return errors.New("invalid INN format")
}

// КПП: 9 цифр
if kpp != "" && !regexp.MustCompile(`^\d{9}$`).MatchString(kpp) {
    return errors.New("invalid KPP format")
}

// Телефон: начинается с 7, 11 цифр
if !regexp.MustCompile(`^7\d{10}$`).MatchString(phone) {
    return errors.New("invalid phone format")
}

// URL: начинается с https://max.ru/
if !strings.HasPrefix(url, "https://max.ru/") {
    return errors.New("invalid MAX URL")
}

// Курс: число от 1 до 6
course, err := strconv.Atoi(courseStr)
if err != nil || course < 1 || course > 6 {
    return errors.New("invalid course number")
}
```

## Обработка ошибок

### Стратегия

1. **Критические ошибки** (остановить обработку строки):
   - Невалидный INN
   - Отсутствует название чата
   - Отсутствует URL чата

2. **Некритические ошибки** (логировать, продолжить):
   - Не удалось получить max_id
   - Пустой КПП
   - Пустое название факультета

3. **Логирование**:
   ```go
   if err != nil {
       logError(MigrationError{
           JobID:            jobID,
           RecordIdentifier: fmt.Sprintf("row_%d", rowNumber),
           ErrorMessage:     err.Error(),
       })
   }
   ```

## Производительность

### Оптимизации

1. **Batch операции**:
   ```go
   // Вместо INSERT для каждой строки
   // Использовать batch insert каждые 100 строк
   ```

2. **Кэширование**:
   ```go
   // Кэшировать найденные университеты, филиалы и т.д.
   universityCache := make(map[string]*University)
   ```

3. **Транзакции**:
   ```go
   // Обрабатывать по 1000 строк в одной транзакции
   tx.Begin()
   // ... process 1000 rows
   tx.Commit()
   ```

## Пример использования

```bash
# Загрузить Excel файл через API
curl -X POST http://localhost:8084/migration/excel \
  -F "file=@Наименование чатов групп МГТУ в МАХ 17.11.25_ИТОГ.xlsx"

# Проверить статус
curl http://localhost:8084/migration/jobs/1

# Результат:
{
  "id": 1,
  "source_type": "excel",
  "status": "completed",
  "total": 155000,
  "processed": 155000,
  "failed": 123
}
```

## Документация

- [Migration Service Implementation](./migration-service/MIGRATION_SERVICE_IMPLEMENTATION.md)
- [Excel Import Implementation](./structure-service/EXCEL_IMPORT_IMPLEMENTATION.md)
- [API Reference](./docs/API_REFERENCE.md)
