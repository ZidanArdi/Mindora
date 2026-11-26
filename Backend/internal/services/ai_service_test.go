package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/groq"
)

// Mock Groq Client
type mockGroqClient struct {
	Response    *groq.ChatCompletionResponse
	Err         error
	CapturedReq groq.ChatCompletionRequest
	ModelName   string
}

func (m *mockGroqClient) CreateChatCompletion(ctx context.Context, req groq.ChatCompletionRequest) (*groq.ChatCompletionResponse, error) {
	m.CapturedReq = req
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Response, nil
}

func (m *mockGroqClient) GetModel() string {
	if m.ModelName != "" {
		return m.ModelName
	}
	return "openai/gpt-oss-20b"
}

// Mock Task Repo
type mockAITaskRepo struct {
	tasks []models.Task
}

func (m *mockAITaskRepo) Create(ctx context.Context, task *models.Task) error { return nil }
func (m *mockAITaskRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Task, error) {
	return nil, nil
}
func (m *mockAITaskRepo) FindAll(ctx context.Context, userID uuid.UUID, filter repositories.TaskFilter) ([]models.Task, int64, error) {
	var filtered []models.Task
	for _, t := range m.tasks {
		if t.UserID == userID {
			if filter.Status != "" && t.Status != filter.Status {
				continue
			}
			filtered = append(filtered, t)
		}
	}
	return filtered, int64(len(filtered)), nil
}
func (m *mockAITaskRepo) Update(ctx context.Context, task *models.Task) error { return nil }
func (m *mockAITaskRepo) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status string, completedAt *time.Time) (*models.Task, error) {
	return nil, nil
}
func (m *mockAITaskRepo) SoftDelete(ctx context.Context, id, userID uuid.UUID) error { return nil }
func (m *mockAITaskRepo) CountByStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	return map[string]int64{}, nil
}
func (m *mockAITaskRepo) CountCompletedByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAITaskRepo) FindCompletedTasksByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.Task, error) {
	return nil, nil
}

// Mock Study Repo
type mockAIStudyRepo struct {
	stats *repositories.StudyDailyStats
}

func (m *mockAIStudyRepo) CreateSession(ctx context.Context, session *models.StudySession) error {
	return nil
}
func (m *mockAIStudyRepo) FindSessionByID(ctx context.Context, id, userID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockAIStudyRepo) FindActiveOrPausedSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockAIStudyRepo) FindSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.StudySession, int64, error) {
	return nil, 0, nil
}
func (m *mockAIStudyRepo) FindCompletedSessions(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.StudySession, error) {
	return nil, nil
}
func (m *mockAIStudyRepo) UpdateSession(ctx context.Context, session *models.StudySession) error {
	return nil
}
func (m *mockAIStudyRepo) GetCompletedDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAIStudyRepo) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*repositories.StudyDailyStats, error) {
	if m.stats != nil {
		return m.stats, nil
	}
	return &repositories.StudyDailyStats{}, nil
}

// Mock Focus Repo
type mockAIFocusRepo struct {
	stats *repositories.FocusDailyStats
}

func (m *mockAIFocusRepo) CreateInterval(ctx context.Context, interval *models.FocusSession) error {
	return nil
}
func (m *mockAIFocusRepo) FindIntervalByID(ctx context.Context, id, userID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *mockAIFocusRepo) FindActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *mockAIFocusRepo) FindHistory(ctx context.Context, userID uuid.UUID, sessionType string, limit, offset int) ([]models.FocusSession, int64, error) {
	return nil, 0, nil
}
func (m *mockAIFocusRepo) FindCompletedIntervals(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.FocusSession, error) {
	return nil, nil
}
func (m *mockAIFocusRepo) UpdateInterval(ctx context.Context, interval *models.FocusSession) error {
	return nil
}
func (m *mockAIFocusRepo) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*repositories.FocusDailyStats, error) {
	if m.stats != nil {
		return m.stats, nil
	}
	return &repositories.FocusDailyStats{}, nil
}
func (m *mockAIFocusRepo) CountCompletedBreaks(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAIFocusRepo) GetCompletedFocusDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}

// Mock Lifestyle Repo
type mockAILifestyleRepo struct {
	record *models.LifestyleRecord
}

func (m *mockAILifestyleRepo) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error) {
	return m.record, nil
}
func (m *mockAILifestyleRepo) UpsertDailyRecord(ctx context.Context, record *models.LifestyleRecord) error {
	return nil
}
func (m *mockAILifestyleRepo) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.LifestyleRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockAILifestyleRepo) FindByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.LifestyleRecord, error) {
	return nil, nil
}

// Mock Prediction Repo
type mockAIPredictionRepo struct {
	latest *models.ProductivityPrediction
	byID   map[uuid.UUID]*models.ProductivityPrediction
}

func (m *mockAIPredictionRepo) Create(ctx context.Context, p *models.ProductivityPrediction) error {
	return nil
}
func (m *mockAIPredictionRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	if m.byID != nil {
		if p, ok := m.byID[id]; ok && p.UserID == userID {
			return p, nil
		}
	}
	return nil, repositories.ErrPredictionNotFound
}
func (m *mockAIPredictionRepo) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ProductivityPrediction, int64, error) {
	return nil, 0, nil
}
func (m *mockAIPredictionRepo) FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	if m.latest != nil && m.latest.UserID == userID {
		return m.latest, nil
	}
	return nil, nil
}
func (m *mockAIPredictionRepo) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.ProductivityPrediction, error) {
	return nil, nil
}

// Mock Recommendation Repo
type mockAIRecRepo struct{}

func (m *mockAIRecRepo) FindActive(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Recommendation, error) {
	return nil, nil
}
func (m *mockAIRecRepo) FindByID(ctx context.Context, id, userID uuid.UUID) (*models.Recommendation, error) {
	return nil, nil
}
func (m *mockAIRecRepo) UpsertPending(ctx context.Context, rec *models.Recommendation) error {
	return nil
}
func (m *mockAIRecRepo) UpdateStatus(ctx context.Context, id, userID uuid.UUID, status string, dismissedUntil *time.Time) (*models.Recommendation, error) {
	return nil, nil
}
func (m *mockAIRecRepo) FindDismissedRecent(ctx context.Context, userID uuid.UUID, ruleID string, now time.Time) (*models.Recommendation, error) {
	return nil, nil
}
func (m *mockAIRecRepo) FindByEvaluationKeyAndStatus(ctx context.Context, userID uuid.UUID, evalKey, status string) (*models.Recommendation, error) {
	return nil, nil
}
func (m *mockAIRecRepo) InvalidatePendingByRule(ctx context.Context, userID uuid.UUID, ruleID string) error {
	return nil
}
func (m *mockAIRecRepo) InvalidatePendingByID(ctx context.Context, id, userID uuid.UUID) error {
	return nil
}
func (m *mockAIRecRepo) ExpirePastValid(ctx context.Context, userID uuid.UUID, now time.Time) error {
	return nil
}

func TestAIService_Chat_Success(t *testing.T) {
	userID := uuid.New()
	mockClient := &mockGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Index: 0,
					Message: groq.ChatMessage{
						Role:    "assistant",
						Content: "Halo! Tetap semangat belajar hari ini. Jangan lupa istirahat.",
					},
					FinishReason: "stop",
				},
			},
			Usage: groq.UsageStats{TotalTokens: 120},
		},
	}

	taskRepo := &mockAITaskRepo{
		tasks: []models.Task{
			{
				ID:       uuid.New(),
				UserID:   userID,
				Title:    "Review Kalkulus",
				Priority: "high",
				Status:   models.TaskStatusPending,
			},
		},
	}
	studyRepo := &mockAIStudyRepo{
		stats: &repositories.StudyDailyStats{
			TotalCompletedSeconds:  3600,
			CompletedSessionsCount: 1,
		},
	}
	focusRepo := &mockAIFocusRepo{
		stats: &repositories.FocusDailyStats{
			CompletedFocusMinutes: 45,
			CompletedBreakMinutes: 10,
			TotalInterruptions:    1,
		},
	}
	sleepVal := 7.5
	phoneVal := 2.0
	stressVal := 2
	lifestyleRepo := &mockAILifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      &sleepVal,
			PhoneUsageHours: &phoneVal,
			StressLevel:     &stressVal,
		},
	}
	predRepo := &mockAIPredictionRepo{
		latest: &models.ProductivityPrediction{
			UserID:         userID,
			PredictedScore: 82.5,
			StateLabel:     "Personalized Estimate",
		},
	}
	recRepo := &mockAIRecRepo{}

	svc := services.NewAIService(mockClient, taskRepo, studyRepo, focusRepo, lifestyleRepo, predRepo, recRepo)

	resp, err := svc.Chat(context.Background(), userID, "Bagaimana cara meningkatkan konsentrasi?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Reply != "Halo! Tetap semangat belajar hari ini. Jangan lupa istirahat." {
		t.Errorf("unexpected reply: %s", resp.Reply)
	}
	if resp.TokensUsed != 120 {
		t.Errorf("expected 120 tokens used, got %d", resp.TokensUsed)
	}

	if len(mockClient.CapturedReq.Messages) != 2 {
		t.Fatalf("expected 2 messages (system and user), got %d", len(mockClient.CapturedReq.Messages))
	}
	if mockClient.CapturedReq.Messages[0].Role != "system" {
		t.Errorf("expected system role, got %s", mockClient.CapturedReq.Messages[0].Role)
	}
	if !strings.Contains(mockClient.CapturedReq.Messages[1].Content, "Bagaimana cara meningkatkan konsentrasi?") {
		t.Errorf("expected user prompt in message content")
	}
	if !strings.Contains(mockClient.CapturedReq.Messages[1].Content, "82.5/100") {
		t.Errorf("expected prediction context in user message")
	}
}

func TestAIService_Chat_Validation(t *testing.T) {
	svc := services.NewAIService(&mockGroqClient{}, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})

	// 1. Empty prompt
	_, err := svc.Chat(context.Background(), uuid.New(), "   ")
	if !errors.Is(err, services.ErrAIPromptRequired) {
		t.Errorf("expected ErrAIPromptRequired, got: %v", err)
	}

	// 2. Too long prompt
	longPrompt := strings.Repeat("a", 1001)
	_, err = svc.Chat(context.Background(), uuid.New(), longPrompt)
	if !errors.Is(err, services.ErrAIPromptTooLong) {
		t.Errorf("expected ErrAIPromptTooLong, got: %v", err)
	}

	// 3. Nil client
	svcNoClient := services.NewAIService(nil, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})
	_, err = svcNoClient.Chat(context.Background(), uuid.New(), "Halo")
	if !errors.Is(err, services.ErrAIClientNotConfigured) {
		t.Errorf("expected ErrAIClientNotConfigured, got: %v", err)
	}
}

func TestAIService_ExplainPrediction_Success(t *testing.T) {
	userID := uuid.New()
	predID := uuid.New()
	sleepHours := 7.0
	phoneHours := 2.5
	stress := 2

	pred := &models.ProductivityPrediction{
		ID:              predID,
		UserID:          userID,
		PredictedScore:  78.4,
		ModelVersion:    "ridge_v1.0.0",
		StudyHours:      3.5,
		SleepHours:      &sleepHours,
		PhoneUsageHours: &phoneHours,
		StressLevel:     &stress,
		StateLabel:      "Personalized Estimate",
	}

	mockClient := &mockGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{
						Content: "Skor Anda 78.4 menunjukkan produktivitas baik...",
					},
				},
			},
			Usage: groq.UsageStats{TotalTokens: 250},
		},
	}

	predRepo := &mockAIPredictionRepo{
		latest: pred,
		byID:   map[uuid.UUID]*models.ProductivityPrediction{predID: pred},
	}

	svc := services.NewAIService(mockClient, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, predRepo, &mockAIRecRepo{})

	// Explain latest
	resp, err := svc.ExplainPrediction(context.Background(), userID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Reply, "78.4") {
		t.Errorf("unexpected reply: %s", resp.Reply)
	}

	// Explain by ID
	resp2, err := svc.ExplainPrediction(context.Background(), userID, &predID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp2.TokensUsed != 250 {
		t.Errorf("expected 250 tokens, got %d", resp2.TokensUsed)
	}
}

func TestAIService_ExplainPrediction_UserIsolation(t *testing.T) {
	userA := uuid.New()
	userB := uuid.New()
	predID := uuid.New()

	predUserB := &models.ProductivityPrediction{
		ID:             predID,
		UserID:         userB,
		PredictedScore: 90.0,
		StateLabel:     "Personalized Estimate",
	}

	predRepo := &mockAIPredictionRepo{
		byID: map[uuid.UUID]*models.ProductivityPrediction{predID: predUserB},
	}

	svc := services.NewAIService(&mockGroqClient{}, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, predRepo, &mockAIRecRepo{})

	// User A tries to access User B's prediction ID
	_, err := svc.ExplainPrediction(context.Background(), userA, &predID)
	if !errors.Is(err, services.ErrAIPredictionNotFound) {
		t.Fatalf("expected ErrAIPredictionNotFound when User A accesses User B's prediction, got: %v", err)
	}
}

func TestAIService_ExplainPrediction_BaselineNotice(t *testing.T) {
	userID := uuid.New()
	baselinePred := &models.ProductivityPrediction{
		ID:             uuid.New(),
		UserID:         userID,
		PredictedScore: 50.0,
		IsBaseline:     true,
		StateLabel:     "Estimasi Awal",
	}

	mockClient := &mockGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{
						Content: "Ini adalah estimasi awal berdasarkan baseline...",
					},
				},
			},
		},
	}

	predRepo := &mockAIPredictionRepo{latest: baselinePred}
	svc := services.NewAIService(mockClient, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, predRepo, &mockAIRecRepo{})

	_, err := svc.ExplainPrediction(context.Background(), userID, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify baseline note in prompt
	userContent := mockClient.CapturedReq.Messages[1].Content
	if !strings.Contains(userContent, "CATATAN KHUSUS: Skor ini adalah estimasi awal berdasarkan baseline") {
		t.Errorf("expected baseline explanation in prompt")
	}
}

func TestAIService_ExplainPrediction_NotFound(t *testing.T) {
	userID := uuid.New()
	svc := services.NewAIService(&mockGroqClient{}, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})

	_, err := svc.ExplainPrediction(context.Background(), userID, nil)
	if !errors.Is(err, services.ErrAIPredictionNotFound) {
		t.Errorf("expected ErrAIPredictionNotFound, got: %v", err)
	}
}

func TestAIService_GenerateStudyPlan_Success(t *testing.T) {
	userID := uuid.New()
	mockClient := &mockGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{
						Content: "## Jadwal Belajar Hari Ini\n1. 09:00 - 09:45: Belajar Fisika Dasar\n2. 09:45 - 09:55: Istirahat",
					},
				},
			},
			Usage: groq.UsageStats{TotalTokens: 180},
		},
	}

	taskRepo := &mockAITaskRepo{
		tasks: []models.Task{
			{
				ID:       uuid.New(),
				UserID:   userID,
				Title:    "Tugas Fisika",
				Priority: "high",
				Status:   models.TaskStatusPending,
			},
		},
	}

	svc := services.NewAIService(mockClient, taskRepo, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})

	resp, err := svc.GenerateStudyPlan(context.Background(), userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(resp.Reply, "Jadwal Belajar") {
		t.Errorf("unexpected reply: %s", resp.Reply)
	}
}

func TestAIService_GroqClientError(t *testing.T) {
	userID := uuid.New()
	mockClient := &mockGroqClient{
		Err: groq.ErrRateLimitExceeded,
	}

	svc := services.NewAIService(mockClient, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})

	_, err := svc.Chat(context.Background(), userID, "Pertanyaan tes")
	if !errors.Is(err, groq.ErrRateLimitExceeded) {
		t.Errorf("expected ErrRateLimitExceeded, got: %v", err)
	}
}

func TestAIService_EmptyAssistantResponseFallback(t *testing.T) {
	userID := uuid.New()
	mockClient := &mockGroqClient{
		Response: &groq.ChatCompletionResponse{
			Model: "openai/gpt-oss-20b",
			Choices: []groq.ChatChoice{
				{
					Message: groq.ChatMessage{
						Content: "   ",
					},
				},
			},
		},
	}

	svc := services.NewAIService(mockClient, &mockAITaskRepo{}, &mockAIStudyRepo{}, &mockAIFocusRepo{}, &mockAILifestyleRepo{}, &mockAIPredictionRepo{}, &mockAIRecRepo{})

	resp, err := svc.Chat(context.Background(), userID, "Pertanyaan tes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Reply, "belum dapat menghasilkan respons") {
		t.Errorf("expected fallback message for empty assistant response, got: %s", resp.Reply)
	}
}
