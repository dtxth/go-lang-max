package worker

import (
	"chat-service/internal/domain"
	"chat-service/internal/infrastructure/logger"
	"context"
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
	
	start := time.Now()
	updated, err := w.updater.UpdateStale(ctx, w.config.StaleThreshold, w.config.BatchSize)
	duration := time.Since(start)
	
	if err != nil {
		w.logger.Error(ctx, "Failed to update stale participants data", map[string]interface{}{
			"error": err.Error(), 
			"duration": duration.String(),
		})
		return
	}
	
	if updated > 0 {
		w.logger.Info(ctx, "Updated stale participants data", map[string]interface{}{
			"updated_count": updated, 
			"duration": duration.String(),
		})
	} else {
		w.logger.Debug(ctx, "No stale participants data to update", map[string]interface{}{
			"duration": duration.String(),
		})
	}
}

// performFullUpdate выполняет полное обновление всех чатов
func (w *ParticipantsWorker) performFullUpdate() {
	ctx, cancel := context.WithTimeout(w.ctx, 2*time.Hour)
	defer cancel()
	
	w.logger.Info(ctx, "Starting full participants update", nil)
	start := time.Now()
	
	updated, err := w.updater.UpdateAll(ctx, w.config.BatchSize)
	duration := time.Since(start)
	
	if err != nil {
		w.logger.Error(ctx, "Failed to perform full participants update", map[string]interface{}{
			"error": err.Error(), 
			"duration": duration.String(),
		})
		return
	}
	
	w.logger.Info(ctx, "Completed full participants update", map[string]interface{}{
		"updated_count": updated, 
		"duration": duration.String(),
	})
}