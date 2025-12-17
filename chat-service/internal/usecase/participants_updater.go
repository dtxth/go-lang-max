package usecase

import (
	"chat-service/internal/domain"
	"context"
	"fmt"
	"strconv"
	"time"

	"chat-service/internal/infrastructure/logger"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

type ParticipantsUpdaterService struct {
	chatRepo       domain.ChatRepository
	cache          domain.ParticipantsCache
	maxService     domain.MaxService
	config         *domain.ParticipantsConfig
	logger         *logger.Logger
	circuitBreaker CircuitBreaker
}

// CircuitBreaker interface for dependency injection
type CircuitBreaker interface {
	CanExecute() bool
	RecordSuccess()
	RecordFailure()
	GetState() CircuitState
}

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

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

func NewParticipantsUpdaterServiceWithCircuitBreaker(
	chatRepo domain.ChatRepository,
	cache domain.ParticipantsCache,
	maxService domain.MaxService,
	config *domain.ParticipantsConfig,
	logger *logger.Logger,
	circuitBreaker CircuitBreaker,
) *ParticipantsUpdaterService {
	return &ParticipantsUpdaterService{
		chatRepo:       chatRepo,
		cache:          cache,
		maxService:     maxService,
		config:         config,
		logger:         logger,
		circuitBreaker: circuitBreaker,
	}
}

func (s *ParticipantsUpdaterService) UpdateSingle(ctx context.Context, chatID int64, maxChatID string) (*domain.ParticipantsInfo, error) {
	updateStart := time.Now()
	
	s.logger.Debug(ctx, "Starting single chat participants update", map[string]interface{}{
		"component":   "participants_updater",
		"operation":   "update_single_start",
		"chat_id":     chatID,
		"max_chat_id": maxChatID,
	})
	
	// Проверяем, есть ли MAX Chat ID
	if maxChatID == "" {
		s.logger.Debug(ctx, "No MAX Chat ID for chat, using fallback", map[string]interface{}{
			"component":   "participants_updater",
			"operation":   "update_single_no_max_id",
			"chat_id":     chatID,
			"fallback":    "database",
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Парсим MAX Chat ID в int64
	maxChatIDInt, err := strconv.ParseInt(maxChatID, 10, 64)
	if err != nil {
		s.logger.Error(ctx, "Invalid MAX Chat ID format", map[string]interface{}{
			"component":   "participants_updater",
			"operation":   "update_single_parse_error",
			"chat_id":     chatID, 
			"max_chat_id": maxChatID, 
			"error":       err.Error(),
			"fallback":    "database",
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Проверяем circuit breaker перед вызовом MAX API
	if s.circuitBreaker != nil && !s.circuitBreaker.CanExecute() {
		s.logger.Warn(ctx, "Circuit breaker is open, using fallback data", map[string]interface{}{
			"component":             "participants_updater",
			"operation":             "update_single_circuit_breaker_open",
			"chat_id":               chatID,
			"max_chat_id":           maxChatID,
			"circuit_breaker_state": s.circuitBreaker.GetState(),
			"fallback":              "database",
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Получаем информацию о чате из MAX API с retry logic
	apiCallStart := time.Now()
	chatInfo, err := s.getChatInfoWithRetry(ctx, maxChatIDInt, chatID, maxChatID)
	apiCallDuration := time.Since(apiCallStart)
	
	if err != nil {
		if s.circuitBreaker != nil {
			s.circuitBreaker.RecordFailure()
		}
		s.logger.Error(ctx, "Failed to get chat info from MAX API after retries", map[string]interface{}{
			"component":        "participants_updater",
			"operation":        "update_single_api_failed",
			"chat_id":          chatID, 
			"max_chat_id":      maxChatID, 
			"error":            err.Error(),
			"api_call_duration": apiCallDuration.String(),
			"fallback":         "database",
		})
		return s.getFallbackInfo(ctx, chatID)
	}
	
	// Записываем успех в circuit breaker
	if s.circuitBreaker != nil {
		s.circuitBreaker.RecordSuccess()
	}
	
	s.logger.Debug(ctx, "Successfully retrieved chat info from MAX API", map[string]interface{}{
		"component":         "participants_updater",
		"operation":         "update_single_api_success",
		"chat_id":           chatID,
		"max_chat_id":       maxChatID,
		"participants_count": chatInfo.ParticipantsCount,
		"api_call_duration": apiCallDuration.String(),
	})
	
	// Создаем информацию об участниках
	info := &domain.ParticipantsInfo{
		Count:     chatInfo.ParticipantsCount,
		UpdatedAt: time.Now(),
		Source:    "api",
	}
	
	// Сохраняем в кэш (если доступен)
	cacheStart := time.Now()
	if s.cache != nil {
		if err := s.cache.Set(ctx, chatID, info.Count, s.config.CacheTTL); err != nil {
			s.logger.Error(ctx, "Failed to cache participants count", map[string]interface{}{
				"component":       "participants_updater",
				"operation":       "update_single_cache_failed",
				"chat_id":         chatID, 
				"count":           info.Count,
				"cache_ttl":       s.config.CacheTTL.String(),
				"error":           err.Error(),
			})
		} else {
			cacheDuration := time.Since(cacheStart)
			s.logger.Debug(ctx, "Successfully cached participants count", map[string]interface{}{
				"component":      "participants_updater",
				"operation":      "update_single_cache_success",
				"chat_id":        chatID,
				"count":          info.Count,
				"cache_ttl":      s.config.CacheTTL.String(),
				"cache_duration": cacheDuration.String(),
			})
		}
	}
	
	// Обновляем в базе данных (опционально)
	dbStart := time.Now()
	if err := s.updateDatabaseCount(ctx, chatID, info.Count); err != nil {
		s.logger.Error(ctx, "Failed to update participants count in database", map[string]interface{}{
			"component": "participants_updater",
			"operation": "update_single_db_failed",
			"chat_id":   chatID, 
			"count":     info.Count,
			"error":     err.Error(),
		})
	} else {
		dbDuration := time.Since(dbStart)
		s.logger.Debug(ctx, "Successfully updated participants count in database", map[string]interface{}{
			"component":   "participants_updater",
			"operation":   "update_single_db_success",
			"chat_id":     chatID,
			"count":       info.Count,
			"db_duration": dbDuration.String(),
		})
	}
	
	totalDuration := time.Since(updateStart)
	s.logger.Info(ctx, "Successfully completed single chat participants update", map[string]interface{}{
		"component":      "participants_updater",
		"operation":      "update_single_completed",
		"chat_id":        chatID, 
		"count":          info.Count,
		"source":         info.Source,
		"total_duration": totalDuration.String(),
	})
	
	// Performance warning for slow updates
	if totalDuration > 10*time.Second {
		s.logger.Warn(ctx, "Single update was slow", map[string]interface{}{
			"component":      "participants_updater",
			"operation":      "update_single_slow",
			"chat_id":        chatID,
			"duration":       totalDuration.String(),
			"expected_max":   "10s",
		})
	}
	
	return info, nil
}

func (s *ParticipantsUpdaterService) UpdateBatch(ctx context.Context, chats []domain.ChatUpdateRequest) (map[int64]*domain.ParticipantsInfo, error) {
	batchStart := time.Now()
	result := make(map[int64]*domain.ParticipantsInfo)
	cacheData := make(map[int64]int)
	errors := make([]error, 0)
	
	s.logger.Info(ctx, "Starting batch participants update", map[string]interface{}{
		"component":   "participants_updater",
		"operation":   "update_batch_start",
		"batch_size":  len(chats),
		"timeout":     s.config.MaxAPITimeout.String(),
	})
	
	for i, chat := range chats {
		// Проверяем контекст на каждой итерации
		select {
		case <-ctx.Done():
			s.logger.Warn(ctx, "Batch update cancelled", map[string]interface{}{
				"component":     "participants_updater",
				"operation":     "update_batch_cancelled",
				"processed":     i,
				"total":         len(chats),
				"successful":    len(result),
				"failed":        len(errors),
				"duration":      time.Since(batchStart).String(),
				"cancel_reason": ctx.Err().Error(),
			})
			return result, ctx.Err()
		default:
		}
		
		itemStart := time.Now()
		info, err := s.UpdateSingle(ctx, chat.ChatID, chat.MaxChatID)
		itemDuration := time.Since(itemStart)
		
		if err != nil {
			errors = append(errors, fmt.Errorf("chat_id %d: %w", chat.ChatID, err))
			s.logger.Error(ctx, "Failed to update single chat in batch", map[string]interface{}{
				"component":      "participants_updater",
				"operation":      "update_batch_item_failed",
				"chat_id":        chat.ChatID, 
				"max_chat_id":    chat.MaxChatID,
				"error":          err.Error(),
				"batch_progress": fmt.Sprintf("%d/%d", i+1, len(chats)),
				"item_duration":  itemDuration.String(),
			})
			continue
		}
		
		result[chat.ChatID] = info
		if info.Source == "api" {
			cacheData[chat.ChatID] = info.Count
		}
		
		// Логируем прогресс для больших батчей
		if len(chats) > 10 && (i+1)%10 == 0 {
			progressDuration := time.Since(batchStart)
			itemsPerSecond := float64(i+1) / progressDuration.Seconds()
			
			s.logger.Info(ctx, "Batch update progress", map[string]interface{}{
				"component":       "participants_updater",
				"operation":       "update_batch_progress",
				"processed":       i + 1,
				"total":           len(chats),
				"successful":      len(result),
				"failed":          len(errors),
				"duration":        progressDuration.String(),
				"items_per_second": fmt.Sprintf("%.2f", itemsPerSecond),
				"progress_percent": fmt.Sprintf("%.1f%%", float64(i+1)/float64(len(chats))*100),
			})
		}
	}
	
	// Батчевое сохранение в кэш (если доступен)
	batchCacheStart := time.Now()
	if s.cache != nil && len(cacheData) > 0 {
		if err := s.cache.SetMultiple(ctx, cacheData, s.config.CacheTTL); err != nil {
			s.logger.Error(ctx, "Failed to batch cache participants counts", map[string]interface{}{
				"component":   "participants_updater",
				"operation":   "update_batch_cache_failed",
				"error":       err.Error(),
				"cache_items": len(cacheData),
				"cache_ttl":   s.config.CacheTTL.String(),
			})
		} else {
			batchCacheDuration := time.Since(batchCacheStart)
			s.logger.Debug(ctx, "Successfully cached batch results", map[string]interface{}{
				"component":          "participants_updater",
				"operation":          "update_batch_cache_success",
				"cache_items":        len(cacheData),
				"cache_ttl":          s.config.CacheTTL.String(),
				"batch_cache_duration": batchCacheDuration.String(),
			})
		}
	}
	
	totalDuration := time.Since(batchStart)
	successRate := float64(len(result)) / float64(len(chats)) * 100
	itemsPerSecond := float64(len(chats)) / totalDuration.Seconds()
	
	logData := map[string]interface{}{
		"component":       "participants_updater",
		"operation":       "update_batch_completed",
		"total":           len(chats), 
		"successful":      len(result),
		"failed":          len(errors),
		"success_rate":    fmt.Sprintf("%.1f%%", successRate),
		"duration":        totalDuration.String(),
		"items_per_second": fmt.Sprintf("%.2f", itemsPerSecond),
		"cached_items":    len(cacheData),
	}
	
	// Добавляем детали ошибок для анализа
	if len(errors) > 0 {
		errorTypes := make(map[string]int)
		for _, err := range errors {
			errorTypes[fmt.Sprintf("%T", err)]++
		}
		logData["error_types"] = errorTypes
		logData["sample_errors"] = errors[:min(3, len(errors))] // Первые 3 ошибки для анализа
		
		s.logger.Warn(ctx, "Batch update completed with errors", logData)
	} else {
		s.logger.Info(ctx, "Batch update completed successfully", logData)
	}
	
	// Performance warnings
	if totalDuration > 5*time.Minute {
		s.logger.Warn(ctx, "Batch update was slow", map[string]interface{}{
			"component":     "participants_updater",
			"operation":     "update_batch_slow",
			"duration":      totalDuration.String(),
			"expected_max":  "5m",
			"batch_size":    len(chats),
		})
	}
	
	if successRate < 90.0 && len(chats) > 5 {
		s.logger.Warn(ctx, "Batch update has low success rate", map[string]interface{}{
			"component":    "participants_updater",
			"operation":    "update_batch_low_success_rate",
			"success_rate": fmt.Sprintf("%.1f%%", successRate),
			"threshold":    "90%",
			"batch_size":   len(chats),
		})
	}
	
	// Возвращаем результат даже если были ошибки (частичный успех)
	return result, nil
}

func (s *ParticipantsUpdaterService) UpdateStale(ctx context.Context, olderThan time.Duration, batchSize int) (int, error) {
	staleUpdateStart := time.Now()
	
	s.logger.Info(ctx, "Starting stale participants update", map[string]interface{}{
		"component":      "participants_updater",
		"operation":      "update_stale_start",
		"older_than":     olderThan.String(),
		"batch_size":     batchSize,
	})
	
	// Получаем устаревшие чаты из кэша
	staleQueryStart := time.Now()
	staleChats, err := s.cache.GetStaleChats(ctx, olderThan, batchSize)
	staleQueryDuration := time.Since(staleQueryStart)
	
	if err != nil {
		s.logger.Error(ctx, "Failed to get stale chats from cache", map[string]interface{}{
			"component":           "participants_updater",
			"operation":           "update_stale_query_failed",
			"older_than":          olderThan.String(),
			"batch_size":          batchSize,
			"error":               err.Error(),
			"stale_query_duration": staleQueryDuration.String(),
		})
		return 0, fmt.Errorf("failed to get stale chats: %w", err)
	}
	
	s.logger.Info(ctx, "Retrieved stale chats from cache", map[string]interface{}{
		"component":           "participants_updater",
		"operation":           "update_stale_query_success",
		"stale_count":         len(staleChats),
		"older_than":          olderThan.String(),
		"stale_query_duration": staleQueryDuration.String(),
	})
	
	if len(staleChats) == 0 {
		s.logger.Debug(ctx, "No stale chats found", map[string]interface{}{
			"component":  "participants_updater",
			"operation":  "update_stale_no_data",
			"older_than": olderThan.String(),
		})
		return 0, nil
	}
	
	// Получаем информацию о чатах из базы данных
	dbQueryStart := time.Now()
	updateRequests := make([]domain.ChatUpdateRequest, 0, len(staleChats))
	dbErrors := 0
	
	for _, chatID := range staleChats {
		chat, err := s.chatRepo.GetByID(chatID)
		if err != nil {
			dbErrors++
			s.logger.Error(ctx, "Failed to get chat from database during stale update", map[string]interface{}{
				"component": "participants_updater",
				"operation": "update_stale_db_query_failed",
				"chat_id":   chatID, 
				"error":     err.Error(),
			})
			continue
		}
		
		updateRequests = append(updateRequests, domain.ChatUpdateRequest{
			ChatID:    chatID,
			MaxChatID: chat.MaxChatID,
		})
	}
	
	dbQueryDuration := time.Since(dbQueryStart)
	s.logger.Info(ctx, "Retrieved chat data from database for stale update", map[string]interface{}{
		"component":        "participants_updater",
		"operation":        "update_stale_db_query_completed",
		"requested_chats":  len(staleChats),
		"valid_chats":      len(updateRequests),
		"db_errors":        dbErrors,
		"db_query_duration": dbQueryDuration.String(),
	})
	
	// Обновляем батчем
	if len(updateRequests) == 0 {
		s.logger.Warn(ctx, "No valid chats to update in stale update", map[string]interface{}{
			"component":       "participants_updater",
			"operation":       "update_stale_no_valid_chats",
			"stale_found":     len(staleChats),
			"db_errors":       dbErrors,
		})
		return 0, nil
	}
	
	results, err := s.UpdateBatch(ctx, updateRequests)
	if err != nil {
		s.logger.Error(ctx, "Failed to update stale chats batch", map[string]interface{}{
			"component":      "participants_updater",
			"operation":      "update_stale_batch_failed",
			"update_requests": len(updateRequests),
			"error":          err.Error(),
		})
		return 0, fmt.Errorf("failed to update stale chats: %w", err)
	}
	
	staleTotalDuration := time.Since(staleUpdateStart)
	s.logger.Info(ctx, "Completed stale participants update", map[string]interface{}{
		"component":       "participants_updater",
		"operation":       "update_stale_completed",
		"stale_found":     len(staleChats),
		"update_requests": len(updateRequests),
		"successful_updates": len(results),
		"db_errors":       dbErrors,
		"total_duration":  staleTotalDuration.String(),
		"older_than":      olderThan.String(),
	})
	
	return len(results), nil
}

func (s *ParticipantsUpdaterService) UpdateAll(ctx context.Context, batchSize int) (int, error) {
	fullUpdateStart := time.Now()
	
	s.logger.Info(ctx, "Starting full participants update", map[string]interface{}{
		"component":  "participants_updater",
		"operation":  "update_all_start",
		"batch_size": batchSize,
	})
	
	// Получаем все чаты с MAX Chat ID
	// Это упрощенная реализация - в реальности нужна пагинация
	dbQueryStart := time.Now()
	filter := &domain.ChatFilter{} // без фильтрации для полного обновления
	chats, _, err := s.chatRepo.GetAllWithSortingAndSearch(10000, 0, "id", "asc", "", filter)
	dbQueryDuration := time.Since(dbQueryStart)
	
	if err != nil {
		s.logger.Error(ctx, "Failed to get all chats for full update", map[string]interface{}{
			"component":        "participants_updater",
			"operation":        "update_all_db_query_failed",
			"error":            err.Error(),
			"db_query_duration": dbQueryDuration.String(),
		})
		return 0, fmt.Errorf("failed to get all chats: %w", err)
	}
	
	s.logger.Info(ctx, "Retrieved all chats from database", map[string]interface{}{
		"component":        "participants_updater",
		"operation":        "update_all_db_query_success",
		"total_chats":      len(chats),
		"db_query_duration": dbQueryDuration.String(),
	})
	
	totalUpdated := 0
	totalBatches := 0
	skippedChats := 0
	updateRequests := make([]domain.ChatUpdateRequest, 0, batchSize)
	
	for i, chat := range chats {
		if chat.MaxChatID == "" {
			skippedChats++
			continue // пропускаем чаты без MAX Chat ID
		}
		
		updateRequests = append(updateRequests, domain.ChatUpdateRequest{
			ChatID:    chat.ID,
			MaxChatID: chat.MaxChatID,
		})
		
		// Обрабатываем батчами
		if len(updateRequests) >= batchSize {
			batchStart := time.Now()
			totalBatches++
			
			results, err := s.UpdateBatch(ctx, updateRequests)
			batchDuration := time.Since(batchStart)
			
			if err != nil {
				s.logger.Error(ctx, "Failed to update batch in full update", map[string]interface{}{
					"component":     "participants_updater",
					"operation":     "update_all_batch_failed",
					"batch_number":  totalBatches,
					"batch_size":    len(updateRequests),
					"error":         err.Error(),
					"batch_duration": batchDuration.String(),
				})
			} else {
				totalUpdated += len(results)
				s.logger.Info(ctx, "Completed batch in full update", map[string]interface{}{
					"component":      "participants_updater",
					"operation":      "update_all_batch_success",
					"batch_number":   totalBatches,
					"batch_size":     len(updateRequests),
					"batch_updated":  len(results),
					"total_updated":  totalUpdated,
					"batch_duration": batchDuration.String(),
					"progress":       fmt.Sprintf("%.1f%%", float64(i+1)/float64(len(chats))*100),
				})
			}
			
			updateRequests = updateRequests[:0] // очищаем слайс
			
			// Небольшая пауза между батчами для снижения нагрузки на MAX API
			s.logger.Debug(ctx, "Pausing between batches", map[string]interface{}{
				"component":    "participants_updater",
				"operation":    "update_all_batch_pause",
				"batch_number": totalBatches,
				"pause_duration": "1s",
			})
			time.Sleep(1 * time.Second)
		}
	}
	
	// Обрабатываем оставшиеся чаты
	if len(updateRequests) > 0 {
		finalBatchStart := time.Now()
		totalBatches++
		
		results, err := s.UpdateBatch(ctx, updateRequests)
		finalBatchDuration := time.Since(finalBatchStart)
		
		if err != nil {
			s.logger.Error(ctx, "Failed to update final batch in full update", map[string]interface{}{
				"component":      "participants_updater",
				"operation":      "update_all_final_batch_failed",
				"batch_number":   totalBatches,
				"batch_size":     len(updateRequests),
				"error":          err.Error(),
				"batch_duration": finalBatchDuration.String(),
			})
		} else {
			totalUpdated += len(results)
			s.logger.Info(ctx, "Completed final batch in full update", map[string]interface{}{
				"component":      "participants_updater",
				"operation":      "update_all_final_batch_success",
				"batch_number":   totalBatches,
				"batch_size":     len(updateRequests),
				"batch_updated":  len(results),
				"total_updated":  totalUpdated,
				"batch_duration": finalBatchDuration.String(),
			})
		}
	}
	
	fullUpdateDuration := time.Since(fullUpdateStart)
	updateRate := float64(totalUpdated) / fullUpdateDuration.Seconds()
	
	s.logger.Info(ctx, "Full participants update completed", map[string]interface{}{
		"component":       "participants_updater",
		"operation":       "update_all_completed",
		"total_chats":     len(chats),
		"skipped_chats":   skippedChats,
		"total_batches":   totalBatches,
		"total_updated":   totalUpdated,
		"update_rate":     fmt.Sprintf("%.2f items/sec", updateRate),
		"total_duration":  fullUpdateDuration.String(),
		"success_rate":    fmt.Sprintf("%.1f%%", float64(totalUpdated)/float64(len(chats)-skippedChats)*100),
	})
	
	// Performance analysis
	if fullUpdateDuration > 2*time.Hour {
		s.logger.Warn(ctx, "Full update took longer than expected", map[string]interface{}{
			"component":     "participants_updater",
			"operation":     "update_all_slow",
			"duration":      fullUpdateDuration.String(),
			"expected_max":  "2h",
			"total_chats":   len(chats),
		})
	}
	
	return totalUpdated, nil
}

// getFallbackInfo возвращает информацию из базы данных как fallback
func (s *ParticipantsUpdaterService) getFallbackInfo(ctx context.Context, chatID int64) (*domain.ParticipantsInfo, error) {
	fallbackStart := time.Now()
	
	s.logger.Debug(ctx, "Using database fallback for participants info", map[string]interface{}{
		"component": "participants_updater",
		"operation": "get_fallback_info",
		"chat_id":   chatID,
	})
	
	chat, err := s.chatRepo.GetByID(chatID)
	fallbackDuration := time.Since(fallbackStart)
	
	if err != nil {
		s.logger.Error(ctx, "Failed to get fallback info from database", map[string]interface{}{
			"component":        "participants_updater",
			"operation":        "get_fallback_info_failed",
			"chat_id":          chatID,
			"error":            err.Error(),
			"fallback_duration": fallbackDuration.String(),
		})
		return nil, fmt.Errorf("failed to get chat from database: %w", err)
	}
	
	info := &domain.ParticipantsInfo{
		Count:     chat.ParticipantsCount,
		UpdatedAt: chat.UpdatedAt,
		Source:    "database",
	}
	
	s.logger.Debug(ctx, "Successfully retrieved fallback info", map[string]interface{}{
		"component":        "participants_updater",
		"operation":        "get_fallback_info_success",
		"chat_id":          chatID,
		"count":            info.Count,
		"data_age":         time.Since(info.UpdatedAt).String(),
		"fallback_duration": fallbackDuration.String(),
	})
	
	return info, nil
}

// updateDatabaseCount обновляет количество участников в базе данных
func (s *ParticipantsUpdaterService) updateDatabaseCount(ctx context.Context, chatID int64, count int) error {
	dbUpdateStart := time.Now()
	
	chat, err := s.chatRepo.GetByID(chatID)
	if err != nil {
		s.logger.Error(ctx, "Failed to get chat for database update", map[string]interface{}{
			"component": "participants_updater",
			"operation": "update_database_count_get_failed",
			"chat_id":   chatID,
			"count":     count,
			"error":     err.Error(),
		})
		return err
	}
	
	oldCount := chat.ParticipantsCount
	chat.ParticipantsCount = count
	
	err = s.chatRepo.Update(chat)
	dbUpdateDuration := time.Since(dbUpdateStart)
	
	if err != nil {
		s.logger.Error(ctx, "Failed to update chat in database", map[string]interface{}{
			"component":        "participants_updater",
			"operation":        "update_database_count_update_failed",
			"chat_id":          chatID,
			"old_count":        oldCount,
			"new_count":        count,
			"error":            err.Error(),
			"db_update_duration": dbUpdateDuration.String(),
		})
		return err
	}
	
	s.logger.Debug(ctx, "Successfully updated chat in database", map[string]interface{}{
		"component":        "participants_updater",
		"operation":        "update_database_count_success",
		"chat_id":          chatID,
		"old_count":        oldCount,
		"new_count":        count,
		"count_change":     count - oldCount,
		"db_update_duration": dbUpdateDuration.String(),
	})
	
	return nil
}

// getChatInfoWithRetry получает информацию о чате с retry logic
func (s *ParticipantsUpdaterService) getChatInfoWithRetry(ctx context.Context, maxChatIDInt int64, chatID int64, maxChatID string) (*domain.ChatInfo, error) {
	retryStart := time.Now()
	maxRetries := s.config.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 1 // At least one attempt
	}
	retryDelay := 1 * time.Second
	
	s.logger.Debug(ctx, "Starting MAX API call with retry logic", map[string]interface{}{
		"component":    "participants_updater",
		"operation":    "get_chat_info_retry_start",
		"chat_id":      chatID,
		"max_chat_id":  maxChatID,
		"max_retries":  maxRetries,
		"api_timeout":  s.config.MaxAPITimeout.String(),
	})
	
	for attempt := 1; attempt <= maxRetries; attempt++ {
		attemptStart := time.Now()
		
		// Создаем контекст с таймаутом для каждой попытки
		attemptCtx, cancel := context.WithTimeout(ctx, s.config.MaxAPITimeout)
		
		chatInfo, err := s.maxService.GetChatInfo(attemptCtx, maxChatIDInt)
		cancel()
		
		attemptDuration := time.Since(attemptStart)
		
		if err == nil {
			totalRetryDuration := time.Since(retryStart)
			
			logData := map[string]interface{}{
				"component":           "participants_updater",
				"operation":           "get_chat_info_retry_success",
				"chat_id":             chatID,
				"max_chat_id":         maxChatID,
				"attempt":             attempt,
				"participants_count":  chatInfo.ParticipantsCount,
				"attempt_duration":    attemptDuration.String(),
				"total_retry_duration": totalRetryDuration.String(),
			}
			
			if attempt > 1 {
				s.logger.Info(ctx, "MAX API call succeeded after retry", logData)
			} else {
				s.logger.Debug(ctx, "MAX API call succeeded on first attempt", logData)
			}
			
			return chatInfo, nil
		}
		
		s.logger.Warn(ctx, "MAX API call attempt failed", map[string]interface{}{
			"component":        "participants_updater",
			"operation":        "get_chat_info_retry_attempt_failed",
			"chat_id":          chatID,
			"max_chat_id":      maxChatID,
			"attempt":          attempt,
			"max_retries":      maxRetries,
			"error":            err.Error(),
			"attempt_duration": attemptDuration.String(),
			"api_timeout":      s.config.MaxAPITimeout.String(),
		})
		
		// Не делаем retry на последней попытке
		if attempt < maxRetries {
			s.logger.Debug(ctx, "Waiting before retry", map[string]interface{}{
				"component":   "participants_updater",
				"operation":   "get_chat_info_retry_wait",
				"chat_id":     chatID,
				"attempt":     attempt,
				"retry_delay": retryDelay.String(),
			})
			
			select {
			case <-ctx.Done():
				s.logger.Warn(ctx, "MAX API retry cancelled due to context", map[string]interface{}{
					"component":   "participants_updater",
					"operation":   "get_chat_info_retry_cancelled",
					"chat_id":     chatID,
					"attempt":     attempt,
					"cancel_reason": ctx.Err().Error(),
				})
				return nil, ctx.Err()
			case <-time.After(retryDelay):
				retryDelay *= 2 // Exponential backoff
			}
		}
	}
	
	totalRetryDuration := time.Since(retryStart)
	s.logger.Error(ctx, "All MAX API retry attempts failed", map[string]interface{}{
		"component":           "participants_updater",
		"operation":           "get_chat_info_retry_exhausted",
		"chat_id":             chatID,
		"max_chat_id":         maxChatID,
		"total_attempts":      maxRetries,
		"total_retry_duration": totalRetryDuration.String(),
	})
	
	return nil, fmt.Errorf("MAX API call failed after %d attempts", maxRetries)
}