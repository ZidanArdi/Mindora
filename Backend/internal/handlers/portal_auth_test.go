package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/hash"
	"mindora/backend/pkg/response"
)

func TestUserAdminPortalAuthorization(t *testing.T) {
	cfg := config.LoadConfig()
	// Disable Turnstile for direct portal test to isolate portal matrix authorization
	cfg.Turnstile.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	userEmail := fmt.Sprintf("portal.student.%s@mindora.app", uniqueSuffix)
	userPassword := "StudentPass123!"

	adminEmail := fmt.Sprintf("portal.admin.%s@mindora.app", uniqueSuffix)
	adminPassword := "AdminPass123!"

	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'portal.%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'portal.%'")
	}()

	// 1. Provision standard user via registration
	regBody, _ := json.Marshal(map[string]interface{}{
		"name":     "Portal Test Student",
		"email":    userEmail,
		"password": userPassword,
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq, 5000)
	if err != nil || regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to register student user: %v, status: %d", err, regResp.StatusCode)
	}

	var regEnvelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
			User        struct {
				ID   string `json:"id"`
				Role string `json:"role"`
			} `json:"user"`
		} `json:"data"`
	}
	_ = json.NewDecoder(regResp.Body).Decode(&regEnvelope)
	userAccessToken := regEnvelope.Data.AccessToken
	studentUserID := regEnvelope.Data.User.ID

	// 2. Provision admin user via controlled bootstrapping
	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	t.Setenv("ADMIN_EMAIL", adminEmail)
	t.Setenv("ADMIN_PASSWORD", adminPassword)
	t.Setenv("ADMIN_NAME", "Portal Test Admin")
	if err := database.BootstrapAdmin(context.Background(), db.DB, hasher, cfg); err != nil {
		t.Fatalf("failed to bootstrap admin user: %v", err)
	}

	// -------------------------------------------------------------------------
	// Requirement 1: User login through User Portal succeeds
	// -------------------------------------------------------------------------
	t.Run("UserLogin_ThroughUserPortal_Succeeds", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    userEmail,
			"password": userPassword,
			"portal":   "user",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK for user logging into user portal, got: %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------------------
	// Requirement 2: User login through Admin Portal returns 403 Forbidden
	// -------------------------------------------------------------------------
	t.Run("UserLogin_ThroughAdminPortal_Returns403Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    userEmail,
			"password": userPassword,
			"portal":   "admin",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for user attempting admin portal login, got: %d", resp.StatusCode)
		}

		var res response.Response
		_ = json.NewDecoder(resp.Body).Decode(&res)
		if res.Success {
			t.Errorf("expected success to be false")
		}
	})

	// -------------------------------------------------------------------------
	// Requirement 3: Admin login through Admin Portal succeeds
	// -------------------------------------------------------------------------
	var adminAccessToken string
	t.Run("AdminLogin_ThroughAdminPortal_Succeeds", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    adminEmail,
			"password": adminPassword,
			"portal":   "admin",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200 OK for admin logging into admin portal, got: %d", resp.StatusCode)
		}

		var adminResp struct {
			Data struct {
				AccessToken string `json:"access_token"`
			} `json:"data"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&adminResp)
		adminAccessToken = adminResp.Data.AccessToken
	})

	// -------------------------------------------------------------------------
	// Requirement 4: Admin login through User Portal returns 403 Forbidden
	// -------------------------------------------------------------------------
	t.Run("AdminLogin_ThroughUserPortal_Returns403Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    adminEmail,
			"password": adminPassword,
			"portal":   "user",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for admin attempting user portal login, got: %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------------------
	// Requirement 5: Invalid portal parameter returns 403 Forbidden
	// -------------------------------------------------------------------------
	t.Run("InvalidPortal_Returns403Forbidden", func(t *testing.T) {
		body, _ := json.Marshal(map[string]interface{}{
			"email":    userEmail,
			"password": userPassword,
			"portal":   "super_admin_portal_hack",
		})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for invalid portal parameter, got: %d", resp.StatusCode)
		}
	})

	// -------------------------------------------------------------------------
	// Requirement 6: Admin API rejects a user JWT with 403 Forbidden
	// -------------------------------------------------------------------------
	t.Run("AdminAPI_RejectsUserJWT_Returns403Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
		req.Header.Set("Authorization", "Bearer "+userAccessToken)

		resp, err := app.Test(req, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if resp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden when user accesses /api/v1/admin/overview, got: %d", resp.StatusCode)
		}

		// Also verify admin endpoint /api/v1/admin/users
		reqUsers := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
		reqUsers.Header.Set("Authorization", "Bearer "+userAccessToken)
		respUsers, _ := app.Test(reqUsers, 5000)
		if respUsers.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden when user accesses /api/v1/admin/users, got: %d", respUsers.StatusCode)
		}

		// Verify that admin token CAN access admin overview
		adminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
		adminReq.Header.Set("Authorization", "Bearer "+adminAccessToken)
		adminResp, _ := app.Test(adminReq, 5000)
		if adminResp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK for admin accessing /api/v1/admin/overview, got: %d", adminResp.StatusCode)
		}
	})

	// -------------------------------------------------------------------------
	// Requirement 7: User API enforces ownership
	// -------------------------------------------------------------------------
	t.Run("UserAPI_EnforcesOwnership", func(t *testing.T) {
		otherUserUUID := uuid.New().String()

		// Accessing other user's tasks with student user token -> MUST return 403 Forbidden
		badReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+otherUserUUID, nil)
		badReq.Header.Set("Authorization", "Bearer "+userAccessToken)
		badResp, err := app.Test(badReq, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if badResp.StatusCode != fiber.StatusForbidden {
			t.Errorf("expected 403 Forbidden for cross-account task access, got: %d", badResp.StatusCode)
		}

		// Accessing own tasks with student user token -> MUST return 200 OK
		ownReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+studentUserID, nil)
		ownReq.Header.Set("Authorization", "Bearer "+userAccessToken)
		ownResp, err := app.Test(ownReq, 5000)
		if err != nil {
			t.Fatalf("request error: %v", err)
		}
		if ownResp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200 OK for accessing own tasks, got: %d", ownResp.StatusCode)
		}
	})
}
