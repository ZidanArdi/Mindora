package routes

import (
	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/handlers"
)

// SetupRoutes registers all application routes
func SetupRoutes(app *fiber.App) {
	api := app.Group("/api")

	healthHandler := handlers.NewHealthHandler()
	api.Get("/health", healthHandler.Check)
}
