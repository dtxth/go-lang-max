# MaxBot Integration Examples for Employee Service

Примеры расширенного использования MaxBot Service в Employee Service.

## Дополнительные методы для MaxClient

Добавьте следующие методы в `internal/infrastructure/max/max_client.go`:

```go
// SendNotification отправляет уведомление сотруднику
func (c *MaxClient) SendNotification(phone, text string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.SendNotification(ctx, &maxbotproto.SendNotificationRequest{
		Phone: phone,
		Text:  text,
	})
	if err != nil {
		return err
	}

	if resp.Error != "" {
		return mapError(resp.ErrorCode, resp.Error)
	}

	if !resp.Success {
		return errors.New("failed to send notification")
	}

	return nil
}

// CheckPhoneNumbers проверяет существование номеров телефонов в Max Messenger
func (c *MaxClient) CheckPhoneNumbers(phones []string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	resp, err := c.client.CheckPhoneNumbers(ctx, &maxbotproto.CheckPhoneNumbersRequest{
		Phones: phones,
	})
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, mapError(resp.ErrorCode, resp.Error)
	}

	return resp.ExistingPhones, nil
}
```

## Расширение Domain интерфейса

Обновите `internal/domain/max_service.go`:

```go
package domain

// MaxService определяет интерфейс для работы с MAX API
type MaxService interface {
	// Существующие методы
	GetMaxIDByPhone(phone string) (string, error)
	ValidatePhone(phone string) bool
	
	// Новые методы
	SendNotification(phone, text string) error
	CheckPhoneNumbers(phones []string) ([]string, error)
}
```


## Use Case: Уведомления сотрудников

Добавьте в `internal/usecase/employee_service.go`:

```go
// NotifyEmployee отправляет уведомление сотруднику
func (s *EmployeeService) NotifyEmployee(employeeID int64, message string) error {
	employee, err := s.employeeRepo.GetByID(employeeID)
	if err != nil {
		return domain.ErrEmployeeNotFound
	}

	if employee.Phone == "" {
		return fmt.Errorf("employee has no phone number")
	}

	if err := s.maxService.SendNotification(employee.Phone, message); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

// NotifyUniversityEmployees отправляет уведомление всем сотрудникам вуза
func (s *EmployeeService) NotifyUniversityEmployees(universityID int64, message string) error {
	// Получаем всех сотрудников вуза
	employees, err := s.employeeRepo.GetByUniversityID(universityID)
	if err != nil {
		return fmt.Errorf("failed to get employees: %w", err)
	}

	if len(employees) == 0 {
		return fmt.Errorf("no employees found for university")
	}

	successCount := 0
	errorCount := 0

	// Отправляем уведомления асинхронно
	for _, employee := range employees {
		if employee.Phone == "" {
			continue
		}

		go func(phone, msg string) {
			if err := s.maxService.SendNotification(phone, msg); err != nil {
				log.Printf("Failed to notify employee %s: %v", maskPhone(phone), err)
			}
		}(employee.Phone, message)

		successCount++
	}

	if successCount == 0 {
		return fmt.Errorf("no employees with phone numbers found")
	}

	log.Printf("Sent notifications to %d employees", successCount)
	return nil
}

func maskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	return "****" + phone[len(phone)-4:]
}
```

## Use Case: Пакетная проверка номеров

```go
// ValidateEmployeePhones проверяет, какие сотрудники зарегистрированы в Max Messenger
func (s *EmployeeService) ValidateEmployeePhones(universityID int64) (*PhoneValidationResult, error) {
	// Получаем всех сотрудников вуза
	employees, err := s.employeeRepo.GetByUniversityID(universityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employees: %w", err)
	}

	if len(employees) == 0 {
		return &PhoneValidationResult{}, nil
	}

	// Собираем все номера телефонов
	phones := make([]string, 0, len(employees))
	phoneToEmployee := make(map[string]*domain.Employee)
	
	for _, employee := range employees {
		if employee.Phone != "" {
			phones = append(phones, employee.Phone)
			phoneToEmployee[employee.Phone] = employee
		}
	}

	// Проверяем номера через MaxBot Service
	existingPhones, err := s.maxService.CheckPhoneNumbers(phones)
	if err != nil {
		return nil, fmt.Errorf("failed to check phones: %w", err)
	}

	// Формируем результат
	result := &PhoneValidationResult{
		Total:    len(phones),
		Existing: len(existingPhones),
		Missing:  len(phones) - len(existingPhones),
	}

	existingMap := make(map[string]bool)
	for _, phone := range existingPhones {
		existingMap[phone] = true
	}

	for phone, employee := range phoneToEmployee {
		if existingMap[phone] {
			result.ExistingEmployees = append(result.ExistingEmployees, employee)
		} else {
			result.MissingEmployees = append(result.MissingEmployees, employee)
		}
	}

	return result, nil
}

type PhoneValidationResult struct {
	Total             int
	Existing          int
	Missing           int
	ExistingEmployees []*domain.Employee
	MissingEmployees  []*domain.Employee
}
```


## HTTP Handler: Новые эндпоинты

Добавьте в `internal/infrastructure/http/handler.go`:

```go
// NotifyEmployee отправляет уведомление сотруднику
// @Summary Notify employee
// @Tags employees
// @Param id path int true "Employee ID"
// @Param request body NotifyRequest true "Notification message"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /employees/{id}/notify [post]
func (h *Handler) NotifyEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID, err := getEmployeeIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "Message is required")
		return
	}

	if err := h.service.NotifyEmployee(employeeID, req.Message); err != nil {
		if errors.Is(err, domain.ErrEmployeeNotFound) {
			respondError(w, http.StatusNotFound, "Employee not found")
			return
		}
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Notification sent successfully",
	})
}

// NotifyUniversityEmployees отправляет уведомление всем сотрудникам вуза
// @Summary Notify all university employees
// @Tags universities
// @Param id path int true "University ID"
// @Param request body NotifyRequest true "Notification message"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /universities/{id}/notify [post]
func (h *Handler) NotifyUniversityEmployees(w http.ResponseWriter, r *http.Request) {
	universityID, err := getUniversityIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid university ID")
		return
	}

	var req NotifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Message == "" {
		respondError(w, http.StatusBadRequest, "Message is required")
		return
	}

	if err := h.service.NotifyUniversityEmployees(universityID, req.Message); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{
		"message": "Notifications sent successfully",
	})
}

// ValidateUniversityPhones проверяет номера телефонов сотрудников вуза
// @Summary Validate university employee phones
// @Tags universities
// @Param id path int true "University ID"
// @Success 200 {object} PhoneValidationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /universities/{id}/validate-phones [get]
func (h *Handler) ValidateUniversityPhones(w http.ResponseWriter, r *http.Request) {
	universityID, err := getUniversityIDFromPath(r)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid university ID")
		return
	}

	result, err := h.service.ValidateEmployeePhones(universityID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, PhoneValidationResponse{
		Total:             result.Total,
		Existing:          result.Existing,
		Missing:           result.Missing,
		ExistingEmployees: result.ExistingEmployees,
		MissingEmployees:  result.MissingEmployees,
	})
}

type NotifyRequest struct {
	Message string `json:"message"`
}

type PhoneValidationResponse struct {
	Total             int                `json:"total"`
	Existing          int                `json:"existing"`
	Missing           int                `json:"missing"`
	ExistingEmployees []*domain.Employee `json:"existing_employees"`
	MissingEmployees  []*domain.Employee `json:"missing_employees"`
}
```

## Обновление роутера

Добавьте новые маршруты в `internal/infrastructure/http/router.go`:

```go
// Уведомления
r.Post("/employees/{id}/notify", h.NotifyEmployee)
r.Post("/universities/{id}/notify", h.NotifyUniversityEmployees)

// Валидация номеров
r.Get("/universities/{id}/validate-phones", h.ValidateUniversityPhones)
```


## Примеры использования API

### Отправка уведомления сотруднику

```bash
curl -X POST http://localhost:8081/employees/123/notify \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Важное уведомление для сотрудника"
  }'
```

Ответ:
```json
{
  "message": "Notification sent successfully"
}
```

### Отправка уведомления всем сотрудникам вуза

```bash
curl -X POST http://localhost:8081/universities/5/notify \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Общее уведомление для всех сотрудников"
  }'
```

Ответ:
```json
{
  "message": "Notifications sent successfully"
}
```

### Проверка номеров телефонов сотрудников

```bash
curl http://localhost:8081/universities/5/validate-phones
```

Ответ:
```json
{
  "total": 50,
  "existing": 45,
  "missing": 5,
  "existing_employees": [
    {
      "id": 1,
      "first_name": "Иван",
      "last_name": "Иванов",
      "phone": "+79991234567",
      "max_id": "79991234567"
    }
  ],
  "missing_employees": [
    {
      "id": 2,
      "first_name": "Петр",
      "last_name": "Петров",
      "phone": "+79997654321",
      "max_id": ""
    }
  ]
}
```

## Добавление метода в репозиторий

Добавьте в `internal/domain/employee_repository.go`:

```go
type EmployeeRepository interface {
	// ... существующие методы
	
	// GetByUniversityID получает всех сотрудников вуза
	GetByUniversityID(universityID int64) ([]*Employee, error)
}
```

Реализация в `internal/infrastructure/repository/employee_postgres.go`:

```go
func (r *EmployeePostgresRepository) GetByUniversityID(universityID int64) ([]*domain.Employee, error) {
	query := `
		SELECT e.id, e.first_name, e.last_name, e.middle_name, e.phone, e.max_id, 
		       e.inn, e.kpp, e.university_id, e.created_at, e.updated_at,
		       u.id, u.name, u.inn, u.kpp, u.created_at, u.updated_at
		FROM employees e
		LEFT JOIN universities u ON e.university_id = u.id
		WHERE e.university_id = $1
		ORDER BY e.last_name, e.first_name
	`

	rows, err := r.db.Query(query, universityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []*domain.Employee
	for rows.Next() {
		employee := &domain.Employee{University: &domain.University{}}
		err := rows.Scan(
			&employee.ID, &employee.FirstName, &employee.LastName, &employee.MiddleName,
			&employee.Phone, &employee.MaxID, &employee.INN, &employee.KPP,
			&employee.UniversityID, &employee.CreatedAt, &employee.UpdatedAt,
			&employee.University.ID, &employee.University.Name, &employee.University.INN,
			&employee.University.KPP, &employee.University.CreatedAt, &employee.University.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		employees = append(employees, employee)
	}

	return employees, nil
}
```

## Конфигурация

Переменные окружения уже настроены в основном проекте. Убедитесь, что в `docker-compose.yml` указано:

```yaml
services:
  employee-service:
    environment:
      - MAXBOT_GRPC_ADDR=maxbot-service:9095
      - MAXBOT_TIMEOUT=5s
    depends_on:
      - maxbot-service
```

## Использование в бизнес-логике

### Пример: Уведомление при добавлении сотрудника

```go
func (s *EmployeeService) AddEmployeeByPhone(phone, firstName, lastName string) (*domain.Employee, error) {
	// ... существующий код создания сотрудника
	
	employee, err := s.employeeRepo.Create(employee)
	if err != nil {
		return nil, err
	}
	
	// Отправляем приветственное уведомление
	go func() {
		message := fmt.Sprintf("Добро пожаловать, %s! Вы добавлены в систему.", employee.FirstName)
		if err := s.maxService.SendNotification(employee.Phone, message); err != nil {
			log.Printf("Failed to send welcome notification: %v", err)
		}
	}()
	
	return employee, nil
}
```

### Пример: Периодическая проверка номеров

```go
// internal/app/phone_validator.go
package app

import (
	"context"
	"log"
	"time"

	"employee-service/internal/usecase"
)

type PhoneValidator struct {
	service  *usecase.EmployeeService
	interval time.Duration
	stop     chan struct{}
}

func NewPhoneValidator(service *usecase.EmployeeService, interval time.Duration) *PhoneValidator {
	return &PhoneValidator{
		service:  service,
		interval: interval,
		stop:     make(chan struct{}),
	}
}

func (v *PhoneValidator) Start(ctx context.Context) {
	ticker := time.NewTicker(v.interval)
	defer ticker.Stop()

	log.Printf("Starting phone validator with interval: %v", v.interval)

	for {
		select {
		case <-ticker.C:
			v.validateAllPhones()
		case <-v.stop:
			log.Println("Stopping phone validator")
			return
		case <-ctx.Done():
			log.Println("Context cancelled, stopping phone validator")
			return
		}
	}
}

func (v *PhoneValidator) Stop() {
	close(v.stop)
}

func (v *PhoneValidator) validateAllPhones() {
	log.Println("Starting phone validation...")
	
	// Здесь можно добавить логику валидации всех номеров
	// Например, получить все университеты и проверить их сотрудников
	
	log.Println("Phone validation completed")
}
```

