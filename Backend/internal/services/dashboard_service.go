package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

type DashboardSummary struct {
	TotalTasks              int64                   `json:"total_tasks"`
	CompletedTasks          int64                   `json:"completed_tasks"`
	PendingTasks            int64                   `json:"pending_tasks"`
	InProgressTasks         int64                   `json:"in_progress_tasks"`
	CompletionPercentage    float64                 `json:"completion_percentage"`
	TotalStudySeconds       int64                   `json:"total_study_seconds"`
	TotalStudyFormatted     string                  `json:"total_study_formatted"`
	CompletedStudySessions  int64                   `json:"completed_study_sessions"`
	CompletedFocusMinutes   int64                   `json:"completed_focus_minutes"`
	CompletedFocusCount     int64                   `json:"completed_focus_count"`
	CompletedBreaksCount    int64                   `json:"completed_breaks_count"`
	TotalInterruptions      int64                   `json:"total_interruptions"`
	RecentTasks             []models.Task           `json:"recent_tasks"`
	UpcomingDeadlines       []models.Task           `json:"upcoming_deadlines"`
	TodayLifestyleLogged    bool                    `json:"today_lifestyle_logged"`
	TodayLifestyle          *models.LifestyleRecord `json:"today_lifestyle,omitempty"`
}

type DashboardService interface {
	GetSummary(ctx context.Context, userID uuid.UUID) (*DashboardSummary, error)
}

type dashboardService struct {
	taskRepo      repositories.TaskRepository
	studyRepo     repositories.StudyRepository
	focusRepo     repositories.FocusRepository
	lifestyleRepo repositories.LifestyleRepository
}

func NewDashboardService(
	taskRepo repositories.TaskRepository,
	studyRepo repositories.StudyRepository,
	focusRepo repositories.FocusRepository,
	lifestyleRepo repositories.LifestyleRepository,
) DashboardService {
	return &dashboardService{
		taskRepo:      taskRepo,
		studyRepo:     studyRepo,
		focusRepo:     focusRepo,
		lifestyleRepo: lifestyleRepo,
	}
}

func formatSeconds(sec int64) string {
	hours := sec / 3600
	mins := (sec % 3600) / 60
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}

func (s *dashboardService) GetSummary(ctx context.Context, userID uuid.UUID) (*DashboardSummary, error) {
	// 1. Task counts
	taskCounts, err := s.taskRepo.CountByStatus(ctx, userID)
	if err != nil {
		return nil, err
	}

	totalTasks := taskCounts[models.TaskStatusPending] +
		taskCounts[models.TaskStatusInProgress] +
		taskCounts[models.TaskStatusCompleted] +
		taskCounts[models.TaskStatusCancelled]
	completedTasks := taskCounts[models.TaskStatusCompleted]
	pendingTasks := taskCounts[models.TaskStatusPending]
	inProgressTasks := taskCounts[models.TaskStatusInProgress]

	var completionPercentage float64
	if totalTasks > 0 {
		completionPercentage = float64(completedTasks) / float64(totalTasks) * 100.0
	}

	// 2. Study stats for today
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	studyStats, err := s.studyRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 3. Focus stats for today
	focusStats, err := s.focusRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	if err != nil {
		return nil, err
	}

	// 4. Recent tasks
	recentTasks, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		SortBy:    "created_at",
		SortOrder: "desc",
		Limit:     5,
	})
	if err != nil {
		return nil, err
	}

	// 5. Upcoming deadlines
	upcomingDeadlines, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Status:    models.TaskStatusPending,
		SortBy:    "deadline",
		SortOrder: "asc",
		Limit:     5,
	})
	if err != nil {
		return nil, err
	}

	// 6. Today's lifestyle record
	todayLifestyle, err := s.lifestyleRepo.FindByDate(ctx, userID, now)
	todayLogged := (err == nil && todayLifestyle != nil)

	return &DashboardSummary{
		TotalTasks:             totalTasks,
		CompletedTasks:         completedTasks,
		PendingTasks:           pendingTasks,
		InProgressTasks:        inProgressTasks,
		CompletionPercentage:   completionPercentage,
		TotalStudySeconds:      studyStats.TotalCompletedSeconds,
		TotalStudyFormatted:    formatSeconds(studyStats.TotalCompletedSeconds),
		CompletedStudySessions: studyStats.CompletedSessionsCount,
		CompletedFocusMinutes:  focusStats.CompletedFocusMinutes,
		CompletedFocusCount:    focusStats.CompletedFocusCount,
		CompletedBreaksCount:   focusStats.CompletedBreaksCount,
		TotalInterruptions:     focusStats.TotalInterruptions,
		RecentTasks:            recentTasks,
		UpcomingDeadlines:      upcomingDeadlines,
		TodayLifestyleLogged:   todayLogged,
		TodayLifestyle:         todayLifestyle,
	}, nil
}
