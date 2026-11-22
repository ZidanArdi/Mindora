package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/hash"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/oauth"
	"mindora/backend/pkg/turnstile"
	"mindora/backend/pkg/validator"
)

// MockGoogleVerifier allows mocking Google OIDC token verification for automated tests.
type MockGoogleVerifier struct {
	Profile *oauth.GoogleProfile
	Err     error
}

func (m *MockGoogleVerifier) VerifyIDToken(ctx context.Context, idTokenStr, expectedNonce string) (*oauth.GoogleProfile, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Profile, nil
}

// MockTokenExchanger allows mocking authorization code exchange with Google.
type MockTokenExchanger struct {
	Token *oauth2.Token
	Err   error
}

func (m *MockTokenExchanger) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Token, nil
}

func setupTestAuthService(
	t *testing.T,
	tx *gorm.DB,
	mockVerifier oauth.IDTokenVerifier,
	mockExchanger oauth.TokenExchanger,
) (services.AuthService, *jwt.TokenManager) {
	userRepo := repositories.NewUserRepository(tx)
	identityRepo := repositories.NewIdentityRepository(tx)
	sessionRepo := repositories.NewSessionRepository(tx)

	hasher := hash.NewBcryptHasher(10)
	tokenManager := jwt.NewTokenManager("test-jwt-secret-key-32chars-minimum!", 15*time.Minute, 24*time.Hour)
	turnstileVerifier := turnstile.NewClient(turnstile.Config{Enabled: false}) // Disabled for unit tests
	val := validator.New()

	googleSvc := oauth.NewGoogleServiceWithExchanger(oauth.Config{
		ClientID:     "mock-client-id",
		ClientSecret: "mock-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Enabled:      true,
	}, mockVerifier, mockExchanger)

	authSvc := services.NewAuthService(
		userRepo,
		identityRepo,
		sessionRepo,
		hasher,
		tokenManager,
		turnstileVerifier,
		val,
		googleSvc,
	)

	return authSvc, tokenManager
}

func TestAuthService_EmailRegisterAndLogin(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tx := db.DB.Begin()
	defer tx.Rollback()

	authSvc, _ := setupTestAuthService(t, tx, nil, nil)
	ctx := context.Background()

	// 1. Valid Registration
	regReq := services.RegisterRequest{
		Name:     "Alice Student",
		Email:    "alice@mindora.app",
		Password: "SecurePassword123!",
	}

	result, err := authSvc.Register(ctx, regReq, "127.0.0.1")
	if err != nil {
		t.Fatalf("registration failed: %v", err)
	}

	if result.AccessToken == "" || result.RefreshToken == "" {
		t.Errorf("expected access and refresh tokens")
	}
	if result.User.Role != "user" {
		t.Errorf("expected auto-assigned role 'user', got '%s'", result.User.Role)
	}
	if result.User.Email != "alice@mindora.app" {
		t.Errorf("expected normalized email, got '%s'", result.User.Email)
	}

	// 2. Duplicate Registration with Case-Insensitive Variant
	dupReq := services.RegisterRequest{
		Name:     "Alice Duplicate",
		Email:    "ALICE@MINDORA.APP",
		Password: "SecurePassword123!",
	}

	_, err = authSvc.Register(ctx, dupReq, "127.0.0.1")
	if !errors.Is(err, services.ErrEmailAlreadyRegistered) {
		t.Errorf("expected ErrEmailAlreadyRegistered, got: %v", err)
	}

	// 3. Weak Password Registration
	weakReq := services.RegisterRequest{
		Name:     "Weak User",
		Email:    "weak@mindora.app",
		Password: "short",
	}

	_, err = authSvc.Register(ctx, weakReq, "127.0.0.1")
	if err == nil {
		t.Errorf("expected validation error for weak password, got nil")
	}

	// 4. Successful Login
	loginReq := services.LoginRequest{
		Email:    "alice@mindora.app",
		Password: "SecurePassword123!",
	}

	loginResult, err := authSvc.Login(ctx, loginReq, "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}
	if loginResult.AccessToken == "" || loginResult.RefreshToken == "" {
		t.Errorf("expected non-empty tokens on login")
	}

	// 5. Invalid Password Login
	badPwdReq := services.LoginRequest{
		Email:    "alice@mindora.app",
		Password: "WrongPassword!",
	}
	_, err = authSvc.Login(ctx, badPwdReq, "127.0.0.1", "TestAgent")
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}

	// 6. Non-existent User Login (Generic error)
	nonExistentReq := services.LoginRequest{
		Email:    "ghost@mindora.app",
		Password: "SecurePassword123!",
	}
	_, err = authSvc.Login(ctx, nonExistentReq, "127.0.0.1", "TestAgent")
	if !errors.Is(err, services.ErrInvalidCredentials) {
		t.Errorf("expected generic ErrInvalidCredentials for non-existent user, got: %v", err)
	}
}

func TestAuthService_AccountStatusValidation(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tx := db.DB.Begin()
	defer tx.Rollback()

	authSvc, _ := setupTestAuthService(t, tx, nil, nil)
	ctx := context.Background()

	// Register user
	regReq := services.RegisterRequest{
		Name:     "Status User",
		Email:    "status.user@mindora.app",
		Password: "SecurePassword123!",
	}
	res, err := authSvc.Register(ctx, regReq, "127.0.0.1")
	if err != nil {
		t.Fatalf("setup registration failed: %v", err)
	}

	// Test Inactive Status
	if err := tx.Model(&models.User{}).Where("id = ?", res.User.ID).Update("status", models.UserStatusInactive).Error; err != nil {
		t.Fatalf("failed to set status inactive: %v", err)
	}

	loginReq := services.LoginRequest{
		Email:    "status.user@mindora.app",
		Password: "SecurePassword123!",
	}
	_, err = authSvc.Login(ctx, loginReq, "127.0.0.1", "TestAgent")
	if !errors.Is(err, services.ErrAccountNotActive) {
		t.Errorf("expected ErrAccountNotActive, got: %v", err)
	}

	// Test Suspended Status
	if err := tx.Model(&models.User{}).Where("id = ?", res.User.ID).Update("status", models.UserStatusSuspended).Error; err != nil {
		t.Fatalf("failed to set status suspended: %v", err)
	}
	_, err = authSvc.Login(ctx, loginReq, "127.0.0.1", "TestAgent")
	if !errors.Is(err, services.ErrAccountSuspended) {
		t.Errorf("expected ErrAccountSuspended, got: %v", err)
	}
}

func TestAuthService_RefreshTokenRotationAndReuseDetection(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tx := db.DB.Begin()
	defer tx.Rollback()

	authSvc, _ := setupTestAuthService(t, tx, nil, nil)
	ctx := context.Background()

	// Register and login
	regReq := services.RegisterRequest{
		Name:     "Rotation User",
		Email:    "rotation@mindora.app",
		Password: "SecurePassword123!",
	}
	authRes, err := authSvc.Register(ctx, regReq, "127.0.0.1")
	if err != nil {
		t.Fatalf("setup register failed: %v", err)
	}

	originalRefresh := authRes.RefreshToken

	// 1. First Refresh (Rotation)
	rotatedRes, err := authSvc.Refresh(ctx, originalRefresh, "op-legit-1", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("first refresh failed: %v", err)
	}

	if rotatedRes.RefreshToken == originalRefresh {
		t.Errorf("expected rotated new refresh token, got identical token")
	}

	// 2. TOKEN REUSE DETECTION: Present old rotated refresh token
	_, err = authSvc.Refresh(ctx, originalRefresh, "op-attacker", "127.0.0.1", "AttackerAgent")
	if !errors.Is(err, services.ErrTokenReuseDetected) {
		t.Fatalf("expected ErrTokenReuseDetected when replaying old refresh token, got: %v", err)
	}

	// 3. Verify that the session family was revoked: Even the new rotated token should now be invalidated
	_, err = authSvc.Refresh(ctx, rotatedRes.RefreshToken, "op-legit-2", "127.0.0.1", "LegitAgent")
	if !errors.Is(err, services.ErrTokenReuseDetected) && !errors.Is(err, services.ErrInvalidRefreshToken) {
		t.Errorf("expected invalidated session after reuse detection, got: %v", err)
	}
}

func TestAuthService_GoogleOAuth_Scenarios(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	tx := db.DB.Begin()
	defer tx.Rollback()

	// Base mock token with id_token extra
	rawOauthTok := (&oauth2.Token{
		AccessToken: "mock-google-access-token",
		TokenType:   "Bearer",
		Expiry:      time.Now().Add(1 * time.Hour),
	}).WithExtra(map[string]interface{}{
		"id_token": "valid.mock.id.token",
	})

	mockExchanger := &MockTokenExchanger{Token: rawOauthTok}

	// SCENARIO 1: Brand New Google User Registration
	newGoogleVerifier := &MockGoogleVerifier{
		Profile: &oauth.GoogleProfile{
			Subject:       "google-sub-alice-123",
			Email:         "alice.google@mindora.app",
			EmailVerified: true,
			Name:          "Alice Google",
		},
	}

	authSvc, _ := setupTestAuthService(t, tx, newGoogleVerifier, mockExchanger)
	ctx := context.Background()

	const validPKCEVerifier = "test-pkce-verifier-cryptographically-secure-string-43-chars"

	result, err := authSvc.HandleGoogleCallback(
		ctx,
		"valid-code",
		"state-abc",
		"state-abc",
		"nonce-xyz",
		validPKCEVerifier,
		"127.0.0.1",
		"TestAgent",
	)
	if err != nil {
		t.Fatalf("google new user registration failed: %v", err)
	}

	if result.User.Email != "alice.google@mindora.app" {
		t.Errorf("expected user email alice.google@mindora.app, got %s", result.User.Email)
	}
	if result.User.Role != "user" {
		t.Errorf("expected role 'user', got %s", result.User.Role)
	}

	// SCENARIO 2: Existing Google Identity Direct Login
	// Call again with same Google sub
	loginResult, err := authSvc.HandleGoogleCallback(
		ctx,
		"valid-code-2",
		"state-abc",
		"state-abc",
		"nonce-xyz",
		validPKCEVerifier,
		"127.0.0.1",
		"TestAgent",
	)
	if err != nil {
		t.Fatalf("google existing user login failed: %v", err)
	}
	if loginResult.User.ID != result.User.ID {
		t.Errorf("expected existing user ID %s, got %s", result.User.ID, loginResult.User.ID)
	}

	// SCENARIO 3: STRICT NO-ACCOUNT-LINKING POLICY
	// Pre-create standard email user "bob.password@mindora.app"
	userRepo := repositories.NewUserRepository(tx)
	userRole, _ := userRepo.GetRoleByName(ctx, "user")
	hasher := hash.NewBcryptHasher(10)
	pwdHash, _ := hasher.HashPassword("SecurePassword123!")

	existingEmailUser := &models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         "Bob Password User",
		Email:        "bob.password@mindora.app",
		PasswordHash: &pwdHash,
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}
	if err := userRepo.Create(ctx, existingEmailUser); err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	// A Google user with the SAME email but new Google sub tries to log in
	conflictVerifier := &MockGoogleVerifier{
		Profile: &oauth.GoogleProfile{
			Subject:       "google-sub-different-bob-999",
			Email:         "bob.password@mindora.app", // Matches existing user
			EmailVerified: true,
			Name:          "Bob Different Google",
		},
	}
	conflictAuthSvc, _ := setupTestAuthService(t, tx, conflictVerifier, mockExchanger)

	_, err = conflictAuthSvc.HandleGoogleCallback(
		ctx,
		"valid-code-3",
		"state-abc",
		"state-abc",
		"nonce-xyz",
		validPKCEVerifier,
		"127.0.0.1",
		"TestAgent",
	)

	if !errors.Is(err, services.ErrEmailConflictExternalIdentity) {
		t.Fatalf("expected ErrEmailConflictExternalIdentity for conflicting email, got: %v", err)
	}

	// SCENARIO 4: Unverified Google Email
	unverifiedVerifier := &MockGoogleVerifier{
		Profile: &oauth.GoogleProfile{
			Subject:       "google-sub-unverified-555",
			Email:         "unverified@mindora.app",
			EmailVerified: false, // Unverified
			Name:          "Unverified User",
		},
	}
	unverifiedAuthSvc, _ := setupTestAuthService(t, tx, unverifiedVerifier, mockExchanger)

	_, err = unverifiedAuthSvc.HandleGoogleCallback(
		ctx,
		"valid-code-4",
		"state-abc",
		"state-abc",
		"nonce-xyz",
		validPKCEVerifier,
		"127.0.0.1",
		"TestAgent",
	)
	if !errors.Is(err, oauth.ErrUnverifiedEmail) {
		t.Errorf("expected ErrUnverifiedEmail, got: %v", err)
	}

	// SCENARIO 5: Invalid OAuth State
	_, err = authSvc.HandleGoogleCallback(
		ctx,
		"valid-code",
		"tampered-state",
		"original-state",
		"nonce",
		validPKCEVerifier,
		"127.0.0.1",
		"TestAgent",
	)
	if !errors.Is(err, oauth.ErrInvalidState) {
		t.Errorf("expected ErrInvalidState, got: %v", err)
	}

	// SCENARIO 6: Missing PKCE Verifier
	_, err = authSvc.HandleGoogleCallback(
		ctx,
		"valid-code",
		"state-abc",
		"state-abc",
		"nonce",
		"", // Empty PKCE verifier
		"127.0.0.1",
		"TestAgent",
	)
	if !errors.Is(err, oauth.ErrInvalidPKCE) {
		t.Errorf("expected ErrInvalidPKCE when code_verifier is missing, got: %v", err)
	}
}
