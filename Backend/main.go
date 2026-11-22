package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/routes"
)

func main() {
	cfg := config.LoadConfig()

	// Validate configuration integrity (enforces fail-secure in production)
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Fatal configuration validation error: %v", err)
	}

	// 1. Establish database connection
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Fatal: Database initialization failed: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// 2. Initialize Fiber application with explicit trusted proxy settings
	app := fiber.New(fiber.Config{
		AppName:                 "Mindora API",
		EnableTrustedProxyCheck: cfg.Proxy.EnableTrustedProxyCheck,
		TrustedProxies:          cfg.Proxy.TrustedProxies,
		ProxyHeader:             fiber.HeaderXForwardedFor,
	})

	// 3. Setup middleware
	app.Use(middleware.SetupCORS(cfg))

	// 4. Setup application routes
	routes.SetupRoutes(app, db, cfg)

	// 5. Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	addr := fmt.Sprintf(":%s", cfg.Port)

	go func() {
		log.Printf("Mindora API running on port %s in %s mode", cfg.Port, cfg.Env)
		if err := app.Listen(addr); err != nil {
			log.Printf("Server event: %v", err)
		}
	}()

	// Block until signal is received
	<-sigChan
	log.Println("Received termination signal, initiating graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Printf("Error shutting down Fiber server: %v", err)
	}

	log.Println("Mindora API shutdown completed.")
}
