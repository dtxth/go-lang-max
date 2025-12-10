package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"migration-service/internal/domain"
	"net/http"
)

// HTTPClient implements ChatService using HTTP REST API
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client for Chat Service
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// CreateChatRequest represents the request to create a chat
type CreateChatRequest struct {
	Name              string  `json:"name"`
	URL               string  `json:"url"`
	ExternalChatID    *string `json:"external_chat_id,omitempty"`
	Source            string  `json:"source"`
	UniversityID      *int    `json:"university_id,omitempty"`
	BranchID          *int    `json:"branch_id,omitempty"`
	FacultyID         *int    `json:"faculty_id,omitempty"`
	ParticipantsCount int     `json:"participants_count"`
}

// CreateChatResponse represents the response from creating a chat
type CreateChatResponse struct {
	ID int `json:"id"`
}

// AddAdministratorRequest represents the request to add an administrator
type AddAdministratorRequest struct {
	Phone    string `json:"phone"`
	MaxID    string `json:"max_id,omitempty"`
	AddUser  bool   `json:"add_user"`
	AddAdmin bool   `json:"add_admin"`
}

// CreateChat creates a new chat
func (c *HTTPClient) CreateChat(ctx context.Context, chat *domain.ChatData) (int, error) {
	var externalChatID *string
	if chat.ExternalChatID != "" {
		externalChatID = &chat.ExternalChatID
	}

	reqBody := CreateChatRequest{
		Name:              chat.Name,
		URL:               chat.URL,
		ExternalChatID:    externalChatID,
		Source:            chat.Source,
		UniversityID:      chat.UniversityID,
		BranchID:          chat.BranchID,
		FacultyID:         chat.FacultyID,
		ParticipantsCount: 0,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chats", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("failed to create chat: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response CreateChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return 0, fmt.Errorf("failed to decode response: %w", err)
	}

	return response.ID, nil
}

// AddAdministrator adds an administrator to a chat
func (c *HTTPClient) AddAdministrator(ctx context.Context, admin *domain.AdministratorData) error {
	reqBody := AddAdministratorRequest{
		Phone:    admin.Phone,
		MaxID:    admin.MaxID,
		AddUser:  admin.AddUser,
		AddAdmin: admin.AddAdmin,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/chats/%d/administrators", c.baseURL, admin.ChatID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to add administrator: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}


