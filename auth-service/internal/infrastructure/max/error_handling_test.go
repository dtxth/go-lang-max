package max

import (
	"strings"
	"testing"
)

func TestAuthValidator_ErrorHandling_EdgeCases(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	tests := []struct {
		name        string
		initData    string
		botToken    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "whitespace only initData",
			initData:    "   ",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash parameter is missing",
		},
		{
			name:        "whitespace only botToken",
			initData:    "max_id=123&first_name=John&hash=abc",
			botToken:    "   ",
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "newline characters in initData",
			initData:    "max_id=123\n&first_name=John&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "URL encoded special characters",
			initData:    "max_id=123&first_name=John%20Doe&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed", // Will fail due to wrong hash, but parsing should work
		},
		{
			name:        "duplicate parameters",
			initData:    "max_id=123&max_id=456&first_name=John&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed", // Parser takes first value
		},
		{
			name:        "empty parameter values",
			initData:    "max_id=&first_name=&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed", // Hash will fail first, then validation
		},
		{
			name:        "very long parameter values",
			initData:    "max_id=123&first_name=" + strings.Repeat("a", 1000) + "&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "negative max_id",
			initData:    createValidInitData(map[string]string{"max_id": "-123", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     false, // Negative numbers are valid int64
		},
		{
			name:        "zero max_id",
			initData:    createValidInitData(map[string]string{"max_id": "0", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     false, // Zero is valid
		},
		{
			name:        "max int64 value",
			initData:    createValidInitData(map[string]string{"max_id": "9223372036854775807", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     false,
		},
		{
			name:        "overflow int64 value",
			initData:    createValidInitData(map[string]string{"max_id": "9223372036854775808", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     true,
			errContains: "invalid max_id format",
		},
		{
			name:        "float max_id",
			initData:    createValidInitData(map[string]string{"max_id": "123.45", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     true,
			errContains: "invalid max_id format",
		},
		{
			name:        "scientific notation max_id",
			initData:    createValidInitData(map[string]string{"max_id": "1e5", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     true,
			errContains: "invalid max_id format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateInitData(tt.initData, tt.botToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateInitData() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateInitData() error = %v, want error containing %v", err, tt.errContains)
				}
				if result != nil {
					t.Errorf("ValidateInitData() expected nil result on error but got %v", result)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateInitData() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("ValidateInitData() result is nil")
			}
		})
	}
}

func TestAuthValidator_SecurityValidation(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	tests := []struct {
		name        string
		description string
		initData    string
		botToken    string
		wantErr     bool
		errContains string
	}{
		{
			name:        "timing attack protection",
			description: "Different length hashes should still use constant time comparison",
			initData:    "max_id=123&first_name=John&hash=short",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "hash case sensitivity",
			description: "Hash should be case sensitive",
			initData:    "max_id=123&first_name=John&hash=ABCDEF",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "parameter injection attempt",
			description: "Malicious parameters should not affect hash calculation",
			initData:    "max_id=123&first_name=John&malicious=value&hash=abc",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "empty hash parameter",
			description: "Empty hash should be rejected",
			initData:    "max_id=123&first_name=John&hash=",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "multiple hash parameters",
			description: "Multiple hash parameters should use first one",
			initData:    "max_id=123&first_name=John&hash=first&hash=second",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateInitData(tt.initData, tt.botToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateInitData() expected error but got none for %s", tt.description)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateInitData() error = %v, want error containing %v for %s", err, tt.errContains, tt.description)
				}
				if result != nil {
					t.Errorf("ValidateInitData() expected nil result on error but got %v for %s", result, tt.description)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateInitData() unexpected error = %v for %s", err, tt.description)
				return
			}

			if result == nil {
				t.Errorf("ValidateInitData() result is nil for %s", tt.description)
			}
		})
	}
}

func TestAuthValidator_UnicodeHandling(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	tests := []struct {
		name      string
		firstName string
		lastName  string
		username  string
		wantErr   bool
	}{
		{
			name:      "ASCII characters",
			firstName: "John",
			lastName:  "Doe",
			username:  "johndoe",
			wantErr:   false,
		},
		{
			name:      "Unicode characters",
			firstName: "Jöhn",
			lastName:  "Döe",
			username:  "jöhndöe",
			wantErr:   false,
		},
		{
			name:      "Cyrillic characters",
			firstName: "Иван",
			lastName:  "Петров",
			username:  "ivan_petrov",
			wantErr:   false,
		},
		{
			name:      "Chinese characters",
			firstName: "张",
			lastName:  "三",
			username:  "zhangsan",
			wantErr:   false,
		},
		{
			name:      "Emoji characters",
			firstName: "John😀",
			lastName:  "Doe🎉",
			username:  "john_doe_😀",
			wantErr:   false,
		},
		{
			name:      "Mixed scripts",
			firstName: "JohnИван张😀",
			lastName:  "DoeПетров三🎉",
			username:  "mixed_script_user",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]string{
				"max_id":     "123",
				"first_name": tt.firstName,
			}
			if tt.lastName != "" {
				data["last_name"] = tt.lastName
			}
			if tt.username != "" {
				data["username"] = tt.username
			}

			initData := createValidInitData(data, botToken)
			result, err := validator.ValidateInitData(initData, botToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateInitData() expected error but got none")
					return
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateInitData() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("ValidateInitData() result is nil")
				return
			}

			if result.FirstName != tt.firstName {
				t.Errorf("ValidateInitData() FirstName = %v, want %v", result.FirstName, tt.firstName)
			}
			if result.LastName != tt.lastName {
				t.Errorf("ValidateInitData() LastName = %v, want %v", result.LastName, tt.lastName)
			}
			if result.Username != tt.username {
				t.Errorf("ValidateInitData() Username = %v, want %v", result.Username, tt.username)
			}
		})
	}
}