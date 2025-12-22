package grpc

import (
	"auth-service/api/proto"
	"auth-service/internal/usecase"
	"context"
	"testing"
)

func TestAuthHandlerImplementsInterface(t *testing.T) {
	// Create a minimal auth service for testing
	authService := &usecase.AuthService{}
	
	// Create the handler
	handler := NewAuthHandler(authService)
	
	// Verify that handler implements the interface
	var _ proto.AuthServiceServer = handler
	
	// Test Health method (should work without dependencies)
	ctx := context.Background()
	resp, err := handler.Health(ctx, &proto.HealthRequest{})
	if err != nil {
		t.Errorf("Health method failed: %v", err)
	}
	if resp.Status != "healthy" {
		t.Errorf("Expected status 'healthy', got '%s'", resp.Status)
	}
	
	// Test that methods with validation return appropriate errors
	
	// Test Register with missing data
	registerResp, err := handler.Register(ctx, &proto.RegisterRequest{})
	if err != nil {
		t.Errorf("Register should not return error, got: %v", err)
	}
	if registerResp.Error == "" {
		t.Error("Register should return validation error for missing email/phone")
	}
	
	// Test Login with missing data
	loginResp, err := handler.Login(ctx, &proto.LoginRequest{})
	if err != nil {
		t.Errorf("Login should not return error, got: %v", err)
	}
	if loginResp.Error == "" {
		t.Error("Login should return validation error for missing email")
	}
	
	// Test LoginByPhone with missing data
	loginPhoneResp, err := handler.LoginByPhone(ctx, &proto.LoginByPhoneRequest{})
	if err != nil {
		t.Errorf("LoginByPhone should not return error, got: %v", err)
	}
	if loginPhoneResp.Error == "" {
		t.Error("LoginByPhone should return validation error for missing phone")
	}
	
	// Test Refresh with missing data
	refreshResp, err := handler.Refresh(ctx, &proto.RefreshRequest{})
	if err != nil {
		t.Errorf("Refresh should not return error, got: %v", err)
	}
	if refreshResp.Error == "" {
		t.Error("Refresh should return validation error for missing refresh token")
	}
	
	// Test Logout with missing data
	logoutResp, err := handler.Logout(ctx, &proto.LogoutRequest{})
	if err != nil {
		t.Errorf("Logout should not return error, got: %v", err)
	}
	if !logoutResp.Success && logoutResp.Error == "" {
		t.Error("Logout should return validation error for missing refresh token")
	}
	
	// Test AuthenticateMAX with missing data
	maxResp, err := handler.AuthenticateMAX(ctx, &proto.AuthenticateMAXRequest{})
	if err != nil {
		t.Errorf("AuthenticateMAX should not return error, got: %v", err)
	}
	if maxResp.Error == "" {
		t.Error("AuthenticateMAX should return validation error for missing init_data")
	}
	
	t.Log("All gRPC methods are properly implemented and callable")
}