package usecase

import (
	"testing"

	"maxbot-service/internal/domain"
)

func TestNormalizePhoneUseCase_Execute(t *testing.T) {
	uc := NewNormalizePhoneUseCase()

	tests := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "Russian phone starting with 8",
			input:       "89991234567",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Russian phone starting with 9",
			input:       "9991234567",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Russian phone with +7",
			input:       "+79991234567",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Russian phone starting with 7",
			input:       "79991234567",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Phone with spaces and dashes",
			input:       "+7 (999) 123-45-67",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Phone with parentheses",
			input:       "8(999)1234567",
			expected:    "+79991234567",
			expectError: false,
		},
		{
			name:        "Empty phone",
			input:       "",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid phone - too short",
			input:       "123",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Invalid phone - only non-digits",
			input:       "abc-def-ghij",
			expected:    "",
			expectError: true,
		},
		{
			name:        "Phone with only spaces",
			input:       "   ",
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := uc.Execute(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				if err != domain.ErrInvalidPhone {
					t.Errorf("Expected ErrInvalidPhone but got: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if result != tt.expected {
					t.Errorf("Expected %s but got %s", tt.expected, result)
				}
			}
		})
	}
}

func TestNormalizePhoneUseCase_E164Validation(t *testing.T) {
	uc := NewNormalizePhoneUseCase()

	// Test that all normalized phones are in E.164 format
	validInputs := []string{
		"89991234567",
		"9991234567",
		"+79991234567",
		"79991234567",
		"+7 999 123 45 67",
		"8 (999) 123-45-67",
	}

	for _, input := range validInputs {
		result, err := uc.Execute(input)
		if err != nil {
			t.Errorf("Unexpected error for input %s: %v", input, err)
			continue
		}

		// Verify E.164 format
		if len(result) != 12 {
			t.Errorf("E.164 format should be 12 characters, got %d for input %s", len(result), input)
		}
		if result[0] != '+' || result[1] != '7' {
			t.Errorf("E.164 format should start with +7, got %s for input %s", result[:2], input)
		}
		// Verify all remaining characters are digits
		for i := 2; i < len(result); i++ {
			if result[i] < '0' || result[i] > '9' {
				t.Errorf("E.164 format should only contain digits after +7, got %c at position %d for input %s", result[i], i, input)
			}
		}
	}
}
