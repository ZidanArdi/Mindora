package models

// DailyAnalytics represents activity and lifestyle records for a single calendar date.
type DailyAnalytics struct {
	Date                   string           `json:"date"` // YYYY-MM-DD
	TotalStudySeconds      int64            `json:"total_study_seconds"`
	TotalStudyFormatted    string           `json:"total_study_formatted"`
	TotalFocusMinutes      int64            `json:"total_focus_minutes"`
	TotalBreakMinutes      int64            `json:"total_break_minutes"`
	CompletedStudySessions int64            `json:"completed_study_sessions"`
	CompletedFocusCount    int64            `json:"completed_focus_count"`
	CompletedBreaksCount   int64            `json:"completed_breaks_count"`
	TotalInterruptions     int64            `json:"total_interruptions"`
	CompletedTasksCount    int64            `json:"completed_tasks_count"`
	Lifestyle              *LifestyleRecord `json:"lifestyle,omitempty"`
	LatestPrediction       *ProductivityPrediction `json:"latest_prediction,omitempty"`
}

// DailyMetricPoint represents a day-by-day datapoint in a trend line.
type DailyMetricPoint struct {
	Date                string   `json:"date"` // YYYY-MM-DD
	StudyDurationHours  float64  `json:"study_duration_hours"`
	FocusDurationHours  float64  `json:"focus_duration_hours"`
	BreakDurationHours  float64  `json:"break_duration_hours"`
	CompletedTasksCount int64    `json:"completed_tasks_count"`
	SleepHours          *float64 `json:"sleep_hours,omitempty"`
	PhoneUsageHours     *float64 `json:"phone_usage_hours,omitempty"`
	StressLevel         *int     `json:"stress_level,omitempty"`
}

// WeeklyAnalytics represents a 7-day rolling or calendar week analytics aggregation.
type WeeklyAnalytics struct {
	StartDate           string             `json:"start_date"` // YYYY-MM-DD
	EndDate             string             `json:"end_date"`   // YYYY-MM-DD
	TotalStudySeconds   int64              `json:"total_study_seconds"`
	TotalStudyFormatted string             `json:"total_study_formatted"`
	TotalFocusMinutes   int64              `json:"total_focus_minutes"`
	TotalBreakMinutes   int64              `json:"total_break_minutes"`
	TotalCompletedTasks int64              `json:"total_completed_tasks"`
	StudyConsistencyPct float64            `json:"study_consistency_percentage"` // e.g. 5/7 days = 71.4%
	DailyPoints         []DailyMetricPoint `json:"daily_points"`
	AverageSleepHours   *float64           `json:"average_sleep_hours,omitempty"`
	AveragePhoneHours   *float64           `json:"average_phone_usage_hours,omitempty"`
	AverageStressLevel  *float64           `json:"average_stress_level,omitempty"`
}

// MonthlyAnalytics represents a 30-day analytics overview.
type MonthlyAnalytics struct {
	StartDate           string             `json:"start_date"`
	EndDate             string             `json:"end_date"`
	TotalStudySeconds   int64              `json:"total_study_seconds"`
	TotalStudyFormatted string             `json:"total_study_formatted"`
	TotalFocusMinutes   int64              `json:"total_focus_minutes"`
	TotalBreakMinutes   int64              `json:"total_break_minutes"`
	TotalCompletedTasks int64              `json:"total_completed_tasks"`
	StudyConsistencyPct float64            `json:"study_consistency_percentage"` // days with >= 1 completed session / total days
	DailyPoints         []DailyMetricPoint `json:"daily_points"`
	AverageSleepHours   *float64           `json:"average_sleep_hours,omitempty"`
	AveragePhoneHours   *float64           `json:"average_phone_usage_hours,omitempty"`
	AverageStressLevel  *float64           `json:"average_stress_level,omitempty"`
}

// AnalyticsSummary provides high-level KPIs across the user's recent productivity history.
type AnalyticsSummary struct {
	TodayStudySeconds     int64                   `json:"today_study_seconds"`
	TodayStudyFormatted   string                  `json:"today_study_formatted"`
	TodayFocusMinutes     int64                   `json:"today_focus_minutes"`
	TodayCompletedTasks   int64                   `json:"today_completed_tasks"`
	WeekStudySeconds      int64                   `json:"week_study_seconds"`
	WeekStudyFormatted    string                  `json:"week_study_formatted"`
	WeekFocusMinutes      int64                   `json:"week_focus_minutes"`
	WeekCompletedTasks    int64                   `json:"week_completed_tasks"`
	StudyConsistencyPct   float64                 `json:"study_consistency_percentage"`
	LatestPrediction      *ProductivityPrediction `json:"latest_prediction,omitempty"`
	TodayLifestyleLogged  bool                    `json:"today_lifestyle_logged"`
	ActiveStudySession    *StudySession           `json:"active_study_session,omitempty"`
	ActiveFocusSession    *FocusSession           `json:"active_focus_session,omitempty"`
}
