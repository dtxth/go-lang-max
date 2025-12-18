package usecase

import (
	"context"

	"maxbot-service/internal/domain"
)

type MaxBotService struct {
	apiClient              domain.MaxAPIClient
	normalizePhoneUC       *NormalizePhoneUseCase
	batchGetUsersByPhoneUC *BatchGetUsersByPhoneUseCase
}

func NewMaxBotService(apiClient domain.MaxAPIClient) *MaxBotService {
	return &MaxBotService{
		apiClient:              apiClient,
		normalizePhoneUC:       NewNormalizePhoneUseCase(),
		batchGetUsersByPhoneUC: NewBatchGetUsersByPhoneUseCase(apiClient),
	}
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

func (s *MaxBotService) NormalizePhone(phone string) (string, error) {
	return s.normalizePhoneUC.Execute(phone)
}

func (s *MaxBotService) BatchGetUsersByPhone(ctx context.Context, phones []string) ([]*domain.UserPhoneMapping, error) {
	return s.batchGetUsersByPhoneUC.Execute(ctx, phones)
}

func (s *MaxBotService) GetMe(ctx context.Context) (*domain.BotInfo, error) {
	return s.apiClient.GetMe(ctx)
}

func (s *MaxBotService) GetUserProfileByPhone(ctx context.Context, phone string) (*domain.UserProfile, error) {
	return s.apiClient.GetUserProfileByPhone(ctx, phone)
}
