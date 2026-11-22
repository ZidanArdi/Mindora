package models

import (
	"time"

	"github.com/google/uuid"
)

// Study session status constants
const (
	StudyStatusActive    = "active"
	StudyStatusPaused    = "paused"
	StudyStatusCompleted = "completed"
	StudyStatusCancelled = "cancelled"
)

// StudySession represents a continuous academic study session.
type StudySession struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_study_sessions_user_id;index:idx_study_sessions_user_status,priority:1" json:"user_id"`
	TaskID          *uuid.UUID `gorm:"type:uuid;index" json:"task_id,omitempty"`
	Subject         string     `gorm:"type:varchar(100);not null" json:"subject"`
	DurationSeconds int        `gorm:"type:integer;not null;default:0" json:"duration_seconds"`
	Status          string     `gorm:"type:varchar(20);not null;default:'active';index:idx_study_sessions_user_status,priority:2" json:"status"`
	StartedAt       time.Time  `gorm:"type:timestamptz;not null;index" json:"started_at"`
	EndedAt         *time.Time `gorm:"type:timestamptz" json:"ended_at,omitempty"`
	CreatedAt       time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`

	User User  `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
	Task *Task `gorm:"foreignKey:TaskID;references:ID;constraint:OnDelete:SET NULL" json:"task,omitempty"`
}

func (StudySession) TableName() string {
	return "study_sessions"
}
