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
	ErrStudySubjectRequired      = errors.New("study subject is required")
	ErrActiveStudySessionExists   = errors.New("an active study session is already in progress")
	ErrInvalidSessionTransition   = errors.New("invalid study session status transition")
	ErrTaskOwnershipInvalid       = errors.New("referenced task does not belong to the user")
)

type StartStudyInput struct {
	Subject string     `json:"subject"`
	TaskID  *uuid.UUID `json:"task_id,omitempty"`
}

type StudyService interface {
	StartSession(ctx context.Context, userID uuid.UUID, input StartStudyInput) (*models.StudySession, error)
	PauseSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error)
	ResumeSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error)
	FinishSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error)
	CancelSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error)
	GetActiveSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error)
	ListSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]models.StudySession, int64, error)
	GetDailyStats(ctx context.Context, userID uuid.UUID, date time.Time) (*repositories.StudyDailyStats, error)
}

type studyService struct {
	studyRepo repositories.StudyRepository
	taskRepo  repositories.TaskRepository
}

func NewStudyService(studyRepo repositories.StudyRepository, taskRepo repositories.TaskRepository) StudyService {
	return &studyService{
		studyRepo: studyRepo,
		taskRepo:  taskRepo,
	}
}

func (s *studyService) StartSession(ctx context.Context, userID uuid.UUID, input StartStudyInput) (*models.StudySession, error) {
	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		return nil, ErrStudySubjectRequired
	}

	// Guard: check if an active or paused session already exists
	active, err := s.studyRepo.FindActiveOrPausedSession(ctx, userID)
	if err != nil {
		return nil, err
	}
	if active != nil {
		return nil, ErrActiveStudySessionExists
	}

	// Validate task ownership if task is specified
	if input.TaskID != nil && *input.TaskID != uuid.Nil {
		_, err := s.taskRepo.FindByID(ctx, *input.TaskID, userID)
		if err != nil {
			return nil, ErrTaskOwnershipInvalid
		}
	}

	now := time.Now().UTC()
	session := &models.StudySession{
		ID:              uuid.New(),
		UserID:          userID,
		TaskID:          input.TaskID,
		Subject:         subject,
		DurationSeconds: 0,
		Status:          models.StudyStatusActive,
		StartedAt:       now,
	}

	if err := s.studyRepo.CreateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *studyService) PauseSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error) {
	session, err := s.studyRepo.FindSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Status != models.StudyStatusActive {
		return nil, ErrInvalidSessionTransition
	}

	// Add elapsed seconds since last update / start
	now := time.Now().UTC()
	elapsed := int(now.Sub(session.UpdatedAt).Seconds())
	if elapsed > 0 {
		session.DurationSeconds += elapsed
	}

	session.Status = models.StudyStatusPaused
	if err := s.studyRepo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *studyService) ResumeSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error) {
	session, err := s.studyRepo.FindSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Status != models.StudyStatusPaused {
		return nil, ErrInvalidSessionTransition
	}

	session.Status = models.StudyStatusActive
	if err := s.studyRepo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *studyService) FinishSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error) {
	session, err := s.studyRepo.FindSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}
	if session.Status != models.StudyStatusActive && session.Status != models.StudyStatusPaused {
		return nil, ErrInvalidSessionTransition
	}

	now := time.Now().UTC()
	if session.Status == models.StudyStatusActive {
		elapsed := int(now.Sub(session.UpdatedAt).Seconds())
		if elapsed > 0 {
			session.DurationSeconds += elapsed
		}
	}

	// Ensure minimum non-negative duration
	if session.DurationSeconds < 0 {
		session.DurationSeconds = 0
	}

	session.Status = models.StudyStatusCompleted
	session.EndedAt = &now

	if err := s.studyRepo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *studyService) CancelSession(ctx context.Context, sessionID uuid.UUID, userID uuid.UUID) (*models.StudySession, error) {
	session, err := s.studyRepo.FindSessionByID(ctx, sessionID, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	session.Status = models.StudyStatusCancelled
	session.EndedAt = &now

	if err := s.studyRepo.UpdateSession(ctx, session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *studyService) GetActiveSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error) {
	return s.studyRepo.FindActiveOrPausedSession(ctx, userID)
}

func (s *studyService) ListSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, page, limit int) ([]models.StudySession, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.studyRepo.FindSessions(ctx, userID, startDate, endDate, limit, offset)
}

func (s *studyService) GetDailyStats(ctx context.Context, userID uuid.UUID, date time.Time) (*repositories.StudyDailyStats, error) {
	// Midnight to end of day in local/provided timezone, converted to UTC window
	loc := date.Location()
	year, month, day := date.Date()
	start := time.Date(year, month, day, 0, 0, 0, 0, loc).UTC()
	end := time.Date(year, month, day, 23, 59, 59, 999999999, loc).UTC()
	return s.studyRepo.GetDailyStatistics(ctx, userID, start, end)
}
