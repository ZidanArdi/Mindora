package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

	StateBaseline     = "baseline"
	StatePersonalized = "personalized"

	LabelBaseline     = "Baseline Estimate"
	LabelPersonalized = "Personalized Estimate"

	DescBaseline     = "This estimate uses reference values for one or more unrecorded habits. Record your daily habits to obtain a more personalized estimate."
	DescPersonalized = "This estimate is calculated using your recorded study duration, sleep, phone usage, and stress level."

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
	ID                 uuid.UUID              `json:"id"`
	PredictedScore     float64                `json:"predicted_score"`
	ModelVersion       string                 `json:"model_version"`
	Algorithm          string                 `json:"algorithm"`
	PredictionDate     string                 `json:"prediction_date"`
	StudyHours         float64                `json:"study_hours"`
	SleepHours         *float64               `json:"sleep_hours,omitempty"`
	PhoneUsageHours    *float64               `json:"phone_usage_hours,omitempty"`
	StressLevel        *int                   `json:"stress_level,omitempty"`
	PredictionState    string                 `json:"prediction_state"`
	StateLabel         string                 `json:"state_label"`
	StateDescription   string                 `json:"state_description"`
	IsBaseline         bool                   `json:"is_baseline"`
	UnrecordedFeatures []string               `json:"unrecorded_features"`
	FeaturesSnapshot   map[string]interface{} `json:"features_snapshot"`
	Disclaimer         string                 `json:"disclaimer"`
	CreatedAt          time.Time              `json:"created_at"`
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
	var studyHoursSupplied float64
	var studyHoursObserved *float64
	var studyImputed bool

	if req.StudyHours != nil {
		if *req.StudyHours < 0 || *req.StudyHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		studyHoursSupplied = *req.StudyHours
		studyHoursObserved = req.StudyHours
		studyImputed = false
	} else {
		// Aggregate from actual completed study sessions
		stats, err := s.studyRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve daily study statistics: %w", err)
		}
		if stats == nil {
			return nil, errors.New("study repository returned unexpected nil statistics")
		}

		if stats.CompletedSessionsCount > 0 {
			// Observed study session exists on this day
			if stats.TotalCompletedSeconds > 0 {
				studyHoursSupplied = math.Round((float64(stats.TotalCompletedSeconds)/3600.0)*100) / 100
			} else {
				studyHoursSupplied = 0.0
			}
			studyHoursObserved = &studyHoursSupplied
			studyImputed = false
		} else {
			// Missing / unrecorded: User has no completed study records for this date
			studyHoursSupplied = math.Round(StudyHoursMean*100) / 100
			studyHoursObserved = nil
			studyImputed = true
		}
	}

	// 2. Resolve Lifestyle Metrics (Sleep, Phone, Stress)
	lifestyleRecord, err := s.lifestyleRepo.FindByDate(ctx, userID, targetDate)
	if err != nil && !errors.Is(err, repositories.ErrLifestyleRecordNotFound) && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to retrieve daily lifestyle record: %w", err)
	}

	var sleepHoursSupplied float64
	var sleepHoursObserved *float64
	var sleepImputed bool

	if req.SleepHours != nil {
		if *req.SleepHours < 0 || *req.SleepHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		sleepHoursSupplied = *req.SleepHours
		sleepHoursObserved = req.SleepHours
		sleepImputed = false
	} else if lifestyleRecord != nil && lifestyleRecord.SleepHours != nil {
		sleepHoursSupplied = *lifestyleRecord.SleepHours
		sleepHoursObserved = lifestyleRecord.SleepHours
		sleepImputed = false
	} else {
		// Impute using training population mean
		sleepHoursSupplied = math.Round(SleepHoursMean*100) / 100
		sleepHoursObserved = nil
		sleepImputed = true
	}

	var phoneHoursSupplied float64
	var phoneHoursObserved *float64
	var phoneImputed bool

	if req.PhoneUsageHours != nil {
		if *req.PhoneUsageHours < 0 || *req.PhoneUsageHours > 24 {
			return nil, ErrInvalidPredictionInput
		}
		phoneHoursSupplied = *req.PhoneUsageHours
		phoneHoursObserved = req.PhoneUsageHours
		phoneImputed = false
	} else if lifestyleRecord != nil && lifestyleRecord.PhoneUsageHours != nil {
		phoneHoursSupplied = *lifestyleRecord.PhoneUsageHours
		phoneHoursObserved = lifestyleRecord.PhoneUsageHours
		phoneImputed = false
	} else {
		// Impute using training population mean
		phoneHoursSupplied = math.Round(PhoneHoursMean*100) / 100
		phoneHoursObserved = nil
		phoneImputed = true
	}

	var stressLevelSupplied int
	var stressLevelObserved *int
	var stressImputed bool

	if req.StressLevel != nil {
		if *req.StressLevel < 1 || *req.StressLevel > 10 {
			return nil, ErrInvalidPredictionInput
		}
		stressLevelSupplied = *req.StressLevel
		stressLevelObserved = req.StressLevel
		stressImputed = false
	} else if lifestyleRecord != nil && lifestyleRecord.StressLevel != nil {
		stressLevelSupplied = *lifestyleRecord.StressLevel
		stressLevelObserved = lifestyleRecord.StressLevel
		stressImputed = false
	} else {
		// Impute using training population mean
		stressLevelSupplied = int(math.Round(StressLevelMean))
		stressLevelObserved = nil
		stressImputed = true
	}

	// Determine Prediction State and List Unrecorded Features
	var unrecordedFeatures []string
	if studyImputed {
		unrecordedFeatures = append(unrecordedFeatures, "study_hours")
	}
	if sleepImputed {
		unrecordedFeatures = append(unrecordedFeatures, "sleep_hours")
	}
	if phoneImputed {
		unrecordedFeatures = append(unrecordedFeatures, "phone_usage_hours")
	}
	if stressImputed {
		unrecordedFeatures = append(unrecordedFeatures, "stress_level")
	}

	isBaseline := len(unrecordedFeatures) > 0
	predState := StatePersonalized
	stateLabel := LabelPersonalized
	stateDesc := DescPersonalized
	if isBaseline {
		predState = StateBaseline
		stateLabel = LabelBaseline
		stateDesc = DescBaseline
	}

	// 3. Compute Ridge Regression Score
	// Feature 0: study_hours_per_day
	scaledStudy := (studyHoursSupplied - StudyHoursMean) / StudyHoursScale
	// Feature 1: sleep_hours
	scaledSleep := (sleepHoursSupplied - SleepHoursMean) / SleepHoursScale
	// Feature 2: phone_usage_hours
	scaledPhone := (phoneHoursSupplied - PhoneHoursMean) / PhoneHoursScale
	// Feature 3: stress_level
	scaledStress := (float64(stressLevelSupplied) - StressLevelMean) / StressLevelScale

	rawScore := ModelIntercept +
		(scaledStudy * StudyHoursCoef) +
		(scaledSleep * SleepHoursCoef) +
		(scaledPhone * PhoneHoursCoef) +
		(scaledStress * StressLevelCoef)

	// Clamp between 0.00 and 100.00
	predictedScore := math.Max(0.0, math.Min(100.0, rawScore))
	predictedScore = math.Round(predictedScore*100) / 100

	// 4. Construct Features Snapshot JSON
	var studyObsVal interface{}
	if studyHoursObserved != nil {
		studyObsVal = *studyHoursObserved
	}
	var sleepObsVal interface{}
	if sleepHoursObserved != nil {
		sleepObsVal = *sleepHoursObserved
	}
	var phoneObsVal interface{}
	if phoneHoursObserved != nil {
		phoneObsVal = *phoneHoursObserved
	}
	var stressObsVal interface{}
	if stressLevelObserved != nil {
		stressObsVal = *stressLevelObserved
	}

	featuresSnapshot := map[string]interface{}{
		"study_hours":          studyHoursSupplied,
		"study_hours_supplied": studyHoursSupplied,
		"study_hours_observed": studyObsVal,
		"study_hours_imputed":  studyImputed,

		"sleep_hours":          sleepHoursSupplied,
		"sleep_hours_supplied": sleepHoursSupplied,
		"sleep_hours_observed": sleepObsVal,
		"sleep_hours_imputed":  sleepImputed,

		"phone_usage_hours":          phoneHoursSupplied,
		"phone_usage_hours_supplied": phoneHoursSupplied,
		"phone_usage_hours_observed": phoneObsVal,
		"phone_usage_hours_imputed":  phoneImputed,

		"stress_level":          stressLevelSupplied,
		"stress_level_supplied": stressLevelSupplied,
		"stress_level_observed": stressObsVal,
		"stress_level_imputed":  stressImputed,

		"prediction_state":    predState,
		"state_label":         stateLabel,
		"state_description":   stateDesc,
		"is_baseline":         isBaseline,
		"unrecorded_features": unrecordedFeatures,

		"model_version":        ModelVersionCurrent,
		"algorithm":            ModelAlgorithmName,
		"raw_score":            math.Round(rawScore*10000) / 10000,
		"disclaimer":           ModelDisclaimerText,
		"prediction_timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	snapshotJSON, _ := json.Marshal(featuresSnapshot)

	// 5. Persist Prediction Record
	predictionModel := &models.ProductivityPrediction{
		ID:                 uuid.New(),
		UserID:             userID,
		PredictedScore:     predictedScore,
		ModelVersion:       ModelVersionCurrent,
		StudyHours:         studyHoursSupplied,
		SleepHours:         sleepHoursObserved,
		PhoneUsageHours:    phoneHoursObserved,
		StressLevel:        stressLevelObserved,
		FeaturesSnapshot:   string(snapshotJSON),
		PredictionDate:     targetDate,
		CreatedAt:          time.Now().UTC(),
		PredictionState:    predState,
		StateLabel:         stateLabel,
		StateDescription:   stateDesc,
		IsBaseline:         isBaseline,
		UnrecordedFeatures: unrecordedFeatures,
	}

	if err := s.predictionRepo.Create(ctx, predictionModel); err != nil {
		return nil, err
	}

	return &PredictionResponse{
		ID:                 predictionModel.ID,
		PredictedScore:     predictionModel.PredictedScore,
		ModelVersion:       predictionModel.ModelVersion,
		Algorithm:          ModelAlgorithmName,
		PredictionDate:     targetDate.Format("2006-01-02"),
		StudyHours:         studyHoursSupplied,
		SleepHours:         sleepHoursObserved,
		PhoneUsageHours:    phoneHoursObserved,
		StressLevel:        stressLevelObserved,
		PredictionState:    predState,
		StateLabel:         stateLabel,
		StateDescription:   stateDesc,
		IsBaseline:         isBaseline,
		UnrecordedFeatures: unrecordedFeatures,
		FeaturesSnapshot:   featuresSnapshot,
		Disclaimer:         ModelDisclaimerText,
		CreatedAt:          predictionModel.CreatedAt,
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
