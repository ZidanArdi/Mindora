package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/handlers"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/groq"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/response"
)

// Comprehensive Integration Test Suite for Mindora AI Chat
type flowGroqClient struct {
	Response    *groq.ChatCompletionResponse
	Err         error
	CapturedReq groq.ChatCompletionRequest
}

func (m *flowGroqClient) CreateChatCompletion(ctx context.Context, req groq.ChatCompletionRequest) (*groq.ChatCompletionResponse, error) {
	m.CapturedReq = req
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Response, nil
}

func (m *flowGroqClient) GetModel() string {
	return "openai/gpt-oss-20b"
}

type flowTaskRepo struct{}

func (m *flowTaskRepo) Create(ctx context.Context, task *models.Task) error { return nil }
func (m *flowTaskRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Task, error) {
	return nil, nil
}
func (m *flowTaskRepo) FindAll(ctx context.Context, userID uuid.UUID, filter repositories.TaskFilter) ([]models.Task, int64, error) {
	return nil, 0, nil
}
func (m *flowTaskRepo) Update(ctx context.Context, task *models.Task) error { return nil }
func (m *flowTaskRepo) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status string, completedAt *time.Time) (*models.Task, error) {
	return nil, nil
}
func (m *flowTaskRepo) SoftDelete(ctx context.Context, id, userID uuid.UUID) error { return nil }
func (m *flowTaskRepo) CountByStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *flowTaskRepo) CountCompletedByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *flowTaskRepo) FindCompletedTasksByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.Task, error) {
	return nil, nil
}

type flowStudyRepo struct{}

func (m *flowStudyRepo) CreateSession(ctx context.Context, session *models.StudySession) error {
	return nil
}
func (m *flowStudyRepo) FindSessionByID(ctx context.Context, id, userID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *flowStudyRepo) FindActiveOrPausedSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *flowStudyRepo) FindSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.StudySession, int64, error) {
	return nil, 0, nil
}
func (m *flowStudyRepo) FindCompletedSessions(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.StudySession, error) {
	return nil, nil
}
func (m *flowStudyRepo) UpdateSession(ctx context.Context, session *models.StudySession) error {
	return nil
}
func (m *flowStudyRepo) GetCompletedDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *flowStudyRepo) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*repositories.StudyDailyStats, error) {
	return &repositories.StudyDailyStats{}, nil
}

type flowFocusRepo struct{}

func (m *flowFocusRepo) CreateInterval(ctx context.Context, interval *models.FocusSession) error {
	return nil
}
func (m *flowFocusRepo) FindIntervalByID(ctx context.Context, id, userID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *flowFocusRepo) FindActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *flowFocusRepo) FindHistory(ctx context.Context, userID uuid.UUID, sessionType string, limit, offset int) ([]models.FocusSession, int64, error) {
	return nil, 0, nil
}
func (m *flowFocusRepo) FindCompletedIntervals(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.FocusSession, error) {
	return nil, nil
}
func (m *flowFocusRepo) UpdateInterval(ctx context.Context, interval *models.FocusSession) error {
	return nil
}
func (m *flowFocusRepo) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*repositories.FocusDailyStats, error) {
	return &repositories.FocusDailyStats{}, nil
}
func (m *flowFocusRepo) CountCompletedBreaks(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *flowFocusRepo) GetCompletedFocusDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}

type flowLifestyleRepo struct{}

func (m *flowLifestyleRepo) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error) {
	return nil, nil
}
func (m *flowLifestyleRepo) UpsertDailyRecord(ctx context.Context, record *models.LifestyleRecord) error {
	return nil
}
func (m *flowLifestyleRepo) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.LifestyleRecord, int64, error) {
	return nil, 0, nil
}
func (m *flowLifestyleRepo) FindByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.LifestyleRecord, error) {
	return nil, nil
}

type flowPredictionRepo struct {
	latest *models.ProductivityPrediction
	byID   map[uuid.UUID]*models.ProductivityPrediction
}

func (m *flowPredictionRepo) Create(ctx context.Context, p *models.ProductivityPrediction) error {
	return nil
}
func (m *flowPredictionRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	if m.byID != nil {
		if p, ok := m.byID[id]; ok && p.UserID == userID {
			return p, nil
		}
	}
	return nil, repositories.ErrPredictionNotFound
}
func (m *flowPredictionRepo) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ProductivityPrediction, int64, error) {
	return nil, 0, nil
}
func (m *flowPredictionRepo) FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	if m.latest != nil && m.latest.UserID == userID {
		return m.latest, nil
	}
	return nil, nil
}
func (m *flowPredictionRepo) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.ProductivityPrediction, error) {
	return nil, nil
}

type flowRecRepo struct{}

func (m *flowRecRepo) FindActive(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Recommendation, error) {
	return nil, nil
}
func (m *flowRecRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Recommendation, error) {
	return nil, nil
}
func (m *flowRecRepo) UpsertPending(ctx context.Context, rec *models.Recommendation) error {
	return nil
}
func (m *flowRecRepo) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status string, dismissedUntil *time.Time) (*models.Recommendation, error) {
	return nil, nil
}
func (m *flowRecRepo) FindDismissedRecent(ctx context.Context, userID uuid.UUID, ruleID string, now time.Time) (*models.Recommendation, error) {
	return nil, nil
}
func (m *flowRecRepo) FindByEvaluationKeyAndStatus(ctx context.Context, userID uuid.UUID, evalKey, status string) (*models.Recommendation, error) {
	return nil, nil
}
func (m *flowRecRepo) InvalidatePendingByRule(ctx context.Context, userID uuid.UUID, ruleID string) error {
	return nil
}
func (m *flowRecRepo) InvalidatePendingByID(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *flowRecRepo) ExpirePastValid(ctx context.Context, userID uuid.UUID, now time.Time) error {
	return nil
}

// Comprehensive Integration Test Suite for Mindora AI Chat
func TestAIFlow_Comprehensive(t *testing.T) {
	tokenMgr := jwt.NewTokenManager("mindora-super-test-jwt-secret-key-32chars", 15*time.Minute, 24*time.Hour)
	userA := uuid.New()
	userB := uuid.New()
	tokenA, _, _ := tokenMgr.GenerateAccessToken(userA, "studentA@mindora.app", "user")
	tokenB, _, _ := tokenMgr.GenerateAccessToken(userB, "studentB@mindora.app", "user")

	predA := &models.ProductivityPrediction{
		ID:             uuid.New(),
		UserID:         userA,
		PredictedScore: 85.0,
		StateLabel:     "Personalized Estimate",
	}

	predRepo := &flowPredictionRepo{
		latest: predA,
		byID:   map[uuid.UUID]*models.ProductivityPrediction{predA.ID: predA},
	}

	mockClient := &flowGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{
						Content: "Halo! Tetap jaga fokus dan kesehatan belajarmu ya.",
					},
					FinishReason: "stop",
				},
			},
			Usage: groq.UsageStats{TotalTokens: 60},
		},
	}

	aiService := services.NewAIService(
		mockClient,
		&flowTaskRepo{},
		&flowStudyRepo{},
		&flowFocusRepo{},
		&flowLifestyleRepo{},
		predRepo,
		&flowRecRepo{},
	)

	app := fiber.New()
	aiHandler := handlers.NewAIHandler(aiService)

	authMiddleware := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" || len(authHeader) < 8 {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
		}
		token := authHeader[7:]
		claims, err := tokenMgr.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Invalid or expired token"))
		}
		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)
		return c.Next()
	}

	aiGroup := app.Group("/api/v1/ai", authMiddleware)
	aiGroup.Post("/chat", aiHandler.Chat)
	aiGroup.Post("/explain-prediction", aiHandler.ExplainPrediction)
	aiGroup.Post("/study-plan", aiHandler.GenerateStudyPlan)

	// 1. Normal AI chat request
	t.Run("1_Normal_AI_Chat_Request", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"prompt": "Bagaimana tips menjaga stamina belajar?"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected app.Test error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// 2. Prediction explanation request
	t.Run("2_Prediction_Explanation_Request", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"prediction_id": predA.ID.String()})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/explain-prediction", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected app.Test error: %v", err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200, got %d", resp.StatusCode)
		}
	})

	// 3. Missing prediction data -> 404
	t.Run("3_Missing_Prediction_Data", func(t *testing.T) {
		// User B has no prediction
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/explain-prediction", bytes.NewReader([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("unexpected app.Test error: %v", err)
		}
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404 for missing prediction, got %d", resp.StatusCode)
		}
	})

	// 4. Invalid Groq API key -> 503
	t.Run("4_Invalid_Groq_API_Key", func(t *testing.T) {
		mockClient.Err = groq.ErrUnauthorized
		defer func() { mockClient.Err = nil }()

		body, _ := json.Marshal(map[string]string{"prompt": "Tes error"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("expected status 503 for invalid groq key, got %d", resp.StatusCode)
		}
	})

	// 5. Groq rate limit -> 429
	t.Run("5_Groq_Rate_Limit", func(t *testing.T) {
		mockClient.Err = groq.ErrRateLimitExceeded
		defer func() { mockClient.Err = nil }()

		body, _ := json.Marshal(map[string]string{"prompt": "Tes rate limit"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected status 429 for rate limit, got %d", resp.StatusCode)
		}
	})

	// 6. Groq timeout -> 504
	t.Run("6_Groq_Timeout", func(t *testing.T) {
		mockClient.Err = context.DeadlineExceeded
		defer func() { mockClient.Err = nil }()

		body, _ := json.Marshal(map[string]string{"prompt": "Tes timeout"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusGatewayTimeout {
			t.Fatalf("expected status 504 for timeout, got %d", resp.StatusCode)
		}
	})

	// 7. Malformed provider response -> handled safely by groq client & service
	t.Run("7_Malformed_Provider_Response", func(t *testing.T) {
		brokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("not-valid-json{{{"))
		}))
		defer brokenServer.Close()

		brokenClient := groq.NewClient("key", "openai/gpt-oss-20b", brokenServer.URL, 5*time.Second)
		brokenSvc := services.NewAIService(brokenClient, &flowTaskRepo{}, &flowStudyRepo{}, &flowFocusRepo{}, &flowLifestyleRepo{}, predRepo, &flowRecRepo{})
		brokenHandler := handlers.NewAIHandler(brokenSvc)

		brokenApp := fiber.New()
		brokenApp.Post("/chat", authMiddleware, brokenHandler.Chat)

		body, _ := json.Marshal(map[string]string{"prompt": "Tes broken json"})
		req := httptest.NewRequest(http.MethodPost, "/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, _ := brokenApp.Test(req)
		if resp.StatusCode != http.StatusInternalServerError {
			t.Fatalf("expected status 500 for unparseable groq json, got %d", resp.StatusCode)
		}
	})

	// 8. Empty assistant response -> fallback message
	t.Run("8_Empty_Assistant_Response", func(t *testing.T) {
		mockClient.Response = &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{Content: "   "},
				},
			},
		}
		defer func() {
			mockClient.Response = &groq.ChatCompletionResponse{
				Model: "openai/gpt-oss-20b",
				Choices: []groq.ChatChoice{
					{Message: groq.ChatMessage{Content: "Respons normal."}},
				},
			}
		}()

		body, _ := json.Marshal(map[string]string{"prompt": "Halo"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenA)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200 with fallback message, got %d", resp.StatusCode)
		}
	})

	// 9. Expired JWT -> 401
	t.Run("9_Expired_JWT", func(t *testing.T) {
		expiredMgr := jwt.NewTokenManager("mindora-super-test-jwt-secret-key-32chars", -1*time.Minute, 24*time.Hour)
		expiredToken, _, _ := expiredMgr.GenerateAccessToken(userA, "studentA@mindora.app", "user")

		body, _ := json.Marshal(map[string]string{"prompt": "Halo"})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("expected status 401 for expired JWT, got %d", resp.StatusCode)
		}
	})

	// 10. User attempting to access another user's AI context
	t.Run("10_Cross_User_Prediction_Access", func(t *testing.T) {
		// User B attempts to access User A's prediction
		body, _ := json.Marshal(map[string]string{"prediction_id": predA.ID.String()})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/explain-prediction", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+tokenB)

		resp, _ := app.Test(req)
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected status 404 for accessing another user's prediction, got %d", resp.StatusCode)
		}
	})

	// 11. Retry after provider failure
	t.Run("11_Retry_After_Provider_Failure", func(t *testing.T) {
		// First attempt fails with rate limit
		mockClient.Err = groq.ErrRateLimitExceeded
		body, _ := json.Marshal(map[string]string{"prompt": "Halo lagi"})
		req1 := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req1.Header.Set("Content-Type", "application/json")
		req1.Header.Set("Authorization", "Bearer "+tokenA)

		resp1, _ := app.Test(req1)
		if resp1.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("expected first request to be 429, got %d", resp1.StatusCode)
		}

		// Second attempt succeeds when provider recovers
		mockClient.Err = nil
		req2 := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+tokenA)

		resp2, _ := app.Test(req2)
		if resp2.StatusCode != http.StatusOK {
			t.Fatalf("expected second request to succeed (200), got %d", resp2.StatusCode)
		}
	})
}
