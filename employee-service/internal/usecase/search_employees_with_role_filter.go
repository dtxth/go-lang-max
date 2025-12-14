package usecase

import (
	"context"
	"employee-service/internal/domain"
)

// SearchEmployeesWithRoleFilterUseCase выполняет поиск сотрудников с фильтрацией по ролям
type SearchEmployeesWithRoleFilterUseCase struct {
	employeeRepo domain.EmployeeRepository
	authService  domain.AuthService
}

// NewSearchEmployeesWithRoleFilterUseCase создает новый use case для поиска сотрудников
func NewSearchEmployeesWithRoleFilterUseCase(
	employeeRepo domain.EmployeeRepository,
	authService domain.AuthService,
) *SearchEmployeesWithRoleFilterUseCase {
	return &SearchEmployeesWithRoleFilterUseCase{
		employeeRepo: employeeRepo,
		authService:  authService,
	}
}

// SearchEmployeeResult представляет результат поиска сотрудника
type SearchEmployeeResult struct {
	ID             int64  `json:"id"`
	FullName       string `json:"full_name"`
	Phone          string `json:"phone"`
	Role           string `json:"role"`
	UniversityName string `json:"university_name"`
}

// Execute выполняет поиск сотрудников с применением ролевой фильтрации
// Requirements: 14.1, 14.2, 14.3, 14.4
func (uc *SearchEmployeesWithRoleFilterUseCase) Execute(
	ctx context.Context,
	query string,
	userRole string,
	universityID *int64,
	limit, offset int,
) ([]*SearchEmployeeResult, error) {
	// Валидация и установка значений по умолчанию для пагинации
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем сотрудников из репозитория
	employees, err := uc.employeeRepo.Search(query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Применяем ролевую фильтрацию
	// Requirements 14.2: Superadmin видит всех, Curator видит только своих
	var filteredEmployees []*domain.Employee
	for _, emp := range employees {
		if uc.shouldIncludeEmployee(emp, userRole, universityID) {
			filteredEmployees = append(filteredEmployees, emp)
		}
	}

	// Преобразуем в результаты поиска с требуемыми полями
	// Requirements 14.4: включаем full name, phone, role, university name
	results := make([]*SearchEmployeeResult, 0, len(filteredEmployees))
	for _, emp := range filteredEmployees {
		result := &SearchEmployeeResult{
			ID:       emp.ID,
			FullName: emp.FullName(),
			Phone:    emp.Phone,
			Role:     emp.Role,
		}
		
		// Добавляем название университета
		if emp.University != nil {
			result.UniversityName = emp.University.Name
		}
		
		results = append(results, result)
	}

	return results, nil
}

// shouldIncludeEmployee определяет, должен ли сотрудник быть включен в результаты
// на основе роли пользователя, выполняющего поиск
func (uc *SearchEmployeesWithRoleFilterUseCase) shouldIncludeEmployee(
	employee *domain.Employee,
	userRole string,
	universityID *int64,
) bool {
	// Requirements 14.2: Superadmin видит всех сотрудников
	if userRole == "superadmin" {
		return true
	}

	// Requirements 14.3: Curator видит только сотрудников своего университета
	if userRole == "curator" {
		if universityID == nil {
			return false
		}
		return employee.UniversityID == *universityID
	}

	// Operator и другие роли не имеют доступа к поиску сотрудников
	return false
}
