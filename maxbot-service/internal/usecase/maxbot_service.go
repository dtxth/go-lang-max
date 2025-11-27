package usecase

import (
	"context"

	"maxbot-service/internal/domain"
)

type MaxBotService struct {
	apiClient domain.MaxAPIClient
}

func NewMaxBotService(apiClient domain.MaxAPIClient) *MaxBotService {
	return &MaxBotService{apiClient: apiClient}
}

func (s *MaxBotService) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
	return s.apiClient.GetMaxIDByPhone(ctx, phone)
}

func (s *MaxBotService) ValidatePhone(phone string) (bool, string, error) {
	return s.apiClient.ValidatePhone(phone)
}
