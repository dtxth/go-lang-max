package max

import (
	"context"
	"errors"
	"fmt"
	"time"

	"employee-service/internal/domain"
	grpcretry "employee-service/internal/infrastructure/grpc"
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

func (c *MaxClient) GetConnection() *grpc.ClientConn {
	return c.conn
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
	// Этот метод больше не используется для валидации телефонов
	// Валидация теперь выполняется локально через PhoneValidator
	// Оставляем реализацию для совместимости с интерфейсом
	return true
}

// BatchGetMaxIDByPhone получает MAX_id для нескольких телефонов
// Обрабатывает до 100 телефонов за раз
func (c *MaxClient) BatchGetMaxIDByPhone(phones []string) (map[string]string, error) {
	result := make(map[string]string)
	
	// Ограничиваем размер батча до 100 (Requirements 4.2, 4.3)
	batchSize := 100
	if len(phones) > batchSize {
		phones = phones[:batchSize]
	}
	
	// Пока нет batch метода в MaxBot Service, вызываем по одному
	// TODO: Когда будет реализован BatchGetUsersByPhone в MaxBot Service, использовать его
	for _, phone := range phones {
		maxID, err := c.GetMaxIDByPhone(phone)
		if err == nil && maxID != "" {
			result[phone] = maxID
		}
		// Игнорируем ошибки для отдельных телефонов
	}
	
	return result, nil
}

// GetUserProfileByPhone получает профиль пользователя по номеру телефона
// Теперь использует GetInternalUsers для получения полной информации
func (c *MaxClient) GetUserProfileByPhone(phone string) (*domain.UserProfile, error) {
	// Используем новый метод GetInternalUsers для получения полной информации
	users, failed, err := c.GetInternalUsers([]string{phone})
	if err != nil {
		return nil, err
	}

	// Если телефон не найден
	if len(users) == 0 {
		if len(failed) > 0 {
			return nil, domain.ErrMaxIDNotFound
		}
		return nil, domain.ErrInvalidPhone
	}

	// Конвертируем InternalUser в UserProfile
	user := users[0]
	return &domain.UserProfile{
		MaxID:     fmt.Sprintf("%d", user.UserID), // Используем UserID как MaxID
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Phone:     user.PhoneNumber,
	}, nil
}

// GetInternalUsers получает детальную информацию о пользователях по номерам телефонов
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
