package domain

import "time"

// Chat представляет групповой чат
type Chat struct {
	ID                int64           `json:"id"`
	Name              string          `json:"name"`                    // Название чата
	URL               string          `json:"url"`                     // Ссылка на чат
	MaxChatID         string          `json:"max_chat_id"`             // ID чата в MAX
	ParticipantsCount int             `json:"participants_count"`      // Количество участников
	UniversityID      *int64          `json:"university_id,omitempty"` // ID вуза (опционально)
	University        *University     `json:"university,omitempty"`    // Вуз (для суперадмина)
	Department        string          `json:"department,omitempty"`    // Подразделение вуза
	Source            string          `json:"source"`                  // Источник: "admin_panel", "bot_registrar", "academic_group"
	Administrators    []Administrator `json:"administrators"`          // Администраторы чата
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// University представляет вуз (переиспользуем из employee-service)
type University struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	INN       string    `json:"inn"`
	KPP       string    `json:"kpp,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
