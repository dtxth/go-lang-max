package domain

import "context"

// EmployeeServiceInterface определяет интерфейс для сервиса сотрудников
type EmployeeServiceInterface interface {
	// AddEmployeeByPhone добавляет сотрудника по номеру телефона
	AddEmployeeByPhone(phone, firstName, lastName, middleName, inn, kpp, universityName string) (*Employee, error)
	
	// SearchEmployees выполняет поиск сотрудников
	SearchEmployees(query string, limit, offset int) ([]*Employee, error)
	
	// GetAllEmployees получает всех сотрудников с пагинацией
	GetAllEmployees(limit, offset int) ([]*Employee, error)
	
	// GetAllEmployeesWithSortingAndSearch получает всех сотрудников с пагинацией, сортировкой и поиском
	GetAllEmployeesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*Employee, int, error)
	
	// GetEmployeeByID получает сотрудника по ID
	GetEmployeeByID(id int64) (*Employee, error)
	
	// GetEmployeeByMaxID получает сотрудника по MAX ID
	GetEmployeeByMaxID(maxID string) (*Employee, error)
	
	// UpdateEmployee обновляет данные сотрудника
	UpdateEmployee(employee *Employee) error
	
	// DeleteEmployee удаляет сотрудника
	DeleteEmployee(id int64) error
	
	// CreateEmployeeWithRole создает сотрудника с назначением роли
	CreateEmployeeWithRole(ctx context.Context, phone, firstName, lastName, middleName, inn, kpp, universityName, role, requesterRole string) (*Employee, error)
	
	// GetUniversityByID получает вуз по ID
	GetUniversityByID(id int64) (*University, error)
	
	// GetUniversityByINN получает вуз по ИНН
	GetUniversityByINN(inn string) (*University, error)
	
	// GetUniversityByINNAndKPP получает вуз по ИНН и КПП
	GetUniversityByINNAndKPP(inn, kpp string) (*University, error)
}