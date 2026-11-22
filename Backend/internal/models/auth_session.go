package models

import (
	"time"

	"github.com/google/uuid"
)

// AuthSession stores authentication session metadata for refresh token validation, rotation, and revocation.
// Raw tokens are never stored; only one-way cryptographic hashes.
// FamilyID groups tokens in the same rotation lineage to detect token reuse.
// ReplacedByID, RotationOpID, EncryptedResponse, and ResponseExpiresAt support distributed idempotency and concurrency.
type AuthSession struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            uuid.UUID  `gorm:"type:uuid;not null;index;index:idx_auth_sessions_user_expires,priority:1" json:"user_id"`
	FamilyID          uuid.UUID  `gorm:"type:uuid;index" json:"family_id"`
	TokenHash         string     `gorm:"type:text;not null;uniqueIndex" json:"-"`
	IPAddress         *string    `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent         *string    `gorm:"type:text" json:"user_agent,omitempty"`
	ExpiresAt         time.Time  `gorm:"type:timestamptz;not null;index;index:idx_auth_sessions_user_expires,priority:2" json:"expires_at"`
	RevokedAt         *time.Time `gorm:"type:timestamptz" json:"revoked_at,omitempty"`
	CreatedAt         time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`

	ReplacedByID      *uuid.UUID `gorm:"type:uuid;index" json:"replaced_by_id,omitempty"`
	RotationOpID      *string    `gorm:"type:varchar(64);index" json:"rotation_op_id,omitempty"`
	EncryptedResponse *string    `gorm:"type:text" json:"-"`
	ResponseExpiresAt *time.Time `gorm:"type:timestamptz" json:"-"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
}

func (AuthSession) TableName() string {
	return "auth_sessions"
}
