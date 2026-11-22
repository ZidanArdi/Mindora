package handlers

import (
	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type DashboardHandler struct {
	dashboardService services.DashboardService
}

func NewDashboardHandler(dashboardService services.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetSummary retrieves comprehensive real-time productivity statistics for the authenticated user.
func (h *DashboardHandler) GetSummary(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	summary, err := h.dashboardService.GetSummary(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to aggregate dashboard summary: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Dashboard summary retrieved successfully", summary))
}
