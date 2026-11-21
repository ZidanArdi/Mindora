package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
)

func setupTestEnvironment(t *testing.T) (*database.Database, services.RecommendationService, *models.User) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}

	ctx := context.Background()
	_ = database.SeedRBAC(ctx, db.DB)

	userRepo := repositories.NewUserRepository(db.DB)
	taskRepo := repositories.NewTaskRepository(db.DB)
	studyRepo := repositories.NewStudyRepository(db.DB)
	focusRepo := repositories.NewFocusRepository(db.DB)
	lifestyleRepo := repositories.NewLifestyleRepository(db.DB)
	predictionRepo := repositories.NewPredictionRepository(db.DB)
	recRepo := repositories.NewRecommendationRepository(db.DB)

	recService := services.NewRecommendationService(
		recRepo,
		taskRepo,
		studyRepo,
		focusRepo,
		lifestyleRepo,
		predictionRepo,
		userRepo,
	)

	// Create unique test user with role "user"
	role, err := userRepo.GetRoleByName(ctx, "user")
	if err != nil {
		t.Fatalf("failed to get user role: %v", err)
	}

	pw := "hashed_dummy_pw"
	testUser := &models.User{
		ID:           uuid.New(),
		Email:        "rec_test_" + uuid.New().String()[:8] + "@mindora.app",
		PasswordHash: &pw,
		Name:         "Recommendation Test Student",
		RoleID:       role.ID,
		Status:       "active",
		CreatedAt:    time.Now().UTC().AddDate(0, 0, -10), // 10 days old account
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	return db, recService, testUser
}

// -----------------------------------------------------------------------------
// 1. Rule Unit Test: REC-TASK-001 (Overdue High Priority Tasks)
// -----------------------------------------------------------------------------
func TestRecommendation_RecTask001(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	taskRepo := repositories.NewTaskRepository(db.DB)

	// Overdue high-priority task
	overdueDeadline := time.Now().UTC().Add(-2 * time.Hour)
	task := &models.Task{
		ID:       uuid.New(),
		UserID:   user.ID,
		Title:    "Urgent Lab Assignment",
		Category: "Study",
		Priority: models.TaskPriorityHigh,
		Status:   models.TaskStatusPending,
		Deadline: &overdueDeadline,
	}
	if err := taskRepo.Create(ctx, task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	recs, tier, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if tier == "" {
		t.Errorf("expected resolved tier, got empty")
	}

	found := false
	for _, r := range recs {
		if r.RuleID == "REC-TASK-001" {
			found = true
			if r.Priority != models.RecPriorityCritical {
				t.Errorf("expected critical priority, got %s", r.Priority)
			}
			if r.ActionRoute == nil || *r.ActionRoute != "/tasks" {
				t.Errorf("expected action route /tasks, got %v", r.ActionRoute)
			}
		}
	}
	if !found {
		t.Errorf("expected REC-TASK-001 recommendation to trigger for overdue task")
	}

	// Now complete the task and re-evaluate -> real-time verification should invalidate it
	completedAt := time.Now().UTC()
	_, err = taskRepo.UpdateStatus(ctx, task.ID, user.ID, models.TaskStatusCompleted, &completedAt)
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	recsAfter, _, err := svc.GetRecommendations(ctx, user.ID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range recsAfter {
		if r.RuleID == "REC-TASK-001" {
			t.Errorf("expected REC-TASK-001 to be invalidated after task completed")
		}
	}
}

// -----------------------------------------------------------------------------
// 2. Rule Unit & NULL vs Zero Test: REC-SLEEP-001
// -----------------------------------------------------------------------------
func TestRecommendation_RecSleep001_NullVsZero(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	lifestyleRepo := repositories.NewLifestyleRepository(db.DB)
	now := time.Now().UTC()

	// To reach Tier 3 (Calibrated), user needs at least 4 lifestyle days or study days past 7.
	// Scenario A: 4 lifestyle records, but only 2 have sleep_hours recorded (2 days have sleep NULL).
	// N_sleep = 2 (< 3 days required).
	// MUST be suppressed! NULL is never treated as 0.
	val5 := 5.0
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		ID:         uuid.New(),
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -1),
		SleepHours: &val5,
	})
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		ID:         uuid.New(),
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -2),
		SleepHours: &val5,
	})
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		ID:         uuid.New(),
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -4),
		SleepHours: nil, // unrecorded
	})
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		ID:         uuid.New(),
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -5),
		SleepHours: nil, // unrecorded
	})

	recs, tier, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tier != "calibrated" {
		t.Fatalf("expected calibrated tier with 4 lifestyle records, got %s", tier)
	}
	for _, r := range recs {
		if r.RuleID == "REC-SLEEP-001" {
			t.Errorf("REC-SLEEP-001 must NOT trigger when recorded sleep days < 3 (NULL must not be coerced to 0)")
		}
	}

	// Scenario B: Add 3rd recorded day with explicit zero (0.0h).
	// Total recorded sleep days = 3. Recorded hours: 5.0 + 5.0 + 0.0 = 10.0.
	// Average = 3.33h < 6.00h -> MUST TRIGGER! Explicit zero is valid evidence.
	valZero := 0.0
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		ID:         uuid.New(),
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -3),
		SleepHours: &valZero,
	})

	recs, _, err = svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, r := range recs {
		if r.RuleID == "REC-SLEEP-001" {
			found = true
			if r.Priority != models.RecPriorityCritical {
				t.Errorf("expected critical priority, got %s", r.Priority)
			}
		}
	}
	if !found {
		t.Errorf("expected REC-SLEEP-001 to trigger when 3 recorded days average < 6.0h with explicit zero")
	}

	// Scenario C: Healthy sleep average (e.g. 7.5h) -> MUST be suppressed.
	val8 := 8.0
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -1),
		SleepHours: &val8,
	})
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -2),
		SleepHours: &val8,
	})
	_ = lifestyleRepo.UpsertDailyRecord(ctx, &models.LifestyleRecord{
		UserID:     user.ID,
		RecordDate: now.AddDate(0, 0, -3),
		SleepHours: &val8,
	})

	recs, _, err = svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range recs {
		if r.RuleID == "REC-SLEEP-001" {
			t.Errorf("REC-SLEEP-001 must NOT trigger when sleep average is >= 6.0h")
		}
	}
}

// -----------------------------------------------------------------------------
// 3. Rule Unit Test: REC-BREAK-001 (Break Ratio in Pomodoro Timer)
// -----------------------------------------------------------------------------
func TestRecommendation_RecBreak001(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	focusRepo := repositories.NewFocusRepository(db.DB)
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// 120 minutes of focus, 5 minutes of break (< 10m break per 100m focus)
	_ = focusRepo.CreateInterval(ctx, &models.FocusSession{
		ID:              uuid.New(),
		UserID:          user.ID,
		SessionType:     models.FocusTypeWork,
		DurationMinutes: 120,
		Status:          models.FocusStatusCompleted,
		StartedAt:       todayStart,
	})
	_ = focusRepo.CreateInterval(ctx, &models.FocusSession{
		ID:              uuid.New(),
		UserID:          user.ID,
		SessionType:     models.FocusTypeBreak,
		DurationMinutes: 5,
		Status:          models.FocusStatusCompleted,
		StartedAt:       todayStart.Add(1 * time.Second),
	})

	recs, _, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, r := range recs {
		if r.RuleID == "REC-BREAK-001" {
			found = true
			if r.Priority != models.RecPriorityMedium {
				t.Errorf("expected medium priority, got %s", r.Priority)
			}
		}
	}
	if !found {
		t.Errorf("expected REC-BREAK-001 to trigger for 120m focus with only 5m break")
	}
}

// -----------------------------------------------------------------------------
// 4. Duplicate Prevention & Atomic Upsert Test
// -----------------------------------------------------------------------------
func TestRecommendation_DuplicatePrevention(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	taskRepo := repositories.NewTaskRepository(db.DB)
	overdueDeadline := time.Now().UTC().Add(-1 * time.Hour)
	_ = taskRepo.Create(ctx, &models.Task{
		ID:       uuid.New(),
		UserID:   user.ID,
		Title:    "Idempotency Task",
		Category: "Study",
		Priority: models.TaskPriorityHigh,
		Status:   models.TaskStatusPending,
		Deadline: &overdueDeadline,
	})

	// Evaluate twice consecutively
	_, _, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("first eval failed: %v", err)
	}

	recs2, _, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("second eval failed: %v", err)
	}

	recCount := 0
	for _, r := range recs2 {
		if r.RuleID == "REC-TASK-001" {
			recCount++
		}
	}
	if recCount != 1 {
		t.Errorf("expected exactly 1 active REC-TASK-001 recommendation, got %d", recCount)
	}
}

// -----------------------------------------------------------------------------
// 5. Dismissal Cooldown (48h) & Completion Lifecycle Test
// -----------------------------------------------------------------------------
func TestRecommendation_DismissalAndCompletion(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	taskRepo := repositories.NewTaskRepository(db.DB)
	overdueDeadline := time.Now().UTC().Add(-1 * time.Hour)
	_ = taskRepo.Create(ctx, &models.Task{
		ID:       uuid.New(),
		UserID:   user.ID,
		Title:    "Dismissal Lifecycle Task",
		Category: "Study",
		Priority: models.TaskPriorityHigh,
		Status:   models.TaskStatusPending,
		Deadline: &overdueDeadline,
	})

	recs, _, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil || len(recs) == 0 {
		t.Fatalf("expected active recommendation, got %v", err)
	}

	targetRec := recs[0]

	// Dismiss recommendation
	updated, err := svc.UpdateStatus(ctx, targetRec.ID, user.ID, models.RecStatusDismissed)
	if err != nil {
		t.Fatalf("failed to dismiss recommendation: %v", err)
	}
	if updated.Status != models.RecStatusDismissed {
		t.Errorf("expected status dismissed, got %s", updated.Status)
	}
	if updated.DismissedUntil == nil || updated.DismissedUntil.Before(time.Now().UTC().Add(47*time.Hour)) {
		t.Errorf("expected dismissed_until to be ~48 hours in future")
	}

	// Re-evaluate: must be suppressed due to 48-hour cooldown
	recsAfterDismiss, _, err := svc.GetRecommendations(ctx, user.ID, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, r := range recsAfterDismiss {
		if r.RuleID == targetRec.RuleID {
			t.Errorf("rule %s should be suppressed during 48-hour dismissal cooldown", targetRec.RuleID)
		}
	}
}

// -----------------------------------------------------------------------------
// 6. Maximum 3 Active Recommendations Cap
// -----------------------------------------------------------------------------
func TestRecommendation_MaxThreeActiveCap(t *testing.T) {
	db, svc, user := setupTestEnvironment(t)
	defer db.Close()
	ctx := context.Background()

	recRepo := repositories.NewRecommendationRepository(db.DB)
	now := time.Now().UTC()

	// Insert 5 pending recommendations directly into DB
	rules := []struct {
		id       string
		priority string
	}{
		{"REC-1", models.RecPriorityCritical},
		{"REC-2", models.RecPriorityHigh},
		{"REC-3", models.RecPriorityHigh},
		{"REC-4", models.RecPriorityMedium},
		{"REC-5", models.RecPriorityLow},
	}

	for _, r := range rules {
		_ = recRepo.UpsertPending(ctx, &models.Recommendation{
			ID:                    uuid.New(),
			UserID:                user.ID,
			RuleID:                r.id,
			Category:              "General",
			Priority:              r.priority,
			Title:                 "Test " + r.id,
			Message:               "Message " + r.id,
			Explanation:           "Explanation " + r.id,
			Evidence:              models.JSONB([]byte("{}")),
			Status:                models.RecStatusPending,
			EvaluationPeriodStart: now,
			EvaluationPeriodEnd:   now.Add(24 * time.Hour),
			EvaluationKey:         r.id + ":" + user.ID.String(),
			ValidUntil:            now.Add(24 * time.Hour),
			CreatedAt:             now,
			UpdatedAt:             now,
		})
	}

	recs, _, err := svc.GetRecommendations(ctx, user.ID, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(recs) > 3 {
		t.Errorf("expected at most 3 active recommendations, got %d", len(recs))
	}
}
