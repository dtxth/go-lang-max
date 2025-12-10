package phone

import (
	"regexp"
	"strings"
)

// NormalizePhone нормализует номер телефона к формату +7XXXXXXXXXX
// Принимает номера в форматах:
// - +79001234567
// - 79001234567  
// - 89001234567 (заменяет 8 на +7)
// - 9001234567 (добавляет +7)
func NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	// Удаляем все пробелы, дефисы и скобки
	phone = regexp.MustCompile(`[\s\-\(\)]`).ReplaceAllString(phone, "")

	// Если номер уже начинается с +7, возвращаем как есть
	if strings.HasPrefix(phone, "+7") && len(phone) == 12 {
		return phone
	}

	// Если номер начинается с 7 (без +), добавляем +
	if strings.HasPrefix(phone, "7") && len(phone) == 11 {
		return "+" + phone
	}

	// Если номер начинается с 8, заменяем на +7
	if strings.HasPrefix(phone, "8") && len(phone) == 11 {
		return "+7" + phone[1:]
	}

	// Если номер начинается с 9 и длина 10 символов, добавляем +7
	if strings.HasPrefix(phone, "9") && len(phone) == 10 {
		return "+7" + phone
	}

	// Если ничего не подошло, возвращаем исходный номер
	return phone
}

// IsValidRussianPhone проверяет, является ли номер валидным российским номером
func IsValidRussianPhone(phone string) bool {
	normalized := NormalizePhone(phone)
	
	// Проверяем формат +7XXXXXXXXXX где X - цифры
	matched, _ := regexp.MatchString(`^\+7\d{10}$`, normalized)
	return matched
}