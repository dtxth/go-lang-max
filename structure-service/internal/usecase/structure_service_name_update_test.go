package usecase

import (
	"testing"
	"time"

	"structure-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStructureRepository для тестирования
type MockStructureRepository struct {
	mock.Mock
}

// University methods
func (m *MockStructureRepository) CreateUniversity(university *domain.University) error {
	args := m.Called(university)
	return args.Error(0)
}

func (m *MockStructureRepository) GetUniversityByID(id int64) (*domain.University, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureRepository) GetUniversityByINN(inn string) (*domain.University, error) {
	args := m.Called(inn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureRepository) GetUniversityByINNAndKPP(inn, kpp string) (*domain.University, error) {
	args := m.Called(inn, kpp)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.University), args.Error(1)
}

func (m *MockStructureRepository) UpdateUniversity(university *domain.University) error {
	args := m.Called(university)
	return args.Error(0)
}

func (m *MockStructureRepository) DeleteUniversity(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Branch methods
func (m *MockStructureRepository) CreateBranch(branch *domain.Branch) error {
	args := m.Called(branch)
	return args.Error(0)
}

func (m *MockStructureRepository) GetBranchByID(id int64) (*domain.Branch, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Branch), args.Error(1)
}

func (m *MockStructureRepository) GetBranchByUniversityAndName(universityID int64, name string) (*domain.Branch, error) {
	args := m.Called(universityID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Branch), args.Error(1)
}

func (m *MockStructureRepository) GetBranchesByUniversityID(universityID int64) ([]*domain.Branch, error) {
	args := m.Called(universityID)
	return args.Get(0).([]*domain.Branch), args.Error(1)
}

func (m *MockStructureRepository) UpdateBranch(branch *domain.Branch) error {
	args := m.Called(branch)
	return args.Error(0)
}

func (m *MockStructureRepository) DeleteBranch(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Faculty methods
func (m *MockStructureRepository) CreateFaculty(faculty *domain.Faculty) error {
	args := m.Called(faculty)
	return args.Error(0)
}

func (m *MockStructureRepository) GetFacultyByID(id int64) (*domain.Faculty, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Faculty), args.Error(1)
}

func (m *MockStructureRepository) GetFacultyByBranchAndName(branchID *int64, name string) (*domain.Faculty, error) {
	args := m.Called(branchID, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Faculty), args.Error(1)
}

func (m *MockStructureRepository) GetFacultiesByBranchID(branchID int64) ([]*domain.Faculty, error) {
	args := m.Called(branchID)
	return args.Get(0).([]*domain.Faculty), args.Error(1)
}

func (m *MockStructureRepository) GetFacultiesByUniversityID(universityID int64) ([]*domain.Faculty, error) {
	args := m.Called(universityID)
	return args.Get(0).([]*domain.Faculty), args.Error(1)
}

func (m *MockStructureRepository) UpdateFaculty(faculty *domain.Faculty) error {
	args := m.Called(faculty)
	return args.Error(0)
}

func (m *MockStructureRepository) DeleteFaculty(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Group methods
func (m *MockStructureRepository) CreateGroup(group *domain.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockStructureRepository) GetGroupByID(id int64) (*domain.Group, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockStructureRepository) GetGroupByFacultyAndNumber(facultyID int64, course int, number string) (*domain.Group, error) {
	args := m.Called(facultyID, course, number)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockStructureRepository) GetGroupsByFacultyID(facultyID int64) ([]*domain.Group, error) {
	args := m.Called(facultyID)
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *MockStructureRepository) UpdateGroup(group *domain.Group) error {
	args := m.Called(group)
	return args.Error(0)
}

func (m *MockStructureRepository) DeleteGroup(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

// Structure methods
func (m *MockStructureRepository) GetStructureByUniversityID(universityID int64) (*domain.StructureNode, error) {
	args := m.Called(universityID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.StructureNode), args.Error(1)
}

func (m *MockStructureRepository) GetAllUniversities() ([]*domain.University, error) {
	args := m.Called()
	return args.Get(0).([]*domain.University), args.Error(1)
}

func (m *MockStructureRepository) GetAllUniversitiesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.University, int, error) {
	args := m.Called(limit, offset, sortBy, sortOrder, search)
	return args.Get(0).([]*domain.University), args.Int(1), args.Error(2)
}

// Chat counting methods
func (m *MockStructureRepository) GetChatCountForUniversity(universityID int64) (int, error) {
	args := m.Called(universityID)
	return args.Int(0), args.Error(1)
}

func (m *MockStructureRepository) GetChatCountForBranch(branchID int64) (int, error) {
	args := m.Called(branchID)
	return args.Int(0), args.Error(1)
}

func (m *MockStructureRepository) GetChatCountForFaculty(facultyID int64) (int, error) {
	args := m.Called(facultyID)
	return args.Int(0), args.Error(1)
}

// Tests for UpdateUniversityName
func TestStructureService_UpdateUniversityName_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Old University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedUniversity := &domain.University{
		ID:        1,
		Name:      "New University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: university.CreatedAt,
		UpdatedAt: university.UpdatedAt,
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", updatedUniversity).Return(nil)

	err := service.UpdateUniversityName(1, "New University Name")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateUniversityName_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetUniversityByID", int64(999)).Return(nil, domain.ErrUniversityNotFound)

	err := service.UpdateUniversityName(999, "New University Name")

	assert.Equal(t, domain.ErrUniversityNotFound, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateUniversityName_UpdateError(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Old University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedUniversity := &domain.University{
		ID:        1,
		Name:      "New University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: university.CreatedAt,
		UpdatedAt: university.UpdatedAt,
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", updatedUniversity).Return(assert.AnError)

	err := service.UpdateUniversityName(1, "New University Name")

	assert.Equal(t, assert.AnError, err)
	mockRepo.AssertExpectations(t)
}

// Tests for UpdateBranchName
func TestStructureService_UpdateBranchName_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branch := &domain.Branch{
		ID:           1,
		UniversityID: 1,
		Name:         "Old Branch Name",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	updatedBranch := &domain.Branch{
		ID:           1,
		UniversityID: 1,
		Name:         "New Branch Name",
		CreatedAt:    branch.CreatedAt,
		UpdatedAt:    branch.UpdatedAt,
	}

	mockRepo.On("GetBranchByID", int64(1)).Return(branch, nil)
	mockRepo.On("UpdateBranch", updatedBranch).Return(nil)

	err := service.UpdateBranchName(1, "New Branch Name")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateBranchName_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetBranchByID", int64(999)).Return(nil, domain.ErrBranchNotFound)

	err := service.UpdateBranchName(999, "New Branch Name")

	assert.Equal(t, domain.ErrBranchNotFound, err)
	mockRepo.AssertExpectations(t)
}

// Tests for UpdateFacultyName
func TestStructureService_UpdateFacultyName_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branchID := int64(1)
	faculty := &domain.Faculty{
		ID:        1,
		BranchID:  &branchID,
		Name:      "Old Faculty Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedFaculty := &domain.Faculty{
		ID:        1,
		BranchID:  &branchID,
		Name:      "New Faculty Name",
		CreatedAt: faculty.CreatedAt,
		UpdatedAt: faculty.UpdatedAt,
	}

	mockRepo.On("GetFacultyByID", int64(1)).Return(faculty, nil)
	mockRepo.On("UpdateFaculty", updatedFaculty).Return(nil)

	err := service.UpdateFacultyName(1, "New Faculty Name")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateFacultyName_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetFacultyByID", int64(999)).Return(nil, domain.ErrFacultyNotFound)

	err := service.UpdateFacultyName(999, "New Faculty Name")

	assert.Equal(t, domain.ErrFacultyNotFound, err)
	mockRepo.AssertExpectations(t)
}

// Tests for UpdateGroupName
func TestStructureService_UpdateGroupName_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	chatID := int64(123)
	group := &domain.Group{
		ID:        1,
		FacultyID: 1,
		Course:    2,
		Number:    "ВБ-21",
		ChatID:    &chatID,
		ChatURL:   "https://max.ru/join/test",
		ChatName:  "Test Chat",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedGroup := &domain.Group{
		ID:        1,
		FacultyID: 1,
		Course:    2,
		Number:    "ВБ-22",
		ChatID:    &chatID,
		ChatURL:   "https://max.ru/join/test",
		ChatName:  "Test Chat",
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}

	mockRepo.On("GetGroupByID", int64(1)).Return(group, nil)
	mockRepo.On("UpdateGroup", updatedGroup).Return(nil)

	err := service.UpdateGroupName(1, "ВБ-22")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateGroupName_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetGroupByID", int64(999)).Return(nil, domain.ErrGroupNotFound)

	err := service.UpdateGroupName(999, "ВБ-22")

	assert.Equal(t, domain.ErrGroupNotFound, err)
	mockRepo.AssertExpectations(t)
}

// Tests for GetBranchByID
func TestStructureService_GetBranchByID_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	expectedBranch := &domain.Branch{
		ID:           1,
		UniversityID: 1,
		Name:         "Test Branch",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	mockRepo.On("GetBranchByID", int64(1)).Return(expectedBranch, nil)

	branch, err := service.GetBranchByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedBranch, branch)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_GetBranchByID_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetBranchByID", int64(999)).Return(nil, domain.ErrBranchNotFound)

	branch, err := service.GetBranchByID(999)

	assert.Nil(t, branch)
	assert.Equal(t, domain.ErrBranchNotFound, err)
	mockRepo.AssertExpectations(t)
}

// Tests for GetFacultyByID
func TestStructureService_GetFacultyByID_Success(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	branchID := int64(1)
	expectedFaculty := &domain.Faculty{
		ID:        1,
		BranchID:  &branchID,
		Name:      "Test Faculty",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.On("GetFacultyByID", int64(1)).Return(expectedFaculty, nil)

	faculty, err := service.GetFacultyByID(1)

	assert.NoError(t, err)
	assert.Equal(t, expectedFaculty, faculty)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_GetFacultyByID_NotFound(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	mockRepo.On("GetFacultyByID", int64(999)).Return(nil, domain.ErrFacultyNotFound)

	faculty, err := service.GetFacultyByID(999)

	assert.Nil(t, faculty)
	assert.Equal(t, domain.ErrFacultyNotFound, err)
	mockRepo.AssertExpectations(t)
}

// Edge case tests
func TestStructureService_UpdateUniversityName_EmptyName(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	university := &domain.University{
		ID:        1,
		Name:      "Old University Name",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedUniversity := &domain.University{
		ID:        1,
		Name:      "",
		INN:       "1234567890",
		KPP:       "123456789",
		FOIV:      "Минобрнауки",
		CreatedAt: university.CreatedAt,
		UpdatedAt: university.UpdatedAt,
	}

	mockRepo.On("GetUniversityByID", int64(1)).Return(university, nil)
	mockRepo.On("UpdateUniversity", updatedUniversity).Return(nil)

	err := service.UpdateUniversityName(1, "")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateGroupName_WithoutChat(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	group := &domain.Group{
		ID:        1,
		FacultyID: 1,
		Course:    2,
		Number:    "ВБ-21",
		ChatID:    nil,
		ChatURL:   "",
		ChatName:  "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedGroup := &domain.Group{
		ID:        1,
		FacultyID: 1,
		Course:    2,
		Number:    "ВБ-22",
		ChatID:    nil,
		ChatURL:   "",
		ChatName:  "",
		CreatedAt: group.CreatedAt,
		UpdatedAt: group.UpdatedAt,
	}

	mockRepo.On("GetGroupByID", int64(1)).Return(group, nil)
	mockRepo.On("UpdateGroup", updatedGroup).Return(nil)

	err := service.UpdateGroupName(1, "ВБ-22")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestStructureService_UpdateFacultyName_WithoutBranch(t *testing.T) {
	mockRepo := new(MockStructureRepository)
	service := NewStructureService(mockRepo)

	faculty := &domain.Faculty{
		ID:        1,
		BranchID:  nil, // Faculty directly under university
		Name:      "Old Faculty Name",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedFaculty := &domain.Faculty{
		ID:        1,
		BranchID:  nil,
		Name:      "New Faculty Name",
		CreatedAt: faculty.CreatedAt,
		UpdatedAt: faculty.UpdatedAt,
	}

	mockRepo.On("GetFacultyByID", int64(1)).Return(faculty, nil)
	mockRepo.On("UpdateFaculty", updatedFaculty).Return(nil)

	err := service.UpdateFacultyName(1, "New Faculty Name")

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}