package usecase

import (
	"context"
	"employee-service/internal/domain"
	"errors"
	"time"
)

// UpdateEmployeeWithRoleSyncUseCase обновляет сотрудника с синхронизацией роли
type UpdateEmployeeWithRoleSyncUseCase struct {
	employeeRepo   domain.EmployeeRepository
	universityRepo domain.UniversityRepository
	maxService     domain.MaxService
	authService    domain.AuthService
}

// NewUpdateEmployeeWithRoleSyncUseCase создает новый use case
func NewUpdateEmployeeWithRoleSyncUseCase(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
	authService domain.AuthService,
) *UpdateEmployeeWithRoleSyncUseCase {
	return &UpdateEmployeeWithRoleSyncUseCase{
		employeeRepo:   employeeRepo,
		universityRepo: universityRepo,
		maxService:     maxService,
		authService:    authService,
	}
}

// Execute выполняет обновление сотрудника с синхронизацией роли
func (uc *UpdateEmployeeWithRoleSyncUseCase) Execute(
	ctx context.Context,
	employeeID int64,
	firstName, lastName, middleName string,
	phone string,
	inn, kpp string,
	universityID int64,
	newRole string,
	requesterRole string,
) (*domain.Employee, error) {
	// Получаем существующего сотрудника
	existingEmployee, err := uc.employeeRepo.GetByID(employeeID)
	if err != nil {
		return nil, domain.ErrEmployeeNotFound
	}

	// Валидация новой роли
	if newRole != "" && newRole != "curator" && newRole != "operator" {
		return nil, errors.New("invalid role: must be 'curator' or 'operator'")
	}

	// Проверка прав на изменение роли
	if existingEmployee.Role != newRole {
		// Curator может изменять роль только на operator или убирать роль
		if requesterRole == "curator" {
			if newRole == "curator" {
				return nil, errors.New("curator cannot assign curator role")
			}
			if existingEmployee.Role == "curator" {
				return nil, errors.New("curator cannot change curator role")
			}
		}
	}

	// Обновляем поля сотрудника
	if firstName != "" {
		existingEmployee.FirstName = firstName
	}
	if lastName != "" {
		existingEmployee.LastName = lastName
	}
	if middleName != "" {
		existingEmployee.MiddleName = middleName
	}

	// Если изменился телефон, обновляем MAX_id
	if phone != "" && phone != existingEmployee.Phone {
		if !uc.maxService.ValidatePhone(phone) {
			return nil, domain.ErrInvalidPhone
		}

		maxID, err := uc.maxService.GetMaxIDByPhone(phone)
		if err == nil && maxID != "" {
			existingEmployee.MaxID = maxID
			now := time.Now()
			existingEmployee.MaxIDUpdatedAt = &now
		}
		existingEmployee.Phone = phone
	}

	if inn != "" {
		existingEmployee.INN = inn
	}
	if kpp != "" {
		existingEmployee.KPP = kpp
	}
	if universityID > 0 {
		existingEmployee.UniversityID = universityID
	}

	// Обрабатываем изменение роли
	roleChanged := existingEmployee.Role != newRole
	oldRole := existingEmployee.Role
	existingEmployee.Role = newRole

	// Если роль изменилась, синхронизируем с Auth Service
	if roleChanged {
		if existingEmployee.UserID != nil {
			// Если была роль, отзываем старую
			if oldRole != "" {
				err := uc.authService.RevokeUserRoles(ctx, *existingEmployee.UserID)
				if err != nil {
					return nil, errors.New("failed to revoke old role: " + err.Error())
				}
			}

			// Если новая роль не пустая, назначаем её
			if newRole != "" {
				err := uc.authService.AssignRole(ctx, *existingEmployee.UserID, newRole, &existingEmployee.UniversityID, nil, nil)
				if err != nil {
					return nil, errors.New("failed to assign new role: " + err.Error())
				}
			} else {
				// Если роль убрали, очищаем user_id
				existingEmployee.UserID = nil
			}
		} else if newRole != "" {
			// Если не было user_id, но теперь есть роль, создаем user_id
			userID := existingEmployee.ID
			existingEmployee.UserID = &userID

			err := uc.authService.AssignRole(ctx, userID, newRole, &existingEmployee.UniversityID, nil, nil)
			if err != nil {
				return nil, errors.New("failed to assign role: " + err.Error())
			}
		}
	}

	// Сохраняем изменения
	if err := uc.employeeRepo.Update(existingEmployee); err != nil {
		return nil, err
	}

	// Загружаем полную информацию о сотруднике
	return uc.employeeRepo.GetByID(employeeID)
}
