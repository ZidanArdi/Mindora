package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
)

func TestStudyHandler_FlowAndTimezone(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// Clean up test user
	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	userEmail := fmt.Sprintf("study.test.%s@mindora.app", uniqueSuffix)
	password := "SecureStudy2026!"

	defer func() {
		_ = db.DB.Exec("DELETE FROM study_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'study.test%')")
		_ = db.DB.Exec("DELETE FROM tasks WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'study.test%')")
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'study.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'study.test%'")
	}()

	// 1. Unauthorized access to /api/v1/study/stats/daily should be 401
	unauthReq := httptest.NewRequest("GET", "/api/v1/study/stats/daily", nil)
	unauthResp, err := app.Test(unauthReq, 2000)
	if err != nil {
		t.Fatalf("failed unauth request: %v", err)
	}
	if unauthResp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized, got %d", unauthResp.StatusCode)
	}

	// 2. Register & Login user
	regBody, _ := json.Marshal(map[string]interface{}{
		"name":     "Study Tester",
		"email":    userEmail,
		"password": password,
	})
	regReq := httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq, 3000)
	if err != nil || regResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to register test user: %v, status: %d", err, regResp.StatusCode)
	}

	loginBody, _ := json.Marshal(map[string]interface{}{
		"email":    userEmail,
		"password": password,
	})
	loginReq := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq, 3000)
	if err != nil || loginResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to login test user: %v", err)
	}

	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	bodyBytes, _ := io.ReadAll(loginResp.Body)
	_ = json.Unmarshal(bodyBytes, &loginData)
	token := loginData.Data.AccessToken
	if token == "" {
		t.Fatalf("empty access token received")
	}

	// 3. Query daily stats for today with tz=Asia/Jakarta (should initially be 0)
	statsReq := httptest.NewRequest("GET", "/api/v1/study/stats/daily?tz=Asia/Jakarta", nil)
	statsReq.Header.Set("Authorization", "Bearer "+token)
	statsResp, err := app.Test(statsReq, 3000)
	if err != nil || statsResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to get daily stats: %v, status: %d", err, statsResp.StatusCode)
	}

	var statsData struct {
		Data struct {
			TotalCompletedSeconds  int64 `json:"total_completed_seconds"`
			CompletedSessionsCount int64 `json:"completed_sessions_count"`
		} `json:"data"`
	}
	statsBody, _ := io.ReadAll(statsResp.Body)
	_ = json.Unmarshal(statsBody, &statsData)

	if statsData.Data.CompletedSessionsCount != 0 || statsData.Data.TotalCompletedSeconds != 0 {
		t.Errorf("expected 0 initial stats, got: %+v", statsData.Data)
	}

	// 4. Start an independent study session (no task linked)
	startBody, _ := json.Marshal(map[string]interface{}{
		"subject": "Matematika Diskrit Pertemuan 3",
	})
	startReq := httptest.NewRequest("POST", "/api/v1/study/sessions/start", bytes.NewReader(startBody))
	startReq.Header.Set("Content-Type", "application/json")
	startReq.Header.Set("Authorization", "Bearer "+token)
	startResp, err := app.Test(startReq, 3000)
	if err != nil || startResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to start study session: %v, status: %d", err, startResp.StatusCode)
	}

	var sessionData struct {
		Data struct {
			ID      string `json:"id"`
			Subject string `json:"subject"`
			Status  string `json:"status"`
			TaskID  *string `json:"task_id"`
		} `json:"data"`
	}
	startRespBody, _ := io.ReadAll(startResp.Body)
	_ = json.Unmarshal(startRespBody, &sessionData)

	if sessionData.Data.TaskID != nil {
		t.Errorf("expected nil task_id for independent study, got: %v", sessionData.Data.TaskID)
	}

	// 5. Complete session
	finishReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/study/sessions/%s/finish", sessionData.Data.ID), nil)
	finishReq.Header.Set("Authorization", "Bearer "+token)
	finishResp, err := app.Test(finishReq, 3000)
	if err != nil || finishResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to finish study session: %v, status: %d", err, finishResp.StatusCode)
	}

	// 6. Query daily stats again with tz=Asia/Jakarta (should now be 1 completed session)
	statsReq2 := httptest.NewRequest("GET", "/api/v1/study/stats/daily?tz=Asia/Jakarta", nil)
	statsReq2.Header.Set("Authorization", "Bearer "+token)
	statsResp2, err := app.Test(statsReq2, 3000)
	if err != nil || statsResp2.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to get daily stats after completion: %v", err)
	}

	statsBody2, _ := io.ReadAll(statsResp2.Body)
	_ = json.Unmarshal(statsBody2, &statsData)

	if statsData.Data.CompletedSessionsCount != 1 {
		t.Errorf("expected 1 completed session, got %d", statsData.Data.CompletedSessionsCount)
	}
}
