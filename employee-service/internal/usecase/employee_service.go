package usecase

import (
	"context"
	"employee-service/internal/domain"
	"strings"
	"time"
)

type EmployeeService struct {
	employeeRepo        domain.EmployeeRepository
	universityRepo      domain.UniversityRepository
	maxService          domain.MaxService
	authService         domain.AuthService
	passwordGenerator   domain.PasswordGenerator
	notificationService domain.NotificationService
}

func NewEmployeeService(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
	authService domain.AuthService,
	passwordGenerator domain.PasswordGenerator,
	notificationService domain.NotificationService,
) *EmployeeService {
	return &EmployeeService{
		employeeRepo:        employeeRepo,
		universityRepo:      universityRepo,
		maxService:          maxService,
		authService:         authService,
		passwordGenerator:   passwordGenerator,
		notificationService: notificationService,
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
	
	// Устанавливаем значения по умолчанию для обязательных полей
	if strings.TrimSpace(firstName) == "" {
		firstName = "Неизвестно"
	}
	if strings.TrimSpace(lastName) == "" {
		lastName = "Неизвестно"
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

// GetAllEmployeesWithSortingAndSearch получает всех сотрудников с пагинацией, сортировкой и поиском
func (s *EmployeeService) GetAllEmployeesWithSortingAndSearch(limit, offset int, sortBy, sortOrder, search string) ([]*domain.Employee, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	// Получаем сотрудников
	employees, err := s.employeeRepo.GetAllWithSortingAndSearch(limit, offset, sortBy, sortOrder, search)
	if err != nil {
		return nil, 0, err
	}
	
	// Получаем общее количество для пагинации
	total, err := s.employeeRepo.CountAllWithSearch(search)
	if err != nil {
		return nil, 0, err
	}
	
	return employees, total, nil
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
	var university *domain.University
	var err error
	
	// Если есть ИНН, пытаемся найти вуз по ИНН и КПП
	if inn != "" {
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
	} else {
		// Если ИНН не указан, ищем существующий университет с пустым ИНН
		if name == "" {
			name = "Неизвестный вуз"
		}
		
		// Пытаемся найти любой университет с пустым ИНН и таким же именем
		universities, err := s.universityRepo.GetAll()
		if err == nil {
			for _, u := range universities {
				if u.INN == "" && u.Name == name {
					return u, nil
				}
			}
		}
	}
	
	// Если вуз не найден, создаем новый
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

// CreateEmployeeWithRole создает сотрудника с назначением роли
func (s *EmployeeService) CreateEmployeeWithRole(
	ctx context.Context,
	phone string,
	firstName, lastName, middleName string,
	inn, kpp string,
	universityName string,
	role string,
	requesterRole string,
) (*domain.Employee, error) {
	// Используем CreateEmployeeWithRoleUseCase
	uc := NewCreateEmployeeWithRoleUseCase(
		s.employeeRepo,
		s.universityRepo,
		s.maxService,
		s.authService,
		s.passwordGenerator,
		s.notificationService,
	)
	
	return uc.Execute(
		ctx,
		phone,
		firstName, lastName, middleName,
		inn, kpp,
		universityName,
		role,
		requesterRole,
	)
}

