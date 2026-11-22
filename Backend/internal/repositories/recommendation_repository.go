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
	ErrRecommendationNotFound = errors.New("recommendation not found or unauthorized")
)

type RecommendationRepository interface {
	FindActive(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Recommendation, error)
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Recommendation, error)
	UpsertPending(ctx context.Context, rec *models.Recommendation) error
	UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string, dismissedUntil *time.Time) (*models.Recommendation, error)
	FindDismissedRecent(ctx context.Context, userID uuid.UUID, ruleID string, now time.Time) (*models.Recommendation, error)
	FindByEvaluationKeyAndStatus(ctx context.Context, userID uuid.UUID, evalKey string, status string) (*models.Recommendation, error)
	InvalidatePendingByRule(ctx context.Context, userID uuid.UUID, ruleID string) error
	InvalidatePendingByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	ExpirePastValid(ctx context.Context, userID uuid.UUID, now time.Time) error
}

type recommendationRepository struct {
	db *gorm.DB
}

func NewRecommendationRepository(db *gorm.DB) RecommendationRepository {
	return &recommendationRepository{db: db}
}

func (r *recommendationRepository) FindActive(ctx context.Context, userID uuid.UUID, now time.Time) ([]models.Recommendation, error) {
	var recs []models.Recommendation
	orderClause := "CASE priority WHEN 'critical' THEN 1 WHEN 'high' THEN 2 WHEN 'medium' THEN 3 WHEN 'low' THEN 4 ELSE 5 END ASC, created_at DESC"
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ? AND valid_until > ?", userID, models.RecStatusPending, now).
		Order(orderClause).
		Find(&recs).Error
	return recs, err
}

func (r *recommendationRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Recommendation, error) {
	var rec models.Recommendation
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRecommendationNotFound
		}
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) UpsertPending(ctx context.Context, rec *models.Recommendation) error {
	if rec.ID == uuid.Nil {
		rec.ID = uuid.New()
	}
	if rec.CreatedAt.IsZero() {
		rec.CreatedAt = time.Now().UTC()
	}
	if rec.UpdatedAt.IsZero() {
		rec.UpdatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO recommendations (
			id, user_id, rule_id, category, priority, title, message, explanation,
			evidence, action_route, action_label, status, evaluation_period_start,
			evaluation_period_end, evaluation_key, dismissed_until, valid_until,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8,
			$9, $10, $11, $12, $13,
			$14, $15, $16, $17,
			$18, $19
		)
		ON CONFLICT (user_id, rule_id) WHERE status = 'pending'
		DO UPDATE SET
			title = EXCLUDED.title,
			message = EXCLUDED.message,
			explanation = EXCLUDED.explanation,
			evidence = EXCLUDED.evidence,
			valid_until = EXCLUDED.valid_until,
			evaluation_period_start = EXCLUDED.evaluation_period_start,
			evaluation_period_end = EXCLUDED.evaluation_period_end,
			evaluation_key = EXCLUDED.evaluation_key,
			action_route = EXCLUDED.action_route,
			action_label = EXCLUDED.action_label,
			priority = EXCLUDED.priority,
			category = EXCLUDED.category,
			updated_at = NOW();
	`

	return r.db.WithContext(ctx).Exec(
		query,
		rec.ID, rec.UserID, rec.RuleID, rec.Category, rec.Priority, rec.Title, rec.Message, rec.Explanation,
		rec.Evidence, rec.ActionRoute, rec.ActionLabel, rec.Status, rec.EvaluationPeriodStart,
		rec.EvaluationPeriodEnd, rec.EvaluationKey, rec.DismissedUntil, rec.ValidUntil,
		rec.CreatedAt, rec.UpdatedAt,
	).Error
}

func (r *recommendationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string, dismissedUntil *time.Time) (*models.Recommendation, error) {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":          status,
		"dismissed_until": dismissedUntil,
		"updated_at":      now,
	}

	result := r.db.WithContext(ctx).Model(&models.Recommendation{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrRecommendationNotFound
	}

	return r.FindByID(ctx, id, userID)
}

func (r *recommendationRepository) FindDismissedRecent(ctx context.Context, userID uuid.UUID, ruleID string, now time.Time) (*models.Recommendation, error) {
	var rec models.Recommendation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND rule_id = ? AND status = ? AND dismissed_until > ?",
			userID, ruleID, models.RecStatusDismissed, now).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) FindByEvaluationKeyAndStatus(ctx context.Context, userID uuid.UUID, evalKey string, status string) (*models.Recommendation, error) {
	var rec models.Recommendation
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND evaluation_key = ? AND status = ?", userID, evalKey, status).
		First(&rec).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &rec, nil
}

func (r *recommendationRepository) InvalidatePendingByRule(ctx context.Context, userID uuid.UUID, ruleID string) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.Recommendation{}).
		Where("user_id = ? AND rule_id = ? AND status = ?", userID, ruleID, models.RecStatusPending).
		Updates(map[string]interface{}{
			"status":     models.RecStatusInvalidated,
			"updated_at": now,
		}).Error
}

func (r *recommendationRepository) InvalidatePendingByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	now := time.Now().UTC()
	return r.db.WithContext(ctx).Model(&models.Recommendation{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, models.RecStatusPending).
		Updates(map[string]interface{}{
			"status":     models.RecStatusInvalidated,
			"updated_at": now,
		}).Error
}

func (r *recommendationRepository) ExpirePastValid(ctx context.Context, userID uuid.UUID, now time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Recommendation{}).
		Where("user_id = ? AND status = ? AND valid_until <= ?", userID, models.RecStatusPending, now).
		Updates(map[string]interface{}{
			"status":     models.RecStatusExpired,
			"updated_at": now,
		}).Error
}
