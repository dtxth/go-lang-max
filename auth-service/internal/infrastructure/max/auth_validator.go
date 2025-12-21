package max

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
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

	// Parse query string into key-value pairs
	values, err := url.ParseQuery(initData)
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
	// Extract max_id (required)
	maxIDStr := values.Get("max_id")
	if maxIDStr == "" {
		return nil, fmt.Errorf("max_id is required")
	}

	maxID, err := strconv.ParseInt(maxIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid max_id format: %w", err)
	}

	// Extract first_name (required)
	firstName := values.Get("first_name")
	if firstName == "" {
		return nil, fmt.Errorf("first_name is required")
	}

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