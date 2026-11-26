package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/response"
	"mindora/backend/pkg/turnstile"
)

func TestTurnstileAuthFlow_MockedCloudflare(t *testing.T) {
	// 1. Setup mock Cloudflare Turnstile siteverify server
	mockCloudflare := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		_ = r.ParseForm()
		secret := r.FormValue("secret")
		token := r.FormValue("response")

		if secret != "mock-turnstile-secret" {
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:    false,
				ErrorCodes: []string{"invalid-input-secret"},
			})
			return
		}

		switch token {
		case "valid-register-token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:     true,
				ChallengeTS: time.Now(),
				Hostname:    "localhost",
				Action:      "register",
			})
		case "valid-login-token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:     true,
				ChallengeTS: time.Now(),
				Hostname:    "localhost",
				Action:      "login",
			})
		case "mismatched-action-token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:     true,
				ChallengeTS: time.Now(),
				Hostname:    "localhost",
				Action:      "wrong_action",
			})
		case "slow-timeout-token":
			// Simulate slow or unresponsive Cloudflare siteverify server
			time.Sleep(300 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success: true,
			})
		case "invalid-token":
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:    false,
				ErrorCodes: []string{"invalid-input-response"},
			})
		default:
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(turnstile.SiteverifyResponse{
				Success:    false,
				ErrorCodes: []string{"invalid-input-response"},
			})
		}
	}))
	defer mockCloudflare.Close()

	// 2. Setup base configuration pointing to mock server with Turnstile strictly enforced
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = true
	cfg.Turnstile.SecretKey = "mock-turnstile-secret"
	cfg.Turnstile.VerifyURL = mockCloudflare.URL
	cfg.Turnstile.ExpectedHostname = "localhost"
	cfg.Turnstile.AllowTestTokens = false
	cfg.Env = "production"

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	testPrefix := "turnstile.test." + time.Now().Format("150405000")
	testEmail := testPrefix + "@mindora.app"
	testPassword := "SecurePass123!"

	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'turnstile.test%')")
		_ = db.DB.Exec("DELETE FROM auth_identities WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'turnstile.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'turnstile.test%'")
	}()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// -------------------------------------------------------------
	// Requirement 8: Missing Turnstile token is rejected
	// -------------------------------------------------------------
	t.Run("Register_MissingToken_RejectsWithForbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":     "Turnstile User",
			"email":    "missing." + testEmail,
			"password": testPassword,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected status 403 Forbidden for missing turnstile token, got: %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// Requirement 9: Invalid Turnstile token is rejected
	// -------------------------------------------------------------
	t.Run("Register_InvalidToken_RejectsWithForbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":            "Turnstile User",
			"email":           "invalid." + testEmail,
			"password":        testPassword,
			"turnstile_token": "invalid-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected status 403 Forbidden for invalid turnstile token, got: %d", resp.StatusCode)
		}
	})

	t.Run("Register_ActionMismatch_RejectsWithForbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":            "Turnstile User",
			"email":           "mismatch." + testEmail,
			"password":        testPassword,
			"turnstile_token": "mismatched-action-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected status 403 Forbidden for action mismatch turnstile token, got: %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------
	// Requirement 10: Valid Turnstile token allows authentication
	// -------------------------------------------------------------
	t.Run("Register_ValidToken_Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"name":            "Turnstile Registered User",
			"email":           testEmail,
			"password":        testPassword,
			"turnstile_token": "valid-register-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			respBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 201 Created, got: %d, body: %s", resp.StatusCode, string(respBytes))
		}

		var res response.Response
		_ = json.NewDecoder(resp.Body).Decode(&res)
		if !res.Success {
			t.Errorf("expected response success to be true")
		}
	})

	t.Run("Login_MissingToken_RejectsWithForbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    testEmail,
			"password": testPassword,
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected status 403 Forbidden for missing turnstile token on login, got: %d", resp.StatusCode)
		}
	})

	t.Run("Login_InvalidToken_RejectsWithForbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":           testEmail,
			"password":        testPassword,
			"turnstile_token": "invalid-token",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected status 403 Forbidden for invalid turnstile token on login, got: %d", resp.StatusCode)
		}
	})

	t.Run("Login_ValidToken_Success", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":           testEmail,
			"password":        testPassword,
			"turnstile_token": "valid-login-token",
			"portal":          "user",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			respBytes, _ := io.ReadAll(resp.Body)
			t.Fatalf("expected status 200 OK for valid login, got: %d, body: %s", resp.StatusCode, string(respBytes))
		}
	})

	// -------------------------------------------------------------
	// Requirement 11: Cloudflare verification timeout fails safely (403 Forbidden)
	// -------------------------------------------------------------
	t.Run("Turnstile_Timeout_FailsSafelyWithForbidden", func(t *testing.T) {
		// Create client with tight 50ms timeout pointing to mock server that delays 300ms
		timeoutClient := turnstile.NewClientWithURL(turnstile.Config{
			Enabled:   true,
			SecretKey: "mock-turnstile-secret",
			Timeout:   50 * time.Millisecond,
		}, mockCloudflare.URL)

		err := timeoutClient.Verify(t.Context(), "slow-timeout-token", "127.0.0.1", "login")
		if err == nil {
			t.Fatalf("expected error on timeout, got nil")
		}
		if !turnstile.IsTurnstileError(err) {
			t.Errorf("expected timeout error to be recognized as Turnstile verification error, got: %v", err)
		}
	})

	// -------------------------------------------------------------
	// Requirement 12: Development bypass works only when explicitly configured
	// -------------------------------------------------------------
	t.Run("DevelopmentBypass_WorksOnlyWhenExplicitlyConfigured", func(t *testing.T) {
		// When explicitly configured in development: dummy token passes
		devBypassAllowed := turnstile.NewClient(turnstile.Config{
			Enabled:         true,
			SecretKey:       "mock-turnstile-secret",
			AllowTestTokens: true,
		})
		if err := devBypassAllowed.Verify(t.Context(), "1x0000000000000000000000000000000AA", "127.0.0.1", "login"); err != nil {
			t.Errorf("expected dummy token to pass when AllowTestTokens is true, got: %v", err)
		}

		// When AllowTestTokens is false in development: dummy token is strictly rejected
		devBypassDisabled := turnstile.NewClient(turnstile.Config{
			Enabled:         true,
			SecretKey:       "mock-turnstile-secret",
			AllowTestTokens: false,
		})
		if err := devBypassDisabled.Verify(t.Context(), "1x0000000000000000000000000000000AA", "127.0.0.1", "login"); err == nil {
			t.Errorf("expected dummy token to be rejected when AllowTestTokens is false, got nil")
		}
	})

	// -------------------------------------------------------------
	// Requirement 13: Production configuration cannot silently bypass Turnstile
	// -------------------------------------------------------------
	t.Run("Production_CannotSilentlyBypassTurnstile", func(t *testing.T) {
		// In production, even if someone attempts to configure AllowTestTokens=true:
		prodCfg := config.LoadConfig()
		prodCfg.Env = "production"
		prodCfg.Turnstile.Enabled = true
		prodCfg.Turnstile.SecretKey = "prod-turnstile-secret"
		prodCfg.Turnstile.AllowTestTokens = true // attempt invalid bypass in prod

		// Config validation MUST catch and reject this unsafe configuration
		if err := prodCfg.Validate(); err == nil {
			t.Errorf("SECURITY VULNERABILITY: production config with AllowTestTokens=true passed validation")
		}

		// Production Turnstile client MUST fail closed against test tokens
		prodClient := turnstile.NewClient(turnstile.Config{
			Enabled:         true,
			SecretKey:       "prod-turnstile-secret",
			AllowTestTokens: false, // Enforced fail closed
		})
		if err := prodClient.Verify(t.Context(), "1x0000000000000000000000000000000AA", "127.0.0.1", "login"); err == nil {
			t.Errorf("SECURITY VULNERABILITY: production client accepted dummy token")
		}
	})
}
