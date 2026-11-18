package models

import (
	"time"

	"github.com/google/uuid"
)

// Role represents an authorization role in the application.
// Only 'user' and 'admin' are persisted system roles. Guest is unauthenticated.
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string       `gorm:"type:varchar(30);not null;uniqueIndex" json:"name"`
	Description *string      `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time    `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt   time.Time    `gorm:"type:timestamptz;not null" json:"updated_at"`
	Permissions []Permission `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:RoleID;references:ID;joinReferences:PermissionID" json:"permissions,omitempty"`
	Users       []User       `gorm:"foreignKey:RoleID;references:ID" json:"-"`
}

func (Role) TableName() string {
	return "roles"
}
