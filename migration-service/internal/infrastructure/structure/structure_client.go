package structure

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"migration-service/internal/domain"
	"net/http"
)

// HTTPClient implements StructureService using HTTP REST API
type HTTPClient struct {
	baseURL string
	client  *http.Client
}

// NewHTTPClient creates a new HTTP client for Structure Service
func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// CreateStructureRequest represents the request to create structure
type CreateStructureRequest struct {
	INN         string `json:"inn"`
	KPP         string `json:"kpp"`
	FOIV        string `json:"foiv"`
	OrgName     string `json:"org_name"`
	BranchName  string `json:"branch_name,omitempty"`
	FacultyName string `json:"faculty_name"`
	Course      int    `json:"course"`
	GroupNumber string `json:"group_number"`
}

// CreateStructureResponse represents the response from creating structure
type CreateStructureResponse struct {
	UniversityID int  `json:"university_id"`
	BranchID     *int `json:"branch_id,omitempty"`
	FacultyID    *int `json:"faculty_id,omitempty"`
	GroupID      int  `json:"group_id"`
}

// LinkGroupToChatRequest represents the request to link a group to a chat
type LinkGroupToChatRequest struct {
	ChatID int `json:"chat_id"`
}

// CreateStructure creates or updates the full structure hierarchy
func (c *HTTPClient) CreateStructure(ctx context.Context, data *domain.StructureData) (*domain.StructureResult, error) {
	reqBody := CreateStructureRequest{
		INN:         data.INN,
		KPP:         data.KPP,
		FOIV:        data.FOIV,
		OrgName:     data.OrgName,
		BranchName:  data.BranchName,
		FacultyName: data.FacultyName,
		Course:      data.Course,
		GroupNumber: data.GroupNumber,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/structure", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create structure: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response CreateStructureResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.StructureResult{
		UniversityID: response.UniversityID,
		BranchID:     response.BranchID,
		FacultyID:    response.FacultyID,
		GroupID:      response.GroupID,
	}, nil
}

// LinkGroupToChat links a group to a chat
func (c *HTTPClient) LinkGroupToChat(ctx context.Context, groupID int, chatID int) error {
	reqBody := LinkGroupToChatRequest{
		ChatID: chatID,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/groups/%d/chat", c.baseURL, groupID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to link group to chat: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}
