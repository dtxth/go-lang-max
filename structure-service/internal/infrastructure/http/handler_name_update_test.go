package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"structure-service/internal/domain"
	"structure-service/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStructureService для тестирования
type MockStructureService struct {
	mock.Mock
}

func (m *MockStructureService) UpdateUniversityName(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(0)
}

func (m *MockStructureService) UpdateBranchName(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(0)
}

func (m *MockStructureService) UpdateFacultyName(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(0)
}

func (m *MockStructureService) UpdateGroupName(id int64, name string) error {
	args := m.Called(id, name)
	return args.Error(0)
}

// Остальные методы интерфейса (заглушки)
func (m *MockStructureService) GetStructure(universityID int64) (*domain.StructureNode, error) {
	args := m.Called(universityID)
	return args.Get(0).(*domain.StructureNode), args.Error(1)
}

func (m *MockStructureService) GetAllUniversities() ([]*domain.University, error) {
	args := m.Called()
	return args.Get(0).([]*domain.University), args.Error(1)
}

func (m *MockStructureService) GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.University, int, error) {
	args := m.Called(limit, offset, sortBy, sortOrder, search)
	return args.Get(0).([]*domain.University), args.Int(1), args.Error(2)
}

func (m *MockStructureService) GetUniversity(id int64) (*domain.University, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureService) GetUniversityByINN(inn string) (*domain.University, error) {
	args := m.Called(inn)
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureService) CreateUniversity(u *domain.University) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockStructureService) CreateOrGetUniversity(inn, kpp, name, foiv string) (*domain.University, error) {
	args := m.Called(inn, kpp, name, foiv)
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureService) UpdateUniversity(u *domain.University) error {
	args := m.Called(u)
	return args.Error(0)
}

func (m *MockStructureService) DeleteUniversity(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStructureService) CreateBranch(b *domain.Branch) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockStructureService) UpdateBranch(b *domain.Branch) error {
	args := m.Called(b)
	return args.Error(0)
}

func (m *MockStructureService) DeleteBranch(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStructureService) CreateFaculty(f *domain.Faculty) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *MockStructureService) UpdateFaculty(f *domain.Faculty) error {
	args := m.Called(f)
	return args.Error(0)
}

func (m *MockStructureService) DeleteFaculty(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStructureService) CreateGroup(g *domain.Group) error {
	args := m.Called(g)
	return args.Error(0)
}

func (m *MockStructureService) GetGroupByID(id int64) (*domain.Group, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockStructureService) UpdateGroup(g *domain.Group) error {
	args := m.Called(g)
	return args.Error(0)
}

func (m *MockStructureService) DeleteGroup(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockStructureService) GetBranchByID(id int64) (*domain.Branch, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Branch), args.Error(1)
}

func (m *MockStructureService) GetFacultyByID(id int64) (*domain.Faculty, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Faculty), args.Error(1)
}

func (m *MockStructureService) ImportFromExcel(rows []*domain.ExcelRow) error {
	args := m.Called(rows)
	return args.Error(0)
}

func TestUpdateUniversityName_Success(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	mockService.On("UpdateUniversityName", int64(1), "New University Name").Return(nil)

	reqBody := UpdateNameRequest{Name: "New University Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/universities/1/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateUniversityName(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "university name updated successfully")
	mockService.AssertExpectations(t)
}

func TestUpdateUniversityName_InvalidID(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	reqBody := UpdateNameRequest{Name: "New University Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/universities/invalid/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateUniversityName(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid university id")
}

func TestUpdateUniversityName_EmptyName(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	reqBody := UpdateNameRequest{Name: "   "}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/universities/1/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateUniversityName(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "name cannot be empty")
}

func TestUpdateUniversityName_NotFound(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	mockService.On("UpdateUniversityName", int64(999), "New University Name").Return(domain.ErrUniversityNotFound)

	reqBody := UpdateNameRequest{Name: "New University Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/universities/999/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateUniversityName(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateBranchName_Success(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	mockService.On("UpdateBranchName", int64(1), "New Branch Name").Return(nil)

	reqBody := UpdateNameRequest{Name: "New Branch Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/branches/1/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateBranchName(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "branch name updated successfully")
	mockService.AssertExpectations(t)
}

func TestUpdateFacultyName_Success(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	mockService.On("UpdateFacultyName", int64(1), "New Faculty Name").Return(nil)

	reqBody := UpdateNameRequest{Name: "New Faculty Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/faculties/1/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateFacultyName(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "faculty name updated successfully")
	mockService.AssertExpectations(t)
}

func TestUpdateGroupName_Success(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	mockService.On("UpdateGroupName", int64(1), "New Group Name").Return(nil)

	reqBody := UpdateNameRequest{Name: "New Group Name"}
	jsonBody, _ := json.Marshal(reqBody)
	
	req := httptest.NewRequest(http.MethodPut, "/groups/1/name", bytes.NewBuffer(jsonBody))
	w := httptest.NewRecorder()

	handler.UpdateGroupName(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "group name updated successfully")
	mockService.AssertExpectations(t)
}

// Тест роутера для проверки правильности маршрутизации
func TestRouter_NameUpdateEndpoints(t *testing.T) {
	mockService := new(MockStructureService)
	
	handler := NewHandler(
		mockService,
		&usecase.GetUniversityStructureUseCase{},
		&usecase.AssignOperatorToDepartmentUseCase{},
		&usecase.ImportStructureFromExcelUseCase{},
		&usecase.CreateStructureFromRowUseCase{},
		nil,
		nil,
	)

	// Тестируем напрямую handlers без middleware
	testCases := []struct {
		name     string
		method   string
		path     string
		handler  func(http.ResponseWriter, *http.Request)
		expected int
	}{
		{"PUT university name", http.MethodPut, "/universities/1/name", handler.UpdateUniversityName, http.StatusBadRequest},
		{"PUT branch name", http.MethodPut, "/branches/1/name", handler.UpdateBranchName, http.StatusBadRequest},
		{"PUT faculty name", http.MethodPut, "/faculties/1/name", handler.UpdateFacultyName, http.StatusBadRequest},
		{"PUT group name", http.MethodPut, "/groups/1/name", handler.UpdateGroupName, http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader("{}"))
			w := httptest.NewRecorder()

			tc.handler(w, req)

			assert.Equal(t, tc.expected, w.Code)
		})
	}
}