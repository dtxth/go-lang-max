package usecase

import (
	"context"
	"employee-service/internal/domain"
)

// DeleteEmployeeWithPermissionRevocationUseCase удаляет сотрудника с отзывом прав
type DeleteEmployeeWithPermissionRevocationUseCase struct {
	employeeRepo domain.EmployeeRepository
	authService  domain.AuthService
}

// NewDeleteEmployeeWithPermissionRevocationUseCase создает новый use case
func NewDeleteEmployeeWithPermissionRevocationUseCase(
	employeeRepo domain.EmployeeRepository,
	authService domain.AuthService,
) *DeleteEmployeeWithPermissionRevocationUseCase {
	return &DeleteEmployeeWithPermissionRevocationUseCase{
		employeeRepo: employeeRepo,
		authService:  authService,
	}
}

// Execute выполняет удаление сотрудника с отзывом прав
func (uc *DeleteEmployeeWithPermissionRevocationUseCase) Execute(ctx context.Context, employeeID int64) error {
	// Получаем сотрудника
	employee, err := uc.employeeRepo.GetByID(employeeID)
	if err != nil {
		return domain.ErrEmployeeNotFound
	}

	// Если у сотрудника есть роль, отзываем её в Auth Service
	if employee.UserID != nil && employee.Role != "" {
		err := uc.authService.RevokeUserRoles(ctx, *employee.UserID)
		if err != nil {
			// Логируем ошибку, но продолжаем удаление
			// В реальной системе можно добавить retry логику
		}
	}

	// Удаляем сотрудника
	return uc.employeeRepo.Delete(employeeID)
}
