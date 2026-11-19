package models

import (
	"time"

	"github.com/google/uuid"
)

// Focus interval type constants
const (
	FocusTypeWork  = "focus"
	FocusTypeBreak = "break"
)

// Focus interval status constants
const (
	FocusStatusActive    = "active"
	FocusStatusCompleted = "completed"
	FocusStatusAbandoned = "abandoned"
)

// FocusSession represents an individual focus or break interval.
// Architectural rule: One database row represents exactly one interval.
type FocusSession struct {
	ID                 uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID  `gorm:"type:uuid;not null;index:idx_focus_sessions_user_id;index:idx_focus_sessions_user_type_status,priority:1" json:"user_id"`
	TaskID             *uuid.UUID `gorm:"type:uuid;index" json:"task_id,omitempty"`
	SessionType        string     `gorm:"type:varchar(20);not null;index:idx_focus_sessions_user_type_status,priority:2" json:"session_type"`
	StartedAt          time.Time  `gorm:"type:timestamptz;not null;index" json:"started_at"`
	EndedAt            *time.Time `gorm:"type:timestamptz" json:"ended_at,omitempty"`
	DurationMinutes    int        `gorm:"type:integer;not null;default:0" json:"duration_minutes"`
	Status             string     `gorm:"type:varchar(20);not null;default:'active';index:idx_focus_sessions_user_type_status,priority:3" json:"status"`
	InterruptionCount  int        `gorm:"type:integer;not null;default:0" json:"interruption_count"`
	CreatedAt          time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`

	User User  `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Task *Task `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:SET NULL" json:"task,omitempty"`
}

func (FocusSession) TableName() string {
	return "focus_sessions"
}
