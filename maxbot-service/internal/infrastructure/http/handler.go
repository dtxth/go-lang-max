package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"maxbot-service/internal/domain"
	"maxbot-service/internal/infrastructure/errors"
	"maxbot-service/internal/usecase"
)

// MaxBotHTTPHandler handles HTTP requests for MaxBot service
type MaxBotHTTPHandler struct {
	service           *usecase.MaxBotService
	webhookHandler    *usecase.WebhookHandlerService
	profileManagement *usecase.ProfileManagementService
	monitoring        domain.MonitoringService
}

// NewMaxBotHTTPHandler creates a new HTTP handler
func NewMaxBotHTTPHandler(service *usecase.MaxBotService, webhookHandler *usecase.WebhookHandlerService, profileManagement *usecase.ProfileManagementService, monitoring domain.MonitoringService) *MaxBotHTTPHandler {
	return &MaxBotHTTPHandler{
		service:           service,
		webhookHandler:    webhookHandler,
		profileManagement: profileManagement,
		monitoring:        monitoring,
	}
}

// BotInfoResponse represents the response for /me endpoint
// @Description Bot information response
type BotInfoResponse struct {
	Name    string `json:"name" example:"MAX Bot"`                    // Bot name
	AddLink string `json:"add_link" example:"https://max.ru/add-bot"` // Link to add the bot
} // @name BotInfoResponse

// ErrorResponse represents an error response
// @Description Error response structure
type ErrorResponse struct {
	Error   string `json:"error" example:"internal_error"`           // Error code
	Message string `json:"message" example:"Internal server error"`  // Error message
} // @name ErrorResponse



// ChatInfoResponse represents the response for chat info endpoint
// @Description Chat information response
type ChatInfoResponse struct {
	ChatID            int64  `json:"chat_id" example:"123456789"`                    // Chat ID
	Title             string `json:"title" example:"Test Chat"`                      // Chat title
	Type              string `json:"type" example:"group"`                           // Chat type
	ParticipantsCount int    `json:"participants_count" example:"25"`               // Number of participants
	Description       string `json:"description" example:"Test chat description"`   // Chat description
} // @name ChatInfoResponse

// GetChatInfo godoc
// @Summary Get chat information
// @Description Get information about a specific chat from MAX Messenger
// @Tags Chat
// @Accept json
// @Produce json
// @Param chat_id path int64 true "Chat ID"
// @Success 200 {object} ChatInfoResponse "Chat information"
// @Failure 400 {object} ErrorResponse "Invalid chat ID"
// @Failure 404 {object} ErrorResponse "Chat not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /chats/{chat_id} [get]
func (h *MaxBotHTTPHandler) GetChatInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	log.Printf("[DEBUG] GetChatInfo called with URL: %s", r.URL.Path)

	// Extract chat_id from URL path
	chatIDStr := extractChatIDFromPath(r)
	log.Printf("[DEBUG] Extracted chat_id: '%s'", chatIDStr)
	
	if chatIDStr == "" {
		log.Printf("[ERROR] chat_id is empty")
		errors.WriteError(w, errors.ValidationError("chat_id is required"), requestID)
		return
	}

	// Parse chat_id as int64
	var chatID int64
	if _, err := fmt.Sscanf(chatIDStr, "%d", &chatID); err != nil {
		errors.WriteError(w, errors.ValidationError("invalid chat_id format"), requestID)
		return
	}

	// Get chat info from service
	chatInfo, err := h.service.GetChatInfo(ctx, chatID)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Create response
	response := ChatInfoResponse{
		ChatID:            chatInfo.ChatID,
		Title:             chatInfo.Title,
		Type:              chatInfo.Type,
		ParticipantsCount: chatInfo.ParticipantsCount,
		Description:       chatInfo.Description,
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// GetMe godoc
// @Summary Get bot information
// @Description Get bot name and add bot link
// @Tags Bot
// @Accept json
// @Produce json
// @Success 200 {object} BotInfoResponse "Bot information"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /me [get]
func (h *MaxBotHTTPHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Call the service
	botInfo, err := h.service.GetMe(ctx)
	if err != nil {
		errors.WriteError(w, err, getRequestID(ctx))
		return
	}

	// Create response
	response := BotInfoResponse{
		Name:    botInfo.Name,
		AddLink: botInfo.AddLink,
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), getRequestID(ctx))
		return
	}
}



// HandleMaxWebhook godoc
// @Summary Handle MAX webhook events
// @Description Process incoming webhook events from MAX Messenger
// @Tags Webhook
// @Accept json
// @Produce json
// @Param event body domain.MaxWebhookEvent true "Webhook event"
// @Success 200 {object} map[string]string "Event processed successfully"
// @Failure 400 {object} ErrorResponse "Invalid request body"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /webhook/max [post]
func (h *MaxBotHTTPHandler) HandleMaxWebhook(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Читаем тело запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read webhook request body: %v", err)
		errors.WriteError(w, errors.ValidationError("Failed to read request body"), requestID)
		return
	}
	defer r.Body.Close()

	// Логируем входящий webhook для отладки
	log.Printf("Received webhook event: %s", string(body))

	// Парсим JSON
	var event domain.MaxWebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		log.Printf("Failed to parse webhook JSON: %v", err)
		// Возвращаем 200 OK даже при ошибке парсинга, как требует спецификация
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok", "message": "Invalid JSON format"})
		return
	}

	// Обрабатываем событие
	err = h.webhookHandler.HandleMaxWebhook(ctx, event)
	if err != nil {
		log.Printf("Error processing webhook event: %v", err)
		// Возвращаем 200 OK даже при ошибке обработки, как требует спецификация
	}

	// Всегда возвращаем 200 OK для подтверждения получения
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// ProfileResponse represents a user profile response
// @Description User profile information
type ProfileResponse struct {
	UserID           string `json:"user_id" example:"12345"`                          // User ID
	MaxFirstName     string `json:"max_first_name" example:"Иван"`                   // First name from MAX
	MaxLastName      string `json:"max_last_name" example:"Петров"`                  // Last name from MAX
	UserProvidedName string `json:"user_provided_name" example:"Иван Петрович"`      // User-provided name
	DisplayName      string `json:"display_name" example:"Иван Петрович"`            // Display name (prioritized)
	Source           string `json:"source" example:"user_input"`                     // Profile source
	LastUpdated      string `json:"last_updated" example:"2024-01-15T10:30:00Z"`    // Last update time
	HasFullName      bool   `json:"has_full_name" example:"true"`                    // Whether profile has full name
} // @name ProfileResponse

// ProfileUpdateRequest represents a profile update request
// @Description Profile update request
type ProfileUpdateRequest struct {
	MaxFirstName     *string `json:"max_first_name,omitempty" example:"Иван"`        // First name from MAX
	MaxLastName      *string `json:"max_last_name,omitempty" example:"Петров"`       // Last name from MAX
	UserProvidedName *string `json:"user_provided_name,omitempty" example:"Иван П."` // User-provided name
} // @name ProfileUpdateRequest

// SetNameRequest represents a set name request
// @Description Set user-provided name request
type SetNameRequest struct {
	Name string `json:"name" example:"Иван Петрович" binding:"required"` // User-provided name
} // @name SetNameRequest

// ProfileStatsResponse represents profile statistics
// @Description Profile statistics response
type ProfileStatsResponse struct {
	TotalProfiles        int64            `json:"total_profiles" example:"1000"`        // Total number of profiles
	ProfilesWithFullName int64            `json:"profiles_with_full_name" example:"750"` // Profiles with full names
	ProfilesBySource     map[string]int64 `json:"profiles_by_source"`                    // Profiles by source
} // @name ProfileStatsResponse

// WebhookStatsResponse represents webhook processing statistics
// @Description Webhook processing statistics response
type WebhookStatsResponse struct {
	Period                TimePeriodResponse `json:"period"`                          // Time period
	TotalEvents           int64              `json:"total_events" example:"5000"`    // Total webhook events
	SuccessfulEvents      int64              `json:"successful_events" example:"4800"` // Successful events
	FailedEvents          int64              `json:"failed_events" example:"200"`    // Failed events
	EventsByType          map[string]int64   `json:"events_by_type"`                  // Events by type
	ProfilesExtracted     int64              `json:"profiles_extracted" example:"3000"` // Profiles extracted from events
	ProfilesStored        int64              `json:"profiles_stored" example:"2900"` // Profiles stored in cache
	AverageProcessingTime float64            `json:"average_processing_time_ms" example:"150.5"` // Average processing time
	ErrorsByType          map[string]int64   `json:"errors_by_type"`                  // Errors by type
} // @name WebhookStatsResponse

// TimePeriodResponse represents a time period
// @Description Time period for statistics
type TimePeriodResponse struct {
	From string `json:"from" example:"2024-01-15T00:00:00Z"` // Start time
	To   string `json:"to" example:"2024-01-16T00:00:00Z"`   // End time
} // @name TimePeriodResponse

// ProfileCoverageResponse represents profile coverage metrics
// @Description Profile coverage metrics response
type ProfileCoverageResponse struct {
	TotalUsers         int64            `json:"total_users" example:"10000"`         // Total users
	UsersWithProfiles  int64            `json:"users_with_profiles" example:"8000"`  // Users with profiles
	UsersWithFullNames int64            `json:"users_with_full_names" example:"6000"` // Users with full names
	CoveragePercentage float64          `json:"coverage_percentage" example:"80.0"`  // Coverage percentage
	FullNamePercentage float64          `json:"full_name_percentage" example:"60.0"` // Full name percentage
	ProfilesBySource   map[string]int64 `json:"profiles_by_source"`                   // Profiles by source
	LastUpdated        string           `json:"last_updated" example:"2024-01-15T10:30:00Z"` // Last update time
} // @name ProfileCoverageResponse

// ProfileQualityReportResponse represents profile quality report
// @Description Profile quality report response
type ProfileQualityReportResponse struct {
	GeneratedAt        string                           `json:"generated_at" example:"2024-01-15T10:30:00Z"` // Report generation time
	TotalProfiles      int64                            `json:"total_profiles" example:"10000"`              // Total profiles
	QualityMetrics     ProfileQualityMetricsResponse    `json:"quality_metrics"`                             // Quality metrics
	SourceBreakdown    map[string]SourceQualityResponse `json:"source_breakdown"`                            // Quality by source
	RecommendedActions []string                         `json:"recommended_actions"`                         // Recommended actions
	DataIssues         []ProfileDataIssueResponse       `json:"data_issues"`                                 // Data issues
} // @name ProfileQualityReportResponse

// ProfileQualityMetricsResponse represents quality metrics
// @Description Profile quality metrics
type ProfileQualityMetricsResponse struct {
	CompleteProfiles  int64   `json:"complete_profiles" example:"7000"`   // Complete profiles
	PartialProfiles   int64   `json:"partial_profiles" example:"2000"`    // Partial profiles
	EmptyProfiles     int64   `json:"empty_profiles" example:"1000"`      // Empty profiles
	StaleProfiles     int64   `json:"stale_profiles" example:"500"`       // Stale profiles
	QualityScore      float64 `json:"quality_score" example:"85.5"`       // Quality score (0-100)
	CompletenessScore float64 `json:"completeness_score" example:"70.0"`  // Completeness score
	FreshnessScore    float64 `json:"freshness_score" example:"90.0"`     // Freshness score
} // @name ProfileQualityMetricsResponse

// SourceQualityResponse represents quality metrics by source
// @Description Quality metrics by source
type SourceQualityResponse struct {
	Count            int64   `json:"count" example:"5000"`           // Count of profiles
	CompleteProfiles int64   `json:"complete_profiles" example:"4000"` // Complete profiles
	AverageAge       float64 `json:"average_age_days" example:"7.5"` // Average age in days
	QualityScore     float64 `json:"quality_score" example:"85.0"`   // Quality score
} // @name SourceQualityResponse

// ProfileDataIssueResponse represents a data issue
// @Description Profile data issue
type ProfileDataIssueResponse struct {
	Type        string `json:"type" example:"incomplete_profiles"`        // Issue type
	Description string `json:"description" example:"Missing full names"` // Issue description
	Count       int64  `json:"count" example:"2000"`                     // Affected profiles count
	Severity    string `json:"severity" example:"medium"`                // Issue severity
} // @name ProfileDataIssueResponse

// GetProfile godoc
// @Summary Get user profile
// @Description Get user profile information by user ID
// @Tags Profile
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} ProfileResponse "User profile"
// @Failure 400 {object} ErrorResponse "Invalid user ID"
// @Failure 404 {object} ErrorResponse "Profile not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /profiles/{user_id} [get]
func (h *MaxBotHTTPHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Извлекаем user_id из URL
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		errors.WriteError(w, errors.ValidationError("user_id is required"), requestID)
		return
	}

	// Получаем профиль
	profile, err := h.profileManagement.GetProfile(ctx, userID)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	response := ProfileResponse{
		UserID:           profile.UserID,
		MaxFirstName:     profile.MaxFirstName,
		MaxLastName:      profile.MaxLastName,
		UserProvidedName: profile.UserProvidedName,
		DisplayName:      profile.GetDisplayName(),
		Source:           string(profile.Source),
		LastUpdated:      profile.LastUpdated.Format(time.RFC3339),
		HasFullName:      profile.HasFullName(),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update user profile information
// @Tags Profile
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param profile body ProfileUpdateRequest true "Profile updates"
// @Success 200 {object} ProfileResponse "Updated profile"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /profiles/{user_id} [put]
func (h *MaxBotHTTPHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Извлекаем user_id из URL
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		errors.WriteError(w, errors.ValidationError("user_id is required"), requestID)
		return
	}

	// Парсим тело запроса
	var req ProfileUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.ValidationError("Invalid JSON format"), requestID)
		return
	}
	defer r.Body.Close()

	// Формируем обновления
	updates := domain.ProfileUpdates{
		MaxFirstName:     req.MaxFirstName,
		MaxLastName:      req.MaxLastName,
		UserProvidedName: req.UserProvidedName,
	}

	// Обновляем профиль
	profile, err := h.profileManagement.UpdateProfile(ctx, userID, updates)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	response := ProfileResponse{
		UserID:           profile.UserID,
		MaxFirstName:     profile.MaxFirstName,
		MaxLastName:      profile.MaxLastName,
		UserProvidedName: profile.UserProvidedName,
		DisplayName:      profile.GetDisplayName(),
		Source:           string(profile.Source),
		LastUpdated:      profile.LastUpdated.Format(time.RFC3339),
		HasFullName:      profile.HasFullName(),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// SetUserProvidedName godoc
// @Summary Set user-provided name
// @Description Set name provided by user
// @Tags Profile
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param name body SetNameRequest true "User-provided name"
// @Success 200 {object} ProfileResponse "Updated profile"
// @Failure 400 {object} ErrorResponse "Invalid request"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /profiles/{user_id}/name [post]
func (h *MaxBotHTTPHandler) SetUserProvidedName(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Извлекаем user_id из URL
	userID := extractUserIDFromPath(r.URL.Path)
	if userID == "" {
		errors.WriteError(w, errors.ValidationError("user_id is required"), requestID)
		return
	}

	// Парсим тело запроса
	var req SetNameRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteError(w, errors.ValidationError("Invalid JSON format"), requestID)
		return
	}
	defer r.Body.Close()

	// Устанавливаем имя
	profile, err := h.profileManagement.SetUserProvidedName(ctx, userID, req.Name)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	response := ProfileResponse{
		UserID:           profile.UserID,
		MaxFirstName:     profile.MaxFirstName,
		MaxLastName:      profile.MaxLastName,
		UserProvidedName: profile.UserProvidedName,
		DisplayName:      profile.GetDisplayName(),
		Source:           string(profile.Source),
		LastUpdated:      profile.LastUpdated.Format(time.RFC3339),
		HasFullName:      profile.HasFullName(),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// GetProfileStats godoc
// @Summary Get profile statistics
// @Description Get statistics about user profiles
// @Tags Profile
// @Accept json
// @Produce json
// @Success 200 {object} ProfileStatsResponse "Profile statistics"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /profiles/stats [get]
func (h *MaxBotHTTPHandler) GetProfileStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Получаем статистику
	stats, err := h.profileManagement.GetProfileStats(ctx)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	profilesBySource := make(map[string]int64)
	for source, count := range stats.ProfilesBySource {
		profilesBySource[string(source)] = count
	}

	response := ProfileStatsResponse{
		TotalProfiles:        stats.TotalProfiles,
		ProfilesWithFullName: stats.ProfilesWithFullName,
		ProfilesBySource:     profilesBySource,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// extractUserIDFromPath извлекает user_id из пути URL
func extractUserIDFromPath(path string) string {
	// Ожидаем путь вида /api/v1/profiles/{user_id} или /api/v1/profiles/{user_id}/name
	parts := strings.Split(strings.Trim(path, "/"), "/")
	
	// Ищем "profiles" и берем следующий элемент
	for i, part := range parts {
		if part == "profiles" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	
	return ""
}

// extractChatIDFromPath извлекает chat_id из пути URL используя mux.Vars
func extractChatIDFromPath(r *http.Request) string {
	vars := mux.Vars(r)
	chatID := vars["chat_id"]
	log.Printf("[DEBUG] mux.Vars: %+v, chat_id: '%s'", vars, chatID)
	return chatID
}

// getRequestID extracts request ID from context, returns empty string if not found
func getRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	return ""
}

// GetWebhookStats godoc
// @Summary Get webhook processing statistics
// @Description Get statistics about webhook event processing
// @Tags Monitoring
// @Accept json
// @Produce json
// @Param period query string false "Time period (hour, day, week, month)" default:"day"
// @Success 200 {object} WebhookStatsResponse "Webhook statistics"
// @Failure 400 {object} ErrorResponse "Invalid period parameter"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /monitoring/webhook/stats [get]
func (h *MaxBotHTTPHandler) GetWebhookStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Получаем параметр периода
	periodParam := r.URL.Query().Get("period")
	if periodParam == "" {
		periodParam = "day"
	}

	// Определяем временной период
	var period domain.TimePeriod
	switch periodParam {
	case "hour":
		period = domain.LastHour()
	case "day":
		period = domain.LastDay()
	case "week":
		period = domain.LastWeek()
	case "month":
		period = domain.LastMonth()
	default:
		errors.WriteError(w, errors.ValidationError("Invalid period parameter. Use: hour, day, week, month"), requestID)
		return
	}

	// Получаем статистику
	stats, err := h.monitoring.GetWebhookStats(ctx, period)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	response := WebhookStatsResponse{
		Period: TimePeriodResponse{
			From: stats.Period.From.Format(time.RFC3339),
			To:   stats.Period.To.Format(time.RFC3339),
		},
		TotalEvents:           stats.TotalEvents,
		SuccessfulEvents:      stats.SuccessfulEvents,
		FailedEvents:          stats.FailedEvents,
		EventsByType:          stats.EventsByType,
		ProfilesExtracted:     stats.ProfilesExtracted,
		ProfilesStored:        stats.ProfilesStored,
		AverageProcessingTime: stats.AverageProcessingTime,
		ErrorsByType:          stats.ErrorsByType,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// GetProfileCoverage godoc
// @Summary Get profile coverage metrics
// @Description Get metrics about profile data coverage
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} ProfileCoverageResponse "Profile coverage metrics"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /monitoring/profiles/coverage [get]
func (h *MaxBotHTTPHandler) GetProfileCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Получаем метрики покрытия
	coverage, err := h.monitoring.GetProfileCoverage(ctx)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	profilesBySource := make(map[string]int64)
	for source, count := range coverage.ProfilesBySource {
		profilesBySource[string(source)] = count
	}

	response := ProfileCoverageResponse{
		TotalUsers:         coverage.TotalUsers,
		UsersWithProfiles:  coverage.UsersWithProfiles,
		UsersWithFullNames: coverage.UsersWithFullNames,
		CoveragePercentage: coverage.CoveragePercentage,
		FullNamePercentage: coverage.FullNamePercentage,
		ProfilesBySource:   profilesBySource,
		LastUpdated:        coverage.LastUpdated.Format(time.RFC3339),
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}

// GetProfileQualityReport godoc
// @Summary Get profile quality report
// @Description Get detailed report about profile data quality
// @Tags Monitoring
// @Accept json
// @Produce json
// @Success 200 {object} ProfileQualityReportResponse "Profile quality report"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /monitoring/profiles/quality [get]
func (h *MaxBotHTTPHandler) GetProfileQualityReport(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Получаем отчет о качестве
	report, err := h.monitoring.GetProfileQualityReport(ctx)
	if err != nil {
		errors.WriteError(w, err, requestID)
		return
	}

	// Формируем ответ
	sourceBreakdown := make(map[string]SourceQualityResponse)
	for source, quality := range report.SourceBreakdown {
		sourceBreakdown[string(source)] = SourceQualityResponse{
			Count:            quality.Count,
			CompleteProfiles: quality.CompleteProfiles,
			AverageAge:       quality.AverageAge,
			QualityScore:     quality.QualityScore,
		}
	}

	dataIssues := make([]ProfileDataIssueResponse, len(report.DataIssues))
	for i, issue := range report.DataIssues {
		dataIssues[i] = ProfileDataIssueResponse{
			Type:        issue.Type,
			Description: issue.Description,
			Count:       issue.Count,
			Severity:    issue.Severity,
		}
	}

	response := ProfileQualityReportResponse{
		GeneratedAt:   report.GeneratedAt.Format(time.RFC3339),
		TotalProfiles: report.TotalProfiles,
		QualityMetrics: ProfileQualityMetricsResponse{
			CompleteProfiles:  report.QualityMetrics.CompleteProfiles,
			PartialProfiles:   report.QualityMetrics.PartialProfiles,
			EmptyProfiles:     report.QualityMetrics.EmptyProfiles,
			StaleProfiles:     report.QualityMetrics.StaleProfiles,
			QualityScore:      report.QualityMetrics.QualityScore,
			CompletenessScore: report.QualityMetrics.CompletenessScore,
			FreshnessScore:    report.QualityMetrics.FreshnessScore,
		},
		SourceBreakdown:    sourceBreakdown,
		RecommendedActions: report.RecommendedActions,
		DataIssues:         dataIssues,
	}

	// Отправляем ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		errors.WriteError(w, errors.InternalError("Failed to encode response", err), requestID)
		return
	}
}