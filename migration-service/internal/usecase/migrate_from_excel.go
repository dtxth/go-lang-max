package usecase

import (
	"context"
	"fmt"
	"migration-service/internal/domain"
	"migration-service/internal/infrastructure/logger"
	"strconv"

	"github.com/xuri/excelize/v2"
)

// MigrateFromExcelUseCase handles migration from Excel files
type MigrateFromExcelUseCase struct {
	jobRepo          domain.MigrationJobRepository
	errorRepo        domain.MigrationErrorRepository
	structureService domain.StructureService
	chatService      domain.ChatService
	logger           *logger.Logger
}

// NewMigrateFromExcelUseCase creates a new MigrateFromExcelUseCase
func NewMigrateFromExcelUseCase(
	jobRepo domain.MigrationJobRepository,
	errorRepo domain.MigrationErrorRepository,
	structureService domain.StructureService,
	chatService domain.ChatService,
	log *logger.Logger,
) *MigrateFromExcelUseCase {
	return &MigrateFromExcelUseCase{
		jobRepo:          jobRepo,
		errorRepo:        errorRepo,
		structureService: structureService,
		chatService:      chatService,
		logger:           log,
	}
}

// ExcelRow represents a row from Excel file
type ExcelRow struct {
	RowNumber   int
	AdminPhone  string
	INN         string
	FOIV        string
	OrgName     string
	BranchName  string
	KPP         string
	FacultyName string
	Course      int
	GroupNumber string
	ChatName    string
	ChatURL     string
}

// Execute executes the Excel migration
func (uc *MigrateFromExcelUseCase) Execute(ctx context.Context, filePath string) (int, error) {
	// Create migration job
	job := &domain.MigrationJob{
		SourceType:       domain.MigrationSourceExcel,
		SourceIdentifier: filePath,
		Status:           domain.MigrationJobStatusPending,
		Total:            0,
		Processed:        0,
		Failed:           0,
	}

	if err := uc.jobRepo.Create(ctx, job); err != nil {
		return 0, fmt.Errorf("failed to create migration job: %w", err)
	}

	// Update status to running
	if err := uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusRunning); err != nil {
		return 0, fmt.Errorf("failed to update job status: %w", err)
	}

	// Read data from Excel file
	rows, err := uc.readFromExcel(filePath)
	if err != nil {
		uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusFailed)
		return 0, fmt.Errorf("failed to read from Excel: %w", err)
	}

	// Update total count
	job.Total = len(rows)
	if err := uc.jobRepo.Update(ctx, job); err != nil {
		uc.logger.Error(ctx, "Failed to update job total", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	uc.logger.Info(ctx, "Migration progress: rows loaded", map[string]interface{}{
		"job_id": job.ID,
		"total":  job.Total,
	})

	// Process each row
	processed := 0
	failed := 0

	for _, row := range rows {
		if err := uc.processRow(ctx, job.ID, &row); err != nil {
			uc.logger.Error(ctx, "Failed to process row", map[string]interface{}{
				"job_id":     job.ID,
				"row_number": row.RowNumber,
				"error":      err.Error(),
			})
			failed++

			// Record error
			migrationErr := &domain.MigrationError{
				JobID:            job.ID,
				RecordIdentifier: fmt.Sprintf("row_%d", row.RowNumber),
				ErrorMessage:     err.Error(),
			}
			if err := uc.errorRepo.Create(ctx, migrationErr); err != nil {
				uc.logger.Error(ctx, "Failed to record error", map[string]interface{}{
					"job_id": job.ID,
					"error":  err.Error(),
				})
			}
		} else {
			processed++
		}

		// Update progress periodically
		if (processed+failed)%100 == 0 {
			if err := uc.jobRepo.UpdateProgress(ctx, job.ID, processed, failed); err != nil {
				uc.logger.Error(ctx, "Failed to update progress", map[string]interface{}{
					"job_id": job.ID,
					"error":  err.Error(),
				})
			}
			uc.logger.Info(ctx, "Migration progress update", map[string]interface{}{
				"job_id":    job.ID,
				"total":     job.Total,
				"processed": processed,
				"failed":    failed,
				"percent":   float64(processed+failed) / float64(job.Total) * 100,
			})
		}
	}

	// Final progress update
	if err := uc.jobRepo.UpdateProgress(ctx, job.ID, processed, failed); err != nil {
		uc.logger.Error(ctx, "Failed to update final progress", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	// Update status to completed
	if err := uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusCompleted); err != nil {
		uc.logger.Error(ctx, "Failed to update job status to completed", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	uc.logger.Info(ctx, "Excel migration completed", map[string]interface{}{
		"job_id":    job.ID,
		"total":     job.Total,
		"processed": processed,
		"failed":    failed,
	})

	return job.ID, nil
}

// readFromExcel reads data from Excel file
func (uc *MigrateFromExcelUseCase) readFromExcel(filePath string) ([]ExcelRow, error) {
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel file must have at least a header row and one data row")
	}

	// Validate header row (expected columns)
	header := rows[0]
	requiredColumns := []string{"phone", "inn", "foiv", "org_name", "branch", "kpp", "faculty", "course", "group", "chat_name", "url"}
	if len(header) < len(requiredColumns) {
		return nil, domain.ErrMissingRequiredColumns
	}

	var excelRows []ExcelRow
	ctx := context.Background() // Create context for logging
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) < len(requiredColumns) {
			uc.logger.Warn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
				"row_number": i + 1,
				"columns":    len(row),
			})
			continue
		}

		// Parse course as integer
		course := 0
		if row[7] != "" {
			course, _ = strconv.Atoi(row[7])
		}

		excelRow := ExcelRow{
			RowNumber:   i + 1,
			AdminPhone:  row[0],
			INN:         row[1],
			FOIV:        row[2],
			OrgName:     row[3],
			BranchName:  row[4],
			KPP:         row[5],
			FacultyName: row[6],
			Course:      course,
			GroupNumber: row[8],
			ChatName:    row[9],
			ChatURL:     row[10],
		}

		// Validate required fields
		if excelRow.INN == "" || excelRow.ChatURL == "" {
			uc.logger.Warn(ctx, "Skipping row: missing required fields", map[string]interface{}{
				"row_number": excelRow.RowNumber,
			})
			continue
		}

		excelRows = append(excelRows, excelRow)
	}

	return excelRows, nil
}

// processRow processes a single row from Excel
func (uc *MigrateFromExcelUseCase) processRow(ctx context.Context, jobID int, row *ExcelRow) error {
	// Create structure hierarchy via Structure Service
	structureData := &domain.StructureData{
		INN:         row.INN,
		KPP:         row.KPP,
		FOIV:        row.FOIV,
		OrgName:     row.OrgName,
		BranchName:  row.BranchName,
		FacultyName: row.FacultyName,
		Course:      row.Course,
		GroupNumber: row.GroupNumber,
		ChatName:    row.ChatName,
		ChatURL:     row.ChatURL,
		AdminPhone:  row.AdminPhone,
	}

	structureResult, err := uc.structureService.CreateStructure(ctx, structureData)
	if err != nil {
		return fmt.Errorf("failed to create structure: %w", err)
	}

	// Create chat with source='academic_group'
	chatData := &domain.ChatData{
		Name:         row.ChatName,
		URL:          row.ChatURL,
		UniversityID: structureResult.UniversityID,
		BranchID:     structureResult.BranchID,
		FacultyID:    structureResult.FacultyID,
		Source:       "academic_group",
		AdminPhone:   row.AdminPhone,
	}

	chatID, err := uc.chatService.CreateChat(ctx, chatData)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// Link group to chat
	if err := uc.structureService.LinkGroupToChat(ctx, structureResult.GroupID, chatID); err != nil {
		// Log error but don't fail the migration
		uc.logger.Warn(ctx, "Failed to link group to chat", map[string]interface{}{
			"group_id": structureResult.GroupID,
			"chat_id":  chatID,
			"error":    err.Error(),
		})
	}

	// Add administrator if phone is provided
	if row.AdminPhone != "" {
		if err := uc.chatService.AddAdministrator(ctx, chatID, row.AdminPhone); err != nil {
			// Log error but don't fail the migration
			uc.logger.Warn(ctx, "Failed to add administrator for chat", map[string]interface{}{
				"chat_id": chatID,
				"phone":   row.AdminPhone,
				"error":   err.Error(),
			})
		}
	}

	return nil
}
