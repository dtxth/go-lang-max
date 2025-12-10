package phone

import (
	"testing"
)

func TestNormalizePhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "already normalized with +7",
			input:    "+79001234567",
			expected: "+79001234567",
		},
		{
			name:     "starts with 7 without +",
			input:    "79001234567",
			expected: "+79001234567",
		},
		{
			name:     "starts with 8",
			input:    "89001234567",
			expected: "+79001234567",
		},
		{
			name:     "starts with 9 (10 digits)",
			input:    "9001234567",
			expected: "+79001234567",
		},
		{
			name:     "with spaces and dashes",
			input:    "+7 900 123-45-67",
			expected: "+79001234567",
		},
		{
			name:     "with parentheses",
			input:    "+7 (900) 123-45-67",
			expected: "+79001234567",
		},
		{
			name:     "8 with formatting",
			input:    "8 (900) 123-45-67",
			expected: "+79001234567",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "invalid format unchanged",
			input:    "123456",
			expected: "123456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizePhone(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizePhone(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsValidRussianPhone(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid +7 format",
			input:    "+79001234567",
			expected: true,
		},
		{
			name:     "valid 7 format",
			input:    "79001234567",
			expected: true,
		},
		{
			name:     "valid 8 format",
			input:    "89001234567",
			expected: true,
		},
		{
			name:     "valid 9 format",
			input:    "9001234567",
			expected: true,
		},
		{
			name:     "invalid - too short",
			input:    "900123456",
			expected: false,
		},
		{
			name:     "invalid - too long",
			input:    "790012345678",
			expected: false,
		},
		{
			name:     "invalid - wrong country code",
			input:    "+19001234567",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidRussianPhone(tt.input)
			if result != tt.expected {
				t.Errorf("IsValidRussianPhone(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}