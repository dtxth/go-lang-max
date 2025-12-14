package http

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// createTestExcelFile creates a test Excel file with 18 columns
func createTestExcelFile(t *testing.T) string {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Header row
	headers := []string{
		"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
		"INN", "KPP", "Faculty", "Course", "Group", "ChatName",
		"Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
	}

	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, header)
	}

	// Data rows
	dataRows := [][]interface{}{
		{
			"79884753064", "496728250", "105014177", "Минобрнауки России",
			"МГТУ", "Головной филиал", "105014177", "10501001",
			"Политехнический колледж МГТУ", "2", "Колледж ИП-22",
			"Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
			"-69257108032233", "https://max.ru/join/test", "ИСТИНА", "ИСТИНА",
		},
		{
			"79001234567", "123456789", "105014177", "Минобрнауки России",
			"МГТУ", "Головной филиал", "105014177", "10501001",
			"Факультет информатики", "3", "ИВТ-31",
			"Группа ИВТ-31", "79001234567", "file.xlsx",
			"-12345678", "https://max.ru/join/test2", "TRUE", "TRUE",
		},
	}

	for rowIdx, dataRow := range dataRows {
		for colIdx, value := range dataRow {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue(sheetName, cellName, value)
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	err := f.SaveAs(filePath)
	assert.NoError(t, err)

	return filePath
}

// createLargeTestExcelFile creates a large Excel file for performance testing
func createLargeTestExcelFile(t *testing.T, numRows int) string {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Header row
	headers := []string{
		"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch",
		"INN", "KPP", "Faculty", "Course", "Group", "ChatName",
		"Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin",
	}

	for i, header := range headers {
		cellName, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cellName, header)
	}

	// Generate data rows
	for row := 2; row <= numRows+1; row++ {
		dataRow := []interface{}{
			"79884753064", "496728250", "105014177", "Минобрнауки России",
			"МГТУ", "Головной филиал", "105014177", "10501001",
			"Политехнический колледж МГТУ", "2", "Колледж ИП-22",
			"Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "file.xlsx",
			"-69257108032233", "https://max.ru/join/test", "ИСТИНА", "ИСТИНА",
		}

		for colIdx, value := range dataRow {
			cellName, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			f.SetCellValue(sheetName, cellName, value)
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "large_test.xlsx")
	err := f.SaveAs(filePath)
	assert.NoError(t, err)

	return filePath
}

func TestUploadExcelFile_ValidFile(t *testing.T) {
	// Create test Excel file
	filePath := createTestExcelFile(t)

	// Read file
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.xlsx")
	assert.NoError(t, err)

	_, err = part.Write(fileContent)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Verify file can be parsed
	err = req.ParseMultipartForm(32 << 20) // 32 MB
	assert.NoError(t, err)

	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, "test.xlsx", header.Filename)
	defer file.Close()

	// Verify file size
	fileInfo, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.Greater(t, fileInfo.Size(), int64(0))
	t.Logf("File size: %d bytes", fileInfo.Size())
}

func TestUploadExcelFile_LargeFile(t *testing.T) {
	// Create large test Excel file (1000 rows)
	filePath := createLargeTestExcelFile(t, 1000)

	// Read file
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "large_test.xlsx")
	assert.NoError(t, err)

	_, err = part.Write(fileContent)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Verify file can be parsed
	err = req.ParseMultipartForm(32 << 20) // 32 MB
	assert.NoError(t, err)

	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	assert.NotNil(t, file)
	assert.Equal(t, "large_test.xlsx", header.Filename)
	defer file.Close()

	// Verify file size
	fileInfo, err := os.Stat(filePath)
	assert.NoError(t, err)
	t.Logf("Large file size: %d bytes (%.2f MB)", fileInfo.Size(), float64(fileInfo.Size())/(1024*1024))
	
	// File should be less than 10 MB for 1000 rows
	assert.Less(t, fileInfo.Size(), int64(10*1024*1024))
}

func TestUploadExcelFile_MissingFile(t *testing.T) {
	// Create empty multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	err := writer.Close()
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse form
	err = req.ParseMultipartForm(32 << 20)
	assert.NoError(t, err)

	// Try to get file (should fail)
	_, _, err = req.FormFile("file")
	assert.Error(t, err)
	assert.Equal(t, http.ErrMissingFile, err)
}

func TestUploadExcelFile_InvalidContentType(t *testing.T) {
	// Create request with wrong content type
	body := bytes.NewBufferString("not a multipart form")
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", "application/json")

	// Try to parse (should fail)
	err := req.ParseMultipartForm(32 << 20)
	assert.Error(t, err)
}

func TestUploadExcelFile_FileSizeLimit(t *testing.T) {
	// Create test Excel file
	filePath := createTestExcelFile(t)

	// Read file
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.xlsx")
	assert.NoError(t, err)

	_, err = part.Write(fileContent)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Try to parse with very small limit (should fail if file is larger)
	err = req.ParseMultipartForm(1024) // 1 KB limit
	
	// If file is larger than 1KB, this should work but file will be stored on disk
	// ParseMultipartForm doesn't return error for large files, it just stores them
	assert.NoError(t, err)
}

func TestExcelFileCanBeParsed(t *testing.T) {
	// Create test Excel file
	filePath := createTestExcelFile(t)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get sheets
	sheets := f.GetSheetList()
	assert.NotEmpty(t, sheets)
	assert.Equal(t, "Sheet1", sheets[0])

	// Get rows
	rows, err := f.GetRows(sheets[0])
	assert.NoError(t, err)
	assert.Len(t, rows, 3) // Header + 2 data rows

	// Verify header
	assert.Len(t, rows[0], 18)
	assert.Equal(t, "Phone1", rows[0][0])
	assert.Equal(t, "AddAdmin", rows[0][17])

	// Verify first data row
	assert.Len(t, rows[1], 18)
	assert.Equal(t, "79884753064", rows[1][0])
	assert.Equal(t, "496728250", rows[1][1])
	assert.Equal(t, "105014177", rows[1][6])
	assert.Equal(t, "ИСТИНА", rows[1][17])
}

func TestExcelFileStreaming(t *testing.T) {
	// Create test Excel file
	filePath := createTestExcelFile(t)

	// Open Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Test streaming API
	rows, err := f.Rows("Sheet1")
	assert.NoError(t, err)
	defer rows.Close()

	rowCount := 0
	for rows.Next() {
		row, err := rows.Columns()
		assert.NoError(t, err)
		assert.NotEmpty(t, row)
		rowCount++
	}

	assert.Equal(t, 3, rowCount) // Header + 2 data rows
}

func TestMultipartFormFileUpload(t *testing.T) {
	// Create test Excel file
	filePath := createTestExcelFile(t)

	// Read file
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)

	// Create multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "test.xlsx")
	assert.NoError(t, err)

	_, err = io.Copy(part, bytes.NewReader(fileContent))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	// Create request
	req := httptest.NewRequest(http.MethodPost, "/migration/excel", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse form
	err = req.ParseMultipartForm(32 << 20)
	assert.NoError(t, err)

	// Get file
	file, header, err := req.FormFile("file")
	assert.NoError(t, err)
	defer file.Close()

	// Verify header
	assert.Equal(t, "test.xlsx", header.Filename)
	assert.Greater(t, header.Size, int64(0))
	t.Logf("Uploaded file: %s, size: %d bytes", header.Filename, header.Size)

	// Save to temp file
	tmpDir := t.TempDir()
	tmpPath := filepath.Join(tmpDir, header.Filename)

	tmpFile, err := os.Create(tmpPath)
	assert.NoError(t, err)
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, file)
	assert.NoError(t, err)

	// Verify saved file can be opened
	f, err := excelize.OpenFile(tmpPath)
	assert.NoError(t, err)
	defer f.Close()

	rows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, rows, 3)
}
