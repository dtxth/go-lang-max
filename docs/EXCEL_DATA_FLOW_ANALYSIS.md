# Анализ потока данных из Excel в БД

## 📊 Структура входных данных Excel

```
Колонка 0:  Нормализованный номер телефона администратора (79884753064)
Колонка 1:  max_id (496728250)
Колонка 2:  ИНН_Справочник (105014177)
Колонка 3:  ФОИВ_Справочник (Минобрнауки России)
Колонка 4:  Наименование организации_Справочник (ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ...)
Колонка 5:  Наименование головного подразделения/филиала (Федеральное государственное...)
Колонка 6:  ИНН юридического лица (105014177)
Колонка 7:  КПП головного подразделения/филиала (10501001)
Колонка 8:  Факультет/институт/иная структурная классификация (Политехнический колледж МГТУ)
Колонка 9:  Курс обучения (2)
Колонка 10: Номер группы (Колледж ИП-22 (2024)
Колонка 11: Название чата (Колледж ИП-22 (2024 ОФО МГТУ)
Колонка 12: Мобильный номер телефона администратора чата (79884753064)
Колонка 13: Наименование файла (Министерство науки и высшего образования...)
Колонка 14: chat_id (-69257108032233)
Колонка 15: link (https://max.ru/join/fqQlVkO6LU-RAw5HkshlQy6giI9kJiFU_a0OoJ75TTQ)
Колонка 16: add_user (ИСТИНА)
Колонка 17: add_admin (ИСТИНА)
```

## ⚠️ ПРОБЛЕМА: Текущая реализация НЕ готова!

### Что не так:

1. **Неправильный маппинг колонок** - текущий код ожидает другую структуру
2. **Отсутствуют поля в БД** - `chat_id`, `add_user`, `add_admin`
3. **Неправильная обработка** - не учитывает max_id и флаги

## ✅ Куда попадут данные

### 1. Structure Service → PostgreSQL (structure_db)

**Таблица: `universities`**
```sql
INSERT INTO universities (name, inn, kpp, foiv)
VALUES (
  'ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ БЮДЖЕТНОЕ ОБРАЗОВАТЕЛЬНОЕ УЧРЕЖДЕНИЕ...',  -- Колонка 4
  '105014177',                                                              -- Колонка 6
  '10501001',                                                               -- Колонка 7
  'Минобрнауки России'                                                      -- Колонка 3
);
```

**Таблица: `branches`**
```sql
INSERT INTO branches (university_id, name)
VALUES (
  1,  -- ID созданного университета
  'Федеральное государственное бюджетное образовательное учреждение...'  -- Колонка 5
);
```

**Таблица: `faculties`**
```sql
INSERT INTO faculties (branch_id, name)
VALUES (
  1,  -- ID созданного филиала
  'Политехнический колледж МГТУ'  -- Колонка 8
);
```

**Таблица: `groups`**
```sql
INSERT INTO groups (faculty_id, course, number, chat_id)
VALUES (
  1,                           -- ID созданного факультета
  2,                           -- Колонка 9
  'Колледж ИП-22 (2024',       -- Колонка 10
  NULL                         -- Будет обновлено после создания чата
);
```

### 2. Chat Service → PostgreSQL (chat_db)

**Таблица: `chats`**
```sql
INSERT INTO chats (name, url, max_chat_id, university_id, source)
VALUES (
  'Колледж ИП-22 (2024 ОФО МГТУ',                                          -- Колонка 11
  'https://max.ru/join/fqQlVkO6LU-RAw5HkshlQy6giI9kJiFU_a0OoJ75TTQ',      -- Колонка 15
  '-69257108032233',                                                        -- Колонка 14 ⚠️ НУЖНО ДОБАВИТЬ
  1,                                                                        -- ID университета
  'academic_group'                                                          -- Константа
);
```

**Таблица: `administrators`**
```sql
INSERT INTO administrators (chat_id, phone, max_id)
VALUES (
  1,                    -- ID созданного чата
  '+79884753064',       -- Колонка 0 или 12 (нормализованный)
  '496728250'           -- Колонка 1
);
```

### 3. Migration Service → PostgreSQL (migration_db)

**Таблица: `migration_jobs`**
```sql
INSERT INTO migration_jobs (source_type, source_identifier, status, total, processed, failed)
VALUES (
  'excel',
  'Наименование чатов групп МГТУ в МАХ 17.11.25_ИТОГ.xlsx',
  'running',
  1,
  0,
  0
);
```

**Таблица: `migration_errors`** (если есть ошибки)
```sql
INSERT INTO migration_errors (job_id, record_identifier, error_message)
VALUES (
  1,
  'row_2',
  'Failed to create chat: ...'
);
```

## 🔧 Необходимые изменения

### 1. Обновить схему БД chat-service

```sql
-- chat-service/migrations/002_add_missing_fields.sql
ALTER TABLE chats ADD COLUMN IF NOT EXISTS external_chat_id TEXT;
CREATE INDEX IF NOT EXISTS idx_chats_external_chat_id ON chats(external_chat_id);

ALTER TABLE administrators ADD COLUMN IF NOT EXISTS add_user BOOLEAN DEFAULT TRUE;
ALTER TABLE administrators ADD COLUMN IF NOT EXISTS add_admin BOOLEAN DEFAULT TRUE;
```

### 2. Обновить Go структуры

**chat-service/internal/domain/chat.go:**
```go
type Chat struct {
    ID             int64       `json:"id"`
    Name           string      `json:"name"`
    URL            string      `json:"url"`
    MaxChatID      string      `json:"max_chat_id"`
    ExternalChatID *string     `json:"external_chat_id,omitempty"` // ⬅️ НОВОЕ (Колонка 14)
    // ... остальные поля
}
```

**chat-service/internal/domain/administrator.go:**
```go
type Administrator struct {
    ID       int64   `json:"id"`
    ChatID   int64   `json:"chat_id"`
    Phone    string  `json:"phone"`
    MaxID    string  `json:"max_id"`
    AddUser  bool    `json:"add_user"`   // ⬅️ НОВОЕ (Колонка 16)
    AddAdmin bool    `json:"add_admin"`  // ⬅️ НОВОЕ (Колонка 17)
    // ... остальные поля
}
```

### 3. Обновить ExcelRow структуру

**migration-service/internal/usecase/migrate_from_excel.go:**
```go
type ExcelRow struct {
    RowNumber        int
    AdminPhone1      string // Колонка 0 - Нормализованный номер телефона
    MaxID            string // Колонка 1 - max_id
    INNReference     string // Колонка 2 - ИНН_Справочник
    FOIVReference    string // Колонка 3 - ФОИВ_Справочник
    OrgNameRef       string // Колонка 4 - Наименование организации_Справочник
    BranchName       string // Колонка 5 - Наименование головного подразделения
    INN              string // Колонка 6 - ИНН юридического лица
    KPP              string // Колонка 7 - КПП
    FacultyName      string // Колонка 8 - Факультет/институт
    Course           int    // Колонка 9 - Курс обучения
    GroupNumber      string // Колонка 10 - Номер группы
    ChatName         string // Колонка 11 - Название чата
    AdminPhone2      string // Колонка 12 - Мобильный номер телефона
    FileName         string // Колонка 13 - Наименование файла
    ChatID           string // Колонка 14 - chat_id
    ChatURL          string // Колонка 15 - link
    AddUser          string // Колонка 16 - add_user
    AddAdmin         string // Колонка 17 - add_admin
}
```

### 4. Обновить функцию readFromExcel

```go
func (uc *MigrateFromExcelUseCase) readFromExcel(filePath string) ([]ExcelRow, error) {
    f, err := excelize.OpenFile(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open Excel file: %w", err)
    }
    defer f.Close()

    sheets := f.GetSheetList()
    if len(sheets) == 0 {
        return nil, fmt.Errorf("no sheets found in Excel file")
    }

    sheetName := sheets[0]
    rows, err := f.GetRows(sheetName)
    if err != nil {
        return nil, fmt.Errorf("failed to get rows: %w", err)
    }

    if len(rows) < 2 {
        return nil, fmt.Errorf("Excel file must have at least a header row and one data row")
    }

    var excelRows []ExcelRow
    ctx := context.Background()
    
    // Пропускаем заголовок (строка 0)
    for i := 1; i < len(rows); i++ {
        row := rows[i]
        
        // Проверяем минимальное количество колонок
        if len(row) < 18 {
            uc.logger.Warn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
                "row_number": i + 1,
                "columns":    len(row),
            })
            continue
        }

        // Парсим курс
        course := 0
        if row[9] != "" {
            course, _ = strconv.Atoi(row[9])
        }

        excelRow := ExcelRow{
            RowNumber:     i + 1,
            AdminPhone1:   row[0],
            MaxID:         row[1],
            INNReference:  row[2],
            FOIVReference: row[3],
            OrgNameRef:    row[4],
            BranchName:    row[5],
            INN:           row[6],
            KPP:           row[7],
            FacultyName:   row[8],
            Course:        course,
            GroupNumber:   row[10],
            ChatName:      row[11],
            AdminPhone2:   row[12],
            FileName:      row[13],
            ChatID:        row[14],
            ChatURL:       row[15],
            AddUser:       row[16],
            AddAdmin:      row[17],
        }

        // Валидация обязательных полей
        if excelRow.INN == "" || excelRow.ChatURL == "" {
            uc.logger.Warn(ctx, "Skipping row: missing required fields", map[string]interface{}{
                "row_number": excelRow.RowNumber,
            })
            continue
        }

        excelRows = append(excelRows, excelRow)
    }

    return excelRows, nil
}
```

### 5. Обновить функцию processRow

```go
func (uc *MigrateFromExcelUseCase) processRow(ctx context.Context, jobID int, row *ExcelRow) error {
    // 1. Создать структуру через Structure Service
    structureData := &domain.StructureData{
        INN:         row.INN,
        KPP:         row.KPP,
        FOIV:        row.FOIVReference,
        OrgName:     row.OrgNameRef,
        BranchName:  row.BranchName,
        FacultyName: row.FacultyName,
        Course:      row.Course,
        GroupNumber: row.GroupNumber,
        ChatName:    row.ChatName,
        ChatURL:     row.ChatURL,
    }

    structureResult, err := uc.structureService.CreateStructure(ctx, structureData)
    if err != nil {
        return fmt.Errorf("failed to create structure: %w", err)
    }

    // 2. Создать чат через Chat Service
    chatData := &domain.ChatData{
        Name:           row.ChatName,
        URL:            row.ChatURL,
        ExternalChatID: row.ChatID,           // ⬅️ НОВОЕ
        UniversityID:   structureResult.UniversityID,
        Source:         "academic_group",
    }

    chatID, err := uc.chatService.CreateChat(ctx, chatData)
    if err != nil {
        return fmt.Errorf("failed to create chat: %w", err)
    }

    // 3. Связать группу с чатом
    if err := uc.structureService.LinkGroupToChat(ctx, structureResult.GroupID, chatID); err != nil {
        uc.logger.Warn(ctx, "Failed to link group to chat", map[string]interface{}{
            "group_id": structureResult.GroupID,
            "chat_id":  chatID,
            "error":    err.Error(),
        })
    }

    // 4. Добавить администратора (если add_admin = ИСТИНА)
    addAdmin := strings.ToUpper(row.AddAdmin) == "ИСТИНА" || 
                strings.ToUpper(row.AddAdmin) == "TRUE"
    addUser := strings.ToUpper(row.AddUser) == "ИСТИНА" || 
               strings.ToUpper(row.AddUser) == "TRUE"

    if addAdmin {
        // Выбираем телефон (приоритет AdminPhone1)
        phone := row.AdminPhone1
        if phone == "" {
            phone = row.AdminPhone2
        }
        
        // Нормализуем телефон
        phone = normalizePhone(phone)
        
        adminData := &domain.AdministratorData{
            ChatID:   chatID,
            Phone:    phone,
            MaxID:    row.MaxID,
            AddUser:  addUser,   // ⬅️ НОВОЕ
            AddAdmin: addAdmin,  // ⬅️ НОВОЕ
        }

        if err := uc.chatService.AddAdministrator(ctx, adminData); err != nil {
            uc.logger.Warn(ctx, "Failed to add administrator", map[string]interface{}{
                "chat_id": chatID,
                "phone":   phone,
                "error":   err.Error(),
            })
        }
    }

    return nil
}

// normalizePhone нормализует номер телефона
func normalizePhone(phone string) string {
    // Удаляем все нецифровые символы
    digits := ""
    for _, r := range phone {
        if r >= '0' && r <= '9' {
            digits += string(r)
        }
    }
    
    // Добавляем + если начинается с 7
    if len(digits) == 11 && digits[0] == '7' {
        return "+" + digits
    }
    
    return digits
}
```

## 📋 Итоговая схема потока данных

```
Excel File (18 колонок)
    ↓
Migration Service (POST /migration/excel)
    ↓
    ├─→ Structure Service (gRPC)
    │   ├─→ universities (INN, KPP, FOIV, Name)
    │   ├─→ branches (Name)
    │   ├─→ faculties (Name)
    │   └─→ groups (Course, Number)
    │
    └─→ Chat Service (gRPC)
        ├─→ chats (Name, URL, ExternalChatID, UniversityID, Source)
        └─→ administrators (Phone, MaxID, AddUser, AddAdmin)
```

## 🎯 Результат после импорта

### Structure Service DB (structure_db):
- ✅ 1 запись в `universities`
- ✅ 1 запись в `branches`
- ✅ 1 запись в `faculties`
- ✅ 1 запись в `groups`

### Chat Service DB (chat_db):
- ✅ 1 запись в `chats` (с external_chat_id)
- ✅ 1 запись в `administrators` (с max_id, add_user, add_admin)

### Migration Service DB (migration_db):
- ✅ 1 запись в `migration_jobs` (статус, прогресс)
- ⚠️ N записей в `migration_errors` (если были ошибки)

## ⏱️ Время на доработку: ~45 минут

1. **Миграции БД** (10 мин) - добавить поля
2. **Go структуры** (15 мин) - обновить domain модели
3. **Excel обработка** (15 мин) - исправить маппинг колонок
4. **Тестирование** (5 мин) - проверить импорт

## 🚀 Готовность: 70%

**Что работает:**
- ✅ Создание иерархии структуры (университет → филиал → факультет → группа)
- ✅ Создание чатов с привязкой к университету
- ✅ Добавление администраторов с телефонами

**Что нужно добавить:**
- 🔧 Поле `external_chat_id` в чатах (Колонка 14)
- 🔧 Поля `add_user`, `add_admin` в администраторах (Колонки 16-17)
- 🔧 Правильный маппинг 18 колонок Excel
- 🔧 Обработка max_id из Excel (Колонка 1)
