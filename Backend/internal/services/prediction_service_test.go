package services_test

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	durationSec    int64
	hasRecords     bool
	returnNilStats bool
	err            error
}

func (m *mockStudyRepo) CreateSession(ctx context.Context, s *models.StudySession) error { return nil }
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
	if m.err != nil {
		return nil, m.err
	}
	if m.returnNilStats {
		return nil, nil
	}
	count := int64(0)
	if m.hasRecords || m.durationSec > 0 {
		count = 1
	}
	return &repositories.StudyDailyStats{TotalCompletedSeconds: m.durationSec, CompletedSessionsCount: count}, nil
}

// MockLifestyleRepo
type mockLifestyleRepo struct {
	record *models.LifestyleRecord
	err    error
}

func (m *mockLifestyleRepo) FindByDate(ctx context.Context, uID uuid.UUID, d time.Time) (*models.LifestyleRecord, error) {
	if m.err != nil {
		return nil, m.err
	}
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
	if !errors.Is(err, services.ErrInvalidPredictionInput) {
		t.Errorf("expected ErrInvalidPredictionInput for study hours > 24, got %v", err)
	}

	// 2. Invalid stress level (> 10)
	invalidStress := 11
	_, err = svc.Predict(context.Background(), userID, services.PredictionRequest{StressLevel: &invalidStress})
	if !errors.Is(err, services.ErrInvalidPredictionInput) {
		t.Errorf("expected ErrInvalidPredictionInput for stress level > 10, got %v", err)
	}

	// 3. Invalid stress level (< 1)
	zeroStress := 0
	_, err = svc.Predict(context.Background(), userID, services.PredictionRequest{StressLevel: &zeroStress})
	if !errors.Is(err, services.ErrInvalidPredictionInput) {
		t.Errorf("expected ErrInvalidPredictionInput for stress level < 1, got %v", err)
	}
}

func TestPredictionService_ImputationWhenLifestyleMissing(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 7200, hasRecords: true} // 2 hours
	lifestyleRepo := &mockLifestyleRepo{record: nil}                 // No lifestyle logged
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
	if resp.IsBaseline != true {
		t.Errorf("expected IsBaseline to be true when lifestyle features are missing")
	}
	if resp.PredictionState != services.StateBaseline {
		t.Errorf("expected state %s, got %s", services.StateBaseline, resp.PredictionState)
	}

	// Score must be valid within [0, 100]
	if resp.PredictedScore < 0 || resp.PredictedScore > 100 {
		t.Errorf("predicted score out of range: %v", resp.PredictedScore)
	}
}

// Scenario 1: All four features available -> Personalized Estimate
func TestPredictionService_PersonalizedEstimate_AllFeaturesAvailable(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 14400, hasRecords: true} // 4.0h
	sleep := 7.5
	phone := 2.0
	stress := 4
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      &sleep,
			PhoneUsageHours: &phone,
			StressLevel:     &stress,
		},
	}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.IsBaseline {
		t.Errorf("expected Personalized Estimate (is_baseline=false), got true")
	}
	if resp.PredictionState != services.StatePersonalized {
		t.Errorf("expected state %s, got %s", services.StatePersonalized, resp.PredictionState)
	}
	if resp.StateLabel != services.LabelPersonalized {
		t.Errorf("expected label %s, got %s", services.LabelPersonalized, resp.StateLabel)
	}
	if len(resp.UnrecordedFeatures) != 0 {
		t.Errorf("expected 0 unrecorded features, got %v", resp.UnrecordedFeatures)
	}
	if resp.StudyHours != 4.0 {
		t.Errorf("expected 4.0 study hours, got %v", resp.StudyHours)
	}
	if resp.SleepHours == nil || *resp.SleepHours != 7.5 {
		t.Errorf("expected 7.5 sleep hours, got %v", resp.SleepHours)
	}
}

// Scenario 2: One lifestyle feature missing -> Baseline Estimate
func TestPredictionService_BaselineEstimate_OneLifestyleMissing(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 10800, hasRecords: true} // 3.0h
	phone := 3.0
	stress := 5
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      nil, // MISSING SLEEP
			PhoneUsageHours: &phone,
			StressLevel:     &stress,
		},
	}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsBaseline {
		t.Errorf("expected Baseline Estimate (is_baseline=true)")
	}
	if resp.PredictionState != services.StateBaseline {
		t.Errorf("expected state %s, got %s", services.StateBaseline, resp.PredictionState)
	}
	if len(resp.UnrecordedFeatures) != 1 || resp.UnrecordedFeatures[0] != "sleep_hours" {
		t.Errorf("expected unrecorded features [sleep_hours], got %v", resp.UnrecordedFeatures)
	}
	if resp.SleepHours != nil {
		t.Errorf("expected observed sleep_hours to be nil (preserved NULL), got %v", resp.SleepHours)
	}
}

// Scenario 4: Study duration is genuinely zero (user completed a session of 0 seconds)
func TestPredictionService_ObservedZeroStudyDuration(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 0, hasRecords: true} // Completed session, 0 seconds
	sleep := 8.0
	phone := 1.5
	stress := 3
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      &sleep,
			PhoneUsageHours: &phone,
			StressLevel:     &stress,
		},
	}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StudyHours != 0.0 {
		t.Errorf("expected observed 0.0 study hours preserved, got %v", resp.StudyHours)
	}
	if resp.FeaturesSnapshot["study_hours_imputed"] != false {
		t.Errorf("expected study_hours_imputed to be false for recorded 0 session")
	}
	// All other features observed -> should be personalized
	if resp.IsBaseline {
		t.Errorf("expected Personalized Estimate when 0h study was genuinely observed")
	}
}

// Scenario 5: No study records exist -> Impute with StudyHoursMean (5.25h) and mark imputed
func TestPredictionService_UnrecordedStudyImputation(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 0, hasRecords: false} // NO study sessions
	sleep := 8.0
	phone := 1.5
	stress := 3
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      &sleep,
			PhoneUsageHours: &phone,
			StressLevel:     &stress,
		},
	}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Study hours should be imputed with reference default (5.25h)
	expectedStudy := 5.25
	if resp.StudyHours != expectedStudy {
		t.Errorf("expected imputed study hours %v, got %v", expectedStudy, resp.StudyHours)
	}
	if resp.FeaturesSnapshot["study_hours_imputed"] != true {
		t.Errorf("expected study_hours_imputed to be true for unrecorded study")
	}
	if !resp.IsBaseline {
		t.Errorf("expected Baseline Estimate due to missing study record")
	}
	foundStudy := false
	for _, f := range resp.UnrecordedFeatures {
		if f == "study_hours" {
			foundStudy = true
		}
	}
	if !foundStudy {
		t.Errorf("expected study_hours in unrecorded_features, got %v", resp.UnrecordedFeatures)
	}
}

// Scenario 6: Cold start: All four required features missing
func TestPredictionService_ColdStart_AllFeaturesMissing(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 0, hasRecords: false} // No study
	lifestyleRepo := &mockLifestyleRepo{record: nil}               // No lifestyle
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !resp.IsBaseline {
		t.Errorf("cold start must be Baseline Estimate")
	}
	if len(resp.UnrecordedFeatures) != 4 {
		t.Errorf("expected all 4 features in unrecorded_features, got %v", resp.UnrecordedFeatures)
	}
	// Score must evaluate to population reference default ~50.76 (with integer stress level = 5)
	if math.Abs(resp.PredictedScore-50.76) > 0.05 {
		t.Errorf("expected cold-start baseline score ~50.76, got %v", resp.PredictedScore)
	}
}

// Scenario 7: Imputation metadata matches model inputs exactly
func TestPredictionService_ImputationMetadataMatchesModelInputs(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{durationSec: 0, hasRecords: false} // Imputed
	phone := 3.5
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			SleepHours:      nil,    // Imputed
			PhoneUsageHours: &phone, // Observed
			StressLevel:     nil,    // Imputed
		},
	}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	snap := resp.FeaturesSnapshot
	if snap["study_hours_supplied"].(float64) != 5.25 || snap["study_hours_observed"] != nil {
		t.Errorf("study metadata mismatch: supplied=%v, observed=%v", snap["study_hours_supplied"], snap["study_hours_observed"])
	}
	if snap["phone_usage_hours_supplied"].(float64) != 3.5 || snap["phone_usage_hours_observed"] == nil {
		t.Errorf("phone metadata mismatch: supplied=%v, observed=%v", snap["phone_usage_hours_supplied"], snap["phone_usage_hours_observed"])
	}
	if snap["sleep_hours_supplied"].(float64) != 6.54 || snap["sleep_hours_observed"] != nil {
		t.Errorf("sleep metadata mismatch: supplied=%v, observed=%v", snap["sleep_hours_supplied"], snap["sleep_hours_observed"])
	}
	if snap["stress_level_supplied"].(int) != 5 || snap["stress_level_observed"] != nil {
		t.Errorf("stress metadata mismatch: supplied=%v, observed=%v", snap["stress_level_supplied"], snap["stress_level_observed"])
	}
}

// Test Study Repository Error Handling (Phase 3.2.1)
func TestPredictionService_StudyRepoErrors(t *testing.T) {
	userID := uuid.New()

	t.Run("Database connection error aborts prediction and does not persist record", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{err: errors.New("db connection refused")}
		lifestyleRepo := &mockLifestyleRepo{}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on study repo failure, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response on failure, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "failed to retrieve daily study statistics") {
			t.Errorf("expected wrapped error message, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted on DB error, got %d", len(predRepo.records))
		}
	})

	t.Run("Context cancellation error propagates properly", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{err: context.Canceled}
		lifestyleRepo := &mockLifestyleRepo{}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on context cancellation, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error to unwrap to context.Canceled, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted, got %d", len(predRepo.records))
		}
	})

	t.Run("Unexpected query error aborts prediction", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{err: errors.New("query failed: deadlock detected")}
		lifestyleRepo := &mockLifestyleRepo{}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on deadlock failure, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "deadlock detected") {
			t.Errorf("expected deadlock error detail, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted, got %d", len(predRepo.records))
		}
	})

	t.Run("Explicit study hours request skips study repo and succeeds even if study repo has error", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{err: errors.New("db connection refused")}
		lifestyleRepo := &mockLifestyleRepo{}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		studyVal := 4.5
		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{
			StudyHours: &studyVal,
		})
		if err != nil {
			t.Fatalf("unexpected error when study hours explicitly provided: %v", err)
		}
		if resp.StudyHours != 4.5 {
			t.Errorf("expected study hours 4.5, got %v", resp.StudyHours)
		}
		if len(predRepo.records) != 1 {
			t.Errorf("expected 1 prediction persisted, got %d", len(predRepo.records))
		}
	})
}

// Test Lifestyle Repository Error Handling (Phase 3.2.1)
func TestPredictionService_LifestyleRepoErrors(t *testing.T) {
	userID := uuid.New()

	t.Run("Database connection error aborts prediction and does not persist record", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{err: errors.New("pq: could not connect to server")}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on lifestyle repo connection failure, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response on failure, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "failed to retrieve daily lifestyle record") {
			t.Errorf("expected wrapped error message, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted on DB error, got %d", len(predRepo.records))
		}
	})

	t.Run("Context cancellation error propagates properly", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{err: context.Canceled}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on context cancellation, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
		if !errors.Is(err, context.Canceled) {
			t.Errorf("expected error to unwrap to context.Canceled, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted, got %d", len(predRepo.records))
		}
	})

	t.Run("Unexpected query error aborts prediction", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{err: errors.New("internal query error: relation does not exist")}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err == nil {
			t.Fatal("expected error on query failure, got nil")
		}
		if resp != nil {
			t.Errorf("expected nil response, got %+v", resp)
		}
		if !strings.Contains(err.Error(), "relation does not exist") {
			t.Errorf("expected error details, got: %v", err)
		}
		if len(predRepo.records) != 0 {
			t.Errorf("expected 0 predictions persisted, got %d", len(predRepo.records))
		}
	})
}

// Test Lifestyle Repository Not Found Conventions (Phase 3.2.1)
func TestPredictionService_LifestyleRepo_NotFoundConventions(t *testing.T) {
	userID := uuid.New()

	t.Run("repositories.ErrLifestyleRecordNotFound proceeds to baseline imputation", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{err: repositories.ErrLifestyleRecordNotFound}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err != nil {
			t.Fatalf("unexpected error when ErrLifestyleRecordNotFound returned: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if !resp.IsBaseline {
			t.Errorf("expected IsBaseline to be true")
		}
		if resp.PredictionState != services.StateBaseline {
			t.Errorf("expected state %s, got %s", services.StateBaseline, resp.PredictionState)
		}
		if len(predRepo.records) != 1 {
			t.Errorf("expected 1 prediction persisted, got %d", len(predRepo.records))
		}
	})

	t.Run("gorm.ErrRecordNotFound proceeds to baseline imputation", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{err: gorm.ErrRecordNotFound}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err != nil {
			t.Fatalf("unexpected error when gorm.ErrRecordNotFound returned: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if !resp.IsBaseline {
			t.Errorf("expected IsBaseline to be true")
		}
		if resp.PredictionState != services.StateBaseline {
			t.Errorf("expected state %s, got %s", services.StateBaseline, resp.PredictionState)
		}
		if len(predRepo.records) != 1 {
			t.Errorf("expected 1 prediction persisted, got %d", len(predRepo.records))
		}
	})

	t.Run("nil record and nil error proceeds to baseline imputation", func(t *testing.T) {
		predRepo := &mockPredictionRepo{}
		studyRepo := &mockStudyRepo{durationSec: 3600, hasRecords: true}
		lifestyleRepo := &mockLifestyleRepo{record: nil, err: nil}
		svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)

		resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
		if err != nil {
			t.Fatalf("unexpected error when (nil, nil) returned: %v", err)
		}
		if resp == nil {
			t.Fatal("expected non-nil response")
		}
		if !resp.IsBaseline {
			t.Errorf("expected IsBaseline to be true")
		}
		if resp.PredictionState != services.StateBaseline {
			t.Errorf("expected state %s, got %s", services.StateBaseline, resp.PredictionState)
		}
		if len(predRepo.records) != 1 {
			t.Errorf("expected 1 prediction persisted, got %d", len(predRepo.records))
		}
	})
}

// Test Unexpected Nil Study Statistics (Phase 3.2.2)
func TestPredictionService_UnexpectedNilStudyStats(t *testing.T) {
	predRepo := &mockPredictionRepo{}
	studyRepo := &mockStudyRepo{returnNilStats: true}
	lifestyleRepo := &mockLifestyleRepo{}
	svc := services.NewPredictionService(predRepo, studyRepo, lifestyleRepo)
	userID := uuid.New()

	resp, err := svc.Predict(context.Background(), userID, services.PredictionRequest{})
	if err == nil {
		t.Fatal("expected error on unexpected nil study stats, got nil")
	}
	if resp != nil {
		t.Errorf("expected nil response on error, got %+v", resp)
	}
	if !strings.Contains(err.Error(), "unexpected nil statistics") {
		t.Errorf("expected error message mentioning unexpected nil statistics, got: %v", err)
	}
	if len(predRepo.records) != 0 {
		t.Errorf("expected 0 predictions persisted on nil stats error, got %d", len(predRepo.records))
	}
}
