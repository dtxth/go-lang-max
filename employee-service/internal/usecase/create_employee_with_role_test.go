package usecase

import (
	"context"
	"testing"
)

// Test: CreateEmployeeWithRole creates new university automatically (Requirements 15.1)
func TestCreateEmployeeWithRole_CreatesNewUniversity(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_role_123"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
profileCache := newMockProfileCacheService()

	useCase := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService, profileCache,
	)

	ctx := context.Background()

	// Create employee with new university INN
	employee, err := useCase.Execute(
		ctx,
		"+79101234567",
		"Екатерина",
		"Морозова",
		"Ивановна",
		"2222222222",
		"222222222",
		"Казанский Федеральный Университет",
		"operator",
		"superadmin",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify employee was created
	if employee.ID == 0 {
		t.Error("Expected employee to have ID assigned")
	}

	// Verify university was created
	university, err := universityRepo.GetByINN("2222222222")
	if err != nil {
		t.Fatalf("Expected university to be created, got error: %v", err)
	}

	if university.Name != "Казанский Федеральный Университет" {
		t.Errorf("Expected university name 'Казанский Федеральный Университет', got '%s'", university.Name)
	}

	if university.INN != "2222222222" {
		t.Errorf("Expected university INN '2222222222', got '%s'", university.INN)
	}

	if university.KPP != "222222222" {
		t.Errorf("Expected university KPP '222222222', got '%s'", university.KPP)
	}

	// Verify employee is linked to the university
	if employee.UniversityID != university.ID {
		t.Errorf("Expected employee university_id %d, got %d", university.ID, employee.UniversityID)
	}

	// Verify role was assigned
	if employee.Role != "operator" {
		t.Errorf("Expected role 'operator', got '%s'", employee.Role)
	}
}

// Test: CreateEmployeeWithRole reuses existing university (Requirements 15.2)
func TestCreateEmployeeWithRole_ReusesExistingUniversity(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_role_456"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
profileCache := newMockProfileCacheService()

	useCase := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService, profileCache,
	)

	ctx := context.Background()

	// Create first employee (this will create the university)
	employee1, err := useCase.Execute(
		ctx,
		"+79201111111",
		"Николай",
		"Соколов",
		"",
		"3333333333",
		"333333333",
		"Уральский Федеральный Университет",
		"curator",
		"superadmin",
	)

	if err != nil {
		t.Fatalf("Expected no error creating first employee, got %v", err)
	}

	universityID1 := employee1.UniversityID

	// Create second employee with same INN (should reuse university)
	employee2, err := useCase.Execute(
		ctx,
		"+79202222222",
		"Светлана",
		"Павлова",
		"",
		"3333333333",
		"333333333",
		"УрФУ Другое Название",
		"operator",
		"curator",
	)

	if err != nil {
		t.Fatalf("Expected no error creating second employee, got %v", err)
	}

	// Verify both employees reference the same university
	if employee2.UniversityID != universityID1 {
		t.Errorf("Expected second employee to reuse university ID %d, got %d", universityID1, employee2.UniversityID)
	}

	// Verify only one university was created
	allUniversities, _ := universityRepo.GetAll()
	if len(allUniversities) != 1 {
		t.Errorf("Expected 1 university, got %d", len(allUniversities))
	}

	// Verify the university name is from the first creation
	university, _ := universityRepo.GetByINN("3333333333")
	if university.Name != "Уральский Федеральный Университет" {
		t.Errorf("Expected university name 'Уральский Федеральный Университет', got '%s'", university.Name)
	}
}

// Test: CreateEmployeeWithRole stores university data correctly (Requirements 15.3)
func TestCreateEmployeeWithRole_StoresUniversityData(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_role_789"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
profileCache := newMockProfileCacheService()

	useCase := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService, profileCache,
	)

	ctx := context.Background()

	employee, err := useCase.Execute(
		ctx,
		"+79303333333",
		"Андрей",
		"Лебедев",
		"Петрович",
		"4444444444",
		"444444444",
		"Новосибирский Государственный Университет",
		"operator",
		"superadmin",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Retrieve the university
	university, err := universityRepo.GetByID(employee.UniversityID)
	if err != nil {
		t.Fatalf("Expected to retrieve university, got error: %v", err)
	}

	// Verify all fields are stored
	if university.Name != "Новосибирский Государственный Университет" {
		t.Errorf("Expected university name 'Новосибирский Государственный Университет', got '%s'", university.Name)
	}

	if university.INN != "4444444444" {
		t.Errorf("Expected university INN '4444444444', got '%s'", university.INN)
	}

	if university.KPP != "444444444" {
		t.Errorf("Expected university KPP '444444444', got '%s'", university.KPP)
	}

	// Verify timestamps are set
	if university.CreatedAt.IsZero() {
		t.Error("Expected university CreatedAt to be set")
	}

	if university.UpdatedAt.IsZero() {
		t.Error("Expected university UpdatedAt to be set")
	}
}

// Test: CreateEmployeeWithRole returns complete record (Requirements 15.5)
func TestCreateEmployeeWithRole_ReturnsCompleteRecord(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_role_complete"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
profileCache := newMockProfileCacheService()

	useCase := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService, profileCache,
	)

	ctx := context.Background()

	employee, err := useCase.Execute(
		ctx,
		"+79404444444",
		"Татьяна",
		"Кузнецова",
		"Владимировна",
		"6666666666",
		"666666666",
		"Томский Политехнический Университет",
		"curator",
		"superadmin",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify employee has all required fields
	if employee.ID == 0 {
		t.Error("Expected employee ID to be set")
	}

	if employee.FirstName != "Татьяна" {
		t.Errorf("Expected first name 'Татьяна', got '%s'", employee.FirstName)
	}

	if employee.LastName != "Кузнецова" {
		t.Errorf("Expected last name 'Кузнецова', got '%s'", employee.LastName)
	}

	if employee.MiddleName != "Владимировна" {
		t.Errorf("Expected middle name 'Владимировна', got '%s'", employee.MiddleName)
	}

	if employee.Phone != "+79404444444" {
		t.Errorf("Expected phone '+79404444444', got '%s'", employee.Phone)
	}

	if employee.MaxID != "max_role_complete" {
		t.Errorf("Expected MAX_id 'max_role_complete', got '%s'", employee.MaxID)
	}

	if employee.INN != "6666666666" {
		t.Errorf("Expected INN '6666666666', got '%s'", employee.INN)
	}

	if employee.KPP != "666666666" {
		t.Errorf("Expected KPP '666666666', got '%s'", employee.KPP)
	}

	if employee.UniversityID == 0 {
		t.Error("Expected university_id to be set")
	}

	if employee.Role != "curator" {
		t.Errorf("Expected role 'curator', got '%s'", employee.Role)
	}

	if employee.UserID == nil {
		t.Error("Expected user_id to be set when role is assigned")
	}

	if employee.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if employee.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// Test: CreateEmployeeWithRole handles INN without KPP (Requirements 15.2)
func TestCreateEmployeeWithRole_HandlesINNWithoutKPP(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_role_no_kpp"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()
profileCache := newMockProfileCacheService()

	useCase := NewCreateEmployeeWithRoleUseCase(
		employeeRepo,
		universityRepo,
		maxService,
		authService,
		passwordGenerator,
		notificationService, profileCache,
	)

	ctx := context.Background()

	// Create employee with INN but no KPP
	employee, err := useCase.Execute(
		ctx,
		"+79505555555",
		"Владимир",
		"Орлов",
		"",
		"7777777777",
		"",
		"Дальневосточный Федеральный Университет",
		"operator",
		"superadmin",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify university was created
	university, err := universityRepo.GetByINN("7777777777")
	if err != nil {
		t.Fatalf("Expected university to be created, got error: %v", err)
	}

	if university.INN != "7777777777" {
		t.Errorf("Expected university INN '7777777777', got '%s'", university.INN)
	}

	if university.KPP != "" {
		t.Errorf("Expected university KPP to be empty, got '%s'", university.KPP)
	}

	// Verify employee is linked to the university
	if employee.UniversityID != university.ID {
		t.Errorf("Expected employee university_id %d, got %d", university.ID, employee.UniversityID)
	}
}
