package usecase

import (
	"context"
	"fmt"
	"migration-service/internal/domain"
	"migration-service/internal/infrastructure/logger"
	"os"
	"strconv"
	"strings"

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

// ExcelRow represents a row from Excel file "Чаты" sheet with 19 columns
type ExcelRow struct {
	RowNumber        int
	RegionNumber     string // Колонка 0 - № Региона
	Region           string // Колонка 1 - Регион
	ChatID           string // Колонка 2 - ID чата
	ChatName         string // Колонка 3 - Название чата
	ChatURL          string // Колонка 4 - Ссылка на чат
	ParticipantCount int    // Колонка 5 - Кол-во участников чата
	OwnerID          string // Колонка 6 - ID владельца чата
	OwnerPhone       string // Колонка 7 - Телефон владельца чата
	CreatedDate      string // Колонка 8 - Дата создания чата
	CreatorID        string // Колонка 9 - ID создателя чата
	Organization     string // Колонка 10 - Организация
	INN              string // Колонка 11 - ИНН
	KPP              string // Колонка 12 - КПП
	HeadOrganization string // Колонка 13 - Головная организация
	Faculty          string // Колонка 14 - Факультет
	Course           int    // Колонка 15 - Курс
	GroupNumber      string // Колонка 16 - Группа
	AddUser          string // Колонка 17 - Добавлени пользователь
	AddAdmin         string // Колонка 18 - Добавлен администратор
}

// logInfo safely logs info message
func (uc *MigrateFromExcelUseCase) logInfo(ctx context.Context, msg string, fields map[string]interface{}) {
	if uc.logger != nil {
		uc.logger.Info(ctx, msg, fields)
	}
}

// logError safely logs error message
func (uc *MigrateFromExcelUseCase) logError(ctx context.Context, msg string, fields map[string]interface{}) {
	if uc.logger != nil {
		uc.logger.Error(ctx, msg, fields)
	}
}

// logWarn safely logs warning message
func (uc *MigrateFromExcelUseCase) logWarn(ctx context.Context, msg string, fields map[string]interface{}) {
	if uc.logger != nil {
		uc.logger.Warn(ctx, msg, fields)
	}
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

	// Check file size to decide on processing strategy
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusFailed)
		return 0, fmt.Errorf("failed to stat file: %w", err)
	}

	fileSizeMB := float64(fileInfo.Size()) / (1024 * 1024)
	uc.logInfo(ctx, "Excel file info", map[string]interface{}{
		"job_id":   job.ID,
		"size_mb":  fileSizeMB,
		"strategy": "streaming",
	})

	// Use streaming processing for all files (memory efficient)
	processed := 0
	failed := 0
	
	uc.logInfo(ctx, "Starting Excel processing", map[string]interface{}{
		"job_id":    job.ID,
		"file_path": filePath,
	})
	
	// Process file with streaming
	err = uc.processExcelStreaming(ctx, job.ID, filePath, &processed, &failed)
	if err != nil {
		uc.logError(ctx, "Excel processing failed", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
		uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusFailed)
		return 0, fmt.Errorf("failed to process Excel: %w", err)
	}
	
	uc.logInfo(ctx, "Excel processing completed", map[string]interface{}{
		"job_id":    job.ID,
		"processed": processed,
		"failed":    failed,
	})

	// Update total count (same as processed + failed for streaming)
	job.Total = processed + failed
	if err := uc.jobRepo.Update(ctx, job); err != nil {
		uc.logError(ctx, "Failed to update job total", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}
	// Final progress update
	if err := uc.jobRepo.UpdateProgress(ctx, job.ID, processed, failed); err != nil {
		uc.logError(ctx, "Failed to update final progress", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	// Update status to completed
	if err := uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusCompleted); err != nil {
		uc.logError(ctx, "Failed to update job status to completed", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	uc.logInfo(ctx, "Excel migration completed", map[string]interface{}{
		"job_id":    job.ID,
		"total":     job.Total,
		"processed": processed,
		"failed":    failed,
	})

	return job.ID, nil
}

// processExcelStreaming processes Excel file with streaming API (memory efficient)
func (uc *MigrateFromExcelUseCase) processExcelStreaming(ctx context.Context, jobID int, filePath string, processed *int, failed *int) error {
	uc.logInfo(ctx, "Opening Excel file for streaming", map[string]interface{}{
		"job_id":    jobID,
		"file_path": filePath,
	})

	f, err := excelize.OpenFile(filePath)
	if err != nil {
		uc.logError(ctx, "Failed to open Excel file", map[string]interface{}{
			"job_id":    jobID,
			"file_path": filePath,
			"error":     err.Error(),
		})
		return fmt.Errorf("failed to open Excel file: %w", err)
	}
	defer f.Close()

	uc.logInfo(ctx, "Excel file opened successfully", map[string]interface{}{
		"job_id": jobID,
	})

	// Get sheets and find the "Чаты" sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		uc.logError(ctx, "No sheets found in Excel file", map[string]interface{}{
			"job_id": jobID,
		})
		return fmt.Errorf("no sheets found in Excel file")
	}

	// Look for "Чаты" sheet, fallback to first sheet if not found
	sheetName := sheets[0] // default
	for _, sheet := range sheets {
		if sheet == "Чаты" {
			sheetName = sheet
			break
		}
	}
	uc.logInfo(ctx, "Starting streaming processing", map[string]interface{}{
		"job_id":      jobID,
		"sheet_name":  sheetName,
		"total_sheets": len(sheets),
	})

	// Use streaming API
	rows, err := f.Rows(sheetName)
	if err != nil {
		return fmt.Errorf("failed to get rows iterator: %w", err)
	}
	defer rows.Close()

	rowNumber := 0
	
	// Read and process rows one by one
	for rows.Next() {
		rowNumber++
		
		// Skip header row
		if rowNumber == 1 {
			continue
		}
		
		row, err := rows.Columns()
		if err != nil {
			uc.logWarn(ctx, "Failed to read row", map[string]interface{}{
				"row_number": rowNumber,
				"error":      err.Error(),
			})
			(*failed)++
			continue
		}

		// Check minimum columns (19 columns for "Чаты" sheet)
		if len(row) < 19 {
			uc.logWarn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
				"row_number": rowNumber,
				"columns":    len(row),
				"expected":   19,
			})
			(*failed)++
			continue
		}

		// Parse course and participant count
		course := 0
		if row[15] != "" {
			course, _ = strconv.Atoi(row[15])
		}
		
		participantCount := 0
		if row[5] != "" {
			participantCount, _ = strconv.Atoi(row[5])
		}

		// Create ExcelRow
		excelRow := ExcelRow{
			RowNumber:        rowNumber,
			RegionNumber:     strings.TrimSpace(row[0]),
			Region:           strings.TrimSpace(row[1]),
			ChatID:           strings.TrimSpace(row[2]),
			ChatName:         cleanChatName(row[3]),
			ChatURL:          strings.TrimSpace(row[4]),
			ParticipantCount: participantCount,
			OwnerID:          strings.TrimSpace(row[6]),
			OwnerPhone:       strings.TrimSpace(row[7]),
			CreatedDate:      strings.TrimSpace(row[8]),
			CreatorID:        strings.TrimSpace(row[9]),
			Organization:     strings.TrimSpace(row[10]),
			INN:              strings.TrimSpace(row[11]),
			KPP:              strings.TrimSpace(row[12]),
			HeadOrganization: strings.TrimSpace(row[13]),
			Faculty:          strings.TrimSpace(row[14]),
			Course:           course,
			GroupNumber:      strings.TrimSpace(row[16]),
			AddUser:          strings.TrimSpace(row[17]),
			AddAdmin:         strings.TrimSpace(row[18]),
		}

		// Validate required fields
		if excelRow.INN == "" || excelRow.ChatURL == "" {
			uc.logWarn(ctx, "Skipping row: missing required fields", map[string]interface{}{
				"row_number": excelRow.RowNumber,
				"inn":        excelRow.INN,
				"chat_url":   excelRow.ChatURL,
			})
			(*failed)++
			continue
		}

		// Process row immediately
		if err := uc.processRow(ctx, jobID, &excelRow); err != nil {
			uc.logError(ctx, "Failed to process row", map[string]interface{}{
				"job_id":     jobID,
				"row_number": excelRow.RowNumber,
				"error":      err.Error(),
			})
			(*failed)++

			// Record error
			migrationErr := &domain.MigrationError{
				JobID:            jobID,
				RecordIdentifier: fmt.Sprintf("row_%d", excelRow.RowNumber),
				ErrorMessage:     err.Error(),
			}
			if err := uc.errorRepo.Create(ctx, migrationErr); err != nil {
				uc.logError(ctx, "Failed to record error", map[string]interface{}{
					"job_id": jobID,
					"error":  err.Error(),
				})
			}
		} else {
			(*processed)++
		}

		// Update progress periodically
		if (*processed+*failed)%100 == 0 {
			if err := uc.jobRepo.UpdateProgress(ctx, jobID, *processed, *failed); err != nil {
				uc.logError(ctx, "Failed to update progress", map[string]interface{}{
					"job_id": jobID,
					"error":  err.Error(),
				})
			}
			uc.logInfo(ctx, "Streaming progress update", map[string]interface{}{
				"job_id":    jobID,
				"processed": *processed,
				"failed":    *failed,
				"total":     *processed + *failed,
			})
		}
	}

	// Check for errors during iteration
	if err := rows.Error(); err != nil {
		return fmt.Errorf("error during row iteration: %w", err)
	}

	uc.logInfo(ctx, "Streaming processing completed", map[string]interface{}{
		"job_id":    jobID,
		"processed": *processed,
		"failed":    *failed,
		"total":     *processed + *failed,
	})

	return nil
}

// readFromExcel reads data from Excel file using streaming API for memory efficiency
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

	// Look for "Чаты" sheet, fallback to first sheet if not found
	sheetName := sheets[0] // default
	for _, sheet := range sheets {
		if sheet == "Чаты" {
			sheetName = sheet
			break
		}
	}
	ctx := context.Background()
	
	uc.logInfo(ctx, "Starting streaming Excel read", map[string]interface{}{
		"sheet_name": sheetName,
		"file_path":  filePath,
	})

	// Use streaming API for memory efficiency
	rows, err := f.Rows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to get rows iterator: %w", err)
	}
	defer rows.Close()

	var excelRows []ExcelRow
	rowNumber := 0
	
	// Read rows one by one
	for rows.Next() {
		rowNumber++
		
		// Skip header row
		if rowNumber == 1 {
			continue
		}
		
		row, err := rows.Columns()
		if err != nil {
			uc.logWarn(ctx, "Failed to read row", map[string]interface{}{
				"row_number": rowNumber,
				"error":      err.Error(),
			})
			continue
		}
		
		// Проверяем минимальное количество колонок (19 для листа "Чаты")
		if len(row) < 19 {
			uc.logWarn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
				"row_number": rowNumber,
				"columns":    len(row),
				"expected":   19,
			})
			continue
		}

		// Парсим курс обучения и количество участников
		course := 0
		if row[15] != "" {
			course, _ = strconv.Atoi(row[15])
		}
		
		participantCount := 0
		if row[5] != "" {
			participantCount, _ = strconv.Atoi(row[5])
		}

		excelRow := ExcelRow{
			RowNumber:        rowNumber,
			RegionNumber:     strings.TrimSpace(row[0]),
			Region:           strings.TrimSpace(row[1]),
			ChatID:           strings.TrimSpace(row[2]),
			ChatName:         cleanChatName(row[3]),
			ChatURL:          strings.TrimSpace(row[4]),
			ParticipantCount: participantCount,
			OwnerID:          strings.TrimSpace(row[6]),
			OwnerPhone:       strings.TrimSpace(row[7]),
			CreatedDate:      strings.TrimSpace(row[8]),
			CreatorID:        strings.TrimSpace(row[9]),
			Organization:     strings.TrimSpace(row[10]),
			INN:              strings.TrimSpace(row[11]),
			KPP:              strings.TrimSpace(row[12]),
			HeadOrganization: strings.TrimSpace(row[13]),
			Faculty:          strings.TrimSpace(row[14]),
			Course:           course,
			GroupNumber:      strings.TrimSpace(row[16]),
			AddUser:          strings.TrimSpace(row[17]),
			AddAdmin:         strings.TrimSpace(row[18]),
		}

		// Валидация обязательных полей
		if excelRow.INN == "" || excelRow.ChatURL == "" {
			uc.logWarn(ctx, "Skipping row: missing required fields", map[string]interface{}{
				"row_number": excelRow.RowNumber,
				"inn":        excelRow.INN,
				"chat_url":   excelRow.ChatURL,
			})
			continue
		}

		excelRows = append(excelRows, excelRow)
		
		// Log progress every 1000 rows
		if len(excelRows)%1000 == 0 {
			uc.logInfo(ctx, "Excel reading progress", map[string]interface{}{
				"rows_read":  rowNumber,
				"valid_rows": len(excelRows),
			})
		}
	}

	// Check for errors during iteration
	if err := rows.Error(); err != nil {
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	uc.logInfo(ctx, "Excel parsing completed", map[string]interface{}{
		"total_rows_read": rowNumber,
		"valid_rows":      len(excelRows),
		"skipped_rows":    rowNumber - 1 - len(excelRows),
	})

	return excelRows, nil
}

// processRow processes a single row from Excel
func (uc *MigrateFromExcelUseCase) processRow(ctx context.Context, jobID int, row *ExcelRow) error {
	// 1. Создать структуру через Structure Service
	structureData := &domain.StructureData{
		INN:         row.INN,
		KPP:         row.KPP,
		FOIV:        row.Region, // Используем регион как FOIV
		OrgName:     row.Organization,
		BranchName:  row.HeadOrganization,
		FacultyName: row.Faculty,
		Course:      row.Course,
		GroupNumber: row.GroupNumber,
		ChatName:    row.ChatName,
		ChatURL:     row.ChatURL,
	}

	structureResult, err := uc.structureService.CreateStructure(ctx, structureData)
	if err != nil {
		return fmt.Errorf("failed to create structure: %w", err)
	}

	// 2. Создать или получить университет в Chat Service
	// Так как chat-service и structure-service используют разные БД,
	// 3. Создать чат через Chat Service (без university_id)
	chatData := &domain.ChatData{
		Name:           row.ChatName,
		URL:            row.ChatURL,
		ExternalChatID: row.ChatID, // Колонка 14
		Source:         "academic_group",
	}

	chatID, err := uc.chatService.CreateChat(ctx, chatData)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// 4. Связать группу с чатом
	if err := uc.structureService.LinkGroupToChat(ctx, structureResult.GroupID, chatID); err != nil {
		// Log error but don't fail the migration
		uc.logWarn(ctx, "Failed to link group to chat", map[string]interface{}{
			"group_id": structureResult.GroupID,
			"chat_id":  chatID,
			"error":    err.Error(),
		})
	}

	// 5. Добавить администратора
	// Используем телефон владельца чата
	phone := row.OwnerPhone
	

	
	// Нормализуем телефон
	phone = normalizePhone(phone)
	
	// Если есть телефон, создаем администратора
	if phone != "" {
		// Парсим флаги add_user и add_admin (1 = true, 0 = false)
		addAdmin := row.AddAdmin == "1" || strings.ToUpper(row.AddAdmin) == "TRUE"
		addUser := row.AddUser == "1" || strings.ToUpper(row.AddUser) == "TRUE"
		
		// Если оба флага пустые, устанавливаем по умолчанию
		if row.AddAdmin == "" && row.AddUser == "" {
			addAdmin = true
			addUser = true
		}

		adminData := &domain.AdministratorData{
			ChatID:   chatID,
			Phone:    phone,
			MaxID:    row.OwnerID, // Используем ID владельца как MaxID
			AddUser:  addUser,
			AddAdmin: addAdmin,
		}

		if err := uc.chatService.AddAdministrator(ctx, adminData); err != nil {
			// Log error but don't fail the migration
			uc.logWarn(ctx, "Failed to add administrator", map[string]interface{}{
				"chat_id": chatID,
				"phone":   phone,
				"error":   err.Error(),
			})
		}
	}

	// Старый код для справки (закомментирован)
	/*
	addAdmin := strings.ToUpper(row.AddAdmin) == "ИСТИНА" || 
	            strings.ToUpper(row.AddAdmin) == "TRUE"
	addUser := strings.ToUpper(row.AddUser) == "ИСТИНА" || 
	           strings.ToUpper(row.AddUser) == "TRUE"

	if addAdmin {
	*/

	return nil
}

// normalizePhone нормализует номер телефона
func normalizePhone(phone string) string {
	// Удаляем все нецифровые символы
	digits := ""
	for _, r := range phone {
		if r >= '0' && r <= '9' {
			digits += string(r)
		}
	}
	
	// Если пусто, возвращаем пустую строку
	if digits == "" {
		return ""
	}
	
	// Если начинается с 8 и длина 11, заменяем на +7
	if len(digits) == 11 && digits[0] == '8' {
		return "+7" + digits[1:]
	}
	
	// Если начинается с 7 и длина 11, добавляем +
	if len(digits) == 11 && digits[0] == '7' {
		return "+" + digits
	}
	
	// Если длина 10 и НЕ начинается с 7 или 8, добавляем +7 (российский номер без кода страны)
	// Например: 9884753064 -> +79884753064
	if len(digits) == 10 && digits[0] != '7' && digits[0] != '8' {
		return "+7" + digits
	}
	
	// Если длина 10 и начинается с 7 или 8, это неправильный номер
	// Но мы все равно попробуем добавить +7
	if len(digits) == 10 {
		return "+7" + digits
	}
	
	// Если другая длина, возвращаем как есть
	return digits
}

// cleanChatName очищает название чата от лишних кавычек и пробелов
func cleanChatName(name string) string {
	// Убираем пробелы по краям
	cleaned := strings.TrimSpace(name)
	
	// Убираем одинарные кавычки по краям
	cleaned = strings.Trim(cleaned, "'")
	
	// Убираем двойные кавычки по краям
	cleaned = strings.Trim(cleaned, "\"")
	
	// Убираем пробелы еще раз после удаления кавычек
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}
