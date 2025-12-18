package maxbot

import (
	"context"

	"auth-service/internal/domain"
)

// MockClient implements domain.MaxBotClient for testing
type MockClient struct {
	botInfo *domain.BotInfo
	err     error
}

// NewMockClient creates a new mock MaxBot client
func NewMockClient() *MockClient {
	return &MockClient{
		botInfo: &domain.BotInfo{
			Name:    "Digital University Bot",
			AddLink: "https://max.ru/bot/digital_university_bot",
		},
	}
}

// SetBotInfo sets the bot info to return
func (m *MockClient) SetBotInfo(botInfo *domain.BotInfo) {
	m.botInfo = botInfo
}

// SetError sets the error to return
func (m *MockClient) SetError(err error) {
	m.err = err
}

// GetBotInfo returns mock bot information
func (m *MockClient) GetBotInfo(ctx context.Context) (*domain.BotInfo, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.botInfo, nil
}