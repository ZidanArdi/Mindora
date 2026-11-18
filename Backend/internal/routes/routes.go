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

	// Utilities & Providers
	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	tokenManager := jwt.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	turnstileVerifier := turnstile.NewClient(turnstile.Config{
		Enabled:          cfg.Turnstile.Enabled,
		SecretKey:        cfg.Turnstile.SecretKey,
		ExpectedHostname: cfg.Turnstile.ExpectedHostname,
		ExpectedAction:   cfg.Turnstile.ExpectedAction,
		AllowTestTokens:  cfg.Env != "production",
	})
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

	// Phase 4: Protected User Route Group
	userHandler := handlers.NewUserHandler(userRepo)
	adminHandler := handlers.NewAdminHandler(userRepo, db.DB)
	jwtAuth := middleware.JWTAuthMiddleware(tokenManager)

	userGroup := v1.Group("/user", jwtAuth, middleware.RequireRole("user", "admin"))
	userGroup.Get("/profile", middleware.RequirePermission(db.DB, "profile", "read"), userHandler.GetProfile)
	userGroup.Put("/profile", middleware.RequirePermission(db.DB, "profile", "update"), userHandler.UpdateProfile)
	userGroup.Get("/tasks/:userId", middleware.RequireOwnership("userId"), userHandler.GetTasks)

	// Phase 4: Protected Admin Route Group
	adminGroup := v1.Group("/admin", jwtAuth, middleware.RequireRole("admin"))
	adminGroup.Get("/overview", middleware.RequirePermission(db.DB, "system", "monitor"), adminHandler.GetOverview)
	adminGroup.Get("/users", middleware.RequirePermission(db.DB, "users", "read"), adminHandler.ListUsers)
}
