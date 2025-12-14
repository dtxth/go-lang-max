package notification

import (
	"context"
	"testing"

	"auth-service/internal/infrastructure/logger"
)

func TestNewMaxNotificationService(t *testing.T) {
	log := logger.NewDefault()
	
	service, err := NewMaxNotificationService("localhost:9999", log)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if service == nil {
		t.Error("Expected service to be created")
	}
	
	// Test close
	err = service.Close()
	if err != nil {
		t.Errorf("Expected no error on close, got %v", err)
	}
}

func TestMaxNotificationService_SendPasswordNotification(t *testing.T) {
	log := logger.NewDefault()
	
	service, err := NewMaxNotificationService("localhost:9999", log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()
	
	ctx := context.Background()
	
	// Test with valid phone and password
	err = service.SendPasswordNotification(ctx, "+79001234567", "testpassword123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMaxNotificationService_SendResetTokenNotification(t *testing.T) {
	log := logger.NewDefault()
	
	service, err := NewMaxNotificationService("localhost:9999", log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()
	
	ctx := context.Background()
	
	// Test with valid phone and token
	err = service.SendResetTokenNotification(ctx, "+79001234567", "ABC123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMaxNotificationService_SendPasswordNotification_EmptyPhone(t *testing.T) {
	log := logger.NewDefault()
	
	service, err := NewMaxNotificationService("localhost:9999", log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()
	
	ctx := context.Background()
	
	// Test with empty phone - should still work in mock implementation
	err = service.SendPasswordNotification(ctx, "", "testpassword123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestMaxNotificationService_SendResetTokenNotification_EmptyToken(t *testing.T) {
	log := logger.NewDefault()
	
	service, err := NewMaxNotificationService("localhost:9999", log)
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}
	defer service.Close()
	
	ctx := context.Background()
	
	// Test with empty token - should still work in mock implementation
	err = service.SendResetTokenNotification(ctx, "+79001234567", "")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}