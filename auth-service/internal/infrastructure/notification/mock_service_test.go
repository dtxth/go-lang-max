package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"auth-service/internal/infrastructure/logger"
)

func TestMockNotificationService_SendPasswordNotification(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		password string
	}{
		{
			name:     "sends notification with sanitized phone",
			phone:    "+71234567890",
			password: "SecurePass123!",
		},
		{
			name:     "handles short phone number",
			phone:    "123",
			password: "AnotherPass456!",
		},
		{
			name:     "handles long phone number",
			phone:    "+7123456789012345",
			password: "LongPhonePass789!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer
			log := logger.New(&buf, logger.INFO)
			
			service := NewMockNotificationService(log)
			ctx := context.Background()
			
			err := service.SendPasswordNotification(ctx, tt.phone, tt.password)
			if err != nil {
				t.Errorf("SendPasswordNotification() error = %v, want nil", err)
			}
			
			// Parse the log output
			logOutput := buf.String()
			
			// Test that password is NOT in the logs
			if strings.Contains(logOutput, tt.password) {
				t.Errorf("Log output contains plaintext password: %s", tt.password)
			}
			
			// Test that the log contains the expected action
			if !strings.Contains(logOutput, "send_password_notification") {
				t.Errorf("Log output missing action field")
			}
			
			// Test that phone is sanitized (only last 4 digits shown)
			var logEntry logger.LogEntry
			if err := json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry); err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}
			
			phoneSuffix, ok := logEntry.Fields["phone_suffix"].(string)
			if !ok {
				t.Fatal("phone_suffix field not found in log")
			}
			
			// Verify phone is sanitized
			if !strings.HasPrefix(phoneSuffix, "****") {
				t.Errorf("Phone not properly sanitized, got: %s", phoneSuffix)
			}
			
			// Verify full phone number is not in logs
			if strings.Contains(logOutput, tt.phone) {
				t.Errorf("Log output contains full phone number: %s", tt.phone)
			}
		})
	}
}

func TestMockNotificationService_SendResetTokenNotification(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		token string
	}{
		{
			name:  "sends reset token notification with sanitized phone",
			phone: "+71234567890",
			token: "reset-token-abc123",
		},
		{
			name:  "handles short phone number",
			phone: "123",
			token: "short-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a buffer to capture log output
			var buf bytes.Buffer
			log := logger.New(&buf, logger.INFO)
			
			service := NewMockNotificationService(log)
			ctx := context.Background()
			
			err := service.SendResetTokenNotification(ctx, tt.phone, tt.token)
			if err != nil {
				t.Errorf("SendResetTokenNotification() error = %v, want nil", err)
			}
			
			// Parse the log output
			logOutput := buf.String()
			
			// Test that token is NOT in the logs
			if strings.Contains(logOutput, tt.token) {
				t.Errorf("Log output contains plaintext token: %s", tt.token)
			}
			
			// Test that the log contains the expected action
			if !strings.Contains(logOutput, "send_reset_token_notification") {
				t.Errorf("Log output missing action field")
			}
			
			// Test that phone is sanitized
			var logEntry logger.LogEntry
			if err := json.Unmarshal([]byte(strings.TrimSpace(logOutput)), &logEntry); err != nil {
				t.Fatalf("Failed to parse log output: %v", err)
			}
			
			phoneSuffix, ok := logEntry.Fields["phone_suffix"].(string)
			if !ok {
				t.Fatal("phone_suffix field not found in log")
			}
			
			// Verify phone is sanitized
			if !strings.HasPrefix(phoneSuffix, "****") {
				t.Errorf("Phone not properly sanitized, got: %s", phoneSuffix)
			}
			
			// Verify full phone number is not in logs
			if strings.Contains(logOutput, tt.phone) {
				t.Errorf("Log output contains full phone number: %s", tt.phone)
			}
		})
	}
}

func TestSanitizePhone(t *testing.T) {
	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "sanitizes normal phone number",
			phone:    "+71234567890",
			expected: "****7890",
		},
		{
			name:     "handles short phone number",
			phone:    "123",
			expected: "****",
		},
		{
			name:     "handles exactly 4 digits",
			phone:    "1234",
			expected: "****",
		},
		{
			name:     "handles 5 digits",
			phone:    "12345",
			expected: "****2345",
		},
		{
			name:     "handles empty string",
			phone:    "",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizePhone(tt.phone)
			if result != tt.expected {
				t.Errorf("sanitizePhone(%s) = %s, want %s", tt.phone, result, tt.expected)
			}
		})
	}
}
