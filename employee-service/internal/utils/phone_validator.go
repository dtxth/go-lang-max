package utils

import (
	"regexp"
	"strings"
)

// PhoneValidator предоставляет методы для валидации номеров телефонов
type PhoneValidator struct{}

// NewPhoneValidator создает новый валидатор телефонов
func NewPhoneValidator() *PhoneValidator {
	return &PhoneValidator{}
}

// ValidatePhone проверяет корректность номера телефона
// Поддерживает российские номера в формате +7XXXXXXXXXX
func (v *PhoneValidator) ValidatePhone(phone string) bool {
	if phone == "" {
		return false
	}

	// Убираем все пробелы и дефисы
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// Проверяем российские номера
	// Формат: +7XXXXXXXXXX (11 цифр после +7)
	russianPhoneRegex := regexp.MustCompile(`^\+7[0-9]{10}$`)
	if russianPhoneRegex.MatchString(cleanPhone) {
		return true
	}

	// Проверяем номера без + (начинающиеся с 7 или 8)
	// Формат: 7XXXXXXXXXX или 8XXXXXXXXXX
	russianPhoneWithoutPlusRegex := regexp.MustCompile(`^[78][0-9]{10}$`)
	if russianPhoneWithoutPlusRegex.MatchString(cleanPhone) {
		return true
	}

	return false
}

// NormalizePhone нормализует номер телефона к формату +7XXXXXXXXXX
func (v *PhoneValidator) NormalizePhone(phone string) string {
	if phone == "" {
		return ""
	}

	// Убираем все пробелы и дефисы
	cleanPhone := strings.ReplaceAll(phone, " ", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "-", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, "(", "")
	cleanPhone = strings.ReplaceAll(cleanPhone, ")", "")

	// Если уже в правильном формате +7XXXXXXXXXX
	russianPhoneRegex := regexp.MustCompile(`^\+7[0-9]{10}$`)
	if russianPhoneRegex.MatchString(cleanPhone) {
		return cleanPhone
	}

	// Если начинается с 8, заменяем на +7
	if strings.HasPrefix(cleanPhone, "8") && len(cleanPhone) == 11 {
		return "+7" + cleanPhone[1:]
	}

	// Если начинается с 7, добавляем +
	if strings.HasPrefix(cleanPhone, "7") && len(cleanPhone) == 11 {
		return "+" + cleanPhone
	}

	// Возвращаем как есть, если не удалось нормализовать
	return phone
}