package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"maxbot-service/internal/infrastructure/monitoring"
)

func TestMonitoringEndpoints(t *testing.T) {
	// Create mock services
	mockMonitoring := monitoring.NewMockMonitoringService()
	
	// Create handler with mock services
	handler := NewMaxBotHTTPHandler(nil, nil, nil, mockMonitoring)

	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
	}{
		{
			name:           "Get webhook stats",
			endpoint:       "/api/v1/monitoring/webhook/stats",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get webhook stats with period",
			endpoint:       "/api/v1/monitoring/webhook/stats?period=day",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get profile coverage",
			endpoint:       "/api/v1/monitoring/profiles/coverage",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get profile quality report",
			endpoint:       "/api/v1/monitoring/profiles/quality",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid period parameter",
			endpoint:       "/api/v1/monitoring/webhook/stats?period=invalid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.endpoint, nil)
			w := httptest.NewRecorder()

			// Route the request to the appropriate handler
			switch tt.endpoint {
			case "/api/v1/monitoring/webhook/stats", "/api/v1/monitoring/webhook/stats?period=day":
				handler.GetWebhookStats(w, req)
			case "/api/v1/monitoring/webhook/stats?period=invalid":
				handler.GetWebhookStats(w, req)
			case "/api/v1/monitoring/profiles/coverage":
				handler.GetProfileCoverage(w, req)
			case "/api/v1/monitoring/profiles/quality":
				handler.GetProfileQualityReport(w, req)
			}

			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			// For successful requests, verify JSON response structure
			if w.Code == http.StatusOK {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
					t.Errorf("Failed to parse JSON response: %v", err)
				}

				// Basic validation that response contains expected fields
				switch tt.endpoint {
				case "/api/v1/monitoring/webhook/stats", "/api/v1/monitoring/webhook/stats?period=day":
					if _, ok := response["total_events"]; !ok {
						t.Error("Expected 'total_events' field in webhook stats response")
					}
				case "/api/v1/monitoring/profiles/coverage":
					if _, ok := response["total_users"]; !ok {
						t.Error("Expected 'total_users' field in profile coverage response")
					}
				case "/api/v1/monitoring/profiles/quality":
					if _, ok := response["total_profiles"]; !ok {
						t.Error("Expected 'total_profiles' field in quality report response")
					}
				}
			}
		})
	}
}

func TestWebhookStatsResponseStructure(t *testing.T) {
	mockMonitoring := monitoring.NewMockMonitoringService()
	handler := NewMaxBotHTTPHandler(nil, nil, nil, mockMonitoring)

	req := httptest.NewRequest("GET", "/api/v1/monitoring/webhook/stats", nil)
	w := httptest.NewRecorder()

	handler.GetWebhookStats(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response WebhookStatsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse webhook stats response: %v", err)
	}

	// Verify response structure
	if response.Period.From == "" {
		t.Error("Expected non-empty period.from")
	}
	if response.Period.To == "" {
		t.Error("Expected non-empty period.to")
	}
	if response.EventsByType == nil {
		t.Error("Expected non-nil events_by_type map")
	}
	if response.ErrorsByType == nil {
		t.Error("Expected non-nil errors_by_type map")
	}
}

func TestProfileQualityReportStructure(t *testing.T) {
	mockMonitoring := monitoring.NewMockMonitoringService()
	handler := NewMaxBotHTTPHandler(nil, nil, nil, mockMonitoring)

	req := httptest.NewRequest("GET", "/api/v1/monitoring/profiles/quality", nil)
	w := httptest.NewRecorder()

	handler.GetProfileQualityReport(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var response ProfileQualityReportResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse quality report response: %v", err)
	}

	// Verify response structure
	if response.GeneratedAt == "" {
		t.Error("Expected non-empty generated_at")
	}
	if response.TotalProfiles == 0 {
		t.Error("Expected non-zero total_profiles")
	}
	if response.SourceBreakdown == nil {
		t.Error("Expected non-nil source_breakdown map")
	}
	if response.RecommendedActions == nil {
		t.Error("Expected non-nil recommended_actions slice")
	}
	if response.DataIssues == nil {
		t.Error("Expected non-nil data_issues slice")
	}

	// Verify quality metrics structure
	if response.QualityMetrics.QualityScore < 0 || response.QualityMetrics.QualityScore > 100 {
		t.Errorf("Quality score should be between 0-100, got %f", response.QualityMetrics.QualityScore)
	}
}