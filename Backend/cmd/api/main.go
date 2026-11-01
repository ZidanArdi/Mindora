package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/routes"
)

func main() {
	cfg := config.LoadConfig()

	app := fiber.New(fiber.Config{
		AppName: "Mindora API",
	})

	// Setup middleware
	app.Use(middleware.SetupCORS(cfg))

	// Setup application routes
	routes.SetupRoutes(app)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Mindora API running on port %s in %s mode", cfg.Port, cfg.Env)
	if err := app.Listen(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
