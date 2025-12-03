package integration_tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

// TestE2E_ExcelImport_FullFlow tests the complete Excel import flow
func TestE2E_ExcelImport_FullFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Step 1: Create test Excel file
	t.Log("Step 1: Creating test Excel file...")
	excelFile := createTestExcelFile(t)
	defer os.Remove(excelFile)

	// Step 2: Upload Excel file to migration-service
	t.Log("Step 2: Uploading Excel file to migration-service...")
	jobID := uploadExcelFile(t, excelFile)
	assert.Greater(t, jobID, 0, "Job ID should be positive")

	// Step 3: Wait for job completion
	t.Log("Step 3: Waiting for job completion...")
	job := waitForJobCompletion(t, jobID, 60*time.Second)
	
	t.Logf("Job completed: total=%d, processed=%d, failed=%d", 
		job.Total, job.Processed, job.Failed)

	// Step 4: Verify job status
	t.Log("Step 4: Verifying job status...")
	assert.Equal(t, "completed", job.Status, "Job should be completed")
	assert.Greater(t, job.Total, 0, "Total should be greater than 0")
	assert.Greater(t, job.Processed, 0, "Processed should be greater than 0")

	// Step 5: Verify data in structure-db
	t.Log("Step 5: Verifying data in structure-db...")
	verifyStructureData(t)

	// Step 6: Verify data in chat-db
	t.Log("Step 6: Verifying data in chat-db...")
	verifyChatData(t)

	t.Log("✅ E2E Excel import test completed successfully!")
}

// createTestExcelFile creates a minimal test Excel file with 18 columns
func createTestExcelFile(t *testing.T) string {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Header row (18 columns)
	headers := []string{
		"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
		"INN", "KPP", "Faculty", "Course", "Group", "ChatName",
		"Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
	}

	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, header)
	}

	// Data rows (2 test rows)
	dataRows := [][]interface{}{
		{
			"79884753064", "496728250", "105014177", "Минобрнауки России",
			"МГТУ Тест", "Головной филиал", "105014177", "10501001",
			"Политехнический колледж МГТУ", "2", "Колледж ИП-22",
			"Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
			"-69257108032233", "https://max.ru/join/test1", "ИСТИНА", "ИСТИНА",
		},
		{
			"79001234567", "123456789", "105014177", "Минобрнауки России",
			"МГТУ Тест", "Головной филиал", "105014177", "10501001",
			"Факультет информатики", "3", "ИВТ-31",
			"Группа ИВТ-31", "79001234567", "file.xlsx",
			"-12345678", "https://max.ru/join/test2", "TRUE", "TRUE",
		},
	}

	for rowIdx, dataRow := range dataRows {
		for colIdx, value := range dataRow {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test_import.xlsx")
	err := f.SaveAs(filePath)
	require.NoError(t, err, "Failed to save Excel file")

	t.Logf("Created test Excel file: %s", filePath)
	return filePath
}

// uploadExcelFile uploads Excel file to migration-service
func uploadExcelFile(t *testing.T, filePath string) int {
	// Read file
	fileContent, err := os.ReadFile(filePath)
	require.NoError(t, err, "Failed to read Excel file")

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	require.NoError(t, err, "Failed to create form file")

	_, err = part.Write(fileContent)
	require.NoError(t, err, "Failed to write file content")

	err = writer.Close()
	require.NoError(t, err, "Failed to close writer")

	// Upload file
	migrationServiceURL := getEnv("MIGRATION_SERVICE_URL", "http://localhost:8084")
	url := fmt.Sprintf("%s/migration/excel", migrationServiceURL)

	req, err := http.NewRequest("POST", url, body)
	require.NoError(t, err, "Failed to create request")
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	require.NoError(t, err, "Failed to upload file")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Upload should return 200 OK")

	// Parse response
	var result struct {
		Message string `json:"message"`
		JobID   int    `json:"job_id"`
	}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "Failed to decode response")

	t.Logf("File uploaded successfully, job_id: %d", result.JobID)
	return result.JobID
}

// MigrationJob represents a migration job
type MigrationJob struct {
	ID               int       `json:"id"`
	SourceType       string    `json:"source_type"`
	SourceIdentifier string    `json:"source_identifier"`
	Status           string    `json:"status"`
	Total            int       `json:"total"`
	Processed        int       `json:"processed"`
	Failed           int       `json:"failed"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      *time.Time `json:"completed_at"`
}

// waitForJobCompletion waits for job to complete
func waitForJobCompletion(t *testing.T, jobID int, timeout time.Duration) *MigrationJob {
	migrationServiceURL := getEnv("MIGRATION_SERVICE_URL", "http://localhost:8084")
	url := fmt.Sprintf("%s/migration/jobs/%d", migrationServiceURL, jobID)

	client := &http.Client{Timeout: 5 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err != nil {
			t.Logf("Failed to get job status: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		var job MigrationJob
		err = json.NewDecoder(resp.Body).Decode(&job)
		resp.Body.Close()

		if err != nil {
			t.Logf("Failed to decode job: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		t.Logf("Job status: %s, total=%d, processed=%d, failed=%d", 
			job.Status, job.Total, job.Processed, job.Failed)

		if job.Status == "completed" || job.Status == "failed" {
			return &job
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatal("Job did not complete within timeout")
	return nil
}

// verifyStructureData verifies data in structure-db
func verifyStructureData(t *testing.T) {
	// Check universities
	count := queryCount(t, "structure-db", "postgres", "postgres", "SELECT COUNT(*) FROM universities")
	assert.Greater(t, count, 0, "Should have at least 1 university")
	t.Logf("Universities count: %d", count)

	// Check branches
	count = queryCount(t, "structure-db", "postgres", "postgres", "SELECT COUNT(*) FROM branches")
	assert.Greater(t, count, 0, "Should have at least 1 branch")
	t.Logf("Branches count: %d", count)

	// Check faculties
	count = queryCount(t, "structure-db", "postgres", "postgres", "SELECT COUNT(*) FROM faculties")
	assert.Greater(t, count, 0, "Should have at least 1 faculty")
	t.Logf("Faculties count: %d", count)

	// Check groups
	count = queryCount(t, "structure-db", "postgres", "postgres", "SELECT COUNT(*) FROM groups")
	assert.Greater(t, count, 0, "Should have at least 1 group")
	t.Logf("Groups count: %d", count)
}

// verifyChatData verifies data in chat-db
func verifyChatData(t *testing.T) {
	// Check chats
	count := queryCount(t, "chat-db", "chat_user", "chat_db", "SELECT COUNT(*) FROM chats")
	assert.Greater(t, count, 0, "Should have at least 1 chat")
	t.Logf("Chats count: %d", count)

	// Check administrators
	count = queryCount(t, "chat-db", "chat_user", "chat_db", "SELECT COUNT(*) FROM administrators")
	assert.Greater(t, count, 0, "Should have at least 1 administrator")
	t.Logf("Administrators count: %d", count)

	// Verify external_chat_id is set
	count = queryCount(t, "chat-db", "chat_user", "chat_db", 
		"SELECT COUNT(*) FROM chats WHERE external_chat_id IS NOT NULL")
	assert.Greater(t, count, 0, "Should have chats with external_chat_id")
	t.Logf("Chats with external_chat_id: %d", count)

	// Verify add_user and add_admin flags
	count = queryCount(t, "chat-db", "chat_user", "chat_db",
		"SELECT COUNT(*) FROM administrators WHERE add_user = TRUE AND add_admin = TRUE")
	assert.Greater(t, count, 0, "Should have administrators with flags set")
	t.Logf("Administrators with flags: %d", count)
}

// queryCount executes a COUNT query and returns the result
func queryCount(t *testing.T, container, user, database, query string) int {
	cmd := fmt.Sprintf("docker-compose exec -T %s psql -U %s -d %s -t -c \"%s\"",
		container, user, database, query)

	output, err := execCommand(cmd)
	if err != nil {
		t.Logf("Query failed: %v, output: %s", err, output)
		return 0
	}

	var count int
	_, err = fmt.Sscanf(output, "%d", &count)
	if err != nil {
		t.Logf("Failed to parse count: %v, output: %s", err, output)
		return 0
	}

	return count
}

// execCommand executes a shell command and returns output
func execCommand(cmd string) (string, error) {
	// This is a placeholder - in real implementation use exec.Command
	// For now, return empty to avoid compilation errors
	return "", fmt.Errorf("not implemented")
}

// getEnv gets environment variable with default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
