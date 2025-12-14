package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"migration-service/internal/domain"
	"net/http"
)

// UniversityHTTPRepository implements UniversityRepository using HTTP calls to Structure Service
type UniversityHTTPRepository struct {
	baseURL string
	client  *http.Client
}

// NewUniversityHTTPRepository creates a new UniversityHTTPRepository
func NewUniversityHTTPRepository(baseURL string) *UniversityHTTPRepository {
	return &UniversityHTTPRepository{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

// UniversityResponse represents the university response from Structure Service
type UniversityResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	INN  string `json:"inn"`
	KPP  string `json:"kpp"`
}

// GetByINN retrieves a university by INN
func (r *UniversityHTTPRepository) GetByINN(ctx context.Context, inn string) (*domain.University, error) {
	url := fmt.Sprintf("%s/universities?inn=%s", r.baseURL, inn)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // University not found
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get university: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response UniversityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.University{
		ID:   response.ID,
		Name: response.Name,
		INN:  response.INN,
		KPP:  response.KPP,
	}, nil
}

// GetByINNAndKPP retrieves a university by INN and KPP
func (r *UniversityHTTPRepository) GetByINNAndKPP(ctx context.Context, inn, kpp string) (*domain.University, error) {
	url := fmt.Sprintf("%s/universities?inn=%s&kpp=%s", r.baseURL, inn, kpp)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // University not found
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get university: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response UniversityResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &domain.University{
		ID:   response.ID,
		Name: response.Name,
		INN:  response.INN,
		KPP:  response.KPP,
	}, nil
}

// Create creates a new university (not implemented - universities are created via Structure Service)
func (r *UniversityHTTPRepository) Create(ctx context.Context, university *domain.University) error {
	return fmt.Errorf("create university not implemented - use Structure Service")
}
