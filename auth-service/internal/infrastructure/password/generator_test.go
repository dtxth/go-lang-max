package password

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestPasswordLengthProperty tests Property 1: Password length requirement
// Feature: secure-password-management, Property 1: Password length requirement
// Validates: Requirements 1.1
func TestPasswordLengthProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	generator := NewSecurePasswordGenerator(12)

	properties.Property("generated passwords are at least 12 characters", prop.ForAll(
		func(length int) bool {
			password, err := generator.Generate(length)
			if err != nil {
				// If there's an error, it should be because length is too short
				return length < 12
			}
			return len(password) >= 12 && len(password) == length
		},
		gen.IntRange(12, 50), // Generate lengths from 12 to 50
	))

	properties.TestingRun(t)
}

// TestPasswordComplexityProperty tests Property 2: Password complexity requirement
// Feature: secure-password-management, Property 2: Password complexity requirement
// Validates: Requirements 1.2
func TestPasswordComplexityProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	generator := NewSecurePasswordGenerator(12)

	properties.Property("generated passwords contain all character types", prop.ForAll(
		func(length int) bool {
			password, err := generator.Generate(length)
			if err != nil {
				return false
			}
			return HasUppercase(password) &&
				HasLowercase(password) &&
				HasDigit(password) &&
				HasSpecial(password)
		},
		gen.IntRange(12, 50),
	))

	properties.TestingRun(t)
}

// TestPasswordUniquenessProperty tests Property 3: Password uniqueness
// Feature: secure-password-management, Property 3: Password uniqueness
// Validates: Requirements 1.4
func TestPasswordUniquenessProperty(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	generator := NewSecurePasswordGenerator(12)

	properties.Property("generated passwords are unique", prop.ForAll(
		func(length int) bool {
			passwords := make(map[string]bool)
			// Generate 100 passwords and check for duplicates
			for i := 0; i < 100; i++ {
				password, err := generator.Generate(length)
				if err != nil {
					return false
				}
				if passwords[password] {
					// Duplicate found
					return false
				}
				passwords[password] = true
			}
			return true
		},
		gen.IntRange(12, 50),
	))

	properties.TestingRun(t)
}
