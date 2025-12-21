package usecase

import (
	"employee-service/internal/domain"
	"errors"
	"testing"
	"time"
)

// Mock repositories and services specific to batch update tests
type mockEmployeeRepoForBatch struct {
	employees          []*domain.Employee
	countWithoutMaxID  int
	updateCalled       int
}

func (m *mockEmployeeRepoForBatch) Create(employee *domain.Employee) error {
	employee.ID = int64(len(m.employees) + 1)
	employee.CreatedAt = time.Now()
	employee.UpdatedAt = time.Now()
	m.employees = append(m.employees, employee)
	return nil
}

func (m *mockEmployeeRepoForBatch) GetByID(id int64) (*domain.Employee, error) {
	for _, emp := range m.employees {
		if emp.ID == id {
			return emp, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockEmployeeRepoForBatch) GetByPhone(phone string) (*domain.Employee, error) {
	return nil, errors.New("not found")
}

func (m *mockEmployeeRepoForBatch) GetByMaxID(maxID string) (*domain.Employee, error) {
	return nil, errors.New("not found")
}

func (m *mockEmployeeRepoForBatch) Search(query string, limit, offset int) ([]*domain.Employee, error) {
	return m.employees, nil
}

func (m *mockEmployeeRepoForBatch) GetAll(limit, offset int) ([]*domain.Employee, error) {
	return m.employees, nil
}

func (m *mockEmployeeRepoForBatch) Update(employee *domain.Employee) error {
	m.updateCalled++
	for i, emp := range m.employees {
		if emp.ID == employee.ID {
			m.employees[i] = employee
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockEmployeeRepoForBatch) Delete(id int64) error {
	return nil
}

func (m *mockEmployeeRepoForBatch) GetWithoutMaxID(limit int) ([]*domain.Employee, error) {
	var result []*domain.Employee
	count := 0
	for _, emp := range m.employees {
		if emp.MaxID == "" && count < limit {
			result = append(result, emp)
			count++
		}
	}
	return result, nil
}

func (m *mockEmployeeRepoForBatch) GetEmployeesWithoutMaxID(limit, offset int) ([]*domain.Employee, error) {
	var result []*domain.Employee
	for _, emp := range m.employees {
		if emp.MaxID == "" {
			result = append(result, emp)
		}
	}
	
	// Apply pagination
	if offset >= len(result) {
		return []*domain.Employee{}, nil
	}
	
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	
	return result[offset:end], nil
}

func (m *mockEmployeeRepoForBatch) CountEmployeesWithoutMaxID() (int, error) {
	return m.countWithoutMaxID, nil
}

func (m *mockEmployeeRepoForBatch) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, error) {
	return m.employees, nil
}

func (m *mockEmployeeRepoForBatch) CountAllWithSearch(search string) (int, error) {
	return len(m.employees), nil
}

type mockMaxServiceForBatch struct {
	maxIDs map[string]string
}

func (m *mockMaxServiceForBatch) GetMaxIDByPhone(phone string) (string, error) {
	if maxID, ok := m.maxIDs[phone]; ok {
		return maxID, nil
	}
	return "", errors.New("not found")
}

func (m *mockMaxServiceForBatch) BatchGetMaxIDByPhone(phones []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, phone := range phones {
		if maxID, ok := m.maxIDs[phone]; ok {
			result[phone] = maxID
		}
	}
	return result, nil
}

func (m *mockMaxServiceForBatch) GetUserProfileByPhone(phone string) (*domain.UserProfile, error) {
	if maxID, ok := m.maxIDs[phone]; ok {
		return &domain.UserProfile{
			MaxID:     maxID,
			Phone:     phone,
			FirstName: "Test",
			LastName:  "User",
		}, nil
	}
	return nil, errors.New("not found")
}

func TestBatchUpdateMaxId_EmptyDatabase(t *testing.T) {
	employeeRepo := &mockEmployeeRepoForBatch{
		employees:         []*domain.Employee{},
		countWithoutMaxID: 0,
	}
	batchJobRepo := newMockBatchUpdateJobRepo()
	maxService := &mockMaxServiceForBatch{
		maxIDs: make(map[string]string),
	}
	
	uc := NewBatchUpdateMaxIdUseCase(employeeRepo, batchJobRepo, maxService)
	
	result, err := uc.StartBatchUpdate()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Total != 0 {
		t.Errorf("Expected total 0, got %d", result.Total)
	}
	
	if result.Success != 0 {
		t.Errorf("Expected success 0, got %d", result.Success)
	}
}

func TestBatchUpdateMaxId_SuccessfulUpdate(t *testing.T) {
	// Create employees without MAX_id
	employees := []*domain.Employee{
		{ID: 1, Phone: "+79001234567", MaxID: "", FirstName: "Ivan", LastName: "Ivanov"},
		{ID: 2, Phone: "+79001234568", MaxID: "", FirstName: "Petr", LastName: "Petrov"},
	}
	
	employeeRepo := &mockEmployeeRepoForBatch{
		employees:         employees,
		countWithoutMaxID: 2,
	}
	
	batchJobRepo := newMockBatchUpdateJobRepo()
	
	maxService := &mockMaxServiceForBatch{
		maxIDs: map[string]string{
			"+79001234567": "max_id_1",
			"+79001234568": "max_id_2",
		},
	}
	
	uc := NewBatchUpdateMaxIdUseCase(employeeRepo, batchJobRepo, maxService)
	
	result, err := uc.StartBatchUpdate()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Total != 2 {
		t.Errorf("Expected total 2, got %d", result.Total)
	}
	
	if result.Success != 2 {
		t.Errorf("Expected success 2, got %d", result.Success)
	}
	
	if result.Failed != 0 {
		t.Errorf("Expected failed 0, got %d", result.Failed)
	}
	
	// Verify employees were updated
	if employeeRepo.updateCalled != 2 {
		t.Errorf("Expected Update to be called 2 times, got %d", employeeRepo.updateCalled)
	}
}

func TestBatchUpdateMaxId_PartialFailure(t *testing.T) {
	// Create employees without MAX_id
	employees := []*domain.Employee{
		{ID: 1, Phone: "+79001234567", MaxID: "", FirstName: "Ivan", LastName: "Ivanov"},
		{ID: 2, Phone: "+79001234568", MaxID: "", FirstName: "Petr", LastName: "Petrov"},
		{ID: 3, Phone: "+79001234569", MaxID: "", FirstName: "Sidor", LastName: "Sidorov"},
	}
	
	employeeRepo := &mockEmployeeRepoForBatch{
		employees:         employees,
		countWithoutMaxID: 3,
	}
	
	batchJobRepo := newMockBatchUpdateJobRepo()
	
	// Only 2 out of 3 phones have MAX_id
	maxService := &mockMaxServiceForBatch{
		maxIDs: map[string]string{
			"+79001234567": "max_id_1",
			"+79001234568": "max_id_2",
			// +79001234569 not found
		},
	}
	
	uc := NewBatchUpdateMaxIdUseCase(employeeRepo, batchJobRepo, maxService)
	
	result, err := uc.StartBatchUpdate()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Total != 3 {
		t.Errorf("Expected total 3, got %d", result.Total)
	}
	
	if result.Success != 2 {
		t.Errorf("Expected success 2, got %d", result.Success)
	}
	
	if result.Failed != 1 {
		t.Errorf("Expected failed 1, got %d", result.Failed)
	}
}

func TestBatchUpdateMaxId_BatchSizeLimit(t *testing.T) {
	// Create 150 employees without MAX_id
	employees := make([]*domain.Employee, 150)
	maxIDs := make(map[string]string)
	
	for i := 0; i < 150; i++ {
		phone := "+7900123" + string(rune('0'+i%10)) + string(rune('0'+(i/10)%10)) + string(rune('0'+(i/100)%10)) + "0"
		employees[i] = &domain.Employee{
			ID:        int64(i + 1),
			Phone:     phone,
			MaxID:     "",
			FirstName: "User",
			LastName:  "Test",
		}
		maxIDs[phone] = "max_id_" + string(rune('0'+i))
	}
	
	employeeRepo := &mockEmployeeRepoForBatch{
		employees:         employees,
		countWithoutMaxID: 150,
	}
	
	batchJobRepo := newMockBatchUpdateJobRepo()
	
	maxService := &mockMaxServiceForBatch{
		maxIDs: maxIDs,
	}
	
	uc := NewBatchUpdateMaxIdUseCase(employeeRepo, batchJobRepo, maxService)
	
	result, err := uc.StartBatchUpdate()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	
	if result.Total != 150 {
		t.Errorf("Expected total 150, got %d", result.Total)
	}
	
	// Should process all employees in batches of 100
	if result.Success < 100 {
		t.Errorf("Expected at least 100 successful updates, got %d", result.Success)
	}
}
