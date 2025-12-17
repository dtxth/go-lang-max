package integration_tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

// Service endpoints
const (
	AuthServiceURL      = "http://localhost:8080"
	EmployeeServiceURL  = "http://localhost:8081"
	ChatServiceURL      = "http://localhost:8082"
	StructureServiceURL = "http://localhost:8083"
	MigrationServiceURL = "http://localhost:8084"
)

// Database connection strings
const (
	AuthDBConnStr      = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	EmployeeDBConnStr  = "postgres://employee_user:employee_pass@localhost:5433/employee_db?sslmode=disable"
	ChatDBConnStr      = "postgres://chat_user:chat_pass@localhost:5434/chat_db?sslmode=disable"
	StructureDBConnStr = "postgres://postgres:postgres@localhost:5435/postgres?sslmode=disable"
	MigrationDBConnStr = "postgres://postgres:postgres@localhost:5436/migration_db?sslmode=disable"
)

// HTTPClient wraps http.Client with helper methods
type HTTPClient struct {
	client *http.Client
	token  string
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient() *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the JWT token for authenticated requests
func (c *HTTPClient) SetToken(token string) {
	c.token = token
}

// POST sends a POST request
func (c *HTTPClient) POST(t *testing.T, url string, body interface{}) (int, []byte) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		require.NoError(t, err)
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest("POST", url, reqBody)
	require.NoError(t, err)

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, respBody
}

// GET sends a GET request
func (c *HTTPClient) GET(t *testing.T, url string) (int, []byte) {
	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, respBody
}

// DELETE sends a DELETE request
func (c *HTTPClient) DELETE(t *testing.T, url string) (int, []byte) {
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(t, err)

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, respBody
}

// PUT sends a PUT request
func (c *HTTPClient) PUT(t *testing.T, url string, body interface{}) (int, []byte) {
	jsonData, err := json.Marshal(body)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp.StatusCode, respBody
}

// ConnectDB connects to a database
func ConnectDB(t *testing.T, connStr string) *sql.DB {
	db, err := sql.Open("postgres", connStr)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	return db
}

// CleanupDB cleans up test data from database
func CleanupDB(t *testing.T, db *sql.DB, tables []string) {
	for _, table := range tables {
		_, err := db.Exec(fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: failed to cleanup table %s: %v", table, err)
		}
	}
}

// WaitForService waits for a service to be ready
func WaitForService(t *testing.T, url string, maxRetries int) {
	client := &http.Client{Timeout: 2 * time.Second}
	
	for i := 0; i < maxRetries; i++ {
		resp, err := client.Get(url + "/health")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	
	t.Fatalf("Service %s not ready after %d retries", url, maxRetries)
}

// CreateTestUser creates a test user and returns JWT token
func CreateTestUser(t *testing.T, role string, universityID int) string {
	client := NewHTTPClient()
	
	// Register user with unique email
	rand.Seed(time.Now().UnixNano())
	timestamp := time.Now().UnixNano()
	randomNum := rand.Intn(1000000)
	registerBody := map[string]interface{}{
		"email":    fmt.Sprintf("test-%s-%d-%d-%d@example.com", role, universityID, timestamp, randomNum),
		"password": "testpassword123",
		"name":     fmt.Sprintf("Test %s", role),
	}
	
	status, respBody := client.POST(t, AuthServiceURL+"/register", registerBody)
	if status != http.StatusOK && status != http.StatusCreated {
		require.Equal(t, http.StatusCreated, status, string(respBody))
	}
	
	var registerResp map[string]interface{}
	err := json.Unmarshal(respBody, &registerResp)
	require.NoError(t, err)
	
	// Login to get token
	loginBody := map[string]interface{}{
		"email":    registerBody["email"],
		"password": registerBody["password"],
	}
	
	status, respBody = client.POST(t, AuthServiceURL+"/login", loginBody)
	require.Equal(t, http.StatusOK, status, string(respBody))
	
	var loginResp map[string]interface{}
	err = json.Unmarshal(respBody, &loginResp)
	require.NoError(t, err)
	
	token, ok := loginResp["access_token"].(string)
	require.True(t, ok, "access_token not found in response")
	
	return token
}

// ParseJSON parses JSON response
func ParseJSON(t *testing.T, data []byte) map[string]interface{} {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	require.NoError(t, err)
	return result
}

// ParseJSONArray parses JSON array response
func ParseJSONArray(t *testing.T, data []byte) []interface{} {
	var result []interface{}
	err := json.Unmarshal(data, &result)
	require.NoError(t, err)
	return result
}
