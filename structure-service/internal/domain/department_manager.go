package domain

import "time"

// DepartmentManager представляет оператора, назначенного на подразделение
type DepartmentManager struct {
	ID         int64     `json:"id"`
	EmployeeID int64     `json:"employee_id"`
	BranchID   *int64    `json:"branch_id,omitempty"`
	FacultyID  *int64    `json:"faculty_id,omitempty"`
	AssignedBy *int64    `json:"assigned_by,omitempty"` // User ID куратора
	AssignedAt time.Time `json:"assigned_at"`
}
