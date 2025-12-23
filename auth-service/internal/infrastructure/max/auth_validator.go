package max

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"auth-service/internal/domain"
)

// AuthValidator implements domain.MaxAuthValidator
type AuthValidator struct{}

// NewAuthValidator creates a new MaxAuthValidator implementation
func NewAuthValidator() domain.MaxAuthValidator {
	return &AuthValidator{}
}

// ValidateInitData validates MAX Mini App initData and extracts user information
func (v *AuthValidator) ValidateInitData(initData string, botToken string) (*domain.MaxUserData, error) {
	if initData == "" {
		return nil, fmt.Errorf("initData cannot be empty")
	}
	
	if botToken == "" {
		return nil, fmt.Errorf("botToken cannot be empty")
	}

	// Try to URL decode the entire initData string first
	// If it fails, assume it's already decoded (old format)
	decodedInitData, err := url.QueryUnescape(initData)
	if err != nil {
		// If URL decode fails, try to parse as-is (old format)
		decodedInitData = initData
	}

	// Parse query string into key-value pairs
	values, err := url.ParseQuery(decodedInitData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	// Extract and remove hash parameter
	hashValues := values["hash"]
	if len(hashValues) == 0 {
		return nil, fmt.Errorf("hash parameter is missing")
	}
	receivedHash := hashValues[0]
	values.Del("hash")

	// Remove WebApp specific parameters that are not part of the hash calculation
	values.Del("WebAppPlatform")
	values.Del("WebAppVersion")

	// Prepare data for verification by sorting parameters alphabetically
	var sortedParams []string
	for key := range values {
		// Get the first value for each key
		value := values.Get(key)
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedParams)

	// Create data string for hash verification
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate expected hash using HMAC-SHA256
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(receivedHash), []byte(expectedHash)) != 1 {
		return nil, fmt.Errorf("hash verification failed")
	}

	// Extract user data from validated parameters
	userData, err := v.extractUserData(values)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user data: %w", err)
	}

	return userData, nil
}

// extractUserData extracts MaxUserData from parsed query values
func (v *AuthValidator) extractUserData(values url.Values) (*domain.MaxUserData, error) {
	// Check if we have user data in JSON format (new format)
	userJSON := values.Get("user")
	if userJSON != "" {
		return v.extractUserDataFromJSON(userJSON)
	}

	// Fallback to old format for backward compatibility
	return v.extractUserDataFromParams(values)
}

// extractUserDataFromJSON extracts user data from JSON string (new format)
func (v *AuthValidator) extractUserDataFromJSON(userJSON string) (*domain.MaxUserData, error) {
	var user struct {
		ID           int64  `json:"id"`
		FirstName    string `json:"first_name"`
		LastName     string `json:"last_name"`
		Username     *string `json:"username"`
		LanguageCode string `json:"language_code"`
		PhotoURL     string `json:"photo_url"`
	}

	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to parse user JSON: %w", err)
	}

	if user.ID == 0 {
		return nil, fmt.Errorf("user id is required")
	}

	// first_name can be empty, no validation required

	username := ""
	if user.Username != nil {
		username = *user.Username
	}

	return &domain.MaxUserData{
		MaxID:     user.ID,
		Username:  username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
	}, nil
}

// extractUserDataFromParams extracts user data from URL parameters (old format)
func (v *AuthValidator) extractUserDataFromParams(values url.Values) (*domain.MaxUserData, error) {
	// Extract max_id (required)
	maxIDStr := values.Get("max_id")
	if maxIDStr == "" {
		return nil, fmt.Errorf("max_id is required")
	}

	maxID, err := strconv.ParseInt(maxIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid max_id format: %w", err)
	}

	// Extract first_name (can be empty)
	firstName := values.Get("first_name")

	// Extract optional fields
	username := values.Get("username")
	lastName := values.Get("last_name")

	return &domain.MaxUserData{
		MaxID:     maxID,
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
	}, nil
}