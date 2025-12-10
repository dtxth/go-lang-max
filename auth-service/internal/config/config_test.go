package config

import (
	"os"
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with mock notification service",
			config: &Config{
				MinPasswordLength:       12,
				ResetTokenExpiration:    15,
				NotificationServiceType: "mock",
			},
			wantErr: false,
		},
		{
			name: "valid config with max notification service",
			config: &Config{
				MinPasswordLength:       12,
				ResetTokenExpiration:    15,
				NotificationServiceType: "max",
				MaxBotServiceAddr:       "localhost:9090",
			},
			wantErr: false,
		},
		{
			name: "invalid - password length too short",
			config: &Config{
				MinPasswordLength:       7,
				ResetTokenExpiration:    15,
				NotificationServiceType: "mock",
			},
			wantErr: true,
			errMsg:  "MIN_PASSWORD_LENGTH must be at least 8",
		},
		{
			name: "invalid - reset token expiration too short",
			config: &Config{
				MinPasswordLength:       12,
				ResetTokenExpiration:    0,
				NotificationServiceType: "mock",
			},
			wantErr: true,
			errMsg:  "RESET_TOKEN_EXPIRATION must be at least 1 minute",
		},
		{
			name: "invalid - unknown notification service type",
			config: &Config{
				MinPasswordLength:       12,
				ResetTokenExpiration:    15,
				NotificationServiceType: "unknown",
			},
			wantErr: true,
			errMsg:  "NOTIFICATION_SERVICE_TYPE must be 'mock' or 'max'",
		},
		{
			name: "invalid - max service without address",
			config: &Config{
				MinPasswordLength:       12,
				ResetTokenExpiration:    15,
				NotificationServiceType: "max",
				MaxBotServiceAddr:       "",
			},
			wantErr: true,
			errMsg:  "MAXBOT_SERVICE_ADDR is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %v", err, tt.errMsg)
				}
			}
		})
	}
}

func TestLoad(t *testing.T) {
	// Save original env vars
	originalVars := map[string]string{
		"MIN_PASSWORD_LENGTH":       os.Getenv("MIN_PASSWORD_LENGTH"),
		"RESET_TOKEN_EXPIRATION":    os.Getenv("RESET_TOKEN_EXPIRATION"),
		"NOTIFICATION_SERVICE_TYPE": os.Getenv("NOTIFICATION_SERVICE_TYPE"),
		"MAXBOT_SERVICE_ADDR":       os.Getenv("MAXBOT_SERVICE_ADDR"),
	}
	
	// Restore env vars after test
	defer func() {
		for key, val := range originalVars {
			if val == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, val)
			}
		}
	}()

	tests := []struct {
		name    string
		envVars map[string]string
		wantErr bool
		check   func(*testing.T, *Config)
	}{
		{
			name: "loads with defaults",
			envVars: map[string]string{
				"MIN_PASSWORD_LENGTH":       "",
				"RESET_TOKEN_EXPIRATION":    "",
				"NOTIFICATION_SERVICE_TYPE": "",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.MinPasswordLength != 12 {
					t.Errorf("MinPasswordLength = %d, want 12", cfg.MinPasswordLength)
				}
				if cfg.ResetTokenExpiration != 15 {
					t.Errorf("ResetTokenExpiration = %d, want 15", cfg.ResetTokenExpiration)
				}
				if cfg.NotificationServiceType != "mock" {
					t.Errorf("NotificationServiceType = %s, want mock", cfg.NotificationServiceType)
				}
			},
		},
		{
			name: "loads with custom values",
			envVars: map[string]string{
				"MIN_PASSWORD_LENGTH":       "16",
				"RESET_TOKEN_EXPIRATION":    "30",
				"NOTIFICATION_SERVICE_TYPE": "max",
				"MAXBOT_SERVICE_ADDR":       "localhost:9090",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.MinPasswordLength != 16 {
					t.Errorf("MinPasswordLength = %d, want 16", cfg.MinPasswordLength)
				}
				if cfg.ResetTokenExpiration != 30 {
					t.Errorf("ResetTokenExpiration = %d, want 30", cfg.ResetTokenExpiration)
				}
				if cfg.NotificationServiceType != "max" {
					t.Errorf("NotificationServiceType = %s, want max", cfg.NotificationServiceType)
				}
				if cfg.MaxBotServiceAddr != "localhost:9090" {
					t.Errorf("MaxBotServiceAddr = %s, want localhost:9090", cfg.MaxBotServiceAddr)
				}
			},
		},
		{
			name: "fails validation with invalid password length",
			envVars: map[string]string{
				"MIN_PASSWORD_LENGTH": "5",
			},
			wantErr: true,
		},
		{
			name: "fails validation with invalid reset token expiration",
			envVars: map[string]string{
				"RESET_TOKEN_EXPIRATION": "0",
			},
			wantErr: true,
		},
		{
			name: "fails validation with invalid notification service type",
			envVars: map[string]string{
				"NOTIFICATION_SERVICE_TYPE": "invalid",
			},
			wantErr: true,
		},
		{
			name: "fails validation when max service type without address",
			envVars: map[string]string{
				"NOTIFICATION_SERVICE_TYPE": "max",
				"MAXBOT_SERVICE_ADDR":       "",
			},
			wantErr: true,
		},
		{
			name: "handles invalid integer for password length - uses default",
			envVars: map[string]string{
				"MIN_PASSWORD_LENGTH":       "invalid",
				"RESET_TOKEN_EXPIRATION":    "",
				"NOTIFICATION_SERVICE_TYPE": "",
				"MAXBOT_SERVICE_ADDR":       "",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.MinPasswordLength != 12 {
					t.Errorf("MinPasswordLength = %d, want 12 (default)", cfg.MinPasswordLength)
				}
			},
		},
		{
			name: "handles invalid integer for reset token expiration - uses default",
			envVars: map[string]string{
				"MIN_PASSWORD_LENGTH":       "",
				"RESET_TOKEN_EXPIRATION":    "invalid",
				"NOTIFICATION_SERVICE_TYPE": "",
				"MAXBOT_SERVICE_ADDR":       "",
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				if cfg.ResetTokenExpiration != 15 {
					t.Errorf("ResetTokenExpiration = %d, want 15 (default)", cfg.ResetTokenExpiration)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for key, val := range tt.envVars {
				if val == "" {
					os.Unsetenv(key)
				} else {
					os.Setenv(key, val)
				}
			}

			cfg, err := Load()
			if (err != nil) != tt.wantErr {
				t.Errorf("Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, cfg)
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
