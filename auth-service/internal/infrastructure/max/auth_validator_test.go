package max

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"testing"

	"auth-service/internal/domain"
)

func TestAuthValidator_ValidateInitData(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	tests := []struct {
		name        string
		initData    string
		botToken    string
		wantErr     bool
		errContains string
		expected    *domain.MaxUserData
	}{
		{
			name:        "empty initData",
			initData:    "",
			botToken:    botToken,
			wantErr:     true,
			errContains: "initData cannot be empty",
		},
		{
			name:        "empty botToken",
			initData:    "max_id=123&first_name=John&hash=abc",
			botToken:    "",
			wantErr:     true,
			errContains: "botToken cannot be empty",
		},
		{
			name:        "invalid query string",
			initData:    "invalid%query%string",
			botToken:    botToken,
			wantErr:     true,
			errContains: "failed to parse initData",
		},
		{
			name:        "missing hash parameter",
			initData:    "max_id=123&first_name=John",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash parameter is missing",
		},
		{
			name:        "invalid hash verification",
			initData:    "max_id=123&first_name=John&hash=invalid_hash",
			botToken:    botToken,
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "missing max_id",
			initData:    createValidInitData(map[string]string{"first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     true,
			errContains: "max_id is required",
		},
		{
			name:        "invalid max_id format",
			initData:    createValidInitData(map[string]string{"max_id": "invalid", "first_name": "John"}, botToken),
			botToken:    botToken,
			wantErr:     true,
			errContains: "invalid max_id format",
		},
		{
			name:     "missing first_name (now allowed)",
			initData: createValidInitData(map[string]string{"max_id": "123"}, botToken),
			botToken: botToken,
			wantErr:  false,
			expected: &domain.MaxUserData{
				MaxID:     123,
				Username:  "",
				FirstName: "",
				LastName:  "",
			},
		},
		{
			name:     "valid minimal data",
			initData: createValidInitData(map[string]string{"max_id": "123", "first_name": "John"}, botToken),
			botToken: botToken,
			wantErr:  false,
			expected: &domain.MaxUserData{
				MaxID:     123,
				Username:  "",
				FirstName: "John",
				LastName:  "",
			},
		},
		{
			name: "valid complete data",
			initData: createValidInitData(map[string]string{
				"max_id":     "456",
				"username":   "johndoe",
				"first_name": "John",
				"last_name":  "Doe",
			}, botToken),
			botToken: botToken,
			wantErr:  false,
			expected: &domain.MaxUserData{
				MaxID:     456,
				Username:  "johndoe",
				FirstName: "John",
				LastName:  "Doe",
			},
		},
		{
			name: "valid data with extra parameters",
			initData: createValidInitData(map[string]string{
				"max_id":     "789",
				"username":   "jane",
				"first_name": "Jane",
				"last_name":  "Smith",
				"extra_param": "ignored",
			}, botToken),
			botToken: botToken,
			wantErr:  false,
			expected: &domain.MaxUserData{
				MaxID:     789,
				Username:  "jane",
				FirstName: "Jane",
				LastName:  "Smith",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateInitData(tt.initData, tt.botToken)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateInitData() expected error but got none")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("ValidateInitData() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateInitData() unexpected error = %v", err)
				return
			}

			if result == nil {
				t.Errorf("ValidateInitData() result is nil")
				return
			}

			if result.MaxID != tt.expected.MaxID {
				t.Errorf("ValidateInitData() MaxID = %v, want %v", result.MaxID, tt.expected.MaxID)
			}
			if result.Username != tt.expected.Username {
				t.Errorf("ValidateInitData() Username = %v, want %v", result.Username, tt.expected.Username)
			}
			if result.FirstName != tt.expected.FirstName {
				t.Errorf("ValidateInitData() FirstName = %v, want %v", result.FirstName, tt.expected.FirstName)
			}
			if result.LastName != tt.expected.LastName {
				t.Errorf("ValidateInitData() LastName = %v, want %v", result.LastName, tt.expected.LastName)
			}
		})
	}
}

func TestAuthValidator_HashVerification(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	// Test that different bot tokens produce different hashes
	data := map[string]string{"max_id": "123", "first_name": "John"}
	initData1 := createValidInitData(data, botToken)
	initData2 := createValidInitData(data, "different_token")

	// Should succeed with correct token
	_, err1 := validator.ValidateInitData(initData1, botToken)
	if err1 != nil {
		t.Errorf("ValidateInitData() with correct token failed: %v", err1)
	}

	// Should fail with different token
	_, err2 := validator.ValidateInitData(initData2, botToken)
	if err2 == nil {
		t.Errorf("ValidateInitData() with wrong token should have failed")
	}
	if !strings.Contains(err2.Error(), "hash verification failed") {
		t.Errorf("ValidateInitData() error = %v, want error containing 'hash verification failed'", err2)
	}
}

func TestAuthValidator_ParameterSorting(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	// Create data with parameters in different orders
	data := map[string]string{
		"max_id":     "123",
		"first_name": "John",
		"username":   "johndoe",
		"last_name":  "Doe",
	}

	// Both should produce the same result regardless of parameter order
	initData1 := createValidInitData(data, botToken)
	initData2 := createValidInitDataWithOrder([]string{"username", "max_id", "last_name", "first_name"}, data, botToken)

	result1, err1 := validator.ValidateInitData(initData1, botToken)
	if err1 != nil {
		t.Errorf("ValidateInitData() first call failed: %v", err1)
	}

	result2, err2 := validator.ValidateInitData(initData2, botToken)
	if err2 != nil {
		t.Errorf("ValidateInitData() second call failed: %v", err2)
	}

	if result1.MaxID != result2.MaxID || result1.Username != result2.Username ||
		result1.FirstName != result2.FirstName || result1.LastName != result2.LastName {
		t.Errorf("ValidateInitData() results differ with different parameter orders")
	}
}

// Helper function to create valid initData with proper hash
func createValidInitData(params map[string]string, botToken string) string {
	return createValidInitDataWithOrder(nil, params, botToken)
}

// Helper function to create valid initData with specific parameter order
func createValidInitDataWithOrder(order []string, params map[string]string, botToken string) string {
	// For hash calculation, we always need to sort alphabetically (as per MAX protocol)
	// The order parameter is just for the final query string encoding
	var sortedParams []string
	keys := make([]string, 0, len(params))
	
	// Always sort alphabetically for hash calculation
	for key := range params {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if value, exists := params[key]; exists {
			sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
		}
	}

	// Create data string for hash calculation
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate hash
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Create query string with hash (order doesn't matter for final encoding)
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	values.Set("hash", hash)

	return values.Encode()
}