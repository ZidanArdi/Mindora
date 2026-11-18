package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/models"
)

var (
	ErrIdentityNotFound = errors.New("auth identity not found")
)

type IdentityRepository interface {
	FindByIdentity(ctx context.Context, provider, subject string) (*models.AuthIdentity, error)
	Create(ctx context.Context, identity *models.AuthIdentity) error
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthIdentity, error)
}

type GormIdentityRepository struct {
	db *gorm.DB
}

func NewIdentityRepository(db *gorm.DB) *GormIdentityRepository {
	return &GormIdentityRepository{db: db}
}

func (r *GormIdentityRepository) FindByIdentity(ctx context.Context, provider, subject string) (*models.AuthIdentity, error) {
	var identity models.AuthIdentity
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("User.Role").
		Where("provider = ? AND provider_subject = ?", provider, subject).
		First(&identity).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrIdentityNotFound
		}
		return nil, err
	}
	return &identity, nil
}

func (r *GormIdentityRepository) Create(ctx context.Context, identity *models.AuthIdentity) error {
	return r.db.WithContext(ctx).Create(identity).Error
}

func (r *GormIdentityRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]models.AuthIdentity, error) {
	var identities []models.AuthIdentity
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&identities).Error
	return identities, err
}
