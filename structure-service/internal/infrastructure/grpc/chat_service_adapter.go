package grpc

import (
	"context"
	"structure-service/internal/domain"
)

// ChatServiceAdapter adapts the gRPC chat client to the domain ChatService interface
type ChatServiceAdapter struct {
	client *ChatClient
}

// NewChatServiceAdapter creates a new ChatServiceAdapter
func NewChatServiceAdapter(client *ChatClient) *ChatServiceAdapter {
	return &ChatServiceAdapter{
		client: client,
	}
}

// GetChatByID retrieves chat details from the chat service via gRPC
func (a *ChatServiceAdapter) GetChatByID(ctx context.Context, chatID int64) (*domain.Chat, error) {
	chatProto, err := a.client.GetChatByID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	if chatProto == nil {
		return nil, nil
	}

	// Convert proto Chat to domain Chat
	chat := &domain.Chat{
		ID:    chatProto.Id,
		Name:  chatProto.Name,
		URL:   chatProto.Url,
		MaxID: chatProto.MaxChatId,
	}

	return chat, nil
}
