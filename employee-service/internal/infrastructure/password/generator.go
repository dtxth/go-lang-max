package password

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// SecurePasswordGenerator implements domain.PasswordGenerator using crypto/rand
type SecurePasswordGenerator struct {
	minLength int
}

// NewSecurePasswordGenerator creates a new SecurePasswordGenerator with the specified minimum length
func NewSecurePasswordGenerator(minLength int) *SecurePasswordGenerator {
	return &SecurePasswordGenerator{
		minLength: minLength,
	}
}

const (
	uppercaseChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowercaseChars = "abcdefghijklmnopqrstuvwxyz"
	digitChars     = "0123456789"
	specialChars   = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	allChars       = uppercaseChars + lowercaseChars + digitChars + specialChars
)

// Generate creates a cryptographically secure random password
func (g *SecurePasswordGenerator) Generate(length int) (string, error) {
	if length < g.minLength {
		return "", fmt.Errorf("password length must be at least %d characters", g.minLength)
	}

	if length < 4 {
		return "", fmt.Errorf("password length must be at least 4 characters to meet complexity requirements")
	}

	// Generate one character from each required type
	password := make([]byte, 0, length)

	// Add one uppercase
	char, err := randomChar(uppercaseChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate uppercase character: %w", err)
	}
	password = append(password, char)

	// Add one lowercase
	char, err = randomChar(lowercaseChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate lowercase character: %w", err)
	}
	password = append(password, char)

	// Add one digit
	char, err = randomChar(digitChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate digit character: %w", err)
	}
	password = append(password, char)

	// Add one special character
	char, err = randomChar(specialChars)
	if err != nil {
		return "", fmt.Errorf("failed to generate special character: %w", err)
	}
	password = append(password, char)

	// Fill the rest with random characters from all types
	for i := 4; i < length; i++ {
		char, err := randomChar(allChars)
		if err != nil {
			return "", fmt.Errorf("failed to generate random character: %w", err)
		}
		password = append(password, char)
	}

	// Shuffle the password to avoid predictable patterns
	if err := shuffle(password); err != nil {
		return "", fmt.Errorf("failed to shuffle password: %w", err)
	}

	return string(password), nil
}

// randomChar returns a random character from the given character set
func randomChar(charset string) (byte, error) {
	if len(charset) == 0 {
		return 0, fmt.Errorf("charset cannot be empty")
	}

	max := big.NewInt(int64(len(charset)))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return 0, err
	}

	return charset[n.Int64()], nil
}

// shuffle randomly shuffles the bytes in place using Fisher-Yates algorithm
func shuffle(data []byte) error {
	n := len(data)
	for i := n - 1; i > 0; i-- {
		max := big.NewInt(int64(i + 1))
		j, err := rand.Int(rand.Reader, max)
		if err != nil {
			return err
		}
		jInt := int(j.Int64())
		data[i], data[jInt] = data[jInt], data[i]
	}
	return nil
}
