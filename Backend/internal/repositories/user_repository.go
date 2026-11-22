package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/models"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrRoleNotFound = errors.New("role not found")
)

type UserRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	Update(ctx context.Context, user *models.User) error
	GetRoleByName(ctx context.Context, name string) (*models.Role, error)
	GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
	List(ctx context.Context, limit, offset int) ([]models.User, int64, error)
	Transaction(ctx context.Context, fn func(txRepo UserRepository) error) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Preload("Role").Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	err := r.db.WithContext(ctx).Preload("Role").Where("LOWER(email) = ?", normalizedEmail).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *GormUserRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	var role models.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRoleNotFound
		}
		return nil, err
	}
	return &role, nil
}

func (r *GormUserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *GormUserRepository) GetUserPermissions(ctx context.Context, roleID uuid.UUID) ([]string, error) {
	var permissions []string
	err := r.db.WithContext(ctx).
		Table("permissions").
		Select("permissions.resource || '.' || permissions.action").
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Order("permissions.resource, permissions.action").
		Pluck("permissions.resource || '.' || permissions.action", &permissions).Error
	if err != nil {
		return nil, err
	}
	return permissions, nil
}

func (r *GormUserRepository) List(ctx context.Context, limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	tx := r.db.WithContext(ctx).Model(&models.User{})
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := tx.Preload("Role").Limit(limit).Offset(offset).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *GormUserRepository) Transaction(ctx context.Context, fn func(txRepo UserRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := NewUserRepository(tx)
		return fn(txRepo)
	})
}
