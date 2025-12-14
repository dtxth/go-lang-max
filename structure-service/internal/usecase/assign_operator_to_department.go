package usecase

import (
	"fmt"
	"structure-service/internal/domain"
)

// AssignOperatorToDepartmentUseCase назначает оператора на подразделение
type AssignOperatorToDepartmentUseCase struct {
	dmRepo          domain.DepartmentManagerRepository
	employeeService domain.EmployeeService
}

// NewAssignOperatorToDepartmentUseCase создает новый use case
func NewAssignOperatorToDepartmentUseCase(
	dmRepo domain.DepartmentManagerRepository,
	employeeService domain.EmployeeService,
) *AssignOperatorToDepartmentUseCase {
	return &AssignOperatorToDepartmentUseCase{
		dmRepo:          dmRepo,
		employeeService: employeeService,
	}
}

// Execute выполняет назначение оператора на подразделение
func (uc *AssignOperatorToDepartmentUseCase) Execute(
	employeeID int64,
	branchID *int64,
	facultyID *int64,
	assignedBy *int64,
) (*domain.DepartmentManager, error) {
	// Проверяем, что указан хотя бы один из department (branch или faculty)
	if branchID == nil && facultyID == nil {
		return nil, domain.ErrInvalidDepartment
	}

	// Проверяем, что сотрудник существует в Employee Service
	employee, err := uc.employeeService.GetEmployeeByID(employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify employee: %w", err)
	}

	// Проверяем, что сотрудник имеет роль оператора
	if employee.Role != "operator" {
		return nil, fmt.Errorf("employee must have operator role, got: %s", employee.Role)
	}

	// Создаем запись department_manager
	dm := &domain.DepartmentManager{
		EmployeeID: employeeID,
		BranchID:   branchID,
		FacultyID:  facultyID,
		AssignedBy: assignedBy,
	}

	if err := uc.dmRepo.CreateDepartmentManager(dm); err != nil {
		return nil, fmt.Errorf("failed to create department manager: %w", err)
	}

	return dm, nil
}
