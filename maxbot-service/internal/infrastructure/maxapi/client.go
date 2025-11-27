package maxapi

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"maxbot-service/internal/domain"
)

// Client инкапсулирует взаимодействие с внешним MAX API.
// Пока реализована заглушка с нормализацией телефона.
var nonDigitRegexp = regexp.MustCompile(`\D`)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *Client) GetMaxIDByPhone(ctx context.Context, phone string) (string, error) {
	valid, normalized, err := c.ValidatePhone(phone)
	if err != nil {
		return "", err
	}
	if !valid {
		return "", domain.ErrInvalidPhone
	}

	// TODO: заменить на реальный вызов API.
	if normalized == "" {
		return "", domain.ErrMaxIDNotFound
	}

	return normalized, nil
}

func (c *Client) ValidatePhone(phone string) (bool, string, error) {
	cleaned := nonDigitRegexp.ReplaceAllString(phone, "")

	if len(cleaned) < 10 || len(cleaned) > 15 {
		return false, "", nil
	}

	normalized := c.normalizePhone(cleaned)
	if normalized == "" {
		return false, "", nil
	}

	return true, normalized, nil
}

func (c *Client) normalizePhone(phone string) string {
	digits := strings.TrimSpace(phone)

	if strings.HasPrefix(digits, "8") && len(digits) == 11 {
		digits = "7" + digits[1:]
	}

	if strings.HasPrefix(digits, "7") && len(digits) == 11 {
		return digits
	}

	if len(digits) == 10 {
		return "7" + digits
	}

	if len(digits) >= 10 && len(digits) <= 15 {
		return digits
	}

	return ""
}
