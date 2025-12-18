package integration_tests

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExcelImportWithFullStructure tests Excel import creating full hierarchy
func TestExcelImportWithFullStructure(t *testing.T) {
	WaitForService(t, StructureServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create a test Excel file
	excelContent := createTestExcelFileBytes(t)
	
	// Upload Excel file
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	part, err := writer.CreateFormFile("file", "test_structure.xlsx")
	require.NoError(t, err)
	
	_, err = io.Copy(part, bytes.NewReader(excelContent))
	require.NoError(t, err)
	
	err = writer.Close()
	require.NoError(t, err)
	
	req, err := http.NewRequest("POST", StructureServiceURL+"/import/excel", body)
	require.NoError(t, err)
	
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	
	require.Equal(t, 202, resp.StatusCode, "Expected 202 Accepted: %s", string(respBody))
	
	response := ParseJSON(t, respBody)
	assert.NotNil(t, response["job_id"], "job_id should be present")
	
	// Wait for import to complete
	time.Sleep(5 * time.Second)
	
	// Verify structure was created
	structureDB := ConnectDB(t, StructureDBConnStr)
	defer structureDB.Close()
	
	var universityCount int
	err = structureDB.QueryRow("SELECT COUNT(*) FROM universities").Scan(&universityCount)
	require.NoError(t, err)
	assert.Greater(t, universityCount, 0, "Universities should be created")
	
	var groupCount int
	err = structureDB.QueryRow("SELECT COUNT(*) FROM groups").Scan(&groupCount)
	require.NoError(t, err)
	assert.Greater(t, groupCount, 0, "Groups should be created")
	
	// Cleanup
	CleanupDB(t, structureDB, []string{"groups", "faculties", "branches", "universities"})
}

// TestUniversityStructureRetrieval tests getting full hierarchy
func TestUniversityStructureRetrieval(t *testing.T) {
	WaitForService(t, StructureServiceURL, 10)
	WaitForService(t, ChatServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Setup test data
	structureDB := ConnectDB(t, StructureDBConnStr)
	defer structureDB.Close()
	
	// Create university
	var universityID int
	err := structureDB.QueryRow(`
		INSERT INTO universities (name, inn, kpp, created_at)
		VALUES ('Test University', '1234567890', '123456789', NOW())
		RETURNING id
	`).Scan(&universityID)
	require.NoError(t, err)
	
	// Create branch
	var branchID int
	err = structureDB.QueryRow(`
		INSERT INTO branches (university_id, name, created_at)
		VALUES ($1, 'Main Branch', NOW())
		RETURNING id
	`, universityID).Scan(&branchID)
	require.NoError(t, err)
	
	// Create faculty
	var facultyID int
	err = structureDB.QueryRow(`
		INSERT INTO faculties (university_id, branch_id, name, created_at)
		VALUES ($1, $2, 'Computer Science', NOW())
		RETURNING id
	`, universityID, branchID).Scan(&facultyID)
	require.NoError(t, err)
	
	// Create group
	var groupID int
	err = structureDB.QueryRow(`
		INSERT INTO groups (faculty_id, name, course, created_at)
		VALUES ($1, 'CS-101', 1, NOW())
		RETURNING id
	`, facultyID).Scan(&groupID)
	require.NoError(t, err)
	
	// Create associated chat
	chatDB := ConnectDB(t, ChatDBConnStr)
	defer chatDB.Close()
	
	var chatID int
	err = chatDB.QueryRow(`
		INSERT INTO chats (name, url, university_id, branch_id, faculty_id, source, created_at)
		VALUES ('CS-101 Chat', 'https://max.com/cs101', $1, $2, $3, 'academic_group', NOW())
		RETURNING id
	`, universityID, branchID, facultyID).Scan(&chatID)
	require.NoError(t, err)
	
	// Link group to chat
	_, err = structureDB.Exec(`
		UPDATE groups SET chat_id = $1, chat_url = $2, chat_name = $3
		WHERE id = $4
	`, chatID, "https://max.com/cs101", "CS-101 Chat", groupID)
	require.NoError(t, err)
	
	// Retrieve structure
	status, respBody := client.GET(t, fmt.Sprintf("%s/universities/%d/structure", StructureServiceURL, universityID))
	require.Equal(t, 200, status, string(respBody))
	
	structure := ParseJSON(t, respBody)
	
	// Validate structure
	assert.Equal(t, "Test University", structure["name"])
	
	branches, ok := structure["branches"].([]interface{})
	require.True(t, ok, "branches should be an array")
	assert.Equal(t, 1, len(branches), "Should have 1 branch")
	
	branch := branches[0].(map[string]interface{})
	assert.Equal(t, "Main Branch", branch["name"])
	
	faculties, ok := branch["faculties"].([]interface{})
	require.True(t, ok, "faculties should be an array")
	assert.Equal(t, 1, len(faculties), "Should have 1 faculty")
	
	faculty := faculties[0].(map[string]interface{})
	assert.Equal(t, "Computer Science", faculty["name"])
	
	groups, ok := faculty["groups"].([]interface{})
	require.True(t, ok, "groups should be an array")
	assert.Equal(t, 1, len(groups), "Should have 1 group")
	
	group := groups[0].(map[string]interface{})
	assert.Equal(t, "CS-101", group["name"])
	
	// Verify chat information is included
	if chatInfo, ok := group["chat"]; ok && chatInfo != nil {
		chat := chatInfo.(map[string]interface{})
		assert.Equal(t, "CS-101 Chat", chat["name"])
		assert.Equal(t, "https://max.com/cs101", chat["url"])
	}
	
	// Cleanup
	CleanupDB(t, structureDB, []string{"groups", "faculties", "branches", "universities"})
	CleanupDB(t, chatDB, []string{"administrators", "chats"})
}

// TestDepartmentManagerAssignment tests operator assignment to departments
func TestDepartmentManagerAssignment(t *testing.T) {
	WaitForService(t, StructureServiceURL, 10)
	WaitForService(t, EmployeeServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "curator", 10)
	client.SetToken(token)
	
	// Setup test data
	structureDB := ConnectDB(t, StructureDBConnStr)
	defer structureDB.Close()
	
	// Create university and branch
	var universityID int
	err := structureDB.QueryRow(`
		INSERT INTO universities (name, inn, kpp, created_at)
		VALUES ('Manager Test University', '9999999999', '999999999', NOW())
		RETURNING id
	`).Scan(&universityID)
	require.NoError(t, err)
	
	var branchID int
	err = structureDB.QueryRow(`
		INSERT INTO branches (university_id, name, created_at)
		VALUES ($1, 'Test Branch', NOW())
		RETURNING id
	`, universityID).Scan(&branchID)
	require.NoError(t, err)
	
	// Create employee (operator)
	employeeDB := ConnectDB(t, EmployeeDBConnStr)
	defer employeeDB.Close()
	
	var employeeID int
	err = employeeDB.QueryRow(`
		INSERT INTO employees (first_name, last_name, phone, email, role, university_id, created_at)
		VALUES ('Operator', 'Test', '+79991234570', $1, 'operator', $2, NOW())
		RETURNING id
	`, fmt.Sprintf("operator.test.%d@university.ru", time.Now().Unix()), universityID).Scan(&employeeID)
	require.NoError(t, err)
	
	// Assign operator to branch
	assignmentData := map[string]interface{}{
		"employee_id": employeeID,
		"branch_id":   branchID,
	}
	
	status, respBody := client.POST(t, StructureServiceURL+"/departments/managers", assignmentData)
	require.Equal(t, 201, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	assert.NotNil(t, response["id"], "Assignment ID should be present")
	assert.Equal(t, float64(employeeID), response["employee_id"])
	assert.Equal(t, float64(branchID), response["branch_id"])
	
	// Verify assignment was created
	var count int
	err = structureDB.QueryRow(`
		SELECT COUNT(*) FROM department_managers 
		WHERE employee_id = $1 AND branch_id = $2
	`, employeeID, branchID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Assignment should be created")
	
	// Cleanup
	CleanupDB(t, structureDB, []string{"department_managers", "groups", "faculties", "branches", "universities"})
	CleanupDB(t, employeeDB, []string{"employees", "universities"})
}

// TestStructureAlphabeticalOrdering tests that entities are ordered alphabetically
func TestStructureAlphabeticalOrdering(t *testing.T) {
	WaitForService(t, StructureServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Setup test data
	structureDB := ConnectDB(t, StructureDBConnStr)
	defer structureDB.Close()
	
	// Create university
	var universityID int
	err := structureDB.QueryRow(`
		INSERT INTO universities (name, inn, kpp, created_at)
		VALUES ('Ordering Test University', '8888888888', '888888888', NOW())
		RETURNING id
	`).Scan(&universityID)
	require.NoError(t, err)
	
	// Create branches in non-alphabetical order
	_, err = structureDB.Exec(`
		INSERT INTO branches (university_id, name, created_at)
		VALUES 
			($1, 'Zeta Branch', NOW()),
			($1, 'Alpha Branch', NOW()),
			($1, 'Beta Branch', NOW())
	`, universityID)
	require.NoError(t, err)
	
	// Retrieve structure
	status, respBody := client.GET(t, fmt.Sprintf("%s/universities/%d/structure", StructureServiceURL, universityID))
	require.Equal(t, 200, status, string(respBody))
	
	structure := ParseJSON(t, respBody)
	branches, ok := structure["branches"].([]interface{})
	require.True(t, ok)
	
	// Verify alphabetical ordering
	if len(branches) >= 3 {
		branch1 := branches[0].(map[string]interface{})
		branch2 := branches[1].(map[string]interface{})
		branch3 := branches[2].(map[string]interface{})
		
		assert.Equal(t, "Alpha Branch", branch1["name"])
		assert.Equal(t, "Beta Branch", branch2["name"])
		assert.Equal(t, "Zeta Branch", branch3["name"])
	}
	
	// Cleanup
	CleanupDB(t, structureDB, []string{"branches", "universities"})
}

// createTestExcelFileBytes creates a minimal Excel file for testing
// In a real implementation, this would use a library like excelize
func createTestExcelFileBytes(t *testing.T) []byte {
	// For this integration test, we'll create a simple CSV-like content
	// In production, you'd use github.com/xuri/excelize/v2
	content := `phone,inn,foiv,org_name,branch_name,kpp,faculty,course,group_number,chat_name,chat_url
+79991234567,1234567890,Минобрнауки,Test University,Main Campus,123456789,Computer Science,1,CS-101,CS-101 Chat,https://max.com/cs101
+79991234568,1234567890,Минобрнауки,Test University,Main Campus,123456789,Computer Science,2,CS-201,CS-201 Chat,https://max.com/cs201
`
	return []byte(content)
}
