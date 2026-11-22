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
	ErrPredictionNotFound = errors.New("prediction record not found or unauthorized")
)

type PredictionRepository interface {
	Create(ctx context.Context, prediction *models.ProductivityPrediction) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ProductivityPrediction, error)
	FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ProductivityPrediction, int64, error)
	FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error)
	FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.ProductivityPrediction, error)
}

type predictionRepository struct {
	db *gorm.DB
}

func NewPredictionRepository(db *gorm.DB) PredictionRepository {
	return &predictionRepository{db: db}
}

func (r *predictionRepository) Create(ctx context.Context, prediction *models.ProductivityPrediction) error {
	if prediction.ID == uuid.Nil {
		prediction.ID = uuid.New()
	}
	if prediction.CreatedAt.IsZero() {
		prediction.CreatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Create(prediction).Error
}

func (r *predictionRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	var prediction models.ProductivityPrediction
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&prediction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPredictionNotFound
		}
		return nil, err
	}
	return &prediction, nil
}

func (r *predictionRepository) FindHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]models.ProductivityPrediction, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.ProductivityPrediction{}).Where("user_id = ?", userID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query = query.Order("created_at DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	var results []models.ProductivityPrediction
	if err := query.Find(&results).Error; err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *predictionRepository) FindLatest(ctx context.Context, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	var prediction models.ProductivityPrediction
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&prediction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // None recorded yet
		}
		return nil, err
	}
	return &prediction, nil
}

func (r *predictionRepository) FindByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.ProductivityPrediction, error) {
	var prediction models.ProductivityPrediction
	dateStr := date.Format("2006-01-02")
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND prediction_date = ?", userID, dateStr).
		Order("created_at DESC").
		First(&prediction).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &prediction, nil
}
