package domain

import "time"

// Employee представляет сотрудника вуза
type Employee struct {
	ID                   int64       `json:"id"`
	FirstName            string      `json:"first_name"`
	LastName             string      `json:"last_name"`
	MiddleName           string      `json:"middle_name,omitempty"`
	Phone                string      `json:"phone"`                    // Номер телефона
	MaxID                string      `json:"max_id"`                   // MAX_id (заменяет телефон)
	INN                  string      `json:"inn,omitempty"`            // ИНН
	KPP                  string      `json:"kpp,omitempty"`            // КПП
	Role                 string      `json:"role,omitempty"`           // Роль: curator, operator, или пусто
	UserID               *int64      `json:"user_id,omitempty"`        // ID пользователя в auth-service
	MaxIDUpdatedAt       *time.Time  `json:"max_id_updated_at,omitempty"` // Время последнего обновления MAX_id
	ProfileSource        string      `json:"profile_source"`           // Источник профильной информации: webhook, user_input, default
	ProfileLastUpdated   *time.Time  `json:"profile_last_updated,omitempty"` // Время последнего обновления профиля
	UniversityID         int64       `json:"university_id"`
	University           *University `json:"university,omitempty"`
	CreatedAt            time.Time   `json:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at"`
}

// FullName возвращает полное ФИО сотрудника
func (e *Employee) FullName() string {
	if e.MiddleName != "" {
		return e.LastName + " " + e.FirstName + " " + e.MiddleName
	}
	return e.LastName + " " + e.FirstName
}
