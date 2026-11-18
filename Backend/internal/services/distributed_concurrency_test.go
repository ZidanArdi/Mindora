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

func TestDistributedMultiInstanceRefreshConcurrency(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	testEmail := "multiinstance.phase36@mindora.app"

	// Clean up after test
	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email = ?)", testEmail)
		_ = db.DB.Exec("DELETE FROM users WHERE email = ?", testEmail)
	}()

	// Instantiate TWO completely independent authentication service instances
	// (Simulating two separate backend container/server replicas sharing ONLY PostgreSQL)
	serviceA, _ := setupTestAuthService(t, db.DB, nil, nil)
	serviceB, _ := setupTestAuthService(t, db.DB, nil, nil)

	ctx := context.Background()

	// Initial user registration via Service A
	regReq := services.RegisterRequest{
		Name:     "Multi-Instance Phase 3.6 User",
		Email:    testEmail,
		Password: "SecurePassword123!",
	}
	authRes, err := serviceA.Register(ctx, regReq, "127.0.0.1")
	if err != nil {
		t.Fatalf("initial register failed: %v", err)
	}

	userID := authRes.User.ID

	// =========================================================================
	// SCENARIO 1: Two concurrent requests with the SAME token and SAME operation identifier
	// =========================================================================
	t.Run("Scenario 1: Concurrent with SAME Operation ID (Idempotent Handoff)", func(t *testing.T) {
		token1 := authRes.RefreshToken
		sharedOpID := "op-shared-12345"

		var wg sync.WaitGroup
		var resA, resB *services.AuthResult
		var errA, errB error

		wg.Add(2)
		go func() {
			defer wg.Done()
			resA, errA = serviceA.Refresh(ctx, token1, sharedOpID, "127.0.0.1", "Replica-A")
		}()
		go func() {
			defer wg.Done()
			resB, errB = serviceB.Refresh(ctx, token1, sharedOpID, "127.0.0.1", "Replica-B")
		}()
		wg.Wait()

		if errA != nil {
			t.Fatalf("Service A failed on shared op ID: %v", errA)
		}
		if errB != nil {
			t.Fatalf("Service B failed on shared op ID: %v", errB)
		}

		// Both instances should receive the exact same valid replacement refresh token
		if resA.RefreshToken == "" || resB.RefreshToken == "" {
			t.Fatalf("expected non-empty refresh tokens, got A=%q, B=%q", resA.RefreshToken, resB.RefreshToken)
		}
		if resA.RefreshToken != resB.RefreshToken {
			t.Errorf("expected identical replacement tokens under idempotent shared op ID, got A=%q, B=%q", resA.RefreshToken, resB.RefreshToken)
		}

		// Verify database has exactly 1 active unrevoked session (no duplicate branch)
		var activeSessions []models.AuthSession
		db.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&activeSessions)
		if len(activeSessions) != 1 {
			t.Fatalf("expected exactly 1 active session in DB, got %d", len(activeSessions))
		}

		// Update active token pointer for next scenario
		authRes.RefreshToken = resA.RefreshToken
	})

	// =========================================================================
	// SCENARIO 2: Two concurrent requests with the SAME token but DIFFERENT operation identifiers
	// =========================================================================
	t.Run("Scenario 2: Concurrent with DIFFERENT Operation IDs (Conflict Resolution)", func(t *testing.T) {
		token2 := authRes.RefreshToken

		var wg sync.WaitGroup
		var resA, resB *services.AuthResult
		var errA, errB error

		wg.Add(2)
		go func() {
			defer wg.Done()
			resA, errA = serviceA.Refresh(ctx, token2, "op-client-A", "127.0.0.1", "Replica-A")
		}()
		go func() {
			defer wg.Done()
			resB, errB = serviceB.Refresh(ctx, token2, "op-client-B", "127.0.0.1", "Replica-B")
		}()
		wg.Wait()

		// Exactly one must succeed and the other must receive ErrConcurrentConflict (HTTP 409)
		// Crucially, NO accidental family revocation must occur!
		var activeSessions []models.AuthSession
		db.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&activeSessions)
		if len(activeSessions) != 1 {
			t.Fatalf("CRITICAL: Concurrent conflict caused accidental family revocation! Active sessions = %d (expected 1)", len(activeSessions))
		}

		var winnerRes *services.AuthResult
		if errA == nil && errors.Is(errB, services.ErrConcurrentConflict) {
			winnerRes = resA
		} else if errB == nil && errors.Is(errA, services.ErrConcurrentConflict) {
			winnerRes = resB
		} else {
			t.Fatalf("expected exactly one winner and one ErrConcurrentConflict, got errA=%v, errB=%v", errA, errB)
		}

		// Update active token pointer
		authRes.RefreshToken = winnerRes.RefreshToken
	})

	// =========================================================================
	// SCENARIO 3: Retry after a simulated network timeout (Idempotency Key Retry)
	// =========================================================================
	t.Run("Scenario 3: Retry After Simulated Network Timeout", func(t *testing.T) {
		token3 := authRes.RefreshToken
		timeoutOpID := "op-timeout-retry-999"

		// 1. Initial refresh attempt
		firstRes, err := serviceA.Refresh(ctx, token3, timeoutOpID, "127.0.0.1", "Replica-A")
		if err != nil {
			t.Fatalf("initial refresh attempt failed: %v", err)
		}

		// 2. Simulate client network drop/timeout: Client retries the same operation on Service B
		retryRes, err := serviceB.Refresh(ctx, token3, timeoutOpID, "127.0.0.1", "Replica-B")
		if err != nil {
			t.Fatalf("retry after network timeout failed: %v", err)
		}

		// Both should return identical replacement token
		if retryRes.RefreshToken != firstRes.RefreshToken {
			t.Errorf("expected idempotent retry to return same token, got %q vs %q", retryRes.RefreshToken, firstRes.RefreshToken)
		}

		// DB session count remains exactly 1 active session
		var activeSessions []models.AuthSession
		db.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&activeSessions)
		if len(activeSessions) != 1 {
			t.Fatalf("expected exactly 1 active session in DB after retry, got %d", len(activeSessions))
		}

		// Update active token pointer
		authRes.RefreshToken = firstRes.RefreshToken
	})

	// =========================================================================
	// SCENARIO 4: Replacement token remains usable after legitimate retry
	// =========================================================================
	t.Run("Scenario 4: Replacement Token Usable After Retry", func(t *testing.T) {
		replacementToken := authRes.RefreshToken

		subsequentRes, err := serviceA.Refresh(ctx, replacementToken, "op-next-cycle", "127.0.0.1", "Replica-A")
		if err != nil {
			t.Fatalf("replacement token failed to rotate: %v", err)
		}

		if subsequentRes.RefreshToken == "" || subsequentRes.RefreshToken == replacementToken {
			t.Errorf("expected valid new rotated token, got %q", subsequentRes.RefreshToken)
		}

		var activeSessions []models.AuthSession
		db.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&activeSessions)
		if len(activeSessions) != 1 {
			t.Fatalf("expected exactly 1 active session in DB, got %d", len(activeSessions))
		}

		// Save the older token for replay test and update active token
		stolenToken := replacementToken
		authRes.RefreshToken = subsequentRes.RefreshToken

		// =========================================================================
		// SCENARIO 5: Sequential replay using an already-rotated token (Token Reuse Detection)
		// =========================================================================
		t.Run("Scenario 5: Sequential Replay of Already-Rotated Token", func(t *testing.T) {
			// Attacker presents the old stolen token with their own operation ID
			_, err := serviceB.Refresh(ctx, stolenToken, "op-attacker-stolen", "127.0.0.1", "AttackerClient")
			if !errors.Is(err, services.ErrTokenReuseDetected) {
				t.Fatalf("expected ErrTokenReuseDetected on sequential replay, got: %v", err)
			}

			// =========================================================================
			// SCENARIO 6: Refresh family revocation after confirmed reuse
			// =========================================================================
			var sessionsAfterReuse []models.AuthSession
			db.DB.Where("user_id = ? AND revoked_at IS NULL", userID).Find(&sessionsAfterReuse)
			if len(sessionsAfterReuse) != 0 {
				t.Fatalf("expected 0 active sessions after confirmed reuse detection, got %d", len(sessionsAfterReuse))
			}

			// =========================================================================
			// SCENARIO 7: Even the previously valid replacement token is now terminated
			// =========================================================================
			_, err = serviceA.Refresh(ctx, authRes.RefreshToken, "op-victim-attempt", "127.0.0.1", "VictimClient")
			if !errors.Is(err, services.ErrTokenReuseDetected) && !errors.Is(err, services.ErrInvalidRefreshToken) {
				t.Errorf("expected terminated session family after reuse detection, got: %v", err)
			}
		})
	})
}
