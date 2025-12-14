package domain

import "context"

// University represents a university entity
type University struct {
	ID   int
	Name string
	INN  string
	KPP  string
}

// UniversityRepository defines the interface for university persistence
type UniversityRepository interface {
	// GetByINN retrieves a university by INN
	GetByINN(ctx context.Context, inn string) (*University, error)

	// GetByINNAndKPP retrieves a university by INN and KPP
	GetByINNAndKPP(ctx context.Context, inn, kpp string) (*University, error)

	// Create creates a new university
	Create(ctx context.Context, university *University) error
}
