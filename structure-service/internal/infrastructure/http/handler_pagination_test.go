package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"structure-service/internal/domain"
	"testing"
	"time"
)

// mockStructureServiceForPagination is a mock for testing pagination
type mockStructureServiceForPagination struct {
	universities []*domain.University
	total        int
}

func (m *mockStructureServiceForPagination) GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.University, int, error) {
	// Simulate filtering by search
	filtered := m.universities
	if search != "" {
		filtered = []*domain.University{}
		for _, uni := range m.universities {
			if containsIgnoreCase(uni.Name, search) ||
				containsIgnoreCase(uni.INN, search) ||
				containsIgnoreCase(uni.KPP, search) ||
				containsIgnoreCase(uni.FOIV, search) {
				filtered = append(filtered, uni)
			}
		}
	}
	
	// Simulate pagination
	start := offset
	end := offset + limit
	if start > len(filtered) {
		return []*domain.University{}, len(filtered), nil
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

func TestGetAllUniversities_WithPagination(t *testing.T) {
	// Create test data
	universities := []*domain.University{
		{
			ID:        1,
			Name:      "МГУ им. М.В. Ломоносова",
			INN:       "1234567890",
			KPP:       "123456789",
			FOIV:      "Минобрнауки",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			Name:      "СПбГУ",
			INN:       "0987654321",
			KPP:       "098765432",
			FOIV:      "Минобрнауки",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	mockService := &mockStructureServiceForPagination{
		universities: universities,
		total:        2,
	}
	
	handler := &Handler{
		structureService: &mockStructureServiceWrapper{mockPagination: mockService},
	}
	
	// Test with pagination parameters
	req := httptest.NewRequest("GET", "/universities?limit=1&offset=0", nil)
	w := httptest.NewRecorder()
	
	handler.GetAllUniversities(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedUniversitiesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 university, got %d", len(response.Data))
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

func TestGetAllUniversities_WithSearch(t *testing.T) {
	// Create test data
	universities := []*domain.University{
		{
			ID:        1,
			Name:      "МГУ им. М.В. Ломоносова",
			INN:       "1234567890",
			KPP:       "123456789",
			FOIV:      "Минобрнауки",
		},
		{
			ID:        2,
			Name:      "СПбГУ",
			INN:       "0987654321",
			KPP:       "098765432",
			FOIV:      "Минобрнауки",
		},
	}
	
	mockService := &mockStructureServiceForPagination{
		universities: universities,
		total:        2,
	}
	
	handler := &Handler{
		structureService: &mockStructureServiceWrapper{mockPagination: mockService},
	}
	
	// Test with search parameter
	params := url.Values{}
	params.Add("search", "МГУ")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/universities?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handler.GetAllUniversities(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedUniversitiesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 1 {
		t.Errorf("Expected 1 university after search, got %d", len(response.Data))
	}
	
	if response.Data[0].Name != "МГУ им. М.В. Ломоносова" {
		t.Errorf("Expected university name 'МГУ им. М.В. Ломоносова', got '%s'", response.Data[0].Name)
	}
}

func TestGetAllUniversities_WithSorting(t *testing.T) {
	// Create test data
	universities := []*domain.University{
		{
			ID:        1,
			Name:      "Я-Университет",
			INN:       "1234567890",
			KPP:       "123456789",
			FOIV:      "Минобрнауки",
		},
		{
			ID:        2,
			Name:      "А-Университет",
			INN:       "0987654321",
			KPP:       "098765432",
			FOIV:      "Минобрнауки",
		},
	}
	
	mockService := &mockStructureServiceForPagination{
		universities: universities,
		total:        2,
	}
	
	handler := &Handler{
		structureService: &mockStructureServiceWrapper{mockPagination: mockService},
	}
	
	// Test with sorting parameters
	params := url.Values{}
	params.Add("sort_by", "name")
	params.Add("sort_order", "desc")
	params.Add("limit", "10")
	params.Add("offset", "0")
	
	req := httptest.NewRequest("GET", "/universities?"+params.Encode(), nil)
	w := httptest.NewRecorder()
	
	handler.GetAllUniversities(w, req)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	
	var response PaginatedUniversitiesResponse
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	
	if len(response.Data) != 2 {
		t.Errorf("Expected 2 universities, got %d", len(response.Data))
	}
}

// mockStructureServiceWrapper wraps the pagination mock to satisfy the interface
type mockStructureServiceWrapper struct {
	mockPagination *mockStructureServiceForPagination
}

func (m *mockStructureServiceWrapper) GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.University, int, error) {
	return m.mockPagination.GetAllUniversitiesWithSortingAndSearch(limit, offset, sortBy, sortOrder, search)
}

// Implement other required methods as no-ops for testing
func (m *mockStructureServiceWrapper) GetStructure(universityID int64) (*domain.StructureNode, error) {
	return nil, nil
}

func (m *mockStructureServiceWrapper) GetAllUniversities() ([]*domain.University, error) {
	return nil, nil
}

func (m *mockStructureServiceWrapper) GetUniversity(id int64) (*domain.University, error) {
	return nil, nil
}

func (m *mockStructureServiceWrapper) GetUniversityByINN(inn string) (*domain.University, error) {
	return nil, nil
}

func (m *mockStructureServiceWrapper) CreateUniversity(u *domain.University) error {
	return nil
}

func (m *mockStructureServiceWrapper) UpdateUniversity(u *domain.University) error {
	return nil
}

func (m *mockStructureServiceWrapper) DeleteUniversity(id int64) error {
	return nil
}

func (m *mockStructureServiceWrapper) CreateBranch(b *domain.Branch) error {
	return nil
}

func (m *mockStructureServiceWrapper) UpdateBranch(b *domain.Branch) error {
	return nil
}

func (m *mockStructureServiceWrapper) DeleteBranch(id int64) error {
	return nil
}

func (m *mockStructureServiceWrapper) CreateFaculty(f *domain.Faculty) error {
	return nil
}

func (m *mockStructureServiceWrapper) UpdateFaculty(f *domain.Faculty) error {
	return nil
}

func (m *mockStructureServiceWrapper) DeleteFaculty(id int64) error {
	return nil
}

func (m *mockStructureServiceWrapper) CreateGroup(g *domain.Group) error {
	return nil
}

func (m *mockStructureServiceWrapper) GetGroupByID(id int64) (*domain.Group, error) {
	return nil, nil
}

func (m *mockStructureServiceWrapper) UpdateGroup(g *domain.Group) error {
	return nil
}

func (m *mockStructureServiceWrapper) DeleteGroup(id int64) error {
	return nil
}

func (m *mockStructureServiceWrapper) ImportFromExcel(rows []*domain.ExcelRow) error {
	return nil
}