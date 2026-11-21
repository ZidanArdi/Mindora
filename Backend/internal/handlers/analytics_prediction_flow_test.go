package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/handlers"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/routes"
	"mindora/backend/internal/services"
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

	// 4.6 Analytics Range (Reports Module)
	rangeReq := httptest.NewRequest(http.MethodGet, "/api/v1/analytics/range?start_date=2026-11-15&end_date=2026-11-21&tz=UTC", nil)
	rangeReq.Header.Set("Authorization", "Bearer "+tokenA)
	rangeResp, _ := app.Test(rangeReq, 5000)
	if rangeResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on /analytics/range, got %d", rangeResp.StatusCode)
	}

	// 4.7 Root /api/analytics/range alias
	rootRangeReq := httptest.NewRequest(http.MethodGet, "/api/analytics/range?start_date=2026-11-15&end_date=2026-11-21&tz=UTC", nil)
	rootRangeReq.Header.Set("Authorization", "Bearer "+tokenA)
	rootRangeResp, _ := app.Test(rootRangeReq, 5000)
	if rootRangeResp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200 OK on root alias /api/analytics/range, got %d", rootRangeResp.StatusCode)
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
			ID                 string   `json:"id"`
			PredictedScore     float64  `json:"predicted_score"`
			ModelVersion       string   `json:"model_version"`
			StudyHours         float64  `json:"study_hours"`
			PredictionState    string   `json:"prediction_state"`
			StateLabel         string   `json:"state_label"`
			StateDescription   string   `json:"state_description"`
			IsBaseline         bool     `json:"is_baseline"`
			UnrecordedFeatures []string `json:"unrecorded_features"`
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
	if predEnvelope.Data.IsBaseline != false {
		t.Errorf("expected Personalized Estimate for User A (is_baseline=false), got true")
	}
	if predEnvelope.Data.PredictionState != "personalized" {
		t.Errorf("expected prediction_state 'personalized', got %v", predEnvelope.Data.PredictionState)
	}
	if predEnvelope.Data.StateLabel != "Personalized Estimate" {
		t.Errorf("expected state_label 'Personalized Estimate', got %v", predEnvelope.Data.StateLabel)
	}
	if len(predEnvelope.Data.UnrecordedFeatures) != 0 {
		t.Errorf("expected 0 unrecorded features for fully tracked User A, got %v", predEnvelope.Data.UnrecordedFeatures)
	}

	// 5.1b POST /api/v1/predictions for User B (Cold Start: zero tracking records)
	predReqCold := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
	predReqCold.Header.Set("Content-Type", "application/json")
	predReqCold.Header.Set("Authorization", "Bearer "+tokenB)
	predRespCold, _ := app.Test(predReqCold, 5000)
	if predRespCold.StatusCode != fiber.StatusCreated {
		t.Fatalf("expected 201 Created on User B cold-start /predictions, got %d", predRespCold.StatusCode)
	}

	var predEnvelopeCold struct {
		Data struct {
			PredictedScore     float64  `json:"predicted_score"`
			PredictionState    string   `json:"prediction_state"`
			StateLabel         string   `json:"state_label"`
			IsBaseline         bool     `json:"is_baseline"`
			UnrecordedFeatures []string `json:"unrecorded_features"`
		} `json:"data"`
	}
	_ = json.NewDecoder(predRespCold.Body).Decode(&predEnvelopeCold)
	if !predEnvelopeCold.Data.IsBaseline {
		t.Errorf("expected User B cold start to be Baseline Estimate (is_baseline=true)")
	}
	if predEnvelopeCold.Data.PredictionState != "baseline" {
		t.Errorf("expected User B state 'baseline', got %s", predEnvelopeCold.Data.PredictionState)
	}
	if predEnvelopeCold.Data.StateLabel != "Baseline Estimate" {
		t.Errorf("expected User B label 'Baseline Estimate', got %s", predEnvelopeCold.Data.StateLabel)
	}
	if len(predEnvelopeCold.Data.UnrecordedFeatures) != 4 {
		t.Errorf("expected 4 unrecorded features for cold start User B, got %v", predEnvelopeCold.Data.UnrecordedFeatures)
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

// Mocks for Handler Repository Error Tests (Phase 3.2.1)
type mockHandlerPredRepo struct {
	records []models.ProductivityPrediction
}

func (m *mockHandlerPredRepo) Create(ctx context.Context, p *models.ProductivityPrediction) error {
	m.records = append(m.records, *p)
	return nil
}
func (m *mockHandlerPredRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	return nil, nil
}
func (m *mockHandlerPredRepo) FindHistory(ctx context.Context, userID uuid.UUID, l, o int) ([]models.ProductivityPrediction, int64, error) {
	return nil, 0, nil
}
func (m *mockHandlerPredRepo) FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	return nil, nil
}
func (m *mockHandlerPredRepo) FindByDate(ctx context.Context, userID uuid.UUID, d time.Time) (*models.ProductivityPrediction, error) {
	return nil, nil
}

type mockHandlerStudyRepo struct {
	returnNilStats bool
	err            error
}

func (m *mockHandlerStudyRepo) CreateSession(ctx context.Context, s *models.StudySession) error {
	return nil
}
func (m *mockHandlerStudyRepo) FindSessionByID(ctx context.Context, id, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockHandlerStudyRepo) FindActiveOrPausedSession(ctx context.Context, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockHandlerStudyRepo) FindSessions(ctx context.Context, uID uuid.UUID, s, e *time.Time, l, o int) ([]models.StudySession, int64, error) {
	return nil, 0, nil
}
func (m *mockHandlerStudyRepo) FindCompletedSessions(ctx context.Context, uID uuid.UUID, s, e time.Time) ([]models.StudySession, error) {
	return nil, nil
}
func (m *mockHandlerStudyRepo) UpdateSession(ctx context.Context, s *models.StudySession) error {
	return nil
}
func (m *mockHandlerStudyRepo) GetCompletedDuration(ctx context.Context, uID uuid.UUID, s, e time.Time) (int64, error) {
	return 0, nil
}
func (m *mockHandlerStudyRepo) GetDailyStatistics(ctx context.Context, uID uuid.UUID, s, e time.Time) (*repositories.StudyDailyStats, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.returnNilStats {
		return nil, nil
	}
	return &repositories.StudyDailyStats{TotalCompletedSeconds: 0, CompletedSessionsCount: 0}, nil
}

type mockHandlerLifestyleRepo struct {
	err    error
	record *models.LifestyleRecord
}

func (m *mockHandlerLifestyleRepo) FindByDate(ctx context.Context, uID uuid.UUID, d time.Time) (*models.LifestyleRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.record, nil
}
func (m *mockHandlerLifestyleRepo) UpsertDailyRecord(ctx context.Context, r *models.LifestyleRecord) error {
	return nil
}
func (m *mockHandlerLifestyleRepo) FindHistory(ctx context.Context, uID uuid.UUID, l, o int) ([]models.LifestyleRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockHandlerLifestyleRepo) FindByDateRange(ctx context.Context, uID uuid.UUID, s, e time.Time) ([]models.LifestyleRecord, error) {
	return nil, nil
}

// TestPredictionHandler_RepositoryDatabaseErrorHandling validates that repository failures
// return 500 Internal Server Error, do not leak internal database details, log errors server-side,
// do not persist predictions, and handle wrapped input validation errors as 400.
func TestPredictionHandler_RepositoryDatabaseErrorHandling(t *testing.T) {
	testUserID := uuid.New()

	t.Run("Study repo query failure returns generic 500, hides DB details, and logs error", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(os.Stderr)

		predRepo := &mockHandlerPredRepo{}
		studyRepo := &mockHandlerStudyRepo{err: errors.New("pq: connection reset by peer")}
		lifestyleRepo := &mockHandlerLifestyleRepo{}

		predSvc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
		handler := handlers.NewPredictionHandler(predSvc)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", testUserID)
			return c.Next()
		})
		app.Post("/api/v1/predictions", handler.Create)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 2000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("expected status 500 Internal Server Error, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body["success"] != false {
			t.Errorf("expected success: false in response, got %v", body["success"])
		}
		msg, _ := body["message"].(string)
		if msg != "Failed to compute productivity prediction" {
			t.Errorf("expected generic error message 'Failed to compute productivity prediction', got: %s", msg)
		}
		// Security Assertion: Ensure internal database details are NOT exposed to HTTP client
		if strings.Contains(msg, "pq:") || strings.Contains(msg, "failed to retrieve") {
			t.Errorf("SECURITY LEAK: raw database error exposed in HTTP response body: %s", msg)
		}

		// Security Assertion: Ensure detailed error IS captured in server logs
		if !strings.Contains(logBuf.String(), "pq: connection reset by peer") {
			t.Errorf("expected detailed error logged server-side, got log: %s", logBuf.String())
		}
		if !strings.Contains(logBuf.String(), testUserID.String()) {
			t.Errorf("expected user ID in server log, got log: %s", logBuf.String())
		}

		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions saved in database on study repo failure, got %d", len(predRepo.records))
		}
	})

	t.Run("Lifestyle repo query failure returns generic 500, hides DB details, and logs error", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(os.Stderr)

		predRepo := &mockHandlerPredRepo{}
		studyRepo := &mockHandlerStudyRepo{}
		lifestyleRepo := &mockHandlerLifestyleRepo{err: errors.New("pq: deadlock detected")}

		predSvc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
		handler := handlers.NewPredictionHandler(predSvc)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", testUserID)
			return c.Next()
		})
		app.Post("/api/v1/predictions", handler.Create)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 2000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("expected status 500 Internal Server Error, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body["success"] != false {
			t.Errorf("expected success: false in response, got %v", body["success"])
		}
		msg, _ := body["message"].(string)
		if msg != "Failed to compute productivity prediction" {
			t.Errorf("expected generic error message 'Failed to compute productivity prediction', got: %s", msg)
		}
		// Security Assertion: Ensure internal database details are NOT exposed to HTTP client
		if strings.Contains(msg, "pq:") || strings.Contains(msg, "failed to retrieve") {
			t.Errorf("SECURITY LEAK: raw database error exposed in HTTP response body: %s", msg)
		}

		// Security Assertion: Ensure detailed error IS captured in server logs
		if !strings.Contains(logBuf.String(), "pq: deadlock detected") {
			t.Errorf("expected detailed error logged server-side, got log: %s", logBuf.String())
		}

		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions saved in database on lifestyle repo failure, got %d", len(predRepo.records))
		}
	})

	t.Run("Unexpected nil study statistics returns generic 500 without panic", func(t *testing.T) {
		var logBuf bytes.Buffer
		log.SetOutput(&logBuf)
		defer log.SetOutput(os.Stderr)

		predRepo := &mockHandlerPredRepo{}
		studyRepo := &mockHandlerStudyRepo{returnNilStats: true}
		lifestyleRepo := &mockHandlerLifestyleRepo{}

		predSvc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
		handler := handlers.NewPredictionHandler(predSvc)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", testUserID)
			return c.Next()
		})
		app.Post("/api/v1/predictions", handler.Create)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 2000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("expected status 500 Internal Server Error on nil statistics, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		msg, _ := body["message"].(string)
		if msg != "Failed to compute productivity prediction" {
			t.Errorf("expected generic error message, got: %s", msg)
		}

		// Verify logged error indicates unexpected nil statistics
		if !strings.Contains(logBuf.String(), "unexpected nil statistics") {
			t.Errorf("expected nil statistics detail logged server-side, got log: %s", logBuf.String())
		}

		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted on nil statistics error, got %d", len(predRepo.records))
		}
	})

	t.Run("Wrapped invalid input error returns HTTP 400", func(t *testing.T) {
		predRepo := &mockHandlerPredRepo{}
		studyRepo := &mockHandlerStudyRepo{}
		lifestyleRepo := &mockHandlerLifestyleRepo{}

		predSvc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
		handler := handlers.NewPredictionHandler(predSvc)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", testUserID)
			return c.Next()
		})
		app.Post("/api/v1/predictions", handler.Create)

		// Send invalid study hours (> 24)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{"study_hours": 26.0}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 2000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected status 400 Bad Request, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body["success"] != false {
			t.Errorf("expected success: false, got %v", body["success"])
		}
		msg, _ := body["message"].(string)
		if msg != "Invalid input parameter range for prediction features" {
			t.Errorf("expected invalid parameter range message, got: %s", msg)
		}

		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted on invalid input, got %d", len(predRepo.records))
		}
	})

	t.Run("Lifestyle repo ErrLifestyleRecordNotFound succeeds and creates Baseline Estimate", func(t *testing.T) {
		predRepo := &mockHandlerPredRepo{}
		studyRepo := &mockHandlerStudyRepo{}
		lifestyleRepo := &mockHandlerLifestyleRepo{err: repositories.ErrLifestyleRecordNotFound}

		predSvc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
		handler := handlers.NewPredictionHandler(predSvc)

		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("userID", testUserID)
			return c.Next()
		})
		app.Post("/api/v1/predictions", handler.Create)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/predictions", bytes.NewReader([]byte(`{}`)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req, 2000)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusCreated {
			t.Errorf("expected status 201 Created on missing lifestyle record, got %d", resp.StatusCode)
		}

		var body map[string]interface{}
		_ = json.NewDecoder(resp.Body).Decode(&body)
		if body["success"] != true {
			t.Errorf("expected success: true, got %v", body["success"])
		}
		data, ok := body["data"].(map[string]interface{})
		if !ok || data["is_baseline"] != true {
			t.Errorf("expected is_baseline: true for missing data, got %v", data["is_baseline"])
		}

		if len(predRepo.records) != 1 {
			t.Errorf("expected 1 prediction persisted, got %d", len(predRepo.records))
		}
	})
}
