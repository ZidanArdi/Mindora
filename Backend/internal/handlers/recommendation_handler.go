package handlers

import (
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type RecommendationHandler struct {
	recService services.RecommendationService
}

func NewRecommendationHandler(recService services.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recService: recService}
}

// GetRecommendations handles GET /api/v1/recommendations
func (h *RecommendationHandler) GetRecommendations(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
			Success: false,
			Message: "Authentication required",
		})
	}

	userRole := middleware.GetUserRole(c)
	if !strings.EqualFold(userRole, "user") {
		return c.Status(fiber.StatusForbidden).JSON(response.Response{
			Success: false,
			Message: "Access forbidden: student recommendation streams are private",
		})
	}

	evaluateParam := strings.ToLower(c.Query("evaluate", "false"))
	evaluate := evaluateParam == "true" || evaluateParam == "1"

	items, tier, err := h.recService.GetRecommendations(c.Context(), userID, evaluate)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch recommendations for user %s: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to retrieve recommendations",
		})
	}

	if items == nil {
		items = []models.Recommendation{}
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Recommendations retrieved successfully",
		Data: fiber.Map{
			"items":          items,
			"total":          len(items),
			"user_data_tier": tier,
			"generated_at":   time.Now().UTC().Format(time.RFC3339),
		},
	})
}

type UpdateRecommendationStatusPayload struct {
	Status string `json:"status"`
}

// UpdateStatus handles PATCH /api/v1/recommendations/:id/status
func (h *RecommendationHandler) UpdateStatus(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
			Success: false,
			Message: "Authentication required",
		})
	}

	userRole := middleware.GetUserRole(c)
	if !strings.EqualFold(userRole, "user") {
		return c.Status(fiber.StatusForbidden).JSON(response.Response{
			Success: false,
			Message: "Access forbidden: student recommendation streams are private",
		})
	}

	idParam := c.Params("id")
	recID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid recommendation identifier format",
		})
	}

	var payload UpdateRecommendationStatusPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid request payload format",
		})
	}

	status := strings.ToLower(strings.TrimSpace(payload.Status))
	if status != models.RecStatusDismissed && status != models.RecStatusCompleted {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid status: only 'dismissed' and 'completed' are permitted",
		})
	}

	updatedRec, err := h.recService.UpdateStatus(c.Context(), recID, userID, status)
	if err != nil {
		if errors.Is(err, repositories.ErrRecommendationNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Response{
				Success: false,
				Message: "Recommendation not found",
			})
		}
		log.Printf("[ERROR] Failed to update recommendation %s for user %s: %v", recID, userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to update recommendation status",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Recommendation status updated successfully",
		Data:    updatedRec,
	})
}
