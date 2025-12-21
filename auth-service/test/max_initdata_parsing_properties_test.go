package test

import (
	"auth-service/internal/infrastructure/max"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty2_InitDataParsingConsistency tests that initData parsing works correctly
// **Feature: max-miniapp-auth, Property 2: InitData parsing consistency**
// **Validates: Requirements 2.1, 2.2, 2.3**
func TestProperty2_InitDataParsingConsistency(t *testing.T) {
	validator := max.NewAuthValidator()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test 1: Valid initData with proper hash should parse successfully
	properties.Property("valid initData with proper hash should parse successfully", prop.ForAll(
		func(maxID int64, usernameSeed int, firstNameSeed int, lastNameSeed int, botTokenSeed int) bool {
			// Generate test data
			username := "user" + strconv.Itoa(usernameSeed%1000000)
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			lastName := "Last" + strconv.Itoa(lastNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Create valid initData with proper hash
			initData := createValidInitData(maxID, username, firstName, lastName, botToken)

			// Parse the initData
			userData, err := validator.ValidateInitData(initData, botToken)
			if err != nil {
				t.Logf("Failed to parse valid initData: %v", err)
				return false
			}

			// Verify extracted data matches input
			if userData.MaxID != maxID {
				t.Logf("MaxID mismatch: expected %d, got %d", maxID, userData.MaxID)
				return false
			}

			if userData.Username != username {
				t.Logf("Username mismatch: expected %s, got %s", username, userData.Username)
				return false
			}

			if userData.FirstName != firstName {
				t.Logf("FirstName mismatch: expected %s, got %s", firstName, userData.FirstName)
				return false
			}

			if userData.LastName != lastName {
				t.Logf("LastName mismatch: expected %s, got %s", lastName, userData.LastName)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // username seed
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // lastName seed
		gen.IntRange(100000, 999999),    // botToken seed
	))

	// Test 2: InitData without optional fields should parse successfully
	properties.Property("initData without optional fields should parse successfully", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Create valid initData without optional fields (username, lastName)
			initData := createValidInitDataMinimal(maxID, firstName, botToken)

			// Parse the initData
			userData, err := validator.ValidateInitData(initData, botToken)
			if err != nil {
				t.Logf("Failed to parse minimal initData: %v", err)
				return false
			}

			// Verify extracted data
			if userData.MaxID != maxID {
				t.Logf("MaxID mismatch: expected %d, got %d", maxID, userData.MaxID)
				return false
			}

			if userData.FirstName != firstName {
				t.Logf("FirstName mismatch: expected %s, got %s", firstName, userData.FirstName)
				return false
			}

			// Optional fields should be empty
			if userData.Username != "" {
				t.Logf("Username should be empty, got %s", userData.Username)
				return false
			}

			if userData.LastName != "" {
				t.Logf("LastName should be empty, got %s", userData.LastName)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
	))

	// Test 3: InitData with wrong hash should fail validation
	properties.Property("initData with wrong hash should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int, wrongTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)
			wrongBotToken := "wrong_token_" + strconv.Itoa(wrongTokenSeed%1000000)

			// Ensure tokens are different
			if botToken == wrongBotToken {
				return true // Skip this test case
			}

			// Create initData with correct hash for botToken
			initData := createValidInitDataMinimal(maxID, firstName, botToken)

			// Try to validate with wrong bot token
			_, err := validator.ValidateInitData(initData, wrongBotToken)
			if err == nil {
				t.Logf("Expected validation to fail with wrong bot token")
				return false
			}

			// Should contain hash verification error
			if !strings.Contains(err.Error(), "hash verification failed") {
				t.Logf("Expected hash verification error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
		gen.IntRange(100000, 999999),    // wrongToken seed
	))

	// Test 4: InitData missing required fields should fail
	properties.Property("initData missing required fields should fail", prop.ForAll(
		func(botTokenSeed int) bool {
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Test cases for missing required fields
			testCases := []struct {
				name     string
				initData string
			}{
				{"missing max_id", createInitDataWithHash("first_name=Test", botToken)},
				{"missing first_name", createInitDataWithHash("max_id=123", botToken)},
				{"empty initData", ""},
			}

			for _, tc := range testCases {
				_, err := validator.ValidateInitData(tc.initData, botToken)
				if err == nil {
					t.Logf("Expected validation to fail for %s", tc.name)
					return false
				}
			}

			return true
		},
		gen.IntRange(100000, 999999), // botToken seed
	))

	// Test 5: Parameter sorting should be consistent regardless of input order
	properties.Property("parameter sorting should be consistent regardless of input order", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Create initData with parameters in different orders
			params1 := fmt.Sprintf("max_id=%d&first_name=%s", maxID, firstName)
			params2 := fmt.Sprintf("first_name=%s&max_id=%d", firstName, maxID)

			initData1 := createInitDataWithHash(params1, botToken)
			initData2 := createInitDataWithHash(params2, botToken)

			// Both should parse successfully
			userData1, err1 := validator.ValidateInitData(initData1, botToken)
			userData2, err2 := validator.ValidateInitData(initData2, botToken)

			if err1 != nil {
				t.Logf("Failed to parse initData1: %v", err1)
				return false
			}

			if err2 != nil {
				t.Logf("Failed to parse initData2: %v", err2)
				return false
			}

			// Results should be identical
			if userData1.MaxID != userData2.MaxID {
				t.Logf("MaxID mismatch between different parameter orders")
				return false
			}

			if userData1.FirstName != userData2.FirstName {
				t.Logf("FirstName mismatch between different parameter orders")
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
	))

	properties.TestingRun(t)
}

// Helper function to create valid initData with proper hash
func createValidInitData(maxID int64, username, firstName, lastName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&username=%s&first_name=%s&last_name=%s",
		maxID, url.QueryEscape(username), url.QueryEscape(firstName), url.QueryEscape(lastName))
	return createInitDataWithHash(params, botToken)
}

// Helper function to create minimal valid initData with proper hash
func createValidInitDataMinimal(maxID int64, firstName, botToken string) string {
	params := fmt.Sprintf("max_id=%d&first_name=%s", maxID, url.QueryEscape(firstName))
	return createInitDataWithHash(params, botToken)
}

// Helper function to create initData with proper hash
func createInitDataWithHash(params, botToken string) string {
	// Parse parameters to sort them
	values, _ := url.ParseQuery(params)
	
	var sortedParams []string
	for key := range values {
		value := values.Get(key)
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedParams)

	// Create data string for hash calculation
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate hash using HMAC-SHA256
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Return initData with hash
	return params + "&hash=" + hash
}