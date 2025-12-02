package grpc

import (
	"context"
	"log"
	chatproto "structure-service/api/proto/chat"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ChatClient struct {
	conn   *grpc.ClientConn
	client chatproto.ChatServiceClient
}

func NewChatClient(address string) (*ChatClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &ChatClient{
		conn:   conn,
		client: chatproto.NewChatServiceClient(conn),
	}, nil
}

func (c *ChatClient) Close() error {
	return c.conn.Close()
}

func (c *ChatClient) GetChatByID(ctx context.Context, chatID int64) (*chatproto.Chat, error) {
	resp, err := c.client.GetChatByID(ctx, &chatproto.GetChatByIDRequest{Id: chatID})
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		log.Printf("Error from chat service: %s", resp.Error)
		return nil, nil
	}
	return resp.Chat, nil
}

func (c *ChatClient) CreateChat(ctx context.Context, name, url, maxChatID, source string, participantsCount int32, universityID *int64, department string) (*chatproto.Chat, error) {
	req := &chatproto.CreateChatRequest{
		Name:              name,
		Url:               url,
		MaxChatId:         maxChatID,
		Source:            source,
		ParticipantsCount: participantsCount,
		Department:        department,
	}
	if universityID != nil {
		req.UniversityId = universityID
	}

	resp, err := c.client.CreateChat(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		log.Printf("Error from chat service: %s", resp.Error)
		return nil, nil
	}
	return resp.Chat, nil
}
