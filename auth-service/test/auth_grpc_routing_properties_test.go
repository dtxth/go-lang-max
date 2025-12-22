package test

import (
	"auth-service/api/proto"
	"auth-service/internal/infrastructure/grpc"
	"auth-service/internal/usecase"
	"context"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/prop"
)

// **Feature: gateway-grpc-implementation, Property 1: HTTP-to-gRPC Routing Correctness (Auth endpoints)**
// **Validates: Requirements 2.1-2.7**

func TestAuthGRPCRoutingCorrectness(t *testing.T) {
	// Create a minimal auth service for testing (without database dependencies)
	authService := &usecase.AuthService{}
	grpcHandler := grpc.NewAuthHandler(authService)

	properties := gopter.NewProperties(nil)

	// Property: Health method always returns healthy status
	properties.Property("Health method returns healthy status", prop.ForAll(
		func() bool {
			ctx := context.Background()
			resp, err := grpcHandler.Health(ctx, &proto.HealthRequest{})
			return err == nil && resp.Status == "healthy"
		},
	))

	// Property: Register method validates empty input correctly
	properties.Property("Register method validates empty input", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.RegisterRequest{
				Email:    "",
				Phone:    "",
				Password: "",
				Role:     "",
			}
			
			resp, err := grpcHandler.Register(ctx, req)
			
			// Should never return a gRPC error, only business logic errors in response
			if err != nil {
				return false
			}
			
			// Should return validation error for missing email/phone
			return resp.Error != "" && strings.Contains(resp.Error, "either email or phone is required")
		},
	))

	// Property: Login method validates empty email
	properties.Property("Login method validates empty email", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.LoginRequest{
				Email:    "",
				Password: "somepassword",
			}
			
			resp, err := grpcHandler.Login(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing email
			return resp.Error != "" && strings.Contains(resp.Error, "email is required")
		},
	))

	// Property: Login method validates empty password
	properties.Property("Login method validates empty password", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.LoginRequest{
				Email:    "test@example.com",
				Password: "",
			}
			
			resp, err := grpcHandler.Login(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing password
			return resp.Error != "" && strings.Contains(resp.Error, "password is required")
		},
	))

	// Property: LoginByPhone method validates empty phone
	properties.Property("LoginByPhone method validates empty phone", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.LoginByPhoneRequest{
				Phone:    "",
				Password: "somepassword",
			}
			
			resp, err := grpcHandler.LoginByPhone(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing phone
			return resp.Error != "" && strings.Contains(resp.Error, "phone is required")
		},
	))

	// Property: LoginByPhone method validates empty password
	properties.Property("LoginByPhone method validates empty password", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.LoginByPhoneRequest{
				Phone:    "+1234567890",
				Password: "",
			}
			
			resp, err := grpcHandler.LoginByPhone(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing password
			return resp.Error != "" && strings.Contains(resp.Error, "password is required")
		},
	))

	// Property: Refresh method validates empty token
	properties.Property("Refresh method validates empty token", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.RefreshRequest{
				RefreshToken: "",
			}
			
			resp, err := grpcHandler.Refresh(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing refresh token
			return resp.Error != "" && strings.Contains(resp.Error, "refresh token is required")
		},
	))

	// Property: Logout method validates empty token
	properties.Property("Logout method validates empty token", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.LogoutRequest{
				RefreshToken: "",
			}
			
			resp, err := grpcHandler.Logout(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing refresh token
			return !resp.Success && resp.Error != "" && strings.Contains(resp.Error, "refresh token is required")
		},
	))

	// Property: AuthenticateMAX method validates empty init_data
	properties.Property("AuthenticateMAX method validates empty init_data", prop.ForAll(
		func() bool {
			ctx := context.Background()
			req := &proto.AuthenticateMAXRequest{
				InitData: "",
			}
			
			resp, err := grpcHandler.AuthenticateMAX(ctx, req)
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should return validation error for missing init_data
			return resp.Error != "" && strings.Contains(resp.Error, "init_data is required")
		},
	))

	// Property: GetBotMe method always returns structured response
	properties.Property("GetBotMe method returns structured response", prop.ForAll(
		func() bool {
			ctx := context.Background()
			resp, err := grpcHandler.GetBotMe(ctx, &proto.GetBotMeRequest{})
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should always return structured response
			return resp != nil
		},
	))

	// Property: GetMetrics method always returns structured response
	properties.Property("GetMetrics method returns structured response", prop.ForAll(
		func() bool {
			ctx := context.Background()
			resp, err := grpcHandler.GetMetrics(ctx, &proto.GetMetricsRequest{})
			
			// Should never return a gRPC error
			if err != nil {
				return false
			}
			
			// Should always return structured response
			return resp != nil
		},
	))

	// Property: All gRPC methods are callable without panicking
	properties.Property("All gRPC methods are callable", prop.ForAll(
		func() bool {
			ctx := context.Background()
			
			// Test that all methods can be called without panicking
			methods := []func() bool{
				func() bool {
					_, err := grpcHandler.Health(ctx, &proto.HealthRequest{})
					return err == nil
				},
				func() bool {
					_, err := grpcHandler.GetBotMe(ctx, &proto.GetBotMeRequest{})
					return err == nil
				},
				func() bool {
					_, err := grpcHandler.GetMetrics(ctx, &proto.GetMetricsRequest{})
					return err == nil
				},
			}
			
			for _, method := range methods {
				if !method() {
					return false
				}
			}
			
			return true
		},
	))

	// Run all properties with 100 iterations each
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

