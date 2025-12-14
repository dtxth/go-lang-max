package usecase

import (
	"structure-service/internal/domain"
	"testing"
)

// Mock implementations for testing
type mockDepartmentManagerRepo struct {
	createFunc func(*domain.DepartmentManager) error
}

func (m *mockDepartmentManagerRepo) CreateDepartmentManager(dm *domain.DepartmentManager) error {
	if m.createFunc != nil {
		return m.createFunc(dm)
	}
	dm.ID = 1
	return nil
}

func (m *mockDepartmentManagerRepo) GetDepartmentManagerByID(id int64) (*domain.DepartmentManager, error) {
	return nil, nil
}

func (m *mockDepartmentManagerRepo) GetDepartmentManagersByEmployeeID(employeeID int64) ([]*domain.DepartmentManager, error) {
	return nil, nil
}

func (m *mockDepartmentManagerRepo) GetDepartmentManagersByBranchID(branchID int64) ([]*domain.DepartmentManager, error) {
	return nil, nil
}

func (m *mockDepartmentManagerRepo) GetDepartmentManagersByFacultyID(facultyID int64) ([]*domain.DepartmentManager, error) {
	return nil, nil
}

func (m *mockDepartmentManagerRepo) GetAllDepartmentManagers() ([]*domain.DepartmentManager, error) {
	return nil, nil
}

func (m *mockDepartmentManagerRepo) DeleteDepartmentManager(id int64) error {
	return nil
}

type mockEmployeeService struct {
	getEmployeeFunc func(int64) (*domain.Employee, error)
}

func (m *mockEmployeeService) GetEmployeeByID(id int64) (*domain.Employee, error) {
	if m.getEmployeeFunc != nil {
		return m.getEmployeeFunc(id)
	}
	return &domain.Employee{
		ID:           id,
		FirstName:    "Test",
		LastName:     "Operator",
		Phone:        "+79991234567",
		Role:         "operator",
		UniversityID: 1,
	}, nil
}

func TestAssignOperatorToDepartmentUseCase_Execute(t *testing.T) {
	tests := []struct {
		name          string
		employeeID    int64
		branchID      *int64
		facultyID     *int64
		assignedBy    *int64
		mockEmployee  func(int64) (*domain.Employee, error)
		mockCreate    func(*domain.DepartmentManager) error
		expectError   bool
		errorContains string
	}{
		{
			name:       "successful assignment to branch",
			employeeID: 1,
			branchID:   int64Ptr(1),
			facultyID:  nil,
			assignedBy: int64Ptr(2),
			mockEmployee: func(id int64) (*domain.Employee, error) {
				return &domain.Employee{
					ID:           id,
					FirstName:    "Test",
					LastName:     "Operator",
					Phone:        "+79991234567",
					Role:         "operator",
					UniversityID: 1,
				}, nil
			},
			expectError: false,
		},
		{
			name:       "successful assignment to faculty",
			employeeID: 1,
			branchID:   nil,
			facultyID:  int64Ptr(1),
			assignedBy: int64Ptr(2),
			mockEmployee: func(id int64) (*domain.Employee, error) {
				return &domain.Employee{
					ID:           id,
					FirstName:    "Test",
					LastName:     "Operator",
					Phone:        "+79991234567",
					Role:         "operator",
					UniversityID: 1,
				}, nil
			},
			expectError: false,
		},
		{
			name:          "error when no department specified",
			employeeID:    1,
			branchID:      nil,
			facultyID:     nil,
			assignedBy:    int64Ptr(2),
			expectError:   true,
			errorContains: "invalid department",
		},
		{
			name:       "error when employee not found",
			employeeID: 999,
			branchID:   int64Ptr(1),
			facultyID:  nil,
			assignedBy: int64Ptr(2),
			mockEmployee: func(id int64) (*domain.Employee, error) {
				return nil, domain.ErrEmployeeNotFound
			},
			expectError:   true,
			errorContains: "failed to verify employee",
		},
		{
			name:       "error when employee is not operator",
			employeeID: 1,
			branchID:   int64Ptr(1),
			facultyID:  nil,
			assignedBy: int64Ptr(2),
			mockEmployee: func(id int64) (*domain.Employee, error) {
				return &domain.Employee{
					ID:           id,
					FirstName:    "Test",
					LastName:     "Curator",
					Phone:        "+79991234567",
					Role:         "curator",
					UniversityID: 1,
				}, nil
			},
			expectError:   true,
			errorContains: "must have operator role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockDepartmentManagerRepo{
				createFunc: tt.mockCreate,
			}
			mockEmpService := &mockEmployeeService{
				getEmployeeFunc: tt.mockEmployee,
			}

			uc := NewAssignOperatorToDepartmentUseCase(mockRepo, mockEmpService)
			dm, err := uc.Execute(tt.employeeID, tt.branchID, tt.facultyID, tt.assignedBy)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error to contain '%s', got '%s'", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if dm == nil {
					t.Errorf("expected department manager but got nil")
				}
			}
		})
	}
}

func int64Ptr(i int64) *int64 {
	return &i
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
