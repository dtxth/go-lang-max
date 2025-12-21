package usecase

import (
	"context"
	"employee-service/internal/domain"
	"employee-service/internal/utils"
	"errors"
	"log"
	"strings"
	"time"
)

// CreateEmployeeWithRoleUseCase создает сотрудника с назначением роли
type CreateEmployeeWithRoleUseCase struct {
	employeeRepo        domain.EmployeeRepository
	universityRepo      domain.UniversityRepository
	maxService          domain.MaxService
	authService         domain.AuthService
	passwordGenerator   domain.PasswordGenerator
	notificationService domain.NotificationService
	profileCache        domain.ProfileCacheService
	phoneValidator      *utils.PhoneValidator
}

// NewCreateEmployeeWithRoleUseCase создает новый use case
func NewCreateEmployeeWithRoleUseCase(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
	authService domain.AuthService,
	passwordGenerator domain.PasswordGenerator,
	notificationService domain.NotificationService,
	profileCache domain.ProfileCacheService,
) *CreateEmployeeWithRoleUseCase {
	return &CreateEmployeeWithRoleUseCase{
		employeeRepo:        employeeRepo,
		universityRepo:      universityRepo,
		maxService:          maxService,
		authService:         authService,
		passwordGenerator:   passwordGenerator,
		notificationService: notificationService,
		profileCache:        profileCache,
		phoneValidator:      utils.NewPhoneValidator(),
	}
}

// Execute выполняет создание сотрудника с ролью
func (uc *CreateEmployeeWithRoleUseCase) Execute(
	ctx context.Context,
	phone string,
	firstName, lastName, middleName string,
	inn, kpp string,
	universityName string,
	role string,
	requesterRole string,
) (*domain.Employee, error) {
	// Валидация роли
	if role != "" && role != "curator" && role != "operator" {
		return nil, errors.New("invalid role: must be 'curator' or 'operator'")
	}

	// Проверка прав: Curator может назначать только Operator
	if requesterRole == "curator" && role == "curator" {
		return nil, errors.New("curator cannot assign curator role")
	}

	// Валидация телефона
	if !uc.phoneValidator.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Нормализуем телефон к стандартному формату
	phone = uc.phoneValidator.NormalizePhone(phone)

	// Проверяем, не существует ли уже сотрудник с таким телефоном
	existing, _ := uc.employeeRepo.GetByPhone(phone)
	if existing != nil {
		return nil, domain.ErrEmployeeExists
	}

	// Получаем профиль пользователя по телефону
	// Это включает MAX_id, first_name и last_name
	var maxID string
	var profileFirstName, profileLastName string
	
	profile, err := uc.maxService.GetUserProfileByPhone(phone)
	if err != nil {
		// Логируем ошибку, но продолжаем без MAX_id и имен
		maxID = ""
		profileFirstName = ""
		profileLastName = ""
	} else {
		maxID = profile.MaxID
		profileFirstName = profile.FirstName
		profileLastName = profile.LastName
	}

	// Находим или создаем вуз
	university, err := uc.findOrCreateUniversity(inn, kpp, universityName)
	if err != nil {
		return nil, err
	}

	// Используем данные из MAX профиля, если они доступны
	// Если переданы пустые значения, используем данные из профиля
	if strings.TrimSpace(firstName) == "" && profileFirstName != "" {
		firstName = profileFirstName
	}
	if strings.TrimSpace(lastName) == "" && profileLastName != "" {
		lastName = profileLastName
	}
	
	// Устанавливаем значения по умолчанию для обязательных полей
	if strings.TrimSpace(firstName) == "" {
		firstName = "Неизвестно"
	}
	if strings.TrimSpace(lastName) == "" {
		lastName = "Неизвестно"
	}

	// Создаем сотрудника
	now := time.Now()
	employee := &domain.Employee{
		FirstName:    strings.TrimSpace(firstName),
		LastName:     strings.TrimSpace(lastName),
		MiddleName:   strings.TrimSpace(middleName),
		Phone:        phone,
		MaxID:        maxID,
		INN:          strings.TrimSpace(inn),
		KPP:          strings.TrimSpace(kpp),
		UniversityID: university.ID,
		Role:         role,
	}

	if maxID != "" {
		employee.MaxIDUpdatedAt = &now
	}

	// Если есть роль, сначала создаем пользователя в Auth Service
	if role != "" {
		log.Printf("Creating user with role %s for phone ending in %s", role, sanitizePhone(phone))
		
		// Проверяем, что authService не nil
		if uc.authService == nil {
			return nil, errors.New("auth service is not available")
		}
		
		// Генерируем криптографически безопасный случайный пароль
		password, err := uc.passwordGenerator.Generate(12)
		if err != nil {
			return nil, errors.New("failed to generate password: " + err.Error())
		}
		

		
		// Создаем пользователя в Auth Service (используем телефон как идентификатор)
		userID, err := uc.authService.CreateUser(ctx, phone, password)
		if err != nil {
			return nil, errors.New("failed to create user in auth service: " + err.Error())
		}
		employee.UserID = &userID
		
		// Логируем сгенерированный пароль для администраторов
		log.Printf("Generated password for new employee with phone ending in %s: %s", 
			sanitizePhone(phone), password)
		
		// Примечание: Отправка пароля через MAX Messenger отключена
		// Пароль логируется выше для администраторов
	}

	// Создаем сотрудника в базе
	if err := uc.employeeRepo.Create(employee); err != nil {
		// Если есть роль и создание сотрудника не удалось, откатываем создание пользователя
		if role != "" && employee.UserID != nil {
			_ = uc.authService.RevokeUserRoles(ctx, *employee.UserID)
		}
		return nil, err
	}

	// Если есть роль, назначаем её через Auth Service
	if role != "" && employee.UserID != nil {
		// Назначаем роль в Auth Service
		err := uc.authService.AssignRole(ctx, *employee.UserID, role, &university.ID, nil, nil)
		if err != nil {
			// Откатываем создание сотрудника и пользователя
			_ = uc.employeeRepo.Delete(employee.ID)
			_ = uc.authService.RevokeUserRoles(ctx, *employee.UserID)
			return nil, errors.New("failed to assign role in auth service: " + err.Error())
		}
	}

	// Загружаем полную информацию о сотруднике с вузом
	return uc.employeeRepo.GetByID(employee.ID)
}

// sanitizePhone returns only the last 4 digits of a phone number for logging
func sanitizePhone(phone string) string {
	if len(phone) < 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}

// findOrCreateUniversity находит существующий вуз или создает новый
func (uc *CreateEmployeeWithRoleUseCase) findOrCreateUniversity(inn, kpp, name string) (*domain.University, error) {
	var university *domain.University
	var err error

	// Если есть ИНН, пытаемся найти вуз по ИНН и КПП
	if inn != "" {
		if kpp != "" {
			university, err = uc.universityRepo.GetByINNAndKPP(inn, kpp)
			if err == nil && university != nil {
				return university, nil
			}
		}

		// Пытаемся найти вуз только по ИНН
		university, err = uc.universityRepo.GetByINN(inn)
		if err == nil && university != nil {
			return university, nil
		}
	}

	// Если вуз не найден или ИНН не указан, создаем новый вуз
	if name == "" {
		name = "Неизвестный вуз"
	}

	university = &domain.University{
		Name: strings.TrimSpace(name),
		INN:  strings.TrimSpace(inn),
		KPP:  strings.TrimSpace(kpp),
	}

	if err := uc.universityRepo.Create(university); err != nil {
		return nil, err
	}

	return university, nil
}
