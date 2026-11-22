package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
)

// MockAnalyticsStudyRepo
type mockAnalyticsStudyRepo struct {
	sessions []models.StudySession
}

func (m *mockAnalyticsStudyRepo) CreateSession(ctx context.Context, s *models.StudySession) error {
	return nil
}
func (m *mockAnalyticsStudyRepo) FindSessionByID(ctx context.Context, id, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockAnalyticsStudyRepo) FindActiveOrPausedSession(ctx context.Context, uID uuid.UUID) (*models.StudySession, error) {
	return nil, nil
}
func (m *mockAnalyticsStudyRepo) FindSessions(ctx context.Context, uID uuid.UUID, s, e *time.Time, l, o int) ([]models.StudySession, int64, error) {
	return nil, 0, nil
}
func (m *mockAnalyticsStudyRepo) FindCompletedSessions(ctx context.Context, uID uuid.UUID, start, end time.Time) ([]models.StudySession, error) {
	var res []models.StudySession
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			res = append(res, s)
		}
	}
	return res, nil
}
func (m *mockAnalyticsStudyRepo) UpdateSession(ctx context.Context, s *models.StudySession) error {
	return nil
}
func (m *mockAnalyticsStudyRepo) GetCompletedDuration(ctx context.Context, uID uuid.UUID, start, end time.Time) (int64, error) {
	var total int64
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			total += int64(s.DurationSeconds)
		}
	}
	return total, nil
}
func (m *mockAnalyticsStudyRepo) GetDailyStatistics(ctx context.Context, uID uuid.UUID, start, end time.Time) (*repositories.StudyDailyStats, error) {
	var total int64
	var count int64
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			total += int64(s.DurationSeconds)
			count++
		}
	}
	return &repositories.StudyDailyStats{TotalCompletedSeconds: total, CompletedSessionsCount: count}, nil
}

// MockAnalyticsFocusRepo
type mockAnalyticsFocusRepo struct {
	intervals []models.FocusSession
}

func (m *mockAnalyticsFocusRepo) CreateInterval(ctx context.Context, i *models.FocusSession) error {
	return nil
}
func (m *mockAnalyticsFocusRepo) FindIntervalByID(ctx context.Context, id, uID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *mockAnalyticsFocusRepo) FindActiveInterval(ctx context.Context, uID uuid.UUID) (*models.FocusSession, error) {
	return nil, nil
}
func (m *mockAnalyticsFocusRepo) FindHistory(ctx context.Context, uID uuid.UUID, sType string, l, o int) ([]models.FocusSession, int64, error) {
	return nil, 0, nil
}
func (m *mockAnalyticsFocusRepo) FindCompletedIntervals(ctx context.Context, uID uuid.UUID, start, end time.Time) ([]models.FocusSession, error) {
	var res []models.FocusSession
	for _, i := range m.intervals {
		if i.UserID == uID && i.Status == models.FocusStatusCompleted && !i.StartedAt.Before(start) && !i.StartedAt.After(end) {
			res = append(res, i)
		}
	}
	return res, nil
}
func (m *mockAnalyticsFocusRepo) UpdateInterval(ctx context.Context, i *models.FocusSession) error {
	return nil
}
func (m *mockAnalyticsFocusRepo) GetDailyStatistics(ctx context.Context, uID uuid.UUID, start, end time.Time) (*repositories.FocusDailyStats, error) {
	stats := &repositories.FocusDailyStats{}
	for _, i := range m.intervals {
		if i.UserID == uID && i.Status == models.FocusStatusCompleted && !i.StartedAt.Before(start) && !i.StartedAt.After(end) {
			if i.SessionType == models.FocusTypeWork {
				stats.CompletedFocusMinutes += int64(i.DurationMinutes)
				stats.CompletedFocusCount++
			} else if i.SessionType == models.FocusTypeBreak {
				stats.CompletedBreakMinutes += int64(i.DurationMinutes)
				stats.CompletedBreaksCount++
			}
			stats.TotalInterruptions += int64(i.InterruptionCount)
		}
	}
	return stats, nil
}
func (m *mockAnalyticsFocusRepo) CountCompletedBreaks(ctx context.Context, uID uuid.UUID, s, e time.Time) (int64, error) {
	return 0, nil
}
func (m *mockAnalyticsFocusRepo) GetCompletedFocusDuration(ctx context.Context, uID uuid.UUID, s, e time.Time) (int64, error) {
	return 0, nil
}

// MockAnalyticsTaskRepo
type mockAnalyticsTaskRepo struct {
	tasks []models.Task
}

func (m *mockAnalyticsTaskRepo) Create(ctx context.Context, t *models.Task) error { return nil }
func (m *mockAnalyticsTaskRepo) FindByID(ctx context.Context, id, uID uuid.UUID) (*models.Task, error) {
	return nil, nil
}
func (m *mockAnalyticsTaskRepo) FindAll(ctx context.Context, uID uuid.UUID, f repositories.TaskFilter) ([]models.Task, int64, error) {
	return nil, 0, nil
}
func (m *mockAnalyticsTaskRepo) Update(ctx context.Context, t *models.Task) error { return nil }
func (m *mockAnalyticsTaskRepo) UpdateStatus(ctx context.Context, id, uID uuid.UUID, s string, cAt *time.Time) (*models.Task, error) {
	return nil, nil
}
func (m *mockAnalyticsTaskRepo) SoftDelete(ctx context.Context, id, uID uuid.UUID) error { return nil }
func (m *mockAnalyticsTaskRepo) CountByStatus(ctx context.Context, uID uuid.UUID) (map[string]int64, error) {
	return nil, nil
}
func (m *mockAnalyticsTaskRepo) CountCompletedByDateRange(ctx context.Context, uID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	for _, t := range m.tasks {
		if t.UserID == uID && t.Status == models.TaskStatusCompleted && t.CompletedAt != nil {
			if !t.CompletedAt.Before(start) && !t.CompletedAt.After(end) {
				count++
			}
		}
	}
	return count, nil
}
func (m *mockAnalyticsTaskRepo) FindCompletedTasksByDateRange(ctx context.Context, uID uuid.UUID, start, end time.Time) ([]models.Task, error) {
	var res []models.Task
	for _, t := range m.tasks {
		if t.UserID == uID && t.Status == models.TaskStatusCompleted && t.CompletedAt != nil {
			if !t.CompletedAt.Before(start) && !t.CompletedAt.After(end) {
				res = append(res, t)
			}
		}
	}
	return res, nil
}

func TestAnalyticsService_DailyAndWeeklyCalculations(t *testing.T) {
	userID := uuid.New()
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, time.UTC)

	studyRepo := &mockAnalyticsStudyRepo{
		sessions: []models.StudySession{
			{
				ID:              uuid.New(),
				UserID:          userID,
				DurationSeconds: 3600, // 1 hour completed
				Status:          models.StudyStatusCompleted,
				StartedAt:       startOfToday,
			},
			{
				ID:              uuid.New(),
				UserID:          userID,
				DurationSeconds: 1800,
				Status:          models.StudyStatusCancelled, // Must be ignored!
				StartedAt:       startOfToday.Add(2 * time.Hour),
			},
		},
	}

	focusRepo := &mockAnalyticsFocusRepo{
		intervals: []models.FocusSession{
			{
				ID:              uuid.New(),
				UserID:          userID,
				SessionType:     models.FocusTypeWork,
				DurationMinutes: 25,
				Status:          models.FocusStatusCompleted,
				StartedAt:       startOfToday.Add(1 * time.Hour),
			},
			{
				ID:              uuid.New(),
				UserID:          userID,
				SessionType:     models.FocusTypeBreak,
				DurationMinutes: 5,
				Status:          models.FocusStatusCompleted,
				StartedAt:       startOfToday.Add(1*time.Hour + 30*time.Minute),
			},
			{
				ID:              uuid.New(),
				UserID:          userID,
				SessionType:     models.FocusTypeWork,
				DurationMinutes: 25,
				Status:          models.FocusStatusAbandoned, // Must be ignored!
				StartedAt:       startOfToday.Add(3 * time.Hour),
			},
		},
	}

	completedAt := startOfToday.Add(4 * time.Hour)
	taskRepo := &mockAnalyticsTaskRepo{
		tasks: []models.Task{
			{
				ID:          uuid.New(),
				UserID:      userID,
				Status:      models.TaskStatusCompleted,
				CompletedAt: &completedAt,
			},
		},
	}

	sleepH := 7.5
	phoneH := 2.0
	lifestyleRepo := &mockLifestyleRepo{
		record: &models.LifestyleRecord{
			UserID:          userID,
			RecordDate:      startOfToday,
			SleepHours:      &sleepH,
			PhoneUsageHours: &phoneH,
		},
	}
	predRepo := &mockPredictionRepo{}

	svc := services.NewAnalyticsService(studyRepo, focusRepo, taskRepo, lifestyleRepo, predRepo)

	// 1. DAILY ANALYTICS
	daily, err := svc.GetDailyAnalytics(context.Background(), userID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error in GetDailyAnalytics: %v", err)
	}

	if daily.TotalStudySeconds != 3600 {
		t.Errorf("expected 3600 study seconds, got %d (cancelled session was likely included)", daily.TotalStudySeconds)
	}
	if daily.TotalFocusMinutes != 25 {
		t.Errorf("expected 25 focus minutes, got %d", daily.TotalFocusMinutes)
	}
	if daily.TotalBreakMinutes != 5 {
		t.Errorf("expected 5 break minutes, got %d", daily.TotalBreakMinutes)
	}
	if daily.CompletedTasksCount != 1 {
		t.Errorf("expected 1 completed task, got %d", daily.CompletedTasksCount)
	}

	// 2. WEEKLY ANALYTICS
	weekly, err := svc.GetWeeklyAnalytics(context.Background(), userID, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error in GetWeeklyAnalytics: %v", err)
	}

	if len(weekly.DailyPoints) != 7 {
		t.Errorf("expected 7 daily points in weekly analytics, got %d", len(weekly.DailyPoints))
	}
	if weekly.TotalStudySeconds != 3600 {
		t.Errorf("expected 3600 total study seconds in weekly, got %d", weekly.TotalStudySeconds)
	}
	// 1 out of 7 days had study/focus -> consistency should be ~14.3%
	if weekly.StudyConsistencyPct <= 0 {
		t.Errorf("expected positive study consistency percentage, got %v", weekly.StudyConsistencyPct)
	}

	// 3. RANGE ANALYTICS (Reports Module)
	startStr := startOfToday.AddDate(0, 0, -9).Format("2006-01-02")
	endStr := startOfToday.Format("2006-01-02")
	rangeRes, err := svc.GetRangeAnalytics(context.Background(), userID, startStr, endStr, nil)
	if err != nil {
		t.Fatalf("unexpected error in GetRangeAnalytics: %v", err)
	}
	if len(rangeRes.DailyPoints) != 10 {
		t.Errorf("expected 10 daily points in 10-day range analytics, got %d", len(rangeRes.DailyPoints))
	}
	if rangeRes.TotalStudySeconds != 3600 {
		t.Errorf("expected 3600 total study seconds in range analytics, got %d", rangeRes.TotalStudySeconds)
	}

	// Range validation: start date after end date
	_, err = svc.GetRangeAnalytics(context.Background(), userID, endStr, startStr, nil)
	if err == nil {
		t.Errorf("expected error when start_date is after end_date")
	}

	// Range validation: range exceeding 90 days
	tooEarlyStr := startOfToday.AddDate(0, 0, -95).Format("2006-01-02")
	_, err = svc.GetRangeAnalytics(context.Background(), userID, tooEarlyStr, endStr, nil)
	if err == nil {
		t.Errorf("expected error when date range exceeds 90 days")
	}

	// Range validation: invalid date format
	_, err = svc.GetRangeAnalytics(context.Background(), userID, "invalid-date", endStr, nil)
	if err == nil {
		t.Errorf("expected error for malformed start_date format")
	}
}
