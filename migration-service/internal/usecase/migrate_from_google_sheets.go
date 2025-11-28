package usecase

import (
	"context"
	"fmt"
	"log"
	"migration-service/internal/domain"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// MigrateFromGoogleSheetsUseCase handles migration from Google Sheets
type MigrateFromGoogleSheetsUseCase struct {
	jobRepo        domain.MigrationJobRepository
	errorRepo      domain.MigrationErrorRepository
	universityRepo domain.UniversityRepository
	chatService    domain.ChatService
	credentialsPath string
}

// NewMigrateFromGoogleSheetsUseCase creates a new MigrateFromGoogleSheetsUseCase
func NewMigrateFromGoogleSheetsUseCase(
	jobRepo domain.MigrationJobRepository,
	errorRepo domain.MigrationErrorRepository,
	universityRepo domain.UniversityRepository,
	chatService domain.ChatService,
	credentialsPath string,
) *MigrateFromGoogleSheetsUseCase {
	return &MigrateFromGoogleSheetsUseCase{
		jobRepo:         jobRepo,
		errorRepo:       errorRepo,
		universityRepo:  universityRepo,
		chatService:     chatService,
		credentialsPath: credentialsPath,
	}
}

// SheetRow represents a row from Google Sheets
type SheetRow struct {
	RowNumber  int
	INN        string
	KPP        string
	URL        string
	AdminPhone string
}

// Execute executes the Google Sheets migration
func (uc *MigrateFromGoogleSheetsUseCase) Execute(ctx context.Context, spreadsheetID string) (int, error) {
	// Create migration job
	job := &domain.MigrationJob{
		SourceType:       domain.MigrationSourceGoogleSheets,
		SourceIdentifier: spreadsheetID,
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

	// Read data from Google Sheets
	rows, err := uc.readFromGoogleSheets(ctx, spreadsheetID)
	if err != nil {
		uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusFailed)
		return 0, fmt.Errorf("failed to read from Google Sheets: %w", err)
	}

	// Update total count
	job.Total = len(rows)
	if err := uc.jobRepo.Update(ctx, job); err != nil {
		log.Printf("Failed to update job total: %v", err)
	}

	// Process each row
	processed := 0
	failed := 0

	for _, row := range rows {
		if err := uc.processRow(ctx, job.ID, &row); err != nil {
			log.Printf("Failed to process row %d: %v", row.RowNumber, err)
			failed++

			// Record error
			migrationErr := &domain.MigrationError{
				JobID:            job.ID,
				RecordIdentifier: fmt.Sprintf("row_%d", row.RowNumber),
				ErrorMessage:     err.Error(),
			}
			if err := uc.errorRepo.Create(ctx, migrationErr); err != nil {
				log.Printf("Failed to record error: %v", err)
			}
		} else {
			processed++
		}

		// Update progress periodically
		if (processed+failed)%50 == 0 {
			if err := uc.jobRepo.UpdateProgress(ctx, job.ID, processed, failed); err != nil {
				log.Printf("Failed to update progress: %v", err)
			}
		}
	}

	// Final progress update
	if err := uc.jobRepo.UpdateProgress(ctx, job.ID, processed, failed); err != nil {
		log.Printf("Failed to update final progress: %v", err)
	}

	// Update status to completed
	if err := uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusCompleted); err != nil {
		log.Printf("Failed to update job status to completed: %v", err)
	}

	log.Printf("Google Sheets migration completed: total=%d, processed=%d, failed=%d", job.Total, processed, failed)

	return job.ID, nil
}

// readFromGoogleSheets reads data from Google Sheets
func (uc *MigrateFromGoogleSheetsUseCase) readFromGoogleSheets(ctx context.Context, spreadsheetID string) ([]SheetRow, error) {
	// Authenticate with Google Sheets API
	srv, err := sheets.NewService(ctx, option.WithCredentialsFile(uc.credentialsPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create sheets service: %w", err)
	}

	// Read data from the sheet
	// Assuming columns: INN, KPP, URL, Admin Phone
	readRange := "Sheet1!A2:D" // Skip header row
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to read from sheet: %w", err)
	}

	var rows []SheetRow
	for i, row := range resp.Values {
		if len(row) < 3 {
			log.Printf("Skipping row %d: insufficient columns", i+2)
			continue
		}

		sheetRow := SheetRow{
			RowNumber: i + 2, // +2 because we start from row 2 (after header)
		}

		// Parse columns
		if len(row) > 0 {
			sheetRow.INN = fmt.Sprintf("%v", row[0])
		}
		if len(row) > 1 {
			sheetRow.KPP = fmt.Sprintf("%v", row[1])
		}
		if len(row) > 2 {
			sheetRow.URL = fmt.Sprintf("%v", row[2])
		}
		if len(row) > 3 {
			sheetRow.AdminPhone = fmt.Sprintf("%v", row[3])
		}

		// Validate required fields
		if sheetRow.INN == "" || sheetRow.URL == "" {
			log.Printf("Skipping row %d: missing required fields", sheetRow.RowNumber)
			continue
		}

		rows = append(rows, sheetRow)
	}

	return rows, nil
}

// processRow processes a single row from Google Sheets
func (uc *MigrateFromGoogleSheetsUseCase) processRow(ctx context.Context, jobID int, row *SheetRow) error {
	// Lookup or create university by INN and KPP
	var university *domain.University
	var err error

	if row.KPP != "" {
		university, err = uc.universityRepo.GetByINNAndKPP(ctx, row.INN, row.KPP)
	} else {
		university, err = uc.universityRepo.GetByINN(ctx, row.INN)
	}

	if err != nil {
		return fmt.Errorf("failed to get university: %w", err)
	}

	if university == nil {
		// University doesn't exist - this shouldn't happen in a real migration
		// In production, we would create it via Structure Service
		return fmt.Errorf("university with INN %s not found", row.INN)
	}

	// Create chat with source='bot_registrar'
	chatData := &domain.ChatData{
		Name:         "", // Name not provided in Google Sheets
		URL:          row.URL,
		UniversityID: university.ID,
		Source:       "bot_registrar",
		AdminPhone:   row.AdminPhone,
	}

	chatID, err := uc.chatService.CreateChat(ctx, chatData)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// Add administrator if phone is provided
	if row.AdminPhone != "" {
		if err := uc.chatService.AddAdministrator(ctx, chatID, row.AdminPhone); err != nil {
			// Log error but don't fail the migration
			log.Printf("Failed to add administrator for chat %d: %v", chatID, err)
		}
	}

	return nil
}
