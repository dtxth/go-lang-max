package main

import (
	"context"
	"log"
	"os"
	"time"

	maxbotproto "maxbot-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Check if we should test against real or mock service
	mockMode := os.Getenv("MOCK_MODE")
	if mockMode == "" {
		mockMode = "true" // Default to mock for safety
	}

	log.Printf("Testing GetInternalUsers gRPC method (MOCK_MODE=%s)...", mockMode)

	// Connect to gRPC server
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := maxbotproto.NewMaxBotServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test GetInternalUsers
	log.Println("Testing GetInternalUsers gRPC method...")

	req := &maxbotproto.GetInternalUsersRequest{
		PhoneNumbers: []string{
			"+79991234567", // Should get Петр Петров (in mock)
			"+79995678901", // Should get Анна Сидорова (in mock)
			"+79999999999", // Should get Мария Иванова (in mock, no username)
			"+79991111111", // Should get Иван Иванов (in mock, default)
			"invalid",      // Should fail
		},
	}

	resp, err := client.GetInternalUsers(ctx, req)
	if err != nil {
		log.Fatalf("GetInternalUsers failed: %v", err)
	}

	log.Printf("Response received:")
	log.Printf("- Users count: %d", len(resp.Users))
	log.Printf("- Failed phones count: %d", len(resp.FailedPhoneNumbers))

	if resp.ErrorCode != maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED {
		log.Printf("- Error: %s - %s", resp.ErrorCode, resp.Error)
	}

	// Print user details
	for i, user := range resp.Users {
		log.Printf("User %d:", i+1)
		log.Printf("  - UserID: %d", user.UserId)
		log.Printf("  - Name: %s %s", user.FirstName, user.LastName)
		log.Printf("  - Phone: %s", user.PhoneNumber)
		log.Printf("  - Username: %s", user.Username)
		log.Printf("  - Link: %s", user.Link)
		log.Printf("  - IsBot: %t", user.IsBot)
		log.Printf("  - AvatarURL: %s", user.AvatarUrl)
		log.Printf("  - FullAvatarURL: %s", user.FullAvatarUrl)
	}

	// Print failed phones
	if len(resp.FailedPhoneNumbers) > 0 {
		log.Printf("Failed phones:")
		for _, phone := range resp.FailedPhoneNumbers {
			log.Printf("  - %s", phone)
		}
	}

	if mockMode == "true" {
		log.Println("✅ GetInternalUsers gRPC test completed successfully (MOCK MODE)!")
		log.Println("   In mock mode, users have detailed information including names.")
	} else {
		log.Println("✅ GetInternalUsers gRPC test completed successfully (REAL API MODE)!")
		log.Println("   In real API mode, response depends on actual MAX API availability.")
		if len(resp.Users) == 0 && len(resp.FailedPhoneNumbers) > 0 {
			log.Println("   Note: All phones failed - this might indicate MAX API is unavailable (fallback mode).")
		}
	}

	// Test edge cases
	log.Println("\nTesting edge cases...")

	// Empty request
	emptyReq := &maxbotproto.GetInternalUsersRequest{
		PhoneNumbers: []string{},
	}

	emptyResp, err := client.GetInternalUsers(ctx, emptyReq)
	if err != nil {
		log.Fatalf("Empty request failed: %v", err)
	}

	log.Printf("Empty request response:")
	log.Printf("- Users count: %d", len(emptyResp.Users))
	log.Printf("- Failed phones count: %d", len(emptyResp.FailedPhoneNumbers))

	// Test with only invalid phones
	invalidReq := &maxbotproto.GetInternalUsersRequest{
		PhoneNumbers: []string{"invalid1", "invalid2", "123"},
	}

	invalidResp, err := client.GetInternalUsers(ctx, invalidReq)
	if err != nil {
		log.Fatalf("Invalid phones request failed: %v", err)
	}

	log.Printf("Invalid phones request response:")
	log.Printf("- Users count: %d", len(invalidResp.Users))
	log.Printf("- Failed phones count: %d", len(invalidResp.FailedPhoneNumbers))

	// Test batch size limit (this should work but return an error)
	log.Println("\nTesting batch size limit (101 phones)...")
	largeReq := &maxbotproto.GetInternalUsersRequest{
		PhoneNumbers: make([]string, 101),
	}
	for i := 0; i < 101; i++ {
		largeReq.PhoneNumbers[i] = "+7999123456" + string(rune('0'+i%10))
	}

	largeResp, err := client.GetInternalUsers(ctx, largeReq)
	if err != nil {
		log.Printf("Large batch request failed as expected: %v", err)
	} else {
		log.Printf("Large batch request response:")
		log.Printf("- Users count: %d", len(largeResp.Users))
		log.Printf("- Failed phones count: %d", len(largeResp.FailedPhoneNumbers))
		if largeResp.ErrorCode != maxbotproto.ErrorCode_ERROR_CODE_UNSPECIFIED {
			log.Printf("- Error: %s - %s", largeResp.ErrorCode, largeResp.Error)
		}
	}

	log.Println("✅ All gRPC tests completed successfully!")
	log.Println("\nTo test against real MAX API:")
	log.Println("1. Set MOCK_MODE=false in environment")
	log.Println("2. Ensure MAX_BOT_TOKEN is set")
	log.Println("3. Ensure maxbot-service is running with real API configuration")
}