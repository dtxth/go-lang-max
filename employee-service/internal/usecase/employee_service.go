package usecase

import (
	"employee-service/internal/domain"
	"errors"
	"strings"
	"time"
)

type EmployeeService struct {
	employeeRepo   domain.EmployeeRepository
	universityRepo domain.UniversityRepository
	maxService     domain.MaxService
}

func NewEmployeeService(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
) *EmployeeService {
	return &EmployeeService{
		employeeRepo:   employeeRepo,
		universityRepo: universityRepo,
		maxService:     maxService,
	}
}

// AddEmployeeByPhone добавляет сотрудника по номеру телефона
// Автоматически получает MAX_id и создает или находит вуз по ИНН/КПП
// Если MAX_id не найден, сотрудник создается без него (Requirements 3.5)
func (s *EmployeeService) AddEmployeeByPhone(
	phone string,
	firstName, lastName, middleName string,
	inn, kpp string,
	universityName string,
) (*domain.Employee, error) {
	// Валидация телефона
	if !s.maxService.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}
	
	// Проверяем, не существует ли уже сотрудник с таким телефоном
	existing, _ := s.employeeRepo.GetByPhone(phone)
	if existing != nil {
		return nil, domain.ErrEmployeeExists
	}
	
	// Получаем MAX_id по телефону (Requirements 3.1)
	maxID, err := s.maxService.GetMaxIDByPhone(phone)
	if err != nil {
		// Логируем ошибку, но продолжаем без MAX_id (Requirements 3.5)
		maxID = ""
	}
	
	// Находим или создаем вуз
	university, err := s.findOrCreateUniversity(inn, kpp, universityName)
	if err != nil {
		return nil, err
	}
	
	// Создаем сотрудника
	employee := &domain.Employee{
		FirstName:    strings.TrimSpace(firstName),
		LastName:      strings.TrimSpace(lastName),
		MiddleName:    strings.TrimSpace(middleName),
		Phone:         phone,
		MaxID:         maxID,
		INN:           strings.TrimSpace(inn),
		KPP:           strings.TrimSpace(kpp),
		UniversityID:  university.ID,
	}
	
	// Если MAX_id получен, сохраняем время обновления (Requirements 3.4)
	if maxID != "" {
		now := time.Now()
		employee.MaxIDUpdatedAt = &now
	}
	
	if err := s.employeeRepo.Create(employee); err != nil {
		return nil, err
	}
	
	// Загружаем полную информацию о сотруднике с вузом
	return s.employeeRepo.GetByID(employee.ID)
}

// SearchEmployees выполняет поиск сотрудников
func (s *EmployeeService) SearchEmployees(query string, limit, offset int) ([]*domain.Employee, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	return s.employeeRepo.Search(query, limit, offset)
}

// GetAllEmployees получает всех сотрудников с пагинацией
func (s *EmployeeService) GetAllEmployees(limit, offset int) ([]*domain.Employee, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	return s.employeeRepo.GetAll(limit, offset)
}

// GetEmployeeByID получает сотрудника по ID
func (s *EmployeeService) GetEmployeeByID(id int64) (*domain.Employee, error) {
	employee, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return nil, domain.ErrEmployeeNotFound
	}
	return employee, nil
}

// UpdateEmployee обновляет данные сотрудника
func (s *EmployeeService) UpdateEmployee(employee *domain.Employee) error {
	// Проверяем существование сотрудника
	_, err := s.employeeRepo.GetByID(employee.ID)
	if err != nil {
		return domain.ErrEmployeeNotFound
	}
	
	// Если изменился телефон, обновляем MAX_id
	if employee.Phone != "" {
		if !s.maxService.ValidatePhone(employee.Phone) {
			return domain.ErrInvalidPhone
		}
		
		maxID, err := s.maxService.GetMaxIDByPhone(employee.Phone)
		if err != nil {
			// Логируем ошибку, но продолжаем без MAX_id
			maxID = ""
		}
		employee.MaxID = maxID
		
		// Если MAX_id получен, обновляем время
		if maxID != "" {
			now := time.Now()
			employee.MaxIDUpdatedAt = &now
		}
	}
	
	return s.employeeRepo.Update(employee)
}

// DeleteEmployee удаляет сотрудника
func (s *EmployeeService) DeleteEmployee(id int64) error {
	_, err := s.employeeRepo.GetByID(id)
	if err != nil {
		return domain.ErrEmployeeNotFound
	}
	
	return s.employeeRepo.Delete(id)
}

// findOrCreateUniversity находит существующий вуз или создает новый
func (s *EmployeeService) findOrCreateUniversity(inn, kpp, name string) (*domain.University, error) {
	if inn == "" {
		return nil, errors.New("INN is required")
	}
	
	var university *domain.University
	var err error
	
	// Пытаемся найти вуз по ИНН и КПП
	if kpp != "" {
		university, err = s.universityRepo.GetByINNAndKPP(inn, kpp)
		if err == nil && university != nil {
			return university, nil
		}
	}
	
	// Пытаемся найти вуз только по ИНН
	university, err = s.universityRepo.GetByINN(inn)
	if err == nil && university != nil {
		return university, nil
	}
	
	// Создаем новый вуз
	if name == "" {
		name = "Неизвестный вуз"
	}
	
	university = &domain.University{
		Name: strings.TrimSpace(name),
		INN:  strings.TrimSpace(inn),
		KPP:  strings.TrimSpace(kpp),
	}
	
	if err := s.universityRepo.Create(university); err != nil {
		return nil, err
	}
	
	return university, nil
}

// GetUniversityByID получает вуз по ID
func (s *EmployeeService) GetUniversityByID(id int64) (*domain.University, error) {
	return s.universityRepo.GetByID(id)
}

// GetUniversityByINN получает вуз по ИНН
func (s *EmployeeService) GetUniversityByINN(inn string) (*domain.University, error) {
	return s.universityRepo.GetByINN(inn)
}

// GetUniversityByINNAndKPP получает вуз по ИНН и КПП
func (s *EmployeeService) GetUniversityByINNAndKPP(inn, kpp string) (*domain.University, error) {
	return s.universityRepo.GetByINNAndKPP(inn, kpp)
}

