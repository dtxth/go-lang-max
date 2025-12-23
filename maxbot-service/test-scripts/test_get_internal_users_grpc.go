package main

import (
	"context"
	"log"
	"time"

	maxbotproto "maxbot-service/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
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
			"+79991234567", // Should get Петр Петров
			"+79995678901", // Should get Анна Сидорова
			"+79999999999", // Should get Мария Иванова (no username)
			"+79991111111", // Should get Иван Иванов (default)
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

	log.Println("✅ GetInternalUsers gRPC test completed successfully!")

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

	log.Println("✅ All gRPC tests completed successfully!")
}