package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/handlers"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/groq"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/response"
)

// MockAIService for handler unit tests
type mockAIService struct {
	chatFunc            func(ctx context.Context, userID uuid.UUID, prompt string) (*services.AIResponse, error)
	explainFunc         func(ctx context.Context, userID uuid.UUID, predictionID *uuid.UUID) (*services.AIResponse, error)
	generateStudyPlanFn func(ctx context.Context, userID uuid.UUID) (*services.AIResponse, error)
}

func (m *mockAIService) Chat(ctx context.Context, userID uuid.UUID, prompt string) (*services.AIResponse, error) {
	if m.chatFunc != nil {
		return m.chatFunc(ctx, userID, prompt)
	}
	return &services.AIResponse{Reply: "Jawaban asisten AI", Model: "openai/gpt-oss-20b", TokensUsed: 50}, nil
}

func (m *mockAIService) ExplainPrediction(ctx context.Context, userID uuid.UUID, predictionID *uuid.UUID) (*services.AIResponse, error) {
	if m.explainFunc != nil {
		return m.explainFunc(ctx, userID, predictionID)
	}
	return &services.AIResponse{Reply: "Penjelasan prediksi", Model: "openai/gpt-oss-20b", TokensUsed: 80}, nil
}

func (m *mockAIService) GenerateStudyPlan(ctx context.Context, userID uuid.UUID) (*services.AIResponse, error) {
	if m.generateStudyPlanFn != nil {
		return m.generateStudyPlanFn(ctx, userID)
	}
	return &services.AIResponse{Reply: "Jadwal belajar", Model: "openai/gpt-oss-20b", TokensUsed: 100}, nil
}

func setupTestApp(svc services.AIService, tokenMgr *jwt.TokenManager) *fiber.App {
	app := fiber.New()
	handler := handlers.NewAIHandler(svc)

	// In-memory mock middleware injecting userID
	authMiddleware := func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
		}
		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := tokenMgr.ValidateAccessToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Invalid token"))
		}
		c.Locals("userID", claims.UserID)
		c.Locals("userRole", claims.Role)
		return c.Next()
	}

	ai := app.Group("/api/v1/ai", authMiddleware)
	ai.Post("/chat", handler.Chat)
	ai.Post("/explain-prediction", handler.ExplainPrediction)
	ai.Post("/study-plan", handler.GenerateStudyPlan)

	return app
}

func TestAIHandler_Chat(t *testing.T) {
	tokenMgr := jwt.NewTokenManager("test-secret-key-32chars-minimum-12345", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	token, _, _ := tokenMgr.GenerateAccessToken(userID, "student@mindora.app", "user")

	mockSvc := &mockAIService{
		chatFunc: func(ctx context.Context, uID uuid.UUID, prompt string) (*services.AIResponse, error) {
			if prompt == "rate_limit_me" {
				return nil, groq.ErrRateLimitExceeded
			}
			return &services.AIResponse{
				Reply:      "Halo dari AI Coach!",
				Model:      "openai/gpt-oss-20b",
				TokensUsed: 75,
			}, nil
		},
	}

	app := setupTestApp(mockSvc, tokenMgr)

	// 1. Success case
	body, _ := json.Marshal(map[string]string{"prompt": "Bagaimana tips belajar kalkulus?"})
	req := httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// 2. Empty prompt validation -> 400
	emptyBody, _ := json.Marshal(map[string]string{"prompt": "   "})
	reqEmpty := httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewReader(emptyBody))
	reqEmpty.Header.Set("Content-Type", "application/json")
	reqEmpty.Header.Set("Authorization", "Bearer "+token)

	respEmpty, _ := app.Test(reqEmpty)
	if respEmpty.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400 for empty prompt, got %d", respEmpty.StatusCode)
	}

	// 3. Prompt too long -> 400
	longPrompt := strings.Repeat("x", 1001)
	longBody, _ := json.Marshal(map[string]string{"prompt": longPrompt})
	reqLong := httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewReader(longBody))
	reqLong.Header.Set("Content-Type", "application/json")
	reqLong.Header.Set("Authorization", "Bearer "+token)

	respLong, _ := app.Test(reqLong)
	if respLong.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400 for prompt > 1000 chars, got %d", respLong.StatusCode)
	}

	// 4. Rate limit mapping -> 429
	rateLimitBody, _ := json.Marshal(map[string]string{"prompt": "rate_limit_me"})
	reqRate := httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewReader(rateLimitBody))
	reqRate.Header.Set("Content-Type", "application/json")
	reqRate.Header.Set("Authorization", "Bearer "+token)

	respRate, _ := app.Test(reqRate)
	if respRate.StatusCode != fiber.StatusTooManyRequests {
		t.Errorf("expected status 429 for rate limit, got %d", respRate.StatusCode)
	}

	// 5. Unauthorized -> 401
	reqUnauth := httptest.NewRequest("POST", "/api/v1/ai/chat", bytes.NewReader(body))
	reqUnauth.Header.Set("Content-Type", "application/json")
	respUnauth, _ := app.Test(reqUnauth)
	if respUnauth.StatusCode != fiber.StatusUnauthorized {
		t.Errorf("expected status 401 when no token provided, got %d", respUnauth.StatusCode)
	}
}

func TestAIHandler_ExplainPrediction(t *testing.T) {
	tokenMgr := jwt.NewTokenManager("test-secret-key-32chars-minimum-12345", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	token, _, _ := tokenMgr.GenerateAccessToken(userID, "student@mindora.app", "user")

	predID := uuid.New()
	mockSvc := &mockAIService{
		explainFunc: func(ctx context.Context, uID uuid.UUID, predictionID *uuid.UUID) (*services.AIResponse, error) {
			if predictionID != nil && *predictionID == uuid.Nil {
				return nil, services.ErrAIPredictionNotFound
			}
			return &services.AIResponse{
				Reply: "Skor Anda dipengaruhi oleh durasi tidur 7 jam.",
				Model: "openai/gpt-oss-20b",
			}, nil
		},
	}

	app := setupTestApp(mockSvc, tokenMgr)

	// 1. Success case with ID
	body, _ := json.Marshal(map[string]string{"prediction_id": predID.String()})
	req := httptest.NewRequest("POST", "/api/v1/ai/explain-prediction", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	// 2. Invalid UUID in body -> 400
	badBody, _ := json.Marshal(map[string]string{"prediction_id": "invalid-uuid"})
	reqBad := httptest.NewRequest("POST", "/api/v1/ai/explain-prediction", bytes.NewReader(badBody))
	reqBad.Header.Set("Content-Type", "application/json")
	reqBad.Header.Set("Authorization", "Bearer "+token)

	respBad, _ := app.Test(reqBad)
	if respBad.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected status 400 for bad UUID, got %d", respBad.StatusCode)
	}

	// 3. Not found -> 404
	mockSvc.explainFunc = func(ctx context.Context, uID uuid.UUID, predictionID *uuid.UUID) (*services.AIResponse, error) {
		return nil, services.ErrAIPredictionNotFound
	}
	reqNotFound := httptest.NewRequest("POST", "/api/v1/ai/explain-prediction", bytes.NewReader(body))
	reqNotFound.Header.Set("Content-Type", "application/json")
	reqNotFound.Header.Set("Authorization", "Bearer "+token)

	respNotFound, _ := app.Test(reqNotFound)
	if respNotFound.StatusCode != fiber.StatusNotFound {
		t.Errorf("expected status 404 for not found prediction, got %d", respNotFound.StatusCode)
	}
}

func TestAIHandler_GenerateStudyPlan(t *testing.T) {
	tokenMgr := jwt.NewTokenManager("test-secret-key-32chars-minimum-12345", 15*time.Minute, 24*time.Hour)
	userID := uuid.New()
	token, _, _ := tokenMgr.GenerateAccessToken(userID, "student@mindora.app", "user")

	mockSvc := &mockAIService{
		generateStudyPlanFn: func(ctx context.Context, uID uuid.UUID) (*services.AIResponse, error) {
			return &services.AIResponse{
				Reply: "## Jadwal Belajar Harian",
				Model: "openai/gpt-oss-20b",
			}, nil
		},
	}

	app := setupTestApp(mockSvc, tokenMgr)

	req := httptest.NewRequest("POST", "/api/v1/ai/study-plan", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
}
