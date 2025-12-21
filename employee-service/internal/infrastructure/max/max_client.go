package max

import (
	"context"
	"errors"
	"time"

	"employee-service/internal/domain"
	grpcretry "employee-service/internal/infrastructure/grpc"
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
// Пока MAX API не предоставляет метод получения профиля, используем GetMaxIDByPhone
// и возвращаем пустые имена
func (c *MaxClient) GetUserProfileByPhone(phone string) (*domain.UserProfile, error) {
	// Получаем MAX_id через существующий метод
	maxID, err := c.GetMaxIDByPhone(phone)
	if err != nil {
		return nil, err
	}

	// Возвращаем профиль с MAX_id, но без имен
	// TODO: Когда MAX API предоставит метод GetUserProfile, заменить на реальный вызов
	return &domain.UserProfile{
		MaxID:     maxID,
		FirstName: "", // Пустое имя - MAX API пока не предоставляет эту информацию
		LastName:  "", // Пустая фамилия - MAX API пока не предоставляет эту информацию
		Phone:     phone,
	}, nil
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
