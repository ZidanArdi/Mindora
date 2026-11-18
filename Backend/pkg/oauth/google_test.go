package oauth_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"net/url"
	"testing"

	"golang.org/x/oauth2"
	"mindora/backend/pkg/oauth"
)

type mockExchangerWithRecording struct {
	recordedOpts []oauth2.AuthCodeOption
	token        *oauth2.Token
	err          error
}

func (m *mockExchangerWithRecording) Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error) {
	m.recordedOpts = opts
	return m.token, m.err
}

func TestGeneratePKCE(t *testing.T) {
	verifier, challenge, err := oauth.GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE failed: %v", err)
	}

	// RFC 7636: code_verifier must be at least 43 characters and at most 128 characters
	if len(verifier) < 43 || len(verifier) > 128 {
		t.Errorf("expected verifier length between 43 and 128 chars, got %d", len(verifier))
	}

	// Verify that challenge matches SHA-256 of verifier base64url-encoded without padding
	h := sha256.Sum256([]byte(verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(h[:])
	if challenge != expectedChallenge {
		t.Errorf("expected challenge %s, got %s", expectedChallenge, challenge)
	}
}

func TestGetAuthURL_IncludesPKCE(t *testing.T) {
	svc := oauth.NewGoogleService(oauth.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Enabled:      true,
	}, nil)

	state := "test-state-12345"
	nonce := "test-nonce-67890"
	challenge := "test-challenge-s256"

	authURL, err := svc.GetAuthURL(state, nonce, challenge)
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}

	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse authURL: %v", err)
	}

	query := parsedURL.Query()
	if query.Get("state") != state {
		t.Errorf("expected state %s, got %s", state, query.Get("state"))
	}
	if query.Get("nonce") != nonce {
		t.Errorf("expected nonce %s, got %s", nonce, query.Get("nonce"))
	}
	if query.Get("code_challenge") != challenge {
		t.Errorf("expected code_challenge %s, got %s", challenge, query.Get("code_challenge"))
	}
	if query.Get("code_challenge_method") != "S256" {
		t.Errorf("expected code_challenge_method S256, got %s", query.Get("code_challenge_method"))
	}
}

func TestExchangeCode_TransmitsVerifier(t *testing.T) {
	mockExch := &mockExchangerWithRecording{
		token: &oauth2.Token{AccessToken: "mock-token"},
	}

	svc := oauth.NewGoogleServiceWithExchanger(oauth.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost:8080/callback",
		Enabled:      true,
	}, nil, mockExch)

	code := "mock-auth-code"
	verifier := "mock-code-verifier-43-characters-long-string-pkce"

	_, err := svc.ExchangeCode(context.Background(), code, verifier)
	if err != nil {
		t.Fatalf("ExchangeCode failed: %v", err)
	}

	if len(mockExch.recordedOpts) == 0 {
		t.Errorf("expected ExchangeCode to pass AuthCodeOptions with code_verifier")
	}
}
