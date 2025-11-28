package domain

// EmployeeService определяет интерфейс для взаимодействия с Employee Service
type EmployeeService interface {
	// GetEmployeeByID получает сотрудника по ID
	GetEmployeeByID(id int64) (*Employee, error)
}

// Employee представляет сотрудника из Employee Service
type Employee struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	Role         string `json:"role"`
	UniversityID int64  `json:"university_id"`
}
