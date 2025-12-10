package cleanup

import (
	"auth-service/internal/domain"
	"context"
	"log"
	"time"
)

// TokenCleanupJob handles periodic cleanup of expired password reset tokens
type TokenCleanupJob struct {
	repo     domain.PasswordResetRepository
	interval time.Duration
	logger   *log.Logger
	stopChan chan struct{}
}

// NewTokenCleanupJob creates a new token cleanup job
func NewTokenCleanupJob(repo domain.PasswordResetRepository, interval time.Duration, logger *log.Logger) *TokenCleanupJob {
	return &TokenCleanupJob{
		repo:     repo,
		interval: interval,
		logger:   logger,
		stopChan: make(chan struct{}),
	}
}

// Start begins the periodic cleanup job
func (j *TokenCleanupJob) Start(ctx context.Context) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	j.logger.Printf("Token cleanup job started (interval: %v)", j.interval)

	// Run cleanup immediately on start
	j.runCleanup()

	for {
		select {
		case <-ticker.C:
			j.runCleanup()
		case <-j.stopChan:
			j.logger.Println("Token cleanup job stopped")
			return
		case <-ctx.Done():
			j.logger.Println("Token cleanup job stopped due to context cancellation")
			return
		}
	}
}

// Stop stops the cleanup job
func (j *TokenCleanupJob) Stop() {
	close(j.stopChan)
}

// runCleanup performs the actual cleanup operation
func (j *TokenCleanupJob) runCleanup() {
	j.logger.Println("Running token cleanup...")

	err := j.repo.DeleteExpired()
	if err != nil {
		j.logger.Printf("ERROR: Token cleanup failed: %v", err)
		return
	}

	j.logger.Println("Token cleanup completed successfully")
}
