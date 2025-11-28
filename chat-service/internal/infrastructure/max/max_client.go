package max

import (
	"context"
	"errors"
	"time"

	"chat-service/internal/domain"
	grpcretry "chat-service/internal/infrastructure/grpc"
	maxbotproto "maxbot-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MaxClient struct {
	conn    *grpc.ClientConn
	client  maxbotproto.MaxBotServiceClient
	timeout time.Duration
}

func NewMaxClient(address string, timeout time.Duration) (*MaxClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &MaxClient{
		conn:    conn,
		client:  maxbotproto.NewMaxBotServiceClient(conn),
		timeout: timeout,
	}, nil
}

func (c *MaxClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *MaxClient) GetMaxIDByPhone(phone string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var resp *maxbotproto.GetMaxIDByPhoneResponse
	err := grpcretry.WithRetry(ctx, "MaxBot.GetMaxIDByPhone", func() error {
		var callErr error
		resp, callErr = c.client.GetMaxIDByPhone(ctx, &maxbotproto.GetMaxIDByPhoneRequest{Phone: phone})
		return callErr
	})
	
	if err != nil {
		return "", err
	}

	if resp.Error != "" {
		return "", mapError(resp.ErrorCode, resp.Error)
	}

	return resp.MaxId, nil
}

func (c *MaxClient) ValidatePhone(phone string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var resp *maxbotproto.ValidatePhoneResponse
	err := grpcretry.WithRetry(ctx, "MaxBot.ValidatePhone", func() error {
		var callErr error
		resp, callErr = c.client.ValidatePhone(ctx, &maxbotproto.ValidatePhoneRequest{Phone: phone})
		return callErr
	})
	
	if err != nil {
		return false
	}

	if resp.Error != "" {
		return false
	}

	return resp.Valid
}

func mapError(code maxbotproto.ErrorCode, message string) error {
	switch code {
	case maxbotproto.ErrorCode_ERROR_CODE_INVALID_PHONE:
		return domain.ErrInvalidPhone
	case maxbotproto.ErrorCode_ERROR_CODE_MAX_ID_NOT_FOUND:
		return domain.ErrMaxIDNotFound
	default:
		return errors.New(message)
	}
}
