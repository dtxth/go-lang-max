package metrics

import (
	"sync"
	"time"
)

// Metrics tracks password operations and notification delivery
type Metrics struct {
	mu sync.RWMutex
	
	// Password operations
	userCreations       int64
	passwordResets      int64
	passwordChanges     int64
	
	// Notification delivery
	notificationsSent   int64
	notificationsFailed int64
	
	// Token operations
	tokensGenerated     int64
	tokensUsed          int64
	tokensExpired       int64
	tokensInvalidated   int64
	
	// Health status
	maxBotHealthy       bool
	lastHealthCheck     time.Time
}

// NewMetrics creates a new metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		maxBotHealthy: true, // Assume healthy initially
	}
}

// IncrementUserCreations increments the user creation counter
func (m *Metrics) IncrementUserCreations() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.userCreations++
}

// IncrementPasswordResets increments the password reset counter
func (m *Metrics) IncrementPasswordResets() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.passwordResets++
}

// IncrementPasswordChanges increments the password change counter
func (m *Metrics) IncrementPasswordChanges() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.passwordChanges++
}

// IncrementNotificationsSent increments the successful notification counter
func (m *Metrics) IncrementNotificationsSent() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notificationsSent++
}

// IncrementNotificationsFailed increments the failed notification counter
func (m *Metrics) IncrementNotificationsFailed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notificationsFailed++
}

// IncrementTokensGenerated increments the token generation counter
func (m *Metrics) IncrementTokensGenerated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensGenerated++
}

// IncrementTokensUsed increments the token usage counter
func (m *Metrics) IncrementTokensUsed() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensUsed++
}

// IncrementTokensExpired increments the expired token counter
func (m *Metrics) IncrementTokensExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensExpired++
}

// IncrementTokensInvalidated increments the invalidated token counter
func (m *Metrics) IncrementTokensInvalidated() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.tokensInvalidated++
}

// SetMaxBotHealth sets the MaxBot service health status
func (m *Metrics) SetMaxBotHealth(healthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.maxBotHealthy = healthy
	m.lastHealthCheck = time.Now()
}

// GetMetrics returns a snapshot of all metrics
func (m *Metrics) GetMetrics() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	return MetricsSnapshot{
		UserCreations:       m.userCreations,
		PasswordResets:      m.passwordResets,
		PasswordChanges:     m.passwordChanges,
		NotificationsSent:   m.notificationsSent,
		NotificationsFailed: m.notificationsFailed,
		TokensGenerated:     m.tokensGenerated,
		TokensUsed:          m.tokensUsed,
		TokensExpired:       m.tokensExpired,
		TokensInvalidated:   m.tokensInvalidated,
		MaxBotHealthy:       m.maxBotHealthy,
		LastHealthCheck:     m.lastHealthCheck,
	}
}

// GetNotificationSuccessRate returns the notification success rate (0.0 to 1.0)
func (m *Metrics) GetNotificationSuccessRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	total := m.notificationsSent + m.notificationsFailed
	if total == 0 {
		return 1.0 // No notifications sent yet, assume 100% success
	}
	
	return float64(m.notificationsSent) / float64(total)
}

// GetNotificationFailureRate returns the notification failure rate (0.0 to 1.0)
func (m *Metrics) GetNotificationFailureRate() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	total := m.notificationsSent + m.notificationsFailed
	if total == 0 {
		return 0.0 // No notifications sent yet, assume 0% failure
	}
	
	return float64(m.notificationsFailed) / float64(total)
}

// IsMaxBotHealthy returns the current MaxBot service health status
func (m *Metrics) IsMaxBotHealthy() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.maxBotHealthy
}

// MetricsSnapshot represents a point-in-time snapshot of metrics
type MetricsSnapshot struct {
	UserCreations       int64
	PasswordResets      int64
	PasswordChanges     int64
	NotificationsSent   int64
	NotificationsFailed int64
	TokensGenerated     int64
	TokensUsed          int64
	TokensExpired       int64
	TokensInvalidated   int64
	MaxBotHealthy       bool
	LastHealthCheck     time.Time
}
