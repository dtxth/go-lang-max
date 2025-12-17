package config

import (
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/assert"
)

// **Feature: participants-background-sync, Property 4: Configuration validation and defaults**
// **Validates: Requirements 3.1, 3.2, 3.3, 3.4**
func TestProperty_ConfigurationValidationAndDefaults(t *testing.T) {
	properties := gopter.NewProperties(nil)
	
	// Property: For any configuration input (valid or invalid), the system should either 
	// apply the configuration correctly or use sensible defaults with appropriate logging
	properties.Property("configuration validation and defaults", prop.ForAll(
		func(configInputs map[string]string) bool {
			// Save original environment
			originalEnv := make(map[string]string)
			envVars := []string{
				"PARTICIPANTS_CACHE_TTL",
				"PARTICIPANTS_UPDATE_INTERVAL", 
				"PARTICIPANTS_FULL_UPDATE_HOUR",
				"PARTICIPANTS_BATCH_SIZE",
				"PARTICIPANTS_MAX_API_TIMEOUT",
				"PARTICIPANTS_STALE_THRESHOLD",
				"PARTICIPANTS_ENABLE_BACKGROUND_SYNC",
				"PARTICIPANTS_ENABLE_LAZY_UPDATE",
				"PARTICIPANTS_MAX_RETRIES",
			}
			
			for _, envVar := range envVars {
				if val, exists := os.LookupEnv(envVar); exists {
					originalEnv[envVar] = val
				}
				os.Unsetenv(envVar)
			}
			
			// Set test environment variables
			for key, value := range configInputs {
				if contains(envVars, key) {
					os.Setenv(key, value)
				}
			}
			
			// Load configuration
			config := LoadParticipantsConfig()
			
			// Restore original environment
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
				if val, exists := originalEnv[envVar]; exists {
					os.Setenv(envVar, val)
				}
			}
			
			// Verify configuration is never nil
			if config == nil {
				return false
			}
			
			// Verify all values are within valid ranges (sensible defaults or valid parsed values)
			if !isValidDuration(config.CacheTTL, 1*time.Minute, 24*time.Hour) {
				return false
			}
			
			if !isValidDuration(config.UpdateInterval, 1*time.Minute, 24*time.Hour) {
				return false
			}
			
			if config.FullUpdateHour < 0 || config.FullUpdateHour > 23 {
				return false
			}
			
			if config.BatchSize < 1 || config.BatchSize > 1000 {
				return false
			}
			
			if !isValidDuration(config.MaxAPITimeout, 1*time.Second, 5*time.Minute) {
				return false
			}
			
			if !isValidDuration(config.StaleThreshold, 1*time.Minute, 24*time.Hour) {
				return false
			}
			
			if config.MaxRetries < 0 || config.MaxRetries > 10 {
				return false
			}
			
			// Verify that invalid inputs result in default values
			for key, value := range configInputs {
				switch key {
				case "PARTICIPANTS_CACHE_TTL":
					if !isValidDurationString(value) {
						// Invalid duration should result in default
						if config.CacheTTL != ParticipantsConfigDefaults.CacheTTL {
							return false
						}
					}
				case "PARTICIPANTS_BATCH_SIZE":
					if intVal, err := strconv.Atoi(value); err != nil || intVal < 1 || intVal > 1000 {
						// Invalid batch size should result in default
						if config.BatchSize != ParticipantsConfigDefaults.BatchSize {
							return false
						}
					}
				case "PARTICIPANTS_FULL_UPDATE_HOUR":
					if intVal, err := strconv.Atoi(value); err != nil || intVal < 0 || intVal > 23 {
						// Invalid hour should result in default
						if config.FullUpdateHour != ParticipantsConfigDefaults.FullUpdateHour {
							return false
						}
					}
				}
			}
			
			return true
		},
		genConfigInputs(),
	))
	
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test property for main configuration validation
func TestProperty_MainConfigurationValidation(t *testing.T) {
	properties := gopter.NewProperties(nil)
	
	properties.Property("main configuration validation", prop.ForAll(
		func(port, grpcPort, redisURL string) bool {
			// Save original environment
			originalEnv := make(map[string]string)
			envVars := []string{"PORT", "GRPC_PORT", "REDIS_URL", "DATABASE_URL"}
			
			for _, envVar := range envVars {
				if val, exists := os.LookupEnv(envVar); exists {
					originalEnv[envVar] = val
				}
				os.Unsetenv(envVar)
			}
			
			// Set required DATABASE_URL to avoid panic
			os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
			
			// Set test values
			os.Setenv("PORT", port)
			os.Setenv("GRPC_PORT", grpcPort)
			os.Setenv("REDIS_URL", redisURL)
			
			// Load configuration (should not panic)
			config := Load()
			
			// Restore environment
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
				if val, exists := originalEnv[envVar]; exists {
					os.Setenv(envVar, val)
				}
			}
			
			// Verify configuration is valid
			if config == nil {
				return false
			}
			
			// Port should be valid or default
			if portNum, err := strconv.Atoi(config.Port); err != nil || portNum < 1 || portNum > 65535 {
				return false
			}
			
			// GRPC Port should be valid or default
			if grpcPortNum, err := strconv.Atoi(config.GRPCPort); err != nil || grpcPortNum < 1 || grpcPortNum > 65535 {
				return false
			}
			
			// Redis URL should be valid or default
			if config.RedisURL == "" {
				return false
			}
			
			return true
		},
		gen.AlphaString(),
		gen.AlphaString(), 
		gen.AlphaString(),
	))
	
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test that defaults are always sensible
func TestProperty_DefaultsAreSensible(t *testing.T) {
	properties := gopter.NewProperties(nil)
	
	properties.Property("defaults are always sensible", prop.ForAll(
		func() bool {
			// Clear all environment variables
			envVars := []string{
				"PARTICIPANTS_CACHE_TTL",
				"PARTICIPANTS_UPDATE_INTERVAL", 
				"PARTICIPANTS_FULL_UPDATE_HOUR",
				"PARTICIPANTS_BATCH_SIZE",
				"PARTICIPANTS_MAX_API_TIMEOUT",
				"PARTICIPANTS_STALE_THRESHOLD",
				"PARTICIPANTS_ENABLE_BACKGROUND_SYNC",
				"PARTICIPANTS_ENABLE_LAZY_UPDATE",
				"PARTICIPANTS_MAX_RETRIES",
				"PARTICIPANTS_DISABLED",
			}
			
			originalEnv := make(map[string]string)
			for _, envVar := range envVars {
				if val, exists := os.LookupEnv(envVar); exists {
					originalEnv[envVar] = val
				}
				os.Unsetenv(envVar)
			}
			
			// Load configuration with no environment variables set
			config := LoadParticipantsConfig()
			
			// Restore environment
			for _, envVar := range envVars {
				if val, exists := originalEnv[envVar]; exists {
					os.Setenv(envVar, val)
				}
			}
			
			// Verify defaults match expected values
			expected := ParticipantsConfigDefaults
			
			return config.CacheTTL == expected.CacheTTL &&
				config.UpdateInterval == expected.UpdateInterval &&
				config.FullUpdateHour == expected.FullUpdateHour &&
				config.BatchSize == expected.BatchSize &&
				config.MaxAPITimeout == expected.MaxAPITimeout &&
				config.StaleThreshold == expected.StaleThreshold &&
				config.EnableBackgroundSync == expected.EnableBackgroundSync &&
				config.EnableLazyUpdate == expected.EnableLazyUpdate &&
				config.MaxRetries == expected.MaxRetries
		},
	))
	
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test participants integration disabled functionality
func TestProperty_ParticipantsIntegrationDisabled(t *testing.T) {
	properties := gopter.NewProperties(nil)
	
	properties.Property("participants integration disabled", prop.ForAll(
		func(disabledValue string) bool {
			// Save original environment
			originalEnv := make(map[string]string)
			envVars := []string{
				"PARTICIPANTS_DISABLED",
				"PARTICIPANTS_ENABLE_BACKGROUND_SYNC",
				"PARTICIPANTS_ENABLE_LAZY_UPDATE",
			}
			
			for _, envVar := range envVars {
				if val, exists := os.LookupEnv(envVar); exists {
					originalEnv[envVar] = val
				}
				os.Unsetenv(envVar)
			}
			
			// Set PARTICIPANTS_DISABLED
			os.Setenv("PARTICIPANTS_DISABLED", disabledValue)
			
			// Load configuration
			config := LoadParticipantsConfig()
			
			// Restore environment
			for _, envVar := range envVars {
				os.Unsetenv(envVar)
				if val, exists := originalEnv[envVar]; exists {
					os.Setenv(envVar, val)
				}
			}
			
			// Check if the value should be parsed as true using our flexible boolean parsing
			shouldBeDisabled := false
			// Replicate the logic from parseFlexibleBool for testing
			if boolVal, err := strconv.ParseBool(disabledValue); err == nil && boolVal {
				shouldBeDisabled = true
			} else {
				// Handle yes/no variants (case insensitive)
				switch strings.ToLower(strings.TrimSpace(disabledValue)) {
				case "yes", "y":
					shouldBeDisabled = true
				case "no", "n":
					shouldBeDisabled = false
				}
			}
			
			// If participants integration should be disabled, both sync options should be false
			if shouldBeDisabled {
				return !config.EnableBackgroundSync && !config.EnableLazyUpdate
			}
			
			// If not disabled, should use defaults
			return config.EnableBackgroundSync == ParticipantsConfigDefaults.EnableBackgroundSync &&
				config.EnableLazyUpdate == ParticipantsConfigDefaults.EnableLazyUpdate
		},
		gen.OneConstOf("true", "false", "1", "0", "yes", "no", "invalid", ""),
	))
	
	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Helper functions

func genConfigInputs() gopter.Gen {
	return gen.MapOf(
		gen.OneConstOf(
			"PARTICIPANTS_CACHE_TTL",
			"PARTICIPANTS_UPDATE_INTERVAL", 
			"PARTICIPANTS_FULL_UPDATE_HOUR",
			"PARTICIPANTS_BATCH_SIZE",
			"PARTICIPANTS_MAX_API_TIMEOUT",
			"PARTICIPANTS_STALE_THRESHOLD",
			"PARTICIPANTS_ENABLE_BACKGROUND_SYNC",
			"PARTICIPANTS_ENABLE_LAZY_UPDATE",
			"PARTICIPANTS_MAX_RETRIES",
		),
		gen.OneConstOf(
			"invalid_string",                     // Invalid strings
			"123",                                // Numeric strings
			"",                                   // Empty strings
			"true",                               // Valid booleans
			"false",
			"1h",                                 // Valid durations
			"30m",
			"invalid_duration",                   // Invalid durations
			"-1",                                 // Invalid numbers
			"999999",                             // Out of range numbers
		),
	)
}

func isValidDuration(d time.Duration, min, max time.Duration) bool {
	return d >= min && d <= max
}

func isValidDurationString(s string) bool {
	_, err := time.ParseDuration(s)
	return err == nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Unit tests for specific validation functions

func TestValidatePort(t *testing.T) {
	tests := []struct {
		port    string
		wantErr bool
	}{
		{"8080", false},
		{"1", false},
		{"65535", false},
		{"0", true},
		{"65536", true},
		{"abc", true},
		{"", true},
		{"-1", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			err := validatePort(tt.port)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateGRPCAddress(t *testing.T) {
	tests := []struct {
		addr    string
		wantErr bool
	}{
		{"localhost:9090", false},
		{"127.0.0.1:8080", false},
		{"service:9092", false},
		{"localhost", true},
		{"", true},
		{"localhost:abc", true},
		{"localhost:0", true},
		{"localhost:65536", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			err := validateGRPCAddress(tt.addr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRedisURL(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"redis://localhost:6379", false},
		{"redis://localhost:6379/0", false},
		{"rediss://secure.redis.com:6380", false},
		{"redis://user:pass@localhost:6379", false},
		{"http://localhost:6379", true},
		{"localhost:6379", true},
		{"", true},
		{"redis://", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			err := validateRedisURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateRedisConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		redisURL string
		wantErr  bool
	}{
		{"valid redis URL", "redis://localhost:6379", false},
		{"valid rediss URL", "rediss://secure.redis.com:6380", false},
		{"empty URL", "", true},
		{"invalid scheme", "http://localhost:6379", true},
		{"invalid format", "not-a-url", true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalRedisURL := os.Getenv("REDIS_URL")
			defer func() {
				if originalRedisURL != "" {
					os.Setenv("REDIS_URL", originalRedisURL)
				} else {
					os.Unsetenv("REDIS_URL")
				}
			}()
			
			// Set test value
			if tt.redisURL == "" {
				os.Unsetenv("REDIS_URL")
			} else {
				os.Setenv("REDIS_URL", tt.redisURL)
			}
			
			err := ValidateRedisConfiguration()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestIsParticipantsIntegrationDisabled(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"1", "1", true},
		{"0", "0", false},
		{"yes", "yes", true},
		{"no", "no", false},
		{"empty", "", false},
		{"invalid", "invalid", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original environment
			originalValue := os.Getenv("PARTICIPANTS_DISABLED")
			defer func() {
				if originalValue != "" {
					os.Setenv("PARTICIPANTS_DISABLED", originalValue)
				} else {
					os.Unsetenv("PARTICIPANTS_DISABLED")
				}
			}()
			
			// Set test value
			if tt.value == "" {
				os.Unsetenv("PARTICIPANTS_DISABLED")
			} else {
				os.Setenv("PARTICIPANTS_DISABLED", tt.value)
			}
			
			result := isParticipantsIntegrationDisabled()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateAllConfiguration(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	envVars := []string{
		"DATABASE_URL", "PORT", "GRPC_PORT", "AUTH_GRPC_ADDR", "MAXBOT_GRPC_ADDR",
		"REDIS_URL", "PARTICIPANTS_CACHE_TTL", "PARTICIPANTS_BATCH_SIZE",
	}
	
	for _, envVar := range envVars {
		if val, exists := os.LookupEnv(envVar); exists {
			originalEnv[envVar] = val
		}
		os.Unsetenv(envVar)
	}
	
	defer func() {
		for _, envVar := range envVars {
			os.Unsetenv(envVar)
			if val, exists := originalEnv[envVar]; exists {
				os.Setenv(envVar, val)
			}
		}
	}()
	
	// Test with valid configuration
	t.Run("valid configuration", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		os.Setenv("REDIS_URL", "redis://localhost:6379")
		os.Setenv("PORT", "8080")
		
		errors := ValidateAllConfiguration()
		assert.Empty(t, errors)
	})
	
	// Test with invalid configuration
	t.Run("invalid configuration", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		os.Setenv("REDIS_URL", "invalid-url")
		os.Setenv("PORT", "invalid-port")
		os.Setenv("PARTICIPANTS_CACHE_TTL", "invalid-duration")
		
		errors := ValidateAllConfiguration()
		assert.NotEmpty(t, errors)
		
		// Should have errors for invalid values
		hasRedisError := false
		hasPortError := false
		hasCacheTTLError := false
		
		for _, err := range errors {
			switch err.Field {
			case "REDIS_URL":
				hasRedisError = true
			case "PORT":
				hasPortError = true
			case "PARTICIPANTS_CACHE_TTL":
				hasCacheTTLError = true
			}
		}
		
		assert.True(t, hasRedisError, "Should have Redis URL validation error")
		assert.True(t, hasPortError, "Should have PORT validation error")
		assert.True(t, hasCacheTTLError, "Should have cache TTL validation error")
	})
	
	// Test with participants disabled
	t.Run("participants disabled", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")
		os.Setenv("PARTICIPANTS_DISABLED", "true")
		
		errors := ValidateAllConfiguration()
		
		// Should not have Redis validation error when participants are disabled
		hasRedisError := false
		for _, err := range errors {
			if err.Field == "REDIS_URL" {
				hasRedisError = true
			}
		}
		
		assert.False(t, hasRedisError, "Should not validate Redis when participants are disabled")
	})
}