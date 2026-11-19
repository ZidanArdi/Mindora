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

func TestCoreProductivityFlow(t *testing.T) {
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
		_ = db.DB.Exec("DELETE FROM tasks WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'prod.test%')")
		_ = db.DB.Exec("DELETE FROM study_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'prod.test%')")
		_ = db.DB.Exec("DELETE FROM focus_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'prod.test%')")
		_ = db.DB.Exec("DELETE FROM lifestyle_records WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'prod.test%')")
		_ = db.DB.Exec("DELETE FROM auth_sessions WHERE user_id IN (SELECT id FROM users WHERE email LIKE 'prod.test%')")
		_ = db.DB.Exec("DELETE FROM users WHERE email LIKE 'prod.test%'")
	}()

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano()%1000000)
	userEmailA := fmt.Sprintf("prod.test.user.a.%s@mindora.app", uniqueSuffix)
	userEmailB := fmt.Sprintf("prod.test.user.b.%s@mindora.app", uniqueSuffix)
	password := "SecureProductivity2026!"

	// 1. REGISTER & LOGIN USER A
	regBodyA, _ := json.Marshal(map[string]interface{}{
		"name":     "Productivity Student A",
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

	// 2. REGISTER & LOGIN USER B (For IDOR / Ownership Checks)
	regBodyB, _ := json.Marshal(map[string]interface{}{
		"name":     "Productivity Student B",
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

	// -------------------------------------------------------------
	// 3. TASK MANAGEMENT FLOW
	// -------------------------------------------------------------
	// 3.1 Create Task
	createTaskBody, _ := json.Marshal(map[string]interface{}{
		"title":       "Complete Operating Systems Lab 3",
		"description": "Kernel memory allocation and page table implementation",
		"category":    "Study",
		"priority":    "high",
		"deadline":    time.Now().Add(48 * time.Hour).Format(time.RFC3339),
	})
	createTaskReq := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewReader(createTaskBody))
	createTaskReq.Header.Set("Content-Type", "application/json")
	createTaskReq.Header.Set("Authorization", "Bearer "+tokenA)
	createTaskResp, _ := app.Test(createTaskReq, 5000)
	if createTaskResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on task creation, got %d", createTaskResp.StatusCode)
	}

	var taskCreatedEnvelope struct {
		Success bool `json:"success"`
		Data    struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Priority string `json:"priority"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	_ = json.NewDecoder(createTaskResp.Body).Decode(&taskCreatedEnvelope)
	taskAID := taskCreatedEnvelope.Data.ID

	if taskCreatedEnvelope.Data.Title != "Complete Operating Systems Lab 3" {
		t.Errorf("unexpected task title: %s", taskCreatedEnvelope.Data.Title)
	}

	// 3.2 Cross-User Task Access (User B attempting to view User A's task -> 404 Not Found)
	crossTaskReq := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+taskAID, nil)
	crossTaskReq.Header.Set("Authorization", "Bearer "+tokenB)
	crossTaskResp, _ := app.Test(crossTaskReq, 5000)
	if crossTaskResp.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected 404 for cross-user task access, got %d", crossTaskResp.StatusCode)
	}

	// 3.3 Update Task Status to "completed"
	statusBody, _ := json.Marshal(map[string]interface{}{"status": "completed"})
	patchReq := httptest.NewRequest(http.MethodPatch, "/api/v1/tasks/"+taskAID+"/status", bytes.NewReader(statusBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchReq.Header.Set("Authorization", "Bearer "+tokenA)
	patchResp, _ := app.Test(patchReq, 5000)
	if patchResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on task status patch, got %d", patchResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 4. STUDY TRACKING FLOW
	// -------------------------------------------------------------
	// 4.1 Start Study Session
	startStudyBody, _ := json.Marshal(map[string]interface{}{
		"subject": "Distributed Systems Raft Consensus",
		"task_id": taskAID,
	})
	startStudyReq := httptest.NewRequest(http.MethodPost, "/api/v1/study/sessions/start", bytes.NewReader(startStudyBody))
	startStudyReq.Header.Set("Content-Type", "application/json")
	startStudyReq.Header.Set("Authorization", "Bearer "+tokenA)
	startStudyResp, _ := app.Test(startStudyReq, 5000)
	if startStudyResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on start study session, got %d", startStudyResp.StatusCode)
	}

	var studyCreatedEnvelope struct {
		Success bool `json:"success"`
		Data    struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startStudyResp.Body).Decode(&studyCreatedEnvelope)
	studySessionID := studyCreatedEnvelope.Data.ID

	// 4.2 Pause Study Session
	pauseReq := httptest.NewRequest(http.MethodPost, "/api/v1/study/sessions/"+studySessionID+"/pause", nil)
	pauseReq.Header.Set("Authorization", "Bearer "+tokenA)
	pauseResp, _ := app.Test(pauseReq, 5000)
	if pauseResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on pause study session, got %d", pauseResp.StatusCode)
	}

	// 4.3 Resume Study Session
	resumeReq := httptest.NewRequest(http.MethodPost, "/api/v1/study/sessions/"+studySessionID+"/resume", nil)
	resumeReq.Header.Set("Authorization", "Bearer "+tokenA)
	resumeResp, _ := app.Test(resumeReq, 5000)
	if resumeResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on resume study session, got %d", resumeResp.StatusCode)
	}

	// 4.4 Finish Study Session
	finishStudyReq := httptest.NewRequest(http.MethodPost, "/api/v1/study/sessions/"+studySessionID+"/finish", nil)
	finishStudyReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishStudyResp, _ := app.Test(finishStudyReq, 5000)
	if finishStudyResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on finish study session, got %d", finishStudyResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 5. FOCUS INTERVAL TRACKING (Individual Interval Architecture)
	// -------------------------------------------------------------
	// 5.1 Start Focus Interval
	startFocusBody, _ := json.Marshal(map[string]interface{}{
		"session_type": "focus",
		"task_id":      taskAID,
	})
	startFocusReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/start", bytes.NewReader(startFocusBody))
	startFocusReq.Header.Set("Content-Type", "application/json")
	startFocusReq.Header.Set("Authorization", "Bearer "+tokenA)
	startFocusResp, _ := app.Test(startFocusReq, 5000)
	if startFocusResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on start focus interval, got %d", startFocusResp.StatusCode)
	}

	var focusCreatedEnvelope struct {
		Success bool `json:"success"`
		Data    struct {
			ID          string `json:"id"`
			SessionType string `json:"session_type"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startFocusResp.Body).Decode(&focusCreatedEnvelope)
	focusIntervalID := focusCreatedEnvelope.Data.ID

	// 5.2 Log Interruption
	interrReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/"+focusIntervalID+"/interruption", nil)
	interrReq.Header.Set("Authorization", "Bearer "+tokenA)
	interrResp, _ := app.Test(interrReq, 5000)
	if interrResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on logging interruption, got %d", interrResp.StatusCode)
	}

	// 5.3 Finish Focus Interval
	finishFocusReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/"+focusIntervalID+"/finish", nil)
	finishFocusReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishFocusResp, _ := app.Test(finishFocusReq, 5000)
	if finishFocusResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on finish focus interval, got %d", finishFocusResp.StatusCode)
	}

	// 5.4 Start & Finish Break Interval (Separate row verification)
	startBreakBody, _ := json.Marshal(map[string]interface{}{
		"session_type": "break",
	})
	startBreakReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/start", bytes.NewReader(startBreakBody))
	startBreakReq.Header.Set("Content-Type", "application/json")
	startBreakReq.Header.Set("Authorization", "Bearer "+tokenA)
	startBreakResp, _ := app.Test(startBreakReq, 5000)
	if startBreakResp.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on break interval, got %d", startBreakResp.StatusCode)
	}

	var breakCreatedEnvelope struct {
		Success bool `json:"success"`
		Data    struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	_ = json.NewDecoder(startBreakResp.Body).Decode(&breakCreatedEnvelope)

	finishBreakReq := httptest.NewRequest(http.MethodPost, "/api/v1/focus/sessions/"+breakCreatedEnvelope.Data.ID+"/finish", nil)
	finishBreakReq.Header.Set("Authorization", "Bearer "+tokenA)
	finishBreakResp, _ := app.Test(finishBreakReq, 5000)
	if finishBreakResp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 OK on finish break interval, got %d", finishBreakResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 6. DAILY LIFESTYLE HABIT LOGGING (8 Variables)
	// -------------------------------------------------------------
	todayStr := time.Now().UTC().Format("2006-01-02")
	sleepH := 7.5
	phoneH := 2.2
	socialH := 1.0
	ytH := 1.5
	gamingH := 0.5
	coffeeMg := 150
	exerciseM := 45
	stress := 4
	notes := "Productive morning study block completed"

	lifestyleBody, _ := json.Marshal(map[string]interface{}{
		"record_date":         todayStr,
		"sleep_hours":         sleepH,
		"phone_usage_hours":   phoneH,
		"social_media_hours":  socialH,
		"youtube_hours":       ytH,
		"gaming_hours":        gamingH,
		"coffee_intake_mg":    coffeeMg,
		"exercise_minutes":    exerciseM,
		"stress_level":        stress,
		"notes":               notes,
	})
	lifestyleReq := httptest.NewRequest(http.MethodPost, "/api/v1/lifestyle/record", bytes.NewReader(lifestyleBody))
	lifestyleReq.Header.Set("Content-Type", "application/json")
	lifestyleReq.Header.Set("Authorization", "Bearer "+tokenA)
	lifestyleResp, _ := app.Test(lifestyleReq, 5000)
	if lifestyleResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on lifestyle upsert, got %d", lifestyleResp.StatusCode)
	}

	// 6.1 Validation Check: Stress Level > 10 must be rejected
	invalidStress := 15
	badLifestyleBody, _ := json.Marshal(map[string]interface{}{
		"record_date":  todayStr,
		"stress_level": invalidStress,
	})
	badLifestyleReq := httptest.NewRequest(http.MethodPost, "/api/v1/lifestyle/record", bytes.NewReader(badLifestyleBody))
	badLifestyleReq.Header.Set("Content-Type", "application/json")
	badLifestyleReq.Header.Set("Authorization", "Bearer "+tokenA)
	badLifestyleResp, _ := app.Test(badLifestyleReq, 5000)
	if badLifestyleResp.StatusCode != fiber.StatusUnprocessableEntity {
		t.Errorf("expected 422 for stress level > 10, got %d", badLifestyleResp.StatusCode)
	}

	// -------------------------------------------------------------
	// 7. REAL DASHBOARD SUMMARY INTEGRATION
	// -------------------------------------------------------------
	dashboardReq := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/summary", nil)
	dashboardReq.Header.Set("Authorization", "Bearer "+tokenA)
	dashboardResp, _ := app.Test(dashboardReq, 5000)
	if dashboardResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on dashboard summary, got %d", dashboardResp.StatusCode)
	}

	var dashEnvelope struct {
		Success bool `json:"success"`
		Data    struct {
			TotalTasks           int64   `json:"total_tasks"`
			CompletedTasks       int64   `json:"completed_tasks"`
			CompletedBreaksCount int64   `json:"completed_breaks_count"`
			TodayLifestyleLogged bool    `json:"today_lifestyle_logged"`
		} `json:"data"`
	}
	_ = json.NewDecoder(dashboardResp.Body).Decode(&dashEnvelope)

	if dashEnvelope.Data.TotalTasks < 1 {
		t.Errorf("expected total tasks >= 1, got %d", dashEnvelope.Data.TotalTasks)
	}
	if dashEnvelope.Data.CompletedTasks < 1 {
		t.Errorf("expected completed tasks >= 1, got %d", dashEnvelope.Data.CompletedTasks)
	}
	if dashEnvelope.Data.CompletedBreaksCount < 1 {
		t.Errorf("expected completed breaks >= 1, got %d", dashEnvelope.Data.CompletedBreaksCount)
	}
	if !dashEnvelope.Data.TodayLifestyleLogged {
		t.Errorf("expected today lifestyle to be logged")
	}
}
