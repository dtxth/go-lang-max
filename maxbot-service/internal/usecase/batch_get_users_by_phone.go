package usecase

import (
	"context"
	"fmt"

	"maxbot-service/internal/domain"
)

const MaxBatchSize = 100

// BatchGetUsersByPhoneUseCase handles batch retrieval of MAX IDs by phone numbers
type BatchGetUsersByPhoneUseCase struct {
	apiClient        domain.MaxAPIClient
	normalizePhoneUC *NormalizePhoneUseCase
}

// NewBatchGetUsersByPhoneUseCase creates a new instance of BatchGetUsersByPhoneUseCase
func NewBatchGetUsersByPhoneUseCase(apiClient domain.MaxAPIClient) *BatchGetUsersByPhoneUseCase {
	return &BatchGetUsersByPhoneUseCase{
		apiClient:        apiClient,
		normalizePhoneUC: NewNormalizePhoneUseCase(),
	}
}

// Execute processes a batch of phone numbers and returns MAX ID mappings
// Validates that batch size does not exceed 100 phones
func (uc *BatchGetUsersByPhoneUseCase) Execute(ctx context.Context, phones []string) ([]*domain.UserPhoneMapping, error) {
	if len(phones) == 0 {
		return []*domain.UserPhoneMapping{}, nil
	}

	if len(phones) > MaxBatchSize {
		return nil, fmt.Errorf("batch size exceeds maximum of %d phones", MaxBatchSize)
	}

	// Normalize all phone numbers first and preserve original phones
	type phoneEntry struct {
		original   string
		normalized string
	}
	
	phoneEntries := make([]phoneEntry, 0, len(phones))
	
	for _, phone := range phones {
		normalized, err := uc.normalizePhoneUC.Execute(phone)
		if err != nil {
			// Skip invalid phones but continue processing
			continue
		}
		phoneEntries = append(phoneEntries, phoneEntry{
			original:   phone,
			normalized: normalized,
		})
	}

	if len(phoneEntries) == 0 {
		return []*domain.UserPhoneMapping{}, nil
	}

	// Get MAX IDs for normalized phones
	mappings := make([]*domain.UserPhoneMapping, 0, len(phoneEntries))
	
	for _, entry := range phoneEntries {
		maxID, err := uc.apiClient.GetMaxIDByPhone(ctx, entry.normalized)
		
		if err != nil {
			// Phone not found or error occurred
			mappings = append(mappings, &domain.UserPhoneMapping{
				Phone: entry.original,
				MaxID: "",
				Found: false,
			})
		} else {
			mappings = append(mappings, &domain.UserPhoneMapping{
				Phone: entry.original,
				MaxID: maxID,
				Found: true,
			})
		}
	}

	return mappings, nil
}
