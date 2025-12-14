package domain

import "context"

// AuthService определяет интерфейс для взаимодействия с Auth Service
type AuthService interface {
	// CreateUser создает нового пользователя и возвращает user_id
	CreateUser(ctx context.Context, phone, password string) (int64, error)
	
	// AssignRole назначает роль пользователю
	AssignRole(ctx context.Context, userID int64, role string, universityID, branchID, facultyID *int64) error
	
	// RevokeUserRoles отзывает все роли пользователя
	RevokeUserRoles(ctx context.Context, userID int64) error
}
