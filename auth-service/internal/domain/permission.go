package domain

// Permission представляет разрешение на выполнение действия с ресурсом
type Permission struct {
	Resource string // Тип ресурса (chat, employee, structure, etc.)
	Action   string // Действие (read, write, delete, etc.)
}

// PermissionContext содержит контекст для проверки разрешений
type PermissionContext struct {
	UserID       int64
	Role         string
	UniversityID *int64
	BranchID     *int64
	FacultyID    *int64
	
	// Контекст ресурса
	ResourceUniversityID *int64
	ResourceBranchID     *int64
	ResourceFacultyID    *int64
}

// PermissionValidator определяет интерфейс для валидации разрешений
type PermissionValidator interface {
	// ValidatePermission проверяет, имеет ли пользователь разрешение на действие
	ValidatePermission(ctx *PermissionContext, permission *Permission) (bool, error)
}
