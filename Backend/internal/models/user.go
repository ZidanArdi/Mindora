package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User status enum constants
const (
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
)

// User represents a registered application user account.
type User struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	RoleID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"role_id"`
	Name            string         `gorm:"type:varchar(100);not null" json:"name"`
	Email           string         `gorm:"type:varchar(255);not null" json:"email"`
	PasswordHash    *string        `gorm:"type:text;column:password_hash" json:"-"`
	Status          string         `gorm:"type:varchar(20);not null;default:'active'" json:"status"`
	EmailVerifiedAt *time.Time     `gorm:"type:timestamptz" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"type:timestamptz;index" json:"-"`

	Role         Role           `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:RESTRICT" json:"role,omitempty"`
	AuthSessions []AuthSession  `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
	Identities   []AuthIdentity `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT" json:"identities,omitempty"`
}

func (User) TableName() string {
	return "users"
}
