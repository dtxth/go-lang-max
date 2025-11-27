package usecase

import (
	"employee-service/internal/domain"
	"errors"
	"testing"
	"time"
)

// Mock implementations for testing

type mockEmployeeRepo struct {
	employees map[int64]*domain.Employee
	nextID    int64
}

func newMockEmployeeRepo() *mockEmployeeRepo {
	return &mockEmployeeRepo{
		employees: make(map[int64]*domain.Employee),
		nextID:    1,
	}
}

func (m *mockEmployeeRepo) Create(e *domain.Employee) error {
	e.ID = m.nextID
	m.nextID++
	e.CreatedAt = time.Now()
	e.UpdatedAt = time.Now()
	m.employees[e.ID] = e
	return nil
}

func (m *mockEmployeeRepo) GetByID(id int64) (*domain.Employee, error) {
	e, ok := m.employees[id]
	if !ok {
		return nil, domain.ErrEmployeeNotFound
	}
	return e, nil
}

func (m *mockEmployeeRepo) GetByPhone(phone string) (*domain.Employee, error) {
	for _, e := range m.employees {
		if e.Phone == phone {
			return e, nil
		}
	}
	return nil, domain.ErrEmployeeNotFound
}

func (m *mockEmployeeRepo) Update(e *domain.Employee) error {
	if _, ok := m.employees[e.ID]; !ok {
		return domain.ErrEmployeeNotFound
	}
	e.UpdatedAt = time.Now()
	m.employees[e.ID] = e
	return nil
}

func (m *mockEmployeeRepo) Delete(id int64) error {
	if _, ok := m.employees[id]; !ok {
		return domain.ErrEmployeeNotFound
	}
	delete(m.employees, id)
	return nil
}

func (m *mockEmployeeRepo) Search(query string, limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetAll(limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetByMaxID(maxID string) (*domain.Employee, error) {
	for _, e := range m.employees {
		if e.MaxID == maxID {
			return e, nil
		}
	}
	return nil, domain.ErrEmployeeNotFound
}

type mockUniversityRepo struct {
	universities map[int64]*domain.University
	nextID       int64
}

func newMockUniversityRepo() *mockUniversityRepo {
	return &mockUniversityRepo{
		universities: make(map[int64]*domain.University),
		nextID:       1,
	}
}

func (m *mockUniversityRepo) Create(u *domain.University) error {
	u.ID = m.nextID
	m.nextID++
	m.universities[u.ID] = u
	return nil
}

func (m *mockUniversityRepo) GetByID(id int64) (*domain.University, error) {
	u, ok := m.universities[id]
	if !ok {
		return nil, domain.ErrUniversityNotFound
	}
	return u, nil
}

func (m *mockUniversityRepo) GetByINN(inn string) (*domain.University, error) {
	for _, u := range m.universities {
		if u.INN == inn {
			return u, nil
		}
	}
	return nil, domain.ErrUniversityNotFound
}

func (m *mockUniversityRepo) GetByINNAndKPP(inn, kpp string) (*domain.University, error) {
	for _, u := range m.universities {
		if u.INN == inn && u.KPP == kpp {
			return u, nil
		}
	}
	return nil, domain.ErrUniversityNotFound
}

func (m *mockUniversityRepo) SearchByName(query string) ([]*domain.University, error) {
	return nil, nil
}

func (m *mockUniversityRepo) GetAll() ([]*domain.University, error) {
	return nil, nil
}

type mockMaxService struct {
	shouldFail bool
	maxID      string
}

func (m *mockMaxService) GetMaxIDByPhone(phone string) (string, error) {
	if m.shouldFail {
		return "", errors.New("MAX API unavailable")
	}
	if m.maxID != "" {
		return m.maxID, nil
	}
	return "max_" + phone, nil
}

func (m *mockMaxService) ValidatePhone(phone string) bool {
	return len(phone) > 5
}

// Test: Employee creation triggers MAX_id lookup (Requirements 3.1)
func TestAddEmployeeByPhone_TriggersMaxIDLookup(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxService{maxID: "max_123456"}

	service := NewEmployeeService(employeeRepo, universityRepo, maxService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Иван",
		"Иванов",
		"Иванович",
		"1234567890",
		"123456789",
		"МГУ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "max_123456" {
		t.Errorf("Expected MAX_id to be 'max_123456', got '%s'", employee.MaxID)
	}
}

// Test: MAX_id is stored when received (Requirements 3.4)
func TestAddEmployeeByPhone_StoresMaxIDWhenReceived(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxService{maxID: "max_987654"}

	service := NewEmployeeService(employeeRepo, universityRepo, maxService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Петр",
		"Петров",
		"",
		"9876543210",
		"987654321",
		"СПбГУ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "max_987654" {
		t.Errorf("Expected MAX_id to be 'max_987654', got '%s'", employee.MaxID)
	}

	if employee.MaxIDUpdatedAt == nil {
		t.Error("Expected MaxIDUpdatedAt to be set, got nil")
	}
}

// Test: Employee creation succeeds without MAX_id (Requirements 3.5)
func TestAddEmployeeByPhone_SucceedsWithoutMaxID(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxService{shouldFail: true}

	service := NewEmployeeService(employeeRepo, universityRepo, maxService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Сергей",
		"Сергеев",
		"Сергеевич",
		"5555555555",
		"555555555",
		"МФТИ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "" {
		t.Errorf("Expected MAX_id to be empty, got '%s'", employee.MaxID)
	}

	if employee.MaxIDUpdatedAt != nil {
		t.Error("Expected MaxIDUpdatedAt to be nil when MAX_id is not set")
	}

	if employee.Phone != "+79001234567" {
		t.Errorf("Expected phone to be '+79001234567', got '%s'", employee.Phone)
	}
}
