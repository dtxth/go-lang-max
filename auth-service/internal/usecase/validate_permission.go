package usecase

import (
	"auth-service/internal/domain"
	"fmt"
)

type ValidatePermissionUseCase struct {
	userRoleRepo domain.UserRoleRepository
	roleRepo     domain.RoleRepository
}

func NewValidatePermissionUseCase(
	userRoleRepo domain.UserRoleRepository,
	roleRepo domain.RoleRepository,
) *ValidatePermissionUseCase {
	return &ValidatePermissionUseCase{
		userRoleRepo: userRoleRepo,
		roleRepo:     roleRepo,
	}
}

// ValidatePermission проверяет, имеет ли пользователь разрешение на действие с ресурсом
func (uc *ValidatePermissionUseCase) ValidatePermission(
	ctx *domain.PermissionContext,
	permission *domain.Permission,
) (bool, error) {
	// Получаем роли пользователя
	userRoles, err := uc.userRoleRepo.GetByUserID(ctx.UserID)
	if err != nil {
		return false, fmt.Errorf("failed to get user roles: %w", err)
	}
	
	if len(userRoles) == 0 {
		return false, nil
	}
	
	// Проверяем каждую роль пользователя
	for _, userRole := range userRoles {
		hasPermission := uc.checkRolePermission(userRole, ctx, permission)
		if hasPermission {
			return true, nil
		}
	}
	
	return false, nil
}

// checkRolePermission проверяет разрешение для конкретной роли
func (uc *ValidatePermissionUseCase) checkRolePermission(
	userRole *domain.UserRoleWithDetails,
	ctx *domain.PermissionContext,
	permission *domain.Permission,
) bool {
	// Superadmin имеет доступ ко всему
	if userRole.RoleName == "superadmin" {
		return true
	}
	
	// Curator имеет доступ только к своему вузу
	if userRole.RoleName == "curator" {
		// Проверяем, что у куратора указан вуз
		if userRole.UniversityID == nil {
			return false
		}
		
		// Проверяем, что ресурс принадлежит вузу куратора
		if ctx.ResourceUniversityID != nil {
			return *userRole.UniversityID == *ctx.ResourceUniversityID
		}
		
		// Если контекст ресурса не указан, разрешаем доступ
		// (это может быть создание нового ресурса)
		return true
	}
	
	// Operator имеет доступ только к своему филиалу или факультету
	if userRole.RoleName == "operator" {
		// Проверяем доступ по филиалу
		if userRole.BranchID != nil && ctx.ResourceBranchID != nil {
			if *userRole.BranchID == *ctx.ResourceBranchID {
				return true
			}
		}
		
		// Проверяем доступ по факультету
		if userRole.FacultyID != nil && ctx.ResourceFacultyID != nil {
			if *userRole.FacultyID == *ctx.ResourceFacultyID {
				return true
			}
		}
		
		// Если у оператора указан вуз, но не указан филиал/факультет,
		// проверяем доступ по вузу
		if userRole.UniversityID != nil && ctx.ResourceUniversityID != nil {
			return *userRole.UniversityID == *ctx.ResourceUniversityID
		}
		
		return false
	}
	
	return false
}

// GetUserPermissions возвращает все разрешения пользователя
func (uc *ValidatePermissionUseCase) GetUserPermissions(userID int64) ([]*domain.UserRoleWithDetails, error) {
	return uc.userRoleRepo.GetByUserID(userID)
}
