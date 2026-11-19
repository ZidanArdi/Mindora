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
	ErrFocusIntervalNotFound = errors.New("focus interval not found or unauthorized")
)

type FocusDailyStats struct {
	CompletedFocusMinutes int64 `json:"completed_focus_minutes"`
	CompletedFocusCount   int64 `json:"completed_focus_count"`
	CompletedBreaksCount  int64 `json:"completed_breaks_count"`
	TotalInterruptions    int64 `json:"total_interruptions"`
}

type FocusRepository interface {
	CreateInterval(ctx context.Context, interval *models.FocusSession) error
	FindIntervalByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.FocusSession, error)
	FindActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error)
	FindHistory(ctx context.Context, userID uuid.UUID, sessionType string, limit, offset int) ([]models.FocusSession, int64, error)
	UpdateInterval(ctx context.Context, interval *models.FocusSession) error
	GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*FocusDailyStats, error)
	CountCompletedBreaks(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error)
	GetCompletedFocusDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error)
}

type focusRepository struct {
	db *gorm.DB
}

func NewFocusRepository(db *gorm.DB) FocusRepository {
	return &focusRepository{db: db}
}

func (r *focusRepository) CreateInterval(ctx context.Context, interval *models.FocusSession) error {
	if interval.ID == uuid.Nil {
		interval.ID = uuid.New()
	}
	now := time.Now().UTC()
	interval.CreatedAt = now
	interval.UpdatedAt = now
	return r.db.WithContext(ctx).Create(interval).Error
}

func (r *focusRepository) FindIntervalByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.FocusSession, error) {
	var interval models.FocusSession
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("id = ? AND user_id = ?", id, userID).
		First(&interval).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrFocusIntervalNotFound
		}
		return nil, err
	}
	return &interval, nil
}

func (r *focusRepository) FindActiveInterval(ctx context.Context, userID uuid.UUID) (*models.FocusSession, error) {
	var interval models.FocusSession
	err := r.db.WithContext(ctx).
		Preload("Task").
		Where("user_id = ? AND status = ?", userID, models.FocusStatusActive).
		Order("started_at DESC").
		First(&interval).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // None currently active
		}
		return nil, err
	}
	return &interval, nil
}

func (r *focusRepository) FindHistory(ctx context.Context, userID uuid.UUID, sessionType string, limit, offset int) ([]models.FocusSession, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.FocusSession{}).Where("user_id = ?", userID)

	if sessionType != "" {
		query = query.Where("session_type = ?", sessionType)
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

	var intervals []models.FocusSession
	if err := query.Find(&intervals).Error; err != nil {
		return nil, 0, err
	}

	return intervals, total, nil
}

func (r *focusRepository) UpdateInterval(ctx context.Context, interval *models.FocusSession) error {
	interval.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.FocusSession{}).
		Where("id = ? AND user_id = ?", interval.ID, interval.UserID).
		Updates(map[string]interface{}{
			"duration_minutes":   interval.DurationMinutes,
			"status":             interval.Status,
			"interruption_count": interval.InterruptionCount,
			"ended_at":           interval.EndedAt,
			"updated_at":         interval.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrFocusIntervalNotFound
	}
	return nil
}

func (r *focusRepository) GetDailyStatistics(ctx context.Context, userID uuid.UUID, start, end time.Time) (*FocusDailyStats, error) {
	type aggRow struct {
		SessionType string
		TotalMins   int64
		Count       int64
		TotalInterr int64
	}
	var rows []aggRow
	err := r.db.WithContext(ctx).Model(&models.FocusSession{}).
		Select("session_type, COALESCE(SUM(duration_minutes), 0) as total_mins, count(*) as count, COALESCE(SUM(interruption_count), 0) as total_interr").
		Where("user_id = ? AND status = ? AND started_at >= ? AND started_at <= ?", userID, models.FocusStatusCompleted, start, end).
		Group("session_type").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	stats := &FocusDailyStats{}
	for _, row := range rows {
		if row.SessionType == models.FocusTypeWork {
			stats.CompletedFocusMinutes = row.TotalMins
			stats.CompletedFocusCount = row.Count
			stats.TotalInterruptions += row.TotalInterr
		} else if row.SessionType == models.FocusTypeBreak {
			stats.CompletedBreaksCount = row.Count
			stats.TotalInterruptions += row.TotalInterr
		}
	}
	return stats, nil
}

func (r *focusRepository) CountCompletedBreaks(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.FocusSession{}).
		Where("user_id = ? AND session_type = ? AND status = ? AND started_at >= ? AND started_at <= ?",
			userID, models.FocusTypeBreak, models.FocusStatusCompleted, start, end).
		Count(&count).Error
	return count, err
}

func (r *focusRepository) GetCompletedFocusDuration(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	var totalMins *int64
	err := r.db.WithContext(ctx).Model(&models.FocusSession{}).
		Select("COALESCE(SUM(duration_minutes), 0)").
		Where("user_id = ? AND session_type = ? AND status = ? AND started_at >= ? AND started_at <= ?",
			userID, models.FocusTypeWork, models.FocusStatusCompleted, start, end).
		Scan(&totalMins).Error
	if err != nil {
		return 0, err
	}
	if totalMins == nil {
		return 0, nil
	}
	return *totalMins, nil
}
