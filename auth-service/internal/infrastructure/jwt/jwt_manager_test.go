package jwt

import (
	"auth-service/internal/domain"
	"testing"
	"time"
)

func TestGenerateTokensWithContext(t *testing.T) {
	manager := NewManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)
	
	userID := int64(123)
	email := "test@example.com"
	role := "curator"
	universityID := int64(456)
	branchID := int64(789)
	
	ctx := &domain.TokenContext{
		UniversityID: &universityID,
		BranchID:     &branchID,
	}
	
	tokens, err := manager.GenerateTokensWithContext(userID, email, role, ctx)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}
	
	if tokens.AccessToken == "" {
		t.Error("Access token is empty")
	}
	
	if tokens.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}
	
	// Verify the access token contains the context
	verifiedUserID, verifiedEmail, verifiedRole, verifiedCtx, err := manager.VerifyAccessTokenWithContext(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}
	
	if verifiedUserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, verifiedUserID)
	}
	
	if verifiedEmail != email {
		t.Errorf("Expected email %s, got %s", email, verifiedEmail)
	}
	
	if verifiedRole != role {
		t.Errorf("Expected role %s, got %s", role, verifiedRole)
	}
	
	if verifiedCtx == nil {
		t.Fatal("Context is nil")
	}
	
	if verifiedCtx.UniversityID == nil || *verifiedCtx.UniversityID != universityID {
		t.Errorf("Expected university ID %d, got %v", universityID, verifiedCtx.UniversityID)
	}
	
	if verifiedCtx.BranchID == nil || *verifiedCtx.BranchID != branchID {
		t.Errorf("Expected branch ID %d, got %v", branchID, verifiedCtx.BranchID)
	}
	
	if verifiedCtx.FacultyID != nil {
		t.Errorf("Expected faculty ID to be nil, got %v", verifiedCtx.FacultyID)
	}
}

func TestGenerateTokensWithoutContext(t *testing.T) {
	manager := NewManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 7*24*time.Hour)
	
	userID := int64(123)
	email := "test@example.com"
	role := "superadmin"
	
	tokens, err := manager.GenerateTokens(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate tokens: %v", err)
	}
	
	// Verify the access token works without context
	verifiedUserID, verifiedEmail, verifiedRole, verifiedCtx, err := manager.VerifyAccessTokenWithContext(tokens.AccessToken)
	if err != nil {
		t.Fatalf("Failed to verify access token: %v", err)
	}
	
	if verifiedUserID != userID {
		t.Errorf("Expected user ID %d, got %d", userID, verifiedUserID)
	}
	
	if verifiedEmail != email {
		t.Errorf("Expected email %s, got %s", email, verifiedEmail)
	}
	
	if verifiedRole != role {
		t.Errorf("Expected role %s, got %s", role, verifiedRole)
	}
	
	if verifiedCtx == nil {
		t.Fatal("Context should not be nil")
	}
	
	// For superadmin, context fields should be nil
	if verifiedCtx.UniversityID != nil {
		t.Errorf("Expected university ID to be nil for superadmin, got %v", verifiedCtx.UniversityID)
	}
	
	if verifiedCtx.BranchID != nil {
		t.Errorf("Expected branch ID to be nil for superadmin, got %v", verifiedCtx.BranchID)
	}
	
	if verifiedCtx.FacultyID != nil {
		t.Errorf("Expected faculty ID to be nil for superadmin, got %v", verifiedCtx.FacultyID)
	}
}
