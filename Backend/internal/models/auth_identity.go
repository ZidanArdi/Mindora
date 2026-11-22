package models

import (
	"time"

	"github.com/google/uuid"
)

// Supported authentication identity providers
const (
	ProviderGoogle = "google"
)

// AuthIdentity maps an external identity provider identity (e.g. Google sub) to a Mindora user account.
type AuthIdentity struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index" json:"user_id"`
	Provider        string    `gorm:"type:varchar(30);not null;uniqueIndex:idx_auth_identities_provider_sub,priority:1" json:"provider"`
	ProviderSubject string    `gorm:"type:varchar(255);not null;uniqueIndex:idx_auth_identities_provider_sub,priority:2" json:"provider_subject"`
	ProviderEmail   *string   `gorm:"type:varchar(255)" json:"provider_email,omitempty"`
	CreatedAt       time.Time `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt       time.Time `gorm:"type:timestamptz;not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:RESTRICT" json:"-"`
}

func (AuthIdentity) TableName() string {
	return "auth_identities"
}
