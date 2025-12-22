package grpc

import (
	"employee-service/api/proto"
	"employee-service/internal/usecase"
	"context"
	"time"
	"strconv"
)

type EmployeeHandler struct {
	employeeService         *usecase.EmployeeService
	batchUpdateMaxIdUseCase *usecase.BatchUpdateMaxIdUseCase
	proto.UnimplementedEmployeeServiceServer
}

func NewEmployeeHandler(employeeService *usecase.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService: employeeService,
	}
}

func NewEmployeeHandlerWithBatch(employeeService *usecase.EmployeeService, batchUpdateMaxIdUseCase *usecase.BatchUpdateMaxIdUseCase) *EmployeeHandler {
	return &EmployeeHandler{
		employeeService:         employeeService,
		batchUpdateMaxIdUseCase: batchUpdateMaxIdUseCase,
	}
}

func (h *EmployeeHandler) GetUniversityByID(ctx context.Context, req *proto.GetUniversityByIDRequest) (*proto.GetUniversityByIDResponse, error) {
	if h.employeeService == nil {
		return &proto.GetUniversityByIDResponse{
			Error: "employee service not available",
		}, nil
	}

	university, err := h.employeeService.GetUniversityByID(req.Id)
	if err != nil {
		return &proto.GetUniversityByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByIDResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *EmployeeHandler) GetUniversityByINN(ctx context.Context, req *proto.GetUniversityByINNRequest) (*proto.GetUniversityByINNResponse, error) {
	if h.employeeService == nil {
		return &proto.GetUniversityByINNResponse{
			Error: "employee service not available",
		}, nil
	}

	university, err := h.employeeService.GetUniversityByINN(req.Inn)
	if err != nil {
		return &proto.GetUniversityByINNResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByINNResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *EmployeeHandler) GetUniversityByINNAndKPP(ctx context.Context, req *proto.GetUniversityByINNAndKPPRequest) (*proto.GetUniversityByINNAndKPPResponse, error) {
	if h.employeeService == nil {
		return &proto.GetUniversityByINNAndKPPResponse{
			Error: "employee service not available",
		}, nil
	}

	university, err := h.employeeService.GetUniversityByINNAndKPP(req.Inn, req.Kpp)
	if err != nil {
		return &proto.GetUniversityByINNAndKPPResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetUniversityByINNAndKPPResponse{
		University: &proto.University{
			Id:        university.ID,
			Name:      university.Name,
			Inn:       university.INN,
			Kpp:       university.KPP,
			CreatedAt: university.CreatedAt.Format(time.RFC3339),
			UpdatedAt: university.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

func (h *EmployeeHandler) GetEmployeeByID(ctx context.Context, req *proto.GetEmployeeByIDRequest) (*proto.GetEmployeeByIDResponse, error) {
	if h.employeeService == nil {
		return &proto.GetEmployeeByIDResponse{
			Error: "employee service not available",
		}, nil
	}

	employee, err := h.employeeService.GetEmployeeByID(req.Id)
	if err != nil {
		return &proto.GetEmployeeByIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.GetEmployeeByIDResponse{
		Employee: &proto.Employee{
			Id:           employee.ID,
			FirstName:    employee.FirstName,
			LastName:     employee.LastName,
			MiddleName:   employee.MiddleName,
			Phone:        employee.Phone,
			Role:         employee.Role,
			UniversityId: employee.UniversityID,
			MaxId:        employee.MaxID,
			CreatedAt:    employee.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    employee.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// GetAllEmployees получает всех сотрудников с пагинацией
func (h *EmployeeHandler) GetAllEmployees(ctx context.Context, req *proto.GetAllEmployeesRequest) (*proto.GetAllEmployeesResponse, error) {
	if h.employeeService == nil {
		return &proto.GetAllEmployeesResponse{
			Error: "employee service not available",
		}, nil
	}

	page := int(req.Page)
	limit := int(req.Limit)
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	
	offset := (page - 1) * limit
	
	employees, total, err := h.employeeService.GetAllEmployeesWithSortingAndSearch(limit, offset, req.SortBy, req.SortOrder, "")
	if err != nil {
		return &proto.GetAllEmployeesResponse{
			Error: err.Error(),
		}, nil
	}

	protoEmployees := make([]*proto.Employee, len(employees))
	for i, emp := range employees {
		protoEmployees[i] = &proto.Employee{
			Id:           emp.ID,
			FirstName:    emp.FirstName,
			LastName:     emp.LastName,
			MiddleName:   emp.MiddleName,
			Phone:        emp.Phone,
			Role:         emp.Role,
			UniversityId: emp.UniversityID,
			MaxId:        emp.MaxID,
			CreatedAt:    emp.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    emp.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &proto.GetAllEmployeesResponse{
		Employees: protoEmployees,
		Total:     int32(total),
		Page:      int32(page),
		Limit:     int32(limit),
	}, nil
}

// SearchEmployees ищет сотрудников по параметрам
func (h *EmployeeHandler) SearchEmployees(ctx context.Context, req *proto.SearchEmployeesRequest) (*proto.SearchEmployeesResponse, error) {
	if h.employeeService == nil {
		return &proto.SearchEmployeesResponse{
			Error: "employee service not available",
		}, nil
	}

	page := int(req.Page)
	limit := int(req.Limit)
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	
	offset := (page - 1) * limit
	
	employees, total, err := h.employeeService.GetAllEmployeesWithSortingAndSearch(limit, offset, req.SortBy, req.SortOrder, req.Query)
	if err != nil {
		return &proto.SearchEmployeesResponse{
			Error: err.Error(),
		}, nil
	}

	protoEmployees := make([]*proto.Employee, len(employees))
	for i, emp := range employees {
		protoEmployees[i] = &proto.Employee{
			Id:           emp.ID,
			FirstName:    emp.FirstName,
			LastName:     emp.LastName,
			MiddleName:   emp.MiddleName,
			Phone:        emp.Phone,
			Role:         emp.Role,
			UniversityId: emp.UniversityID,
			MaxId:        emp.MaxID,
			CreatedAt:    emp.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    emp.UpdatedAt.Format(time.RFC3339),
		}
	}

	return &proto.SearchEmployeesResponse{
		Employees: protoEmployees,
		Total:     int32(total),
		Page:      int32(page),
		Limit:     int32(limit),
	}, nil
}

// CreateEmployee создает нового сотрудника
func (h *EmployeeHandler) CreateEmployee(ctx context.Context, req *proto.CreateEmployeeRequest) (*proto.CreateEmployeeResponse, error) {
	if h.employeeService == nil {
		return &proto.CreateEmployeeResponse{
			Error: "employee service not available",
		}, nil
	}

	employee, err := h.employeeService.AddEmployeeByPhone(
		req.Phone,
		req.FirstName,
		req.LastName,
		req.MiddleName,
		req.Inn,
		req.Kpp,
		req.UniversityName,
	)
	if err != nil {
		return &proto.CreateEmployeeResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateEmployeeResponse{
		Employee: &proto.Employee{
			Id:           employee.ID,
			FirstName:    employee.FirstName,
			LastName:     employee.LastName,
			MiddleName:   employee.MiddleName,
			Phone:        employee.Phone,
			Role:         employee.Role,
			UniversityId: employee.UniversityID,
			MaxId:        employee.MaxID,
			CreatedAt:    employee.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    employee.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// CreateEmployeeSimple создает сотрудника с минимальными данными
func (h *EmployeeHandler) CreateEmployeeSimple(ctx context.Context, req *proto.CreateEmployeeSimpleRequest) (*proto.CreateEmployeeSimpleResponse, error) {
	if h.employeeService == nil {
		return &proto.CreateEmployeeSimpleResponse{
			Error: "employee service not available",
		}, nil
	}

	employee, err := h.employeeService.AddEmployeeByPhone(
		req.Phone,
		req.Name, // Используем name как FirstName
		"",       // LastName пустой
		"",       // MiddleName пустой
		"",       // INN пустой
		"",       // KPP пустой
		"",       // UniversityName пустой
	)
	if err != nil {
		return &proto.CreateEmployeeSimpleResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateEmployeeSimpleResponse{
		Employee: &proto.Employee{
			Id:           employee.ID,
			FirstName:    employee.FirstName,
			LastName:     employee.LastName,
			MiddleName:   employee.MiddleName,
			Phone:        employee.Phone,
			Role:         employee.Role,
			UniversityId: employee.UniversityID,
			MaxId:        employee.MaxID,
			CreatedAt:    employee.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    employee.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// CreateEmployeeByPhone создает сотрудника только по телефону
func (h *EmployeeHandler) CreateEmployeeByPhone(ctx context.Context, req *proto.CreateEmployeeByPhoneRequest) (*proto.CreateEmployeeByPhoneResponse, error) {
	if h.employeeService == nil {
		return &proto.CreateEmployeeByPhoneResponse{
			Error: "employee service not available",
		}, nil
	}

	employee, err := h.employeeService.AddEmployeeByPhone(
		req.Phone,
		"",  // FirstName пустой
		"",  // LastName пустой
		"",  // MiddleName пустой
		"",  // INN пустой
		"",  // KPP пустой
		"",  // UniversityName пустой
	)
	if err != nil {
		return &proto.CreateEmployeeByPhoneResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.CreateEmployeeByPhoneResponse{
		Employee: &proto.Employee{
			Id:           employee.ID,
			FirstName:    employee.FirstName,
			LastName:     employee.LastName,
			MiddleName:   employee.MiddleName,
			Phone:        employee.Phone,
			Role:         employee.Role,
			UniversityId: employee.UniversityID,
			MaxId:        employee.MaxID,
			CreatedAt:    employee.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    employee.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateEmployee обновляет данные сотрудника
func (h *EmployeeHandler) UpdateEmployee(ctx context.Context, req *proto.UpdateEmployeeRequest) (*proto.UpdateEmployeeResponse, error) {
	if h.employeeService == nil {
		return &proto.UpdateEmployeeResponse{
			Error: "employee service not available",
		}, nil
	}

	// Получаем существующего сотрудника
	existingEmployee, err := h.employeeService.GetEmployeeByID(req.Id)
	if err != nil {
		return &proto.UpdateEmployeeResponse{
			Error: err.Error(),
		}, nil
	}

	// Обновляем поля
	existingEmployee.FirstName = req.FirstName
	existingEmployee.LastName = req.LastName
	existingEmployee.MiddleName = req.MiddleName
	existingEmployee.Phone = req.Phone
	existingEmployee.Role = req.Role

	err = h.employeeService.UpdateEmployee(existingEmployee)
	if err != nil {
		return &proto.UpdateEmployeeResponse{
			Error: err.Error(),
		}, nil
	}

	// Получаем обновленного сотрудника
	updatedEmployee, err := h.employeeService.GetEmployeeByID(req.Id)
	if err != nil {
		return &proto.UpdateEmployeeResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.UpdateEmployeeResponse{
		Employee: &proto.Employee{
			Id:           updatedEmployee.ID,
			FirstName:    updatedEmployee.FirstName,
			LastName:     updatedEmployee.LastName,
			MiddleName:   updatedEmployee.MiddleName,
			Phone:        updatedEmployee.Phone,
			Role:         updatedEmployee.Role,
			UniversityId: updatedEmployee.UniversityID,
			MaxId:        updatedEmployee.MaxID,
			CreatedAt:    updatedEmployee.CreatedAt.Format(time.RFC3339),
			UpdatedAt:    updatedEmployee.UpdatedAt.Format(time.RFC3339),
		},
	}, nil
}

// DeleteEmployee удаляет сотрудника
func (h *EmployeeHandler) DeleteEmployee(ctx context.Context, req *proto.DeleteEmployeeRequest) (*proto.DeleteEmployeeResponse, error) {
	if h.employeeService == nil {
		return &proto.DeleteEmployeeResponse{
			Success: false,
			Error:   "employee service not available",
		}, nil
	}

	err := h.employeeService.DeleteEmployee(req.Id)
	if err != nil {
		return &proto.DeleteEmployeeResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	return &proto.DeleteEmployeeResponse{
		Success: true,
	}, nil
}

// BatchUpdateMaxID обновляет MAX ID для группы сотрудников
func (h *EmployeeHandler) BatchUpdateMaxID(ctx context.Context, req *proto.BatchUpdateMaxIDRequest) (*proto.BatchUpdateMaxIDResponse, error) {
	if h.batchUpdateMaxIdUseCase == nil {
		return &proto.BatchUpdateMaxIDResponse{
			Error: "batch update service not available",
		}, nil
	}

	result, err := h.batchUpdateMaxIdUseCase.StartBatchUpdate()
	if err != nil {
		return &proto.BatchUpdateMaxIDResponse{
			Error: err.Error(),
		}, nil
	}

	return &proto.BatchUpdateMaxIDResponse{
		JobId: strconv.FormatInt(result.JobID, 10),
	}, nil
}

// GetBatchStatus получает статус всех batch операций
func (h *EmployeeHandler) GetBatchStatus(ctx context.Context, req *proto.GetBatchStatusRequest) (*proto.GetBatchStatusResponse, error) {
	if h.batchUpdateMaxIdUseCase == nil {
		return &proto.GetBatchStatusResponse{
			Error: "batch update service not available",
		}, nil
	}

	page := int(req.Page)
	limit := int(req.Limit)
	
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	
	offset := (page - 1) * limit

	jobs, err := h.batchUpdateMaxIdUseCase.GetAllBatchJobs(limit, offset)
	if err != nil {
		return &proto.GetBatchStatusResponse{
			Error: err.Error(),
		}, nil
	}

	protoJobs := make([]*proto.BatchJob, len(jobs))
	for i, job := range jobs {
		protoJob := &proto.BatchJob{
			Id:             strconv.FormatInt(job.ID, 10),
			Status:         job.Status,
			TotalItems:     int32(job.Total),
			ProcessedItems: int32(job.Processed),
			FailedItems:    int32(job.Failed),
			CreatedAt:      job.StartedAt.Format(time.RFC3339),
			UpdatedAt:      job.StartedAt.Format(time.RFC3339),
		}
		
		if job.CompletedAt != nil {
			protoJob.UpdatedAt = job.CompletedAt.Format(time.RFC3339)
		}
		
		protoJobs[i] = protoJob
	}

	return &proto.GetBatchStatusResponse{
		Jobs:  protoJobs,
		Total: int32(len(jobs)), // Note: This should ideally be the total count from repository
		Page:  int32(page),
		Limit: int32(limit),
	}, nil
}

// GetBatchStatusByID получает статус конкретной batch операции
func (h *EmployeeHandler) GetBatchStatusByID(ctx context.Context, req *proto.GetBatchStatusByIDRequest) (*proto.GetBatchStatusByIDResponse, error) {
	if h.batchUpdateMaxIdUseCase == nil {
		return &proto.GetBatchStatusByIDResponse{
			Error: "batch update service not available",
		}, nil
	}

	jobID, err := strconv.ParseInt(req.JobId, 10, 64)
	if err != nil {
		return &proto.GetBatchStatusByIDResponse{
			Error: "invalid job ID format",
		}, nil
	}

	job, err := h.batchUpdateMaxIdUseCase.GetBatchJobStatus(jobID)
	if err != nil {
		return &proto.GetBatchStatusByIDResponse{
			Error: err.Error(),
		}, nil
	}

	protoJob := &proto.BatchJob{
		Id:             strconv.FormatInt(job.ID, 10),
		Status:         job.Status,
		TotalItems:     int32(job.Total),
		ProcessedItems: int32(job.Processed),
		FailedItems:    int32(job.Failed),
		CreatedAt:      job.StartedAt.Format(time.RFC3339),
		UpdatedAt:      job.StartedAt.Format(time.RFC3339),
	}
	
	if job.CompletedAt != nil {
		protoJob.UpdatedAt = job.CompletedAt.Format(time.RFC3339)
	}

	return &proto.GetBatchStatusByIDResponse{
		Job: protoJob,
	}, nil
}

// Health проверяет состояние сервиса
func (h *EmployeeHandler) Health(ctx context.Context, req *proto.HealthRequest) (*proto.HealthResponse, error) {
	return &proto.HealthResponse{
		Status: "OK",
	}, nil
}

