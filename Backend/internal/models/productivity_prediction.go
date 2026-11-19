package models

import (
	"time"

	"github.com/google/uuid"
)

// ProductivityPrediction stores historical ML-generated productivity score estimates.
// Predictions represent probabilistic regression estimates based on habit indicators,
// and are strictly distinct from measured activity metrics.
type ProductivityPrediction struct {
	ID               uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	UserID           uuid.UUID `gorm:"type:uuid;not null;index:idx_productivity_predictions_user_date,priority:1;index:idx_productivity_predictions_user_created,priority:1" json:"user_id"`
	PredictedScore   float64   `gorm:"type:numeric(5,2);not null" json:"predicted_score"`
	ModelVersion     string    `gorm:"type:varchar(50);not null" json:"model_version"`
	StudyHours       float64   `gorm:"type:numeric(4,2);not null" json:"study_hours"`
	SleepHours       *float64  `gorm:"type:numeric(4,2)" json:"sleep_hours,omitempty"`
	PhoneUsageHours  *float64  `gorm:"type:numeric(4,2)" json:"phone_usage_hours,omitempty"`
	StressLevel      *int      `gorm:"type:integer" json:"stress_level,omitempty"`
	FeaturesSnapshot string    `gorm:"type:jsonb;not null;default:'{}'" json:"features_snapshot"`
	PredictionDate   time.Time `gorm:"type:date;not null;index:idx_productivity_predictions_user_date,priority:2" json:"prediction_date"`
	CreatedAt        time.Time `gorm:"type:timestamptz;not null" json:"created_at"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProductivityPrediction) TableName() string {
	return "productivity_predictions"
}
