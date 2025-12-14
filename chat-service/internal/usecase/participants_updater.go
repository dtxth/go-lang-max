package usecase

import (
	"chat-service/internal/domain"
	"context"
	"fmt"
	"strconv"
	"time"

	"chat-service/internal/infrastructure/logger"
)

type ParticipantsUpdaterService struct {
	chatRepo    domain.ChatRepository
	cache       domain.ParticipantsCache
	maxService  domain.MaxService
	config      *domain.ParticipantsConfig
	logger      *logger.Logger
}

func NewParticipantsUpdaterService(
	chatRepo domain.ChatRepository,
	cache domain.ParticipantsCache,
	maxService domain.MaxService,
	config *domain.ParticipantsConfig,
	logger *logger.Logger,
) *ParticipantsUpdaterService {
	return &ParticipantsUpdaterService{
		chatRepo:   chatRepo,
		cache:      cache,
		maxService: maxService,
		config:     config,
		logger:     logger,
	}
}

func (s *ParticipantsUpdaterService) UpdateSingle(ctx context.Context, chatID int64, maxChatID string) (*domain.ParticipantsInfo, error) {
	// Проверяем, есть ли MAX Chat ID
	if maxChatID == "" {
		s.logger.Debug(ctx, "No MAX Chat ID for chat", map[string]interface{}{"chat_id": chatID})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Парсим MAX Chat ID в int64
	maxChatIDInt, err := strconv.ParseInt(maxChatID, 10, 64)
	if err != nil {
		s.logger.Error(ctx, "Invalid MAX Chat ID format", map[string]interface{}{
			"chat_id": chatID, 
			"max_chat_id": maxChatID, 
			"error": err.Error(),
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Получаем информацию о чате из MAX API
	chatInfo, err := s.maxService.GetChatInfo(ctx, maxChatIDInt)
	if err != nil {
		s.logger.Error(ctx, "Failed to get chat info from MAX API", map[string]interface{}{
			"chat_id": chatID, 
			"max_chat_id": maxChatID, 
			"error": err.Error(),
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Создаем информацию об участниках
	info := &domain.ParticipantsInfo{
		Count:     chatInfo.ParticipantsCount,
		UpdatedAt: time.Now(),
		Source:    "api",
	}
	
	// Сохраняем в кэш
	if err := s.cache.Set(ctx, chatID, info.Count, s.config.CacheTTL); err != nil {
		s.logger.Error(ctx, "Failed to cache participants count", map[string]interface{}{
			"chat_id": chatID, 
			"error": err.Error(),
		})
	}
	
	// Обновляем в базе данных (опционально)
	if err := s.updateDatabaseCount(ctx, chatID, info.Count); err != nil {
		s.logger.Error(ctx, "Failed to update participants count in database", map[string]interface{}{
			"chat_id": chatID, 
			"error": err.Error(),
		})
	}
	
	s.logger.Debug(ctx, "Successfully updated participants count", map[string]interface{}{
		"chat_id": chatID, 
		"count": info.Count,
	})
	return info, nil
}

func (s *ParticipantsUpdaterService) UpdateBatch(ctx context.Context, chats []domain.ChatUpdateRequest) (map[int64]*domain.ParticipantsInfo, error) {
	result := make(map[int64]*domain.ParticipantsInfo)
	cacheData := make(map[int64]int)
	
	for _, chat := range chats {
		info, err := s.UpdateSingle(ctx, chat.ChatID, chat.MaxChatID)
		if err != nil {
			s.logger.Error(ctx, "Failed to update single chat in batch", map[string]interface{}{
				"chat_id": chat.ChatID, 
				"error": err.Error(),
			})
			continue
		}
		
		result[chat.ChatID] = info
		if info.Source == "api" {
			cacheData[chat.ChatID] = info.Count
		}
	}
	
	// Батчевое сохранение в кэш
	if len(cacheData) > 0 {
		if err := s.cache.SetMultiple(ctx, cacheData, s.config.CacheTTL); err != nil {
			s.logger.Error(ctx, "Failed to batch cache participants counts", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}
	
	s.logger.Info(ctx, "Batch update completed", map[string]interface{}{
		"total": len(chats), 
		"successful": len(result),
	})
	return result, nil
}

func (s *ParticipantsUpdaterService) UpdateStale(ctx context.Context, olderThan time.Duration, batchSize int) (int, error) {
	// Получаем устаревшие чаты из кэша
	staleChats, err := s.cache.GetStaleChats(ctx, olderThan, batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get stale chats: %w", err)
	}
	
	if len(staleChats) == 0 {
		return 0, nil
	}
	
	// Получаем информацию о чатах из базы данных
	updateRequests := make([]domain.ChatUpdateRequest, 0, len(staleChats))
	for _, chatID := range staleChats {
		chat, err := s.chatRepo.GetByID(chatID)
		if err != nil {
			s.logger.Error(ctx, "Failed to get chat from database", map[string]interface{}{
				"chat_id": chatID, 
				"error": err.Error(),
			})
			continue
		}
		
		updateRequests = append(updateRequests, domain.ChatUpdateRequest{
			ChatID:    chatID,
			MaxChatID: chat.MaxChatID,
		})
	}
	
	// Обновляем батчем
	results, err := s.UpdateBatch(ctx, updateRequests)
	if err != nil {
		return 0, fmt.Errorf("failed to update stale chats: %w", err)
	}
	
	return len(results), nil
}

func (s *ParticipantsUpdaterService) UpdateAll(ctx context.Context, batchSize int) (int, error) {
	// Получаем все чаты с MAX Chat ID
	// Это упрощенная реализация - в реальности нужна пагинация
	filter := &domain.ChatFilter{} // без фильтрации для полного обновления
	chats, _, err := s.chatRepo.GetAllWithSortingAndSearch(10000, 0, "id", "asc", "", filter)
	if err != nil {
		return 0, fmt.Errorf("failed to get all chats: %w", err)
	}
	
	totalUpdated := 0
	updateRequests := make([]domain.ChatUpdateRequest, 0, batchSize)
	
	for _, chat := range chats {
		if chat.MaxChatID == "" {
			continue // пропускаем чаты без MAX Chat ID
		}
		
		updateRequests = append(updateRequests, domain.ChatUpdateRequest{
			ChatID:    chat.ID,
			MaxChatID: chat.MaxChatID,
		})
		
		// Обрабатываем батчами
		if len(updateRequests) >= batchSize {
			results, err := s.UpdateBatch(ctx, updateRequests)
			if err != nil {
				s.logger.Error(ctx, "Failed to update batch", map[string]interface{}{
					"error": err.Error(),
				})
			} else {
				totalUpdated += len(results)
			}
			
			updateRequests = updateRequests[:0] // очищаем слайс
			
			// Небольшая пауза между батчами для снижения нагрузки на MAX API
			time.Sleep(1 * time.Second)
		}
	}
	
	// Обрабатываем оставшиеся чаты
	if len(updateRequests) > 0 {
		results, err := s.UpdateBatch(ctx, updateRequests)
		if err != nil {
			s.logger.Error(ctx, "Failed to update final batch", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			totalUpdated += len(results)
		}
	}
	
	s.logger.Info(ctx, "Full update completed", map[string]interface{}{
		"total_updated": totalUpdated,
	})
	return totalUpdated, nil
}

// getFallbackInfo возвращает информацию из базы данных как fallback
func (s *ParticipantsUpdaterService) getFallbackInfo(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return nil, fmt.Errorf("failed to get chat from database: %w", err)
	}
	
	return &domain.ParticipantsInfo{
		Count:     chat.ParticipantsCount,
		UpdatedAt: chat.UpdatedAt,
		Source:    "database",
	}, nil
}

// updateDatabaseCount обновляет количество участников в базе данных
func (s *ParticipantsUpdaterService) updateDatabaseCount(ctx context.Context, chatID int64, count int) error {
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		return err
	}
	
	chat.ParticipantsCount = count
	return s.chatRepo.Update(chat)
}