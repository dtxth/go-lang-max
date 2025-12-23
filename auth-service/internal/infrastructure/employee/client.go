package employee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	
	"auth-service/internal/domain"
)

// Client представляет HTTP клиент для employee-service
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient создает новый клиент для employee-service
func NewClient(baseURL string) domain.EmployeeClient {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// UpdateEmployeeByMaxIDRequest представляет запрос на обновление сотрудника по MAX ID
type UpdateEmployeeByMaxIDRequest struct {
	MaxID     string `json:"max_id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// UpdateEmployeeByMaxIDResponse представляет ответ на обновление сотрудника
type UpdateEmployeeByMaxIDResponse struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	MaxID        string `json:"max_id"`
	UniversityID int64  `json:"university_id"`
}

// UpdateEmployeeByMaxID обновляет данные сотрудника по MAX ID
func (c *Client) UpdateEmployeeByMaxID(maxID, firstName, lastName, username string) error {
	req := UpdateEmployeeByMaxIDRequest{
		MaxID:     maxID,
		FirstName: firstName,
		LastName:  lastName,
		Username:  username,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/employees/update-by-max-id", c.baseURL)
	httpReq, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// Сотрудник не найден - это не критическая ошибка
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("employee service returned status %d", resp.StatusCode)
	}

	return nil
}