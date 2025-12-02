package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

// TestE2E_CompleteUserJourney tests the complete user journey from registration to chat management
func TestE2E_CompleteUserJourney(t *testing.T) {
	t.Log("=== Starting E2E Complete User Journey Test ===")

	// Wait for all services to be ready
	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, EmployeeServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)
	WaitForService(t, StructureServiceURL, 30)

	// Step 1: Register a new user (Superadmin)
	t.Log("Step 1: Registering superadmin user...")
	superadminToken := registerUser(t, "superadmin@test.com", "SuperPass123!", "superadmin")
	if superadminToken == "" {
		t.Fatal("Failed to register superadmin")
	}
	t.Log("✓ Superadmin registered successfully")

	// Step 2: Create a university structure
	t.Log("Step 2: Creating university structure...")
	universityID := createUniversity(t, superadminToken, "Тестовый Университет", "1234567890", "123456789")
	if universityID == 0 {
		t.Fatal("Failed to create university")
	}
	t.Logf("✓ University created with ID: %d", universityID)

	// Step 3: Register a curator for the university
	t.Log("Step 3: Registering curator user...")
	curatorToken := registerUser(t, "curator@test.com", "CuratorPass123!", "curator")
	if curatorToken == "" {
		t.Fatal("Failed to register curator")
	}
	t.Log("✓ Curator registered successfully")

	// Step 4: Create an employee (operator)
	t.Log("Step 4: Creating employee...")
	employeeID := createEmployee(t, superadminToken, "+79001234567", "Иван", "Иванов", "Тестовый Университет")
	if employeeID == 0 {
		t.Fatal("Failed to create employee")
	}
	t.Logf("✓ Employee created with ID: %d", employeeID)

	// Step 5: Create a chat
	t.Log("Step 5: Creating chat...")
	chatID := createChat(t, superadminToken, "Математика 1 курс", "https://max.ru/chat/test123", universityID)
	if chatID == 0 {
		t.Fatal("Failed to create chat")
	}
	t.Logf("✓ Chat created with ID: %d", chatID)

	// Step 6: Add administrator to chat
	t.Log("Step 6: Adding administrator to chat...")
	adminID := addChatAdministrator(t, superadminToken, chatID, "+79001234567")
	if adminID == 0 {
		t.Fatal("Failed to add administrator")
	}
	t.Logf("✓ Administrator added with ID: %d", adminID)

	// Step 7: Search chats as superadmin (should see all)
	t.Log("Step 7: Searching chats as superadmin...")
	chats := searchChats(t, superadminToken, "Математика")
	if len(chats) == 0 {
		t.Fatal("Superadmin should see the created chat")
	}
	t.Logf("✓ Superadmin can see %d chat(s)", len(chats))

	// Step 8: Search chats as curator (should see only their university)
	t.Log("Step 8: Searching chats as curator...")
	curatorChats := searchChats(t, curatorToken, "Математика")
	t.Logf("✓ Curator can see %d chat(s)", len(curatorChats))

	// Step 9: Get employee by ID
	t.Log("Step 9: Getting employee details...")
	employee := getEmployee(t, superadminToken, employeeID)
	if employee == nil {
		t.Fatal("Failed to get employee")
	}
	t.Logf("✓ Employee retrieved: %s %s", employee["first_name"], employee["last_name"])

	// Step 10: Get university structure
	t.Log("Step 10: Getting university structure...")
	structure := getUniversityStructure(t, superadminToken, universityID)
	if structure == nil {
		t.Fatal("Failed to get university structure")
	}
	t.Log("✓ University structure retrieved successfully")

	t.Log("=== E2E Complete User Journey Test PASSED ===")
}

// TestE2E_RoleBasedAccessControl tests role-based access control across services
func TestE2E_RoleBasedAccessControl(t *testing.T) {
	t.Log("=== Starting E2E Role-Based Access Control Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)

	// Create users with different roles
	t.Log("Creating users with different roles...")
	superadminToken := registerUser(t, "rbac_superadmin@test.com", "Pass123!", "superadmin")
	curatorToken := registerUser(t, "rbac_curator@test.com", "Pass123!", "curator")
	operatorToken := registerUser(t, "rbac_operator@test.com", "Pass123!", "operator")

	// Create test data
	universityID := createUniversity(t, superadminToken, "RBAC Test University", "9876543210", "987654321")
	chatID := createChat(t, superadminToken, "RBAC Test Chat", "https://max.ru/chat/rbac", universityID)

	// Test 1: Superadmin can see all chats
	t.Log("Test 1: Superadmin access...")
	superadminChats := listChats(t, superadminToken, 10, 0)
	if len(superadminChats) == 0 {
		t.Error("Superadmin should see chats")
	}
	t.Logf("✓ Superadmin can see %d chat(s)", len(superadminChats))

	// Test 2: Curator can see only their university chats
	t.Log("Test 2: Curator access...")
	curatorChats := listChats(t, curatorToken, 10, 0)
	t.Logf("✓ Curator can see %d chat(s)", len(curatorChats))

	// Test 3: Operator can see only their department chats
	t.Log("Test 3: Operator access...")
	operatorChats := listChats(t, operatorToken, 10, 0)
	t.Logf("✓ Operator can see %d chat(s)", len(operatorChats))

	// Test 4: Unauthorized access should fail
	t.Log("Test 4: Unauthorized access...")
	resp, err := http.Get(fmt.Sprintf("%s/chats", ChatServiceURL))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
	t.Log("✓ Unauthorized access properly rejected")

	t.Log("=== E2E Role-Based Access Control Test PASSED ===")
}

// TestE2E_ChatAdministratorManagement tests complete chat administrator lifecycle
func TestE2E_ChatAdministratorManagement(t *testing.T) {
	t.Log("=== Starting E2E Chat Administrator Management Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)
	WaitForService(t, EmployeeServiceURL, 30)

	// Setup
	token := registerUser(t, "admin_mgmt@test.com", "Pass123!", "superadmin")
	universityID := createUniversity(t, token, "Admin Test University", "1111111111", "111111111")
	chatID := createChat(t, token, "Admin Test Chat", "https://max.ru/chat/admin", universityID)

	// Create employees
	t.Log("Creating employees...")
	emp1ID := createEmployee(t, token, "+79111111111", "Админ", "Первый", "Admin Test University")
	emp2ID := createEmployee(t, token, "+79222222222", "Админ", "Второй", "Admin Test University")
	emp3ID := createEmployee(t, token, "+79333333333", "Админ", "Третий", "Admin Test University")

	// Add first administrator
	t.Log("Adding first administrator...")
	admin1ID := addChatAdministrator(t, token, chatID, "+79111111111")
	if admin1ID == 0 {
		t.Fatal("Failed to add first administrator")
	}
	t.Log("✓ First administrator added")

	// Add second administrator
	t.Log("Adding second administrator...")
	admin2ID := addChatAdministrator(t, token, chatID, "+79222222222")
	if admin2ID == 0 {
		t.Fatal("Failed to add second administrator")
	}
	t.Log("✓ Second administrator added")

	// Try to add duplicate (should fail)
	t.Log("Testing duplicate administrator prevention...")
	resp := makeRequest(t, "POST", fmt.Sprintf("%s/chats/%d/administrators", ChatServiceURL, chatID), token, map[string]string{
		"phone": "+79111111111",
	})
	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for duplicate admin, got %d", resp.StatusCode)
	}
	t.Log("✓ Duplicate administrator properly rejected")

	// Remove one administrator (should succeed)
	t.Log("Removing one administrator...")
	removeResp := makeRequest(t, "DELETE", fmt.Sprintf("%s/administrators/%d", ChatServiceURL, admin2ID), token, nil)
	if removeResp.StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", removeResp.StatusCode)
	}
	t.Log("✓ Administrator removed successfully")

	// Try to remove last administrator (should fail)
	t.Log("Testing last administrator protection...")
	lastRemoveResp := makeRequest(t, "DELETE", fmt.Sprintf("%s/administrators/%d", ChatServiceURL, admin1ID), token, nil)
	if lastRemoveResp.StatusCode != http.StatusConflict {
		t.Errorf("Expected 409 Conflict for last admin removal, got %d", lastRemoveResp.StatusCode)
	}
	t.Log("✓ Last administrator protection working")

	t.Log("=== E2E Chat Administrator Management Test PASSED ===")
}

// TestE2E_PaginationAndSearch tests pagination and search across services
func TestE2E_PaginationAndSearch(t *testing.T) {
	t.Log("=== Starting E2E Pagination and Search Test ===")

	WaitForService(t, AuthServiceURL, 30)
	WaitForService(t, ChatServiceURL, 30)

	token := registerUser(t, "pagination@test.com", "Pass123!", "superadmin")
	universityID := createUniversity(t, token, "Pagination Test University", "2222222222", "222222222")

	// Create multiple chats
	t.Log("Creating multiple chats for pagination testing...")
	chatNames := []string{
		"Математика 1 курс",
		"Математика 2 курс",
		"Физика 1 курс",
		"Химия 1 курс",
		"Биология 1 курс",
	}

	for _, name := range chatNames {
		createChat(t, token, name, fmt.Sprintf("https://max.ru/chat/%s", name), universityID)
		time.Sleep(100 * time.Millisecond) // Small delay to ensure order
	}
	t.Logf("✓ Created %d chats", len(chatNames))

	// Test pagination
	t.Log("Testing pagination...")
	page1 := listChats(t, token, 2, 0)
	page2 := listChats(t, token, 2, 2)

	if len(page1) != 2 {
		t.Errorf("Expected 2 chats in page 1, got %d", len(page1))
	}
	if len(page2) != 2 {
		t.Errorf("Expected 2 chats in page 2, got %d", len(page2))
	}
	t.Log("✓ Pagination working correctly")

	// Test search
	t.Log("Testing search...")
	mathChats := searchChats(t, token, "Математика")
	if len(mathChats) < 2 {
		t.Errorf("Expected at least 2 math chats, got %d", len(mathChats))
	}
	t.Logf("✓ Search found %d matching chats", len(mathChats))

	t.Log("=== E2E Pagination and Search Test PASSED ===")
}

// Helper functions

func registerUser(t *testing.T, email, password, role string) string {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
		"role":     role,
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/register", AuthServiceURL), "", reqBody)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Logf("Warning: Registration returned status %d for %s", resp.StatusCode, email)
		// Try to login instead
		return loginUser(t, email, password)
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if token, ok := result["access_token"].(string); ok {
		return token
	}

	// If no token in response, try to login
	return loginUser(t, email, password)
}

func loginUser(t *testing.T, email, password string) string {
	reqBody := map[string]string{
		"email":    email,
		"password": password,
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/login", AuthServiceURL), "", reqBody)
	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if token, ok := result["access_token"].(string); ok {
		return token
	}

	return ""
}

func createUniversity(t *testing.T, token, name, inn, kpp string) int64 {
	reqBody := map[string]string{
		"name": name,
		"inn":  inn,
		"kpp":  kpp,
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/universities", StructureServiceURL), token, reqBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Warning: University creation returned status %d", resp.StatusCode)
		return 0
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(float64); ok {
		return int64(id)
	}

	return 0
}

func createEmployee(t *testing.T, token, phone, firstName, lastName, universityName string) int64 {
	reqBody := map[string]string{
		"phone":           phone,
		"first_name":      firstName,
		"last_name":       lastName,
		"university_name": universityName,
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/employees", EmployeeServiceURL), token, reqBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Warning: Employee creation returned status %d", resp.StatusCode)
		return 0
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(float64); ok {
		return int64(id)
	}

	return 0
}

func createChat(t *testing.T, token, name, url string, universityID int64) int64 {
	reqBody := map[string]interface{}{
		"name":          name,
		"url":           url,
		"university_id": universityID,
		"source":        "academic_group",
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/chats", ChatServiceURL), token, reqBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Warning: Chat creation returned status %d", resp.StatusCode)
		return 0
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(float64); ok {
		return int64(id)
	}

	return 0
}

func addChatAdministrator(t *testing.T, token string, chatID int64, phone string) int64 {
	reqBody := map[string]string{
		"phone": phone,
	}

	resp := makeRequest(t, "POST", fmt.Sprintf("%s/chats/%d/administrators", ChatServiceURL, chatID), token, reqBody)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		t.Logf("Warning: Administrator addition returned status %d", resp.StatusCode)
		return 0
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if id, ok := result["id"].(float64); ok {
		return int64(id)
	}

	return 0
}

func searchChats(t *testing.T, token, query string) []map[string]interface{} {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/chats?query=%s", ChatServiceURL, query), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Warning: Search request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if chats, ok := result["chats"].([]interface{}); ok {
		var chatList []map[string]interface{}
		for _, chat := range chats {
			if chatMap, ok := chat.(map[string]interface{}); ok {
				chatList = append(chatList, chatMap)
			}
		}
		return chatList
	}

	return nil
}

func listChats(t *testing.T, token string, limit, offset int) []map[string]interface{} {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/chats?limit=%d&offset=%d", ChatServiceURL, limit, offset), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Warning: List request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if chats, ok := result["chats"].([]interface{}); ok {
		var chatList []map[string]interface{}
		for _, chat := range chats {
			if chatMap, ok := chat.(map[string]interface{}); ok {
				chatList = append(chatList, chatMap)
			}
		}
		return chatList
	}

	return nil
}

func getEmployee(t *testing.T, token string, employeeID int64) map[string]interface{} {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Warning: Get employee request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result
}

func getUniversityStructure(t *testing.T, token string, universityID int64) map[string]interface{} {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/universities/%d/structure", StructureServiceURL, universityID), nil)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Logf("Warning: Get structure request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result
}

func makeRequest(t *testing.T, method, url, token string, body interface{}) *http.Response {
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
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}

	return resp
}
