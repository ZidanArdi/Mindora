package database_test

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/config"
	"mindora/backend/internal/database"
	"mindora/backend/internal/models"
)

func TestConstraintsAndIntegrity(t *testing.T) {
	cfg := config.LoadConfig()
	db, err := database.Connect(cfg)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Execute tests in a transaction and ALWAYS ROLLBACK at end
	tx := db.DB.Begin()
	defer tx.Rollback()

	// 1. Fetch user role ID
	var userRole models.Role
	if err := tx.Where("name = ?", "user").First(&userRole).Error; err != nil {
		t.Fatalf("user role not found: %v", err)
	}

	testPwdHash := "$2a$12$eX4mP1eH4sHvaL1dP4ssW0rdH4sh"

	// 2. Insert valid user
	validUser := models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         "Audit Test User",
		Email:        "audit.test@mindora.app",
		PasswordHash: &testPwdHash,
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if err := tx.Create(&validUser).Error; err != nil {
		t.Fatalf("failed to insert valid user: %v", err)
	}

	// 3. Test Invalid Status Check Constraint
	invalidStatusUser := models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         "Invalid Status User",
		Email:        "invalid.status@mindora.app",
		PasswordHash: &testPwdHash,
		Status:       "banned", // Not allowed by CHECK constraint
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	// Use sub-transaction (savepoint) to isolate expected failure
	err = tx.Transaction(func(subTx *gorm.DB) error {
		return subTx.Create(&invalidStatusUser).Error
	})
	if err == nil {
		t.Errorf("expected error when inserting user with invalid status 'banned', got nil")
	} else if !strings.Contains(err.Error(), "chk_users_status") {
		t.Logf("check constraint triggered as expected: %v", err)
	}

	// 4. Test Case-Insensitive Email Uniqueness
	duplicateEmailUser := models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         "Duplicate Email User",
		Email:        "AUDIT.TEST@MINDORA.APP", // Uppercase variant
		PasswordHash: &testPwdHash,
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	err = tx.Transaction(func(subTx *gorm.DB) error {
		return subTx.Create(&duplicateEmailUser).Error
	})
	if err == nil {
		t.Errorf("expected unique constraint violation on case-insensitive email, got nil")
	} else if !strings.Contains(err.Error(), "idx_users_email_lower") {
		t.Logf("case-insensitive email uniqueness triggered as expected: %v", err)
	}

	// 5. Test Soft Delete & Email Reservation Policy
	// Soft delete the valid user
	if err := tx.Delete(&validUser).Error; err != nil {
		t.Fatalf("failed to soft delete user: %v", err)
	}

	// Verify user is soft-deleted (not returned in normal queries)
	var lookupUser models.User
	if err := tx.Where("id = ?", validUser.ID).First(&lookupUser).Error; err == nil {
		t.Errorf("expected soft-deleted user to be omitted from standard query, but found record")
	}

	// Verify soft-deleted user is still present with Unscoped()
	var unscopedUser models.User
	if err := tx.Unscoped().Where("id = ?", validUser.ID).First(&unscopedUser).Error; err != nil {
		t.Errorf("expected soft-deleted user to be retrievable via Unscoped(), got: %v", err)
	}

	// Verify email reservation: attempting to create a new user with the same email MUST FAIL
	reusedEmailUser := models.User{
		ID:           uuid.New(),
		RoleID:       userRole.ID,
		Name:         "Reused Email User",
		Email:        "audit.test@mindora.app", // Same email as soft-deleted user
		PasswordHash: &testPwdHash,
		Status:       models.UserStatusActive,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	err = tx.Transaction(func(subTx *gorm.DB) error {
		return subTx.Create(&reusedEmailUser).Error
	})
	if err == nil {
		t.Errorf("expected error: email reservation policy violated (deleted user's email was reused), got nil")
	} else {
		t.Logf("email reservation policy successfully enforced on soft-deleted account: %v", err)
	}

	// 6. Test Foreign Key Restriction on Role Deletion
	err = tx.Transaction(func(subTx *gorm.DB) error {
		return subTx.Delete(&userRole).Error
	})
	if err == nil {
		t.Errorf("expected foreign key restrict error when deleting role referenced by user, got nil")
	} else {
		t.Logf("foreign key restrict correctly prevented role deletion: %v", err)
	}
}
