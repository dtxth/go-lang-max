package domain

// EmployeeClient определяет интерфейс для взаимодействия с employee-service
type EmployeeClient interface {
	// UpdateEmployeeByMaxID обновляет данные сотрудника по MAX ID
	UpdateEmployeeByMaxID(maxID, firstName, lastName, username string) error
}