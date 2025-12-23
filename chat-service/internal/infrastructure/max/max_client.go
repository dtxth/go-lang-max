package max

import (
	"context"
	"errors"
	"time"

	"chat-service/internal/domain"
	grpcretry "chat-service/internal/infrastructure/grpc"
	maxbotproto "maxbot-service/api/proto/maxbotproto"

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

func (c *MaxClient) GetChatInfo(ctx context.Context, chatID int64) (*domain.ChatInfo, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	var resp *maxbotproto.GetChatInfoResponse
	err := grpcretry.WithRetry(ctx, "MaxBot.GetChatInfo", func() error {
		var callErr error
		resp, callErr = c.client.GetChatInfo(ctx, &maxbotproto.GetChatInfoRequest{ChatId: chatID})
		return callErr
	})
	
	if err != nil {
		return nil, err
	}

	if resp.Error != "" {
		return nil, mapError(resp.ErrorCode, resp.Error)
	}

	if resp.Chat == nil {
		return nil, errors.New("chat info not found")
	}

	return &domain.ChatInfo{
		ChatID:            resp.Chat.ChatId,
		Title:             resp.Chat.Title,
		Type:              resp.Chat.Type,
		ParticipantsCount: int(resp.Chat.ParticipantsCount),
		Description:       resp.Chat.Description,
	}, nil
}

func (c *MaxClient) GetInternalUsers(phones []string) ([]*domain.InternalUser, []string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	var resp *maxbotproto.GetInternalUsersResponse
	err := grpcretry.WithRetry(ctx, "MaxBot.GetInternalUsers", func() error {
		var callErr error
		resp, callErr = c.client.GetInternalUsers(ctx, &maxbotproto.GetInternalUsersRequest{
			PhoneNumbers: phones,
		})
		return callErr
	})

	if err != nil {
		return nil, phones, err
	}

	if resp.Error != "" {
		return nil, phones, mapError(resp.ErrorCode, resp.Error)
	}

	// Конвертируем protobuf объекты в domain объекты
	users := make([]*domain.InternalUser, 0, len(resp.Users))
	for _, protoUser := range resp.Users {
		user := &domain.InternalUser{
			UserID:        protoUser.UserId,
			FirstName:     protoUser.FirstName,
			LastName:      protoUser.LastName,
			IsBot:         protoUser.IsBot,
			Username:      protoUser.Username,
			AvatarURL:     protoUser.AvatarUrl,
			FullAvatarURL: protoUser.FullAvatarUrl,
			Link:          protoUser.Link,
			PhoneNumber:   protoUser.PhoneNumber,
		}
		users = append(users, user)
	}

	return users, resp.FailedPhoneNumbers, nil
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
