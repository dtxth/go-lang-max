package utils

import (
	"testing"
)

func TestPhoneValidator_ValidatePhone(t *testing.T) {
	validator := NewPhoneValidator()

	tests := []struct {
		name     string
		phone    string
		expected bool
	}{
		{
			name:     "Valid Russian phone with +7",
			phone:    "+79001234567",
			expected: true,
		},
		{
			name:     "Valid Russian phone starting with 7",
			phone:    "79001234567",
			expected: true,
		},
		{
			name:     "Valid Russian phone starting with 8",
			phone:    "89001234567",
			expected: true,
		},
		{
			name:     "Valid phone with spaces and dashes",
			phone:    "+7 900 123-45-67",
			expected: true,
		},
		{
			name:     "Valid phone with parentheses",
			phone:    "+7 (900) 123-45-67",
			expected: true,
		},
		{
			name:     "Invalid phone - too short",
			phone:    "+790012345",
			expected: false,
		},
		{
			name:     "Invalid phone - too long",
			phone:    "+790012345678",
			expected: false,
		},
		{
			name:     "Invalid phone - wrong country code",
			phone:    "+19001234567",
			expected: false,
		},
		{
			name:     "Empty phone",
			phone:    "",
			expected: false,
		},
		{
			name:     "Invalid phone - letters",
			phone:    "+7900abc4567",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ValidatePhone(tt.phone)
			if result != tt.expected {
				t.Errorf("ValidatePhone(%s) = %v, expected %v", tt.phone, result, tt.expected)
			}
		})
	}
}

func TestPhoneValidator_NormalizePhone(t *testing.T) {
	validator := NewPhoneValidator()

	tests := []struct {
		name     string
		phone    string
		expected string
	}{
		{
			name:     "Already normalized phone",
			phone:    "+79001234567",
			expected: "+79001234567",
		},
		{
			name:     "Phone starting with 8",
			phone:    "89001234567",
			expected: "+79001234567",
		},
		{
			name:     "Phone starting with 7",
			phone:    "79001234567",
			expected: "+79001234567",
		},
		{
			name:     "Phone with spaces and dashes",
			phone:    "8 900 123-45-67",
			expected: "+79001234567",
		},
		{
			name:     "Phone with parentheses",
			phone:    "8 (900) 123-45-67",
			expected: "+79001234567",
		},
		{
			name:     "Empty phone",
			phone:    "",
			expected: "",
		},
		{
			name:     "Invalid phone - return as is",
			phone:    "+19001234567",
			expected: "+19001234567",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.NormalizePhone(tt.phone)
			if result != tt.expected {
				t.Errorf("NormalizePhone(%s) = %s, expected %s", tt.phone, result, tt.expected)
			}
		})
	}
}