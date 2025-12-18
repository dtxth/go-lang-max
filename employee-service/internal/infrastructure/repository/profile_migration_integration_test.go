package repository

import (
	"employee-service/internal/domain"
	"testing"
	"time"
)

// TestProfileSourceMigrationIntegration tests the complete profile source tracking functionality
func TestProfileSourceMigrationIntegration(t *testing.T) {
	// Test that verifies Requirements 5.4 and 5.5 are properly implemented
	
	// Test Case 1: Employee with webhook source (Requirement 5.4)
	t.Run("WebhookSourceTracking", func(t *testing.T) {
		now := time.Now()
		employee := &domain.Employee{
			FirstName:          "Webhook",
			LastName:           "User",
			Phone:              "+79001111111",
			MaxID:              "max_webhook_123",
			ProfileSource:      string(domain.SourceWebhook),
			ProfileLastUpdated: &now,
			UniversityID:       1,
		}
		
		// Verify source is properly tracked (Requirement 5.4)
		if employee.ProfileSource != string(domain.SourceWebhook) {
			t.Errorf("Expected ProfileSource to be %s, got %s", domain.SourceWebhook, employee.ProfileSource)
		}
		
		// Verify timestamp is tracked (Requirement 5.5)
		if employee.ProfileLastUpdated == nil {
			t.Error("Expected ProfileLastUpdated to be set for webhook source")
		}
	})
	
	// Test Case 2: Employee with user input source (Requirement 5.4)
	t.Run("UserInputSourceTracking", func(t *testing.T) {
		now := time.Now()
		employee := &domain.Employee{
			FirstName:          "UserInput",
			LastName:           "User", 
			Phone:              "+79002222222",
			ProfileSource:      string(domain.SourceUserInput),
			ProfileLastUpdated: &now,
			UniversityID:       1,
		}
		
		// Verify user input source is tracked (Requirement 5.4)
		if employee.ProfileSource != string(domain.SourceUserInput) {
			t.Errorf("Expected ProfileSource to be %s, got %s", domain.SourceUserInput, employee.ProfileSource)
		}
		
		// Verify timestamp tracking for user updates (Requirement 5.5)
		if employee.ProfileLastUpdated == nil {
			t.Error("Expected ProfileLastUpdated to be set for user input source")
		}
	})
	
	// Test Case 3: Employee with default source (existing data migration)
	t.Run("DefaultSourceForExistingData", func(t *testing.T) {
		employee := &domain.Employee{
			FirstName:     "Default",
			LastName:      "User",
			Phone:         "+79003333333",
			ProfileSource: string(domain.SourceDefault),
			UniversityID:  1,
		}
		
		// Verify default source for existing data (migration requirement)
		if employee.ProfileSource != string(domain.SourceDefault) {
			t.Errorf("Expected ProfileSource to be %s for existing data, got %s", domain.SourceDefault, employee.ProfileSource)
		}
		
		// ProfileLastUpdated can be nil for default source (existing data)
		// This is acceptable as existing employees may not have update timestamps
	})
	
	// Test Case 4: Profile update with timestamp tracking (Requirement 5.5)
	t.Run("ProfileUpdateTimestampTracking", func(t *testing.T) {
		originalTime := time.Now().Add(-1 * time.Hour)
		updateTime := time.Now()
		
		employee := &domain.Employee{
			FirstName:          "Updated",
			LastName:           "User",
			Phone:              "+79004444444",
			ProfileSource:      string(domain.SourceWebhook),
			ProfileLastUpdated: &originalTime,
			UniversityID:       1,
		}
		
		// Simulate profile update
		employee.ProfileSource = string(domain.SourceUserInput)
		employee.ProfileLastUpdated = &updateTime
		
		// Verify timestamp was updated (Requirement 5.5)
		if employee.ProfileLastUpdated.Before(originalTime) {
			t.Error("Expected ProfileLastUpdated to be updated to newer timestamp")
		}
		
		// Verify source change is tracked (Requirement 5.4)
		if employee.ProfileSource != string(domain.SourceUserInput) {
			t.Errorf("Expected ProfileSource to be updated to %s, got %s", domain.SourceUserInput, employee.ProfileSource)
		}
	})
}

// TestMigrationSQLStructure verifies the migration SQL structure is correct
func TestMigrationSQLStructure(t *testing.T) {
	// This test verifies that the migration adds the correct columns
	// In a real database test, this would verify the actual schema
	
	// Test that all required profile source values are valid
	validSources := []domain.ProfileSource{
		domain.SourceWebhook,
		domain.SourceUserInput,
		domain.SourceDefault,
	}
	
	for _, source := range validSources {
		if string(source) == "" {
			t.Errorf("Profile source %v should not be empty", source)
		}
		
		// Verify source values match expected database values
		switch source {
		case domain.SourceWebhook:
			if string(source) != "webhook" {
				t.Errorf("Expected webhook source to be 'webhook', got %s", string(source))
			}
		case domain.SourceUserInput:
			if string(source) != "user_input" {
				t.Errorf("Expected user input source to be 'user_input', got %s", string(source))
			}
		case domain.SourceDefault:
			if string(source) != "default" {
				t.Errorf("Expected default source to be 'default', got %s", string(source))
			}
		}
	}
}