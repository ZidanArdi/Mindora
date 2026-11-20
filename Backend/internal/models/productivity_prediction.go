package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	// Virtual fields (not database columns) populated from FeaturesSnapshot
	PredictionState    string   `gorm:"-" json:"prediction_state"`
	StateLabel         string   `gorm:"-" json:"state_label"`
	StateDescription   string   `gorm:"-" json:"state_description"`
	IsBaseline         bool     `gorm:"-" json:"is_baseline"`
	UnrecordedFeatures []string `gorm:"-" json:"unrecorded_features,omitempty"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (ProductivityPrediction) TableName() string {
	return "productivity_predictions"
}

// AfterFind unmarshals metadata from FeaturesSnapshot into virtual fields for API consumers.
func (p *ProductivityPrediction) AfterFind(tx *gorm.DB) error {
	if p.FeaturesSnapshot != "" && p.FeaturesSnapshot != "{}" {
		var snap map[string]interface{}
		if err := json.Unmarshal([]byte(p.FeaturesSnapshot), &snap); err == nil {
			if state, ok := snap["prediction_state"].(string); ok && state != "" {
				p.PredictionState = state
			}
			if label, ok := snap["state_label"].(string); ok && label != "" {
				p.StateLabel = label
			}
			if desc, ok := snap["state_description"].(string); ok && desc != "" {
				p.StateDescription = desc
			}
			if isBase, ok := snap["is_baseline"].(bool); ok {
				p.IsBaseline = isBase
			}
			if unrec, ok := snap["unrecorded_features"].([]interface{}); ok {
				var list []string
				for _, item := range unrec {
					if str, sOk := item.(string); sOk {
						list = append(list, str)
					}
				}
				p.UnrecordedFeatures = list
			}
		}
	}

	// Fallback for legacy records without explicit state metadata
	if p.PredictionState == "" {
		if p.SleepHours == nil || p.PhoneUsageHours == nil || p.StressLevel == nil {
			p.PredictionState = "baseline"
			p.StateLabel = "Baseline Estimate"
			p.StateDescription = "This estimate uses reference values for one or more unrecorded habits. Record your daily habits to obtain a more personalized estimate."
			p.IsBaseline = true
		} else {
			p.PredictionState = "personalized"
			p.StateLabel = "Personalized Estimate"
			p.StateDescription = "This estimate is calculated using your recorded study duration, sleep, phone usage, and stress level."
			p.IsBaseline = false
		}
	}
	return nil
}
