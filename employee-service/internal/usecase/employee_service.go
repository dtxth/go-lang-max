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
	profileCache        domain.ProfileCacheService
}

func NewEmployeeService(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
	authService domain.AuthService,
	passwordGenerator domain.PasswordGenerator,
	notificationService domain.NotificationService,
	profileCache domain.ProfileCacheService,
) *EmployeeService {
	return &EmployeeService{
		employeeRepo:        employeeRepo,
		universityRepo:      universityRepo,
		maxService:          maxService,
		authService:         authService,
		passwordGenerator:   passwordGenerator,
		notificationService: notificationService,
		profileCache:        profileCache,
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
	
	// Получаем профиль пользователя по телефону (Requirements 3.1)
	// Это включает MAX_id, first_name и last_name
	var maxID string
	var profileFirstName, profileLastName string
	var profileSource domain.ProfileSource = domain.SourceDefault
	
	// Сначала пытаемся получить MAX_id через MAX API
	profile, err := s.maxService.GetUserProfileByPhone(phone)
	if err != nil {
		// Логируем ошибку, но продолжаем без MAX_id (Requirements 3.5, 7.5)
		maxID = ""
	} else {
		maxID = profile.MaxID
	}
	
	// Если у нас есть MAX_id, пытаемся получить кэшированный профиль (Requirements 3.4, 7.2)
	if maxID != "" {
		cachedProfile, _ := s.safeGetProfileFromCache(context.Background(), maxID)
		if cachedProfile != nil {
			// Используем данные из кэша с приоритетом (Requirements 2.3, 5.3)
			displayFirstName, displayLastName := cachedProfile.GetDisplayName()
			if displayFirstName != "" || displayLastName != "" {
				profileFirstName = displayFirstName
				profileLastName = displayLastName
				profileSource = cachedProfile.GetPrioritySource()
			}
		}
	}
	
	// Fallback: если кэш недоступен или пуст, используем данные из MAX API
	if profileFirstName == "" && profileLastName == "" && profile != nil {
		profileFirstName = profile.FirstName
		profileLastName = profile.LastName
		if profileFirstName != "" || profileLastName != "" {
			profileSource = domain.SourceWebhook
		}
	}
	
	// Находим или создаем вуз
	university, err := s.findOrCreateUniversity(inn, kpp, universityName)
	if err != nil {
		return nil, err
	}
	
	// Реализуем приоритетную логику имен (Requirements 2.3, 5.3, 7.1, 7.2)
	// Приоритет: user_provided (переданные параметры) > cached profile > max profile > default
	finalFirstName := strings.TrimSpace(firstName)
	finalLastName := strings.TrimSpace(lastName)
	finalSource := profileSource
	
	// Проверяем, были ли имена предоставлены явно (Requirements 7.1)
	userProvidedFirstName := finalFirstName != ""
	userProvidedLastName := finalLastName != ""
	
	// Если имена не переданы явно, используем данные из профиля (Requirements 7.2, 7.3)
	if !userProvidedFirstName && profileFirstName != "" {
		finalFirstName = profileFirstName
	}
	if !userProvidedLastName && profileLastName != "" {
		finalLastName = profileLastName
	}
	
	// Если имена переданы явно, это user_provided источник (Requirements 7.1)
	if userProvidedFirstName || userProvidedLastName {
		finalSource = domain.SourceUserInput
	}
	
	// Устанавливаем значения по умолчанию для обязательных полей (Requirements 7.3, 7.5)
	if finalFirstName == "" {
		finalFirstName = "Неизвестно"
		finalSource = domain.SourceDefault
	}
	if finalLastName == "" {
		finalLastName = "Неизвестно"
		finalSource = domain.SourceDefault
	}
	
	// Создаем сотрудника
	employee := &domain.Employee{
		FirstName:        finalFirstName,
		LastName:         finalLastName,
		MiddleName:       strings.TrimSpace(middleName),
		Phone:            phone,
		MaxID:            maxID,
		INN:              strings.TrimSpace(inn),
		KPP:              strings.TrimSpace(kpp),
		ProfileSource:    string(finalSource),
		UniversityID:     university.ID,
	}
	
	// Если MAX_id получен, сохраняем время обновления (Requirements 3.4)
	if maxID != "" {
		now := time.Now()
		employee.MaxIDUpdatedAt = &now
	}
	
	// Сохраняем время обновления профиля (Requirements 5.5)
	if finalSource != domain.SourceDefault {
		now := time.Now()
		employee.ProfileLastUpdated = &now
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

// safeGetProfileFromCache безопасно получает профиль из кэша с обработкой ошибок
func (s *EmployeeService) safeGetProfileFromCache(ctx context.Context, userID string) (*domain.CachedUserProfile, error) {
	if s.profileCache == nil {
		return nil, nil
	}
	
	// Устанавливаем короткий таймаут для кэша (Requirements 3.4)
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	profile, err := s.profileCache.GetProfile(ctx, userID)
	if err != nil {
		// Логируем ошибку, но не возвращаем её для graceful degradation (Requirements 7.5)
		return nil, nil
	}
	
	return profile, nil
}

// isProfileCacheHealthy проверяет доступность кэша профилей
func (s *EmployeeService) isProfileCacheHealthy(ctx context.Context) bool {
	if s.profileCache == nil {
		return false
	}
	
	// Простая проверка через попытку получения несуществующего профиля
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	
	_, err := s.profileCache.GetProfile(ctx, "health-check")
	return err == nil
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
		s.profileCache,
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

