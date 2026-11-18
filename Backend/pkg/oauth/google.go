package oauth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var (
	ErrOAuthDisabled      = errors.New("google oauth is disabled or not configured")
	ErrInvalidState       = errors.New("invalid or missing oauth state parameter")
	ErrInvalidPKCE        = errors.New("invalid or missing pkce code verifier")
	ErrCodeExchangeFailed = errors.New("failed to exchange authorization code with google")
	ErrInvalidIDToken     = errors.New("invalid google id token")
	ErrUnverifiedEmail    = errors.New("google account email is not verified")
	ErrNonceMismatch      = errors.New("id token nonce mismatch")
)

// GoogleProfile contains verified identity claims extracted from a Google OIDC ID token.
type GoogleProfile struct {
	Subject       string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

// Config holds Google OAuth 2.0 / OIDC credentials and settings.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Enabled      bool
}

// IDTokenVerifier defines an interface for validating Google ID tokens.
type IDTokenVerifier interface {
	VerifyIDToken(ctx context.Context, idTokenStr, expectedNonce string) (*GoogleProfile, error)
}

// RealGoogleVerifier validates ID tokens using Google's public keys.
type RealGoogleVerifier struct {
	clientID string
}

// NewRealGoogleVerifier initializes a real Google ID token verifier.
func NewRealGoogleVerifier(clientID string) *RealGoogleVerifier {
	return &RealGoogleVerifier{clientID: clientID}
}

// VerifyIDToken validates signature, issuer, audience, expiration, and claims.
func (v *RealGoogleVerifier) VerifyIDToken(ctx context.Context, idTokenStr, expectedNonce string) (*GoogleProfile, error) {
	payload, err := idtoken.Validate(ctx, idTokenStr, v.clientID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidIDToken, err)
	}

	// Validate issuer
	if payload.Issuer != "https://accounts.google.com" && payload.Issuer != "accounts.google.com" {
		return nil, fmt.Errorf("%w: unexpected issuer %s", ErrInvalidIDToken, payload.Issuer)
	}

	// Validate nonce if provided
	if expectedNonce != "" {
		tokenNonce, _ := payload.Claims["nonce"].(string)
		if tokenNonce != expectedNonce {
			return nil, ErrNonceMismatch
		}
	}

	// Validate email verification claim
	emailVerified, _ := payload.Claims["email_verified"].(bool)
	if !emailVerified {
		return nil, ErrUnverifiedEmail
	}

	email, _ := payload.Claims["email"].(string)
	name, _ := payload.Claims["name"].(string)
	picture, _ := payload.Claims["picture"].(string)

	return &GoogleProfile{
		Subject:       payload.Subject,
		Email:         strings.ToLower(strings.TrimSpace(email)),
		EmailVerified: emailVerified,
		Name:          name,
		Picture:       picture,
	}, nil
}

// TokenExchanger defines an interface for exchanging authorization codes for tokens.
type TokenExchanger interface {
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

// RealTokenExchanger implements TokenExchanger using oauth2.Config.
type RealTokenExchanger struct {
	oauthConfig *oauth2.Config
}

func (e *RealTokenExchanger) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	if e.oauthConfig == nil {
		return nil, ErrOAuthDisabled
	}
	return e.oauthConfig.Exchange(ctx, code, opts...)
}

// GoogleService encapsulates Google OAuth 2.0 operations.
type GoogleService struct {
	cfg         Config
	oauthConfig *oauth2.Config
	verifier    IDTokenVerifier
	exchanger   TokenExchanger
}

// NewGoogleService initializes a new GoogleService.
func NewGoogleService(cfg Config, verifier IDTokenVerifier) *GoogleService {
	return NewGoogleServiceWithExchanger(cfg, verifier, nil)
}

// NewGoogleServiceWithExchanger initializes a GoogleService with custom verifier and exchanger (useful for unit testing).
func NewGoogleServiceWithExchanger(cfg Config, verifier IDTokenVerifier, exchanger TokenExchanger) *GoogleService {
	var oauthConf *oauth2.Config
	if cfg.Enabled && cfg.ClientID != "" && cfg.ClientSecret != "" {
		oauthConf = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     google.Endpoint,
			Scopes: []string{
				"openid",
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		}
	}

	if verifier == nil {
		verifier = NewRealGoogleVerifier(cfg.ClientID)
	}

	if exchanger == nil {
		exchanger = &RealTokenExchanger{oauthConfig: oauthConf}
	}

	return &GoogleService{
		cfg:         cfg,
		oauthConfig: oauthConf,
		verifier:    verifier,
		exchanger:   exchanger,
	}
}

// IsEnabled returns whether Google OAuth is operational.
func (s *GoogleService) IsEnabled() bool {
	return s.cfg.Enabled && s.oauthConfig != nil
}

// GenerateRandomString creates cryptographically secure random url-safe string.
func GenerateRandomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// GeneratePKCE creates an RFC 7636 compliant cryptographically random code_verifier
// and calculates its SHA-256 S256 code_challenge.
func GeneratePKCE() (verifier, challenge string, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes for pkce: %w", err)
	}
	verifier = base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(verifier))
	challenge = base64.RawURLEncoding.EncodeToString(h[:])
	return verifier, challenge, nil
}

// GetAuthURL generates the Google login consent redirect URL with state, nonce, and PKCE challenge.
func (s *GoogleService) GetAuthURL(state, nonce, codeChallenge string) (string, error) {
	if !s.IsEnabled() {
		return "", ErrOAuthDisabled
	}

	opts := []oauth2.AuthCodeOption{
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("nonce", nonce),
		oauth2.SetAuthURLParam("prompt", "select_account"),
	}

	if codeChallenge != "" {
		opts = append(opts,
			oauth2.SetAuthURLParam("code_challenge", codeChallenge),
			oauth2.SetAuthURLParam("code_challenge_method", "S256"),
		)
	}

	return s.oauthConfig.AuthCodeURL(state, opts...), nil
}

// ExchangeCode exchanges the authorization code for tokens, transmitting the PKCE code_verifier.
func (s *GoogleService) ExchangeCode(ctx context.Context, code, codeVerifier string) (*oauth2.Token, error) {
	if !s.IsEnabled() {
		return nil, ErrOAuthDisabled
	}

	var opts []oauth2.AuthCodeOption
	if codeVerifier != "" {
		opts = append(opts, oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	}

	token, err := s.exchanger.Exchange(ctx, code, opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCodeExchangeFailed, err)
	}

	return token, nil
}

// VerifyIDToken verifies the raw ID token string using the configured verifier.
func (s *GoogleService) VerifyIDToken(ctx context.Context, idTokenStr, expectedNonce string) (*GoogleProfile, error) {
	return s.verifier.VerifyIDToken(ctx, idTokenStr, expectedNonce)
}
