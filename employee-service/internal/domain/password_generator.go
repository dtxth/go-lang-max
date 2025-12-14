package domain

// PasswordGenerator defines the interface for password generation
type PasswordGenerator interface {
	// Generate creates a cryptographically secure random password
	Generate(length int) (string, error)
}
