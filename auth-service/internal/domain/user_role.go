package domain

import "time"

// UserRole представляет связь пользователя с ролью и контекстом
type UserRole struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	RoleID       int64     `json:"role_id"`
	UniversityID *int64    `json:"university_id,omitempty"` // NULL для superadmin
	BranchID     *int64    `json:"branch_id,omitempty"`     // NULL для curator
	FacultyID    *int64    `json:"faculty_id,omitempty"`    // NULL для curator
	AssignedBy   *int64    `json:"assigned_by,omitempty"`
	AssignedAt   time.Time `json:"assigned_at"`
}

// UserRoleWithDetails содержит информацию о роли пользователя с деталями роли
type UserRoleWithDetails struct {
	UserRole
	RoleName string `json:"role_name"`
}

// UserRoleRepository определяет интерфейс для работы с назначениями ролей
type UserRoleRepository interface {
	// Create создает новое назначение роли
	Create(ur *UserRole) error
	
	// GetByUserID возвращает все роли пользователя
	GetByUserID(userID int64) ([]*UserRoleWithDetails, error)
	
	// Delete удаляет назначение роли
	Delete(id int64) error
	
	// DeleteByUserID удаляет все роли пользователя
	DeleteByUserID(userID int64) error
	
	// GetByUserIDAndRole возвращает роль пользователя по имени роли
	GetByUserIDAndRole(userID int64, roleName string) (*UserRoleWithDetails, error)
	
	// GetRoleByName возвращает роль по имени
	GetRoleByName(name string) (*Role, error)
}
