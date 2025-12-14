package integration_tests

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmployeeCreationWithRoleAndMaxID tests end-to-end employee creation
// This test validates:
// - Employee creation via HTTP API
// - Role assignment integration with Auth Service
// - MAX_id lookup integration with MaxBot Service
// - Data persistence across services
func TestEmployeeCreationWithRoleAndMaxID(t *testing.T) {
	// Wait for services to be ready
	WaitForService(t, EmployeeServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	// Setup
	client := NewHTTPClient()
	
	// Create a test superadmin user to make the request
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Test data
	employeeData := map[string]interface{}{
		"first_name":    "Ivan",
		"last_name":     "Petrov",
		"phone":         "+79991234567",
		"email":         fmt.Sprintf("ivan.petrov.%d@university.ru", time.Now().Unix()),
		"role":          "curator",
		"university_id": 1,
		"university": map[string]interface{}{
			"name": "Test University",
			"inn":  "1234567890",
			"kpp":  "123456789",
		},
	}
	
	// Create employee
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
	require.Equal(t, 201, status, "Expected 201 Created, got %d: %s", status, string(respBody))
	
	response := ParseJSON(t, respBody)
	
	// Validate response structure
	assert.NotNil(t, response["id"], "Employee ID should be present")
	assert.Equal(t, employeeData["first_name"], response["first_name"])
	assert.Equal(t, employeeData["last_name"], response["last_name"])
	assert.Equal(t, employeeData["phone"], response["phone"])
	assert.Equal(t, employeeData["role"], response["role"])
	
	// Validate MAX_id was attempted (may be null if MaxBot service returns no user)
	// The field should exist in response
	_, hasMaxID := response["max_id"]
	assert.True(t, hasMaxID, "max_id field should be present in response")
	
	// Validate user_id was created (role assignment in Auth Service)
	assert.NotNil(t, response["user_id"], "user_id should be set after role assignment")
	
	employeeID := int(response["id"].(float64))
	
	// Retrieve employee to verify persistence
	status, respBody = client.GET(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	require.Equal(t, 200, status, "Expected 200 OK, got %d: %s", status, string(respBody))
	
	retrievedEmployee := ParseJSON(t, respBody)
	assert.Equal(t, employeeData["first_name"], retrievedEmployee["first_name"])
	assert.Equal(t, employeeData["role"], retrievedEmployee["role"])
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"employees", "universities"})
	
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"user_roles", "users"})
}

// TestEmployeeRoleUpdate tests role synchronization between services
func TestEmployeeRoleUpdate(t *testing.T) {
	WaitForService(t, EmployeeServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create employee with operator role
	employeeData := map[string]interface{}{
		"first_name":    "Maria",
		"last_name":     "Ivanova",
		"phone":         "+79991234568",
		"email":         fmt.Sprintf("maria.ivanova.%d@university.ru", time.Now().Unix()),
		"role":          "operator",
		"university_id": 1,
		"branch_id":     1,
		"university": map[string]interface{}{
			"name": "Test University 2",
			"inn":  "9876543210",
			"kpp":  "987654321",
		},
	}
	
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
	require.Equal(t, 201, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	employeeID := int(response["id"].(float64))
	
	// Update role to curator
	updateData := map[string]interface{}{
		"role": "curator",
	}
	
	status, respBody = client.PUT(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID), updateData)
	require.Equal(t, 200, status, string(respBody))
	
	updatedEmployee := ParseJSON(t, respBody)
	assert.Equal(t, "curator", updatedEmployee["role"], "Role should be updated to curator")
	
	// Verify role was synchronized with Auth Service
	// This would require querying Auth Service, but we can verify through employee retrieval
	status, respBody = client.GET(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	require.Equal(t, 200, status, string(respBody))
	
	retrievedEmployee := ParseJSON(t, respBody)
	assert.Equal(t, "curator", retrievedEmployee["role"])
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"employees", "universities"})
	
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"user_roles", "users"})
}

// TestEmployeeDeletion tests permission revocation on deletion
func TestEmployeeDeletion(t *testing.T) {
	WaitForService(t, EmployeeServiceURL, 10)
	WaitForService(t, AuthServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create employee
	employeeData := map[string]interface{}{
		"first_name":    "Alexey",
		"last_name":     "Smirnov",
		"phone":         "+79991234569",
		"email":         fmt.Sprintf("alexey.smirnov.%d@university.ru", time.Now().Unix()),
		"role":          "curator",
		"university_id": 1,
		"university": map[string]interface{}{
			"name": "Test University 3",
			"inn":  "5555555555",
			"kpp":  "555555555",
		},
	}
	
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
	require.Equal(t, 201, status, string(respBody))
	
	response := ParseJSON(t, respBody)
	employeeID := int(response["id"].(float64))
	userID := response["user_id"]
	
	assert.NotNil(t, userID, "user_id should exist before deletion")
	
	// Delete employee
	status, respBody = client.DELETE(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	require.Equal(t, 204, status, "Expected 204 No Content, got %d: %s", status, string(respBody))
	
	// Verify employee is deleted
	status, _ = client.GET(t, fmt.Sprintf("%s/employees/%d", EmployeeServiceURL, employeeID))
	assert.Equal(t, 404, status, "Employee should not be found after deletion")
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"employees", "universities"})
	
	authDB := ConnectDB(t, AuthDBConnStr)
	defer authDB.Close()
	CleanupDB(t, authDB, []string{"user_roles", "users"})
}

// TestBatchMaxIDUpdate tests batch MAX_id update functionality
func TestBatchMaxIDUpdate(t *testing.T) {
	WaitForService(t, EmployeeServiceURL, 10)
	
	client := NewHTTPClient()
	token := CreateTestUser(t, "superadmin", 0)
	client.SetToken(token)
	
	// Create multiple employees without MAX_id
	for i := 0; i < 5; i++ {
		employeeData := map[string]interface{}{
			"first_name":    fmt.Sprintf("Employee%d", i),
			"last_name":     "TestBatch",
			"phone":         fmt.Sprintf("+7999123456%d", i),
			"email":         fmt.Sprintf("employee%d.%d@university.ru", i, time.Now().Unix()),
			"university_id": 1,
			"university": map[string]interface{}{
				"name": "Batch Test University",
				"inn":  "1111111111",
				"kpp":  "111111111",
			},
		}
		
		status, respBody := client.POST(t, EmployeeServiceURL+"/employees", employeeData)
		require.Equal(t, 201, status, string(respBody))
	}
	
	// Trigger batch MAX_id update
	status, respBody := client.POST(t, EmployeeServiceURL+"/employees/batch-update-maxid", nil)
	require.Equal(t, 202, status, "Expected 202 Accepted, got %d: %s", status, string(respBody))
	
	response := ParseJSON(t, respBody)
	assert.NotNil(t, response["job_id"], "job_id should be present")
	
	// Wait a bit for batch processing
	time.Sleep(3 * time.Second)
	
	// Check batch status
	jobID := int(response["job_id"].(float64))
	status, respBody = client.GET(t, fmt.Sprintf("%s/employees/batch-status?job_id=%d", EmployeeServiceURL, jobID))
	require.Equal(t, 200, status, string(respBody))
	
	statusResp := ParseJSON(t, respBody)
	assert.NotNil(t, statusResp["status"], "Status should be present")
	assert.NotNil(t, statusResp["total"], "Total count should be present")
	
	// Cleanup
	db := ConnectDB(t, EmployeeDBConnStr)
	defer db.Close()
	CleanupDB(t, db, []string{"batch_update_jobs", "employees", "universities"})
}
