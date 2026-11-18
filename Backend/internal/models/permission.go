package models

import (
	"time"

	"github.com/google/uuid"
)

// Permission defines a granular permission mapped to a resource and action.
// Uniqueness is enforced on the tuple (resource, action).
type Permission struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Resource    string    `gorm:"type:varchar(60);not null;uniqueIndex:idx_permissions_resource_action,priority:1" json:"resource"`
	Action      string    `gorm:"type:varchar(30);not null;uniqueIndex:idx_permissions_resource_action,priority:2" json:"action"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	CreatedAt   time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	Roles       []Role    `gorm:"many2many:role_permissions;foreignKey:ID;joinForeignKey:PermissionID;references:ID;joinReferences:RoleID" json:"-"`
}

func (Permission) TableName() string {
	return "permissions"
}

// Code returns the unified permission identifier in resource.action format.
func (p Permission) Code() string {
	return p.Resource + "." + p.Action
}
