package worker

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/logger"
	"context"
	"fmt"
	"sync"
	"time"
)

// ParticipantsWorker выполняет фоновое обновление количества участников
type ParticipantsWorker struct {
	updater domain.ParticipantsUpdater
	config  *domain.ParticipantsConfig
	logger  *logger.Logger
	
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

func NewParticipantsWorker(
	updater domain.ParticipantsUpdater,
	config *domain.ParticipantsConfig,
	logger *logger.Logger,
) *ParticipantsWorker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ParticipantsWorker{
		updater: updater,
		config:  config,
		logger:  logger,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start запускает фоновые задачи
func (w *ParticipantsWorker) Start() {
	if !w.config.EnableBackgroundSync {
		w.logger.Info(context.Background(), "Background sync disabled, skipping participants worker", nil)
		return
	}
	
	w.logger.Info(context.Background(), "Starting participants worker", map[string]interface{}{
		"update_interval": w.config.UpdateInterval.String(),
		"full_update_hour": w.config.FullUpdateHour,
		"batch_size": w.config.BatchSize,
	})
	
	// Запускаем периодическое обновление устаревших данных
	w.wg.Add(1)
	go w.runStaleUpdater()
	
	// Запускаем полное обновление раз в сутки
	w.wg.Add(1)
	go w.runFullUpdater()
}

// Stop останавливает фоновые задачи
func (w *ParticipantsWorker) Stop() {
	w.logger.Info(context.Background(), "Stopping participants worker", nil)
	w.cancel()
	w.wg.Wait()
	w.logger.Info(context.Background(), "Participants worker stopped", nil)
}

// runStaleUpdater периодически обновляет устаревшие данные
func (w *ParticipantsWorker) runStaleUpdater() {
	defer w.wg.Done()
	
	ticker := time.NewTicker(w.config.UpdateInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.updateStaleData()
		}
	}
}

// runFullUpdater выполняет полное обновление раз в сутки
func (w *ParticipantsWorker) runFullUpdater() {
	defer w.wg.Done()
	
	// Вычисляем время до следующего полного обновления
	now := time.Now()
	nextUpdate := time.Date(now.Year(), now.Month(), now.Day(), w.config.FullUpdateHour, 0, 0, 0, now.Location())
	if nextUpdate.Before(now) {
		nextUpdate = nextUpdate.Add(24 * time.Hour)
	}
	
	// Ждем до времени первого обновления
	timer := time.NewTimer(time.Until(nextUpdate))
	defer timer.Stop()
	
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-timer.C:
			w.performFullUpdate()
			// Устанавливаем таймер на следующие сутки
			timer.Reset(24 * time.Hour)
		}
	}
}

// updateStaleData обновляет устаревшие данные
func (w *ParticipantsWorker) updateStaleData() {
	ctx, cancel := context.WithTimeout(w.ctx, 5*time.Minute)
	defer cancel()
	
	w.logger.Debug(ctx, "Starting stale data update", map[string]interface{}{
		"stale_threshold": w.config.StaleThreshold.String(),
		"batch_size": w.config.BatchSize,
	})
	
	start := time.Now()
	updated, err := w.updater.UpdateStale(ctx, w.config.StaleThreshold, w.config.BatchSize)
	duration := time.Since(start)
	
	logData := map[string]interface{}{
		"duration": duration.String(),
		"stale_threshold": w.config.StaleThreshold.String(),
		"batch_size": w.config.BatchSize,
	}
	
	if err != nil {
		logData["error"] = err.Error()
		w.logger.Error(ctx, "Failed to update stale participants data", logData)
		
		// Проверяем производительность даже при ошибке
		if duration > 2*time.Minute {
			w.logger.Warn(ctx, "Stale data update took longer than expected", map[string]interface{}{
				"duration": duration.String(),
				"expected_max": "2m",
			})
		}
		return
	}
	
	logData["updated_count"] = updated
	
	// Проверяем производительность
	if duration > 2*time.Minute {
		logData["performance_warning"] = "slow_execution"
		w.logger.Warn(ctx, "Stale data update completed but was slow", logData)
	} else if updated > 0 {
		w.logger.Info(ctx, "Updated stale participants data", logData)
	} else {
		w.logger.Debug(ctx, "No stale participants data to update", logData)
	}
	
	// Дополнительная метрика производительности
	if updated > 0 {
		itemsPerSecond := float64(updated) / duration.Seconds()
		w.logger.Debug(ctx, "Stale update performance metrics", map[string]interface{}{
			"items_per_second": fmt.Sprintf("%.2f", itemsPerSecond),
			"total_items": updated,
		})
	}
}

// performFullUpdate выполняет полное обновление всех чатов
func (w *ParticipantsWorker) performFullUpdate() {
	ctx, cancel := context.WithTimeout(w.ctx, 2*time.Hour)
	defer cancel()
	
	w.logger.Info(ctx, "Starting full participants update", map[string]interface{}{
		"batch_size": w.config.BatchSize,
		"timeout": "2h",
		"scheduled_hour": w.config.FullUpdateHour,
	})
	
	start := time.Now()
	updated, err := w.updater.UpdateAll(ctx, w.config.BatchSize)
	duration := time.Since(start)
	
	logData := map[string]interface{}{
		"duration": duration.String(),
		"batch_size": w.config.BatchSize,
	}
	
	if err != nil {
		logData["error"] = err.Error()
		w.logger.Error(ctx, "Failed to perform full participants update", logData)
		
		// Проверяем, не превысили ли мы таймаут
		if duration > 90*time.Minute {
			w.logger.Error(ctx, "Full update approaching timeout limit", map[string]interface{}{
				"duration": duration.String(),
				"timeout_limit": "2h",
			})
		}
		return
	}
	
	logData["updated_count"] = updated
	
	// Анализ производительности
	if updated > 0 {
		itemsPerSecond := float64(updated) / duration.Seconds()
		itemsPerMinute := itemsPerSecond * 60
		
		logData["items_per_second"] = fmt.Sprintf("%.2f", itemsPerSecond)
		logData["items_per_minute"] = fmt.Sprintf("%.0f", itemsPerMinute)
		
		// Предупреждения о производительности
		if duration > 1*time.Hour {
			logData["performance_warning"] = "slow_execution"
			w.logger.Warn(ctx, "Full update completed but took longer than expected", logData)
		} else {
			w.logger.Info(ctx, "Completed full participants update", logData)
		}
		
		// Дополнительные метрики
		w.logger.Info(ctx, "Full update performance summary", map[string]interface{}{
			"total_processed": updated,
			"processing_rate": fmt.Sprintf("%.0f items/min", itemsPerMinute),
			"total_duration": duration.String(),
			"average_time_per_item": fmt.Sprintf("%.2fms", duration.Seconds()*1000/float64(updated)),
		})
	} else {
		w.logger.Warn(ctx, "Full update completed but no items were updated", logData)
	}
}