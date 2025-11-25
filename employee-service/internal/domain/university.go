package domain

import "time"

// University представляет вуз
type University struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	INN       string    `json:"inn"`
	KPP       string    `json:"kpp,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

