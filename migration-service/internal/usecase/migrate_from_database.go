package usecase

import (
	"context"
	"database/sql"
	"fmt"
	"migration-service/internal/domain"
	"migration-service/internal/infrastructure/logger"
)

// MigrateFromDatabaseUseCase handles migration from existing database
type MigrateFromDatabaseUseCase struct {
	sourceDB          *sql.DB
	jobRepo           domain.MigrationJobRepository
	errorRepo         domain.MigrationErrorRepository
	universityRepo    domain.UniversityRepository
	chatService       domain.ChatService
	logger            *logger.Logger
}

// NewMigrateFromDatabaseUseCase creates a new MigrateFromDatabaseUseCase
func NewMigrateFromDatabaseUseCase(
	sourceDB *sql.DB,
	jobRepo domain.MigrationJobRepository,
	errorRepo domain.MigrationErrorRepository,
	universityRepo domain.UniversityRepository,
	chatService domain.ChatService,
	log *logger.Logger,
) *MigrateFromDatabaseUseCase {
	return &MigrateFromDatabaseUseCase{
		sourceDB:       sourceDB,
		jobRepo:        jobRepo,
		errorRepo:      errorRepo,
		universityRepo: universityRepo,
		chatService:    chatService,
		logger:         log,
	}
}

// ChatRecord represents a chat record from the source database
type ChatRecord struct {
	ID         int
	INN        string
	Name       string
	URL        string
	AdminPhone string
}

// Execute executes the database migration
func (uc *MigrateFromDatabaseUseCase) Execute(ctx context.Context, sourceIdentifier string) (int, error) {
	// Create migration job
	job := &domain.MigrationJob{
		SourceType:       domain.MigrationSourceDatabase,
		SourceIdentifier: sourceIdentifier,
		Status:           domain.MigrationJobStatusPending,
		Total:            0,
		Processed:        0,
		Failed:           0,
	}

	if err := uc.jobRepo.Create(ctx, job); err != nil {
		return 0, fmt.Errorf("failed to create migration job: %w", err)
	}

	uc.logger.Info(ctx, "Migration job created", map[string]interface{}{
		"job_id":            job.ID,
		"source_type":       string(domain.MigrationSourceDatabase),
		"source_identifier": sourceIdentifier,
	})

	// Update status to running
	if err := uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusRunning); err != nil {
		return 0, fmt.Errorf("failed to update job status: %w", err)
	}

	uc.logger.Info(ctx, "Migration job started", map[string]interface{}{
		"job_id": job.ID,
	})

	// Read chat data from source database
	chats, err := uc.readChatsFromDatabase(ctx)
	if err != nil {
		uc.jobRepo.UpdateStatus(ctx, job.ID, domain.MigrationJobStatusFailed)
		return 0, fmt.Errorf("failed to read chats from database: %w", err)
	}

	// Update total count
	job.Total = len(chats)
	if err := uc.jobRepo.Update(ctx, job); err != nil {
		uc.logger.Error(ctx, "Failed to update job total", map[string]interface{}{
			"job_id": job.ID,
			"error":  err.Error(),
		})
	}

	uc.logger.Info(ctx, "Migration progress: chats loaded", map[string]interface{}{
		"job_id": job.ID,
		"total":  job.Total,
	})

	// Process each chat
	processed := 0
	failed := 0

	for _, chat := range chats {
		if err := uc.processChat(ctx, job.ID, &chat); err != nil {
			uc.logger.Error(ctx, "Failed to process chat", map[string]interface{}{
				"job_id":  job.ID,
				"chat_id": chat.ID,
				"error":   err.Error(),
			})
			failed++
			
			// Record error
			migrationErr := &domain.MigrationError{
				JobID:            job.ID,
				RecordIdentifier: fmt.Sprintf("chat_id_%d", chat.ID),
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

	uc.logger.Info(ctx, "Migration completed", map[string]interface{}{
		"job_id":    job.ID,
		"total":     job.Total,
		"processed": processed,
		"failed":    failed,
	})

	return job.ID, nil
}

// readChatsFromDatabase reads chat data from the source database
func (uc *MigrateFromDatabaseUseCase) readChatsFromDatabase(ctx context.Context) ([]ChatRecord, error) {
	// This query assumes the source database has a table with these columns
	// Adjust the query based on the actual schema
	query := `
		SELECT id, inn, name, url, admin_phone
		FROM chats
		WHERE source = 'admin_panel' OR source IS NULL
		ORDER BY id
	`

	rows, err := uc.sourceDB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query chats: %w", err)
	}
	defer rows.Close()

	var chats []ChatRecord
	for rows.Next() {
		var chat ChatRecord
		var adminPhone sql.NullString

		if err := rows.Scan(&chat.ID, &chat.INN, &chat.Name, &chat.URL, &adminPhone); err != nil {
			return nil, fmt.Errorf("failed to scan chat row: %w", err)
		}

		if adminPhone.Valid {
			chat.AdminPhone = adminPhone.String
		}

		chats = append(chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating chat rows: %w", err)
	}

	return chats, nil
}

// processChat processes a single chat record
func (uc *MigrateFromDatabaseUseCase) processChat(ctx context.Context, jobID int, chat *ChatRecord) error {
	// Lookup or create university by INN
	university, err := uc.universityRepo.GetByINN(ctx, chat.INN)
	if err != nil {
		return fmt.Errorf("failed to get university by INN: %w", err)
	}

	if university == nil {
		// University doesn't exist - this shouldn't happen in a real migration
		// In production, we would create it via Structure Service
		return fmt.Errorf("university with INN %s not found", chat.INN)
	}

	// Create chat with source='admin_panel'
	chatData := &domain.ChatData{
		Name:         chat.Name,
		URL:          chat.URL,
		UniversityID: university.ID,
		Source:       "admin_panel",
		AdminPhone:   chat.AdminPhone,
	}

	chatID, err := uc.chatService.CreateChat(ctx, chatData)
	if err != nil {
		return fmt.Errorf("failed to create chat: %w", err)
	}

	// Add administrator if phone is provided
	if chat.AdminPhone != "" {
		adminData := &domain.AdministratorData{
			ChatID: chatID,
			Phone:  chat.AdminPhone,
		}
		if err := uc.chatService.AddAdministrator(ctx, adminData); err != nil {
			// Log error but don't fail the migration
			uc.logger.Warn(ctx, "Failed to add administrator for chat", map[string]interface{}{
				"chat_id": chatID,
				"phone":   chat.AdminPhone,
				"error":   err.Error(),
			})
		}
	}

	return nil
}
