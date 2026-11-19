package services_test

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
)

// MockPredictionRepo
type mockPredictionRepo struct {
	records []models.ProductivityPrediction
}

func (m *mockPredictionRepo) Create(ctx context.Context, p *models.ProductivityPrediction) error {
	m.records = append(m.records, *p)
	return nil
}

func (m *mockPredictionRepo) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	for _, r := range m.records {
		if r.ID == id && r.UserID == userID {
			return &r, nil
		}
	}
	return nil, repositories.ErrPredictionNotFound
}

func (m *mockPredictionRepo) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ProductivityPrediction, int64, error) {
	var userRecords []models.ProductivityPrediction
	for _, r := range m.records {
		if r.UserID == userID {
			userRecords = append(userRecords, r)
		}
	}
	return userRecords, int64(len(userRecords)), nil
}

func (m *mockPredictionRepo) FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	for i := len(m.records) - 1; i >= 0; i-- {
		if m.records[i].UserID == userID {
			return &m.records[i], nil
		}
	}
	return nil, nil
}

func (m *mockPredictionRepo) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.ProductivityPrediction, error) {
	dateStr := date.Format("2006-01-02")
	for _, r := range m.records {
		if r.UserID == userID && r.PredictionDate.Format("2006-01-02") == dateStr {
			return &r, nil
		}
	}
	return nil, nil
}

// MockStudyRepo
type mockStudyRepo struct {
	durationSec int64
}

func (m *mockStudyRepo) CreateSession(ctx context.Context, s *models.StudySession) error      { return nil }
func (m *mockStudyRepo) FindSessionByID(ctx context.Context, id, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockStudyRepo) FindActiveOrPausedSession(ctx context.Context, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockStudyRepo) FindSessions(ctx context.Context, uID uuid.UUID, s, e *time.Time, l, o int) ([]models.StudySession, int64, error) {
	return nil, 0, nil
}
func (m *mockStudyRepo) FindCompletedSessions(ctx context.Context, uID uuid.UUID, s, e time.Time) ([]models.StudySession, error) {
	return nil, nil
}
func (m *mockStudyRepo) UpdateSession(ctx context.Context, s *models.StudySession) error { return nil }
func (m *mockStudyRepo) GetCompletedDuration(ctx context.Context, uID uuid.UUID, s, e time.Time) (int64, error) {
	return m.durationSec, nil
}
func (m *mockStudyRepo) GetDailyStatistics(ctx context.Context, uID uuid.UUID, s, e time.Time) (*repositories.StudyDailyStats, error) {
	return &repositories.StudyDailyStats{TotalCompletedSeconds: m.durationSec, CompletedSessionsCount: 1}, nil
}

// MockLifestyleRepo
type mockLifestyleRepo struct {
	record *models.LifestyleRecord
}

func (m *mockLifestyleRepo) FindByDate(ctx context.Context, uID uuid.UUID, d time.Time) (*models.LifestyleRecord, error) {
	return m.record, nil
}
func (m *mockLifestyleRepo) UpsertDailyRecord(ctx context.Context, r *models.LifestyleRecord) error {
	return nil
}
func (m *mockLifestyleRepo) FindHistory(ctx context.Context, uID uuid.UUID, l, o int) ([]models.LifestyleRecord, int64, error) {
	return nil, 0, nil
}
func (m *mockLifestyleRepo) FindByDateRange(ctx context.Context, uID uuid.UUID, s, e time.Time) ([]models.LifestyleRecord, error) {
	return nil, nil
}

func TestPredictionService_ExactScoringAgainstPythonArtifact(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{}
	lifestyleRepo := &mockLifestyleRepo{}

	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

	userID := uuid.New()
	study := 5.0
	sleep := 8.0
	phone := 1.5
	stress := 3

	req := services.PredictionRequest{
		StudyHours:      &study,
		SleepHours:      &sleep,
		PhoneUsageHours: &phone,
		StressLevel:     &stress,
	}

	resp, err := svc.Predict(context.Background(), userID, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Python CLI output for study=5, sleep=8, phone=1.5, stress=3 was exactly 63.49
	expectedScore := 63.49
	if math.Abs(resp.PredictedScore-expectedScore) > 0.05 {
		t.Errorf("expected score ~%.2f, got %.2f", expectedScore, resp.PredictedScore)
	}

	if resp.ModelVersion != services.ModelVersionCurrent {
		t.Errorf("expected model version %s, got %s", services.ModelVersionCurrent, resp.ModelVersion)
	}
}

func TestPredictionService_InputValidation(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{}
	lifestyleRepo := &mockLifestyleRepo{}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	// 1. Invalid study hours (> 24)
	invalidStudy := 25.0
	_, err := svc.Predict(context.Background(), userID, services.PredictionRequest{StudyHours: &invalidStudy})
	if err != services.ErrInvalidPredictionInput {
		t.Errorf("expected ErrInvalidPredictionInput for study hours > 24, got %v", err)
	}

	// 2. Invalid stress level (> 10)
	invalidStress := 11
	_, err = svc.Predict(context.Background(), userID, services.PredictionRequest{StressLevel: &invalidStress})
	if err != services.ErrInvalidPredictionInput {
		t.Errorf("expected ErrInvalidPredictionInput for stress level > 10, got %v", err)
	}

	// 3. Invalid stress level (< 1)
	zeroStress := 0
	_, err = svc.Predict(context.Background(), userID, services.PredictionRequest{StressLevel: &zeroStress})
	if err != services.ErrInvalidPredictionInput {
		t.Errorf("expected ErrInvalidPredictionInput for stress level < 1, got %v", err)
	}
}

func TestPredictionService_ImputationWhenLifestyleMissing(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 7200} // 2 hours
	lifestyleRepo := &mockLifestyleRepo{record: nil} // No lifestyle logged
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Study hours aggregated from 7200 seconds = 2.0 hours
	if resp.StudyHours != 2.0 {
		t.Errorf("expected study hours 2.0, got %v", resp.StudyHours)
	}

	// Sleep, phone, and stress should be imputed
	snapshot := resp.FeaturesSnapshot
	if snapshot["sleep_hours_imputed"] != true {
		t.Errorf("expected sleep_hours_imputed to be true")
	}
	if snapshot["phone_usage_hours_imputed"] != true {
		t.Errorf("expected phone_usage_hours_imputed to be true")
	}
	if snapshot["stress_level_imputed"] != true {
		t.Errorf("expected stress_level_imputed to be true")
	}

	// Score must be valid within [0, 100]
	if resp.PredictedScore < 0 || resp.PredictedScore > 100 {
		t.Errorf("predicted score out of range: %v", resp.PredictedScore)
	}
}
