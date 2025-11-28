package integration_tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestChatFilteringBySuperadmin tests that superadmin sees all chats
func TestChatFilteringBySuperadmin(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	
	// Create superadmin user
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create chats in different universities
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	// Insert test chats
	_, err := chatDB.Exec(`
		INSERT INTO chats (name, url, university_id, source, created_at)
		VALUES 
			('Chat University 1', 'https://max.com/chat1', 1, 'admin_panel', NOW()),
			('Chat University 2', 'https://max.com/chat2', 2, 'admin_panel', NOW()),
			('Chat University 3', 'https://max.com/chat3', 3, 'admin_panel', NOW())
	`)
	require.NoError(t, err)
	
	// Request chat list as superadmin
	status, respBody := client.GET(t, ChatServiceURL+"/chats")
	require.Equal(t, 200, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	chats, ok := response["chats"].([]interface{})
	require.True(t, ok, "chats should be an array")
	
	// Superadmin should see all chats
	assert.GreaterOrEqual(t, len(chats), 3, "Superadmin should see at least 3 chats")
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestChatFilteringByCurator tests that curator sees only their university chats
func TestChatFilteringByCurator(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	// Setup test data
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	// Insert test chats for different universities
	_, err := chatDB.Exec(`
		INSERT INTO chats (name, url, university_id, source, created_at)
		VALUES 
			('Curator Chat 1', 'https://max.com/curator1', 10, 'admin_panel', NOW()),
			('Curator Chat 2', 'https://max.com/curator2', 10, 'admin_panel', NOW()),
			('Other University Chat', 'https://max.com/other', 20, 'admin_panel', NOW())
	`)
	require.NoError(t, err)
	
	// Create curator user for university 10
	client := NewHTTPClient()
	token := CreateTestUser(t, "curator", 10)
	client.SetToken(token)
	
	// Request chat list as curator
	status, respBody := client.GET(t, ChatServiceURL+"/chats")
	require.Equal(t, 200, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	chats, ok := response["chats"].([]interface{})
	require.True(t, ok, "chats should be an array")
	
	// Curator should only see chats from their university
	for _, chat := range chats {
		chatMap := chat.(map[string]interface{})
		universityID := int(chatMap["university_id"].(float64))
		assert.Equal(t, 10, universityID, "Curator should only see chats from university 10")
	}
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestChatFilteringByOperator tests that operator sees only their department chats
func TestChatFilteringByOperator(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	// Setup test data
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	// Insert test chats for different branches
	_, err := chatDB.Exec(`
		INSERT INTO chats (name, url, university_id, branch_id, source, created_at)
		VALUES 
			('Operator Branch Chat 1', 'https://max.com/op1', 10, 5, 'admin_panel', NOW()),
			('Operator Branch Chat 2', 'https://max.com/op2', 10, 5, 'admin_panel', NOW()),
			('Other Branch Chat', 'https://max.com/other', 10, 6, 'admin_panel', NOW())
	`)
	require.NoError(t, err)
	
	// Create operator user for branch 5
	client := NewHTTPClient()
	token := CreateTestUser(t, "operator", 10)
	client.SetToken(token)
	
	// Request chat list as operator
	status, respBody := client.GET(t, ChatServiceURL+"/chats")
	require.Equal(t, 200, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	chats, ok := response["chats"].([]interface{})
	require.True(t, ok, "chats should be an array")
	
	// Operator should only see chats from their branch
	for _, chat := range chats {
		chatMap := chat.(map[string]interface{})
		if branchID, exists := chatMap["branch_id"]; exists && branchID != nil {
			assert.Equal(t, float64(5), branchID.(float64), "Operator should only see chats from branch 5")
		}
	}
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestChatAdministratorManagement tests adding and removing administrators
func TestChatAdministratorManagement(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "curator", 10)
	client.SetToken(token)
	
	// Create a chat
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	var chatID int
	err := chatDB.QueryRow(`
		INSERT INTO chats (name, url, university_id, source, created_at)
		VALUES ('Admin Test Chat', 'https://max.com/admintest', 10, 'admin_panel', NOW())
		RETURNING id
	`).Scan(&chatID)
	require.NoError(t, err)
	
	// Add first administrator
	adminData := map[string]interface{}{
		"chat_id": chatID,
		"phone":   "+79991111111",
		"name":    "Admin One",
	}
	
	status, respBody := client.POST(t, ChatServiceURL+"/administrators", adminData)
	require.Equal(t, 201, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	adminID1 := int(response["id"].(float64))
	
	// Add second administrator
	adminData2 := map[string]interface{}{
		"chat_id": chatID,
		"phone":   "+79992222222",
		"name":    "Admin Two",
	}
	
	status, respBody = client.POST(t, ChatServiceURL+"/administrators", adminData2)
	require.Equal(t, 201, status, string(respBody))
	
	response = ParseJSON(t, respBody)
	adminID2 := int(response["id"].(float64))
	
	// Try to remove first administrator (should succeed - second admin exists)
	status, _ = client.DELETE(t, fmt.Sprintf("%s/administrators/%d", ChatServiceURL, adminID1))
	assert.Equal(t, 204, status, "Should be able to remove admin when another exists")
	
	// Try to remove last administrator (should fail)
	status, respBody = client.DELETE(t, fmt.Sprintf("%s/administrators/%d", ChatServiceURL, adminID2))
	assert.Equal(t, 400, status, "Should not be able to remove last administrator: %s", string(respBody))
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestChatPagination tests pagination functionality
func TestChatPagination(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create multiple chats
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	for i := 0; i < 25; i++ {
		_, err := chatDB.Exec(`
			INSERT INTO chats (name, url, university_id, source, created_at)
			VALUES ($1, $2, 1, 'admin_panel', NOW())
		`, fmt.Sprintf("Pagination Chat %d", i), fmt.Sprintf("https://max.com/page%d", i))
		require.NoError(t, err)
	}
	
	// Test first page with limit 10
	status, respBody := client.GET(t, ChatServiceURL+"/chats?limit=10&offset=0")
	require.Equal(t, 200, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	chats, ok := response["chats"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 10, len(chats), "First page should have 10 chats")
	
	// Test second page
	status, respBody = client.GET(t, ChatServiceURL+"/chats?limit=10&offset=10")
	require.Equal(t, 200, status, string(respBody))
	
	response = ParseJSON(t, respBody)
	chats, ok = response["chats"].([]interface{})
	require.True(t, ok)
	assert.Equal(t, 10, len(chats), "Second page should have 10 chats")
	
	// Test limit capping at 100
	status, respBody = client.GET(t, ChatServiceURL+"/chats?limit=200")
	require.Equal(t, 200, status, string(respBody))
	
	response = ParseJSON(t, respBody)
	chats, ok = response["chats"].([]interface{})
	require.True(t, ok)
	assert.LessOrEqual(t, len(chats), 100, "Limit should be capped at 100")
	
	// Verify total count in metadata
	if metadata, ok := response["metadata"].(map[string]interface{}); ok {
		total := int(metadata["total"].(float64))
		assert.GreaterOrEqual(t, total, 25, "Total count should include all chats")
	}
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestChatSearch tests search functionality
func TestChatSearch(t *testing.T) {
	WaitForService(t, ChatServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create chats with searchable names
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	_, err := chatDB.Exec(`
		INSERT INTO chats (name, url, university_id, source, created_at)
		VALUES 
			('Математика 1 курс', 'https://max.com/math1', 1, 'academic_group', NOW()),
			('Физика 2 курс', 'https://max.com/physics2', 1, 'academic_group', NOW()),
			('Математика 2 курс', 'https://max.com/math2', 1, 'academic_group', NOW())
	`)
	require.NoError(t, err)
	
	// Search for "Математика"
	status, respBody := client.GET(t, ChatServiceURL+"/chats/search?q=Математика")
	require.Equal(t, 200, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	chats, ok := response["chats"].([]interface{})
	require.True(t, ok)
	
	// Should find 2 chats with "Математика"
	assert.GreaterOrEqual(t, len(chats), 2, "Should find at least 2 chats with Математика")
	
	for _, chat := range chats {
		chatMap := chat.(map[string]interface{})
		name := chatMap["name"].(string)
		assert.Contains(t, name, "Математика", "Search results should contain search term")
	}
	
	// Cleanup
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}
