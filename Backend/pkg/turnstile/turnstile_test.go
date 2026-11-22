package turnstile_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mindora/backend/pkg/turnstile"
)

func TestTurnstileDisabled(t *testing.T) {
	client := turnstile.NewClient(turnstile.Config{
		Enabled: false,
	})

	if client.IsEnabled() {
		t.Errorf("expected client to be disabled")
	}

	// Should pass even with empty token when disabled
	if err := client.Verify(context.Background(), "", "", ""); err != nil {
		t.Errorf("expected nil error when disabled, got: %v", err)
	}
}

func TestTurnstileDummyToken(t *testing.T) {
	client := turnstile.NewClient(turnstile.Config{
		Enabled:         true,
		SecretKey:       "mock-secret",
		AllowTestTokens: true,
	})

	// Cloudflare test tokens in non-production
	if err := client.Verify(context.Background(), "1x0000000000000000000000000000000AA", "127.0.0.1", "login"); err != nil {
		t.Errorf("expected dummy token to pass when AllowTestTokens is true, got: %v", err)
	}
	if err := client.Verify(context.Background(), "dummy-test-token", "127.0.0.1", "login"); err != nil {
		t.Errorf("expected dummy token to pass when AllowTestTokens is true, got: %v", err)
	}

	// Production mode (AllowTestTokens = false): dummy test tokens must NOT bypass
	prodClient := turnstile.NewClient(turnstile.Config{
		Enabled:         true,
		SecretKey:       "mock-secret",
		AllowTestTokens: false,
	})
	if err := prodClient.Verify(context.Background(), "dummy-test-token", "127.0.0.1", "login"); err == nil {
		t.Errorf("SECURITY VULNERABILITY: dummy-test-token was accepted when AllowTestTokens is false")
	}
}

func TestTurnstileMockServer(t *testing.T) {
	// Setup mock Cloudflare endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.FormValue("response")
		secret := r.FormValue("secret")

		if secret != "valid-secret" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:    false,
				ErrorCodes: []string{"invalid-input-secret"},
			})
			return
		}

		if token == "success-token" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:  true,
				Hostname: "localhost",
				Action:   "login",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
			Success:    false,
			ErrorCodes: []string{"invalid-input-response"},
		})
	}))
	defer ts.Close()

	// Direct test via custom client pointing to mock server
	client := turnstile.NewClientWithURL(turnstile.Config{
		Enabled:          true,
		SecretKey:        "valid-secret",
		ExpectedHostname: "localhost",
		ExpectedAction:   "login",
		Timeout:          2 * time.Second,
	}, ts.URL)

	ctx := context.Background()

	// 1. Successful verification
	if err := client.Verify(ctx, "success-token", "127.0.0.1", "login"); err != nil {
		t.Errorf("expected success, got: %v", err)
	}

	// 2. Failed verification (invalid token)
	if err := client.Verify(ctx, "bad-token", "127.0.0.1", "login"); err == nil {
		t.Errorf("expected error for bad token, got nil")
	}

	// 3. Action mismatch
	if err := client.Verify(ctx, "success-token", "127.0.0.1", "register"); err == nil {
		t.Errorf("expected action mismatch error, got nil")
	}
}
