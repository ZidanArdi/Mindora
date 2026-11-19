package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mindora/backend/internal/models"
)

var (
	ErrLifestyleRecordNotFound = errors.New("lifestyle record not found for specified date")
)

type LifestyleRepository interface {
	FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error)
	UpsertDailyRecord(ctx context.Context, record *models.LifestyleRecord) error
	FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.LifestyleRecord, int64, error)
	FindByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.LifestyleRecord, error)
}

type lifestyleRepository struct {
	db *gorm.DB
}

func NewLifestyleRepository(db *gorm.DB) LifestyleRepository {
	return &lifestyleRepository{db: db}
}

func (r *lifestyleRepository) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error) {
	var record models.LifestyleRecord
	dateStr := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND record_date = ?", userID, dateStr).
		First(&record).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrLifestyleRecordNotFound
		}
		return nil, err
	}
	return &record, nil
}

func (r *lifestyleRepository) UpsertDailyRecord(ctx context.Context, record *models.LifestyleRecord) error {
	if record.ID == uuid.Nil {
		record.ID = uuid.New()
	}
	now := time.Now().UTC()
	record.CreatedAt = now
	record.UpdatedAt = now

	// Postgres ON CONFLICT (user_id, record_date) DO UPDATE
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "user_id"},
			{Name: "record_date"},
		},
		DoUpdates: clause.AssignmentColumns([]string{
			"sleep_hours",
			"phone_usage_hours",
			"social_media_hours",
			"youtube_hours",
			"gaming_hours",
			"coffee_intake_mg",
			"exercise_minutes",
			"stress_level",
			"notes",
			"updated_at",
		}),
	}).Create(record).Error
}

func (r *lifestyleRepository) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.LifestyleRecord, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.LifestyleRecord{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("record_date DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var records []models.LifestyleRecord
	if err := query.Find(&records).Error; err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

func (r *lifestyleRepository) FindByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.LifestyleRecord, error) {
	var records []models.LifestyleRecord
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND record_date >= ? AND record_date <= ?", userID, start.Format("2006-01-02"), end.Format("2006-01-02")).
		Order("record_date ASC").
		Find(&records).Error
	return records, err
}
