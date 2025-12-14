package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestMetricsIncrement tests that metrics counters increment correctly
func TestMetricsIncrement(t *testing.T) {
	m := NewMetrics()

	// Test user creation metrics
	m.IncrementUserCreations()
	m.IncrementUserCreations()
	snapshot := m.GetMetrics()
	assert.Equal(t, int64(2), snapshot.UserCreations, "User creations should be 2")

	// Test password reset metrics
	m.IncrementPasswordResets()
	snapshot = m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.PasswordResets, "Password resets should be 1")

	// Test password change metrics
	m.IncrementPasswordChanges()
	m.IncrementPasswordChanges()
	m.IncrementPasswordChanges()
	snapshot = m.GetMetrics()
	assert.Equal(t, int64(3), snapshot.PasswordChanges, "Password changes should be 3")
}

// TestMetricsNotifications tests notification metrics
func TestMetricsNotifications(t *testing.T) {
	m := NewMetrics()

	// Test successful notifications
	m.IncrementNotificationsSent()
	m.IncrementNotificationsSent()
	m.IncrementNotificationsSent()
	
	// Test failed notifications
	m.IncrementNotificationsFailed()

	snapshot := m.GetMetrics()
	assert.Equal(t, int64(3), snapshot.NotificationsSent, "Notifications sent should be 3")
	assert.Equal(t, int64(1), snapshot.NotificationsFailed, "Notifications failed should be 1")

	// Test success rate calculation
	successRate := m.GetNotificationSuccessRate()
	assert.Equal(t, 0.75, successRate, "Success rate should be 75%")

	// Test failure rate calculation
	failureRate := m.GetNotificationFailureRate()
	assert.Equal(t, 0.25, failureRate, "Failure rate should be 25%")
}

// TestMetricsTokenOperations tests token operation metrics
func TestMetricsTokenOperations(t *testing.T) {
	m := NewMetrics()

	// Test token generation
	m.IncrementTokensGenerated()
	m.IncrementTokensGenerated()
	snapshot := m.GetMetrics()
	assert.Equal(t, int64(2), snapshot.TokensGenerated, "Tokens generated should be 2")

	// Test token usage
	m.IncrementTokensUsed()
	snapshot = m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.TokensUsed, "Tokens used should be 1")

	// Test token expiration
	m.IncrementTokensExpired()
	snapshot = m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.TokensExpired, "Tokens expired should be 1")

	// Test token invalidation
	m.IncrementTokensInvalidated()
	snapshot = m.GetMetrics()
	assert.Equal(t, int64(1), snapshot.TokensInvalidated, "Tokens invalidated should be 1")
}

// TestMetricsHealthStatus tests MaxBot health status tracking
func TestMetricsHealthStatus(t *testing.T) {
	m := NewMetrics()

	// Initially should be healthy
	assert.True(t, m.IsMaxBotHealthy(), "MaxBot should be healthy initially")

	// Set to unhealthy
	m.SetMaxBotHealth(false)
	assert.False(t, m.IsMaxBotHealthy(), "MaxBot should be unhealthy after setting")

	snapshot := m.GetMetrics()
	assert.False(t, snapshot.MaxBotHealthy, "Snapshot should show unhealthy")
	assert.True(t, time.Since(snapshot.LastHealthCheck) < time.Second, 
		"Last health check should be recent")

	// Set back to healthy
	m.SetMaxBotHealth(true)
	assert.True(t, m.IsMaxBotHealthy(), "MaxBot should be healthy again")
}

// TestMetricsNotificationRatesWithNoData tests rate calculations with no data
func TestMetricsNotificationRatesWithNoData(t *testing.T) {
	m := NewMetrics()

	// With no notifications, success rate should be 100%
	successRate := m.GetNotificationSuccessRate()
	assert.Equal(t, 1.0, successRate, "Success rate should be 100% with no data")

	// With no notifications, failure rate should be 0%
	failureRate := m.GetNotificationFailureRate()
	assert.Equal(t, 0.0, failureRate, "Failure rate should be 0% with no data")
}

// TestMetricsNotificationRatesAllFailed tests rate calculations when all fail
func TestMetricsNotificationRatesAllFailed(t *testing.T) {
	m := NewMetrics()

	// All notifications fail
	m.IncrementNotificationsFailed()
	m.IncrementNotificationsFailed()
	m.IncrementNotificationsFailed()

	// Success rate should be 0%
	successRate := m.GetNotificationSuccessRate()
	assert.Equal(t, 0.0, successRate, "Success rate should be 0% when all fail")

	// Failure rate should be 100%
	failureRate := m.GetNotificationFailureRate()
	assert.Equal(t, 1.0, failureRate, "Failure rate should be 100% when all fail")
}

// TestMetricsConcurrency tests that metrics are thread-safe
func TestMetricsConcurrency(t *testing.T) {
	m := NewMetrics()

	// Run concurrent increments
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.IncrementUserCreations()
				m.IncrementNotificationsSent()
				m.IncrementTokensGenerated()
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify counts
	snapshot := m.GetMetrics()
	assert.Equal(t, int64(1000), snapshot.UserCreations, "User creations should be 1000")
	assert.Equal(t, int64(1000), snapshot.NotificationsSent, "Notifications sent should be 1000")
	assert.Equal(t, int64(1000), snapshot.TokensGenerated, "Tokens generated should be 1000")
}

// TestMetricsSnapshot tests that snapshots are independent
func TestMetricsSnapshot(t *testing.T) {
	m := NewMetrics()

	// Take first snapshot
	m.IncrementUserCreations()
	snapshot1 := m.GetMetrics()

	// Modify metrics
	m.IncrementUserCreations()
	m.IncrementPasswordResets()

	// Take second snapshot
	snapshot2 := m.GetMetrics()

	// Verify snapshots are independent
	assert.Equal(t, int64(1), snapshot1.UserCreations, "First snapshot should have 1 user creation")
	assert.Equal(t, int64(2), snapshot2.UserCreations, "Second snapshot should have 2 user creations")
	assert.Equal(t, int64(0), snapshot1.PasswordResets, "First snapshot should have 0 password resets")
	assert.Equal(t, int64(1), snapshot2.PasswordResets, "Second snapshot should have 1 password reset")
}
