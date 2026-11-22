package services

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

type AnalyticsService interface {
	GetDailyAnalytics(ctx context.Context, userID uuid.UUID, dateStr *string, tzName *string) (*models.DailyAnalytics, error)
	GetWeeklyAnalytics(ctx context.Context, userID uuid.UUID, endDateStr *string, tzName *string) (*models.WeeklyAnalytics, error)
	GetMonthlyAnalytics(ctx context.Context, userID uuid.UUID, endDateStr *string, tzName *string) (*models.MonthlyAnalytics, error)
	GetRangeAnalytics(ctx context.Context, userID uuid.UUID, startDateStr, endDateStr string, tzName *string) (*models.WeeklyAnalytics, error)
	GetAnalyticsSummary(ctx context.Context, userID uuid.UUID, tzName *string) (*models.AnalyticsSummary, error)
}

type analyticsService struct {
	studyRepo      repositories.StudyRepository
	focusRepo      repositories.FocusRepository
	taskRepo       repositories.TaskRepository
	lifestyleRepo  repositories.LifestyleRepository
	predictionRepo repositories.PredictionRepository
}

func NewAnalyticsService(
	studyRepo repositories.StudyRepository,
	focusRepo repositories.FocusRepository,
	taskRepo repositories.TaskRepository,
	lifestyleRepo repositories.LifestyleRepository,
	predictionRepo repositories.PredictionRepository,
) AnalyticsService {
	return &analyticsService{
		studyRepo:      studyRepo,
		focusRepo:      focusRepo,
		taskRepo:       taskRepo,
		lifestyleRepo:  lifestyleRepo,
		predictionRepo: predictionRepo,
	}
}

func (s *analyticsService) GetDailyAnalytics(ctx context.Context, userID uuid.UUID, dateStr *string, tzName *string) (*models.DailyAnalytics, error) {
	loc := parseUserLocation(tzName)

	var targetDate time.Time
	if dateStr != nil && *dateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *dateStr, loc)
		if err != nil {
			return nil, err
		}
		targetDate = parsed
	} else {
		now := time.Now().In(loc)
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc).UTC()
	endOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, loc).UTC()

	// 1. Study Sessions
	studyStats, err := s.studyRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 2. Focus Sessions (Focus + Break)
	focusStats, err := s.focusRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 3. Completed Tasks on this day
	tasksCount, err := s.taskRepo.CountCompletedByDateRange(ctx, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 4. Lifestyle Record for this day
	lifestyle, _ := s.lifestyleRepo.FindByDate(ctx, userID, targetDate)

	// 5. Prediction for this day
	prediction, _ := s.predictionRepo.FindByDate(ctx, userID, targetDate)

	return &models.DailyAnalytics{
		Date:                   targetDate.Format("2006-01-02"),
		TotalStudySeconds:      studyStats.TotalCompletedSeconds,
		TotalStudyFormatted:    formatSeconds(studyStats.TotalCompletedSeconds),
		TotalFocusMinutes:      focusStats.CompletedFocusMinutes,
		TotalBreakMinutes:      focusStats.CompletedBreakMinutes,
		CompletedStudySessions: studyStats.CompletedSessionsCount,
		CompletedFocusCount:    focusStats.CompletedFocusCount,
		CompletedBreaksCount:   focusStats.CompletedBreaksCount,
		TotalInterruptions:     focusStats.TotalInterruptions,
		CompletedTasksCount:    tasksCount,
		Lifestyle:              lifestyle,
		LatestPrediction:       prediction,
	}, nil
}

func (s *analyticsService) GetWeeklyAnalytics(ctx context.Context, userID uuid.UUID, endDateStr *string, tzName *string) (*models.WeeklyAnalytics, error) {
	loc := parseUserLocation(tzName)

	var endDate time.Time
	if endDateStr != nil && *endDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *endDateStr, loc)
		if err != nil {
			return nil, err
		}
		endDate = parsed
	} else {
		now := time.Now().In(loc)
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	// 7 days window: [endDate - 6 days, endDate]
	startDate := endDate.AddDate(0, 0, -6)
	startWindowUTC := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc).UTC()
	endWindowUTC := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc).UTC()

	return s.aggregateTrend(ctx, userID, startDate, endDate, startWindowUTC, endWindowUTC, loc)
}

func (s *analyticsService) GetMonthlyAnalytics(ctx context.Context, userID uuid.UUID, endDateStr *string, tzName *string) (*models.MonthlyAnalytics, error) {
	loc := parseUserLocation(tzName)

	var endDate time.Time
	if endDateStr != nil && *endDateStr != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *endDateStr, loc)
		if err != nil {
			return nil, err
		}
		endDate = parsed
	} else {
		now := time.Now().In(loc)
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	// 30 days window: [endDate - 29 days, endDate]
	startDate := endDate.AddDate(0, 0, -29)
	startWindowUTC := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc).UTC()
	endWindowUTC := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc).UTC()

	weekly, err := s.aggregateTrend(ctx, userID, startDate, endDate, startWindowUTC, endWindowUTC, loc)
	if err != nil {
		return nil, err
	}

	return &models.MonthlyAnalytics{
		StartDate:           weekly.StartDate,
		EndDate:             weekly.EndDate,
		TotalStudySeconds:   weekly.TotalStudySeconds,
		TotalStudyFormatted: weekly.TotalStudyFormatted,
		TotalFocusMinutes:   weekly.TotalFocusMinutes,
		TotalBreakMinutes:   weekly.TotalBreakMinutes,
		TotalCompletedTasks: weekly.TotalCompletedTasks,
		StudyConsistencyPct: weekly.StudyConsistencyPct,
		DailyPoints:         weekly.DailyPoints,
		AverageSleepHours:   weekly.AverageSleepHours,
		AveragePhoneHours:   weekly.AveragePhoneHours,
		AverageStressLevel:  weekly.AverageStressLevel,
	}, nil
}

func (s *analyticsService) GetRangeAnalytics(ctx context.Context, userID uuid.UUID, startDateStr, endDateStr string, tzName *string) (*models.WeeklyAnalytics, error) {
	loc := parseUserLocation(tzName)

	if startDateStr == "" || endDateStr == "" {
		return nil, errors.New("both start_date and end_date are required (YYYY-MM-DD)")
	}

	startDate, err := time.ParseInLocation("2006-01-02", startDateStr, loc)
	if err != nil {
		return nil, errors.New("invalid start_date format, expected YYYY-MM-DD")
	}

	endDate, err := time.ParseInLocation("2006-01-02", endDateStr, loc)
	if err != nil {
		return nil, errors.New("invalid end_date format, expected YYYY-MM-DD")
	}

	if startDate.After(endDate) {
		return nil, errors.New("start_date cannot be after end_date")
	}

	// Safety cap: maximum 90 days range to prevent denial of service
	if endDate.Sub(startDate) > 90*24*time.Hour {
		return nil, errors.New("date range cannot exceed 90 days")
	}

	startWindowUTC := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, loc).UTC()
	endWindowUTC := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, loc).UTC()

	return s.aggregateTrend(ctx, userID, startDate, endDate, startWindowUTC, endWindowUTC, loc)
}

// aggregateTrend builds a unified daily points series across a date range.
func (s *analyticsService) aggregateTrend(
	ctx context.Context,
	userID uuid.UUID,
	startDate, endDate time.Time,
	startUTC, endUTC time.Time,
	loc *time.Location,
) (*models.WeeklyAnalytics, error) {
	// 1. Fetch completed study sessions in window
	studySessions, err := s.studyRepo.FindCompletedSessions(ctx, userID, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	// 2. Fetch completed focus intervals in window
	focusIntervals, err := s.focusRepo.FindCompletedIntervals(ctx, userID, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	// 3. Fetch completed tasks in window
	tasks, err := s.taskRepo.FindCompletedTasksByDateRange(ctx, userID, startUTC, endUTC)
	if err != nil {
		return nil, err
	}

	// 4. Fetch lifestyle records in window
	lifestyleRecords, err := s.lifestyleRepo.FindByDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Index lifestyle records by date string
	lifestyleMap := make(map[string]models.LifestyleRecord)
	for _, rec := range lifestyleRecords {
		dateKey := rec.RecordDate.Format("2006-01-02")
		lifestyleMap[dateKey] = rec
	}

	// Group study seconds by day
	studySecByDay := make(map[string]int64)
	for _, ss := range studySessions {
		dayKey := ss.StartedAt.In(loc).Format("2006-01-02")
		studySecByDay[dayKey] += int64(ss.DurationSeconds)
	}

	// Group focus & break minutes by day
	focusMinsByDay := make(map[string]int64)
	breakMinsByDay := make(map[string]int64)
	for _, fs := range focusIntervals {
		dayKey := fs.StartedAt.In(loc).Format("2006-01-02")
		if fs.SessionType == models.FocusTypeWork {
			focusMinsByDay[dayKey] += int64(fs.DurationMinutes)
		} else if fs.SessionType == models.FocusTypeBreak {
			breakMinsByDay[dayKey] += int64(fs.DurationMinutes)
		}
	}

	// Group tasks by completion day
	tasksByDay := make(map[string]int64)
	for _, t := range tasks {
		if t.CompletedAt != nil {
			dayKey := t.CompletedAt.In(loc).Format("2006-01-02")
			tasksByDay[dayKey]++
		}
	}

	// Build continuous sequence of days
	var dailyPoints []models.DailyMetricPoint
	var totalStudySec int64
	var totalFocusMins int64
	var totalBreakMins int64
	var totalTasksCount int64
	var activeStudyDays int64

	var sleepSum float64
	var sleepCount int
	var phoneSum float64
	var phoneCount int
	var stressSum int
	var stressCount int

	curr := startDate
	totalDays := 0
	for !curr.After(endDate) {
		totalDays++
		dayStr := curr.Format("2006-01-02")
		studySec := studySecByDay[dayStr]
		focusMins := focusMinsByDay[dayStr]
		breakMins := breakMinsByDay[dayStr]
		dayTasks := tasksByDay[dayStr]

		totalStudySec += studySec
		totalFocusMins += focusMins
		totalBreakMins += breakMins
		totalTasksCount += dayTasks

		if studySec > 0 || focusMins > 0 {
			activeStudyDays++
		}

		dp := models.DailyMetricPoint{
			Date:                dayStr,
			StudyDurationHours:  math.Round((float64(studySec)/3600.0)*100) / 100,
			FocusDurationHours:  math.Round((float64(focusMins)/60.0)*100) / 100,
			BreakDurationHours:  math.Round((float64(breakMins)/60.0)*100) / 100,
			CompletedTasksCount: dayTasks,
		}

		if lf, exists := lifestyleMap[dayStr]; exists {
			if lf.SleepHours != nil {
				dp.SleepHours = lf.SleepHours
				sleepSum += *lf.SleepHours
				sleepCount++
			}
			if lf.PhoneUsageHours != nil {
				dp.PhoneUsageHours = lf.PhoneUsageHours
				phoneSum += *lf.PhoneUsageHours
				phoneCount++
			}
			if lf.StressLevel != nil {
				dp.StressLevel = lf.StressLevel
				stressSum += *lf.StressLevel
				stressCount++
			}
		}

		dailyPoints = append(dailyPoints, dp)
		curr = curr.AddDate(0, 0, 1)
	}

	var consistencyPct float64
	if totalDays > 0 {
		consistencyPct = math.Round((float64(activeStudyDays)/float64(totalDays))*1000) / 10
	}

	var avgSleep *float64
	if sleepCount > 0 {
		val := math.Round((sleepSum/float64(sleepCount))*100) / 100
		avgSleep = &val
	}

	var avgPhone *float64
	if phoneCount > 0 {
		val := math.Round((phoneSum/float64(phoneCount))*100) / 100
		avgPhone = &val
	}

	var avgStress *float64
	if stressCount > 0 {
		val := math.Round((float64(stressSum)/float64(stressCount))*10) / 10
		avgStress = &val
	}

	return &models.WeeklyAnalytics{
		StartDate:           startDate.Format("2006-01-02"),
		EndDate:             endDate.Format("2006-01-02"),
		TotalStudySeconds:   totalStudySec,
		TotalStudyFormatted: formatSeconds(totalStudySec),
		TotalFocusMinutes:   totalFocusMins,
		TotalBreakMinutes:   totalBreakMins,
		TotalCompletedTasks: totalTasksCount,
		StudyConsistencyPct: consistencyPct,
		DailyPoints:         dailyPoints,
		AverageSleepHours:   avgSleep,
		AveragePhoneHours:   avgPhone,
		AverageStressLevel:  avgStress,
	}, nil
}

func (s *analyticsService) GetAnalyticsSummary(ctx context.Context, userID uuid.UUID, tzName *string) (*models.AnalyticsSummary, error) {
	loc := parseUserLocation(tzName)
	now := time.Now().In(loc)

	todayDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	startTodayUTC := todayDate.UTC()
	endTodayUTC := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, loc).UTC()

	// Today study stats
	todayStudy, err := s.studyRepo.GetDailyStatistics(ctx, userID, startTodayUTC, endTodayUTC)
	if err != nil {
		return nil, err
	}

	// Today focus stats
	todayFocus, err := s.focusRepo.GetDailyStatistics(ctx, userID, startTodayUTC, endTodayUTC)
	if err != nil {
		return nil, err
	}

	// Today completed tasks
	todayTasks, err := s.taskRepo.CountCompletedByDateRange(ctx, userID, startTodayUTC, endTodayUTC)
	if err != nil {
		return nil, err
	}

	// 7-day rolling window
	startWeekDate := todayDate.AddDate(0, 0, -6)
	startWeekUTC := startWeekDate.UTC()
	weeklyTrend, err := s.aggregateTrend(ctx, userID, startWeekDate, todayDate, startWeekUTC, endTodayUTC, loc)
	if err != nil {
		return nil, err
	}

	// Lifestyle today check
	lifestyle, _ := s.lifestyleRepo.FindByDate(ctx, userID, todayDate)
	todayLifestyleLogged := (lifestyle != nil)

	// Latest prediction
	latestPred, _ := s.predictionRepo.FindLatest(ctx, userID)

	// Active sessions
	activeStudy, _ := s.studyRepo.FindActiveOrPausedSession(ctx, userID)
	activeFocus, _ := s.focusRepo.FindActiveInterval(ctx, userID)

	return &models.AnalyticsSummary{
		TodayStudySeconds:    todayStudy.TotalCompletedSeconds,
		TodayStudyFormatted:  formatSeconds(todayStudy.TotalCompletedSeconds),
		TodayFocusMinutes:    todayFocus.CompletedFocusMinutes,
		TodayCompletedTasks:  todayTasks,
		WeekStudySeconds:     weeklyTrend.TotalStudySeconds,
		WeekStudyFormatted:   weeklyTrend.TotalStudyFormatted,
		WeekFocusMinutes:     weeklyTrend.TotalFocusMinutes,
		WeekCompletedTasks:   weeklyTrend.TotalCompletedTasks,
		StudyConsistencyPct:  weeklyTrend.StudyConsistencyPct,
		LatestPrediction:     latestPred,
		TodayLifestyleLogged: todayLifestyleLogged,
		ActiveStudySession:   activeStudy,
		ActiveFocusSession:   activeFocus,
	}, nil
}
