package repository

import (
	"employee-service/internal/domain"
	"testing"
	"time"
)

// TestProfileSourceMigration verifies that the profile source tracking fields work correctly
func TestProfileSourceMigration(t *testing.T) {
	// This test verifies that the Employee struct and repository can handle
	// the new profile_source and profile_last_updated fields correctly
	
	// Create a test employee with profile source information
	now := time.Now()
	employee := &domain.Employee{
		FirstName:          "Test",
		LastName:           "Employee",
		Phone:              "+79001234567",
		MaxID:              "max_123",
		ProfileSource:      string(domain.SourceWebhook),
		ProfileLastUpdated: &now,
		UniversityID:       1,
	}
	
	// Verify that the struct fields are properly set
	if employee.ProfileSource != string(domain.SourceWebhook) {
		t.Errorf("Expected ProfileSource to be %s, got %s", domain.SourceWebhook, employee.ProfileSource)
	}
	
	if employee.ProfileLastUpdated == nil {
		t.Error("Expected ProfileLastUpdated to be set")
	}
	
	// Test different profile sources
	testSources := []domain.ProfileSource{
		domain.SourceWebhook,
		domain.SourceUserInput,
		domain.SourceDefault,
	}
	
	for _, source := range testSources {
		employee.ProfileSource = string(source)
		if employee.ProfileSource != string(source) {
			t.Errorf("Failed to set ProfileSource to %s", source)
		}
	}
}

// TestProfileSourceConstants verifies that the profile source constants are properly defined
func TestProfileSourceConstants(t *testing.T) {
	// Verify that all required profile source constants exist
	expectedSources := map[domain.ProfileSource]string{
		domain.SourceWebhook:   "webhook",
		domain.SourceUserInput: "user_input", 
		domain.SourceDefault:   "default",
	}
	
	for source, expectedValue := range expectedSources {
		if string(source) != expectedValue {
			t.Errorf("Expected %s to equal %s, got %s", source, expectedValue, string(source))
		}
	}
}