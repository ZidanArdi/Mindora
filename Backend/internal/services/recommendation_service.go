package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

type RecommendationService interface {
	GetRecommendations(ctx context.Context, userID uuid.UUID, evaluate bool) ([]models.Recommendation, string, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string) (*models.Recommendation, error)
}

type recommendationService struct {
	recRepo        repositories.RecommendationRepository
	taskRepo       repositories.TaskRepository
	studyRepo      repositories.StudyRepository
	focusRepo      repositories.FocusRepository
	lifestyleRepo  repositories.LifestyleRepository
	predictionRepo repositories.PredictionRepository
	userRepo       repositories.UserRepository
}

func NewRecommendationService(
	recRepo repositories.RecommendationRepository,
	taskRepo repositories.TaskRepository,
	studyRepo repositories.StudyRepository,
	focusRepo repositories.FocusRepository,
	lifestyleRepo repositories.LifestyleRepository,
	predictionRepo repositories.PredictionRepository,
	userRepo repositories.UserRepository,
) RecommendationService {
	return &recommendationService{
		recRepo:        recRepo,
		taskRepo:       taskRepo,
		studyRepo:      studyRepo,
		focusRepo:      focusRepo,
		lifestyleRepo:  lifestyleRepo,
		predictionRepo: predictionRepo,
		userRepo:       userRepo,
	}
}

func (s *recommendationService) GetRecommendations(ctx context.Context, userID uuid.UUID, evaluate bool) ([]models.Recommendation, string, error) {
	now := time.Now().UTC()

	// 1. Expire past valid recommendations
	if err := s.recRepo.ExpirePastValid(ctx, userID, now); err != nil {
		return nil, "", fmt.Errorf("failed to clean expired recommendations: %w", err)
	}

	// 2. Query user metadata for tier resolution
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch user context: %w", err)
	}

	userCreatedAt := user.CreatedAt
	if userCreatedAt.IsZero() {
		userCreatedAt = now
	}

	// 3. Resolve user data tier
	tier, err := s.resolveUserDataTier(ctx, userID, userCreatedAt, now)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve user tier: %w", err)
	}

	// 4. If evaluate is requested, execute deterministic rule evaluation pipeline
	if evaluate {
		if err := s.evaluateRules(ctx, userID, userCreatedAt, tier, now); err != nil {
			return nil, "", fmt.Errorf("failed to evaluate recommendation rules: %w", err)
		}
	}

	// 5. Query active recommendations (status = 'pending' AND valid_until > now)
	activeRecs, err := s.recRepo.FindActive(ctx, userID, now)
	if err != nil {
		return nil, "", fmt.Errorf("failed to retrieve active recommendations: %w", err)
	}

	// 6. Real-time task verification for REC-TASK-001:
	// Verify that overdue high-priority tasks still exist in real-time.
	// If the user already completed or rescheduled them, invalidate on the fly.
	var verifiedRecs []models.Recommendation
	for _, rec := range activeRecs {
		if rec.RuleID == "REC-TASK-001" {
			hasOverdue, checkErr := s.hasOverdueHighPriorityTasks(ctx, userID, now)
			if checkErr == nil && !hasOverdue {
				_ = s.recRepo.InvalidatePendingByID(ctx, rec.ID, userID)
				continue
			}
		}
		verifiedRecs = append(verifiedRecs, rec)
	}

	// 7. Enforce cognitive load cap: at most 3 active recommendations
	if len(verifiedRecs) > 3 {
		verifiedRecs = verifiedRecs[:3]
	}

	return verifiedRecs, tier, nil
}

func (s *recommendationService) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string) (*models.Recommendation, error) {
	now := time.Now().UTC()
	var dismissedUntil *time.Time

	switch status {
	case models.RecStatusDismissed:
		cooldown := now.Add(48 * time.Hour)
		dismissedUntil = &cooldown
	case models.RecStatusCompleted:
		dismissedUntil = nil
	default:
		return nil, fmt.Errorf("invalid recommendation status update: %s", status)
	}

	return s.recRepo.UpdateStatus(ctx, id, userID, status, dismissedUntil)
}

// =========================================================================
// Deterministic Tier Resolution (Waterfall Specification)
// =========================================================================

func (s *recommendationService) resolveUserDataTier(ctx context.Context, userID uuid.UUID, userCreatedAt, now time.Time) (string, error) {
	accountAgeDays := int(math.Floor(now.Sub(userCreatedAt).Hours() / 24.0))

	// Rolling 7-day window
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	sevenDaysAgo := todayStart.AddDate(0, 0, -6)

	// Check completed study sessions past 7 days (duration >= 600s)
	studySessions, err := s.studyRepo.FindCompletedSessions(ctx, userID, sevenDaysAgo, now)
	if err != nil {
		return "", err
	}
	activeStudyDates := make(map[string]struct{})
	hasActivityToday := false
	for _, sess := range studySessions {
		if sess.DurationSeconds >= 600 {
			dateKey := sess.StartedAt.UTC().Format("2006-01-02")
			activeStudyDates[dateKey] = struct{}{}
		}
		if sess.StartedAt.UTC().After(todayStart) {
			hasActivityToday = true
		}
	}
	activeStudyDaysPast7 := len(activeStudyDates)

	// Check lifestyle records past 7 days
	lifestyleRecords, err := s.lifestyleRepo.FindByDateRange(ctx, userID, sevenDaysAgo, now)
	if err != nil {
		return "", err
	}
	recordedLifestyleDaysPast7 := len(lifestyleRecords)
	for _, lr := range lifestyleRecords {
		if lr.RecordDate.UTC().Equal(todayStart) || lr.RecordDate.UTC().After(todayStart) {
			hasActivityToday = true
		}
	}

	// Check today's focus intervals
	focusIntervals, err := s.focusRepo.FindCompletedIntervals(ctx, userID, todayStart, now)
	if err == nil && len(focusIntervals) > 0 {
		hasActivityToday = true
	}

	// Check Tier 4: Longitudinal Trajectory
	// AccountAge >= 14d AND N_hist(PersonalizedPredictions past 14d) >= 3
	fourteenDaysAgo := todayStart.AddDate(0, 0, -14)
	yesterdayEnd := todayStart.Add(-time.Nanosecond)
	predictionHistory, _, predErr := s.predictionRepo.FindHistory(ctx, userID, 30, 0)
	personalizedDaysPast14 := 0
	if predErr == nil {
		distinctPersonalizedDates := make(map[string]struct{})
		for _, pred := range predictionHistory {
			if pred.PredictionDate.After(fourteenDaysAgo.AddDate(0, 0, -1)) && pred.PredictionDate.Before(yesterdayEnd.Add(time.Second)) {
				// Must satisfy personalization criteria
				if !pred.IsBaseline && len(pred.UnrecordedFeatures) == 0 {
					dateKey := pred.PredictionDate.Format("2006-01-02")
					distinctPersonalizedDates[dateKey] = struct{}{}
				}
			}
		}
		personalizedDaysPast14 = len(distinctPersonalizedDates)
	}

	if accountAgeDays >= 14 && personalizedDaysPast14 >= 3 {
		return "longitudinal", nil
	}

	// Check Tier 3: Habit Calibrated
	if activeStudyDaysPast7 >= 4 || recordedLifestyleDaysPast7 >= 4 {
		return "calibrated", nil
	}

	// Check Tier 2: Nascent / Intraday Active
	if hasActivityToday || (activeStudyDaysPast7 >= 1 && activeStudyDaysPast7 < 4) {
		return "nascent", nil
	}

	// Tier 1: Cold Start / Inactive Fallback
	if accountAgeDays <= 2 {
		return "cold_start", nil
	}
	return "dormant", nil
}

// =========================================================================
// Deterministic Rule Evaluation Pipeline
// =========================================================================

func (s *recommendationService) evaluateRules(ctx context.Context, userID uuid.UUID, userCreatedAt time.Time, tier string, now time.Time) error {
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	// -------------------------------------------------------------------------
	// Rule 1: REC-TASK-001 — Overdue High-Priority Academic Deadlines
	// -------------------------------------------------------------------------
	if err := s.evaluateRecTask001(ctx, userID, now); err != nil {
		return err
	}

	// -------------------------------------------------------------------------
	// Rule 2: REC-SLEEP-001 — Multi-Day Sleep Deficit (< 6.0h over >= 3 recorded days)
	// Permitted scope: Tier 3 & Tier 4 (suppressed in Tier 1 and Tier 2)
	// -------------------------------------------------------------------------
	if tier == "calibrated" || tier == "longitudinal" {
		if err := s.evaluateRecSleep001(ctx, userID, todayStart, todayEnd, now); err != nil {
			return err
		}
	}

	// -------------------------------------------------------------------------
	// Rule 3: REC-COLD-001 — Cold-Start Habit Calibration Prompt
	// Permitted scope: Tier 1A, Tier 2, Tier 3 (when habits missing)
	// -------------------------------------------------------------------------
	if tier == "cold_start" || tier == "nascent" || tier == "calibrated" {
		if err := s.evaluateRecCold001(ctx, userID, todayStart, todayEnd, now); err != nil {
			return err
		}
	}

	// -------------------------------------------------------------------------
	// Rule 4: REC-BREAK-001 — Focus-to-Break Ratio Imbalance in Pomodoro Timer
	// Permitted scope: Tier 2, Tier 3, Tier 4
	// -------------------------------------------------------------------------
	if tier == "nascent" || tier == "calibrated" || tier == "longitudinal" {
		if err := s.evaluateRecBreak001(ctx, userID, todayStart, todayEnd, now); err != nil {
			return err
		}
	}

	// -------------------------------------------------------------------------
	// Rule 5: REC-STUDY-001 — Study Consistency & Habit Cadence
	// Permitted scope: Tier 3, Tier 4 (Account age >= 7d)
	// -------------------------------------------------------------------------
	if tier == "calibrated" || tier == "longitudinal" {
		if err := s.evaluateRecStudy001(ctx, userID, userCreatedAt, todayStart, todayEnd, now); err != nil {
			return err
		}
	}

	return nil
}

// -----------------------------------------------------------------------------
// Rule 1 Implementation: REC-TASK-001
// -----------------------------------------------------------------------------
func (s *recommendationService) evaluateRecTask001(ctx context.Context, userID uuid.UUID, now time.Time) error {
	ruleID := "REC-TASK-001"

	// Check dismissal cooldown
	if dismissed, err := s.recRepo.FindDismissedRecent(ctx, userID, ruleID, now); err != nil {
		return err
	} else if dismissed != nil {
		return nil
	}

	// Query pending/in_progress high-priority tasks
	tasks, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Priority: models.TaskPriorityHigh,
		Limit:    100,
	})
	if err != nil {
		return err
	}

	var overdueTasks []models.Task
	for _, t := range tasks {
		if (t.Status == models.TaskStatusPending || t.Status == models.TaskStatusInProgress) &&
			t.Deadline != nil && t.Deadline.Before(now) {
			overdueTasks = append(overdueTasks, t)
		}
	}

	// If no overdue high-priority tasks, invalidate any pending REC-TASK-001
	if len(overdueTasks) == 0 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	// Select the most urgent task (oldest deadline)
	topTask := overdueTasks[0]
	for _, t := range overdueTasks {
		if t.Deadline.Before(*topTask.Deadline) {
			topTask = t
		}
	}

	evalKey := fmt.Sprintf("REC-TASK-001:%s:%s", userID.String(), topTask.ID.String())

	// Check period completion suppression
	if completed, err := s.recRepo.FindByEvaluationKeyAndStatus(ctx, userID, evalKey, models.RecStatusCompleted); err != nil {
		return err
	} else if completed != nil {
		return nil
	}

	deadlineFormatted := topTask.Deadline.UTC().Format("Jan 02, 2006 15:04 MST")
	deadlineRFC3339 := topTask.Deadline.UTC().Format(time.RFC3339)

	evidenceMap := map[string]interface{}{
		"overdue_high_priority_count": len(overdueTasks),
		"top_task_id":                 topTask.ID.String(),
		"top_task_title":              topTask.Title,
		"top_task_category":           topTask.Category,
		"deadline_utc":                deadlineRFC3339,
		"data_provenance":             "tasks",
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	actionRoute := "/tasks"
	actionLabel := "Review Tasks"
	validUntil := now.Add(24 * time.Hour)

	rec := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                userID,
		RuleID:                ruleID,
		Category:              "Task Prioritization",
		Priority:              models.RecPriorityCritical,
		Title:                 "Action Required: Overdue High-Priority Tasks",
		Message:               fmt.Sprintf("You have %d overdue high-priority task(s), including '%s' (Due: %s).", len(overdueTasks), topTask.Title, deadlineFormatted),
		Explanation:           "Overdue high-priority deliverables generate persistent background task anxiety. Completing or rescheduling the most urgent item clears working memory for ongoing study.",
		Evidence:              models.JSONB(evidenceBytes),
		ActionRoute:           &actionRoute,
		ActionLabel:           &actionLabel,
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: now,
		EvaluationPeriodEnd:   now.Add(24 * time.Hour),
		EvaluationKey:         evalKey,
		ValidUntil:            validUntil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return s.recRepo.UpsertPending(ctx, rec)
}

func (s *recommendationService) hasOverdueHighPriorityTasks(ctx context.Context, userID uuid.UUID, now time.Time) (bool, error) {
	tasks, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Priority: models.TaskPriorityHigh,
		Limit:    100,
	})
	if err != nil {
		return false, err
	}
	for _, t := range tasks {
		if (t.Status == models.TaskStatusPending || t.Status == models.TaskStatusInProgress) &&
			t.Deadline != nil && t.Deadline.Before(now) {
			return true, nil
		}
	}
	return false, nil
}

// -----------------------------------------------------------------------------
// Rule 2 Implementation: REC-SLEEP-001
// -----------------------------------------------------------------------------
func (s *recommendationService) evaluateRecSleep001(ctx context.Context, userID uuid.UUID, todayStart, todayEnd, now time.Time) error {
	ruleID := "REC-SLEEP-001"

	// Check dismissal cooldown
	if dismissed, err := s.recRepo.FindDismissedRecent(ctx, userID, ruleID, now); err != nil {
		return err
	} else if dismissed != nil {
		return nil
	}

	// 7-day window: [today - 6d, today]
	startDate := todayStart.AddDate(0, 0, -6)
	records, err := s.lifestyleRepo.FindByDateRange(ctx, userID, startDate, todayEnd)
	if err != nil {
		return err
	}

	recordedDays := 0
	totalSleepHours := 0.0

	for _, r := range records {
		// Strict SQL NULL vs explicit zero check:
		// NULL (nil) is strictly excluded from count and sum.
		// Explicit zero (*r.SleepHours == 0.0) is included.
		if r.SleepHours != nil {
			recordedDays++
			totalSleepHours += *r.SleepHours
		}
	}

	// Mathematical trigger condition:
	// N_sleep >= 3 AND avg_sleep < 6.00
	if recordedDays < 3 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	avgSleep := totalSleepHours / float64(recordedDays)
	if avgSleep >= 6.00 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	dateStartStr := startDate.Format("2006-01-02")
	dateEndStr := todayStart.Format("2006-01-02")
	evalKey := fmt.Sprintf("REC-SLEEP-001:%s:%s:%s", userID.String(), dateStartStr, dateEndStr)

	// Check period completion suppression
	if completed, err := s.recRepo.FindByEvaluationKeyAndStatus(ctx, userID, evalKey, models.RecStatusCompleted); err != nil {
		return err
	} else if completed != nil {
		return nil
	}

	evidenceMap := map[string]interface{}{
		"average_recorded_sleep_hours": math.Round(avgSleep*100) / 100,
		"recorded_sleep_days":          recordedDays,
		"evaluation_window_days":       7,
		"wellness_reference_target":    "7.0 to 8.5 hours per night",
		"data_provenance":              "lifestyle_records",
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	actionRoute := "/lifestyle"
	actionLabel := "Log Lifestyle Habits"
	validUntil := now.Add(48 * time.Hour)

	rec := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                userID,
		RuleID:                ruleID,
		Category:              "Sleep Hygiene & Rest",
		Priority:              models.RecPriorityCritical,
		Title:                 "Average Sleep Below Recommended Target",
		Message:               fmt.Sprintf("Your recorded sleep averaged %.1f hours over %d days in the past week.", avgSleep, recordedDays),
		Explanation:           "Sustained sleep duration below 6 hours is associated with decreased working memory capacity, slower problem-solving speed, and heightened task difficulty.",
		Evidence:              models.JSONB(evidenceBytes),
		ActionRoute:           &actionRoute,
		ActionLabel:           &actionLabel,
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: startDate,
		EvaluationPeriodEnd:   todayEnd,
		EvaluationKey:         evalKey,
		ValidUntil:            validUntil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return s.recRepo.UpsertPending(ctx, rec)
}

// -----------------------------------------------------------------------------
// Rule 3 Implementation: REC-COLD-001
// -----------------------------------------------------------------------------
func (s *recommendationService) evaluateRecCold001(ctx context.Context, userID uuid.UUID, todayStart, todayEnd, now time.Time) error {
	ruleID := "REC-COLD-001"

	// Check dismissal cooldown
	if dismissed, err := s.recRepo.FindDismissedRecent(ctx, userID, ruleID, now); err != nil {
		return err
	} else if dismissed != nil {
		return nil
	}

	// Query today's lifestyle record
	todayLifestyle, err := s.lifestyleRepo.FindByDate(ctx, userID, todayStart)
	if err != nil && !errorsIsLifestyleNotFound(err) {
		return err
	}

	// Query latest prediction
	latestPred, err := s.predictionRepo.FindLatest(ctx, userID)
	if err != nil {
		return err
	}

	// Determine unrecorded features for today
	var unrecordedFeatures []string
	if todayLifestyle == nil {
		unrecordedFeatures = []string{"sleep_hours", "phone_usage_hours", "stress_level"}
	} else {
		if todayLifestyle.SleepHours == nil {
			unrecordedFeatures = append(unrecordedFeatures, "sleep_hours")
		}
		if todayLifestyle.PhoneUsageHours == nil {
			unrecordedFeatures = append(unrecordedFeatures, "phone_usage_hours")
		}
		if todayLifestyle.StressLevel == nil {
			unrecordedFeatures = append(unrecordedFeatures, "stress_level")
		}
	}

	// Trigger condition:
	// Missing daily habit features OR latest prediction is baseline
	isTriggered := false
	if len(unrecordedFeatures) > 0 {
		isTriggered = true
	} else if latestPred == nil || latestPred.IsBaseline {
		isTriggered = true
	}

	// If all habits are recorded and prediction is calibrated, invalidate pending
	if !isTriggered {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	evalKey := fmt.Sprintf("REC-COLD-001:%s:calibration", userID.String())

	// Check period completion suppression
	if completed, err := s.recRepo.FindByEvaluationKeyAndStatus(ctx, userID, evalKey, models.RecStatusCompleted); err != nil {
		return err
	} else if completed != nil {
		return nil
	}

	unrecordedFormatted := "daily habits (sleep, screen time, stress)"
	if len(unrecordedFeatures) > 0 {
		unrecordedFormatted = strings.Join(unrecordedFeatures, ", ")
	}

	evidenceMap := map[string]interface{}{
		"prediction_state":    "baseline",
		"unrecorded_features": unrecordedFeatures,
		"action_route":        "/lifestyle",
		"data_provenance":     "productivity_predictions, lifestyle_records",
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	actionRoute := "/lifestyle"
	actionLabel := "Calibrate Habits"
	validUntil := now.Add(48 * time.Hour)

	rec := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                userID,
		RuleID:                ruleID,
		Category:              "Lifestyle Balance",
		Priority:              models.RecPriorityMedium,
		Title:                 "Calibrate Your Productivity Baseline",
		Message:               fmt.Sprintf("Your productivity estimate is currently using reference defaults because %s have not been recorded today.", unrecordedFormatted),
		Explanation:           "Mindora uses your verified study time, sleep, screen time, and stress levels to produce personalized productivity scores and recommendations. Logging your habits unlocks calibrated guidance.",
		Evidence:              models.JSONB(evidenceBytes),
		ActionRoute:           &actionRoute,
		ActionLabel:           &actionLabel,
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: todayStart,
		EvaluationPeriodEnd:   todayEnd,
		EvaluationKey:         evalKey,
		ValidUntil:            validUntil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return s.recRepo.UpsertPending(ctx, rec)
}

func errorsIsLifestyleNotFound(err error) bool {
	return err != nil && (err == repositories.ErrLifestyleRecordNotFound || strings.Contains(err.Error(), "not found"))
}

// -----------------------------------------------------------------------------
// Rule 4 Implementation: REC-BREAK-001
// -----------------------------------------------------------------------------
func (s *recommendationService) evaluateRecBreak001(ctx context.Context, userID uuid.UUID, todayStart, todayEnd, now time.Time) error {
	ruleID := "REC-BREAK-001"

	// Check dismissal cooldown
	if dismissed, err := s.recRepo.FindDismissedRecent(ctx, userID, ruleID, now); err != nil {
		return err
	} else if dismissed != nil {
		return nil
	}

	// Operates exclusively on Pomodoro intervals logged in focus_sessions
	dailyStats, err := s.focusRepo.GetDailyStatistics(ctx, userID, todayStart, todayEnd)
	if err != nil {
		return err
	}

	focusMinutes := dailyStats.CompletedFocusMinutes
	breakMinutes := dailyStats.CompletedBreakMinutes

	// Mathematical trigger condition:
	// CompletedFocusMinutes >= 100 AND CompletedBreakMinutes < 10
	if focusMinutes < 100 || breakMinutes >= 10 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	dateStr := todayStart.Format("2006-01-02")
	evalKey := fmt.Sprintf("REC-BREAK-001:%s:%s", userID.String(), dateStr)

	// Check period completion suppression
	if completed, err := s.recRepo.FindByEvaluationKeyAndStatus(ctx, userID, evalKey, models.RecStatusCompleted); err != nil {
		return err
	} else if completed != nil {
		return nil
	}

	breakRatioPct := 0.0
	if focusMinutes > 0 {
		breakRatioPct = math.Round((float64(breakMinutes)/float64(focusMinutes))*1000) / 10
	}

	evidenceMap := map[string]interface{}{
		"today_completed_focus_minutes": focusMinutes,
		"today_completed_break_minutes": breakMinutes,
		"derived_break_ratio_pct":       breakRatioPct,
		"product_heuristic_threshold":   "At least 10 minutes of break per 100 minutes of focus (1:10 ratio)",
		"data_provenance":               "focus_sessions",
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	actionRoute := "/focus"
	actionLabel := "Take a Break"
	validUntil := todayEnd

	rec := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                userID,
		RuleID:                ruleID,
		Category:              "Break Management & Recovery",
		Priority:              models.RecPriorityMedium,
		Title:                 "Rest Interval Recommended for Timer Sessions",
		Message:               fmt.Sprintf("You have recorded %d minutes of focus intervals today with only %d minutes of logged break intervals.", focusMinutes, breakMinutes),
		Explanation:           "Structured interval pacing methods (such as the 25:5 Pomodoro technique) incorporate short recovery periods to sustain task vigilance and prevent mental exhaustion.",
		Evidence:              models.JSONB(evidenceBytes),
		ActionRoute:           &actionRoute,
		ActionLabel:           &actionLabel,
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: todayStart,
		EvaluationPeriodEnd:   todayEnd,
		EvaluationKey:         evalKey,
		ValidUntil:            validUntil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return s.recRepo.UpsertPending(ctx, rec)
}

// -----------------------------------------------------------------------------
// Rule 5 Implementation: REC-STUDY-001
// -----------------------------------------------------------------------------
func (s *recommendationService) evaluateRecStudy001(ctx context.Context, userID uuid.UUID, userCreatedAt, todayStart, todayEnd, now time.Time) error {
	ruleID := "REC-STUDY-001"

	// Check dismissal cooldown
	if dismissed, err := s.recRepo.FindDismissedRecent(ctx, userID, ruleID, now); err != nil {
		return err
	} else if dismissed != nil {
		return nil
	}

	// Suppression condition: AccountAgeDays >= 7
	accountAgeDays := int(math.Floor(now.Sub(userCreatedAt).Hours() / 24.0))
	if accountAgeDays < 7 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	// Rolling 7-day study history: study_sessions (status = 'completed', duration >= 600s)
	sevenDaysAgo := todayStart.AddDate(0, 0, -6)
	studySessions, err := s.studyRepo.FindCompletedSessions(ctx, userID, sevenDaysAgo, todayEnd)
	if err != nil {
		return err
	}

	activeDates := make(map[string]struct{})
	for _, sess := range studySessions {
		if sess.DurationSeconds >= 600 {
			activeDates[sess.StartedAt.UTC().Format("2006-01-02")] = struct{}{}
		}
	}
	activeStudyDays := len(activeDates)

	// Pending study tasks count
	tasks, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Category: "Study",
		Limit:    100,
	})
	if err != nil {
		return err
	}

	pendingStudyTasks := 0
	for _, t := range tasks {
		if t.Status == models.TaskStatusPending || t.Status == models.TaskStatusInProgress {
			pendingStudyTasks++
		}
	}

	// Mathematical trigger condition:
	// ActiveStudyDays(D-6..D) <= 2 AND PendingStudyTasks >= 1
	if activeStudyDays > 2 || pendingStudyTasks < 1 {
		return s.recRepo.InvalidatePendingByRule(ctx, userID, ruleID)
	}

	// Supplementary context: focus intervals count
	focusIntervals, _ := s.focusRepo.FindCompletedIntervals(ctx, userID, sevenDaysAgo, todayEnd)
	focusCount := len(focusIntervals)

	dateStartStr := sevenDaysAgo.Format("2006-01-02")
	dateEndStr := todayStart.Format("2006-01-02")
	evalKey := fmt.Sprintf("REC-STUDY-001:%s:%s:%s", userID.String(), dateStartStr, dateEndStr)

	// Check period completion suppression
	if completed, err := s.recRepo.FindByEvaluationKeyAndStatus(ctx, userID, evalKey, models.RecStatusCompleted); err != nil {
		return err
	} else if completed != nil {
		return nil
	}

	evidenceMap := map[string]interface{}{
		"observed_study_days":                 activeStudyDays,
		"evaluation_window_days":              7,
		"pending_academic_tasks":              pendingStudyTasks,
		"supplementary_focus_intervals_count": focusCount,
		"product_heuristic_threshold":         "At least 3 active study days per week",
		"data_provenance":                     "study_sessions (primary), tasks (primary), focus_sessions (supplementary)",
	}
	evidenceBytes, _ := json.Marshal(evidenceMap)

	actionRoute := "/study"
	actionLabel := "Schedule Study Block"
	validUntil := now.Add(24 * time.Hour)

	rec := &models.Recommendation{
		ID:                    uuid.New(),
		UserID:                userID,
		RuleID:                ruleID,
		Category:              "Study Consistency",
		Priority:              models.RecPriorityHigh,
		Title:                 "Establish a Consistent Study Schedule",
		Message:               fmt.Sprintf("You recorded academic study sessions on %d of the past 7 days, while having %d active study task(s) in your queue.", activeStudyDays, pendingStudyTasks),
		Explanation:           "Educational psychology research demonstrates that distributed practice (spacing study across regular daily sessions) yields greater retention and lower cognitive strain than intermittent cramming.",
		Evidence:              models.JSONB(evidenceBytes),
		ActionRoute:           &actionRoute,
		ActionLabel:           &actionLabel,
		Status:                models.RecStatusPending,
		EvaluationPeriodStart: sevenDaysAgo,
		EvaluationPeriodEnd:   todayEnd,
		EvaluationKey:         evalKey,
		ValidUntil:            validUntil,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	return s.recRepo.UpsertPending(ctx, rec)
}
