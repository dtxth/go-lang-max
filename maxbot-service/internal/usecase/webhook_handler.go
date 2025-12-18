package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"maxbot-service/internal/domain"
)

// WebhookHandlerService реализует обработку webhook событий от MAX
type WebhookHandlerService struct {
	profileCache domain.ProfileCacheService
	monitoring   domain.MonitoringService
}

// NewWebhookHandlerService создает новый обработчик webhook событий
func NewWebhookHandlerService(profileCache domain.ProfileCacheService, monitoring domain.MonitoringService) *WebhookHandlerService {
	return &WebhookHandlerService{
		profileCache: profileCache,
		monitoring:   monitoring,
	}
}

// HandleMaxWebhook обрабатывает входящее webhook событие от MAX
func (h *WebhookHandlerService) HandleMaxWebhook(ctx context.Context, event domain.MaxWebhookEvent) error {
	startTime := time.Now()
	log.Printf("Processing webhook event: type=%s", event.Type)

	// Добавляем таймаут для обработки webhook события (Requirements 4.4, 4.5)
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var userInfo *domain.UserInfo
	var eventType string
	var messageText string
	var processingError error
	var profileFound bool
	var profileStored bool

	// Извлекаем информацию о пользователе в зависимости от типа события
	switch event.Type {
	case "message_new":
		if event.Message != nil {
			userInfo = &event.Message.From
			eventType = "message_new"
			messageText = event.Message.Text
			profileFound = userInfo.FirstName != "" || userInfo.LastName != ""
		}
	case "callback_query":
		if event.Callback != nil {
			userInfo = &event.Callback.User
			eventType = "callback_query"
			profileFound = userInfo.FirstName != "" || userInfo.LastName != ""
		}
	default:
		log.Printf("Unknown webhook event type: %s", event.Type)
		// Записываем метрику для неизвестного события
		h.recordWebhookMetric(ctx, domain.WebhookEventMetric{
			EventType:      event.Type,
			ProcessedAt:    startTime,
			Success:        false,
			ErrorMessage:   "unknown event type",
			ProcessingTime: time.Since(startTime).Milliseconds(),
			ProfileFound:   false,
			ProfileStored:  false,
		})
		// Возвращаем nil, чтобы вернуть 200 OK даже для неизвестных событий (Requirements 4.4)
		return nil
	}

	// Если не удалось извлечь информацию о пользователе, логируем и продолжаем (Requirements 4.4)
	if userInfo == nil || userInfo.UserID == "" {
		log.Printf("No user info found in webhook event type: %s", event.Type)
		h.recordWebhookMetric(ctx, domain.WebhookEventMetric{
			EventType:      eventType,
			ProcessedAt:    startTime,
			Success:        false,
			ErrorMessage:   "no user info",
			ProcessingTime: time.Since(startTime).Milliseconds(),
			ProfileFound:   false,
			ProfileStored:  false,
		})
		return nil
	}

	// Валидируем данные пользователя
	if err := h.validateUserInfo(userInfo); err != nil {
		log.Printf("Invalid user info in webhook event: %v", err)
		h.recordWebhookMetric(ctx, domain.WebhookEventMetric{
			EventType:      eventType,
			UserID:         userInfo.UserID,
			ProcessedAt:    startTime,
			Success:        false,
			ErrorMessage:   err.Error(),
			ProcessingTime: time.Since(startTime).Milliseconds(),
			ProfileFound:   profileFound,
			ProfileStored:  false,
		})
		// Возвращаем nil для graceful degradation (Requirements 4.4)
		return nil
	}

	// Сначала обрабатываем профиль пользователя с retry логикой
	err := h.processUserProfileWithRetry(ctx, userInfo, eventType)
	if err != nil {
		// Логируем ошибку, но не возвращаем её, чтобы webhook получил 200 OK (Requirements 4.5)
		log.Printf("Error processing user profile for user_id=%s: %v", userInfo.UserID, err)
		processingError = err
	} else {
		profileStored = profileFound // Если обработка успешна и профиль найден, значит он сохранен
	}

	// Затем обрабатываем пользовательский ввод для обновления имени (Requirements 2.2, 2.4)
	// Это должно быть после обработки webhook профиля, чтобы user_input имел приоритет
	if eventType == "message_new" && messageText != "" {
		err := h.processUserNameInput(ctx, userInfo.UserID, messageText)
		if err != nil {
			log.Printf("Error processing user name input for user_id=%s: %v", userInfo.UserID, err)
			if processingError == nil {
				processingError = err
			}
		}
	}

	// Записываем метрику обработки события (Requirements 6.1, 6.3)
	metric := domain.WebhookEventMetric{
		EventType:      eventType,
		UserID:         userInfo.UserID,
		ProcessedAt:    startTime,
		Success:        processingError == nil,
		ProcessingTime: time.Since(startTime).Milliseconds(),
		ProfileFound:   profileFound,
		ProfileStored:  profileStored,
	}
	
	if processingError != nil {
		metric.ErrorMessage = processingError.Error()
	}
	
	h.recordWebhookMetric(ctx, metric)

	return nil
}

// validateUserInfo валидирует данные пользователя из webhook события
func (h *WebhookHandlerService) validateUserInfo(userInfo *domain.UserInfo) error {
	if userInfo.UserID == "" {
		return fmt.Errorf("user_id is required")
	}
	
	// Проверяем длину user_id (разумные ограничения)
	if len(userInfo.UserID) > 100 {
		return fmt.Errorf("user_id too long: %d characters", len(userInfo.UserID))
	}
	
	// Проверяем длину имен (если они есть)
	if len(userInfo.FirstName) > 100 {
		return fmt.Errorf("first_name too long: %d characters", len(userInfo.FirstName))
	}
	if len(userInfo.LastName) > 100 {
		return fmt.Errorf("last_name too long: %d characters", len(userInfo.LastName))
	}
	
	return nil
}

// processUserProfileWithRetry обрабатывает профиль пользователя с retry логикой
func (h *WebhookHandlerService) processUserProfileWithRetry(ctx context.Context, userInfo *domain.UserInfo, eventType string) error {
	const maxRetries = 3
	const baseDelay = 100 * time.Millisecond
	
	var lastErr error
	
	for attempt := 0; attempt < maxRetries; attempt++ {
		err := h.processUserProfile(ctx, userInfo, eventType)
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// Если это последняя попытка, не ждем
		if attempt == maxRetries-1 {
			break
		}
		
		// Exponential backoff
		delay := baseDelay * time.Duration(1<<attempt)
		log.Printf("Profile processing failed for user_id=%s (attempt %d/%d), retrying in %v: %v", 
			userInfo.UserID, attempt+1, maxRetries, delay, err)
		
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Продолжаем к следующей попытке
		}
	}
	
	return fmt.Errorf("failed to process profile after %d attempts: %w", maxRetries, lastErr)
}

// processUserProfile обрабатывает и сохраняет профиль пользователя
func (h *WebhookHandlerService) processUserProfile(ctx context.Context, userInfo *domain.UserInfo, eventType string) error {
	// Получаем существующий профиль из кэша с таймаутом
	existingProfile, err := h.getExistingProfileSafely(ctx, userInfo.UserID)
	if err != nil {
		log.Printf("Error getting existing profile for user_id=%s: %v", userInfo.UserID, err)
		// Продолжаем обработку даже если не удалось получить существующий профиль (Requirements 3.4)
	}

	// Создаем новый профиль или обновляем существующий
	profile := domain.UserProfileCache{
		UserID:      userInfo.UserID,
		LastUpdated: time.Now(),
		Source:      domain.SourceWebhook,
	}

	// Если есть существующий профиль, сохраняем user_provided_name (Requirements 5.2)
	if existingProfile != nil {
		profile.UserProvidedName = existingProfile.UserProvidedName
	}

	// Обновляем данные из webhook события (Requirements 1.2, 1.3)
	if userInfo.FirstName != "" {
		profile.MaxFirstName = userInfo.FirstName
	}
	if userInfo.LastName != "" {
		profile.MaxLastName = userInfo.LastName
	}

	// Если у нас есть существующий профиль, сохраняем данные которых нет в новом событии (Requirements 5.2)
	if existingProfile != nil {
		if profile.MaxFirstName == "" && existingProfile.MaxFirstName != "" {
			profile.MaxFirstName = existingProfile.MaxFirstName
		}
		if profile.MaxLastName == "" && existingProfile.MaxLastName != "" {
			profile.MaxLastName = existingProfile.MaxLastName
		}
	}

	// Сохраняем профиль в кэше с обработкой ошибок
	err = h.storeProfileSafely(ctx, userInfo.UserID, profile)
	if err != nil {
		return fmt.Errorf("failed to store profile for user_id=%s: %w", userInfo.UserID, err)
	}

	log.Printf("Profile updated for user_id=%s, first_name=%s, last_name=%s, event_type=%s", 
		userInfo.UserID, userInfo.FirstName, userInfo.LastName, eventType)

	return nil
}

// getExistingProfileSafely безопасно получает существующий профиль с обработкой ошибок
func (h *WebhookHandlerService) getExistingProfileSafely(ctx context.Context, userID string) (*domain.UserProfileCache, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	
	return h.profileCache.GetProfile(ctx, userID)
}

// storeProfileSafely безопасно сохраняет профиль с обработкой ошибок
func (h *WebhookHandlerService) storeProfileSafely(ctx context.Context, userID string, profile domain.UserProfileCache) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	
	return h.profileCache.StoreProfile(ctx, userID, profile)
}

// processUserNameInput обрабатывает пользовательский ввод для обновления имени (Requirements 2.2, 2.4)
func (h *WebhookHandlerService) processUserNameInput(ctx context.Context, userID, messageText string) error {
	// Проверяем, является ли сообщение командой для обновления имени
	if !h.isNameUpdateCommand(messageText) {
		return nil // Не команда обновления имени, игнорируем
	}

	// Извлекаем имя из сообщения
	userName := h.extractUserNameFromMessage(messageText)
	if userName == "" {
		log.Printf("Empty name provided by user_id=%s", userID)
		return nil
	}

	// Валидируем предоставленное имя
	if err := h.validateUserProvidedName(userName); err != nil {
		log.Printf("Invalid name provided by user_id=%s: %v", userID, err)
		return nil
	}

	// Обновляем профиль с пользовательским именем
	updates := domain.ProfileUpdates{
		UserProvidedName: &userName,
		Source:           &[]domain.ProfileSource{domain.SourceUserInput}[0],
	}

	err := h.profileCache.UpdateProfile(ctx, userID, updates)
	if err != nil {
		return fmt.Errorf("failed to update profile with user-provided name: %w", err)
	}

	log.Printf("User-provided name updated for user_id=%s: %s", userID, userName)
	return nil
}

// isNameUpdateCommand проверяет, является ли сообщение командой для обновления имени
func (h *WebhookHandlerService) isNameUpdateCommand(messageText string) bool {
	// Простые команды для обновления имени
	commands := []string{
		"/setname",
		"/имя",
		"меня зовут",
		"мое имя",
		"моё имя",
	}

	messageTextLower := strings.ToLower(strings.TrimSpace(messageText))
	
	for _, cmd := range commands {
		if strings.HasPrefix(messageTextLower, cmd) {
			return true
		}
	}
	
	return false
}

// extractUserNameFromMessage извлекает имя пользователя из сообщения
func (h *WebhookHandlerService) extractUserNameFromMessage(messageText string) string {
	messageText = strings.TrimSpace(messageText)
	
	// Удаляем команду из начала сообщения
	commands := []string{
		"/setname",
		"/имя",
		"меня зовут",
		"мое имя",
		"моё имя",
	}
	
	messageTextLower := strings.ToLower(messageText)
	
	for _, cmd := range commands {
		if strings.HasPrefix(messageTextLower, cmd) {
			// Удаляем команду и возвращаем оставшуюся часть
			name := strings.TrimSpace(messageText[len(cmd):])
			return name
		}
	}
	
	return ""
}

// validateUserProvidedName валидирует имя, предоставленное пользователем
func (h *WebhookHandlerService) validateUserProvidedName(name string) error {
	name = strings.TrimSpace(name)
	
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}
	
	if len(name) > 100 {
		return fmt.Errorf("name too long: %d characters (max 100)", len(name))
	}
	
	// Проверяем на недопустимые символы (только буквы, пробелы, дефисы)
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			 (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я') || 
			 r == ' ' || r == '-' || r == 'ё' || r == 'Ё') {
			return fmt.Errorf("name contains invalid characters")
		}
	}
	
	return nil
}

// recordWebhookMetric записывает метрику обработки webhook события
func (h *WebhookHandlerService) recordWebhookMetric(ctx context.Context, metric domain.WebhookEventMetric) {
	if h.monitoring == nil {
		return // Мониторинг не настроен
	}
	
	// Используем отдельный контекст с коротким таймаутом для записи метрик
	metricCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	if err := h.monitoring.RecordWebhookEvent(metricCtx, metric); err != nil {
		log.Printf("Failed to record webhook metric: %v", err)
		// Не возвращаем ошибку, чтобы не влиять на основную обработку
	}
}