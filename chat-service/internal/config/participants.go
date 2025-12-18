package config

import (
	"chat-service/internal/domain"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// ConfigValidationError represents a configuration validation error
type ConfigValidationError struct {
	Field   string
	Value   string
	Message string
}

func (e ConfigValidationError) Error() string {
	return fmt.Sprintf("invalid configuration for %s='%s': %s", e.Field, e.Value, e.Message)
}

// ParticipantsConfigDefaults contains default values for participants configuration
var ParticipantsConfigDefaults = domain.ParticipantsConfig{
	CacheTTL:              1 * time.Hour,
	UpdateInterval:        15 * time.Minute,
	FullUpdateHour:        3,
	BatchSize:             50,
	MaxAPITimeout:         30 * time.Second,
	StaleThreshold:        1 * time.Hour,
	EnableBackgroundSync:  true,
	EnableLazyUpdate:      true,
	MaxRetries:            3,
}

// LoadParticipantsConfig loads and validates participants configuration from environment variables
// Returns configuration with validated values or sensible defaults for invalid inputs
// Implements comprehensive validation as per Requirements 3.1, 3.2, 3.3, 3.4, 3.5
func LoadParticipantsConfig() *domain.ParticipantsConfig {
	log.Printf("Loading participants configuration from environment variables")
	
	config := ParticipantsConfigDefaults // Start with defaults
	
	// Check if participants integration is explicitly disabled
	if isParticipantsIntegrationDisabled() {
		log.Printf("Participants integration is disabled via PARTICIPANTS_DISABLED=true")
		config.EnableBackgroundSync = false
		config.EnableLazyUpdate = false
		return &config
	}
	
	// Load and validate each configuration parameter with enhanced validation
	config.CacheTTL = loadDurationWithValidation("PARTICIPANTS_CACHE_TTL", config.CacheTTL, 1*time.Minute, 24*time.Hour)
	config.UpdateInterval = loadDurationWithValidation("PARTICIPANTS_UPDATE_INTERVAL", config.UpdateInterval, 1*time.Minute, 24*time.Hour)
	config.FullUpdateHour = loadIntWithValidation("PARTICIPANTS_FULL_UPDATE_HOUR", config.FullUpdateHour, 0, 23)
	config.BatchSize = loadIntWithValidation("PARTICIPANTS_BATCH_SIZE", config.BatchSize, 1, 1000)
	config.MaxAPITimeout = loadDurationWithValidation("PARTICIPANTS_MAX_API_TIMEOUT", config.MaxAPITimeout, 1*time.Second, 5*time.Minute)
	config.StaleThreshold = loadDurationWithValidation("PARTICIPANTS_STALE_THRESHOLD", config.StaleThreshold, 1*time.Minute, 24*time.Hour)
	config.EnableBackgroundSync = loadBoolWithValidation("PARTICIPANTS_ENABLE_BACKGROUND_SYNC", config.EnableBackgroundSync)
	config.EnableLazyUpdate = loadBoolWithValidation("PARTICIPANTS_ENABLE_LAZY_UPDATE", config.EnableLazyUpdate)
	config.MaxRetries = loadIntWithValidation("PARTICIPANTS_MAX_RETRIES", config.MaxRetries, 0, 10)
	
	// Validate configuration consistency and log configuration summary
	validateConfigurationConsistency(&config)
	logConfigurationSummary(&config)
	
	return &config
}

// loadDurationWithValidation loads a duration from environment variable with validation
func loadDurationWithValidation(envVar string, defaultValue time.Duration, min, max time.Duration) time.Duration {
	val := os.Getenv(envVar)
	if val == "" {
		return defaultValue
	}
	
	duration, err := time.ParseDuration(val)
	if err != nil {
		logConfigWarning(ConfigValidationError{
			Field:   envVar,
			Value:   val,
			Message: fmt.Sprintf("invalid duration format, using default %v", defaultValue),
		})
		return defaultValue
	}
	
	if duration < min || duration > max {
		logConfigWarning(ConfigValidationError{
			Field:   envVar,
			Value:   val,
			Message: fmt.Sprintf("duration out of range [%v, %v], using default %v", min, max, defaultValue),
		})
		return defaultValue
	}
	
	return duration
}



// loadBoolWithValidation loads a boolean from environment variable with validation
func loadBoolWithValidation(envVar string, defaultValue bool) bool {
	val := os.Getenv(envVar)
	if val == "" {
		return defaultValue
	}
	
	boolVal, err := parseFlexibleBool(val)
	if err != nil {
		logConfigWarning(ConfigValidationError{
			Field:   envVar,
			Value:   val,
			Message: fmt.Sprintf("invalid boolean format, using default %t", defaultValue),
		})
		return defaultValue
	}
	
	return boolVal
}

// validateConfigurationConsistency validates that configuration values are consistent with each other
func validateConfigurationConsistency(config *domain.ParticipantsConfig) {
	// Ensure stale threshold is not smaller than cache TTL
	if config.StaleThreshold < config.CacheTTL {
		log.Printf("WARNING: PARTICIPANTS_STALE_THRESHOLD (%v) is smaller than PARTICIPANTS_CACHE_TTL (%v), this may cause frequent updates", 
			config.StaleThreshold, config.CacheTTL)
	}
	
	// Ensure update interval is reasonable compared to stale threshold
	if config.UpdateInterval > config.StaleThreshold {
		log.Printf("WARNING: PARTICIPANTS_UPDATE_INTERVAL (%v) is larger than PARTICIPANTS_STALE_THRESHOLD (%v), data may become stale between updates", 
			config.UpdateInterval, config.StaleThreshold)
	}
	
	// Warn if both background sync and lazy update are disabled
	if !config.EnableBackgroundSync && !config.EnableLazyUpdate {
		log.Printf("WARNING: Both PARTICIPANTS_ENABLE_BACKGROUND_SYNC and PARTICIPANTS_ENABLE_LAZY_UPDATE are disabled, participants count will not be updated automatically")
	}
}

// isParticipantsIntegrationDisabled checks if participants integration is explicitly disabled
// Requirement 3.5: WHEN participants integration is disabled THEN the system SHALL skip all participants-related initialization
func isParticipantsIntegrationDisabled() bool {
	val := os.Getenv("PARTICIPANTS_DISABLED")
	if val == "" {
		return false
	}
	
	disabled, err := parseFlexibleBool(val)
	if err != nil {
		logConfigWarning(ConfigValidationError{
			Field:   "PARTICIPANTS_DISABLED",
			Value:   val,
			Message: "invalid boolean format, treating as false",
		})
		return false
	}
	
	return disabled
}

// parseFlexibleBool parses boolean values with support for yes/no in addition to standard boolean formats
func parseFlexibleBool(val string) (bool, error) {
	// First try standard boolean parsing
	if boolVal, err := strconv.ParseBool(val); err == nil {
		return boolVal, nil
	}
	
	// Handle yes/no variants (case insensitive)
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "yes", "y":
		return true, nil
	case "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value: %s", val)
	}
}

// logConfigurationSummary logs a summary of the loaded configuration for monitoring
// Requirement 3.4: Enhanced logging for configuration validation
func logConfigurationSummary(config *domain.ParticipantsConfig) {
	log.Printf("Participants configuration loaded successfully:")
	log.Printf("  Cache TTL: %v", config.CacheTTL)
	log.Printf("  Update Interval: %v", config.UpdateInterval)
	log.Printf("  Full Update Hour: %d", config.FullUpdateHour)
	log.Printf("  Batch Size: %d", config.BatchSize)
	log.Printf("  MAX API Timeout: %v", config.MaxAPITimeout)
	log.Printf("  Stale Threshold: %v", config.StaleThreshold)
	log.Printf("  Background Sync Enabled: %t", config.EnableBackgroundSync)
	log.Printf("  Lazy Update Enabled: %t", config.EnableLazyUpdate)
	log.Printf("  Max Retries: %d", config.MaxRetries)
}

// validateRedisConfiguration validates Redis URL configuration specifically for participants
// Requirement 3.1: WHEN configuring the system THEN the system SHALL support Redis URL configuration via environment variables
func ValidateRedisConfiguration() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return fmt.Errorf("REDIS_URL is required for participants integration")
	}
	
	if err := validateRedisURL(redisURL); err != nil {
		return fmt.Errorf("invalid REDIS_URL configuration: %w", err)
	}
	
	log.Printf("Redis configuration validated successfully: %s", redisURL)
	return nil
}

// ValidateAllConfiguration performs comprehensive validation of all configuration parameters
// This function can be called during application startup to validate the entire configuration
func ValidateAllConfiguration() []ConfigValidationError {
	var errors []ConfigValidationError
	
	// Validate main configuration
	if err := validateMainConfigurationParameters(); err != nil {
		errors = append(errors, err...)
	}
	
	// Validate participants configuration
	if err := validateParticipantsConfigurationParameters(); err != nil {
		errors = append(errors, err...)
	}
	
	// Validate Redis configuration if participants integration is enabled
	if !isParticipantsIntegrationDisabled() {
		if err := ValidateRedisConfiguration(); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_URL",
				Value:   os.Getenv("REDIS_URL"),
				Message: err.Error(),
			})
		}
		
		// Validate Redis reconnection parameters
		if err := validateRedisReconnectionParameters(); err != nil {
			errors = append(errors, err...)
		}
	}
	
	return errors
}

// validateMainConfigurationParameters validates core service configuration parameters
func validateMainConfigurationParameters() []ConfigValidationError {
	var errors []ConfigValidationError
	
	// Validate required DATABASE_URL
	if os.Getenv("DATABASE_URL") == "" {
		errors = append(errors, ConfigValidationError{
			Field:   "DATABASE_URL",
			Value:   "",
			Message: "required environment variable is not set",
		})
	}
	
	// Validate PORT
	if port := os.Getenv("PORT"); port != "" {
		if err := validatePort(port); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "PORT",
				Value:   port,
				Message: err.Error(),
			})
		}
	}
	
	// Validate GRPC_PORT
	if grpcPort := os.Getenv("GRPC_PORT"); grpcPort != "" {
		if err := validatePort(grpcPort); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "GRPC_PORT",
				Value:   grpcPort,
				Message: err.Error(),
			})
		}
	}
	
	// Validate gRPC addresses
	if authAddr := os.Getenv("AUTH_GRPC_ADDR"); authAddr != "" {
		if err := validateGRPCAddress(authAddr); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "AUTH_GRPC_ADDR",
				Value:   authAddr,
				Message: err.Error(),
			})
		}
	}
	
	if maxbotAddr := os.Getenv("MAXBOT_GRPC_ADDR"); maxbotAddr != "" {
		if err := validateGRPCAddress(maxbotAddr); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "MAXBOT_GRPC_ADDR",
				Value:   maxbotAddr,
				Message: err.Error(),
			})
		}
	}
	
	return errors
}

// validateParticipantsConfigurationParameters validates participants-specific configuration parameters
func validateParticipantsConfigurationParameters() []ConfigValidationError {
	var errors []ConfigValidationError
	
	// Validate duration parameters
	durationParams := map[string]struct {
		min, max time.Duration
	}{
		"PARTICIPANTS_CACHE_TTL":         {1 * time.Minute, 24 * time.Hour},
		"PARTICIPANTS_UPDATE_INTERVAL":   {1 * time.Minute, 24 * time.Hour},
		"PARTICIPANTS_MAX_API_TIMEOUT":   {1 * time.Second, 5 * time.Minute},
		"PARTICIPANTS_STALE_THRESHOLD":   {1 * time.Minute, 24 * time.Hour},
	}
	
	for param, bounds := range durationParams {
		if val := os.Getenv(param); val != "" {
			if duration, err := time.ParseDuration(val); err != nil {
				errors = append(errors, ConfigValidationError{
					Field:   param,
					Value:   val,
					Message: fmt.Sprintf("invalid duration format: %v", err),
				})
			} else if duration < bounds.min || duration > bounds.max {
				errors = append(errors, ConfigValidationError{
					Field:   param,
					Value:   val,
					Message: fmt.Sprintf("duration out of range [%v, %v]", bounds.min, bounds.max),
				})
			}
		}
	}
	
	// Validate integer parameters
	intParams := map[string]struct {
		min, max int
	}{
		"PARTICIPANTS_FULL_UPDATE_HOUR": {0, 23},
		"PARTICIPANTS_BATCH_SIZE":       {1, 1000},
		"PARTICIPANTS_MAX_RETRIES":      {0, 10},
	}
	
	for param, bounds := range intParams {
		if val := os.Getenv(param); val != "" {
			if intVal, err := strconv.Atoi(val); err != nil {
				errors = append(errors, ConfigValidationError{
					Field:   param,
					Value:   val,
					Message: fmt.Sprintf("invalid integer format: %v", err),
				})
			} else if intVal < bounds.min || intVal > bounds.max {
				errors = append(errors, ConfigValidationError{
					Field:   param,
					Value:   val,
					Message: fmt.Sprintf("value out of range [%d, %d]", bounds.min, bounds.max),
				})
			}
		}
	}
	
	// Validate boolean parameters
	boolParams := []string{
		"PARTICIPANTS_ENABLE_BACKGROUND_SYNC",
		"PARTICIPANTS_ENABLE_LAZY_UPDATE",
		"PARTICIPANTS_DISABLED",
	}
	
	for _, param := range boolParams {
		if val := os.Getenv(param); val != "" {
			if _, err := parseFlexibleBool(val); err != nil {
				errors = append(errors, ConfigValidationError{
					Field:   param,
					Value:   val,
					Message: fmt.Sprintf("invalid boolean format: %v", err),
				})
			}
		}
	}
	
	return errors
}

// validateRedisReconnectionParameters validates Redis automatic reconnection configuration
func validateRedisReconnectionParameters() []ConfigValidationError {
	var errors []ConfigValidationError
	
	// Validate REDIS_MAX_RETRIES
	if val := os.Getenv("REDIS_MAX_RETRIES"); val != "" {
		if retries, err := strconv.Atoi(val); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_MAX_RETRIES",
				Value:   val,
				Message: fmt.Sprintf("invalid integer format: %v", err),
			})
		} else if retries < 1 || retries > 20 {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_MAX_RETRIES",
				Value:   val,
				Message: "value out of range [1, 20]",
			})
		}
	}
	
	// Validate REDIS_RETRY_DELAY
	if val := os.Getenv("REDIS_RETRY_DELAY"); val != "" {
		if delay, err := time.ParseDuration(val); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_RETRY_DELAY",
				Value:   val,
				Message: fmt.Sprintf("invalid duration format: %v", err),
			})
		} else if delay < 100*time.Millisecond || delay > 30*time.Second {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_RETRY_DELAY",
				Value:   val,
				Message: "duration out of range [100ms, 30s]",
			})
		}
	}
	
	// Validate REDIS_HEALTH_CHECK_INTERVAL
	if val := os.Getenv("REDIS_HEALTH_CHECK_INTERVAL"); val != "" {
		if interval, err := time.ParseDuration(val); err != nil {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_HEALTH_CHECK_INTERVAL",
				Value:   val,
				Message: fmt.Sprintf("invalid duration format: %v", err),
			})
		} else if interval < 10*time.Second || interval > 5*time.Minute {
			errors = append(errors, ConfigValidationError{
				Field:   "REDIS_HEALTH_CHECK_INTERVAL",
				Value:   val,
				Message: "duration out of range [10s, 5m]",
			})
		}
	}
	
	return errors
}

// logConfigWarning logs configuration validation warnings
func logConfigWarning(err ConfigValidationError) {
	log.Printf("CONFIG WARNING: %s", err.Error())
}