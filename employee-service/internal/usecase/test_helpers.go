package usecase

import (
	"context"
	"employee-service/internal/domain"
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

func (m *mockEmployeeRepo) GetByMaxID(maxID string) (*domain.Employee, error) {
	for _, e := range m.employees {
		if e.MaxID == maxID {
			return e, nil
		}
	}
	return nil, domain.ErrEmployeeNotFound
}

func (m *mockEmployeeRepo) Search(query string, limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
}

func (m *mockEmployeeRepo) GetAll(limit, offset int) ([]*domain.Employee, error) {
	return nil, nil
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

func (m *mockEmployeeRepo) GetWithoutMaxID(limit int) ([]*domain.Employee, error) {
	var result []*domain.Employee
	count := 0
	for _, e := range m.employees {
		if e.MaxID == "" && count < limit {
			result = append(result, e)
			count++
		}
	}
	return result, nil
}

func (m *mockEmployeeRepo) GetEmployeesWithoutMaxID(limit, offset int) ([]*domain.Employee, error) {
	var result []*domain.Employee
	for _, e := range m.employees {
		if e.MaxID == "" {
			result = append(result, e)
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

func (m *mockEmployeeRepo) CountEmployeesWithoutMaxID() (int, error) {
	count := 0
	for _, e := range m.employees {
		if e.MaxID == "" {
			count++
		}
	}
	return count, nil
}

func (m *mockEmployeeRepo) GetAllWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, error) {
	var result []*domain.Employee
	for _, e := range m.employees {
		result = append(result, e)
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

func (m *mockEmployeeRepo) CountAllWithSearch(search string) (int, error) {
	return len(m.employees), nil
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
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()
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

func (m *mockUniversityRepo) Update(u *domain.University) error {
	if _, ok := m.universities[u.ID]; !ok {
		return domain.ErrUniversityNotFound
	}
	u.UpdatedAt = time.Now()
	m.universities[u.ID] = u
	return nil
}

func (m *mockUniversityRepo) Delete(id int64) error {
	if _, ok := m.universities[id]; !ok {
		return domain.ErrUniversityNotFound
	}
	delete(m.universities, id)
	return nil
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
	var result []*domain.University
	for _, u := range m.universities {
		result = append(result, u)
	}
	return result, nil
}

type mockMaxService struct {
	users map[string]string // phone -> maxID
}

func newMockMaxService() *mockMaxService {
	return &mockMaxService{
		users: make(map[string]string),
	}
}

func (m *mockMaxService) GetMaxIDByPhone(phone string) (string, error) {
	if maxID, ok := m.users[phone]; ok {
		return maxID, nil
	}
	return "", domain.ErrMaxIDNotFound
}

func (m *mockMaxService) ValidatePhone(phone string) bool {
	return len(phone) > 5
}

func (m *mockMaxService) BatchGetMaxIDByPhone(phones []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, phone := range phones {
		if maxID, ok := m.users[phone]; ok {
			result[phone] = maxID
		}
	}
	return result, nil
}

type mockAuthService struct {
	nextUserID int64
}

func newMockAuthService() *mockAuthService {
	return &mockAuthService{
		nextUserID: 1,
	}
}

func (m *mockAuthService) CreateUser(ctx context.Context, phone, password string) (int64, error) {
	userID := m.nextUserID
	m.nextUserID++
	return userID, nil
}

func (m *mockAuthService) AssignRole(ctx context.Context, userID int64, role string, universityID, branchID, facultyID *int64) error {
	return nil
}

func (m *mockAuthService) RevokeUserRoles(ctx context.Context, userID int64) error {
	return nil
}

type mockPasswordGenerator struct{}

func newMockPasswordGenerator() *mockPasswordGenerator {
	return &mockPasswordGenerator{}
}

func (m *mockPasswordGenerator) Generate(length int) (string, error) {
	// Return a fixed password for testing
	return "TestPass123!", nil
}

type mockNotificationService struct {
	sentNotifications []struct {
		phone    string
		password string
	}
}

func newMockNotificationService() *mockNotificationService {
	return &mockNotificationService{
		sentNotifications: make([]struct {
			phone    string
			password string
		}, 0),
	}
}

func (m *mockNotificationService) SendPasswordNotification(ctx context.Context, phone, password string) error {
	m.sentNotifications = append(m.sentNotifications, struct {
		phone    string
		password string
	}{phone: phone, password: password})
	return nil
}

type mockBatchUpdateJobRepo struct {
	jobs   map[int64]*domain.BatchUpdateJob
	nextID int64
}

func newMockBatchUpdateJobRepo() *mockBatchUpdateJobRepo {
	return &mockBatchUpdateJobRepo{
		jobs:   make(map[int64]*domain.BatchUpdateJob),
		nextID: 1,
	}
}

func (m *mockBatchUpdateJobRepo) Create(job *domain.BatchUpdateJob) error {
	job.ID = m.nextID
	m.nextID++
	job.StartedAt = time.Now()
	m.jobs[job.ID] = job
	return nil
}

func (m *mockBatchUpdateJobRepo) Update(job *domain.BatchUpdateJob) error {
	if _, ok := m.jobs[job.ID]; !ok {
		return domain.ErrEmployeeNotFound
	}
	m.jobs[job.ID] = job
	return nil
}

func (m *mockBatchUpdateJobRepo) GetByID(id int64) (*domain.BatchUpdateJob, error) {
	job, ok := m.jobs[id]
	if !ok {
		return nil, domain.ErrEmployeeNotFound
	}
	return job, nil
}

func (m *mockBatchUpdateJobRepo) GetAll(limit, offset int) ([]*domain.BatchUpdateJob, error) {
	return nil, nil
}
