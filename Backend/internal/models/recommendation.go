package models

import (
	"database/sql/driver"
	"errors"
	"time"

	"github.com/google/uuid"
)

// JSONB represents a raw JSON object compatible with PostgreSQL JSONB columns and Fiber JSON responses.
type JSONB []byte

func (j JSONB) Value() (driver.Value, error) {
	if len(j) == 0 {
		return "{}", nil
	}
	return string(j), nil
}

func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = []byte("{}")
		return nil
	}
	switch v := value.(type) {
	case []byte:
		*j = append((*j)[0:0], v...)
	case string:
		*j = []byte(v)
	default:
		return errors.New("cannot scan type into JSONB")
	}
	return nil
}

func (j JSONB) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("{}"), nil
	}
	return j, nil
}

func (j *JSONB) UnmarshalJSON(b []byte) error {
	if j == nil {
		return errors.New("JSONB: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], b...)
	return nil
}

const (
	RecStatusPending     = "pending"
	RecStatusDismissed   = "dismissed"
	RecStatusCompleted   = "completed"
	RecStatusExpired     = "expired"
	RecStatusInvalidated = "invalidated"

	RecPriorityCritical = "critical"
	RecPriorityHigh     = "high"
	RecPriorityMedium   = "medium"
	RecPriorityLow      = "low"
)

// Recommendation represents an individualized, explainable productivity recommendation.
type Recommendation struct {
	ID                    uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"`
	UserID                uuid.UUID  `gorm:"type:uuid;not null;index:idx_recommendations_user_status_valid,priority:1" json:"user_id"`
	RuleID                string     `gorm:"type:varchar(50);not null" json:"rule_id"`
	Category              string     `gorm:"type:varchar(50);not null" json:"category"`
	Priority              string     `gorm:"type:varchar(20);not null;default:'medium'" json:"priority"`
	Title                 string     `gorm:"type:varchar(255);not null" json:"title"`
	Message               string     `gorm:"type:text;not null" json:"message"`
	Explanation           string     `gorm:"type:text;not null" json:"explanation"`
	Evidence              JSONB      `gorm:"type:jsonb;not null;default:'{}'" json:"evidence"`
	ActionRoute           *string    `gorm:"type:varchar(100)" json:"action_route,omitempty"`
	ActionLabel           *string    `gorm:"type:varchar(100)" json:"action_label,omitempty"`
	Status                string     `gorm:"type:varchar(20);not null;default:'pending';index:idx_recommendations_user_status_valid,priority:2" json:"status"`
	EvaluationPeriodStart time.Time  `gorm:"type:timestamptz;not null" json:"evaluation_period_start"`
	EvaluationPeriodEnd   time.Time  `gorm:"type:timestamptz;not null" json:"evaluation_period_end"`
	EvaluationKey         string     `gorm:"type:varchar(150);not null;index:idx_recommendations_user_eval_key,priority:2" json:"evaluation_key"`
	DismissedUntil        *time.Time `gorm:"type:timestamptz" json:"dismissed_until,omitempty"`
	ValidUntil            time.Time  `gorm:"type:timestamptz;not null;index:idx_recommendations_user_status_valid,priority:3" json:"valid_until"`
	CreatedAt             time.Time  `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt             time.Time  `gorm:"type:timestamptz;not null" json:"updated_at"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (Recommendation) TableName() string {
	return "recommendations"
}
