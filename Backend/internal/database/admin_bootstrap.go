package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/config"
	"mindora/backend/internal/models"
	"mindora/backend/pkg/hash"
)

var (
	ErrAdminEmailRequired    = errors.New("ADMIN_EMAIL environment variable or -email flag is required")
	ErrAdminPasswordRequired = errors.New("ADMIN_PASSWORD environment variable or -password flag is required")
	ErrAdminPasswordTooShort = errors.New("ADMIN_PASSWORD must be at least 8 characters")
	ErrProductionRestricted  = errors.New("admin bootstrap command is strictly disabled in production environment")
	ErrPrivilegeEscalation   = errors.New("target account already exists with non-admin role; privilege escalation rejected")
)

// BootstrapAdmin provisions or updates an administrative user account for development and initial bootstrapping.
// It explicitly enforces security guards:
// - Rejects execution in production environments
// - Verifies the existence of the deterministic 'admin' role
// - Enforces minimum password strength requirements
// - Rejects privilege escalation if target email belongs to a non-admin account
// - Preserves existing admin account ID and marks email as verified
// - Never logs raw passwords or password hashes
func BootstrapAdmin(ctx context.Context, db *gorm.DB, hasher hash.Hasher, cfg *config.Config) error {
	// 1. Production safety guard
	if cfg != nil && strings.ToLower(cfg.Env) == "production" {
		return ErrProductionRestricted
	}
	if strings.ToLower(os.Getenv("APP_ENV")) == "production" {
		return ErrProductionRestricted
	}

	adminEmail := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_EMAIL")))
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	adminName := strings.TrimSpace(os.Getenv("ADMIN_NAME"))
	if adminName == "" {
		adminName = "Mindora Admin"
	}

	if adminEmail == "" {
		return ErrAdminEmailRequired
	}
	if adminPassword == "" {
		return ErrAdminPasswordRequired
	}
	if len(adminPassword) < 8 {
		return ErrAdminPasswordTooShort
	}

	if db == nil {
		return errors.New("database connection is required")
	}

	// 2. Locate the deterministic 'admin' role in the database
	var adminRole models.Role
	if err := db.WithContext(ctx).Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("admin role not found; please run database seeder first: %w", err)
		}
		return fmt.Errorf("failed to query admin role: %w", err)
	}

	// 3. Hash the admin password using standard bcrypt hasher
	pwdHash, err := hasher.HashPassword(adminPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	now := time.Now().UTC()

	// 4. Check if user already exists
	var existingUser models.User
	err = db.WithContext(ctx).Where("LOWER(email) = ?", adminEmail).First(&existingUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("failed to query existing user: %w", err)
	}

	if err == nil {
		// Existing account found: verify role to prevent privilege escalation
		if existingUser.RoleID != adminRole.ID {
			return fmt.Errorf("%w: user '%s' holds role_id %s (expected admin %s)",
				ErrPrivilegeEscalation, adminEmail, existingUser.RoleID, adminRole.ID)
		}

		// Update existing admin account while strictly preserving existing user ID
		existingUser.Name = adminName
		existingUser.PasswordHash = &pwdHash
		existingUser.Status = models.UserStatusActive
		existingUser.EmailVerifiedAt = &now
		existingUser.UpdatedAt = now

		if err := db.WithContext(ctx).Save(&existingUser).Error; err != nil {
			return fmt.Errorf("failed to update admin account: %w", err)
		}

		log.Printf("Admin account '%s' (ID: %s) securely updated with role 'admin'.\n", adminEmail, existingUser.ID)
		return nil
	}

	// 5. Account does not exist: create new admin account
	newAdmin := models.User{
		ID:              uuid.New(),
		RoleID:          adminRole.ID,
		Name:            adminName,
		Email:           adminEmail,
		PasswordHash:    &pwdHash,
		Status:          models.UserStatusActive,
		EmailVerifiedAt: &now,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := db.WithContext(ctx).Create(&newAdmin).Error; err != nil {
		return fmt.Errorf("failed to create admin account: %w", err)
	}

	log.Printf("Admin account '%s' (ID: %s) securely created with role 'admin'.\n", adminEmail, newAdmin.ID)
	return nil
}
