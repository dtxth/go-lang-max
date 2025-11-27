package domain

import "context"

type MaxAPIClient interface {
	GetMaxIDByPhone(ctx context.Context, phone string) (string, error)
	ValidatePhone(phone string) (bool, string, error)
}
