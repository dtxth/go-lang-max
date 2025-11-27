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

func (s *MaxBotService) SendMessage(ctx context.Context, chatID, userID int64, text string) (string, error) {
	return s.apiClient.SendMessage(ctx, chatID, userID, text)
}

func (s *MaxBotService) SendNotification(ctx context.Context, phone, text string) error {
	return s.apiClient.SendNotification(ctx, phone, text)
}

func (s *MaxBotService) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	return s.apiClient.GetChatInfo(ctx, chatID)
}

func (s *MaxBotService) GetChatMembers(ctx context.Context, chatID int64, limit int, marker int64) (*domain.ChatMembersList, error) {
	return s.apiClient.GetChatMembers(ctx, chatID, limit, marker)
}

func (s *MaxBotService) GetChatAdmins(ctx context.Context, chatID int64) ([]*domain.ChatMember, error) {
	return s.apiClient.GetChatAdmins(ctx, chatID)
}

func (s *MaxBotService) CheckPhoneNumbers(ctx context.Context, phones []string) ([]string, error) {
	return s.apiClient.CheckPhoneNumbers(ctx, phones)
}
