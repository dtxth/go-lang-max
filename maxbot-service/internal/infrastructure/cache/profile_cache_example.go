package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"maxbot-service/internal/config"
	"maxbot-service/internal/domain"
)

// ExampleProfileCacheUsage демонстрирует использование ProfileCacheService
func ExampleProfileCacheUsage() {
	// Загружаем конфигурацию
	cfg := config.Load()
	
	// Создаем сервис кэширования профилей
	profileCache, err := NewProfileCacheService(cfg)
	if err != nil {
		log.Printf("Failed to create profile cache service: %v", err)
		return
	}
	
	ctx := context.Background()
	
	// Пример 1: Сохранение профиля из webhook события
	userID := "max_user_123456"
	webhookProfile := domain.UserProfileCache{
		UserID:       userID,
		MaxFirstName: "Иван",
		MaxLastName:  "Петров",
		Source:       domain.SourceWebhook,
	}
	
	err = profileCache.StoreProfile(ctx, userID, webhookProfile)
	if err != nil {
		log.Printf("Failed to store profile: %v", err)
		return
	}
	
	fmt.Printf("Stored profile for user %s\n", userID)
	
	// Пример 2: Получение профиля
	retrievedProfile, err := profileCache.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Failed to get profile: %v", err)
		return
	}
	
	if retrievedProfile != nil {
		fmt.Printf("Retrieved profile: %s (source: %s)\n", 
			retrievedProfile.GetDisplayName(), retrievedProfile.Source)
	}
	
	// Пример 3: Обновление профиля пользовательским вводом
	userProvidedName := "Иван Петрович Петров"
	source := domain.SourceUserInput
	
	updates := domain.ProfileUpdates{
		UserProvidedName: &userProvidedName,
		Source:           &source,
	}
	
	err = profileCache.UpdateProfile(ctx, userID, updates)
	if err != nil {
		log.Printf("Failed to update profile: %v", err)
		return
	}
	
	fmt.Printf("Updated profile with user-provided name\n")
	
	// Пример 4: Получение обновленного профиля
	updatedProfile, err := profileCache.GetProfile(ctx, userID)
	if err != nil {
		log.Printf("Failed to get updated profile: %v", err)
		return
	}
	
	if updatedProfile != nil {
		fmt.Printf("Updated profile: %s (source: %s, has full name: %t)\n", 
			updatedProfile.GetDisplayName(), updatedProfile.Source, updatedProfile.HasFullName())
	}
	
	// Пример 5: Получение статистики
	stats, err := profileCache.GetProfileStats(ctx)
	if err != nil {
		log.Printf("Failed to get profile stats: %v", err)
		return
	}
	
	fmt.Printf("Profile statistics:\n")
	fmt.Printf("  Total profiles: %d\n", stats.TotalProfiles)
	fmt.Printf("  Profiles with full name: %d\n", stats.ProfilesWithFullName)
	fmt.Printf("  Profiles by source:\n")
	for source, count := range stats.ProfilesBySource {
		fmt.Printf("    %s: %d\n", source, count)
	}
}

// ExampleWebhookProfileExtraction демонстрирует извлечение профиля из webhook события
func ExampleWebhookProfileExtraction() {
	// Пример структуры webhook события (упрощенная версия)
	type WebhookEvent struct {
		Type    string `json:"type"`
		Message struct {
			From struct {
				UserID    string `json:"user_id"`
				FirstName string `json:"first_name"`
				LastName  string `json:"last_name"`
			} `json:"from"`
		} `json:"message,omitempty"`
	}
	
	// Симуляция webhook события
	event := WebhookEvent{
		Type: "message_new",
	}
	event.Message.From.UserID = "max_user_789"
	event.Message.From.FirstName = "Мария"
	event.Message.From.LastName = "Иванова"
	
	// Извлекаем профиль из события
	profile := domain.UserProfileCache{
		UserID:       event.Message.From.UserID,
		MaxFirstName: event.Message.From.FirstName,
		MaxLastName:  event.Message.From.LastName,
		Source:       domain.SourceWebhook,
		LastUpdated:  time.Now(),
	}
	
	fmt.Printf("Extracted profile from webhook: %s (%s)\n", 
		profile.GetDisplayName(), profile.UserID)
	
	// Здесь бы мы сохранили профиль в кэше
	// profileCache.StoreProfile(ctx, profile.UserID, profile)
}

// ExampleNamePriorityLogic демонстрирует логику приоритета имен
func ExampleNamePriorityLogic() {
	profiles := []domain.UserProfileCache{
		{
			UserID:       "user1",
			MaxFirstName: "Иван",
			MaxLastName:  "Петров",
			Source:       domain.SourceWebhook,
		},
		{
			UserID:           "user2",
			MaxFirstName:     "Мария",
			MaxLastName:      "Иванова",
			UserProvidedName: "Мария Петровна Иванова",
			Source:           domain.SourceUserInput,
		},
		{
			UserID:       "user3",
			MaxFirstName: "Алексей",
			Source:       domain.SourceWebhook,
		},
		{
			UserID: "user4",
			Source: domain.SourceDefault,
		},
	}
	
	fmt.Println("Name priority examples:")
	for _, profile := range profiles {
		displayName := profile.GetDisplayName()
		if displayName == "" {
			displayName = "[no name]"
		}
		fmt.Printf("  User %s: %s (source: %s, has full name: %t)\n", 
			profile.UserID, displayName, profile.Source, profile.HasFullName())
	}
}