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

// TestProperty1_HashValidationCorrectness tests that hash validation works correctly
// **Feature: max-miniapp-auth, Property 1: Hash validation correctness**
// **Validates: Requirements 1.1, 2.4, 2.6**
func TestProperty1_HashValidationCorrectness(t *testing.T) {
	validator := max.NewAuthValidator()

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Test 1: InitData signed with correct bot token should validate successfully
	properties.Property("initData signed with correct bot token should validate successfully", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int, extraParamSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)
			extraParam := "extra_value_" + strconv.Itoa(extraParamSeed%1000000)

			// Create initData with various parameters and proper hash
			params := fmt.Sprintf("max_id=%d&first_name=%s&extra_param=%s",
				maxID, url.QueryEscape(firstName), url.QueryEscape(extraParam))
			initData := createInitDataWithCorrectHash(params, botToken)

			// Validation should succeed
			userData, err := validator.ValidateInitData(initData, botToken)
			if err != nil {
				t.Logf("Hash validation failed for correctly signed initData: %v", err)
				return false
			}

			// Verify data was extracted correctly
			if userData.MaxID != maxID {
				t.Logf("MaxID mismatch: expected %d, got %d", maxID, userData.MaxID)
				return false
			}

			if userData.FirstName != firstName {
				t.Logf("FirstName mismatch: expected %s, got %s", firstName, userData.FirstName)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
		gen.IntRange(100000, 999999),    // extraParam seed
	))

	// Test 2: InitData signed with wrong bot token should fail validation
	properties.Property("initData signed with wrong bot token should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int, correctTokenSeed int, wrongTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			correctToken := "correct_token_" + strconv.Itoa(correctTokenSeed%1000000)
			wrongToken := "wrong_token_" + strconv.Itoa(wrongTokenSeed%1000000)

			// Ensure tokens are different
			if correctToken == wrongToken {
				return true // Skip this test case
			}

			// Create initData signed with correct token
			params := fmt.Sprintf("max_id=%d&first_name=%s", maxID, url.QueryEscape(firstName))
			initData := createInitDataWithCorrectHash(params, correctToken)

			// Try to validate with wrong token
			_, err := validator.ValidateInitData(initData, wrongToken)
			if err == nil {
				t.Logf("Expected hash validation to fail with wrong bot token")
				return false
			}

			// Should be a hash verification error
			if !strings.Contains(err.Error(), "hash verification failed") {
				t.Logf("Expected hash verification error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // correctToken seed
		gen.IntRange(100000, 999999),    // wrongToken seed
	))

	// Test 3: InitData with tampered parameters should fail validation
	properties.Property("initData with tampered parameters should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int, tamperedValueSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)
			tamperedValue := "tampered_" + strconv.Itoa(tamperedValueSeed%1000000)

			// Create valid initData
			params := fmt.Sprintf("max_id=%d&first_name=%s", maxID, url.QueryEscape(firstName))
			initData := createInitDataWithCorrectHash(params, botToken)

			// Tamper with the first_name parameter while keeping the original hash
			tamperedInitData := strings.Replace(initData, url.QueryEscape(firstName), url.QueryEscape(tamperedValue), 1)

			// Validation should fail due to hash mismatch
			_, err := validator.ValidateInitData(tamperedInitData, botToken)
			if err == nil {
				t.Logf("Expected validation to fail for tampered initData")
				return false
			}

			// Should be a hash verification error
			if !strings.Contains(err.Error(), "hash verification failed") {
				t.Logf("Expected hash verification error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
		gen.IntRange(100000, 999999),    // tamperedValue seed
	))

	// Test 4: InitData with missing hash should fail validation
	properties.Property("initData with missing hash should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Create initData without hash parameter
			initDataWithoutHash := fmt.Sprintf("max_id=%d&first_name=%s", maxID, url.QueryEscape(firstName))

			// Validation should fail
			_, err := validator.ValidateInitData(initDataWithoutHash, botToken)
			if err == nil {
				t.Logf("Expected validation to fail for initData without hash")
				return false
			}

			// Should be a missing hash error
			if !strings.Contains(err.Error(), "hash parameter is missing") {
				t.Logf("Expected missing hash error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
	))

	// Test 5: InitData with invalid hash format should fail validation
	properties.Property("initData with invalid hash format should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int, invalidHashSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)
			invalidHash := "invalid_hash_" + strconv.Itoa(invalidHashSeed%1000000)

			// Create initData with invalid hash
			initDataWithInvalidHash := fmt.Sprintf("max_id=%d&first_name=%s&hash=%s",
				maxID, url.QueryEscape(firstName), invalidHash)

			// Validation should fail
			_, err := validator.ValidateInitData(initDataWithInvalidHash, botToken)
			if err == nil {
				t.Logf("Expected validation to fail for initData with invalid hash")
				return false
			}

			// Should be a hash verification error
			if !strings.Contains(err.Error(), "hash verification failed") {
				t.Logf("Expected hash verification error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
		gen.IntRange(100000, 999999),    // invalidHash seed
	))

	// Test 6: Empty bot token should fail validation
	properties.Property("empty bot token should fail validation", prop.ForAll(
		func(maxID int64, firstNameSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)

			// Create any initData (doesn't matter since bot token is empty)
			initData := fmt.Sprintf("max_id=%d&first_name=%s&hash=anyhash", maxID, url.QueryEscape(firstName))

			// Validation should fail with empty bot token
			_, err := validator.ValidateInitData(initData, "")
			if err == nil {
				t.Logf("Expected validation to fail for empty bot token")
				return false
			}

			// Should be a bot token error
			if !strings.Contains(err.Error(), "botToken cannot be empty") {
				t.Logf("Expected bot token error, got: %v", err)
				return false
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
	))

	// Test 7: Hash validation should be consistent for same input
	properties.Property("hash validation should be consistent for same input", prop.ForAll(
		func(maxID int64, firstNameSeed int, botTokenSeed int) bool {
			firstName := "First" + strconv.Itoa(firstNameSeed%1000000)
			botToken := "bot_token_" + strconv.Itoa(botTokenSeed%1000000)

			// Create valid initData
			params := fmt.Sprintf("max_id=%d&first_name=%s", maxID, url.QueryEscape(firstName))
			initData := createInitDataWithCorrectHash(params, botToken)

			// Validate multiple times - should always succeed
			for i := 0; i < 3; i++ {
				userData, err := validator.ValidateInitData(initData, botToken)
				if err != nil {
					t.Logf("Hash validation failed on attempt %d: %v", i+1, err)
					return false
				}

				// Results should be consistent
				if userData.MaxID != maxID {
					t.Logf("Inconsistent MaxID on attempt %d", i+1)
					return false
				}

				if userData.FirstName != firstName {
					t.Logf("Inconsistent FirstName on attempt %d", i+1)
					return false
				}
			}

			return true
		},
		gen.Int64Range(1, 999999999),    // maxID
		gen.IntRange(100000, 999999),    // firstName seed
		gen.IntRange(100000, 999999),    // botToken seed
	))

	properties.TestingRun(t)
}

// Helper function to create initData with correct hash using MAX Mini App algorithm
func createInitDataWithCorrectHash(params, botToken string) string {
	// Parse parameters to sort them alphabetically
	values, _ := url.ParseQuery(params)
	
	var sortedParams []string
	for key := range values {
		value := values.Get(key)
		sortedParams = append(sortedParams, fmt.Sprintf("%s=%s", key, value))
	}
	sort.Strings(sortedParams)

	// Create data string for hash calculation (newline separated)
	dataCheckString := strings.Join(sortedParams, "\n")

	// Calculate hash using HMAC-SHA256 with SHA256(botToken) as secret
	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(dataCheckString))
	hash := hex.EncodeToString(mac.Sum(nil))

	// Return initData with hash parameter
	return params + "&hash=" + hash
}