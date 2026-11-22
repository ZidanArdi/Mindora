package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

var (
	ErrInvalidRecordDate       = errors.New("invalid record date (must be YYYY-MM-DD and not in the future)")
	ErrInvalidSleepHours       = errors.New("sleep hours must be between 0 and 24")
	ErrInvalidScreenTimeHours   = errors.New("screen time and leisure hours must be between 0 and 24")
	ErrInvalidCoffeeIntake      = errors.New("coffee intake must be non-negative")
	ErrInvalidExerciseMinutes  = errors.New("exercise minutes must be between 0 and 1440")
	ErrInvalidStressLevel      = errors.New("stress level must be an integer between 1 and 10")
)

type UpsertLifestyleInput struct {
	RecordDate       string   `json:"record_date"` // "YYYY-MM-DD"
	SleepHours       *float64 `json:"sleep_hours"`
	PhoneUsageHours  *float64 `json:"phone_usage_hours"`
	SocialMediaHours *float64 `json:"social_media_hours"`
	YouTubeHours     *float64 `json:"youtube_hours"`
	GamingHours      *float64 `json:"gaming_hours"`
	CoffeeIntakeMg   *int     `json:"coffee_intake_mg"`
	ExerciseMinutes  *int     `json:"exercise_minutes"`
	StressLevel      *int     `json:"stress_level"`
	Notes            *string  `json:"notes,omitempty"`
}

type LifestyleService interface {
	GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error)
	UpsertRecord(ctx context.Context, userID uuid.UUID, input UpsertLifestyleInput) (*models.LifestyleRecord, error)
	ListHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.LifestyleRecord, int64, error)
}

type lifestyleService struct {
	repo repositories.LifestyleRepository
}

func NewLifestyleService(repo repositories.LifestyleRepository) LifestyleService {
	return &lifestyleService{repo: repo}
}

func (s *lifestyleService) GetRecordByDate(ctx context.Context, userID uuid.UUID, date time.Time) (*models.LifestyleRecord, error) {
	return s.repo.FindByDate(ctx, userID, date)
}

func (s *lifestyleService) UpsertRecord(ctx context.Context, userID uuid.UUID, input UpsertLifestyleInput) (*models.LifestyleRecord, error) {
	recordDate, err := time.Parse("2006-01-02", input.RecordDate)
	if err != nil {
		return nil, ErrInvalidRecordDate
	}

	// Validate numeric ranges if provided
	if input.SleepHours != nil {
		if *input.SleepHours < 0 || *input.SleepHours > 24 {
			return nil, ErrInvalidSleepHours
		}
	}
	if input.PhoneUsageHours != nil {
		if *input.PhoneUsageHours < 0 || *input.PhoneUsageHours > 24 {
			return nil, ErrInvalidScreenTimeHours
		}
	}
	if input.SocialMediaHours != nil {
		if *input.SocialMediaHours < 0 || *input.SocialMediaHours > 24 {
			return nil, ErrInvalidScreenTimeHours
		}
	}
	if input.YouTubeHours != nil {
		if *input.YouTubeHours < 0 || *input.YouTubeHours > 24 {
			return nil, ErrInvalidScreenTimeHours
		}
	}
	if input.GamingHours != nil {
		if *input.GamingHours < 0 || *input.GamingHours > 24 {
			return nil, ErrInvalidScreenTimeHours
		}
	}
	if input.CoffeeIntakeMg != nil {
		if *input.CoffeeIntakeMg < 0 {
			return nil, ErrInvalidCoffeeIntake
		}
	}
	if input.ExerciseMinutes != nil {
		if *input.ExerciseMinutes < 0 || *input.ExerciseMinutes > 1440 {
			return nil, ErrInvalidExerciseMinutes
		}
	}
	if input.StressLevel != nil {
		if *input.StressLevel < 1 || *input.StressLevel > 10 {
			return nil, ErrInvalidStressLevel
		}
	}

	record := &models.LifestyleRecord{
		UserID:           userID,
		RecordDate:       recordDate,
		SleepHours:       input.SleepHours,
		PhoneUsageHours:  input.PhoneUsageHours,
		SocialMediaHours: input.SocialMediaHours,
		YouTubeHours:     input.YouTubeHours,
		GamingHours:      input.GamingHours,
		CoffeeIntakeMg:   input.CoffeeIntakeMg,
		ExerciseMinutes:  input.ExerciseMinutes,
		StressLevel:      input.StressLevel,
		Notes:            input.Notes,
	}

	if err := s.repo.UpsertDailyRecord(ctx, record); err != nil {
		return nil, err
	}

	// Return updated record from database
	return s.repo.FindByDate(ctx, userID, recordDate)
}

func (s *lifestyleService) ListHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.LifestyleRecord, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	return s.repo.FindHistory(ctx, userID, limit, offset)
}
