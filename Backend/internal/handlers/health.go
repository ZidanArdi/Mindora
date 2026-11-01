package handlers

import (
	"github.com/gofiber/fiber/v2"
	"mindora/backend/pkg/response"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) Check(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Mindora API is running",
	})
}
