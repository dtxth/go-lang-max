package usecase

import (
	"regexp"
	"strings"

	"maxbot-service/internal/domain"
)

// NormalizePhoneUseCase handles phone number normalization to E.164 format
type NormalizePhoneUseCase struct{}

// NewNormalizePhoneUseCase creates a new instance of NormalizePhoneUseCase
func NewNormalizePhoneUseCase() *NormalizePhoneUseCase {
	return &NormalizePhoneUseCase{}
}

// Execute normalizes a phone number to E.164 format
// Handles Russian phone formats:
// - Starting with 8: replace with +7
// - Starting with 9: prepend +7
// - Starting with +7: keep as is
// - Removes all non-digit characters except leading +
func (uc *NormalizePhoneUseCase) Execute(phone string) (string, error) {
	if phone == "" {
		return "", domain.ErrInvalidPhone
	}

	// Remove all whitespace, dashes, parentheses, and other non-digit characters
	// but preserve the leading + if present
	hasPlus := strings.HasPrefix(phone, "+")
	
	// Remove all non-digit characters
	re := regexp.MustCompile(`[^\d]`)
	cleaned := re.ReplaceAllString(phone, "")
	
	if cleaned == "" {
		return "", domain.ErrInvalidPhone
	}

	// Handle Russian phone formats
	var normalized string
	
	if hasPlus && strings.HasPrefix(phone, "+7") {
		// Already has +7, just use cleaned digits
		if len(cleaned) == 11 && cleaned[0] == '7' {
			normalized = "+" + cleaned
		} else if len(cleaned) == 10 {
			// +7 was removed, add it back
			normalized = "+7" + cleaned
		} else {
			return "", domain.ErrInvalidPhone
		}
	} else if cleaned[0] == '8' && len(cleaned) == 11 {
		// Replace leading 8 with +7
		normalized = "+7" + cleaned[1:]
	} else if cleaned[0] == '9' && len(cleaned) == 10 {
		// Prepend +7 to 10-digit number starting with 9
		normalized = "+7" + cleaned
	} else if cleaned[0] == '7' && len(cleaned) == 11 {
		// Already starts with 7, just add +
		normalized = "+" + cleaned
	} else {
		return "", domain.ErrInvalidPhone
	}

	// Validate E.164 format: +7 followed by 10 digits
	if !isValidE164(normalized) {
		return "", domain.ErrInvalidPhone
	}

	return normalized, nil
}

// isValidE164 validates that the phone is in E.164 format for Russian numbers
func isValidE164(phone string) bool {
	// E.164 format for Russia: +7XXXXXXXXXX (12 characters total)
	if len(phone) != 12 {
		return false
	}
	
	if !strings.HasPrefix(phone, "+7") {
		return false
	}
	
	// Check that all characters after +7 are digits
	for i := 2; i < len(phone); i++ {
		if phone[i] < '0' || phone[i] > '9' {
			return false
		}
	}
	
	return true
}
