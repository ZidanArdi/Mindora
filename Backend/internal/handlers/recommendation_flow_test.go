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
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/routes"
	"mindora/backend/pkg/jwt"
)

func TestRecommendationFlow_SecurityAndAPI(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	ctx := context.Background()
	_ = database.SeedRBAC(ctx, db.DB)

	userRepo := repositories.NewUserRepository(db.DB)
	recRepo := repositories.NewRecommendationRepository(db.DB)

	studentRole, err := userRepo.GetRoleByName(ctx, "user")
	if err != nil {
		t.Fatalf("failed to get user role: %v", err)
	}
	adminRole, err := userRepo.GetRoleByName(ctx, "admin")
	if err != nil {
		t.Fatalf("failed to get admin role: %v", err)
	}

	// 1. Create Student A
	pw := "hashed_pw"
	studentA := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("rec_flow_a_%s@mindora.app", uuid.New().String()[:8]),
		PasswordHash: &pw,
		Name:         "Student A",
		RoleID:       studentRole.ID,
		Status:       "active",
		CreatedAt:    time.Now().UTC().AddDate(0, 0, -10),
	}
	if err := userRepo.Create(ctx, studentA); err != nil {
		t.Fatalf("failed to create student A: %v", err)
	}

	// 2. Create Student B
	studentB := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("rec_flow_b_%s@mindora.app", uuid.New().String()[:8]),
		PasswordHash: &pw,
		Name:         "Student B",
		RoleID:       studentRole.ID,
		Status:       "active",
		CreatedAt:    time.Now().UTC().AddDate(0, 0, -10),
	}
	if err := userRepo.Create(ctx, studentB); err != nil {
		t.Fatalf("failed to create student B: %v", err)
	}

	// 3. Create Admin
	adminUser := &models.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("rec_flow_admin_%s@mindora.app", uuid.New().String()[:8]),
		PasswordHash: &pw,
		Name:         "Administrator",
		RoleID:       adminRole.ID,
		Status:       "active",
		CreatedAt:    time.Now().UTC().AddDate(0, 0, -30),
	}
	if err := userRepo.Create(ctx, adminUser); err != nil {
		t.Fatalf("failed to create admin: %v", err)
	}

	defer func() {
		_ = db.DB.Exec("DELETE FROM tasks WHERE user_id IN (?, ?, ?)", studentA.ID, studentB.ID, adminUser.ID)
		_ = db.DB.Exec("DELETE FROM recommendations WHERE user_id IN (?, ?, ?)", studentA.ID, studentB.ID, adminUser.ID)
		_ = db.DB.Exec("DELETE FROM users WHERE id IN (?, ?, ?)", studentA.ID, studentB.ID, adminUser.ID)
	}()

	// Generate JWT tokens
	tokenManager := jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	tokenA, _, err := tokenManager.GenerateAccessToken(studentA.ID, studentA.Email, "user")
	if err != nil {
		t.Fatalf("failed to generate token A: %v", err)
	}
	tokenB, _, err := tokenManager.GenerateAccessToken(studentB.ID, studentB.Email, "user")
	if err != nil {
		t.Fatalf("failed to generate token B: %v", err)
	}
	tokenAdmin, _, err := tokenManager.GenerateAccessToken(adminUser.ID, adminUser.Email, "admin")
	if err != nil {
		t.Fatalf("failed to generate token admin: %v", err)
	}

	// Seed an actual overdue high-priority task for Student A
	now := time.Now().UTC()
	taskRepo := repositories.NewTaskRepository(db.DB)
	overdueTime := now.Add(-1 * time.Hour)
	taskA := &models.Task{
		ID:       uuid.New(),
		UserID:   studentA.ID,
		Title:    "Overdue History Paper",
		Category: "Study",
		Priority: models.TaskPriorityHigh,
		Status:   models.TaskStatusPending,
		Deadline: &overdueTime,
	}
	if err := taskRepo.Create(ctx, taskA); err != nil {
		t.Fatalf("failed to create taskA: %v", err)
	}

	// Seed a recommendation for Student A
	recA := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                studentA.ID,
		RuleID:                "REC-TASK-001",
		Category:              "Task Prioritization",
		Priority:              models.RecPriorityCritical,
		Title:                 "Action Required: Overdue Task",
		Message:               "You have 1 overdue task",
		Explanation:           "Explanation",
		Evidence:              models.JSONB([]byte(`{"overdue_high_priority_count": 1}`)),
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: now,
		EvaluationPeriodEnd:   now.Add(24 * time.Hour),
		EvaluationKey:         fmt.Sprintf("REC-TASK-001:%s:%s", studentA.ID.String(), taskA.ID.String()),
		ValidUntil:            now.Add(24 * time.Hour),
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := recRepo.UpsertPending(ctx, recA); err != nil {
		t.Fatalf("failed to seed recA: %v", err)
	}

	// -------------------------------------------------------------------------
	// Test 1: Unauthenticated Request -> 401 Unauthorized
	// -------------------------------------------------------------------------
	req := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 Unauthorized for missing token, got %d", resp.StatusCode)
	}

	// -------------------------------------------------------------------------
	// Test 2: Admin Request -> 403 Forbidden (Admin Access Denial Rule 8)
	// -------------------------------------------------------------------------
	reqAdmin := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	reqAdmin.Header.Set("Authorization", "Bearer "+tokenAdmin)
	respAdmin, err := app.Test(reqAdmin)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if respAdmin.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when Admin accesses recommendations, got %d", respAdmin.StatusCode)
	}

	// Admin PATCH -> 403 Forbidden
	reqAdminPatch := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/recommendations/%s/status", recA.ID.String()), bytes.NewReader([]byte(`{"status":"dismissed"}`)))
	reqAdminPatch.Header.Set("Authorization", "Bearer "+tokenAdmin)
	reqAdminPatch.Header.Set("Content-Type", "application/json")
	respAdminPatch, err := app.Test(reqAdminPatch)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if respAdminPatch.StatusCode != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when Admin modifies recommendations, got %d", respAdminPatch.StatusCode)
	}

	// -------------------------------------------------------------------------
	// Test 3: Student A GET Recommendations -> 200 OK
	// -------------------------------------------------------------------------
	reqStudentA := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	reqStudentA.Header.Set("Authorization", "Bearer "+tokenA)
	respStudentA, err := app.Test(reqStudentA)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if respStudentA.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for Student A, got %d", respStudentA.StatusCode)
	}

	var jsonResp struct {
		Success bool `json:"success"`
		Data    struct {
			Items        []models.Recommendation `json:"items"`
			Total        int                     `json:"total"`
			UserDataTier string                  `json:"user_data_tier"`
		} `json:"data"`
	}
	_ = json.NewDecoder(respStudentA.Body).Decode(&jsonResp)
	if !jsonResp.Success || jsonResp.Data.Total < 1 {
		t.Errorf("expected at least 1 recommendation returned, got %d", jsonResp.Data.Total)
	}

	// -------------------------------------------------------------------------
	// Test 4: IDOR Protection: Student B attempts to PATCH Student A's recommendation -> 404 Not Found
	// -------------------------------------------------------------------------
	reqIDOR := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/recommendations/%s/status", recA.ID.String()), bytes.NewReader([]byte(`{"status":"dismissed"}`)))
	reqIDOR.Header.Set("Authorization", "Bearer "+tokenB)
	reqIDOR.Header.Set("Content-Type", "application/json")
	respIDOR, err := app.Test(reqIDOR)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if respIDOR.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404 Not Found for cross-user IDOR attempt, got %d", respIDOR.StatusCode)
	}

	// -------------------------------------------------------------------------
	// Test 5: Legitimate Status Mutation: Student A dismisses own recommendation -> 200 OK
	// -------------------------------------------------------------------------
	reqDismiss := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/recommendations/%s/status", recA.ID.String()), bytes.NewReader([]byte(`{"status":"dismissed"}`)))
	reqDismiss.Header.Set("Authorization", "Bearer "+tokenA)
	reqDismiss.Header.Set("Content-Type", "application/json")
	respDismiss, err := app.Test(reqDismiss)
	if err != nil {
		t.Fatalf("HTTP request failed: %v", err)
	}
	if respDismiss.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for valid dismissal, got %d", respDismiss.StatusCode)
	}

	// Verify dismissed record is no longer returned as active pending
	reqAfter := httptest.NewRequest(http.MethodGet, "/api/v1/recommendations", nil)
	reqAfter.Header.Set("Authorization", "Bearer "+tokenA)
	respAfter, _ := app.Test(reqAfter)
	var jsonRespAfter struct {
		Data struct {
			Items []models.Recommendation `json:"items"`
		} `json:"data"`
	}
	_ = json.NewDecoder(respAfter.Body).Decode(&jsonRespAfter)
	for _, item := range jsonRespAfter.Data.Items {
		if item.ID == recA.ID {
			t.Errorf("dismissed recommendation must not be returned in active list")
		}
	}
}
