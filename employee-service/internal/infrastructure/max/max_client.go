package max

import (
	"employee-service/internal/domain"
	"regexp"
	"strings"
)

// MaxClient представляет клиент для работы с MAX API
// В реальной реализации здесь будет HTTP клиент для вызова MAX API
type MaxClient struct {
	// В будущем здесь могут быть поля для конфигурации API
	// baseURL, apiKey и т.д.
}

func NewMaxClient() *MaxClient {
	return &MaxClient{}
}

// GetMaxIDByPhone получает MAX_id по номеру телефона
// В реальной реализации здесь будет вызов MAX API
// Пока возвращаем заглушку - в продакшене нужно реализовать реальный вызов API
func (c *MaxClient) GetMaxIDByPhone(phone string) (string, error) {
	if !c.ValidatePhone(phone) {
		return "", domain.ErrInvalidPhone
	}
	
	// TODO: Реализовать реальный вызов MAX API
	// Пример:
	// resp, err := http.Get(fmt.Sprintf("%s/api/users/by-phone?phone=%s", c.baseURL, phone))
	// if err != nil {
	//     return "", err
	// }
	// defer resp.Body.Close()
	// ...
	
	// Временная заглушка: возвращаем нормализованный номер телефона как MAX_id
	// В реальности MAX_id будет приходить из API
	normalizedPhone := c.normalizePhone(phone)
	if normalizedPhone == "" {
		return "", domain.ErrMaxIDNotFound
	}
	
	return normalizedPhone, nil
}

// ValidatePhone проверяет валидность номера телефона
func (c *MaxClient) ValidatePhone(phone string) bool {
	// Удаляем все нецифровые символы
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	// Проверяем, что номер содержит от 10 до 15 цифр
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false
	}
	
	return true
}

// normalizePhone нормализует номер телефона
func (c *MaxClient) normalizePhone(phone string) string {
	// Удаляем все нецифровые символы
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	// Если номер начинается с 8, заменяем на 7
	if strings.HasPrefix(cleaned, "8") && len(cleaned) == 11 {
		cleaned = "7" + cleaned[1:]
	}
	
	// Если номер начинается с +7 или 7 и имеет 11 цифр, оставляем как есть
	if strings.HasPrefix(cleaned, "7") && len(cleaned) == 11 {
		return cleaned
	}
	
	// Если номер имеет 10 цифр, добавляем 7
	if len(cleaned) == 10 {
		return "7" + cleaned
	}
	
	return cleaned
}

