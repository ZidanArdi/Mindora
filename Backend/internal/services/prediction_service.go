package services

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

var (
	ErrInvalidPredictionInput = errors.New("invalid prediction input values")
)

const (
	ModelVersionCurrent = "mindora-productivity-ridge-v1.0.0"
	ModelAlgorithmName  = "Ridge Regression (L2-Regularized OLS)"
	ModelDisclaimerText = "Predictions are mathematical estimates based on daily behavioral habits and must NOT be interpreted as actual measured productivity."

	// Model Weights & Preprocessing Parameters (Fitted strictly on 70% training split of dataset)
	ModelIntercept = 50.23716964285723

	StudyHoursMean  = 5.247772321428595
	StudyHoursScale = 2.7395364656044996
	StudyHoursCoef  = 11.789324586911583

	SleepHoursMean  = 6.536850892857139
	SleepHoursScale = 2.0372977820552127
	SleepHoursCoef  = 5.497139844559539

	PhoneHoursMean  = 6.245928571428571
	PhoneHoursScale = 3.3099733887425202
	PhoneHoursCoef  = -5.37446525483219

	StressLevelMean  = 5.469821428571429
	StressLevelScale = 2.8705057985555413
	StressLevelCoef  = -3.0970200605794975
)

// PredictionRequest represents client input for generating a productivity prediction.
// If values are nil, the service automatically derives them from the user's recorded activities.
type PredictionRequest struct {
	Date            *string  `json:"date,omitempty"` // YYYY-MM-DD
	Timezone        *string  `json:"timezone,omitempty"`
	StudyHours      *float64 `json:"study_hours,omitempty"`
	SleepHours      *float64 `json:"sleep_hours,omitempty"`
	PhoneUsageHours *float64 `json:"phone_usage_hours,omitempty"`
	StressLevel     *int     `json:"stress_level,omitempty"`
}

// PredictionResponse represents the verified prediction output.
type PredictionResponse struct {
	ID               uuid.UUID              `json:"id"`
	PredictedScore   float64                `json:"predicted_score"`
	ModelVersion     string                 `json:"model_version"`
	Algorithm        string                 `json:"algorithm"`
	PredictionDate   string                 `json:"prediction_date"`
	StudyHours       float64                `json:"study_hours"`
	SleepHours       *float64               `json:"sleep_hours,omitempty"`
	PhoneUsageHours  *float64               `json:"phone_usage_hours,omitempty"`
	StressLevel      *int                   `json:"stress_level,omitempty"`
	FeaturesSnapshot map[string]interface{} `json:"features_snapshot"`
	Disclaimer       string                 `json:"disclaimer"`
	CreatedAt        time.Time              `json:"created_at"`
}

type PredictionService interface {
	Predict(ctx context.Context, userID uuid.UUID, req PredictionRequest) (*PredictionResponse, error)
	GetHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.ProductivityPrediction, int64, error)
	GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ProductivityPrediction, error)
}

type predictionService struct {
	predictionRepo repositories.PredictionRepository
	studyRepo      repositories.StudyRepository
	lifestyleRepo  repositories.LifestyleRepository
}

func NewPredictionService(
	predictionRepo repositories.PredictionRepository,
	studyRepo repositories.StudyRepository,
	lifestyleRepo repositories.LifestyleRepository,
) PredictionService {
	return &predictionService{
		predictionRepo: predictionRepo,
		studyRepo:      studyRepo,
		lifestyleRepo:  lifestyleRepo,
	}
}

// parseTimezone parses timezone name or falls back to UTC.
func parseUserLocation(tzName *string) *time.Location {
	if tzName == nil || *tzName == "" {
		return time.UTC
	}
	loc, err := time.LoadLocation(*tzName)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (s *predictionService) Predict(ctx context.Context, userID uuid.UUID, req PredictionRequest) (*PredictionResponse, error) {
	loc := parseUserLocation(req.Timezone)

	var targetDate time.Time
	if req.Date != nil && *req.Date != "" {
		parsed, err := time.ParseInLocation("2006-01-02", *req.Date, loc)
		if err != nil {
			return nil, ErrInvalidPredictionInput
		}
		targetDate = parsed
	} else {
		now := time.Now().In(loc)
		targetDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	startOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, loc).UTC()
	endOfDay := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, loc).UTC()

	// 1. Resolve Study Hours
	var studyHours float64
	var studyImputed bool
	if req.StudyHours != nil {
		if *req.StudyHours < 0 || *req.StudyHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		studyHours = *req.StudyHours
	} else {
		// Aggregate from actual completed study sessions
		durationSec, err := s.studyRepo.GetCompletedDuration(ctx, userID, startOfDay, endOfDay)
		if err == nil && durationSec > 0 {
			studyHours = math.Round((float64(durationSec)/3600.0)*100) / 100
		} else {
			// If no study session recorded, study hours is 0.0
			studyHours = 0.0
		}
	}

	// 2. Resolve Lifestyle Metrics (Sleep, Phone, Stress)
	lifestyleRecord, _ := s.lifestyleRepo.FindByDate(ctx, userID, targetDate)

	var sleepHours *float64
	var sleepImputed bool
	if req.SleepHours != nil {
		if *req.SleepHours < 0 || *req.SleepHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		sleepHours = req.SleepHours
	} else if lifestyleRecord != nil && lifestyleRecord.SleepHours != nil {
		sleepHours = lifestyleRecord.SleepHours
	} else {
		// Impute using training population mean
		defaultSleep := math.Round(SleepHoursMean*100) / 100
		sleepHours = &defaultSleep
		sleepImputed = true
	}

	var phoneHours *float64
	var phoneImputed bool
	if req.PhoneUsageHours != nil {
		if *req.PhoneUsageHours < 0 || *req.PhoneUsageHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		phoneHours = req.PhoneUsageHours
	} else if lifestyleRecord != nil && lifestyleRecord.PhoneUsageHours != nil {
		phoneHours = lifestyleRecord.PhoneUsageHours
	} else {
		defaultPhone := math.Round(PhoneHoursMean*100) / 100
		phoneHours = &defaultPhone
		phoneImputed = true
	}

	var stressLevel *int
	var stressImputed bool
	if req.StressLevel != nil {
		if *req.StressLevel < 1 || *req.StressLevel > 10 {
			return nil, ErrInvalidPredictionInput
		}
		stressLevel = req.StressLevel
	} else if lifestyleRecord != nil && lifestyleRecord.StressLevel != nil {
		stressLevel = lifestyleRecord.StressLevel
	} else {
		defaultStress := int(math.Round(StressLevelMean))
		stressLevel = &defaultStress
		stressImputed = true
	}

	// 3. Compute Ridge Regression Score
	// Feature 0: study_hours_per_day
	scaledStudy := (studyHours - StudyHoursMean) / StudyHoursScale
	// Feature 1: sleep_hours
	scaledSleep := (*sleepHours - SleepHoursMean) / SleepHoursScale
	// Feature 2: phone_usage_hours
	scaledPhone := (*phoneHours - PhoneHoursMean) / PhoneHoursScale
	// Feature 3: stress_level
	scaledStress := (float64(*stressLevel) - StressLevelMean) / StressLevelScale

	rawScore := ModelIntercept +
		(scaledStudy * StudyHoursCoef) +
		(scaledSleep * SleepHoursCoef) +
		(scaledPhone * PhoneHoursCoef) +
		(scaledStress * StressLevelCoef)

	// Clamp between 0.00 and 100.00
	predictedScore := math.Max(0.0, math.Min(100.0, rawScore))
	predictedScore = math.Round(predictedScore*100) / 100

	// 4. Construct Features Snapshot JSON
	featuresSnapshot := map[string]interface{}{
		"study_hours":               studyHours,
		"study_hours_imputed":       studyImputed,
		"sleep_hours":               *sleepHours,
		"sleep_hours_imputed":       sleepImputed,
		"phone_usage_hours":         *phoneHours,
		"phone_usage_hours_imputed": phoneImputed,
		"stress_level":              *stressLevel,
		"stress_level_imputed":      stressImputed,
		"model_version":             ModelVersionCurrent,
		"algorithm":                 ModelAlgorithmName,
		"raw_score":                 math.Round(rawScore*10000) / 10000,
		"disclaimer":                ModelDisclaimerText,
	}

	snapshotJSON, _ := json.Marshal(featuresSnapshot)

	// 5. Persist Prediction Record
	predictionModel := &models.ProductivityPrediction{
		ID:               uuid.New(),
		UserID:           userID,
		PredictedScore:   predictedScore,
		ModelVersion:     ModelVersionCurrent,
		StudyHours:       studyHours,
		SleepHours:       sleepHours,
		PhoneUsageHours:  phoneHours,
		StressLevel:      stressLevel,
		FeaturesSnapshot: string(snapshotJSON),
		PredictionDate:   targetDate,
		CreatedAt:        time.Now().UTC(),
	}

	if err := s.predictionRepo.Create(ctx, predictionModel); err != nil {
		return nil, err
	}

	return &PredictionResponse{
		ID:               predictionModel.ID,
		PredictedScore:   predictionModel.PredictedScore,
		ModelVersion:     predictionModel.ModelVersion,
		Algorithm:        ModelAlgorithmName,
		PredictionDate:   targetDate.Format("2006-01-02"),
		StudyHours:       studyHours,
		SleepHours:       sleepHours,
		PhoneUsageHours:  phoneHours,
		StressLevel:      stressLevel,
		FeaturesSnapshot: featuresSnapshot,
		Disclaimer:       ModelDisclaimerText,
		CreatedAt:        predictionModel.CreatedAt,
	}, nil
}

func (s *predictionService) GetHistory(ctx context.Context, userID uuid.UUID, page, limit int) ([]models.ProductivityPrediction, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit
	return s.predictionRepo.FindHistory(ctx, userID, limit, offset)
}

func (s *predictionService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.ProductivityPrediction, error) {
	return s.predictionRepo.FindByID(ctx, id, userID)
}
