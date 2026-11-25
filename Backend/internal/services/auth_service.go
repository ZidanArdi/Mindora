package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/sync/singleflight"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/pkg/crypto"
	"mindora/backend/pkg/hash"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/oauth"
	"mindora/backend/pkg/turnstile"
	"mindora/backend/pkg/validator"
)

var (
	ErrInvalidCredentials             = errors.New("invalid email or password")
	ErrEmailAlreadyRegistered         = errors.New("email address is already registered")
	ErrAccountNotActive               = errors.New("account is inactive")
	ErrAccountSuspended               = errors.New("account has been suspended")
	ErrInvalidRefreshToken            = errors.New("invalid or expired refresh token")
	ErrTokenReuseDetected             = errors.New("refresh token reuse detected; all associated sessions have been revoked")
	ErrConcurrentConflict             = errors.New("concurrent refresh in progress; please retry with current token")
	ErrIdempotencyExpired             = errors.New("refresh idempotency window expired")
	ErrEmailConflictExternalIdentity  = errors.New("this email is already registered using another sign-in method; please sign in with your email and password")
	ErrAdminPublicRegistrationBlocked = errors.New("administrative accounts cannot be created via public registration")
	ErrAdminPortalAccessDenied        = errors.New("Akun Anda tidak memiliki izin untuk masuk melalui Konsol Admin.")
	ErrUserPortalAccessDenied         = errors.New("Akun Administrator tidak dapat masuk melalui Ruang Kerja Pengguna.")
	ErrInvalidPortal                  = errors.New("Portal yang diminta tidak valid.")
)

// RegisterRequest represents the user registration payload.
type RegisterRequest struct {
	Name           string `json:"name" validate:"required,min=2,max=100"`
	Email          string `json:"email" validate:"required,email,max=255"`
	Password       string `json:"password" validate:"required,password_strength"`
	TurnstileToken string `json:"turnstile_token"`
}

// LoginRequest represents the user login payload.
type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	TurnstileToken string `json:"turnstile_token"`
	Portal         string `json:"portal"`
}

// UserResponse is the safe public representation of a User.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResult is returned upon successful authentication.
type AuthResult struct {
	AccessToken  string       `json:"access_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"` // in seconds
	RefreshToken string       `json:"refresh_token,omitempty"`
	User         UserResponse `json:"user"`
}

// AuthService defines the authentication business logic operations.
type AuthService interface {
	Register(ctx context.Context, req RegisterRequest, remoteIP string) (*AuthResult, error)
	Login(ctx context.Context, req LoginRequest, remoteIP, userAgent string) (*AuthResult, error)
	Refresh(ctx context.Context, rawRefreshToken, operationID, remoteIP, userAgent string) (*AuthResult, error)
	Logout(ctx context.Context, rawRefreshToken string) error
	HandleGoogleCallback(ctx context.Context, code, state, expectedState, expectedNonce, codeVerifier, remoteIP, userAgent string) (*AuthResult, error)
	SetEncryptionKey(key []byte)
}

type authService struct {
	userRepo      repositories.UserRepository
	identityRepo  repositories.IdentityRepository
	sessionRepo   repositories.SessionRepository
	hasher        hash.Hasher
	tokenManager  *jwt.TokenManager
	turnstile     turnstile.Verifier
	validator     *validator.CustomValidator
	googleService *oauth.GoogleService
	refreshFlight singleflight.Group
	encryptionKey []byte
}

// NewAuthService instantiates a new AuthService.
func NewAuthService(
	userRepo repositories.UserRepository,
	identityRepo repositories.IdentityRepository,
	sessionRepo repositories.SessionRepository,
	hasher hash.Hasher,
	tokenManager *jwt.TokenManager,
	turnstileVerifier turnstile.Verifier,
	val *validator.CustomValidator,
	googleService *oauth.GoogleService,
) AuthService {
	// Derive default 32-byte authenticated encryption key
	defaultKey := crypto.DeriveKey([]byte("mindora-default-session-encryption-salt-2026"))

	return &authService{
		userRepo:      userRepo,
		identityRepo:  identityRepo,
		sessionRepo:   sessionRepo,
		hasher:        hasher,
		tokenManager:  tokenManager,
		turnstile:     turnstileVerifier,
		validator:     val,
		googleService: googleService,
		encryptionKey: defaultKey,
	}
}

// SetEncryptionKey allows configuring a dedicated secret key for AES-256-GCM idempotency encryption.
func (s *authService) SetEncryptionKey(key []byte) {
	if len(key) > 0 {
		s.encryptionKey = crypto.DeriveKey(key)
	}
}

// Register registers a new student user account.
func (s *authService) Register(ctx context.Context, req RegisterRequest, remoteIP string) (*AuthResult, error) {
	// 1. Validate DTO constraints
	if err := s.validator.Validate(req); err != nil {
		return nil, err
	}

	// 2. Verify Cloudflare Turnstile token
	if err := s.turnstile.Verify(ctx, req.TurnstileToken, remoteIP, "register"); err != nil {
		return nil, err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// 3. Check for existing email reservation
	existingUser, err := s.userRepo.FindByEmail(ctx, normalizedEmail)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyRegistered
	}
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, fmt.Errorf("database query error: %w", err)
	}

	// 4. Hash password with bcrypt
	pwdHash, err := s.hasher.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Look up default role "user"
	userRole, err := s.userRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("system user role unavailable: %w", err)
	}

	now := time.Now().UTC()
	newUser := &models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         strings.TrimSpace(req.Name),
		Email:        normalizedEmail,
		PasswordHash: &pwdHash,
		Status:       models.UserStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
		Role:         *userRole,
	}

	// 6. Create user in database
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	// 7. Issue JWT Access Token & Refresh Token
	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(newUser.ID, newUser.Email, userRole.Name)
	if err != nil {
		return nil, err
	}

	rawRefresh, tokenHash, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	session := &models.AuthSession{
		ID:        uuid.New(),
		UserID:    newUser.ID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		IPAddress: &remoteIP,
		ExpiresAt: refreshExpiresAt,
		CreatedAt: now,
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session record: %w", err)
	}

	expiresIn := int64(time.Until(accessExpiresAt).Seconds())

	return &AuthResult{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: rawRefresh,
		User: UserResponse{
			ID:        newUser.ID,
			Name:      newUser.Name,
			Email:     newUser.Email,
			Role:      userRole.Name,
			Status:    newUser.Status,
			CreatedAt: newUser.CreatedAt,
		},
	}, nil
}

// Login authenticates a user via email and password.
func (s *authService) Login(ctx context.Context, req LoginRequest, remoteIP, userAgent string) (*AuthResult, error) {
	// 1. Validate DTO
	if err := s.validator.Validate(req); err != nil {
		return nil, err
	}

	// 2. Verify Cloudflare Turnstile token
	if err := s.turnstile.Verify(ctx, req.TurnstileToken, remoteIP, "login"); err != nil {
		return nil, err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))

	// 3. Find user account
	user, err := s.userRepo.FindByEmail(ctx, normalizedEmail)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// 4. Validate account status
	if user.Status == models.UserStatusInactive {
		return nil, ErrAccountNotActive
	}
	if user.Status == models.UserStatusSuspended {
		return nil, ErrAccountSuspended
	}
	if user.Status != models.UserStatusActive {
		return nil, ErrInvalidCredentials
	}

	// 5. Verify password hash
	if user.PasswordHash == nil {
		return nil, ErrInvalidCredentials
	}
	if err := s.hasher.ComparePassword(*user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 6. Enforce Strict Server-Side Portal Authorization Policy
	// The portal requested by the client is only the destination.
	// We strictly evaluate the actual authenticated account role against the portal matrix.
	userRole := strings.ToLower(strings.TrimSpace(user.Role.Name))
	if userRole != "user" && userRole != "admin" {
		return nil, ErrInvalidCredentials
	}

	requestedPortal := strings.ToLower(strings.TrimSpace(req.Portal))
	if requestedPortal != "" {
		switch requestedPortal {
		case "admin":
			if userRole != "admin" {
				return nil, ErrAdminPortalAccessDenied
			}
		case "user":
			if userRole != "user" {
				return nil, ErrUserPortalAccessDenied
			}
		default:
			return nil, ErrInvalidPortal
		}
	}

	// 7. Issue Access and Refresh tokens
	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID, user.Email, user.Role.Name)
	if err != nil {
		return nil, err
	}

	rawRefresh, tokenHash, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	familyID := uuid.New()
	session := &models.AuthSession{
		ID:        uuid.New(),
		UserID:    user.ID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		IPAddress: &remoteIP,
		UserAgent: &userAgent,
		ExpiresAt: refreshExpiresAt,
		CreatedAt: now,
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to persist session: %w", err)
	}

	expiresIn := int64(time.Until(accessExpiresAt).Seconds())

	return &AuthResult{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		RefreshToken: rawRefresh,
		User: UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			Role:      user.Role.Name,
			Status:    user.Status,
			CreatedAt: user.CreatedAt,
		},
	}, nil
}

// Refresh rotates the refresh token and issues a new access token.
// Distributed synchronization is enforced via PostgreSQL Transactional Advisory Locks (pg_advisory_xact_lock).
// Retries with the same operation identifier return the identical rotation result idempotently.
// Different operation identifiers presenting an already-rotated token are denied (409 Conflict if concurrent,
// or full family revocation if sequential reuse is detected).
func (s *authService) Refresh(ctx context.Context, rawRefreshToken, operationID, remoteIP, userAgent string) (*AuthResult, error) {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil, ErrInvalidRefreshToken
	}

	trimmedOpID := strings.TrimSpace(operationID)
	if trimmedOpID == "" {
		trimmedOpID = uuid.New().String()
	}

	tokenHash := jwt.HashToken(rawRefreshToken)

	// Singleflight optimizes concurrency locally within the replica
	flightKey := tokenHash + ":" + trimmedOpID
	v, err, _ := s.refreshFlight.Do(flightKey, func() (interface{}, error) {
		return s.executeRefreshDistributed(ctx, tokenHash, trimmedOpID, remoteIP, userAgent)
	})

	if err != nil {
		return nil, err
	}

	res, ok := v.(*AuthResult)
	if !ok || res == nil {
		return nil, ErrInvalidRefreshToken
	}

	return res, nil
}

func (s *authService) executeRefreshDistributed(ctx context.Context, tokenHash, operationID, remoteIP, userAgent string) (*AuthResult, error) {
	now := time.Now().UTC()

	params := repositories.RotateDistributedParams{
		TokenHash:   tokenHash,
		OperationID: operationID,
		ResponseTTL: 60 * time.Second,
		EncryptFunc: func(plaintext []byte) (string, error) {
			return crypto.EncryptAESGCM(plaintext, s.encryptionKey)
		},
		DecryptFunc: func(ciphertextB64 string) ([]byte, error) {
			return crypto.DecryptAESGCM(ciphertextB64, s.encryptionKey)
		},
		OnRotate: func(oldSession *models.AuthSession) (*models.AuthSession, []byte, error) {
			// Verify expiration
			if oldSession.ExpiresAt.Before(now) {
				return nil, nil, ErrInvalidRefreshToken
			}

			// Validate account status
			if oldSession.User.Status != models.UserStatusActive {
				return nil, nil, ErrAccountNotActive
			}

			// Generate replacement refresh token
			newRawRefresh, newTokenHash, newExpiresAt, err := s.tokenManager.GenerateRefreshToken()
			if err != nil {
				return nil, nil, err
			}

			newSession := &models.AuthSession{
				ID:        uuid.New(),
				UserID:    oldSession.UserID,
				FamilyID:  oldSession.FamilyID,
				TokenHash: newTokenHash,
				IPAddress: &remoteIP,
				UserAgent: &userAgent,
				ExpiresAt: newExpiresAt,
				CreatedAt: now,
			}

			// Issue new access token
			accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(
				oldSession.User.ID,
				oldSession.User.Email,
				oldSession.User.Role.Name,
			)
			if err != nil {
				return nil, nil, err
			}

			expiresIn := int64(time.Until(accessExpiresAt).Seconds())

			result := &AuthResult{
				AccessToken:  accessToken,
				TokenType:    "Bearer",
				ExpiresIn:    expiresIn,
				RefreshToken: newRawRefresh,
				User: UserResponse{
					ID:        oldSession.User.ID,
					Name:      oldSession.User.Name,
					Email:     oldSession.User.Email,
					Role:      oldSession.User.Role.Name,
					Status:    oldSession.User.Status,
					CreatedAt: oldSession.User.CreatedAt,
				},
			}

			payloadBytes, err := json.Marshal(result)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to marshal auth result: %w", err)
			}

			return newSession, payloadBytes, nil
		},
	}

	payload, err := s.sessionRepo.RotateSessionDistributed(ctx, params)
	if err != nil {
		if errors.Is(err, repositories.ErrSessionNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		if errors.Is(err, repositories.ErrSessionAlreadyRotated) {
			return nil, ErrInvalidRefreshToken
		}
		if errors.Is(err, repositories.ErrConcurrentRefreshConflict) {
			return nil, ErrConcurrentConflict
		}
		if errors.Is(err, repositories.ErrTokenReuseDetected) {
			return nil, ErrTokenReuseDetected
		}
		if errors.Is(err, repositories.ErrIdempotencyExpired) {
			return nil, ErrIdempotencyExpired
		}
		return nil, err
	}

	var res AuthResult
	if err := json.Unmarshal(payload, &res); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rotation response payload: %w", err)
	}

	return &res, nil
}

// Logout revokes the session associated with the provided refresh token.
func (s *authService) Logout(ctx context.Context, rawRefreshToken string) error {
	if strings.TrimSpace(rawRefreshToken) == "" {
		return nil
	}

	tokenHash := jwt.HashToken(rawRefreshToken)
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokenHash)
	if err != nil {
		// If session does not exist, nothing to revoke
		return nil
	}

	return s.sessionRepo.RevokeSession(ctx, session.ID)
}

// HandleGoogleCallback processes the OAuth code exchange, validates the ID token,
// enforces PKCE verification, enforces the No-Account-Linking policy, and authenticates or registers the user.
func (s *authService) HandleGoogleCallback(
	ctx context.Context,
	code, state, expectedState, expectedNonce, codeVerifier, remoteIP, userAgent string,
) (*AuthResult, error) {
	// 1. Verify CSRF State token
	if state == "" || state != expectedState {
		return nil, oauth.ErrInvalidState
	}

	// 2. Enforce PKCE verification
	if strings.TrimSpace(codeVerifier) == "" {
		return nil, oauth.ErrInvalidPKCE
	}

	// 2. Exchange authorization code for tokens with PKCE code_verifier
	token, err := s.googleService.ExchangeCode(ctx, code, codeVerifier)
	if err != nil {
		return nil, err
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return nil, oauth.ErrInvalidIDToken
	}

	// 3. Validate Google ID Token cryptographically
	profile, err := s.googleService.VerifyIDToken(ctx, rawIDToken, expectedNonce)
	if err != nil {
		return nil, err
	}

	if !profile.EmailVerified {
		return nil, oauth.ErrUnverifiedEmail
	}

	// 4. Check if Google identity is already mapped
	identity, err := s.identityRepo.FindByIdentity(ctx, models.ProviderGoogle, profile.Subject)
	if err != nil && !errors.Is(err, repositories.ErrIdentityNotFound) {
		return nil, fmt.Errorf("identity lookup failed: %w", err)
	}

	now := time.Now().UTC()

	// SCENARIO 1: Existing Google Identity -> Direct Login
	if identity != nil {
		user := &identity.User
		if user.Status == models.UserStatusInactive {
			return nil, ErrAccountNotActive
		}
		if user.Status == models.UserStatusSuspended {
			return nil, ErrAccountSuspended
		}
		if user.Status != models.UserStatusActive {
			return nil, ErrInvalidCredentials
		}

		// Google Sign-In is strictly authorized for user workspace; administrator accounts cannot sign in via Google
		if strings.ToLower(strings.TrimSpace(user.Role.Name)) != "user" {
			return nil, ErrUserPortalAccessDenied
		}

		// Issue tokens
		accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(user.ID, user.Email, user.Role.Name)
		if err != nil {
			return nil, err
		}

		rawRefresh, tokenHash, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken()
		if err != nil {
			return nil, err
		}

		familyID := uuid.New()
		session := &models.AuthSession{
			ID:        uuid.New(),
			UserID:    user.ID,
			FamilyID:  familyID,
			TokenHash: tokenHash,
			IPAddress: &remoteIP,
			UserAgent: &userAgent,
			ExpiresAt: refreshExpiresAt,
			CreatedAt: now,
		}

		if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
			return nil, fmt.Errorf("failed to create session: %w", err)
		}

		return &AuthResult{
			AccessToken:  accessToken,
			TokenType:    "Bearer",
			ExpiresIn:    int64(time.Until(accessExpiresAt).Seconds()),
			RefreshToken: rawRefresh,
			User: UserResponse{
				ID:        user.ID,
				Name:      user.Name,
				Email:     user.Email,
				Role:      user.Role.Name,
				Status:    user.Status,
				CreatedAt: user.CreatedAt,
			},
		}, nil
	}

	// SCENARIO 2: Google Identity not found.
	// Check if an account already exists with the same email.
	existingUser, err := s.userRepo.FindByEmail(ctx, profile.Email)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		return nil, fmt.Errorf("user email check failed: %w", err)
	}

	// STRICT ACCOUNT LINKING POLICY: If email exists without matching Google identity,
	// DO NOT auto-link. Reject with clear message instructing user to use existing sign-in method.
	if existingUser != nil {
		return nil, ErrEmailConflictExternalIdentity
	}

	// SCENARIO 3: Brand new Google user -> Registration
	userRole, err := s.userRepo.GetRoleByName(ctx, "user")
	if err != nil {
		return nil, fmt.Errorf("user role not found: %w", err)
	}

	userName := strings.TrimSpace(profile.Name)
	if userName == "" {
		userName = strings.Split(profile.Email, "@")[0]
	}

	newUser := &models.User{
		ID:              uuid.New(),
		RoleID:          userRole.ID,
		Name:            userName,
		Email:           profile.Email,
		PasswordHash:    nil, // Google OAuth accounts do not have local passwords
		Status:          models.UserStatusActive,
		EmailVerifiedAt: &now,
		CreatedAt:       now,
		UpdatedAt:       now,
		Role:            *userRole,
	}

	newIdentity := &models.AuthIdentity{
		ID:              uuid.New(),
		UserID:          newUser.ID,
		Provider:        models.ProviderGoogle,
		ProviderSubject: profile.Subject,
		ProviderEmail:   &profile.Email,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Create User and Identity in database
	if err := s.userRepo.Create(ctx, newUser); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	if err := s.identityRepo.Create(ctx, newIdentity); err != nil {
		return nil, fmt.Errorf("failed to create auth identity: %w", err)
	}

	// Issue tokens
	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccessToken(newUser.ID, newUser.Email, userRole.Name)
	if err != nil {
		return nil, err
	}

	rawRefresh, tokenHash, refreshExpiresAt, err := s.tokenManager.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	familyID := uuid.New()
	session := &models.AuthSession{
		ID:        uuid.New(),
		UserID:    newUser.ID,
		FamilyID:  familyID,
		TokenHash: tokenHash,
		IPAddress: &remoteIP,
		UserAgent: &userAgent,
		ExpiresAt: refreshExpiresAt,
		CreatedAt: now,
	}

	if err := s.sessionRepo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &AuthResult{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(time.Until(accessExpiresAt).Seconds()),
		RefreshToken: rawRefresh,
		User: UserResponse{
			ID:        newUser.ID,
			Name:      newUser.Name,
			Email:     newUser.Email,
			Role:      userRole.Name,
			Status:    newUser.Status,
			CreatedAt: newUser.CreatedAt,
		},
	}, nil
}
