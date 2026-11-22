package handlers_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/response"
)

func TestAuthHTTPFlow(t *testing.T) {
	cfg := config.LoadConfig()
	// Disable Turnstile in test config
	cfg.Turnstile.Enabled = false
	cfg.GoogleOAuth.Enabled = true
	cfg.GoogleOAuth.ClientID = "test-client-id"
	cfg.GoogleOAuth.ClientSecret = "test-client-secret"

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'http.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'http.test%'")
	}()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	uniqueSuffix := time.Now().Format("150405.000")
	testEmail := "http.test." + strings.ReplaceAll(uniqueSuffix, ".", "") + "@mindora.app"
	testPassword := "SecurePassword123!"

	// 1. Test POST /api/v1/auth/register
	regBody, _ := json.Marshal(map[string]interface{}{
		"name":     "HTTP Test User",
		"email":    testEmail,
		"password": testPassword,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}

	if resp.StatusCode != fiber.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected status 201 Created, got %d: %s", resp.StatusCode, string(body))
	}

	// Verify Set-Cookie header exists for refresh_token
	cookieHeader := resp.Header.Get("Set-Cookie")
	if !strings.Contains(cookieHeader, "refresh_token=") || !strings.Contains(cookieHeader, "HttpOnly") {
		t.Errorf("expected HttpOnly refresh_token cookie, got: %s", cookieHeader)
	}

	var regResp response.Response
	_ = json.NewDecoder(resp.Body).Decode(&regResp)
	if !regResp.Success {
		t.Errorf("expected success true, got false")
	}

	// 2. Test Duplicate Registration (HTTP 409 Conflict)
	dupReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	dupReq.Header.Set("Content-Type", "application/json")
	dupResp, err := app.Test(dupReq, 5000)
	if err != nil {
		t.Fatalf("duplicate register request failed: %v", err)
	}
	if dupResp.StatusCode != fiber.StatusConflict {
		t.Errorf("expected status 409 Conflict, got %d", dupResp.StatusCode)
	}

	// 3. Test POST /api/v1/auth/login
	loginBody, _ := json.Marshal(map[string]interface{}{
		"email":    testEmail,
		"password": testPassword,
	})

	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq, 5000)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}

	if loginResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200 OK, got %d", loginResp.StatusCode)
	}

	loginCookie := loginResp.Header.Get("Set-Cookie")
	if !strings.Contains(loginCookie, "refresh_token=") {
		t.Errorf("expected refresh_token cookie on login, got: %s", loginCookie)
	}

	// 4. Test POST /api/v1/auth/login with Invalid Password (HTTP 401)
	badLoginBody, _ := json.Marshal(map[string]interface{}{
		"email":    testEmail,
		"password": "WrongPassword123!",
	})
	badLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(badLoginBody))
	badLoginReq.Header.Set("Content-Type", "application/json")
	badLoginResp, _ := app.Test(badLoginReq, 5000)
	if badLoginResp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized for bad password, got %d", badLoginResp.StatusCode)
	}

	// 5. Test POST /api/v1/auth/refresh using Cookie
	refreshReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	// Extract raw cookie value from login response
	cookieParts := strings.Split(loginCookie, ";")
	refreshReq.Header.Set("Cookie", cookieParts[0])
	refreshResp, err := app.Test(refreshReq, 5000)
	if err != nil {
		t.Fatalf("refresh request failed: %v", err)
	}
	if refreshResp.StatusCode != fiber.StatusOK {
		body, _ := io.ReadAll(refreshResp.Body)
		t.Fatalf("expected status 200 OK on refresh, got %d: %s", refreshResp.StatusCode, string(body))
	}

	// 6. Test POST /api/v1/auth/logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	logoutReq.Header.Set("Cookie", cookieParts[0])
	logoutResp, err := app.Test(logoutReq, 5000)
	if err != nil {
		t.Fatalf("logout request failed: %v", err)
	}
	if logoutResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200 OK on logout, got %d", logoutResp.StatusCode)
	}
	// Check that cookie is cleared (Max-Age=-1 or expires in past)
	logoutCookie := logoutResp.Header.Get("Set-Cookie")
	if !strings.Contains(logoutCookie, "refresh_token=;") && !strings.Contains(logoutCookie, "Max-Age=-1") {
		t.Errorf("expected cleared cookie on logout, got: %s", logoutCookie)
	}

	// 7. Test GET /api/v1/auth/google (Initiation endpoint)
	googleReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google?redirect=false", nil)
	googleReq.Header.Set("Accept", "application/json")
	googleResp, err := app.Test(googleReq, 5000)
	if err != nil {
		t.Fatalf("google init request failed: %v", err)
	}
	if googleResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200 OK on google init, got %d", googleResp.StatusCode)
	}

	googleCookies := googleResp.Header.Values("Set-Cookie")
	var hasState, hasNonce, hasPKCE bool
	for _, c := range googleCookies {
		if strings.Contains(c, "oauth_state=") {
			hasState = true
		}
		if strings.Contains(c, "oauth_nonce=") {
			hasNonce = true
		}
		if strings.Contains(c, "oauth_code_verifier=") {
			hasPKCE = true
		}
	}
	if !hasState || !hasNonce || !hasPKCE {
		t.Errorf("expected oauth_state, oauth_nonce, and oauth_code_verifier cookies, got: %v", googleCookies)
	}
}

func TestRateLimiting(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'http.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'http.test%'")
	}()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	loginBody, _ := json.Marshal(map[string]interface{}{
		"email":    "ratelimit@mindora.app",
		"password": "Password123!",
	})

	var reached429 bool
	// Burst request 15 times (limit is 10 per minute)
	for i := 0; i < 15; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := app.Test(req, 2000)
		if resp.StatusCode == fiber.StatusTooManyRequests {
			reached429 = true
			break
		}
	}

	if !reached429 {
		t.Errorf("expected rate limiter to trigger HTTP 429 Too Many Requests within 15 attempts")
	}
}

func TestLogoutRateLimiting(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	var reached429 bool
	// Normal limiter limit is 30 per minute; burst 35 requests
	for i := 0; i < 35; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
		resp, _ := app.Test(req, 2000)
		if resp.StatusCode == fiber.StatusTooManyRequests {
			reached429 = true
			break
		}
	}

	if !reached429 {
		t.Errorf("expected rate limiter on /logout to trigger HTTP 429 within 35 attempts")
	}
}

func TestGoogleOAuthRedirect_NoTokenLeakage(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false
	cfg.GoogleOAuth.Enabled = true
	cfg.GoogleOAuth.ClientID = "test-client-id"
	cfg.GoogleOAuth.ClientSecret = "test-client-secret"
	cfg.FrontendURL = "http://localhost:5173"

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// In HTML browser callback without code, verify error handling
	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/google/callback", nil)
	req.Header.Set("Accept", "text/html")
	resp, err := app.Test(req, 2000)
	if err != nil {
		t.Fatalf("callback request failed: %v", err)
	}

	loc := resp.Header.Get("Location")
	if strings.Contains(loc, "token=") || strings.Contains(loc, "access_token=") || strings.Contains(loc, "refresh_token=") {
		t.Errorf("SECURITY VULNERABILITY: token exposed in redirect Location URL: %s", loc)
	}
}
