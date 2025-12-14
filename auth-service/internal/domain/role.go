package domain

import "time"

// Role представляет роль в системе
type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// RoleRepository определяет интерфейс для работы с ролями
type RoleRepository interface {
	// GetByName возвращает роль по имени
	GetByName(name string) (*Role, error)
	
	// GetByID возвращает роль по ID
	GetByID(id int64) (*Role, error)
	
	// List возвращает все роли
	List() ([]*Role, error)
}
