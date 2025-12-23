package chat

import (
	"context"
	"migration-service/internal/domain"
)

// CompositeClient combines HTTP and gRPC clients for Chat Service
// Uses gRPC for all operations during migration
type CompositeClient struct {
	HTTPClient *HTTPClient
	GRPCClient interface {
		CreateChat(ctx context.Context, chat *domain.ChatData) (int, error)
		AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error
	}
}

// CreateChat uses gRPC client
func (c *CompositeClient) CreateChat(ctx context.Context, chat *domain.ChatData) (int, error) {
	return c.GRPCClient.CreateChat(ctx, chat)
}

// AddAdministrator uses gRPC client (without phone validation)
func (c *CompositeClient) AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error {
	// Using gRPC to skip phone validation during migration
	return c.GRPCClient.AddAdministrator(ctx, admin)
}
