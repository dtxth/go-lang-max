package domain

// DepartmentManagerRepository определяет интерфейс для работы с операторами подразделений
type DepartmentManagerRepository interface {
	// CreateDepartmentManager создает назначение оператора на подразделение
	CreateDepartmentManager(dm *DepartmentManager) error
	
	// GetDepartmentManagerByID получает назначение по ID
	GetDepartmentManagerByID(id int64) (*DepartmentManager, error)
	
	// GetDepartmentManagersByEmployeeID получает все назначения оператора
	GetDepartmentManagersByEmployeeID(employeeID int64) ([]*DepartmentManager, error)
	
	// GetDepartmentManagersByBranchID получает всех операторов филиала
	GetDepartmentManagersByBranchID(branchID int64) ([]*DepartmentManager, error)
	
	// GetDepartmentManagersByFacultyID получает всех операторов факультета
	GetDepartmentManagersByFacultyID(facultyID int64) ([]*DepartmentManager, error)
	
	// GetAllDepartmentManagers получает все назначения
	GetAllDepartmentManagers() ([]*DepartmentManager, error)
	
	// DeleteDepartmentManager удаляет назначение
	DeleteDepartmentManager(id int64) error
}
