package maxbot

import (
	"context"
	"fmt"
	"time"

	"auth-service/internal/domain"
	maxbotproto "maxbot-service/api/proto/maxbotproto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client implements domain.MaxBotClient using gRPC
type Client struct {
	conn   *grpc.ClientConn
	client maxbotproto.MaxBotServiceClient
}

// NewClient creates a new MaxBot gRPC client
func NewClient(addr string) (*Client, error) {
	// Set up connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MaxBot service at %s: %w", addr, err)
	}

	client := maxbotproto.NewMaxBotServiceClient(conn)

	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Close closes the gRPC connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetBotInfo retrieves bot information from MaxBot service
func (c *Client) GetBotInfo(ctx context.Context) (*domain.BotInfo, error) {
	// Set timeout for the request
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &maxbotproto.GetMeRequest{}
	
	resp, err := c.client.GetMe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot info from MaxBot service: %w", err)
	}

	// Check for error in response
	if resp.ErrorCode != maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED {
		return nil, fmt.Errorf("MaxBot service error: %s", resp.Error)
	}

	if resp.Bot == nil {
		return nil, fmt.Errorf("empty bot info received from MaxBot service")
	}

	return &domain.BotInfo{
		Name:    resp.Bot.Name,
		AddLink: resp.Bot.AddLink,
	}, nil
}