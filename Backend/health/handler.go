package health

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"mindora/backend/shared/response"
)

type HealthHandler struct {
	DB *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{
		DB: db,
	}
}

// Check verifies that the API is running.
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Mindora API is running",
	})
}

// CheckDatabase verifies the PostgreSQL connection.
func (h *HealthHandler) CheckDatabase(c *fiber.Ctx) error {
	sqlDB, err := h.DB.DB()
	if err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Failed to access database connection",
		})
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Database connection is unavailable",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Database connection is healthy",
	})
}
