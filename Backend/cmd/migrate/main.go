package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
)

func main() {
	var (
		action = flag.String("action", "up", "Migration action: up, down, status, seed")
		steps  = flag.Int("steps", 1, "Number of migration steps to roll back (for down action)")
		migDir = flag.String("dir", "migrations", "Path to migrations directory")
	)
	flag.Parse()

	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Locate migrations directory
	resolvedDir := *migDir
	if !filepath.IsAbs(resolvedDir) {
		candidates := []string{
			resolvedDir,
			filepath.Join("migrations"),
			filepath.Join("Backend", "migrations"),
			filepath.Join("..", "migrations"),
			filepath.Join("..", "..", "migrations"),
		}
		for _, candidate := range candidates {
			if info, err := os.Stat(candidate); err == nil && info.IsDir() {
				resolvedDir = candidate
				break
			}
		}
	}

	migrator := database.NewMigrator(db.SQLDB(), resolvedDir)

	switch *action {
	case "up":
		count, err := migrator.MigrateUp(ctx)
		if err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		log.Printf("Successfully applied %d migration(s).", count)
	case "down":
		count, err := migrator.MigrateDown(ctx, *steps)
		if err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		log.Printf("Successfully rolled back %d migration(s).", count)
	case "status":
		statuses, err := migrator.GetStatus(ctx)
		if err != nil {
			log.Fatalf("Failed to retrieve migration status: %v", err)
		}
		fmt.Println("=== Migration Status ===")
		for _, s := range statuses {
			statusStr := "Pending"
			if s.Applied {
				statusStr = fmt.Sprintf("Applied at %s", s.AppliedAt.Format(time.RFC3339))
			}
			fmt.Printf("[%06d] %-35s -> %s\n", s.Version, s.Name, statusStr)
		}
	case "seed":
		if err := database.SeedRBAC(ctx, db.DB); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}
		log.Println("Database seeded successfully.")
	default:
		log.Fatalf("Unknown action: %s. Supported actions: up, down, status, seed", *action)
	}
}
