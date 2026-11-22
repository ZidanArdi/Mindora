package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// LifestyleRecord represents a student's daily lifestyle habits for one calendar date.
// Unrecorded metrics remain NULL to prevent treating missing data as zero.
type LifestyleRecord struct {
	ID                uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID            uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:uq_lifestyle_records_user_date,priority:1;index:idx_lifestyle_records_user_date,priority:1" json:"user_id"`
	RecordDate        time.Time      `gorm:"type:date;not null;uniqueIndex:uq_lifestyle_records_user_date,priority:2;index:idx_lifestyle_records_user_date,priority:2" json:"record_date"`
	SleepHours        *float64       `gorm:"type:numeric(4,2)" json:"sleep_hours"`
	PhoneUsageHours   *float64       `gorm:"type:numeric(4,2)" json:"phone_usage_hours"`
	SocialMediaHours  *float64       `gorm:"type:numeric(4,2)" json:"social_media_hours"`
	YouTubeHours      *float64       `gorm:"type:numeric(4,2);column:youtube_hours" json:"youtube_hours"`
	GamingHours       *float64       `gorm:"type:numeric(4,2)" json:"gaming_hours"`
	CoffeeIntakeMg    *int           `gorm:"type:integer" json:"coffee_intake_mg"`
	ExerciseMinutes   *int           `gorm:"type:integer" json:"exercise_minutes"`
	StressLevel       *int           `gorm:"type:integer" json:"stress_level"`
	Notes             *string        `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt         time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt         time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"type:timestamptz;index" json:"-"`

	User User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (LifestyleRecord) TableName() string {
	return "lifestyle_records"
}
