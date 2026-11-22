package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
	"mindora/backend/internal/repositories"
	"mindora/backend/pkg/response"
)

type AdminHandler struct {
	userRepo repositories.UserRepository
	db       *gorm.DB
}

func NewAdminHandler(userRepo repositories.UserRepository, db *gorm.DB) *AdminHandler {
	return &AdminHandler{
		userRepo: userRepo,
		db:       db,
	}
}

// GetOverview returns administrative overview metrics and system status.
func (h *AdminHandler) GetOverview(c *fiber.Ctx) error {
	var totalUsers int64
	var activeUsers int64
	var activeSessions int64

	_ = h.db.Table("users").Count(&totalUsers).Error
	_ = h.db.Table("users").Where("status = ?", "active").Count(&activeUsers).Error
	_ = h.db.Table("auth_sessions").Where("revoked_at IS NULL AND expires_at > ?", time.Now().UTC()).Count(&activeSessions).Error

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Admin system overview retrieved successfully",
		Data: fiber.Map{
			"total_users":     totalUsers,
			"active_users":    activeUsers,
			"active_sessions": activeSessions,
			"system_status":   "operational",
			"database_status": "healthy",
			"timestamp":       time.Now().UTC(),
		},
	})
}

// ListUsers returns paginated student user accounts without private productivity data.
func (h *AdminHandler) ListUsers(c *fiber.Ctx) error {
	limitStr := c.Query("limit", "20")
	offsetStr := c.Query("offset", "0")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	users, total, err := h.userRepo.List(c.Context(), limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to list users: " + err.Error(),
		})
	}

	userSummaries := make([]fiber.Map, 0, len(users))
	for _, u := range users {
		userSummaries = append(userSummaries, fiber.Map{
			"id":         u.ID,
			"name":       u.Name,
			"email":      u.Email,
			"role":       u.Role.Name,
			"status":     u.Status,
			"created_at": u.CreatedAt,
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Users retrieved successfully",
		Data: fiber.Map{
			"total":  total,
			"limit":  limit,
			"offset": offset,
			"users":  userSummaries,
		},
	})
}
