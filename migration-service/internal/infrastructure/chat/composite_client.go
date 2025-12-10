package chat

import (
	"context"
	"migration-service/internal/domain"
)

// CompositeClient combines HTTP and gRPC clients for Chat Service
// Uses gRPC for administrator operations (no phone validation) and HTTP for other operations
type CompositeClient struct {
	HTTPClient *HTTPClient
	GRPCClient interface {
		AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error
	}
}

// CreateChat uses HTTP client
func (c *CompositeClient) CreateChat(ctx context.Context, chat *domain.ChatData) (int, error) {
	return c.HTTPClient.CreateChat(ctx, chat)
}

// AddAdministrator uses gRPC client (without phone validation)
func (c *CompositeClient) AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error {
	// Using gRPC to skip phone validation during migration
	return c.GRPCClient.AddAdministrator(ctx, admin)
}
