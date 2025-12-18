package usecase

import (
	"context"
	"testing"
)

func TestMaxBotService_GetMe(t *testing.T) {
	// Create mock client
	mockClient := NewMockMaxAPIClient()
	
	// Create service
	service := NewMaxBotService(mockClient)
	
	// Test GetMe
	ctx := context.Background()
	botInfo, err := service.GetMe(ctx)
	
	// Verify results
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	
	if botInfo == nil {
		t.Fatal("Expected bot info, got nil")
	}
	
	if botInfo.Name == "" {
		t.Error("Expected bot name to be set")
	}
	
	if botInfo.AddLink == "" {
		t.Error("Expected add link to be set")
	}
	
	// Verify expected values from mock
	expectedName := "Test Mock Bot"
	expectedLink := "https://max.ru/test-bot"
	
	if botInfo.Name != expectedName {
		t.Errorf("Expected name %s, got %s", expectedName, botInfo.Name)
	}
	
	if botInfo.AddLink != expectedLink {
		t.Errorf("Expected link %s, got %s", expectedLink, botInfo.AddLink)
	}
}