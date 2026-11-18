package handlers

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/database"
	"mindora/backend/pkg/response"
)

type HealthHandler struct {
	db *database.Database
}

func NewHealthHandler(db *database.Database) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check returns the system operational status (preserves existing API contract).
func (h *HealthHandler) Check(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Mindora API is running",
	})
}

// DatabaseCheck verifies the PostgreSQL connection status.
func (h *HealthHandler) DatabaseCheck(c *fiber.Ctx) error {
	if h.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Database connection not configured",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Database connection unreachable",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Database is connected and healthy",
		Data: map[string]string{
			"database": "postgresql",
			"status":   "up",
		},
	})
}
