package jwt_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/pkg/jwt"
)

func TestTokenManager_AccessAndRefresh(t *testing.T) {
	tm := jwt.NewTokenManager("my-test-super-secret-key-32chars", 10*time.Minute, 24*time.Hour)

	userID := uuid.New()
	email := "test.user@mindora.app"
	role := "user"

	// 1. Generate Access Token
	tokenStr, expiresAt, err := tm.GenerateAccessToken(userID, email, role)
	if err != nil {
		t.Fatalf("unexpected token generation error: %v", err)
	}

	if tokenStr == "" {
		t.Errorf("expected non-empty token string")
	}
	if expiresAt.Before(time.Now().UTC()) {
		t.Errorf("expected future expiration")
	}

	// 2. Validate Access Token
	claims, err := tm.ValidateAccessToken(tokenStr)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected user ID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}

	// 3. Test Invalid Token
	if _, err := tm.ValidateAccessToken("invalid.jwt.token"); err == nil {
		t.Errorf("expected error for malformed token, got nil")
	}

	// 4. Test Expired Token
	expiredTM := jwt.NewTokenManager("my-test-super-secret-key-32chars", -1*time.Minute, 24*time.Hour)
	expiredToken, _, _ := expiredTM.GenerateAccessToken(userID, email, role)
	if _, err := tm.ValidateAccessToken(expiredToken); err == nil {
		t.Errorf("expected error for expired token, got nil")
	}

	// 5. Test Refresh Token Generation & Hashing
	rawRefresh, tokenHash, refreshExpiry, err := tm.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected refresh generation error: %v", err)
	}

	if len(rawRefresh) != 64 {
		t.Errorf("expected 64 hex characters for 32 bytes, got %d", len(rawRefresh))
	}
	if tokenHash != jwt.HashToken(rawRefresh) {
		t.Errorf("expected hash to match deterministic HashToken output")
	}
	if refreshExpiry.Before(time.Now().UTC()) {
		t.Errorf("expected future refresh token expiration")
	}
}
