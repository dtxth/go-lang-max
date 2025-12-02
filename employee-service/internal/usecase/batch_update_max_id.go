package usecase

import (
	"employee-service/internal/domain"
	"fmt"
	"log"
	"time"
)

// BatchUpdateMaxIdUseCase handles batch updating of MAX_id for employees
type BatchUpdateMaxIdUseCase struct {
	employeeRepo       domain.EmployeeRepository
	batchUpdateJobRepo domain.BatchUpdateJobRepository
	maxService         domain.MaxService
}

func NewBatchUpdateMaxIdUseCase(
	employeeRepo domain.EmployeeRepository,
	batchUpdateJobRepo domain.BatchUpdateJobRepository,
	maxService domain.MaxService,
) *BatchUpdateMaxIdUseCase {
	return &BatchUpdateMaxIdUseCase{
		employeeRepo:       employeeRepo,
		batchUpdateJobRepo: batchUpdateJobRepo,
		maxService:         maxService,
	}
}

// StartBatchUpdate initiates a batch update job for employees without MAX_id
// Requirements: 4.1, 4.2, 4.4, 4.5
func (uc *BatchUpdateMaxIdUseCase) StartBatchUpdate() (*domain.BatchUpdateResult, error) {
	// Count total employees without MAX_id (Requirements 4.1)
	total, err := uc.employeeRepo.CountEmployeesWithoutMaxID()
	if err != nil {
		return nil, fmt.Errorf("failed to count employees: %w", err)
	}
	
	if total == 0 {
		return &domain.BatchUpdateResult{
			Total:   0,
			Success: 0,
			Failed:  0,
		}, nil
	}
	
	// Create batch update job
	job := &domain.BatchUpdateJob{
		JobType:   "max_id_update",
		Status:    "running",
		Total:     total,
		Processed: 0,
		Failed:    0,
	}
	
	if err := uc.batchUpdateJobRepo.Create(job); err != nil {
		return nil, fmt.Errorf("failed to create batch job: %w", err)
	}
	
	// Process employees in batches of 100 (Requirements 4.2)
	batchSize := 100
	offset := 0
	successCount := 0
	failedCount := 0
	var errors []string
	
	for offset < total {
		// Get batch of employees without MAX_id
		employees, err := uc.employeeRepo.GetEmployeesWithoutMaxID(batchSize, offset)
		if err != nil {
			log.Printf("Error fetching employees at offset %d: %v", offset, err)
			errors = append(errors, fmt.Sprintf("Failed to fetch batch at offset %d: %v", offset, err))
			break
		}
		
		if len(employees) == 0 {
			break
		}
		
		// Collect phone numbers for batch request
		phones := make([]string, 0, len(employees))
		phoneToEmployee := make(map[string]*domain.Employee)
		
		for _, emp := range employees {
			if emp.Phone != "" {
				phones = append(phones, emp.Phone)
				phoneToEmployee[emp.Phone] = emp
			}
		}
		
		// Call MaxBot Service in batches (Requirements 4.2)
		if len(phones) > 0 {
			maxIDs, err := uc.maxService.BatchGetMaxIDByPhone(phones)
			if err != nil {
				log.Printf("Error calling MaxBot service: %v", err)
				errors = append(errors, fmt.Sprintf("MaxBot service error: %v", err))
				failedCount += len(phones)
			} else {
				// Update employees with received MAX_ids (Requirements 4.4)
				now := time.Now()
				for phone, maxID := range maxIDs {
					if emp, ok := phoneToEmployee[phone]; ok {
						emp.MaxID = maxID
						emp.MaxIDUpdatedAt = &now
						
						if err := uc.employeeRepo.Update(emp); err != nil {
							log.Printf("Error updating employee %d: %v", emp.ID, err)
							errors = append(errors, fmt.Sprintf("Failed to update employee %d: %v", emp.ID, err))
							failedCount++
						} else {
							successCount++
						}
					}
				}
				
				// Count failed lookups (phones not in maxIDs map)
				for phone := range phoneToEmployee {
					if _, found := maxIDs[phone]; !found {
						failedCount++
					}
				}
			}
		}
		
		// Update job progress
		job.Processed = successCount + failedCount
		job.Failed = failedCount
		if err := uc.batchUpdateJobRepo.Update(job); err != nil {
			log.Printf("Error updating job progress: %v", err)
		}
		
		offset += batchSize
	}
	
	// Mark job as completed
	completedAt := time.Now()
	job.Status = "completed"
	job.CompletedAt = &completedAt
	job.Processed = successCount + failedCount
	job.Failed = failedCount
	
	if err := uc.batchUpdateJobRepo.Update(job); err != nil {
		log.Printf("Error marking job as completed: %v", err)
	}
	
	// Generate report (Requirements 4.5)
	return &domain.BatchUpdateResult{
		JobID:   job.ID,
		Total:   total,
		Success: successCount,
		Failed:  failedCount,
		Errors:  errors,
	}, nil
}

// GetBatchJobStatus retrieves the status of a batch update job
func (uc *BatchUpdateMaxIdUseCase) GetBatchJobStatus(jobID int64) (*domain.BatchUpdateJob, error) {
	return uc.batchUpdateJobRepo.GetByID(jobID)
}

// GetAllBatchJobs retrieves all batch update jobs with pagination
func (uc *BatchUpdateMaxIdUseCase) GetAllBatchJobs(limit, offset int) ([]*domain.BatchUpdateJob, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	
	return uc.batchUpdateJobRepo.GetAll(limit, offset)
}
