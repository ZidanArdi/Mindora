package services_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
	"mindora/backend/internal/services"
)

func TestRefreshTokenConcurrencyAndReuseScenarios(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Clean up after test
	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email = 'concurrency.test@mindora.app')")
		_ = db.DB.Exec("DELETE FROM users WHERE email = 'concurrency.test@mindora.app'")
	}()

	authSvc, _ := setupTestAuthService(t, db.DB, nil, nil)
	ctx := context.Background()

	// Register user
	regReq := services.RegisterRequest{
		Name:     "Concurrency Test User",
		Email:    "concurrency.test@mindora.app",
		Password: "SecurePassword123!",
	}
	authRes, err := authSvc.Register(ctx, regReq, "127.0.0.1")
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	initialRefreshToken := authRes.RefreshToken

	// =========================================================================
	// SCENARIO 1: Two simultaneous requests using the exact same refresh token
	// =========================================================================
	var wg sync.WaitGroup
	results := make([]*services.AuthResult, 2)
	errs := make([]error, 2)

	wg.Add(2)
	for i := 0; i < 2; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = authSvc.Refresh(ctx, initialRefreshToken, "shared-op-id", "127.0.0.1", "ConcurrentAgent")
		}(i)
	}
	wg.Wait()

	if errs[0] != nil {
		t.Errorf("expected concurrent request 0 to succeed, got error: %v", errs[0])
	}
	if errs[1] != nil {
		t.Errorf("expected concurrent request 1 to succeed, got error: %v", errs[1])
	}

	// Verify that exactly ONE replacement refresh token was issued
	var newlyIssuedRefresh string
	if results[0].RefreshToken != "" {
		newlyIssuedRefresh = results[0].RefreshToken
	} else if results[1].RefreshToken != "" {
		newlyIssuedRefresh = results[1].RefreshToken
	} else {
		t.Fatalf("expected at least one request to return the newly rotated refresh token")
	}

	// Verify that the user still has an active session in the database (family was NOT destroyed)
	var activeSessions []models.AuthSession
	db.DB.Where("user_id = ? AND revoked_at IS NULL", authRes.User.ID).Find(&activeSessions)
	if len(activeSessions) != 1 {
		t.Fatalf("expected exactly 1 active unrevoked session in DB, got %d", len(activeSessions))
	}

	// =========================================================================
	// SCENARIO 2: Valid newly issued session remains usable
	// =========================================================================
	subsequentRes, err := authSvc.Refresh(ctx, newlyIssuedRefresh, "legit-op-2", "127.0.0.1", "LegitAgent")
	if err != nil {
		t.Fatalf("expected newly issued refresh token to be usable, got error: %v", err)
	}
	if subsequentRes.RefreshToken == "" || subsequentRes.RefreshToken == newlyIssuedRefresh {
		t.Errorf("expected new replacement token from subsequent refresh, got: %s", subsequentRes.RefreshToken)
	}

	activeTokenAfterSubsequent := subsequentRes.RefreshToken

	// =========================================================================
	// SCENARIO 3: Genuine token replay / sequential reuse of an already-rotated token
	// =========================================================================
	_, err = authSvc.Refresh(ctx, initialRefreshToken, "attacker-op-3", "127.0.0.1", "AttackerAgent")
	if !errors.Is(err, services.ErrTokenReuseDetected) {
		t.Fatalf("expected ErrTokenReuseDetected for replaying rotated token, got: %v", err)
	}

	// =========================================================================
	// SCENARIO 4: Session family revocation after confirmed reuse
	// =========================================================================
	var activeSessionsAfterReplay []models.AuthSession
	db.DB.Where("user_id = ? AND revoked_at IS NULL", authRes.User.ID).Find(&activeSessionsAfterReplay)
	if len(activeSessionsAfterReplay) != 0 {
		t.Fatalf("expected 0 active sessions after reuse detection, got %d", len(activeSessionsAfterReplay))
	}

	// =========================================================================
	// SCENARIO 5: Previously valid active token is now also invalidated due to family revocation
	// =========================================================================
	_, err = authSvc.Refresh(ctx, activeTokenAfterSubsequent, "legit-op-4", "127.0.0.1", "LegitAgent")
	if !errors.Is(err, services.ErrTokenReuseDetected) && !errors.Is(err, services.ErrInvalidRefreshToken) {
		t.Errorf("expected session to be unusable after family revocation, got: %v", err)
	}
}
