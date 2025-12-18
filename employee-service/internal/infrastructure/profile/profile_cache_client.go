package profile

import (
	"context"
	"employee-service/internal/domain"

	"google.golang.org/grpc"
	maxbotproto "maxbot-service/api/proto"
)

// ProfileCacheClient реализует ProfileCacheService через gRPC вызовы к maxbot-service
// Временно использует NoOp реализацию, так как profile cache не экспонирован через gRPC
type ProfileCacheClient struct {
	client maxbotproto.MaxBotServiceClient
	noop   *NoOpProfileCacheClient
}

// NewProfileCacheClient создает новый клиент для работы с кэшем профилей
func NewProfileCacheClient(conn *grpc.ClientConn) domain.ProfileCacheService {
	return &ProfileCacheClient{
		client: maxbotproto.NewMaxBotServiceClient(conn),
		noop:   &NoOpProfileCacheClient{},
	}
}

// GetProfile получает профиль пользователя из кэша по MAX_id
// Временно использует NoOp реализацию для graceful degradation (Requirements 3.4, 7.5)
func (c *ProfileCacheClient) GetProfile(ctx context.Context, userID string) (*domain.CachedUserProfile, error) {
	// Используем NoOp реализацию для обеспечения обратной совместимости
	// Profile cache функциональность реализована внутри maxbot-service
	// и не экспонирована через gRPC API
	return c.noop.GetProfile(ctx, userID)
}