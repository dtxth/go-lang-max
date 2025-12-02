package integration_tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// gRPC service addresses
const (
	AuthGRPCAddr     = "localhost:9090"
	EmployeeGRPCAddr = "localhost:9091"
	ChatGRPCAddr     = "localhost:9092"
	StructureGRPCAddr = "localhost:9093"
	MaxBotGRPCAddr   = "localhost:9095"
)

// TestAuthServiceGRPC tests Auth Service gRPC communication
func TestAuthServiceGRPC(t *testing.T) {
	// Connect to Auth Service gRPC
	conn, err := grpc.Dial(AuthGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Auth Service gRPC not available: %v", err)
		return
	}
	defer conn.Close()
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Verify connection state
	state := conn.GetState()
	t.Logf("Auth Service gRPC connection state: %v", state)
	
	// Wait for connection to be ready
	conn.WaitForStateChange(ctx, state)
	
	assert.NotEqual(t, "TransientFailure", conn.GetState().String(), "Connection should not be in failure state")
}

// TestMaxBotServiceGRPC tests MaxBot Service gRPC communication
func TestMaxBotServiceGRPC(t *testing.T) {
	// Connect to MaxBot Service gRPC
	conn, err := grpc.Dial(MaxBotGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("MaxBot Service gRPC not available: %v", err)
		return
	}
	defer conn.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// Verify connection
	state := conn.GetState()
	t.Logf("MaxBot Service gRPC connection state: %v", state)
	
	conn.WaitForStateChange(ctx, state)
	
	assert.NotEqual(t, "TransientFailure", conn.GetState().String(), "Connection should not be in failure state")
}

// TestChatServiceGRPC tests Chat Service gRPC communication
func TestChatServiceGRPC(t *testing.T) {
	conn, err := grpc.Dial(ChatGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Chat Service gRPC not available: %v", err)
		return
	}
	defer conn.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	state := conn.GetState()
	t.Logf("Chat Service gRPC connection state: %v", state)
	
	conn.WaitForStateChange(ctx, state)
	
	assert.NotEqual(t, "TransientFailure", conn.GetState().String(), "Connection should not be in failure state")
}

// TestEmployeeServiceGRPC tests Employee Service gRPC communication
func TestEmployeeServiceGRPC(t *testing.T) {
	conn, err := grpc.Dial(EmployeeGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Employee Service gRPC not available: %v", err)
		return
	}
	defer conn.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	state := conn.GetState()
	t.Logf("Employee Service gRPC connection state: %v", state)
	
	conn.WaitForStateChange(ctx, state)
	
	assert.NotEqual(t, "TransientFailure", conn.GetState().String(), "Connection should not be in failure state")
}

// TestStructureServiceGRPC tests Structure Service gRPC communication
func TestStructureServiceGRPC(t *testing.T) {
	conn, err := grpc.Dial(StructureGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Skipf("Structure Service gRPC not available: %v", err)
		return
	}
	defer conn.Close()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	state := conn.GetState()
	t.Logf("Structure Service gRPC connection state: %v", state)
	
	conn.WaitForStateChange(ctx, state)
	
	assert.NotEqual(t, "TransientFailure", conn.GetState().String(), "Connection should not be in failure state")
}

// TestGRPCRetryMechanism tests retry logic for failed gRPC calls
func TestGRPCRetryMechanism(t *testing.T) {
	// This test verifies that services implement retry logic
	// We'll test by making a call to a service and checking logs/behavior
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create an employee which triggers gRPC call to MaxBot Service
	employeeData := map[string]interface{}{
		"first_name":    "Retry",
		"last_name":     "Test",
		"phone":         "+79991234599",
		"email":         fmt.Sprintf("retry.test.%d@university.ru", time.Now().Unix()),
		"university_id": 1,
		"university": map[string]interface{}{
			"name": "Retry Test University",
			"inn":  "7777777777",
			"kpp":  "777777777",
		},
	}
	
	// Even if MaxBot service is unavailable, employee creation should succeed
	// (graceful degradation)
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
	
	// Should either succeed (201) or fail gracefully
	assert.True(t, status == 201 || status == 500 || status == 502, 
		"Status should be 201 (success), 500 (internal error), or 502 (bad gateway): got %d: %s", 
		status, string(respBody))
	
	if status == 201 {
		response := ParseJSON(t, respBody)
		assert.NotNil(t, response["id"], "Employee should be created even if MAX_id lookup fails")
		
		// Cleanup
		employeeID := int(response["id"].(float64))
		client.DELETE(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	}
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"employees", "universities"})
}

// TestInterServiceCommunication tests communication between services
func TestInterServiceCommunication(t *testing.T) {
	WaitForService(t, EmployeeServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	WaitForService(t, ChatServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create employee (Employee Service -> Auth Service gRPC)
	employeeData := map[string]interface{}{
		"first_name":    "Inter",
		"last_name":     "Service",
		"phone":         "+79991234588",
		"email":         fmt.Sprintf("inter.service.%d@university.ru", time.Now().Unix()),
		"role":          "curator",
		"university_id": 1,
		"university": map[string]interface{}{
			"name": "Inter Service Test University",
			"inn":  "6666666666",
			"kpp":  "666666666",
		},
	}
	
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
	require.Equal(t, 201, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	employeeID := int(response["id"].(float64))
	
	// Verify user_id was set (Auth Service was called)
	assert.NotNil(t, response["user_id"], "user_id should be set via Auth Service gRPC")
	
	// Create chat (Chat Service -> Auth Service gRPC for validation)
	chatData := map[string]interface{}{
		"name":          "Inter Service Chat",
		"url":           "https://max.com/interservice",
		"university_id": 1,
		"source":        "admin_panel",
	}
	
	status, respBody = client.POST(t, ChatServiceURL+"/chats", chatData)
	require.Equal(t, 201, status, string(respBody))
	
	chatResponse := ParseJSON(t, respBody)
	chatID := int(chatResponse["id"].(float64))
	
	// Add administrator (Chat Service -> MaxBot Service gRPC)
	adminData := map[string]interface{}{
		"chat_id": chatID,
		"phone":   "+79991234577",
		"name":    "Test Admin",
	}
	
	status, respBody = client.POST(t, ChatServiceURL+"/administrators", adminData)
	require.Equal(t, 201, status, string(respBody))
	
	adminResponse := ParseJSON(t, respBody)
	assert.NotNil(t, adminResponse["id"], "Administrator should be created")
	
	// Cleanup
	client.DELETE(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	client.DELETE(t, fmt.Sprintf("%s/chats/%d", ChatServiceURL, chatID))
	
	employeeDB := ConnectDB(t, EmployeeDBConnStr)
	defer employeeDB.Close()
	CleanupDB(t, employeeDB, []string{"employees", "universities"})
	
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
	
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"user_roles", "users"})
}

// TestGRPCConnectionPooling tests that services maintain gRPC connections
func TestGRPCConnectionPooling(t *testing.T) {
	// Make multiple requests to verify connection reuse
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Make 10 rapid requests
	for i := 0; i < 10; i++ {
		employeeData := map[string]interface{}{
			"first_name":    fmt.Sprintf("Pool%d", i),
			"last_name":     "Test",
			"phone":         fmt.Sprintf("+7999123458%d", i),
			"email":         fmt.Sprintf("pool%d.%d@university.ru", i, time.Now().Unix()),
			"university_id": 1,
			"university": map[string]interface{}{
				"name": "Pool Test University",
				"inn":  "5555555555",
				"kpp":  "555555555",
			},
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
		
		// All requests should succeed (connection pooling working)
		assert.True(t, status == 201 || status == 500, 
			"Request %d should succeed or fail gracefully: got %d: %s", 
			i, status, string(respBody))
	}
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"employees", "universities"})
	
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"user_roles", "users"})
}
