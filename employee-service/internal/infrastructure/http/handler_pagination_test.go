package http

import (
	"context"
	"employee-service/internal/domain"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// mockEmployeeServiceForPagination is a mock for testing pagination
type mockEmployeeServiceForPagination struct {
	employees []*domain.Employee
	total     int
}

func (m *mockEmployeeServiceForPagination) GetAllEmployeesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, int, error) {
	// Simulate filtering by search
	filtered := m.employees
	if search != "" {
		filtered = []*domain.Employee{}
		for _, emp := range m.employees {
			if containsIgnoreCase(emp.FirstName, search) ||
				containsIgnoreCase(emp.LastName, search) ||
				containsIgnoreCase(emp.Phone, search) {
				filtered = append(filtered, emp)
			}
		}
	}
	
	// Simulate pagination
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []*domain.Employee{}, len(filtered), nil
	}
	if end > len(filtered) {
		end = len(filtered)
	}
	
	return filtered[start:end], len(filtered), nil
}

func containsIgnoreCase(str, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	if len(str) == 0 {
		return false
	}
	
	// Simple case-insensitive contains check
	strLower := strings.ToLower(str)
	substrLower := strings.ToLower(substr)
	
	for i := 0; i <= len(strLower)-len(substrLower); i++ {
		if strLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
}

func TestGetAllEmployees_WithPagination(t *testing.T) {
	// Create test data
	university := &domain.University{
		ID:   1,
		Name: "Test University",
		INN:  "1234567890",
		KPP:  "123456789",
	}
	
	employees := []*domain.Employee{
		{
			ID:           1,
			FirstName:    "Иван",
			LastName:     "Иванов",
			Phone:        "+79001234567",
			UniversityID: 1,
			University:   university,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           2,
			FirstName:    "Петр",
			LastName:     "Петров",
			Phone:        "+79001234568",
			UniversityID: 1,
			University:   university,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
	
	mockService := &mockEmployeeServiceForPagination{
		employees: employees,
		total:     2,
	}
	
	handler := &Handler{
		employeeService: &mockEmployeeServiceWrapper{mockPagination: mockService},
	}
	
	// Test with pagination parameters
	req := httptest.NewRequest("GET", "/employees/all?limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	
	handler.GetAllEmployees(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedEmployeesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 employee, got %d", len(response.Data))
	}
	
	if response.Total != 2 {
		t.Errorf("Expected total 2, got %d", response.Total)
	}
	
	if response.Limit != 1 {
		t.Errorf("Expected limit 1, got %d", response.Limit)
	}
	
	if response.Offset != 0 {
		t.Errorf("Expected offset 0, got %d", response.Offset)
	}
	
	if response.TotalPages != 2 {
		t.Errorf("Expected total pages 2, got %d", response.TotalPages)
	}
}

func TestGetAllEmployees_WithSearch(t *testing.T) {
	// Create test data
	university := &domain.University{
		ID:   1,
		Name: "Test University",
		INN:  "1234567890",
		KPP:  "123456789",
	}
	
	employees := []*domain.Employee{
		{
			ID:           1,
			FirstName:    "Иван",
			LastName:     "Иванов",
			Phone:        "+79001234567",
			UniversityID: 1,
			University:   university,
		},
		{
			ID:           2,
			FirstName:    "Петр",
			LastName:     "Петров",
			Phone:        "+79001234568",
			UniversityID: 1,
			University:   university,
		},
	}
	
	mockService := &mockEmployeeServiceForPagination{
		employees: employees,
		total:     2,
	}
	
	handler := &Handler{
		employeeService: &mockEmployeeServiceWrapper{mockPagination: mockService},
	}
	
	// Test with search parameter
	params := url.Values{}
	params.Add("search", "Иван")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/employees/all?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handler.GetAllEmployees(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedEmployeesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 employee after search, got %d", len(response.Data))
	}
	
	if response.Data[0].FirstName != "Иван" {
		t.Errorf("Expected first name 'Иван', got '%s'", response.Data[0].FirstName)
	}
}

func TestGetAllEmployees_WithSorting(t *testing.T) {
	// Create test data
	university := &domain.University{
		ID:   1,
		Name: "Test University",
		INN:  "1234567890",
		KPP:  "123456789",
	}
	
	employees := []*domain.Employee{
		{
			ID:           1,
			FirstName:    "Иван",
			LastName:     "Иванов",
			Phone:        "+79001234567",
			UniversityID: 1,
			University:   university,
		},
		{
			ID:           2,
			FirstName:    "Петр",
			LastName:     "Петров",
			Phone:        "+79001234568",
			UniversityID: 1,
			University:   university,
		},
	}
	
	mockService := &mockEmployeeServiceForPagination{
		employees: employees,
		total:     2,
	}
	
	handler := &Handler{
		employeeService: &mockEmployeeServiceWrapper{mockPagination: mockService},
	}
	
	// Test with sorting parameters
	params := url.Values{}
	params.Add("sort_by", "first_name")
	params.Add("sort_order", "desc")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/employees/all?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handler.GetAllEmployees(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedEmployeesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 2 {
		t.Errorf("Expected 2 employees, got %d", len(response.Data))
	}
}

// mockEmployeeServiceWrapper wraps the pagination mock to satisfy the interface
type mockEmployeeServiceWrapper struct {
	mockPagination *mockEmployeeServiceForPagination
}

func (m *mockEmployeeServiceWrapper) GetAllEmployeesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, int, error) {
	return m.mockPagination.GetAllEmployeesWithSortingAndSearch(limit, offset, sortBy, sortOrder, search)
}

// Implement other required methods as no-ops for testing
func (m *mockEmployeeServiceWrapper) AddEmployeeByPhone(phone, firstName, lastName, middleName, inn, kpp, universityName string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) SearchEmployees(query string, limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) GetAllEmployees(limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) GetEmployeeByID(id int64) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) UpdateEmployee(employee *domain.Employee) error {
	return nil
}

func (m *mockEmployeeServiceWrapper) DeleteEmployee(id int64) error {
	return nil
}

func (m *mockEmployeeServiceWrapper) CreateEmployeeWithRole(ctx context.Context, phone, firstName, lastName, middleName, inn, kpp, universityName, role, requesterRole string) (*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) GetUniversityByID(id int64) (*domain.University, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) GetUniversityByINN(inn string) (*domain.University, error) {
	return nil, nil
}

func (m *mockEmployeeServiceWrapper) GetUniversityByINNAndKPP(inn, kpp string) (*domain.University, error) {
	return nil, nil
}