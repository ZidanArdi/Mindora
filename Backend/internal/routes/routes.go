package routes

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/handlers"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/groq"
	"mindora/backend/pkg/hash"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/oauth"
	"mindora/backend/pkg/turnstile"
	"mindora/backend/pkg/validator"
)

// SetupRoutes registers all application routes.
func SetupRoutes(app *fiber.App, db *database.Database, cfg *config.Config) {
	// Health check endpoints (preserves existing contracts)
	api := app.Group("/api")
	healthHandler := handlers.NewHealthHandler(db)
	api.Get("/health", healthHandler.Check)
	api.Get("/health/db", healthHandler.DatabaseCheck)

	// API v1 route group
	v1 := app.Group("/api/v1")
	v1.Get("/health", healthHandler.Check)
	v1.Get("/health/db", healthHandler.DatabaseCheck)

	// Infrastructure & Repositories
	userRepo := repositories.NewUserRepository(db.DB)
	identityRepo := repositories.NewIdentityRepository(db.DB)
	sessionRepo := repositories.NewSessionRepository(db.DB)

	taskRepo := repositories.NewTaskRepository(db.DB)
	studyRepo := repositories.NewStudyRepository(db.DB)
	focusRepo := repositories.NewFocusRepository(db.DB)
	lifestyleRepo := repositories.NewLifestyleRepository(db.DB)
	predictionRepo := repositories.NewPredictionRepository(db.DB)
	recRepo := repositories.NewRecommendationRepository(db.DB)

	// Utilities & Providers
	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	tokenManager := jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	turnstileVerifier := turnstile.NewClientWithURL(turnstile.Config{
		Enabled:          cfg.Turnstile.Enabled,
		SecretKey:        cfg.Turnstile.SecretKey,
		ExpectedHostname: cfg.Turnstile.ExpectedHostname,
		ExpectedAction:   cfg.Turnstile.ExpectedAction,
		AllowTestTokens:  cfg.Turnstile.AllowTestTokens && cfg.Env != "production",
	}, cfg.Turnstile.VerifyURL)
	val := validator.New()
	googleService := oauth.NewGoogleService(oauth.Config{
		ClientID:     cfg.GoogleOAuth.ClientID,
		ClientSecret: cfg.GoogleOAuth.ClientSecret,
		RedirectURL:  cfg.GoogleOAuth.RedirectURL,
		Enabled:      cfg.GoogleOAuth.Enabled,
	}, nil)

	// Authentication Service
	authService := services.NewAuthService(
		userRepo,
		identityRepo,
		sessionRepo,
		hasher,
		tokenManager,
		turnstileVerifier,
		val,
		googleService,
	)

	authHandler := handlers.NewAuthHandler(authService, googleService, turnstileVerifier, cfg)

	// Authentication Routes with Rate Limiting
	auth := v1.Group("/auth")
	authLimiterStrict := middleware.AuthRateLimiter(10, 1*time.Minute)
	authLimiterNormal := middleware.AuthRateLimiter(30, 1*time.Minute)

	// Public endpoints
	auth.Post("/register", authLimiterStrict, authHandler.Register)
	auth.Post("/login", authLimiterStrict, authHandler.Login)
	auth.Get("/google", authLimiterNormal, authHandler.GoogleAuth)
	auth.Get("/google/callback", authLimiterNormal, authHandler.GoogleCallback)

	// Session endpoints (rate limited)
	auth.Post("/refresh", authLimiterNormal, authHandler.Refresh)
	auth.Post("/logout", authLimiterNormal, authHandler.Logout)

	// Protected User Profile & Admin Groups
	userHandler := handlers.NewUserHandler(userRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, db.DB)
	jwtAuth := middleware.JWTAuthMiddleware(tokenManager)

	userGroup := v1.Group("/user", jwtAuth, middleware.RequireRole("user", "admin"))
	userGroup.Get("/profile", middleware.RequirePermission(db.DB, "profile", "read"), userHandler.GetProfile)
	userGroup.Put("/profile", middleware.RequirePermission(db.DB, "profile", "update"), userHandler.UpdateProfile)
	userGroup.Get("/tasks/:userId", middleware.RequireOwnership("userId"), userHandler.GetTasks)

	adminGroup := v1.Group("/admin", jwtAuth, middleware.RequireRole("admin"))
	adminGroup.Get("/overview", middleware.RequirePermission(db.DB, "system", "monitor"), adminHandler.GetOverview)
	adminGroup.Get("/users", middleware.RequirePermission(db.DB, "users", "read"), adminHandler.ListUsers)

	// =========================================================================
	// Phase 2: Core Productivity Feature Services & Route Groups
	// =========================================================================
	taskService := services.NewTaskService(taskRepo)
	studyService := services.NewStudyService(studyRepo, taskRepo)
	focusService := services.NewFocusService(focusRepo, taskRepo)
	lifestyleService := services.NewLifestyleService(lifestyleRepo)
	dashboardService := services.NewDashboardService(taskRepo, studyRepo, focusRepo, lifestyleRepo)
	predictionService := services.NewPredictionService(predictionRepo, studyRepo, lifestyleRepo)
	analyticsService := services.NewAnalyticsService(studyRepo, focusRepo, taskRepo, lifestyleRepo, predictionRepo)
	recService := services.NewRecommendationService(recRepo, taskRepo, studyRepo, focusRepo, lifestyleRepo, predictionRepo, userRepo)
	groqClient := groq.NewClient(cfg.Groq.APIKey, cfg.Groq.Model, cfg.Groq.BaseURL, cfg.Groq.Timeout)
	aiService := services.NewAIService(groqClient, taskRepo, studyRepo, focusRepo, lifestyleRepo, predictionRepo, recRepo)

	taskHandler := handlers.NewTaskHandler(taskService)
	studyHandler := handlers.NewStudyHandler(studyService)
	focusHandler := handlers.NewFocusHandler(focusService)
	lifestyleHandler := handlers.NewLifestyleHandler(lifestyleService)
	dashboardHandler := handlers.NewDashboardHandler(dashboardService)
	predictionHandler := handlers.NewPredictionHandler(predictionService)
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsService)
	recHandler := handlers.NewRecommendationHandler(recService)
	aiHandler := handlers.NewAIHandler(aiService)

	// 1. Tasks Management
	tasks := v1.Group("/tasks", jwtAuth, middleware.RequireRole("user"))
	tasks.Get("/", taskHandler.List)
	tasks.Post("/", taskHandler.Create)
	tasks.Get("/:id", taskHandler.GetByID)
	tasks.Put("/:id", taskHandler.Update)
	tasks.Patch("/:id/status", taskHandler.UpdateStatus)
	tasks.Delete("/:id", taskHandler.Delete)

	// 2. Study Tracking
	study := v1.Group("/study", jwtAuth, middleware.RequireRole("user"))
	study.Post("/sessions/start", studyHandler.Start)
	study.Post("/sessions/:id/pause", studyHandler.Pause)
	study.Post("/sessions/:id/resume", studyHandler.Resume)
	study.Post("/sessions/:id/finish", studyHandler.Finish)
	study.Post("/sessions/:id/cancel", studyHandler.Cancel)
	study.Get("/sessions/active", studyHandler.GetActive)
	study.Get("/sessions", studyHandler.List)
	study.Get("/stats/daily", studyHandler.GetDailyStats)

	// 3. Focus Tracking (Individual Interval Architecture)
	focus := v1.Group("/focus", jwtAuth, middleware.RequireRole("user"))
	focus.Post("/sessions/start", focusHandler.Start)
	focus.Post("/sessions/:id/interruption", focusHandler.RecordInterruption)
	focus.Post("/sessions/:id/finish", focusHandler.Finish)
	focus.Post("/sessions/:id/abandon", focusHandler.Abandon)
	focus.Get("/sessions/active", focusHandler.GetActive)
	focus.Get("/sessions/history", focusHandler.List)
	focus.Get("/stats/daily", focusHandler.GetDailyStats)

	// 4. Daily Lifestyle Tracking
	lifestyle := v1.Group("/lifestyle", jwtAuth, middleware.RequireRole("user"))
	lifestyle.Get("/record", lifestyleHandler.GetRecord)
	lifestyle.Post("/record", lifestyleHandler.UpsertRecord)
	lifestyle.Put("/record", lifestyleHandler.UpsertRecord)
	lifestyle.Get("/history", lifestyleHandler.ListHistory)

	// 5. Dashboard Summary
	dashboard := v1.Group("/dashboard", jwtAuth, middleware.RequireRole("user"))
	dashboard.Get("/summary", dashboardHandler.GetSummary)

	// =========================================================================
	// Phase 3: Productivity Analytics & Machine Learning Prediction
	// =========================================================================
	// 6. Productivity Analytics
	analytics := v1.Group("/analytics", jwtAuth, middleware.RequireRole("user"))
	analytics.Get("/daily", analyticsHandler.GetDaily)
	analytics.Get("/weekly", analyticsHandler.GetWeekly)
	analytics.Get("/monthly", analyticsHandler.GetMonthly)
	analytics.Get("/range", analyticsHandler.GetRange)
	analytics.Get("/summary", analyticsHandler.GetSummary)

	// 7. Productivity Prediction
	predictions := v1.Group("/predictions", jwtAuth, middleware.RequireRole("user"))
	predictions.Post("/", predictionHandler.Create)
	predictions.Get("/", predictionHandler.List)
	predictions.Get("/:id", predictionHandler.GetByID)

	// =========================================================================
	// Phase 4: Personalized Productivity Recommendation Engine
	// In accordance with Mindora privacy rules, Admin cannot view recommendations
	// =========================================================================
	recommendations := v1.Group("/recommendations", jwtAuth, middleware.RequireRole("user"))
	recommendations.Get("/", recHandler.GetRecommendations)
	recommendations.Patch("/:id/status", recHandler.UpdateStatus)

	// =========================================================================
	// Phase 5: Groq LLM Academic Productivity Coach
	// =========================================================================
	ai := v1.Group("/ai", jwtAuth, middleware.RequireRole("user"))
	ai.Post("/chat", aiHandler.Chat)
	ai.Post("/explain-prediction", aiHandler.ExplainPrediction)
	ai.Post("/study-plan", aiHandler.GenerateStudyPlan)

	// Root /api aliases for contract compatibility
	apiAnalytics := api.Group("/analytics", jwtAuth, middleware.RequireRole("user"))
	apiAnalytics.Get("/daily", analyticsHandler.GetDaily)
	apiAnalytics.Get("/weekly", analyticsHandler.GetWeekly)
	apiAnalytics.Get("/monthly", analyticsHandler.GetMonthly)
	apiAnalytics.Get("/range", analyticsHandler.GetRange)
	apiAnalytics.Get("/summary", analyticsHandler.GetSummary)

	apiPredictions := api.Group("/predictions", jwtAuth, middleware.RequireRole("user"))
	apiPredictions.Post("/", predictionHandler.Create)
	apiPredictions.Get("/", predictionHandler.List)
	apiPredictions.Get("/:id", predictionHandler.GetByID)

	apiRecs := api.Group("/recommendations", jwtAuth, middleware.RequireRole("user"))
	apiRecs.Get("/", recHandler.GetRecommendations)
	apiRecs.Patch("/:id/status", recHandler.UpdateStatus)

	apiAI := api.Group("/ai", jwtAuth, middleware.RequireRole("user"))
	apiAI.Post("/chat", aiHandler.Chat)
	apiAI.Post("/explain-prediction", aiHandler.ExplainPrediction)
	apiAI.Post("/study-plan", aiHandler.GenerateStudyPlan)
}
