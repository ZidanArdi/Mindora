package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/models"
)

var (
	ErrStudySessionNotFound = errors.New("study session not found or unauthorized")
)

type StudyDailyStats struct {
	TotalCompletedSeconds int64 `json:"total_completed_seconds"`
	CompletedSessionsCount int64 `json:"completed_sessions_count"`
}

type StudyRepository interface {
	CreateSession(ctx context.Context, session *models.StudySession) error
	FindSessionByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.StudySession, error)
	FindActiveOrPausedSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error)
	FindSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.StudySession, int64, error)
	UpdateSession(ctx context.Context, session *models.StudySession) error
	GetCompletedDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error)
	GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*StudyDailyStats, error)
}

type studyRepository struct {
	db *gorm.DB
}

func NewStudyRepository(db *gorm.DB) StudyRepository {
	return &studyRepository{db: db}
}

func (r *studyRepository) CreateSession(ctx context.Context, session *models.StudySession) error {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	now := time.Now().UTC()
	session.CreatedAt = now
	session.UpdatedAt = now
	return r.db.WithContext(ctx).Create(session).Error
}

func (r *studyRepository) FindSessionByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.StudySession, error) {
	var session models.StudySession
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("id = ? AND user_id = ?", id, userID).
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStudySessionNotFound
		}
		return nil, err
	}
	return &session, nil
}

func (r *studyRepository) FindActiveOrPausedSession(ctx context.Context, userID uuid.UUID) (*models.StudySession, error) {
	var session models.StudySession
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("user_id = ? AND status IN (?, ?)", userID, models.StudyStatusActive, models.StudyStatusPaused).
		Order("started_at DESC").
		First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // None currently active
		}
		return nil, err
	}
	return &session, nil
}

func (r *studyRepository) FindSessions(ctx context.Context, userID uuid.UUID, startDate, endDate *time.Time, limit, offset int) ([]models.StudySession, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.StudySession{}).Where("user_id = ?", userID)

	if startDate != nil {
		query = query.Where("started_at >= ?", *startDate)
	}
	if endDate != nil {
		query = query.Where("started_at <= ?", *endDate)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("started_at DESC").Preload("Task")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var sessions []models.StudySession
	if err := query.Find(&sessions).Error; err != nil {
		return nil, 0, err
	}

	return sessions, total, nil
}

func (r *studyRepository) UpdateSession(ctx context.Context, session *models.StudySession) error {
	session.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.StudySession{}).
		Where("id = ? AND user_id = ?", session.ID, session.UserID).
		Updates(map[string]interface{}{
			"duration_seconds": session.DurationSeconds,
			"status":           session.Status,
			"ended_at":         session.EndedAt,
			"updated_at":       session.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrStudySessionNotFound
	}
	return nil
}

func (r *studyRepository) GetCompletedDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	var totalSeconds *int64
	err := r.db.WithContext(ctx).Model(&models.StudySession{}).
		Select("COALESCE(SUM(duration_seconds), 0)").
		Where("user_id = ? AND status = ? AND started_at >= ? AND started_at <= ?", userID, models.StudyStatusCompleted, start, end).
		Scan(&totalSeconds).Error
	if err != nil {
		return 0, err
	}
	if totalSeconds == nil {
		return 0, nil
	}
	return *totalSeconds, nil
}

func (r *studyRepository) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*StudyDailyStats, error) {
	type aggResult struct {
		TotalSec int64
		Count    int64
	}
	var res aggResult
	err := r.db.WithContext(ctx).Model(&models.StudySession{}).
		Select("COALESCE(SUM(duration_seconds), 0) as total_sec, count(*) as count").
		Where("user_id = ? AND status = ? AND started_at >= ? AND started_at <= ?", userID, models.StudyStatusCompleted, start, end).
		Scan(&res).Error
	if err != nil {
		return nil, err
	}
	return &StudyDailyStats{
		TotalCompletedSeconds: res.TotalSec,
		CompletedSessionsCount: res.Count,
	}, nil
}
