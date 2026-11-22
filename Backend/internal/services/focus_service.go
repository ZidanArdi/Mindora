package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

var (
	ErrInvalidSessionType       = errors.New("invalid session type (must be 'focus' or 'break')")
	ErrActiveFocusIntervalExists = errors.New("an active focus or break interval is already in progress")
	ErrNoActiveIntervalFound     = errors.New("no active interval found to finish or abandon")
	ErrInvalidIntervalStatus     = errors.New("interval status does not permit this action")
)

type StartFocusInput struct {
	SessionType string     `json:"session_type"` // "focus" or "break"
	TaskID      *uuid.UUID `json:"task_id,omitempty"`
	TargetMinutes int      `json:"target_minutes,omitempty"`
}

type FocusService interface {
	StartInterval(ctx context.Context, userID uuid.UUID, input StartFocusInput) (*models.FocusSession, error)
	RecordInterruption(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error)
	FinishInterval(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error)
	AbandonInterval(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error)
	GetActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error)
	ListHistory(ctx context.Context, userID uuid.UUID, sessionType string, page, limit int) ([]models.FocusSession, int64, error)
	GetDailyStats(ctx context.Context, userID uuid.UUID, date time.Time) (*repositories.FocusDailyStats, error)
}

type focusService struct {
	focusRepo repositories.FocusRepository
	taskRepo  repositories.TaskRepository
}

func NewFocusService(focusRepo repositories.FocusRepository, taskRepo repositories.TaskRepository) FocusService {
	return &focusService{
		focusRepo: focusRepo,
		taskRepo:  taskRepo,
	}
}

func (s *focusService) StartInterval(ctx context.Context, userID uuid.UUID, input StartFocusInput) (*models.FocusSession, error) {
	sType := strings.ToLower(strings.TrimSpace(input.SessionType))
	if sType != models.FocusTypeWork && sType != models.FocusTypeBreak {
		return nil, ErrInvalidSessionType
	}

	// Check if active interval already exists
	active, err := s.focusRepo.FindActiveInterval(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrActiveFocusIntervalExists
	}

	// Validate task ownership if task is provided
	if input.TaskID != nil && *input.TaskID != uuid.Nil {
		_, err := s.taskRepo.FindByID(ctx, *input.TaskID, userID)
		if err != nil {
			return nil, ErrTaskOwnershipInvalid
		}
	}

	now := time.Now().UTC()
	interval := &models.FocusSession{
		ID:                uuid.New(),
		UserID:            userID,
		TaskID:            input.TaskID,
		SessionType:       sType,
		Status:            models.FocusStatusActive,
		StartedAt:         now,
		DurationMinutes:   0,
		InterruptionCount: 0,
	}

	if err := s.focusRepo.CreateInterval(ctx, interval); err != nil {
		return nil, err
	}
	return interval, nil
}

func (s *focusService) RecordInterruption(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error) {
	interval, err := s.focusRepo.FindIntervalByID(ctx, intervalID, userID)
	if err != nil {
		return nil, err
	}
	if interval.Status != models.FocusStatusActive {
		return nil, ErrInvalidIntervalStatus
	}

	interval.InterruptionCount++
	if err := s.focusRepo.UpdateInterval(ctx, interval); err != nil {
		return nil, err
	}
	return interval, nil
}

func (s *focusService) FinishInterval(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error) {
	interval, err := s.focusRepo.FindIntervalByID(ctx, intervalID, userID)
	if err != nil {
		return nil, err
	}
	if interval.Status != models.FocusStatusActive {
		return nil, ErrInvalidIntervalStatus
	}

	now := time.Now().UTC()
	elapsedSeconds := now.Sub(interval.StartedAt).Seconds()
	// Convert elapsed to minutes (round to at least 1 minute if ran > 30s)
	mins := int(elapsedSeconds / 60.0)
	if mins < 1 && elapsedSeconds >= 30.0 {
		mins = 1
	}

	interval.DurationMinutes = mins
	interval.Status = models.FocusStatusCompleted
	interval.EndedAt = &now

	if err := s.focusRepo.UpdateInterval(ctx, interval); err != nil {
		return nil, err
	}
	return interval, nil
}

func (s *focusService) AbandonInterval(ctx context.Context, intervalID uuid.UUID, userID uuid.UUID) (*models.FocusSession, error) {
	interval, err := s.focusRepo.FindIntervalByID(ctx, intervalID, userID)
	if err != nil {
		return nil, err
	}
	if interval.Status != models.FocusStatusActive {
		return nil, ErrInvalidIntervalStatus
	}

	now := time.Now().UTC()
	interval.Status = models.FocusStatusAbandoned
	interval.EndedAt = &now

	if err := s.focusRepo.UpdateInterval(ctx, interval); err != nil {
		return nil, err
	}
	return interval, nil
}

func (s *focusService) GetActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error) {
	return s.focusRepo.FindActiveInterval(ctx, userID)
}

func (s *focusService) ListHistory(ctx context.Context, userID uuid.UUID, sessionType string, page, limit int) ([]models.FocusSession, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.focusRepo.FindHistory(ctx, userID, sessionType, limit, offset)
}

func (s *focusService) GetDailyStats(ctx context.Context, userID uuid.UUID, date time.Time) (*repositories.FocusDailyStats, error) {
	year, month, day := date.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	end := time.Date(year, month, day, 23, 59, 59, 999999999, time.UTC)
	return s.focusRepo.GetDailyStatistics(ctx, userID, start, end)
}
