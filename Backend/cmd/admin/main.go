package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/pkg/hash"
)

func main() {
	var (
		email    = flag.String("email", "", "Admin email address (overrides ADMIN_EMAIL env)")
		password = flag.String("password", "", "Admin password (overrides ADMIN_PASSWORD env)")
		name     = flag.String("name", "", "Admin display name (overrides ADMIN_NAME env)")
	)
	flag.Parse()

	// If flags are provided, set environment variables
	if *email != "" {
		_ = os.Setenv("ADMIN_EMAIL", *email)
	}
	if *password != "" {
		_ = os.Setenv("ADMIN_PASSWORD", *password)
	}
	if *name != "" {
		_ = os.Setenv("ADMIN_NAME", *name)
	}

	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	if err := database.BootstrapAdmin(ctx, db.DB, hasher, cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error provisioning admin: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Admin provisioning completed successfully.")
}
