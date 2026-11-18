package database_test

import (
	"context"
	"os"
	"testing"
	"time"

	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
	"mindora/backend/pkg/hash"
)

func TestBootstrapAdmin_ValidationAndSafety(t *testing.T) {
	hasher := hash.NewBcryptHasher(4)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	devCfg := &config.Config{Env: "development"}
	prodCfg := &config.Config{Env: "production"}

	// Case 1: Rejection in production
	os.Setenv("ADMIN_EMAIL", "admin@mindora.app")
	os.Setenv("ADMIN_PASSWORD", "SuperSecure123!")
	defer func() {
		os.Unsetenv("ADMIN_EMAIL")
		os.Unsetenv("ADMIN_PASSWORD")
	}()

	err := database.BootstrapAdmin(ctx, nil, hasher, prodCfg)
	if err != database.ErrProductionRestricted {
		t.Fatalf("expected ErrProductionRestricted, got %v", err)
	}

	// Case 2: Missing email
	os.Unsetenv("ADMIN_EMAIL")
	os.Unsetenv("ADMIN_PASSWORD")
	err = database.BootstrapAdmin(ctx, nil, hasher, devCfg)
	if err != database.ErrAdminEmailRequired {
		t.Fatalf("expected ErrAdminEmailRequired, got %v", err)
	}

	// Case 3: Missing password
	os.Setenv("ADMIN_EMAIL", "admin@mindora.app")
	err = database.BootstrapAdmin(ctx, nil, hasher, devCfg)
	if err != database.ErrAdminPasswordRequired {
		t.Fatalf("expected ErrAdminPasswordRequired, got %v", err)
	}

	// Case 4: Password too short (< 8 chars)
	os.Setenv("ADMIN_PASSWORD", "short")
	err = database.BootstrapAdmin(ctx, nil, hasher, devCfg)
	if err != database.ErrAdminPasswordTooShort {
		t.Fatalf("expected ErrAdminPasswordTooShort, got %v", err)
	}
}

func TestBootstrapAdmin_Integration(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Ensure RBAC roles are seeded
	if err := database.SeedRBAC(ctx, db.DB); err != nil {
		t.Fatalf("failed to seed RBAC roles: %v", err)
	}

	hasher := hash.NewBcryptHasher(cfg.BcryptCost)
	adminEmail := "admin@mindora.app"
	adminPassword := "AdminSecret2026!#"

	os.Setenv("ADMIN_EMAIL", adminEmail)
	os.Setenv("ADMIN_PASSWORD", adminPassword)
	os.Setenv("ADMIN_NAME", "Mindora Admin")
	defer func() {
		os.Unsetenv("ADMIN_EMAIL")
		os.Unsetenv("ADMIN_PASSWORD")
		os.Unsetenv("ADMIN_NAME")
	}()

	// 1. Initial provision or update
	if err := database.BootstrapAdmin(ctx, db.DB, hasher, cfg); err != nil {
		t.Fatalf("failed to bootstrap admin: %v", err)
	}

	// 2. Verify account in database
	var adminUser models.User
	if err := db.DB.Preload("Role").Where("LOWER(email) = ?", adminEmail).First(&adminUser).Error; err != nil {
		t.Fatalf("admin user not found in database: %v", err)
	}

	if adminUser.Role.Name != "admin" {
		t.Errorf("expected role 'admin', got '%s'", adminUser.Role.Name)
	}
	if adminUser.Status != models.UserStatusActive {
		t.Errorf("expected status 'active', got '%s'", adminUser.Status)
	}
	if adminUser.PasswordHash == nil {
		t.Fatal("expected non-nil password hash")
	}
	if err := hasher.ComparePassword(*adminUser.PasswordHash, adminPassword); err != nil {
		t.Errorf("password verification failed: %v", err)
	}

	// 3. Second execution (idempotent update, preserving ID)
	originalID := adminUser.ID
	newPassword := "AdminUpdatedPass2026!"
	os.Setenv("ADMIN_PASSWORD", newPassword)

	if err := database.BootstrapAdmin(ctx, db.DB, hasher, cfg); err != nil {
		t.Fatalf("failed to update admin: %v", err)
	}

	var updatedUser models.User
	if err := db.DB.Where("LOWER(email) = ?", adminEmail).First(&updatedUser).Error; err != nil {
		t.Fatalf("updated admin user not found: %v", err)
	}

	if updatedUser.ID != originalID {
		t.Errorf("expected user ID to be preserved (%s), got %s", originalID, updatedUser.ID)
	}
	if err := hasher.ComparePassword(*updatedUser.PasswordHash, newPassword); err != nil {
		t.Errorf("new password verification failed: %v", err)
	}
}
