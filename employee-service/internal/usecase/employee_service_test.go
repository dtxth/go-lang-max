package usecase

import (
	"errors"
	"testing"
)

// mockMaxServiceForEmployeeTest is a specific mock for employee service tests
type mockMaxServiceForEmployeeTest struct {
	shouldFail bool
	maxID      string
}

func (m *mockMaxServiceForEmployeeTest) GetMaxIDByPhone(phone string) (string, error) {
	if m.shouldFail {
		return "", errors.New("MAX API unavailable")
	}
	if m.maxID != "" {
		return m.maxID, nil
	}
	return "max_" + phone, nil
}

func (m *mockMaxServiceForEmployeeTest) ValidatePhone(phone string) bool {
	return len(phone) > 5
}

func (m *mockMaxServiceForEmployeeTest) BatchGetMaxIDByPhone(phones []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, phone := range phones {
		if !m.shouldFail {
			if m.maxID != "" {
				result[phone] = m.maxID
			} else {
				result[phone] = "max_" + phone
			}
		}
	}
	return result, nil
}

// Test: Employee creation triggers MAX_id lookup (Requirements 3.1)
func TestAddEmployeeByPhone_TriggersMaxIDLookup(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_123456"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Иван",
		"Иванов",
		"Иванович",
		"1234567890",
		"123456789",
		"МГУ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "max_123456" {
		t.Errorf("Expected MAX_id to be 'max_123456', got '%s'", employee.MaxID)
	}
}

// Test: MAX_id is stored when received (Requirements 3.4)
func TestAddEmployeeByPhone_StoresMaxIDWhenReceived(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_987654"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Петр",
		"Петров",
		"",
		"9876543210",
		"987654321",
		"СПбГУ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "max_987654" {
		t.Errorf("Expected MAX_id to be 'max_987654', got '%s'", employee.MaxID)
	}

	if employee.MaxIDUpdatedAt == nil {
		t.Error("Expected MaxIDUpdatedAt to be set, got nil")
	}
}

// Test: Employee creation succeeds without MAX_id (Requirements 3.5)
func TestAddEmployeeByPhone_SucceedsWithoutMaxID(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{shouldFail: true}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Сергей",
		"Сергеев",
		"Сергеевич",
		"5555555555",
		"555555555",
		"МФТИ",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if employee.MaxID != "" {
		t.Errorf("Expected MAX_id to be empty, got '%s'", employee.MaxID)
	}

	if employee.MaxIDUpdatedAt != nil {
		t.Error("Expected MaxIDUpdatedAt to be nil when MAX_id is not set")
	}

	if employee.Phone != "+79001234567" {
		t.Errorf("Expected phone to be '+79001234567', got '%s'", employee.Phone)
	}
}

// Test: New universities are created automatically (Requirements 15.1)
func TestAddEmployeeByPhone_CreatesNewUniversity(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_123"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	// Create employee with new university INN
	employee, err := service.AddEmployeeByPhone(
		"+79001234567",
		"Анна",
		"Смирнова",
		"",
		"7707083893",
		"770701001",
		"Московский Государственный Университет",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify employee was created
	if employee.ID == 0 {
		t.Error("Expected employee to have ID assigned")
	}

	// Verify university was created
	university, err := universityRepo.GetByINN("7707083893")
	if err != nil {
		t.Fatalf("Expected university to be created, got error: %v", err)
	}

	if university.Name != "Московский Государственный Университет" {
		t.Errorf("Expected university name 'Московский Государственный Университет', got '%s'", university.Name)
	}

	if university.INN != "7707083893" {
		t.Errorf("Expected university INN '7707083893', got '%s'", university.INN)
	}

	if university.KPP != "770701001" {
		t.Errorf("Expected university KPP '770701001', got '%s'", university.KPP)
	}

	// Verify employee is linked to the university
	if employee.UniversityID != university.ID {
		t.Errorf("Expected employee university_id %d, got %d", university.ID, employee.UniversityID)
	}
}

// Test: Existing universities are reused (Requirements 15.2)
func TestAddEmployeeByPhone_ReusesExistingUniversity(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_456"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	// Create first employee (this will create the university)
	employee1, err := service.AddEmployeeByPhone(
		"+79001111111",
		"Иван",
		"Иванов",
		"",
		"1234567890",
		"123456789",
		"СПбГУ",
	)

	if err != nil {
		t.Fatalf("Expected no error creating first employee, got %v", err)
	}

	universityID1 := employee1.UniversityID

	// Create second employee with same INN (should reuse university)
	employee2, err := service.AddEmployeeByPhone(
		"+79002222222",
		"Петр",
		"Петров",
		"",
		"1234567890",
		"123456789",
		"СПбГУ Повторно",
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
	university, _ := universityRepo.GetByINN("1234567890")
	if university.Name != "СПбГУ" {
		t.Errorf("Expected university name 'СПбГУ', got '%s'", university.Name)
	}
}

// Test: University stores name, INN, and KPP (Requirements 15.3)
func TestAddEmployeeByPhone_StoresUniversityData(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_789"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	employee, err := service.AddEmployeeByPhone(
		"+79003333333",
		"Мария",
		"Сидорова",
		"Александровна",
		"9876543210",
		"987654321",
		"МФТИ",
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
	if university.Name != "МФТИ" {
		t.Errorf("Expected university name 'МФТИ', got '%s'", university.Name)
	}

	if university.INN != "9876543210" {
		t.Errorf("Expected university INN '9876543210', got '%s'", university.INN)
	}

	if university.KPP != "987654321" {
		t.Errorf("Expected university KPP '987654321', got '%s'", university.KPP)
	}

	// Verify timestamps are set
	if university.CreatedAt.IsZero() {
		t.Error("Expected university CreatedAt to be set")
	}

	if university.UpdatedAt.IsZero() {
		t.Error("Expected university UpdatedAt to be set")
	}
}

// Test: University reuse by INN+KPP combination (Requirements 15.2)
func TestAddEmployeeByPhone_ReusesUniversityByINNAndKPP(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_999"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	// Create first employee with INN and KPP
	employee1, err := service.AddEmployeeByPhone(
		"+79004444444",
		"Алексей",
		"Козлов",
		"",
		"5555555555",
		"555555555",
		"ВШЭ",
	)

	if err != nil {
		t.Fatalf("Expected no error creating first employee, got %v", err)
	}

	// Create second employee with same INN and KPP
	employee2, err := service.AddEmployeeByPhone(
		"+79005555555",
		"Ольга",
		"Новикова",
		"",
		"5555555555",
		"555555555",
		"ВШЭ Филиал",
	)

	if err != nil {
		t.Fatalf("Expected no error creating second employee, got %v", err)
	}

	// Verify both employees reference the same university
	if employee1.UniversityID != employee2.UniversityID {
		t.Errorf("Expected employees to share university ID, got %d and %d", employee1.UniversityID, employee2.UniversityID)
	}

	// Verify only one university exists
	allUniversities, _ := universityRepo.GetAll()
	if len(allUniversities) != 1 {
		t.Errorf("Expected 1 university, got %d", len(allUniversities))
	}
}

// Test: Employee creation returns complete record with university details (Requirements 15.5)
func TestAddEmployeeByPhone_ReturnsCompleteRecord(t *testing.T) {
	employeeRepo := newMockEmployeeRepo()
	universityRepo := newMockUniversityRepo()
	maxService := &mockMaxServiceForEmployeeTest{maxID: "max_complete"}
	authService := newMockAuthService()
	passwordGenerator := newMockPasswordGenerator()
	notificationService := newMockNotificationService()

	service := NewEmployeeService(employeeRepo, universityRepo, maxService, authService, passwordGenerator, notificationService)

	employee, err := service.AddEmployeeByPhone(
		"+79006666666",
		"Дмитрий",
		"Волков",
		"Сергеевич",
		"1111111111",
		"111111111",
		"ИТМО",
	)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify employee has all required fields
	if employee.ID == 0 {
		t.Error("Expected employee ID to be set")
	}

	if employee.FirstName != "Дмитрий" {
		t.Errorf("Expected first name 'Дмитрий', got '%s'", employee.FirstName)
	}

	if employee.LastName != "Волков" {
		t.Errorf("Expected last name 'Волков', got '%s'", employee.LastName)
	}

	if employee.MiddleName != "Сергеевич" {
		t.Errorf("Expected middle name 'Сергеевич', got '%s'", employee.MiddleName)
	}

	if employee.Phone != "+79006666666" {
		t.Errorf("Expected phone '+79006666666', got '%s'", employee.Phone)
	}

	if employee.MaxID != "max_complete" {
		t.Errorf("Expected MAX_id 'max_complete', got '%s'", employee.MaxID)
	}

	if employee.INN != "1111111111" {
		t.Errorf("Expected INN '1111111111', got '%s'", employee.INN)
	}

	if employee.KPP != "111111111" {
		t.Errorf("Expected KPP '111111111', got '%s'", employee.KPP)
	}

	if employee.UniversityID == 0 {
		t.Error("Expected university_id to be set")
	}

	if employee.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}

	if employee.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}
