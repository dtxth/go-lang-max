package repository

import (
	"database/sql"
	"os"
	"testing"
	"time"

	"auth-service/internal/domain"
	"auth-service/internal/infrastructure/database"
	_ "github.com/lib/pq"
)

// setupTestDB creates a test database connection
func setupTestDB(t *testing.T) *database.DB {
	// Use test database connection string from environment or default
	connStr := os.Getenv("TEST_DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	sqlDB, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Skipping database tests - cannot connect to test database: %v", err)
		return nil
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("Skipping database tests - database not available: %v", err)
		return nil
	}

	// Wrap with database wrapper
	db := database.NewDBFromConnection(sqlDB, nil)

	// Ensure the password_reset_tokens table exists
	ensureTableExists(t, sqlDB)

	// Clean up any existing test data
	cleanupTestData(t, sqlDB)

	return db
}

// ensureTableExists creates the password_reset_tokens table if it doesn't exist
func ensureTableExists(t *testing.T, db *sql.DB) {
	// Create the table (using IF NOT EXISTS so it's safe to run multiple times)
	createTableSQL := `
		CREATE TABLE IF NOT EXISTS password_reset_tokens (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL,
			token VARCHAR(64) NOT NULL UNIQUE,
			expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
			used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
		);

		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token ON password_reset_tokens(token);
		CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
	`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		t.Skipf("Skipping database tests - cannot create table: %v", err)
	}
}

// cleanupTestData removes test data from the database
func cleanupTestData(t *testing.T, db *sql.DB) {
	// Delete all password reset tokens
	_, err := db.Exec("DELETE FROM password_reset_tokens")
	if err != nil {
		t.Logf("Warning: failed to cleanup password_reset_tokens: %v", err)
	}
}

func TestPasswordResetPostgres_Create(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPasswordResetPostgres(db)

	t.Run("valid token creation", func(t *testing.T) {
		token := &domain.PasswordResetToken{
			UserID:    1,
			Token:     "test-token-create-valid",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		err := repo.Create(token)
		if err != nil {
			t.Fatalf("Create() error = %v, expected no error", err)
		}

		// Verify token was created with ID and CreatedAt
		if token.ID == 0 {
			t.Error("Expected token ID to be set")
		}
		if token.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}

		// Verify we can retrieve the token
		retrieved, err := repo.GetByToken(token.Token)
		if err != nil {
			t.Fatalf("Failed to retrieve created token: %v", err)
		}
		if retrieved.UserID != token.UserID {
			t.Errorf("UserID = %v, want %v", retrieved.UserID, token.UserID)
		}
		if retrieved.Token != token.Token {
			t.Errorf("Token = %v, want %v", retrieved.Token, token.Token)
		}

		// Cleanup
		cleanupTestData(t, db.GetUnderlyingDB())
	})

	t.Run("duplicate token should fail", func(t *testing.T) {
		token1 := &domain.PasswordResetToken{
			UserID:    1,
			Token:     "test-token-duplicate",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		err := repo.Create(token1)
		if err != nil {
			t.Fatalf("First Create() error = %v, expected no error", err)
		}

		// Try to create duplicate
		token2 := &domain.PasswordResetToken{
			UserID:    2,
			Token:     "test-token-duplicate", // Same token
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}

		err = repo.Create(token2)
		if err == nil {
			t.Error("Expected error for duplicate token, got nil")
		}

		// Cleanup
		cleanupTestData(t, db.GetUnderlyingDB())
	})
}

func TestPasswordResetPostgres_GetByToken(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPasswordResetPostgres(db)

	t.Run("existing token", func(t *testing.T) {
		// Create a test token
		testToken := &domain.PasswordResetToken{
			UserID:    1,
			Token:     "test-get-token-exists",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		err := repo.Create(testToken)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		result, err := repo.GetByToken("test-get-token-exists")
		if err != nil {
			t.Fatalf("GetByToken() error = %v, expected no error", err)
		}

		if result == nil {
			t.Fatal("Expected token to be returned")
		}
		if result.Token != "test-get-token-exists" {
			t.Errorf("GetByToken() token = %v, want %v", result.Token, "test-get-token-exists")
		}
		if result.UserID != testToken.UserID {
			t.Errorf("GetByToken() userID = %v, want %v", result.UserID, testToken.UserID)
		}
		if result.UsedAt != nil {
			t.Error("Expected UsedAt to be nil for new token")
		}

		// Cleanup
		cleanupTestData(t, db.GetUnderlyingDB())
	})

	t.Run("non-existent token", func(t *testing.T) {
		result, err := repo.GetByToken("non-existent-token")
		if err != domain.ErrNotFound {
			t.Errorf("GetByToken() error = %v, want ErrNotFound", err)
		}
		if result != nil {
			t.Error("Expected nil result for non-existent token")
		}
	})
}

func TestPasswordResetPostgres_Invalidate(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPasswordResetPostgres(db)

	t.Run("invalidate existing token", func(t *testing.T) {
		// Create a test token
		testToken := &domain.PasswordResetToken{
			UserID:    1,
			Token:     "test-invalidate-token-valid",
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
		err := repo.Create(testToken)
		if err != nil {
			t.Fatalf("Failed to create test token: %v", err)
		}

		// Verify token is initially valid
		result, err := repo.GetByToken("test-invalidate-token-valid")
		if err != nil {
			t.Fatalf("Failed to get token before invalidation: %v", err)
		}
		if result.UsedAt != nil {
			t.Error("Expected UsedAt to be nil before invalidation")
		}

		// Invalidate the token
		err = repo.Invalidate("test-invalidate-token-valid")
		if err != nil {
			t.Fatalf("Invalidate() error = %v, expected no error", err)
		}

		// Verify token was invalidated
		result, err = repo.GetByToken("test-invalidate-token-valid")
		if err != nil {
			t.Fatalf("Failed to get token after invalidation: %v", err)
		}
		if result.UsedAt == nil {
			t.Error("Expected UsedAt to be set after invalidation")
		}
		if !result.IsUsed() {
			t.Error("Expected token to be marked as used")
		}
		if result.IsValid() {
			t.Error("Expected token to be invalid after invalidation")
		}

		// Cleanup
		cleanupTestData(t, db.GetUnderlyingDB())
	})

	t.Run("invalidate non-existent token should not error", func(t *testing.T) {
		err := repo.Invalidate("non-existent-token-invalidate")
		if err != nil {
			t.Errorf("Invalidate() error = %v, expected no error for non-existent token", err)
		}
	})
}

func TestPasswordResetPostgres_DeleteExpired(t *testing.T) {
	db := setupTestDB(t)
	if db == nil {
		return
	}
	defer db.Close()

	repo := NewPasswordResetPostgres(db)

	// Create expired token
	expiredToken := &domain.PasswordResetToken{
		UserID:    1,
		Token:     "expired-token-cleanup",
		ExpiresAt: time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
	}
	err := repo.Create(expiredToken)
	if err != nil {
		t.Fatalf("Failed to create expired token: %v", err)
	}

	// Create another expired token
	expiredToken2 := &domain.PasswordResetToken{
		UserID:    1,
		Token:     "expired-token-cleanup-2",
		ExpiresAt: time.Now().Add(-30 * time.Minute), // Expired 30 minutes ago
	}
	err = repo.Create(expiredToken2)
	if err != nil {
		t.Fatalf("Failed to create second expired token: %v", err)
	}

	// Create valid token
	validToken := &domain.PasswordResetToken{
		UserID:    2,
		Token:     "valid-token-cleanup",
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	err = repo.Create(validToken)
	if err != nil {
		t.Fatalf("Failed to create valid token: %v", err)
	}

	// Verify all tokens exist before cleanup
	_, err = repo.GetByToken("expired-token-cleanup")
	if err != nil {
		t.Fatalf("Expected expired token to exist before cleanup: %v", err)
	}
	_, err = repo.GetByToken("expired-token-cleanup-2")
	if err != nil {
		t.Fatalf("Expected second expired token to exist before cleanup: %v", err)
	}
	_, err = repo.GetByToken("valid-token-cleanup")
	if err != nil {
		t.Fatalf("Expected valid token to exist before cleanup: %v", err)
	}

	// Delete expired tokens
	err = repo.DeleteExpired()
	if err != nil {
		t.Fatalf("DeleteExpired() error = %v", err)
	}

	// Verify expired tokens were deleted
	_, err = repo.GetByToken("expired-token-cleanup")
	if err != domain.ErrNotFound {
		t.Errorf("Expected first expired token to be deleted, got error: %v", err)
	}

	_, err = repo.GetByToken("expired-token-cleanup-2")
	if err != domain.ErrNotFound {
		t.Errorf("Expected second expired token to be deleted, got error: %v", err)
	}

	// Verify valid token still exists
	result, err := repo.GetByToken("valid-token-cleanup")
	if err != nil {
		t.Fatalf("Expected valid token to still exist: %v", err)
	}
	if result == nil {
		t.Fatal("Expected valid token to be returned")
	}
	if result.Token != "valid-token-cleanup" {
		t.Errorf("Token = %v, want %v", result.Token, "valid-token-cleanup")
	}

	// Cleanup
	cleanupTestData(t, db.GetUnderlyingDB())
}

func TestPasswordResetToken_IsValid(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token *domain.PasswordResetToken
		want  bool
	}{
		{
			name: "valid token",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(15 * time.Minute),
				UsedAt:    nil,
			},
			want: true,
		},
		{
			name: "expired token",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(-1 * time.Minute),
				UsedAt:    nil,
			},
			want: false,
		},
		{
			name: "used token",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(15 * time.Minute),
				UsedAt:    &now,
			},
			want: false,
		},
		{
			name: "used and expired token",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(-1 * time.Minute),
				UsedAt:    &now,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token *domain.PasswordResetToken
		want  bool
	}{
		{
			name: "not expired",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(15 * time.Minute),
			},
			want: false,
		},
		{
			name: "expired",
			token: &domain.PasswordResetToken{
				ExpiresAt: now.Add(-1 * time.Minute),
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordResetToken_IsUsed(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name  string
		token *domain.PasswordResetToken
		want  bool
	}{
		{
			name: "not used",
			token: &domain.PasswordResetToken{
				UsedAt: nil,
			},
			want: false,
		},
		{
			name: "used",
			token: &domain.PasswordResetToken{
				UsedAt: &now,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsUsed(); got != tt.want {
				t.Errorf("IsUsed() = %v, want %v", got, tt.want)
			}
		})
	}
}
