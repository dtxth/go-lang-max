package excel

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"

	"structure-service/internal/domain"
	"github.com/xuri/excelize/v2"
)

// ParseExcel парсит Excel файл и возвращает массив строк
func ParseExcel(fileBytes []byte) ([]*domain.ExcelRow, error) {
	f, err := excelize.OpenReader(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to open excel file: %w", err)
	}
	defer f.Close()

	// Получаем имя первого листа
	sheetName := f.GetSheetName(0)
	if sheetName == "" {
		return nil, fmt.Errorf("no sheets found")
	}

	// Читаем все строки
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to read rows: %w", err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("file must contain at least header and one data row")
	}

	// Парсим заголовки (первая строка)
	headerMap := make(map[string]int)
	headers := rows[0]
	for i, header := range headers {
		headerMap[strings.ToLower(strings.TrimSpace(header))] = i
	}

	// Определяем индексы колонок
	adminPhoneIdx := findColumnIndex(headerMap, []string{"номер телефона администратора", "телефон администратора", "admin phone"})
	innIdx := findColumnIndex(headerMap, []string{"инн", "inn"})
	foivIdx := findColumnIndex(headerMap, []string{"фоив", "foiv"})
	orgIdx := findColumnIndex(headerMap, []string{"наименование организации", "организация", "organization"})
	branchIdx := findColumnIndex(headerMap, []string{"наименование головного подразделения", "филиал", "branch", "подразделение"})
	kppIdx := findColumnIndex(headerMap, []string{"кпп", "kpp"})
	facultyIdx := findColumnIndex(headerMap, []string{"факультет", "институт", "структурная классификация", "faculty", "institute"})
	courseIdx := findColumnIndex(headerMap, []string{"курс обучения", "курс", "course"})
	groupIdx := findColumnIndex(headerMap, []string{"номер группы", "группа", "group", "group number"})
	chatNameIdx := findColumnIndex(headerMap, []string{"название чата", "chat name", "name"})
	chatURLIdx := findColumnIndex(headerMap, []string{"ссылка на чат", "chat url", "url", "link"})

	// Обязательные поля
	if innIdx == -1 || orgIdx == -1 || facultyIdx == -1 || groupIdx == -1 {
		return nil, fmt.Errorf("required columns not found: INN, Organization, Faculty, Group are required")
	}

	var excelRows []*domain.ExcelRow

	// Парсим данные (начиная со второй строки)
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		if len(row) == 0 {
			continue
		}

		excelRow := &domain.ExcelRow{}

		if adminPhoneIdx >= 0 && adminPhoneIdx < len(row) {
			excelRow.AdminPhone = strings.TrimSpace(row[adminPhoneIdx])
		}
		if innIdx < len(row) {
			excelRow.INN = strings.TrimSpace(row[innIdx])
		}
		if foivIdx >= 0 && foivIdx < len(row) {
			excelRow.FOIV = strings.TrimSpace(row[foivIdx])
		}
		if orgIdx < len(row) {
			excelRow.Organization = strings.TrimSpace(row[orgIdx])
		}
		if branchIdx >= 0 && branchIdx < len(row) {
			excelRow.Branch = strings.TrimSpace(row[branchIdx])
		}
		if kppIdx >= 0 && kppIdx < len(row) {
			excelRow.KPP = strings.TrimSpace(row[kppIdx])
		}
		if facultyIdx < len(row) {
			excelRow.Faculty = strings.TrimSpace(row[facultyIdx])
		}
		if courseIdx >= 0 && courseIdx < len(row) {
			courseStr := strings.TrimSpace(row[courseIdx])
			if courseStr != "" {
				if course, err := strconv.Atoi(courseStr); err == nil {
					excelRow.Course = course
				}
			}
		}
		if groupIdx < len(row) {
			excelRow.GroupNumber = strings.TrimSpace(row[groupIdx])
		}
		if chatNameIdx >= 0 && chatNameIdx < len(row) {
			excelRow.ChatName = strings.TrimSpace(row[chatNameIdx])
		}
		if chatURLIdx >= 0 && chatURLIdx < len(row) {
			excelRow.ChatURL = strings.TrimSpace(row[chatURLIdx])
			// Извлекаем ID чата из URL (если есть)
			excelRow.ChatID = extractChatIDFromURL(excelRow.ChatURL)
		}

		// Пропускаем пустые строки
		if excelRow.INN == "" && excelRow.Organization == "" {
			continue
		}

		excelRows = append(excelRows, excelRow)
	}

	return excelRows, nil
}

// findColumnIndex ищет индекс колонки по различным вариантам названий
func findColumnIndex(headerMap map[string]int, variants []string) int {
	for _, variant := range variants {
		if idx, ok := headerMap[variant]; ok {
			return idx
		}
	}
	return -1
}

// extractChatIDFromURL извлекает ID чата из URL
func extractChatIDFromURL(url string) string {
	// Пример: https://vk.me/join/AbCdEfGhIjKlMnOpQrStUvWxYz123456
	// Или: https://vk.me/join/chat_id
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// Убираем параметры запроса, если есть
		if idx := strings.Index(lastPart, "?"); idx != -1 {
			lastPart = lastPart[:idx]
		}
		return lastPart
	}
	return ""
}

