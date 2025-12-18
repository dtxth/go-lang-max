package cache

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"maxbot-service/internal/domain"
)

func TestProfileRedisCache_StoreAndGetProfile(t *testing.T) {
	// Используем in-memory Redis для тестов
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // Используем отдельную БД для тестов
	})
	
	// Проверяем соединение
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	
	// Очищаем тестовую БД
	client.FlushDB(ctx)
	defer client.FlushDB(ctx)
	
	cache := NewProfileRedisCache(client, time.Hour)
	
	// Тестовый профиль
	userID := "test_user_123"
	profile := domain.UserProfileCache{
		UserID:       userID,
		MaxFirstName: "Иван",
		MaxLastName:  "Петров",
		Source:       domain.SourceWebhook,
	}
	
	// Сохраняем профиль
	err = cache.StoreProfile(ctx, userID, profile)
	if err != nil {
		t.Fatalf("Failed to store profile: %v", err)
	}
	
	// Получаем профиль
	retrieved, err := cache.GetProfile(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get profile: %v", err)
	}
	
	if retrieved == nil {
		t.Fatal("Retrieved profile is nil")
	}
	
	// Проверяем данные
	if retrieved.UserID != profile.UserID {
		t.Errorf("Expected UserID %s, got %s", profile.UserID, retrieved.UserID)
	}
	if retrieved.MaxFirstName != profile.MaxFirstName {
		t.Errorf("Expected MaxFirstName %s, got %s", profile.MaxFirstName, retrieved.MaxFirstName)
	}
	if retrieved.MaxLastName != profile.MaxLastName {
		t.Errorf("Expected MaxLastName %s, got %s", profile.MaxLastName, retrieved.MaxLastName)
	}
	if retrieved.Source != profile.Source {
		t.Errorf("Expected Source %s, got %s", profile.Source, retrieved.Source)
	}
}

func TestProfileRedisCache_GetNonExistentProfile(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	
	client.FlushDB(ctx)
	defer client.FlushDB(ctx)
	
	cache := NewProfileRedisCache(client, time.Hour)
	
	// Пытаемся получить несуществующий профиль
	profile, err := cache.GetProfile(ctx, "nonexistent_user")
	if err != nil {
		t.Fatalf("Expected no error for non-existent profile, got: %v", err)
	}
	
	if profile != nil {
		t.Fatal("Expected nil profile for non-existent user")
	}
}

func TestProfileRedisCache_UpdateProfile(t *testing.T) {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15,
	})
	
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Skipf("Redis not available: %v", err)
	}
	
	client.FlushDB(ctx)
	defer client.FlushDB(ctx)
	
	cache := NewProfileRedisCache(client, time.Hour)
	
	userID := "test_user_456"
	
	// Создаем начальный профиль
	initialProfile := domain.UserProfileCache{
		UserID:       userID,
		MaxFirstName: "Иван",
		Source:       domain.SourceWebhook,
	}
	
	err = cache.StoreProfile(ctx, userID, initialProfile)
	if err != nil {
		t.Fatalf("Failed to store initial profile: %v", err)
	}
	
	// Обновляем профиль
	lastName := "Петров"
	userProvidedName := "Иван Петрович Петров"
	source := domain.SourceUserInput
	
	updates := domain.ProfileUpdates{
		MaxLastName:      &lastName,
		UserProvidedName: &userProvidedName,
		Source:           &source,
	}
	
	err = cache.UpdateProfile(ctx, userID, updates)
	if err != nil {
		t.Fatalf("Failed to update profile: %v", err)
	}
	
	// Получаем обновленный профиль
	updated, err := cache.GetProfile(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get updated profile: %v", err)
	}
	
	if updated == nil {
		t.Fatal("Updated profile is nil")
	}
	
	// Проверяем обновления
	if updated.MaxFirstName != "Иван" {
		t.Errorf("Expected MaxFirstName to remain 'Иван', got %s", updated.MaxFirstName)
	}
	if updated.MaxLastName != lastName {
		t.Errorf("Expected MaxLastName %s, got %s", lastName, updated.MaxLastName)
	}
	if updated.UserProvidedName != userProvidedName {
		t.Errorf("Expected UserProvidedName %s, got %s", userProvidedName, updated.UserProvidedName)
	}
	if updated.Source != source {
		t.Errorf("Expected Source %s, got %s", source, updated.Source)
	}
}

func TestUserProfileCache_GetDisplayName(t *testing.T) {
	tests := []struct {
		name     string
		profile  domain.UserProfileCache
		expected string
	}{
		{
			name: "User provided name has priority",
			profile: domain.UserProfileCache{
				MaxFirstName:     "Иван",
				MaxLastName:      "Петров",
				UserProvidedName: "Иван Петрович",
			},
			expected: "Иван Петрович",
		},
		{
			name: "Full MAX name when no user provided name",
			profile: domain.UserProfileCache{
				MaxFirstName: "Иван",
				MaxLastName:  "Петров",
			},
			expected: "Иван Петров",
		},
		{
			name: "Only first name when no last name",
			profile: domain.UserProfileCache{
				MaxFirstName: "Иван",
			},
			expected: "Иван",
		},
		{
			name: "Empty when no names",
			profile: domain.UserProfileCache{
				UserID: "123",
			},
			expected: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.GetDisplayName()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestUserProfileCache_HasFullName(t *testing.T) {
	tests := []struct {
		name     string
		profile  domain.UserProfileCache
		expected bool
	}{
		{
			name: "Has full name with user provided name",
			profile: domain.UserProfileCache{
				UserProvidedName: "Иван Петрович",
			},
			expected: true,
		},
		{
			name: "Has full name with MAX first and last name",
			profile: domain.UserProfileCache{
				MaxFirstName: "Иван",
				MaxLastName:  "Петров",
			},
			expected: true,
		},
		{
			name: "No full name with only first name",
			profile: domain.UserProfileCache{
				MaxFirstName: "Иван",
			},
			expected: false,
		},
		{
			name: "No full name when empty",
			profile: domain.UserProfileCache{
				UserID: "123",
			},
			expected: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.profile.HasFullName()
			if result != tt.expected {
				t.Errorf("Expected %t, got %t", tt.expected, result)
			}
		})
	}
}