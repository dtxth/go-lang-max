package health

import (
	"context"
	"time"
)

// HealthChecker performs health checks on external services
type HealthChecker struct {
	timeout time.Duration
}

// NewHealthChecker creates a new health checker
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		timeout: 5 * time.Second,
	}
}

// CheckMaxBotHealth checks if MaxBot service is healthy
// In mock implementation, always returns false since there's no real MaxBot connection
func (h *HealthChecker) CheckMaxBotHealth(ctx context.Context) bool {
	// Mock implementation - no real MaxBot service to check
	return false
}