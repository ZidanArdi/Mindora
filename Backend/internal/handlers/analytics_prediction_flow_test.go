package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/routes"
)

func TestAnalyticsAndPredictionFlow(t *testing.T) {
	cfg := config.LoadConfig()
	cfg.Turnstile.Enabled = false

	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("database connect failed: %v", err)
	}
	defer db.Close()

	app := fiber.New()
	routes.SetupRoutes(app, db, cfg)

	// Cleanup test users at end
	defer func() {
		_ = db.DB.Exec("DELETE FROM productivity_predictions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM tasks WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM study_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM focus_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM lifestyle_records WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'phase3.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'phase3.test%'")
	}()

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	userEmailA := fmt.Sprintf("phase3.test.user.a.%s@mindora.app", uniqueSuffix)
	userEmailB := fmt.Sprintf("phase3.test.user.b.%s@mindora.app", uniqueSuffix)
	password := "Phase3SecurePassword2026!"

	// 1. REGISTER & LOGIN USER A
	regBodyA, _ := json.Marshal(map[string]interface{}{
		"name":     "Phase 3 Student A",
		"email":    userEmailA,
		"password": password,
	})
	regReqA := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBodyA))
	regReqA.Header.Set("Content-Type", "application/json")
	regRespA, err := app.Test(regReqA, 5000)
	if err != nil || regRespA.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to register User A, status: %d", regRespA.StatusCode)
	}

	var envA APIEnvelope
	_ = json.NewDecoder(regRespA.Body).Decode(&envA)
	tokenA := envA.Data.AccessToken

	// 2. REGISTER & LOGIN USER B
	regBodyB, _ := json.Marshal(map[string]interface{}{
		"name":     "Phase 3 Student B",
		"email":    userEmailB,
		"password": password,
	})
	regReqB := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBodyB))
	regReqB.Header.Set("Content-Type", "application/json")
	regRespB, err := app.Test(regReqB, 5000)
	if err != nil || regRespB.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to register User B, status: %d", regRespB.StatusCode)
	}

	var envB APIEnvelope
	_ = json.NewDecoder(regRespB.Body).Decode(&envB)
	tokenB := envB.Data.AccessToken

	// 3. LOG PERSISTED HABIT DATA FOR USER A
	// 3.1 Study Session: Start and Finish 3600 seconds (1 hour)
	startStudyBody, _ := json.Marshal(map[string]interface{}{
		"subject": "Machine Learning Integration",
	})
	startStudyReq := httptest.NewRequest(http.MethodPost, "/api/v1/study/sessions/start", bytes.NewReader(startStudyBody))
	startStudyReq.Header.Set("Content-Type", "application/json")
	startStudyReq.Header.Set("Authorization", "Bearer "+tokenA)
	startStudyResp, _ := app.Test(startStudyReq, 5000)
	if startStudyResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to start study session, status: %d", startStudyResp.StatusCode)
	}

	var studyCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startStudyResp.Body).Decode(&studyCreated)
	studySessionID := studyCreated.Data.ID

	// Finish study session with 3600s
	finishStudyReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/study/sessions/%s/finish", studySessionID), nil)
	finishStudyReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishStudyResp, _ := app.Test(finishStudyReq, 5000)
	if finishStudyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to finish study session, status: %d", finishStudyResp.StatusCode)
	}
	// Explicitly set 3600 seconds for reliable test assertions
	_ = db.DB.Exec("UPDATE study_sessions SET duration_seconds = 3600 WHERE id = ?", studySessionID)

	// 3.2 Focus Interval: Start and Finish (25 min focus)
	startFocusBody, _ := json.Marshal(map[string]interface{}{
		"session_type": "focus",
	})
	startFocusReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/start", bytes.NewReader(startFocusBody))
	startFocusReq.Header.Set("Content-Type", "application/json")
	startFocusReq.Header.Set("Authorization", "Bearer "+tokenA)
	startFocusResp, _ := app.Test(startFocusReq, 5000)
	if startFocusResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to start focus interval, status: %d", startFocusResp.StatusCode)
	}

	var focusCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startFocusResp.Body).Decode(&focusCreated)
	focusSessionID := focusCreated.Data.ID

	// Finish focus interval
	finishFocusReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/focus/sessions/%s/finish", focusSessionID), nil)
	finishFocusReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishFocusResp, _ := app.Test(finishFocusReq, 5000)
	if finishFocusResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to finish focus interval, status: %d", finishFocusResp.StatusCode)
	}
	_ = db.DB.Exec("UPDATE focus_sessions SET duration_minutes = 25 WHERE id = ?", focusSessionID)

	// 3.3 Break Interval: Start and Finish (5 min break)
	startBreakBody, _ := json.Marshal(map[string]interface{}{
		"session_type": "break",
	})
	startBreakReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/start", bytes.NewReader(startBreakBody))
	startBreakReq.Header.Set("Content-Type", "application/json")
	startBreakReq.Header.Set("Authorization", "Bearer "+tokenA)
	startBreakResp, _ := app.Test(startBreakReq, 5000)
	if startBreakResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to start break interval, status: %d", startBreakResp.StatusCode)
	}

	var breakCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startBreakResp.Body).Decode(&breakCreated)
	breakSessionID := breakCreated.Data.ID

	finishBreakReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/focus/sessions/%s/finish", breakSessionID), nil)
	finishBreakReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishBreakResp, _ := app.Test(finishBreakReq, 5000)
	if finishBreakResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to finish break interval, status: %d", finishBreakResp.StatusCode)
	}
	_ = db.DB.Exec("UPDATE focus_sessions SET duration_minutes = 5 WHERE id = ?", breakSessionID)

	// 3.4 Task: Create and Mark Completed
	createTaskBody, _ := json.Marshal(map[string]interface{}{
		"title":    "Finish Phase 3 Verification",
		"category": "Study",
		"priority": "high",
	})
	createTaskReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(createTaskBody))
	createTaskReq.Header.Set("Content-Type", "application/json")
	createTaskReq.Header.Set("Authorization", "Bearer "+tokenA)
	createTaskResp, _ := app.Test(createTaskReq, 5000)
	if createTaskResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to create task, status: %d", createTaskResp.StatusCode)
	}

	var taskCreated struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(createTaskResp.Body).Decode(&taskCreated)
	taskID := taskCreated.Data.ID

	// Complete task
	toggleTaskReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/v1/tasks/%s/status", taskID), bytes.NewReader([]byte(`{"status":"completed"}`)))
	toggleTaskReq.Header.Set("Content-Type", "application/json")
	toggleTaskReq.Header.Set("Authorization", "Bearer "+tokenA)
	toggleTaskResp, _ := app.Test(toggleTaskReq, 5000)
	if toggleTaskResp.StatusCode != fiber.StatusOK {
		t.Fatalf("failed to mark task completed, status: %d", toggleTaskResp.StatusCode)
	}

	// 3.5 Lifestyle Record: Log today's lifestyle
	lifestyleBody, _ := json.Marshal(map[string]interface{}{
		"record_date":        time.Now().UTC().Format("2006-01-02"),
		"sleep_hours":        7.5,
		"phone_usage_hours":  2.0,
		"social_media_hours": 1.0,
		"stress_level":       4,
	})
	lifestyleReq := httptest.NewRequest(http.MethodPost, "/api/v1/lifestyle/record", bytes.NewReader(lifestyleBody))
	lifestyleReq.Header.Set("Content-Type", "application/json")
	lifestyleReq.Header.Set("Authorization", "Bearer "+tokenA)
	lifestyleResp, _ := app.Test(lifestyleReq, 5000)
	if lifestyleResp.StatusCode != fiber.StatusOK && lifestyleResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("failed to upsert lifestyle record, status: %d", lifestyleResp.StatusCode)
	}

	// -------------------------------------------------------------------------
	// 4. VERIFY ANALYTICS ENDPOINTS
	// -------------------------------------------------------------------------
	// 4.1 Daily Analytics
	dailyReq := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/daily", nil)
	dailyReq.Header.Set("Authorization", "Bearer "+tokenA)
	dailyResp, _ := app.Test(dailyReq, 5000)
	if dailyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /analytics/daily, got %d", dailyResp.StatusCode)
	}

	var dailyEnvelope struct {
		Data struct {
			TotalStudySeconds   int64 `json:"total_study_seconds"`
			TotalFocusMinutes   int64 `json:"total_focus_minutes"`
			TotalBreakMinutes   int64 `json:"total_break_minutes"`
			CompletedTasksCount int64 `json:"completed_tasks_count"`
		} `json:"data"`
	}
	_ = json.NewDecoder(dailyResp.Body).Decode(&dailyEnvelope)
	if dailyEnvelope.Data.TotalStudySeconds != 3600 {
		t.Errorf("expected 3600 study seconds, got %d", dailyEnvelope.Data.TotalStudySeconds)
	}
	if dailyEnvelope.Data.TotalFocusMinutes != 25 {
		t.Errorf("expected 25 focus minutes, got %d", dailyEnvelope.Data.TotalFocusMinutes)
	}
	if dailyEnvelope.Data.TotalBreakMinutes != 5 {
		t.Errorf("expected 5 break minutes, got %d", dailyEnvelope.Data.TotalBreakMinutes)
	}
	if dailyEnvelope.Data.CompletedTasksCount != 1 {
		t.Errorf("expected 1 completed task, got %d", dailyEnvelope.Data.CompletedTasksCount)
	}

	// 4.2 Weekly Analytics
	weeklyReq := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/weekly", nil)
	weeklyReq.Header.Set("Authorization", "Bearer "+tokenA)
	weeklyResp, _ := app.Test(weeklyReq, 5000)
	if weeklyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /analytics/weekly, got %d", weeklyResp.StatusCode)
	}

	// 4.3 Monthly Analytics
	monthlyReq := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/monthly", nil)
	monthlyReq.Header.Set("Authorization", "Bearer "+tokenA)
	monthlyResp, _ := app.Test(monthlyReq, 5000)
	if monthlyResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /analytics/monthly, got %d", monthlyResp.StatusCode)
	}

	// 4.4 Analytics Summary
	summaryReq := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/summary", nil)
	summaryReq.Header.Set("Authorization", "Bearer "+tokenA)
	summaryResp, _ := app.Test(summaryReq, 5000)
	if summaryResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /analytics/summary, got %d", summaryResp.StatusCode)
	}

	// 4.5 Root /api/analytics/summary alias
	rootAnalyticsReq := httptest.NewRequest(http.MethodGet, "/api/analytics/summary", nil)
	rootAnalyticsReq.Header.Set("Authorization", "Bearer "+tokenA)
	rootAnalyticsResp, _ := app.Test(rootAnalyticsReq, 5000)
	if rootAnalyticsResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on root alias /api/analytics/summary, got %d", rootAnalyticsResp.StatusCode)
	}

	// -------------------------------------------------------------------------
	// 5. VERIFY PREDICTION ENDPOINTS & OWNERSHIP ENFORCEMENT
	// -------------------------------------------------------------------------
	// 5.1 POST /api/v1/predictions (automatic aggregation from logged habits)
	predReq := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
	predReq.Header.Set("Content-Type", "application/json")
	predReq.Header.Set("Authorization", "Bearer "+tokenA)
	predResp, _ := app.Test(predReq, 5000)
	if predResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on /predictions, got %d", predResp.StatusCode)
	}

	var predEnvelope struct {
		Data struct {
			ID             string  `json:"id"`
			PredictedScore float64 `json:"predicted_score"`
			ModelVersion   string  `json:"model_version"`
			StudyHours     float64 `json:"study_hours"`
		} `json:"data"`
	}
	_ = json.NewDecoder(predResp.Body).Decode(&predEnvelope)
	predictionID := predEnvelope.Data.ID

	if predEnvelope.Data.PredictedScore < 0 || predEnvelope.Data.PredictedScore > 100 {
		t.Errorf("predicted score out of range [0, 100]: %v", predEnvelope.Data.PredictedScore)
	}
	if predEnvelope.Data.StudyHours != 1.0 {
		t.Errorf("expected 1.0 study hour aggregated from completed session, got %v", predEnvelope.Data.StudyHours)
	}

	// 5.2 GET /api/v1/predictions (History)
	listPredReq := httptest.NewRequest(http.MethodGet, "/api/v1/predictions?limit=5", nil)
	listPredReq.Header.Set("Authorization", "Bearer "+tokenA)
	listPredResp, _ := app.Test(listPredReq, 5000)
	if listPredResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /predictions list, got %d", listPredResp.StatusCode)
	}

	// 5.3 GET /api/v1/predictions/:id by Owner (User A) -> 200 OK
	getPredReqA := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/predictions/%s", predictionID), nil)
	getPredReqA.Header.Set("Authorization", "Bearer "+tokenA)
	getPredRespA, _ := app.Test(getPredReqA, 5000)
	if getPredRespA.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK when owner accesses prediction, got %d", getPredRespA.StatusCode)
	}

	// 5.4 GET /api/v1/predictions/:id by Unauthorized User B (IDOR Check) -> 404 NOT FOUND
	getPredReqB := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/predictions/%s", predictionID), nil)
	getPredReqB.Header.Set("Authorization", "Bearer "+tokenB)
	getPredRespB, _ := app.Test(getPredReqB, 5000)
	if getPredRespB.StatusCode != fiber.StatusNotFound {
		t.Errorf("CRITICAL SECURITY VIOLATION: expected 404 Not Found for cross-user IDOR access, got %d", getPredRespB.StatusCode)
	}

	// 5.5 Input validation error check: study_hours > 24
	invalidPredReq := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{"study_hours": 25.5}`)))
	invalidPredReq.Header.Set("Content-Type", "application/json")
	invalidPredReq.Header.Set("Authorization", "Bearer "+tokenA)
	invalidPredResp, _ := app.Test(invalidPredReq, 5000)
	if invalidPredResp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400 Bad Request on invalid study hours (>24), got %d", invalidPredResp.StatusCode)
	}
}
