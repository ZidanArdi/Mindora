package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Task priority constants
const (
	TaskPriorityLow    = "low"
	TaskPriorityMedium = "medium"
	TaskPriorityHigh   = "high"
)

// Task status constants
const (
	TaskStatusPending    = "pending"
	TaskStatusInProgress = "in_progress"
	TaskStatusCompleted  = "completed"
	TaskStatusCancelled  = "cancelled"
)

// Task represents a student productivity task with priority, category, status, and deadline tracking.
type Task struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID      uuid.UUID      `gorm:"type:uuid;not null;index:idx_tasks_user_id;index:idx_tasks_user_status,priority:1;index:idx_tasks_user_deadline,priority:1" json:"user_id"`
	Title       string         `gorm:"type:varchar(255);not null" json:"title"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	Category    string         `gorm:"type:varchar(50);not null;default:'Study'" json:"category"`
	Priority    string         `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	Status      string         `gorm:"type:varchar(20);not null;default:'pending';index:idx_tasks_user_status,priority:2" json:"status"`
	Deadline    *time.Time     `gorm:"type:timestamptz;index:idx_tasks_user_deadline,priority:2" json:"deadline,omitempty"`
	CompletedAt *time.Time     `gorm:"type:timestamptz" json:"completed_at,omitempty"`
	CreatedAt   time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"type:timestamptz;index" json:"-"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Task) TableName() string {
	return "tasks"
}
