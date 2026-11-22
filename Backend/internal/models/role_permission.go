package models

import (
	"time"

	"github.com/google/uuid"
)

// RolePermission is the junction table connecting Roles and Permissions.
// It uses a composite primary key (role_id, permission_id).
type RolePermission struct {
	RoleID       uuid.UUID  `gorm:"type:uuid;primaryKey;not null" json:"role_id"`
	PermissionID uuid.UUID  `gorm:"type:uuid;primaryKey;not null" json:"permission_id"`
	CreatedAt    time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	Role         Role       `gorm:"foreignKey:RoleID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
	Permission   Permission `gorm:"foreignKey:PermissionID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
}

func (RolePermission) TableName() string {
	return "role_permissions"
}
