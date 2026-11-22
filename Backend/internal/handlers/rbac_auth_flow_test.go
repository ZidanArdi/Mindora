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
)

type AuthResponseData struct {
	AccessToken string `json:"access_token"`
	User        struct {
		ID    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
		Role  string `json:"role"`
	} `json:"user"`
}

type APIEnvelope struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Data    AuthResponseData `json:"data"`
}

func TestGuestUserAdminAccessControl(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// Clean up any test records at the end
	defer func() {
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'flow.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'flow.test%'")
	}()

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	userEmail := fmt.Sprintf("flow.test.student.%s@mindora.app", uniqueSuffix)
	userPassword := "StudentSecurePass123!"

	// -------------------------------------------------------------
	// 1. GUEST ACCESS VERIFICATION
	// -------------------------------------------------------------
	// Unauthenticated request to /api/v1/user/profile must be 401 Unauthorized
	guestUserReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	guestUserResp, err := app.Test(guestUserReq, 5000)
	if err != nil {
		t.Fatalf("guest user request failed: %v", err)
	}
	if guestUserResp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for guest profile access, got %d", guestUserResp.StatusCode)
	}

	// Unauthenticated request to /api/v1/admin/overview must be 401 Unauthorized
	guestAdminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
	guestAdminResp, err := app.Test(guestAdminReq, 5000)
	if err != nil {
		t.Fatalf("guest admin request failed: %v", err)
	}
	if guestAdminResp.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for guest admin access, got %d", guestAdminResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 2. USER REGISTRATION & LOGIN VERIFICATION
	// -------------------------------------------------------------
	regBody, _ := json.Marshal(map[string]interface{}{
		"name":     "Test Student",
		"email":    userEmail,
		"password": userPassword,
	})
	regReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq, 5000)
	if err != nil {
		t.Fatalf("user registration failed: %v", err)
	}
	if regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on registration, got %d", regResp.StatusCode)
	}

	var regEnvelope APIEnvelope
	_ = json.NewDecoder(regResp.Body).Decode(&regEnvelope)
	userAccessToken := regEnvelope.Data.AccessToken
	studentUserID := regEnvelope.Data.User.ID

	if regEnvelope.Data.User.Role != "user" {
		t.Fatalf("expected role 'user' on public registration, got '%s'", regEnvelope.Data.User.Role)
	}

	// -------------------------------------------------------------
	// 3. ADMIN AUTHENTICATION VERIFICATION
	// -------------------------------------------------------------
	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	adminEmail := "admin@mindora.app"
	adminPassword := "AdminMindora2026!"
	t.Setenv("ADMIN_EMAIL", adminEmail)
	t.Setenv("ADMIN_PASSWORD", adminPassword)
	t.Setenv("ADMIN_NAME", "Mindora Admin")

	ctx := context.Background()
	if err := database.BootstrapAdmin(ctx, db.DB, hasher, cfg); err != nil {
		t.Fatalf("failed to bootstrap admin for test: %v", err)
	}

	adminLoginBody, _ := json.Marshal(map[string]interface{}{
		"email":    adminEmail,
		"password": adminPassword,
	})
	adminLoginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(adminLoginBody))
	adminLoginReq.Header.Set("Content-Type", "application/json")
	adminLoginResp, err := app.Test(adminLoginReq, 5000)
	if err != nil {
		t.Fatalf("admin login request failed: %v", err)
	}
	if adminLoginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on admin login, got %d", adminLoginResp.StatusCode)
	}

	var adminEnvelope APIEnvelope
	_ = json.NewDecoder(adminLoginResp.Body).Decode(&adminEnvelope)
	adminAccessToken := adminEnvelope.Data.AccessToken

	if adminEnvelope.Data.User.Role != "admin" {
		t.Fatalf("expected role 'admin' for admin account, got '%s'", adminEnvelope.Data.User.Role)
	}

	// -------------------------------------------------------------
	// 4. USER ROLE ACCESS MATRIX VERIFICATION
	// -------------------------------------------------------------
	// User accessing own profile: MUST BE 200 OK
	userProfReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/profile", nil)
	userProfReq.Header.Set("Authorization", "Bearer "+userAccessToken)
	userProfResp, _ := app.Test(userProfReq, 5000)
	if userProfResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK for user profile access, got %d", userProfResp.StatusCode)
	}

	// User accessing admin overview: MUST BE 403 FORBIDDEN
	userAdminReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
	userAdminReq.Header.Set("Authorization", "Bearer "+userAccessToken)
	userAdminResp, _ := app.Test(userAdminReq, 5000)
	if userAdminResp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 Forbidden for user accessing admin overview, got %d", userAdminResp.StatusCode)
	}

	// User accessing admin user list: MUST BE 403 FORBIDDEN
	userListReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	userListReq.Header.Set("Authorization", "Bearer "+userAccessToken)
	userListResp, _ := app.Test(userListReq, 5000)
	if userListResp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 Forbidden for user accessing admin user list, got %d", userListResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 5. ADMIN ROLE ACCESS MATRIX & OWNERSHIP RESTRICTION
	// -------------------------------------------------------------
	// Admin accessing admin overview: MUST BE 200 OK
	adminOverviewReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/overview", nil)
	adminOverviewReq.Header.Set("Authorization", "Bearer "+adminAccessToken)
	adminOverviewResp, _ := app.Test(adminOverviewReq, 5000)
	if adminOverviewResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK for admin overview access, got %d", adminOverviewResp.StatusCode)
	}

	// Admin accessing admin user list: MUST BE 200 OK
	adminUsersReq := httptest.NewRequest(http.MethodGet, "/api/v1/admin/users", nil)
	adminUsersReq.Header.Set("Authorization", "Bearer "+adminAccessToken)
	adminUsersResp, _ := app.Test(adminUsersReq, 5000)
	if adminUsersResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK for admin user list, got %d", adminUsersResp.StatusCode)
	}

	// Admin accessing personal productivity tasks of another user: MUST BE 403 FORBIDDEN (Privacy Protection)
	adminTaskReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+studentUserID, nil)
	adminTaskReq.Header.Set("Authorization", "Bearer "+adminAccessToken)
	adminTaskResp, _ := app.Test(adminTaskReq, 5000)
	if adminTaskResp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 Forbidden for admin accessing student personal tasks, got %d", adminTaskResp.StatusCode)
	}

	// User accessing personal tasks with mismatched arbitrary UUID: MUST BE 403 FORBIDDEN (Ownership Protection)
	arbitraryID := uuid.New().String()
	mismatchedTaskReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/tasks/"+arbitraryID, nil)
	mismatchedTaskReq.Header.Set("Authorization", "Bearer "+userAccessToken)
	mismatchedTaskResp, _ := app.Test(mismatchedTaskReq, 5000)
	if mismatchedTaskResp.StatusCode != fiber.StatusForbidden {
		t.Errorf("expected 403 Forbidden for mismatched ownership task access, got %d", mismatchedTaskResp.StatusCode)
	}
}
