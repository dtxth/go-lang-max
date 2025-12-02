package usecase

import (
	"context"
	"employee-service/internal/domain"
	"errors"
	"strings"
	"time"
)

// CreateEmployeeWithRoleUseCase создает сотрудника с назначением роли
type CreateEmployeeWithRoleUseCase struct {
	employeeRepo   domain.EmployeeRepository
	universityRepo domain.UniversityRepository
	maxService     domain.MaxService
	authService    domain.AuthService
}

// NewCreateEmployeeWithRoleUseCase создает новый use case
func NewCreateEmployeeWithRoleUseCase(
	employeeRepo domain.EmployeeRepository,
	universityRepo domain.UniversityRepository,
	maxService domain.MaxService,
	authService domain.AuthService,
) *CreateEmployeeWithRoleUseCase {
	return &CreateEmployeeWithRoleUseCase{
		employeeRepo:   employeeRepo,
		universityRepo: universityRepo,
		maxService:     maxService,
		authService:    authService,
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
	if !uc.maxService.ValidatePhone(phone) {
		return nil, domain.ErrInvalidPhone
	}

	// Проверяем, не существует ли уже сотрудник с таким телефоном
	existing, _ := uc.employeeRepo.GetByPhone(phone)
	if existing != nil {
		return nil, domain.ErrEmployeeExists
	}

	// Получаем MAX_id по телефону
	maxID, err := uc.maxService.GetMaxIDByPhone(phone)
	if err != nil {
		// Логируем ошибку, но продолжаем без MAX_id
		maxID = ""
	}

	// Находим или создаем вуз
	university, err := uc.findOrCreateUniversity(inn, kpp, universityName)
	if err != nil {
		return nil, err
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

	// Если есть роль, создаем пользователя в Auth Service
	if role != "" {
		// Для простоты используем ID сотрудника как user_id
		// В реальной системе нужно создать пользователя через Auth Service
		// и получить user_id
		// Пока оставим user_id пустым, он будет заполнен после создания
	}

	if err := uc.employeeRepo.Create(employee); err != nil {
		return nil, err
	}

	// Если есть роль, назначаем её через Auth Service
	if role != "" {
		// Используем ID сотрудника как user_id
		userID := employee.ID
		employee.UserID = &userID

		// Назначаем роль в Auth Service
		err := uc.authService.AssignRole(ctx, userID, role, &university.ID, nil, nil)
		if err != nil {
			// Откатываем создание сотрудника
			_ = uc.employeeRepo.Delete(employee.ID)
			return nil, errors.New("failed to assign role in auth service: " + err.Error())
		}

		// Обновляем сотрудника с user_id
		if err := uc.employeeRepo.Update(employee); err != nil {
			// Пытаемся отозвать роль
			_ = uc.authService.RevokeUserRoles(ctx, userID)
			_ = uc.employeeRepo.Delete(employee.ID)
			return nil, err
		}
	}

	// Загружаем полную информацию о сотруднике с вузом
	return uc.employeeRepo.GetByID(employee.ID)
}

// findOrCreateUniversity находит существующий вуз или создает новый
func (uc *CreateEmployeeWithRoleUseCase) findOrCreateUniversity(inn, kpp, name string) (*domain.University, error) {
	if inn == "" {
		return nil, errors.New("INN is required")
	}

	var university *domain.University
	var err error

	// Пытаемся найти вуз по ИНН и КПП
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

	// Создаем новый вуз
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
