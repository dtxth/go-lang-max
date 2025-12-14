package domain

import "time"

// University представляет вуз
type University struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	INN        string    `json:"inn"`
	KPP        string    `json:"kpp,omitempty"`
	FOIV       string    `json:"foiv,omitempty"` // ФОИВ
	ChatsCount int       `json:"chats_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Branch представляет филиал/головное подразделение
type Branch struct {
	ID           int64     `json:"id"`
	UniversityID int64     `json:"university_id"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Faculty представляет факультет/институт/иная структурная классификация
type Faculty struct {
	ID        int64     `json:"id"`
	BranchID  *int64    `json:"branch_id,omitempty"` // Может быть NULL, если факультет напрямую привязан к вузу
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Group представляет группу
type Group struct {
	ID         int64     `json:"id"`
	FacultyID  int64     `json:"faculty_id"`
	Course     int       `json:"course"`      // Курс обучения
	Number     string    `json:"number"`     // Номер группы
	ChatID     *int64    `json:"chat_id,omitempty"` // ID чата (может быть NULL)
	ChatURL    string    `json:"chat_url,omitempty"` // URL чата
	ChatName   string    `json:"chat_name,omitempty"` // Название чата
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Chat представляет чат (связь с chat-service)
type Chat struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	MaxID string `json:"max_id,omitempty"`
}

// StructureNode представляет узел иерархической структуры для отображения
type StructureNode struct {
	Type      string           `json:"type"` // "university", "branch", "faculty", "group"
	ID        int64            `json:"id"`
	Name      string           `json:"name"`
	Children  []*StructureNode `json:"children,omitempty"`
	Chat      *Chat            `json:"chat,omitempty"`
	Course    *int             `json:"course,omitempty"`
	GroupNum  *string          `json:"group_num,omitempty"`
	ChatCount *int             `json:"chat_count"`
}

// ExcelRow представляет строку из Excel файла
type ExcelRow struct {
	AdminPhone      string // Номер телефона администратора
	INN             string // ИНН
	FOIV            string // ФОИВ
	Organization    string // Наименование организации
	Branch          string // Наименование головного подразделения/филиала
	KPP             string // КПП
	Faculty         string // Факультет/институт/иная структурная классификация
	Course          int    // Курс обучения
	GroupNumber     string // Номер группы
	ChatName        string // Название чата
	ChatURL         string // Ссылка на чат
	ChatID          string // ID чата (из ссылки)
}

// ImportResult представляет результат импорта структуры
type ImportResult struct {
	Created int      `json:"created"` // Количество созданных записей
	Updated int      `json:"updated"` // Количество обновленных записей
	Failed  int      `json:"failed"`  // Количество неудачных записей
	Errors  []string `json:"errors,omitempty"` // Список ошибок
}

