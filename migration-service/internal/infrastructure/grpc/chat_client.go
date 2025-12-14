package grpc

import (
	"context"
	"fmt"
	"migration-service/internal/domain"
	"time"

	chatpb "migration-service/api/proto/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ChatClient implements ChatService using gRPC
type ChatClient struct {
	client chatpb.ChatServiceClient
	conn   *grpc.ClientConn
}

// NewChatClient creates a new gRPC client for Chat Service
func NewChatClient(address string) (*ChatClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to chat service: %w", err)
	}

	client := chatpb.NewChatServiceClient(conn)

	return &ChatClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close closes the gRPC connection
func (c *ChatClient) Close() error {
	return c.conn.Close()
}

// CreateChat creates a new chat
func (c *ChatClient) CreateChat(ctx context.Context, chat *domain.ChatData) (int, error) {
	req := &chatpb.CreateChatRequest{
		Name:              chat.Name,
		Url:               chat.URL,
		MaxChatId:         chat.ExternalChatID,
		Source:            chat.Source,
		ParticipantsCount: 0,
		Department:        "",
	}

	if chat.UniversityID != nil && *chat.UniversityID != 0 {
		universityID := int64(*chat.UniversityID)
		req.UniversityId = &universityID
	}

	resp, err := c.client.CreateChat(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("failed to create chat: %w", err)
	}

	if resp.Error != "" {
		return 0, fmt.Errorf("chat service error: %s", resp.Error)
	}

	return int(resp.Chat.Id), nil
}

// AddAdministrator adds an administrator to a chat using gRPC (without phone validation for migration)
func (c *ChatClient) AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error {
	req := &chatpb.AddAdministratorForMigrationRequest{
		ChatId:   int64(admin.ChatID),
		Phone:    admin.Phone,
		MaxId:    admin.MaxID,
		AddUser:  admin.AddUser,
		AddAdmin: admin.AddAdmin,
	}

	resp, err := c.client.AddAdministratorForMigration(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC call failed: %w", err)
	}

	if resp.Error != "" {
		return fmt.Errorf("chat service returned error: %s", resp.Error)
	}

	return nil
}
