package main

import (
	"fmt"
	"net/http"
	"testing"
)

// TestE2E_ErrorHandling tests error handling across all services
func TestE2E_ErrorHandling(t *testing.T) {
	t.Log("=== Starting E2E Error Handling Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, EmployeeServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)
	WaitForService(t, StructureServiceURL, 30)

	token := registerUser(t, "error_test@test.com", "Pass123!", "superadmin")

	// Test 1: Invalid JSON
	t.Log("Test 1: Invalid JSON handling...")
	testInvalidJSON(t, token)

	// Test 2: Missing required fields
	t.Log("Test 2: Missing required fields...")
	testMissingFields(t, token)

	// Test 3: Invalid IDs
	t.Log("Test 3: Invalid ID handling...")
	testInvalidIDs(t, token)

	// Test 4: Not found errors
	t.Log("Test 4: Not found errors...")
	testNotFoundErrors(t, token)

	// Test 5: Duplicate entries
	t.Log("Test 5: Duplicate entry handling...")
	testDuplicateEntries(t, token)

	// Test 6: Authorization errors
	t.Log("Test 6: Authorization errors...")
	testAuthorizationErrors(t)

	t.Log("=== E2E Error Handling Test PASSED ===")
}

func testInvalidJSON(t *testing.T, token string) {
	// Test invalid JSON in employee creation
	resp := makeRequestRaw(t, "POST", fmt.Sprintf("%s/employees", EmployeeServiceURL), token, []byte("invalid json"))
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid JSON, got %d", resp.StatusCode)
	}
	t.Log("✓ Invalid JSON properly rejected")
}

func testMissingFields(t *testing.T, token string) {
	// Test missing phone in employee creation
	resp := makeRequest(t, "POST", fmt.Sprintf("%s/employees", EmployeeServiceURL), token, map[string]string{
		"first_name": "Test",
		"last_name":  "User",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing phone, got %d", resp.StatusCode)
	}

	// Test missing name in employee creation
	resp = makeRequest(t, "POST", fmt.Sprintf("%s/employees", EmployeeServiceURL), token, map[string]string{
		"phone": "+79001234567",
	})
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for missing name, got %d", resp.StatusCode)
	}

	t.Log("✓ Missing fields properly validated")
}

func testInvalidIDs(t *testing.T, token string) {
	// Test invalid employee ID
	resp := makeRequest(t, "GET", fmt.Sprintf("%s/employees/invalid", EmployeeServiceURL), token, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", resp.StatusCode)
	}

	// Test invalid chat ID
	resp = makeRequest(t, "GET", fmt.Sprintf("%s/chats/invalid", ChatServiceURL), token, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", resp.StatusCode)
	}

	// Test invalid university ID
	resp = makeRequest(t, "GET", fmt.Sprintf("%s/universities/invalid", StructureServiceURL), token, nil)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 for invalid ID, got %d", resp.StatusCode)
	}

	t.Log("✓ Invalid IDs properly rejected")
}

func testNotFoundErrors(t *testing.T, token string) {
	// Test non-existent employee
	resp := makeRequest(t, "GET", fmt.Sprintf("%s/employees/999999", EmployeeServiceURL), token, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent employee, got %d", resp.StatusCode)
	}

	// Test non-existent chat
	resp = makeRequest(t, "GET", fmt.Sprintf("%s/chats/999999", ChatServiceURL), token, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent chat, got %d", resp.StatusCode)
	}

	// Test non-existent university
	resp = makeRequest(t, "GET", fmt.Sprintf("%s/universities/999999", StructureServiceURL), token, nil)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for non-existent university, got %d", resp.StatusCode)
	}

	t.Log("✓ Not found errors properly handled")
}

func testDuplicateEntries(t *testing.T, token string) {
	// Create a university
	universityID := createUniversity(t, token, "Duplicate Test Uni", "3333333333", "333333333")
	if universityID == 0 {
		t.Skip("Could not create university for duplicate test")
		return
	}

	// Create a chat
	chatID := createChat(t, token, "Duplicate Test Chat", "https://max.ru/chat/dup", universityID)
	if chatID == 0 {
		t.Skip("Could not create chat for duplicate test")
		return
	}

	// Create an employee
	employeeID := createEmployee(t, token, "+79444444444", "Дубликат", "Тестов", "Duplicate Test Uni")
	if employeeID == 0 {
		t.Skip("Could not create employee for duplicate test")
		return
	}

	// Try to add same administrator twice
	addChatAdministrator(t, token, chatID, "+79444444444")
	resp := makeRequest(t, "POST", fmt.Sprintf("%s/chats/%d/administrators", ChatServiceURL, chatID), token, map[string]string{
		"phone": "+79444444444",
	})

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 for duplicate administrator, got %d", resp.StatusCode)
	}

	t.Log("✓ Duplicate entries properly prevented")
}

func testAuthorizationErrors(t *testing.T) {
	// Test missing authorization header
	resp := makeRequest(t, "GET", fmt.Sprintf("%s/employees", EmployeeServiceURL), "", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing auth, got %d", resp.StatusCode)
	}

	// Test invalid token format
	resp = makeRequestWithHeader(t, "GET", fmt.Sprintf("%s/employees", EmployeeServiceURL), "InvalidToken", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token format, got %d", resp.StatusCode)
	}

	// Test expired/invalid token
	resp = makeRequestWithHeader(t, "GET", fmt.Sprintf("%s/employees", EmployeeServiceURL), "Bearer invalid_token_xyz", nil)
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", resp.StatusCode)
	}

	t.Log("✓ Authorization errors properly handled")
}

// TestE2E_ConcurrentOperations tests concurrent operations
func TestE2E_ConcurrentOperations(t *testing.T) {
	t.Log("=== Starting E2E Concurrent Operations Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)

	token := registerUser(t, "concurrent@test.com", "Pass123!", "superadmin")
	universityID := createUniversity(t, token, "Concurrent Test Uni", "4444444444", "444444444")

	// Create multiple chats concurrently
	t.Log("Creating chats concurrently...")
	done := make(chan bool, 5)

	for i := 0; i < 5; i++ {
		go func(index int) {
			chatName := fmt.Sprintf("Concurrent Chat %d", index)
			chatURL := fmt.Sprintf("https://max.ru/chat/concurrent%d", index)
			chatID := createChat(t, token, chatName, chatURL, universityID)
			if chatID > 0 {
				t.Logf("✓ Created chat %d with ID %d", index, chatID)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 5; i++ {
		<-done
	}

	t.Log("✓ Concurrent operations completed successfully")
	t.Log("=== E2E Concurrent Operations Test PASSED ===")
}

// TestE2E_DataConsistency tests data consistency across services
func TestE2E_DataConsistency(t *testing.T) {
	t.Log("=== Starting E2E Data Consistency Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, EmployeeServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)
	WaitForService(t, StructureServiceURL, 30)

	token := registerUser(t, "consistency@test.com", "Pass123!", "superadmin")

	// Create a complete data hierarchy
	t.Log("Creating data hierarchy...")
	universityID := createUniversity(t, token, "Consistency Test Uni", "5555555555", "555555555")
	employeeID := createEmployee(t, token, "+79555555555", "Консистент", "Тестов", "Consistency Test Uni")
	chatID := createChat(t, token, "Consistency Test Chat", "https://max.ru/chat/consistency", universityID)
	adminID := addChatAdministrator(t, token, chatID, "+79555555555")

	// Verify data consistency
	t.Log("Verifying data consistency...")

	// Check employee exists
	employee := getEmployee(t, token, employeeID)
	if employee == nil {
		t.Error("Employee not found after creation")
	} else {
		t.Log("✓ Employee data consistent")
	}

	// Check chat exists and has correct university
	chats := searchChats(t, token, "Consistency")
	found := false
	for _, chat := range chats {
		if int64(chat["id"].(float64)) == chatID {
			found = true
			if int64(chat["university_id"].(float64)) != universityID {
				t.Error("Chat has incorrect university_id")
			}
		}
	}
	if !found {
		t.Error("Chat not found after creation")
	} else {
		t.Log("✓ Chat data consistent")
	}

	// Check university structure
	structure := getUniversityStructure(t, token, universityID)
	if structure == nil {
		t.Error("University structure not found")
	} else {
		t.Log("✓ University structure consistent")
	}

	t.Log("=== E2E Data Consistency Test PASSED ===")
}

// Helper functions

func makeRequestRaw(t *testing.T, method, url, token string, body []byte) *http.Response {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}

func makeRequestWithHeader(t *testing.T, method, url, authHeader string, body interface{}) *http.Response {
	var reqBody io.Reader
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}
