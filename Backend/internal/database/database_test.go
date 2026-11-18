package database_test

import (
	"context"
	"testing"
	"time"

	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
)

func TestDatabaseConnectionAndPing(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		t.Errorf("expected successful ping, got error: %v", err)
	}
}

func TestSeederIdempotencyAndCount(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Run seeder once
	if err := database.SeedRBAC(ctx, db.DB); err != nil {
		t.Fatalf("seeder first execution failed: %v", err)
	}

	// Run seeder second time to verify idempotency
	if err := database.SeedRBAC(ctx, db.DB); err != nil {
		t.Fatalf("seeder second execution failed: %v", err)
	}

	// Verify exact role count
	var roleCount int64
	if err := db.DB.Model(&models.Role{}).Count(&roleCount).Error; err != nil {
		t.Fatalf("failed to count roles: %v", err)
	}
	if roleCount != 2 {
		t.Errorf("expected exactly 2 roles, got %d", roleCount)
	}

	// Verify exact permission count
	var permCount int64
	if err := db.DB.Model(&models.Permission{}).Count(&permCount).Error; err != nil {
		t.Fatalf("failed to count permissions: %v", err)
	}
	if permCount != 13 {
		t.Errorf("expected exactly 13 permissions, got %d", permCount)
	}

	// Verify User role has 6 permissions
	var userRole models.Role
	if err := db.DB.Preload("Permissions").Where("name = ?", "user").First(&userRole).Error; err != nil {
		t.Fatalf("failed to load user role with permissions: %v", err)
	}
	if len(userRole.Permissions) != 6 {
		t.Errorf("expected user role to have 6 permissions, got %d", len(userRole.Permissions))
	}

	// Verify Admin role has 7 permissions
	var adminRole models.Role
	if err := db.DB.Preload("Permissions").Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		t.Fatalf("failed to load admin role with permissions: %v", err)
	}
	if len(adminRole.Permissions) != 7 {
		t.Errorf("expected admin role to have 7 permissions, got %d", len(adminRole.Permissions))
	}
}
