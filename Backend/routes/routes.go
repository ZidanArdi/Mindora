package routes

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"mindora/backend/health"
)

// SetupRoutes registers all application routes.
func SetupRoutes(app *fiber.App, db *gorm.DB) {
	api := app.Group("/api/v1")

	healthHandler := health.NewHealthHandler(db)

	// Health check routes.
	api.Get("/health", healthHandler.Check)
	api.Get("/health/database", healthHandler.CheckDatabase)
}
