package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"maxbot-service/internal/domain"
)

// ProfileManagementService предоставляет API для управления профилями пользователей
type ProfileManagementService struct {
	profileCache domain.ProfileCacheService
	maxAPIClient domain.MaxAPIClient
}

// NewProfileManagementService создает новый сервис управления профилями
func NewProfileManagementService(profileCache domain.ProfileCacheService, maxAPIClient domain.MaxAPIClient) *ProfileManagementService {
	return &ProfileManagementService{
		profileCache: profileCache,
		maxAPIClient: maxAPIClient,
	}
}

// GetProfile получает профиль пользователя по user_id (Requirements 5.4)
func (s *ProfileManagementService) GetProfile(ctx context.Context, userID string) (*domain.UserProfileCache, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	profile, err := s.profileCache.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Если профиль не найден, возвращаем пустой профиль (Requirements 3.5)
	if profile == nil {
		return &domain.UserProfileCache{
			UserID:      userID,
			Source:      domain.SourceDefault,
			LastUpdated: time.Now(),
		}, nil
	}

	return profile, nil
}

// UpdateProfile обновляет профиль пользователя (Requirements 2.4, 5.5)
func (s *ProfileManagementService) UpdateProfile(ctx context.Context, userID string, updates domain.ProfileUpdates) (*domain.UserProfileCache, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Валидируем обновления
	if err := s.validateProfileUpdates(updates); err != nil {
		return nil, fmt.Errorf("invalid profile updates: %w", err)
	}

	// Применяем обновления
	err := s.profileCache.UpdateProfile(ctx, userID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	// Возвращаем обновленный профиль
	updatedProfile, err := s.profileCache.GetProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated profile: %w", err)
	}

	log.Printf("Profile updated for user_id=%s", userID)
	return updatedProfile, nil
}

// SetUserProvidedName устанавливает имя, предоставленное пользователем (Requirements 2.2, 2.4)
func (s *ProfileManagementService) SetUserProvidedName(ctx context.Context, userID, name string) (*domain.UserProfileCache, error) {
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	// Валидируем имя
	if err := s.validateUserProvidedName(name); err != nil {
		return nil, fmt.Errorf("invalid name: %w", err)
	}

	// Обновляем профиль
	updates := domain.ProfileUpdates{
		UserProvidedName: &name,
		Source:           &[]domain.ProfileSource{domain.SourceUserInput}[0],
	}

	return s.UpdateProfile(ctx, userID, updates)
}

// GetProfileStats возвращает статистику профилей (Requirements 6.1, 6.3)
func (s *ProfileManagementService) GetProfileStats(ctx context.Context) (*domain.ProfileStats, error) {
	stats, err := s.profileCache.GetProfileStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile stats: %w", err)
	}

	return stats, nil
}

// RequestNameFromUser отправляет запрос пользователю на предоставление имени (Requirements 2.1)
func (s *ProfileManagementService) RequestNameFromUser(ctx context.Context, userID string, chatID int64) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	// Проверяем, нужно ли запрашивать имя
	profile, err := s.profileCache.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Error getting profile for name request check: %v", err)
		// Продолжаем, так как это не критическая ошибка
	}

	// Если у пользователя уже есть полное имя, не запрашиваем
	if profile != nil && profile.HasFullName() {
		log.Printf("User %s already has full name, skipping request", userID)
		return nil
	}

	// Формируем сообщение с запросом имени
	message := s.buildNameRequestMessage(profile)

	// Отправляем сообщение пользователю
	_, err = s.maxAPIClient.SendMessage(ctx, chatID, 0, message) // userID в int64 не нужен для отправки в чат
	if err != nil {
		return fmt.Errorf("failed to send name request message: %w", err)
	}

	log.Printf("Name request sent to user_id=%s in chat_id=%d", userID, chatID)
	return nil
}

// buildNameRequestMessage формирует сообщение с запросом имени
func (s *ProfileManagementService) buildNameRequestMessage(profile *domain.UserProfileCache) string {
	if profile == nil || (profile.MaxFirstName == "" && profile.MaxLastName == "") {
		return "Привет! Для лучшего взаимодействия, пожалуйста, укажите ваше имя и фамилию.\n" +
			"Напишите: \"Меня зовут [Ваше имя и фамилия]\""
	}

	if profile.MaxLastName == "" {
		return fmt.Sprintf("Привет, %s! Для полного профиля, пожалуйста, укажите вашу фамилию.\n"+
			"Напишите: \"Меня зовут %s [Ваша фамилия]\"", 
			profile.MaxFirstName, profile.MaxFirstName)
	}

	return "Ваш профиль уже заполнен. Если хотите изменить имя, напишите: \"Меня зовут [Новое имя]\""
}

// ListProfiles возвращает список профилей с пагинацией (для административных целей)
func (s *ProfileManagementService) ListProfiles(ctx context.Context, limit, offset int) ([]*domain.UserProfileCache, error) {
	// Эта функция может быть реализована позже для административного интерфейса
	// Пока возвращаем ошибку "не реализовано"
	return nil, fmt.Errorf("list profiles not implemented yet")
}

// DeleteProfile удаляет профиль пользователя (для административных целей)
func (s *ProfileManagementService) DeleteProfile(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}

	// Эта функция может быть реализована позже
	// Пока возвращаем ошибку "не реализовано"
	return fmt.Errorf("delete profile not implemented yet")
}

// validateProfileUpdates валидирует обновления профиля
func (s *ProfileManagementService) validateProfileUpdates(updates domain.ProfileUpdates) error {
	if updates.MaxFirstName != nil {
		if err := s.validateName(*updates.MaxFirstName); err != nil {
			return fmt.Errorf("invalid max_first_name: %w", err)
		}
	}

	if updates.MaxLastName != nil {
		if err := s.validateName(*updates.MaxLastName); err != nil {
			return fmt.Errorf("invalid max_last_name: %w", err)
		}
	}

	if updates.UserProvidedName != nil {
		if err := s.validateUserProvidedName(*updates.UserProvidedName); err != nil {
			return fmt.Errorf("invalid user_provided_name: %w", err)
		}
	}

	if updates.Source != nil {
		if err := s.validateProfileSource(*updates.Source); err != nil {
			return fmt.Errorf("invalid source: %w", err)
		}
	}

	return nil
}

// validateName валидирует имя или фамилию
func (s *ProfileManagementService) validateName(name string) error {
	name = strings.TrimSpace(name)
	
	if len(name) > 50 {
		return fmt.Errorf("name too long: %d characters (max 50)", len(name))
	}
	
	return nil
}

// validateUserProvidedName валидирует имя, предоставленное пользователем
func (s *ProfileManagementService) validateUserProvidedName(name string) error {
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

// validateProfileSource валидирует источник профиля
func (s *ProfileManagementService) validateProfileSource(source domain.ProfileSource) error {
	switch source {
	case domain.SourceWebhook, domain.SourceUserInput, domain.SourceDefault:
		return nil
	default:
		return fmt.Errorf("unknown profile source: %s", source)
	}
}