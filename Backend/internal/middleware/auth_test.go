package middleware_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/response"
)

func TestAuthorizationAndRBAC(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false
	cfg.GoogleOAuth.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	// Ensure RBAC is seeded
	if err := database.SeedRBAC(t.Context(), db.DB); err != nil {
		t.Fatalf("seed RBAC failed: %v", err)
	}

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	tokenManager := jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	expiredTokenManager := jwt.NewTokenManager(cfg.JWT.Secret, -1*time.Minute, -1*time.Minute)

	// Create test users in DB
	userRoleID := database.DeterministicRoleID("user")
	adminRoleID := database.DeterministicRoleID("admin")

	userAID := uuid.New()
	userBID := uuid.New()
	adminID := uuid.New()

	testUsers := []models.User{
		{
			ID:       userAID,
			Name:     "Test User A",
			Email:    "test.usera." + userAID.String()[:8] + "@mindora.app",
			RoleID:   userRoleID,
			Status:   "active",
		},
		{
			ID:       userBID,
			Name:     "Test User B",
			Email:    "test.userb." + userBID.String()[:8] + "@mindora.app",
			RoleID:   userRoleID,
			Status:   "active",
		},
		{
			ID:       adminID,
			Name:     "Test Admin",
			Email:    "test.admin." + adminID.String()[:8] + "@mindora.app",
			RoleID:   adminRoleID,
			Status:   "active",
		},
	}

	for _, u := range testUsers {
		if err := db.DB.Create(&u).Error; err != nil {
			t.Fatalf("failed to insert test user: %v", err)
		}
	}

	defer func() {
		_ = db.DB.Exec("DELETE FROM users WHERE id IN (?, ?, ?)", userAID, userBID, adminID)
	}()

	// Mint tokens
	userAToken, _, err := tokenManager.GenerateAccessToken(userAID, testUsers[0].Email, "user")
	if err != nil {
		t.Fatalf("failed to generate userA token: %v", err)
	}

	adminToken, _, err := tokenManager.GenerateAccessToken(adminID, testUsers[2].Email, "admin")
	if err != nil {
		t.Fatalf("failed to generate admin token: %v", err)
	}

	expiredToken, _, err := expiredTokenManager.GenerateAccessToken(userAID, testUsers[0].Email, "user")
	if err != nil {
		t.Fatalf("failed to generate expired token: %v", err)
	}

	// 1. Guest Access Restrictions
	t.Run("Guest access to /api/v1/user/profile is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got: %d", resp.StatusCode)
		}
	})

	t.Run("Malformed Authorization header is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got: %d", resp.StatusCode)
		}
	})

	// 2. Invalid and Expired JWT
	t.Run("Invalid JWT signature is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.signature")
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got: %d", resp.StatusCode)
		}
	})

	t.Run("Expired JWT access token is rejected", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		req.Header.Set("Authorization", "Bearer "+expiredToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got: %d", resp.StatusCode)
		}
	})

	// 3. User Authorization
	t.Run("User can access own profile", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK, got: %d", resp.StatusCode)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		var apiResp response.Response
		_ = json.Unmarshal(bodyBytes, &apiResp)
		if !apiResp.Success {
			t.Errorf("expected success: true, got: %v", apiResp.Success)
		}
	})

	t.Run("User can update own profile", func(t *testing.T) {
		payload, _ := json.Marshal(map[string]string{"name": "Updated User A"})
		req := httptest.NewRequest(http.MethodPut, "/api/v1/user/profile", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK, got: %d", resp.StatusCode)
		}
	})

	t.Run("User is forbidden from admin overview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got: %d", resp.StatusCode)
		}
	})

	t.Run("User is forbidden from admin users list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got: %d", resp.StatusCode)
		}
	})

	// 4. Admin Authorization
	t.Run("Admin can access admin overview", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK, got: %d", resp.StatusCode)
		}
	})

	t.Run("Admin can access admin users list", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users?limit=5", nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK, got: %d", resp.StatusCode)
		}
	})

	// 5. Cross-User Resource Access Rejection (Ownership Guard)
	t.Run("User A can access own productivity tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+userAID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK, got: %d", resp.StatusCode)
		}
	})

	t.Run("User A is rejected from accessing User B's productivity tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+userBID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+userAToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-user access, got: %d", resp.StatusCode)
		}
	})

	// 6. Admin Privacy Restrictions
	t.Run("Admin cannot freely read private productivity records of User A", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+userAID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+adminToken)
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden: Admin privacy violation, got: %d", resp.StatusCode)
		}
	})

	// 7. Role Escalation Prevention on Public Registration
	t.Run("Public registration payload attempting to inject role 'admin' is strictly forced to role 'user'", func(t *testing.T) {
		hackerEmail := "hacker." + uuid.New().String()[:8] + "@mindora.app"
		regPayload, _ := json.Marshal(map[string]interface{}{
			"name":     "Hacker Attempt",
			"email":    hackerEmail,
			"password": "SecurePassword123!",
			"role":     "admin", // malicious privilege escalation injection attempt
		})

		defer func() {
			_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email = ?)", hackerEmail)
			_ = db.DB.Exec("DELETE FROM users WHERE email = ?", hackerEmail)
		}()

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regPayload))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusCreated {
			t.Fatalf("expected 201 Created, got: %d", resp.StatusCode)
		}

		bodyBytes, _ := io.ReadAll(resp.Body)
		var apiResp response.Response
		_ = json.Unmarshal(bodyBytes, &apiResp)

		dataMap, ok := apiResp.Data.(map[string]interface{})
		if !ok {
			t.Fatalf("expected data map in response")
		}
		userMap, ok := dataMap["user"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected user map in response data")
		}

		if userMap["role"] != "user" {
			t.Errorf("SECURITY CRITICAL: role escalation succeeded! Expected 'user', got '%v'", userMap["role"])
		}
	})

	// 8. Suspended Account Rejection
	t.Run("Suspended user account login is rejected with 403 Forbidden", func(t *testing.T) {
		suspendedEmail := "suspended." + uuid.New().String()[:8] + "@mindora.app"
		suspendedUser := models.User{
			ID:       uuid.New(),
			Name:     "Suspended Student",
			Email:    suspendedEmail,
			RoleID:   userRoleID,
			Status:   "suspended",
		}
		if err := db.DB.Create(&suspendedUser).Error; err != nil {
			t.Fatalf("failed to insert suspended user: %v", err)
		}
		defer func() {
			_ = db.DB.Exec("DELETE FROM users WHERE id = ?", suspendedUser.ID)
		}()

		loginPayload, _ := json.Marshal(map[string]string{
			"email":    suspendedEmail,
			"password": "Password123!",
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginPayload))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for suspended account, got: %d", resp.StatusCode)
		}
	})

	// 9. Revoked Session Refresh Rejection
	t.Run("Refresh token on revoked session is rejected with 401 Unauthorized", func(t *testing.T) {
		rawToken, tokenHash, exp, err := tokenManager.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("failed to generate refresh token: %v", err)
		}

		now := time.Now().UTC()
		revokedSession := models.AuthSession{
			ID:        uuid.New(),
			UserID:    userAID,
			FamilyID:  uuid.New(),
			TokenHash: tokenHash,
			ExpiresAt: exp,
			RevokedAt: &now, // Revoked!
			CreatedAt: now,
		}
		if err := db.DB.Create(&revokedSession).Error; err != nil {
			t.Fatalf("failed to insert revoked session: %v", err)
		}
		defer func() {
			_ = db.DB.Exec("DELETE FROM auth_sessions WHERE id = ?", revokedSession.ID)
		}()

		refreshPayload, _ := json.Marshal(map[string]string{
			"refresh_token": rawToken,
			"operation_id":  uuid.New().String(),
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", bytes.NewReader(refreshPayload))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 3000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for revoked session, got: %d", resp.StatusCode)
		}
	})
}
