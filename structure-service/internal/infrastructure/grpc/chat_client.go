package grpc

import (
	"context"
	"log"

	chatproto "chat-service/api/proto"
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
	var resp *chatproto.GetChatByIDResponse
	err := WithRetry(ctx, "Chat.GetChatByID", func() error {
		var callErr error
		resp, callErr = c.client.GetChatByID(ctx, &chatproto.GetChatByIDRequest{Id: chatID})
		return callErr
	})
	
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

	var resp *chatproto.CreateChatResponse
	err := WithRetry(ctx, "Chat.CreateChat", func() error {
		var callErr error
		resp, callErr = c.client.CreateChat(ctx, req)
		return callErr
	})
	
	if err != nil {
		return nil, err
	}
	if resp.Error != "" {
		log.Printf("Error from chat service: %s", resp.Error)
		return nil, nil
	}
	return resp.Chat, nil
}
