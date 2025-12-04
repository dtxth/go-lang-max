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

// ExcelRow represents a row from Excel file with 18 columns
type ExcelRow struct {
	RowNumber        int
	AdminPhone1      string // Колонка 0 - Нормализованный номер телефона администратора
	MaxID            string // Колонка 1 - max_id
	INNReference     string // Колонка 2 - ИНН_Справочник
	FOIVReference    string // Колонка 3 - ФОИВ_Справочник
	OrgNameRef       string // Колонка 4 - Наименование организации_Справочник
	BranchName       string // Колонка 5 - Наименование головного подразделения/филиала
	INN              string // Колонка 6 - ИНН юридического лица
	KPP              string // Колонка 7 - КПП головного подразделения/филиала
	FacultyName      string // Колонка 8 - Факультет/институт/иная структурная классификация
	Course           int    // Колонка 9 - Курс обучения
	GroupNumber      string // Колонка 10 - Номер группы
	ChatName         string // Колонка 11 - Название чата
	AdminPhone2      string // Колонка 12 - Мобильный номер телефона администратора чата
	FileName         string // Колонка 13 - Наименование файла
	ChatID           string // Колонка 14 - chat_id
	ChatURL          string // Колонка 15 - link
	AddUser          string // Колонка 16 - add_user
	AddAdmin         string // Колонка 17 - add_admin
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

	// Get the first sheet
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		uc.logError(ctx, "No sheets found in Excel file", map[string]interface{}{
			"job_id": jobID,
		})
		return fmt.Errorf("no sheets found in Excel file")
	}

	sheetName := sheets[0]
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

		// Check minimum columns
		if len(row) < 18 {
			uc.logWarn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
				"row_number": rowNumber,
				"columns":    len(row),
				"expected":   18,
			})
			(*failed)++
			continue
		}

		// Parse course
		course := 0
		if row[9] != "" {
			course, _ = strconv.Atoi(row[9])
		}

		// Create ExcelRow
		excelRow := ExcelRow{
			RowNumber:     rowNumber,
			AdminPhone1:   strings.TrimSpace(row[0]),
			MaxID:         strings.TrimSpace(row[1]),
			INNReference:  strings.TrimSpace(row[2]),
			FOIVReference: strings.TrimSpace(row[3]),
			OrgNameRef:    strings.TrimSpace(row[4]),
			BranchName:    strings.TrimSpace(row[5]),
			INN:           strings.TrimSpace(row[6]),
			KPP:           strings.TrimSpace(row[7]),
			FacultyName:   strings.TrimSpace(row[8]),
			Course:        course,
			GroupNumber:   strings.TrimSpace(row[10]),
			ChatName:      cleanChatName(row[11]),
			AdminPhone2:   strings.TrimSpace(row[12]),
			FileName:      strings.TrimSpace(row[13]),
			ChatID:        strings.TrimSpace(row[14]),
			ChatURL:       strings.TrimSpace(row[15]),
			AddUser:       strings.TrimSpace(row[16]),
			AddAdmin:      strings.TrimSpace(row[17]),
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

	sheetName := sheets[0]
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
		
		// Проверяем минимальное количество колонок (18)
		if len(row) < 18 {
			uc.logWarn(ctx, "Skipping row: insufficient columns", map[string]interface{}{
				"row_number": rowNumber,
				"columns":    len(row),
				"expected":   18,
			})
			continue
		}

		// Парсим курс обучения
		course := 0
		if row[9] != "" {
			course, _ = strconv.Atoi(row[9])
		}

		excelRow := ExcelRow{
			RowNumber:     rowNumber,
			AdminPhone1:   strings.TrimSpace(row[0]),
			MaxID:         strings.TrimSpace(row[1]),
			INNReference:  strings.TrimSpace(row[2]),
			FOIVReference: strings.TrimSpace(row[3]),
			OrgNameRef:    strings.TrimSpace(row[4]),
			BranchName:    strings.TrimSpace(row[5]),
			INN:           strings.TrimSpace(row[6]),
			KPP:           strings.TrimSpace(row[7]),
			FacultyName:   strings.TrimSpace(row[8]),
			Course:        course,
			GroupNumber:   strings.TrimSpace(row[10]),
			ChatName:      cleanChatName(row[11]),
			AdminPhone2:   strings.TrimSpace(row[12]),
			FileName:      strings.TrimSpace(row[13]),
			ChatID:        strings.TrimSpace(row[14]),
			ChatURL:       strings.TrimSpace(row[15]),
			AddUser:       strings.TrimSpace(row[16]),
			AddAdmin:      strings.TrimSpace(row[17]),
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
		FOIV:        row.FOIVReference,
		OrgName:     row.OrgNameRef,
		BranchName:  row.BranchName,
		FacultyName: row.FacultyName,
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
	// нужно создать университет в обеих базах
	universityData := &domain.UniversityData{
		INN:  row.INN,
		KPP:  row.KPP,
		Name: row.OrgNameRef,
	}
	
	chatUniversityID, err := uc.chatService.CreateOrGetUniversity(ctx, universityData)
	if err != nil {
		return fmt.Errorf("failed to create university in chat service: %w", err)
	}

	// 3. Создать чат через Chat Service
	chatData := &domain.ChatData{
		Name:           row.ChatName,
		URL:            row.ChatURL,
		ExternalChatID: row.ChatID, // Колонка 14
		UniversityID:   chatUniversityID, // Используем ID из chat-service
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
	// Проверяем, есть ли телефон администратора
	phone := row.AdminPhone1
	if phone == "" {
		phone = row.AdminPhone2
	}
	
	// DEBUG: Логируем первые 5 строк
	if row.RowNumber <= 5 {
		fmt.Printf("[DEBUG] Row %d: phone1='%s', phone2='%s', normalized='%s', maxid='%s'\n",
			row.RowNumber, row.AdminPhone1, row.AdminPhone2, normalizePhone(phone), row.MaxID)
	}
	
	// Нормализуем телефон
	phone = normalizePhone(phone)
	
	// Если есть телефон, создаем администратора
	if phone != "" {
		// Парсим флаги add_user и add_admin
		addAdmin := strings.ToUpper(row.AddAdmin) == "ИСТИНА" || 
		            strings.ToUpper(row.AddAdmin) == "TRUE" ||
		            row.AddAdmin != "" // Если не пусто, считаем что нужно добавить
		addUser := strings.ToUpper(row.AddUser) == "ИСТИНА" || 
		           strings.ToUpper(row.AddUser) == "TRUE" ||
		           row.AddUser != "" // Если не пусто, считаем что нужно добавить
		
		// Если оба флага пустые, устанавливаем по умолчанию
		if row.AddAdmin == "" && row.AddUser == "" {
			addAdmin = true
			addUser = true
		}

		adminData := &domain.AdministratorData{
			ChatID:   chatID,
			Phone:    phone,
			MaxID:    row.MaxID,
			AddUser:  addUser,
			AddAdmin: addAdmin,
		}

		// DEBUG: Log first 5 administrator creations
		if row.RowNumber <= 5 {
			fmt.Printf("[DEBUG] Adding administrator: chat_id=%d, phone=%s, max_id=%s, add_user=%v, add_admin=%v\n",
				chatID, phone, row.MaxID, addUser, addAdmin)
		}

		if err := uc.chatService.AddAdministrator(ctx, adminData); err != nil {
			// Log error but don't fail the migration
			uc.logWarn(ctx, "Failed to add administrator", map[string]interface{}{
				"chat_id": chatID,
				"phone":   phone,
				"error":   err.Error(),
			})
			// DEBUG: Log first 5 errors
			if row.RowNumber <= 5 {
				fmt.Printf("[DEBUG] Error adding administrator: %v\n", err)
			}
		} else if row.RowNumber <= 5 {
			fmt.Printf("[DEBUG] Administrator added successfully\n")
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
