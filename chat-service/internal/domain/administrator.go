package domain

import "time"

// Administrator представляет администратора чата
type Administrator struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Phone     string    `json:"phone"`      // Номер телефона администратора
	MaxID     string    `json:"max_id"`     // MAX_id администратора
	AddUser   bool      `json:"add_user"`   // Может ли добавлять пользователей
	AddAdmin  bool      `json:"add_admin"`  // Может ли добавлять администраторов
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

