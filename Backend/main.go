package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"

	"mindora/backend/config"
	"mindora/backend/database"
	"mindora/backend/middleware"
	"mindora/backend/routes"
)

func main() {
	// Load application configuration.
	cfg := config.LoadConfig()

	// Initialize PostgreSQL connection.
	db, err := database.ConnectPostgres(cfg.DatabaseURL, cfg.Env)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}

	// Close database connection when application exits.
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	// Initialize Fiber application.
	app := fiber.New(fiber.Config{
		AppName: "Mindora API",
	})

	// Setup middleware.
	app.Use(middleware.SetupCORS(cfg))

	// Setup application routes with database dependency.
	routes.SetupRoutes(app, db)

	// Start HTTP server.
	addr := fmt.Sprintf(":%s", cfg.Port)

	log.Printf(
		"Mindora API running on port %s in %s mode",
		cfg.Port,
		cfg.Env,
	)

	if err := app.Listen(addr); err != nil {
		log.Printf("Failed to start server: %v", err)
	}
}
