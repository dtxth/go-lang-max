package max

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"strings"
	"testing"

	"auth-service/internal/domain"
)

func TestAuthValidator_ValidateInitData_NewFormat(t *testing.T) {
	validator := NewAuthValidator()
	botToken := "test_bot_token_123"

	// Test data similar to your example
	testInitData := `auth_date%3D1766484807%26hash%3D0a044bdc3320af8c9871bcf61a125f795bcb0385980bd295ee405b4a83a7b230%26chat%3D%257B%2522id%2522%253A123809879%252C%2522type%2522%253A%2522DIALOG%2522%257D%26ip%3D5.227.65.9%26user%3D%257B%2522id%2522%253A18963527%252C%2522first_name%2522%253A%2522%25D0%2590%25D0%25BD%25D0%25B4%25D1%2580%25D0%25B5%25D0%25B9%2522%252C%2522last_name%2522%253A%2522%2522%252C%2522username%2522%253Anull%252C%2522language_code%2522%253A%2522ru%2522%252C%2522photo_url%2522%253A%2522https%253A%252F%252Fi.oneme.ru%252Fi%253Fr%253DBTGBPUwtwgYUeoFhO7rESmr8zwrhZ_p4jWvgWajP77Kf1UYIInL726n0cb3v5ARlyCo%2522%257D%26query_id%3D40a5e9e4-3160-480d-b6dc-12b2b7fce31b&WebAppPlatform=web&WebAppVersion=25.12.13`

	// Create a valid test case with correct hash
	validInitData := createValidNewFormatInitData(botToken)

	tests := []struct {
		name        string
		initData    string
		botToken    string
		expected    *domain.MaxUserData
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid new format",
			initData: validInitData,
			botToken: botToken,
			expected: &domain.MaxUserData{
				MaxID:     18963527,
				Username:  "",
				FirstName: "Андрей",
				LastName:  "",
			},
			wantErr: false,
		},
		{
			name:     "valid new format with last name",
			initData: createValidNewFormatInitDataWithLastName(botToken),
			botToken: botToken,
			expected: &domain.MaxUserData{
				MaxID:     18963527,
				Username:  "testuser",
				FirstName: "Андрей",
				LastName:  "Иванов",
			},
			wantErr: false,
		},
		{
			name:     "valid new format with empty first name",
			initData: createValidNewFormatInitDataWithEmptyFirstName(botToken),
			botToken: botToken,
			expected: &domain.MaxUserData{
				MaxID:     18963527,
				Username:  "testuser",
				FirstName: "",
				LastName:  "Иванов",
			},
			wantErr: false,
		},
		{
			name:        "invalid hash in new format",
			initData:    testInitData,
			botToken:    "wrong_token",
			wantErr:     true,
			errContains: "hash verification failed",
		},
		{
			name:        "empty init data",
			initData:    "",
			botToken:    botToken,
			wantErr:     true,
			errContains: "initData cannot be empty",
		},
		{
			name:        "empty bot token",
			initData:    validInitData,
			botToken:    "",
			wantErr:     true,
			errContains: "botToken cannot be empty",
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

// createValidNewFormatInitData creates a valid initData string in new format with correct hash
func createValidNewFormatInitData(botToken string) string {
	// Create test data
	authDate := "1766484807"
	chatJSON := `{"id":123809879,"type":"DIALOG"}`
	ip := "5.227.65.9"
	userJSON := `{"id":18963527,"first_name":"Андрей","last_name":"","username":null,"language_code":"ru","photo_url":"https://i.oneme.ru/i?r=BTGBPUwtwgYUeoFhO7rESmr8zwrhZ_p4jWvgWajP77Kf1UYIInL726n0cb3v5ARlyCo"}`
	queryID := "40a5e9e4-3160-480d-b6dc-12b2b7fce31b"

	// Create the data string for hash calculation (sorted alphabetically)
	dataParams := []string{
		"auth_date=" + authDate,
		"chat=" + chatJSON,
		"ip=" + ip,
		"query_id=" + queryID,
		"user=" + userJSON,
	}

	// Calculate hash using the same logic as the validator
	dataCheckString := strings.Join(dataParams, "\n")
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Create the full initData string
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("hash", hash)
	params.Set("chat", chatJSON)
	params.Set("ip", ip)
	params.Set("user", userJSON)
	params.Set("query_id", queryID)

	// Add WebApp parameters
	fullParams := params.Encode() + "&WebAppPlatform=web&WebAppVersion=25.12.13"

	// URL encode the entire string
	return url.QueryEscape(fullParams)
}

// createValidNewFormatInitDataWithLastName creates a valid initData string with last name
func createValidNewFormatInitDataWithLastName(botToken string) string {
	// Create test data with last name
	authDate := "1766484807"
	chatJSON := `{"id":123809879,"type":"DIALOG"}`
	ip := "5.227.65.9"
	userJSON := `{"id":18963527,"first_name":"Андрей","last_name":"Иванов","username":"testuser","language_code":"ru","photo_url":"https://i.oneme.ru/i?r=BTGBPUwtwgYUeoFhO7rESmr8zwrhZ_p4jWvgWajP77Kf1UYIInL726n0cb3v5ARlyCo"}`
	queryID := "40a5e9e4-3160-480d-b6dc-12b2b7fce31b"

	// Create the data string for hash calculation (sorted alphabetically)
	dataParams := []string{
		"auth_date=" + authDate,
		"chat=" + chatJSON,
		"ip=" + ip,
		"query_id=" + queryID,
		"user=" + userJSON,
	}

	// Calculate hash using the same logic as the validator
	dataCheckString := strings.Join(dataParams, "\n")
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Create the full initData string
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("hash", hash)
	params.Set("chat", chatJSON)
	params.Set("ip", ip)
	params.Set("user", userJSON)
	params.Set("query_id", queryID)

	// Add WebApp parameters
	fullParams := params.Encode() + "&WebAppPlatform=web&WebAppVersion=25.12.13"

	// URL encode the entire string
	return url.QueryEscape(fullParams)
}

// createValidNewFormatInitDataWithEmptyFirstName creates a valid initData string with empty first name
func createValidNewFormatInitDataWithEmptyFirstName(botToken string) string {
	// Create test data with empty first name
	authDate := "1766484807"
	chatJSON := `{"id":123809879,"type":"DIALOG"}`
	ip := "5.227.65.9"
	userJSON := `{"id":18963527,"first_name":"","last_name":"Иванов","username":"testuser","language_code":"ru","photo_url":"https://i.oneme.ru/i?r=BTGBPUwtwgYUeoFhO7rESmr8zwrhZ_p4jWvgWajP77Kf1UYIInL726n0cb3v5ARlyCo"}`
	queryID := "40a5e9e4-3160-480d-b6dc-12b2b7fce31b"

	// Create the data string for hash calculation (sorted alphabetically)
	dataParams := []string{
		"auth_date=" + authDate,
		"chat=" + chatJSON,
		"ip=" + ip,
		"query_id=" + queryID,
		"user=" + userJSON,
	}

	// Calculate hash using the same logic as the validator
	dataCheckString := strings.Join(dataParams, "\n")
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Create the full initData string
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("hash", hash)
	params.Set("chat", chatJSON)
	params.Set("ip", ip)
	params.Set("user", userJSON)
	params.Set("query_id", queryID)

	// Add WebApp parameters
	fullParams := params.Encode() + "&WebAppPlatform=web&WebAppVersion=25.12.13"

	// URL encode the entire string
	return url.QueryEscape(fullParams)
}