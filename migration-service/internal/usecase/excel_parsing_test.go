package usecase

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/xuri/excelize/v2"
)

// Простые тесты без моков - только проверка парсинга Excel

func createSimpleTestExcelFile(t *testing.T, rows [][]string) string {
	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"

	// Write rows
	for i, row := range rows {
		for j, cell := range row {
			cellName, _ := excelize.CoordinatesToCellName(j+1, i+1)
			f.SetCellValue(sheetName, cellName, cell)
		}
	}

	// Save to temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.xlsx")
	err := f.SaveAs(filePath)
	assert.NoError(t, err)

	return filePath
}

func TestExcelFileParsing_18Columns(t *testing.T) {
	// Create test Excel file with 18 columns
	rows := [][]string{
		// Header
		{"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch", "INN", "KPP", "Faculty", "Course", "Group", "ChatName", "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin"},
		// Data row
		{"79884753064", "496728250", "105014177", "Минобрнауки", "МГТУ", "Головной филиал", "105014177", "10501001", "Политехнический колледж", "2", "ИП-22", "Колледж ИП-22", "79884753064", "file.xlsx", "-69257108032233", "https://max.ru/join/test", "ИСТИНА", "ИСТИНА"},
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get sheets
	sheets := f.GetSheetList()
	assert.NotEmpty(t, sheets)

	// Get rows
	excelRows, err := f.GetRows(sheets[0])
	assert.NoError(t, err)
	assert.Len(t, excelRows, 2) // Header + 1 data row

	// Verify header
	assert.Len(t, excelRows[0], 18)
	assert.Equal(t, "Phone1", excelRows[0][0])
	assert.Equal(t, "AddAdmin", excelRows[0][17])

	// Verify data row
	dataRow := excelRows[1]
	assert.Len(t, dataRow, 18)
	assert.Equal(t, "79884753064", dataRow[0])
	assert.Equal(t, "496728250", dataRow[1])
	assert.Equal(t, "105014177", dataRow[6])
	assert.Equal(t, "10501001", dataRow[7])
	assert.Equal(t, "2", dataRow[9])
	assert.Equal(t, "-69257108032233", dataRow[14])
	assert.Equal(t, "https://max.ru/join/test", dataRow[15])
	assert.Equal(t, "ИСТИНА", dataRow[16])
	assert.Equal(t, "ИСТИНА", dataRow[17])
}

func TestExcelFileParsing_MultipleRows(t *testing.T) {
	// Create test Excel file with multiple rows
	rows := [][]string{
		{"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch", "INN", "KPP", "Faculty", "Course", "Group", "ChatName", "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin"},
		{"79884753064", "496728250", "105014177", "Минобрнауки", "МГТУ", "Филиал 1", "105014177", "10501001", "Факультет 1", "1", "Группа 1", "Чат 1", "79884753064", "file1.xlsx", "-111", "https://max.ru/1", "ИСТИНА", "ИСТИНА"},
		{"79001234567", "123456789", "105014177", "Минобрнауки", "МГТУ", "Филиал 2", "105014177", "10501002", "Факультет 2", "2", "Группа 2", "Чат 2", "79001234567", "file2.xlsx", "-222", "https://max.ru/2", "TRUE", "TRUE"},
		{"79111111111", "111111111", "105014177", "Минобрнауки", "МГТУ", "Филиал 3", "105014177", "10501003", "Факультет 3", "3", "Группа 3", "Чат 3", "79111111111", "file3.xlsx", "-333", "https://max.ru/3", "FALSE", "FALSE"},
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get rows
	excelRows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, excelRows, 4) // Header + 3 data rows

	// Verify each data row has 18 columns
	for i := 1; i < len(excelRows); i++ {
		assert.Len(t, excelRows[i], 18, "Row %d should have 18 columns", i)
	}
}

func TestNormalizePhone_VariousFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"11 digits starting with 7", "79884753064", "+79884753064"},
		{"11 digits starting with 8", "89884753064", "+79884753064"},
		{"10 digits", "9884753064", "+79884753064"},
		{"With spaces", "7 988 475 30 64", "+79884753064"},
		{"With dashes", "7-988-475-30-64", "+79884753064"},
		{"With parentheses", "7(988)475-30-64", "+79884753064"},
		{"With plus already", "+79884753064", "+79884753064"},
		{"Empty string", "", ""},
		{"Only letters", "abc", ""},
		{"Mixed", "7abc988def4753064", "+79884753064"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizePhone(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExcelFileParsing_InsufficientColumns(t *testing.T) {
	// Create test Excel file with only 10 columns
	rows := [][]string{
		{"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch", "INN", "KPP", "Faculty", "Course"},
		{"79884753064", "496728250", "105014177", "Минобрнауки", "МГТУ", "Филиал", "105014177", "10501001", "Факультет", "2"},
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get rows
	excelRows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, excelRows, 2)

	// Data row should have only 10 columns (insufficient)
	assert.Len(t, excelRows[1], 10)
	assert.Less(t, len(excelRows[1]), 18, "Row has insufficient columns")
}

func TestExcelFileParsing_EmptyFile(t *testing.T) {
	// Create empty Excel file
	rows := [][]string{
		{"Header"},
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get rows
	excelRows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, excelRows, 1) // Only header
}

func TestExcelFileParsing_LargeFile(t *testing.T) {
	// Create large Excel file (1000 rows)
	rows := make([][]string, 1001) // Header + 1000 data rows

	// Header
	rows[0] = []string{"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch", "INN", "KPP", "Faculty", "Course", "Group", "ChatName", "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin"}

	// Data rows
	for i := 1; i <= 1000; i++ {
		rows[i] = []string{
			"79884753064", "496728250", "105014177", "Минобрнауки", "МГТУ",
			"Филиал", "105014177", "10501001", "Факультет", "2",
			"Группа", "Чат", "79884753064", "file.xlsx", "-123",
			"https://max.ru/test", "ИСТИНА", "ИСТИНА",
		}
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get rows
	excelRows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, excelRows, 1001) // Header + 1000 data rows

	t.Logf("Successfully parsed large file with %d rows", len(excelRows))
}

func TestExcelFileParsing_CyrillicText(t *testing.T) {
	// Create test Excel file with Cyrillic text
	rows := [][]string{
		{"Phone1", "MaxID", "INN_Ref", "FOIV", "OrgName", "Branch", "INN", "KPP", "Faculty", "Course", "Group", "ChatName", "Phone2", "FileName", "ChatID", "Link", "AddUser", "AddAdmin"},
		{"79884753064", "496728250", "105014177", "Минобрнауки России", "ФЕДЕРАЛЬНОЕ ГОСУДАРСТВЕННОЕ БЮДЖЕТНОЕ ОБРАЗОВАТЕЛЬНОЕ УЧРЕЖДЕНИЕ", "Федеральное государственное бюджетное", "105014177", "10501001", "Политехнический колледж МГТУ", "2", "Колледж ИП-22 (2024", "Колледж ИП-22 (2024 ОФО МГТУ", "79884753064", "Министерство науки.xlsx", "-69257108032233", "https://max.ru/join/fqQlVkO6LU", "ИСТИНА", "ИСТИНА"},
	}

	filePath := createSimpleTestExcelFile(t, rows)

	// Open and parse Excel file
	f, err := excelize.OpenFile(filePath)
	assert.NoError(t, err)
	defer f.Close()

	// Get rows
	excelRows, err := f.GetRows("Sheet1")
	assert.NoError(t, err)
	assert.Len(t, excelRows, 2)

	// Verify Cyrillic text is preserved
	dataRow := excelRows[1]
	assert.Contains(t, dataRow[3], "Минобрнауки")
	assert.Contains(t, dataRow[4], "ФЕДЕРАЛЬНОЕ")
	assert.Contains(t, dataRow[8], "Политехнический")
	assert.Equal(t, "ИСТИНА", dataRow[16])
	assert.Equal(t, "ИСТИНА", dataRow[17])
}
