package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"mindora/backend/internal/models"
)

// MindoraNamespace is the base UUID namespace used to deterministically generate system role and permission UUIDs.
var MindoraNamespace = uuid.MustParse("e07f7b58-86d3-4e89-b883-9366e60d3d5f")

// DeterministicRoleID generates a consistent UUID for a system role.
func DeterministicRoleID(roleName string) uuid.UUID {
	return uuid.NewSHA1(MindoraNamespace, []byte("role:"+roleName))
}

// DeterministicPermissionID generates a consistent UUID for a system permission.
func DeterministicPermissionID(resource, action string) uuid.UUID {
	return uuid.NewSHA1(MindoraNamespace, []byte("permission:"+resource+":"+action))
}

// PermissionDefinition defines metadata for a system permission.
type PermissionDefinition struct {
	Resource    string
	Action      string
	Description string
}

// AllSystemPermissions defines the complete catalog of 13 system permissions.
var AllSystemPermissions = []PermissionDefinition{
	{Resource: "profile", Action: "read", Description: "View own profile information"},
	{Resource: "profile", Action: "update", Description: "Update own profile information"},
	{Resource: "tasks", Action: "create", Description: "Create personal productivity tasks"},
	{Resource: "tasks", Action: "read", Description: "View personal productivity tasks"},
	{Resource: "tasks", Action: "update", Description: "Update personal productivity tasks"},
	{Resource: "tasks", Action: "delete", Description: "Delete personal productivity tasks"},
	{Resource: "users", Action: "create", Description: "Create student user accounts"},
	{Resource: "users", Action: "read", Description: "View student user accounts"},
	{Resource: "users", Action: "update", Description: "Update student user accounts"},
	{Resource: "users", Action: "deactivate", Description: "Deactivate student user accounts"},
	{Resource: "users", Action: "activate", Description: "Activate student user accounts"},
	{Resource: "system", Action: "monitor", Description: "Monitor system health and service status"},
	{Resource: "audit_logs", Action: "read", Description: "View security audit and system activity logs"},
}

// UserPermissions defines permissions assigned to the 'user' role.
var UserPermissions = []string{
	"profile.read",
	"profile.update",
	"tasks.create",
	"tasks.read",
	"tasks.update",
	"tasks.delete",
}

// AdminPermissions defines permissions assigned to the 'admin' role.
var AdminPermissions = []string{
	"users.create",
	"users.read",
	"users.update",
	"users.deactivate",
	"users.activate",
	"system.monitor",
	"audit_logs.read",
}

// SeedRBAC executes an idempotent database seeder for system roles and permissions.
func SeedRBAC(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()

		// 1. Seed Roles
		userDesc := "Standard student user with personal productivity access"
		adminDesc := "System administrator with user management and monitoring access"

		roles := []models.Role{
			{
				ID:          DeterministicRoleID("user"),
				Name:        "user",
				Description: &userDesc,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
			{
				ID:          DeterministicRoleID("admin"),
				Name:        "admin",
				Description: &adminDesc,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		}

		for _, r := range roles {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "name"}},
				DoUpdates: clause.AssignmentColumns([]string{"description", "updated_at"}),
			}).Create(&r).Error; err != nil {
				return fmt.Errorf("failed to seed role %s: %w", r.Name, err)
			}
		}

		// 2. Seed Permissions
		permMap := make(map[string]uuid.UUID)
		for _, pDef := range AllSystemPermissions {
			desc := pDef.Description
			permID := DeterministicPermissionID(pDef.Resource, pDef.Action)
			code := pDef.Resource + "." + pDef.Action
			permMap[code] = permID

			perm := models.Permission{
				ID:          permID,
				Resource:    pDef.Resource,
				Action:      pDef.Action,
				Description: &desc,
				CreatedAt:   now,
			}

			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "resource"}, {Name: "action"}},
				DoUpdates: clause.AssignmentColumns([]string{"description"}),
			}).Create(&perm).Error; err != nil {
				return fmt.Errorf("failed to seed permission %s: %w", code, err)
			}
		}

		// 3. Assign Permissions to Roles
		userRoleID := DeterministicRoleID("user")
		adminRoleID := DeterministicRoleID("admin")

		var rolePerms []models.RolePermission

		for _, code := range UserPermissions {
			pID, exists := permMap[code]
			if !exists {
				return fmt.Errorf("user permission code not found in catalog: %s", code)
			}
			rolePerms = append(rolePerms, models.RolePermission{
				RoleID:       userRoleID,
				PermissionID: pID,
				CreatedAt:    now,
			})
		}

		for _, code := range AdminPermissions {
			pID, exists := permMap[code]
			if !exists {
				return fmt.Errorf("admin permission code not found in catalog: %s", code)
			}
			rolePerms = append(rolePerms, models.RolePermission{
				RoleID:       adminRoleID,
				PermissionID: pID,
				CreatedAt:    now,
			})
		}

		for _, rp := range rolePerms {
			if err := tx.Clauses(clause.OnConflict{
				DoNothing: true,
			}).Create(&rp).Error; err != nil {
				return fmt.Errorf("failed to assign permission %s to role %s: %w",
					rp.PermissionID, rp.RoleID, err)
			}
		}

		log.Println("RBAC seeder completed successfully: 2 roles, 13 permissions, 13 assignments verified.")
		return nil
	})
}
