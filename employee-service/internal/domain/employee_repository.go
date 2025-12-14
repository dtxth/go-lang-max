package domain

// EmployeeRepository определяет интерфейс для работы с сотрудниками
type EmployeeRepository interface {
	// Create создает нового сотрудника
	Create(employee *Employee) error
	
	// GetByID получает сотрудника по ID
	GetByID(id int64) (*Employee, error)
	
	// GetByPhone получает сотрудника по номеру телефона
	GetByPhone(phone string) (*Employee, error)
	
	// GetByMaxID получает сотрудника по MAX_id
	GetByMaxID(maxID string) (*Employee, error)
	
	// Search выполняет поиск сотрудников по имени, фамилии и названию вуза
	Search(query string, limit, offset int) ([]*Employee, error)
	
	// GetAll получает всех сотрудников с пагинацией
	GetAll(limit, offset int) ([]*Employee, error)
	
	// GetAllWithSortingAndSearch получает всех сотрудников с пагинацией, сортировкой и поиском
	GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*Employee, error)
	
	// CountAllWithSearch подсчитывает общее количество сотрудников с учетом поиска
	CountAllWithSearch(search string) (int, error)
	
	// Update обновляет данные сотрудника
	Update(employee *Employee) error
	
	// Delete удаляет сотрудника
	Delete(id int64) error
	
	// GetEmployeesWithoutMaxID получает сотрудников без MAX_id
	GetEmployeesWithoutMaxID(limit, offset int) ([]*Employee, error)
	
	// CountEmployeesWithoutMaxID подсчитывает количество сотрудников без MAX_id
	CountEmployeesWithoutMaxID() (int, error)
}

