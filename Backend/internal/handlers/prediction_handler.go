package handlers

import (
	"errors"
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type PredictionHandler struct {
	predictionService services.PredictionService
}

func NewPredictionHandler(predictionService services.PredictionService) *PredictionHandler {
	return &PredictionHandler{
		predictionService: predictionService,
	}
}

// Create triggers a new productivity score prediction for the authenticated student.
// POST /api/v1/predictions or /api/predictions
func (h *PredictionHandler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req services.PredictionRequest
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.Error("Malformed request body JSON"))
		}
	}

	pred, err := h.predictionService.Predict(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, services.ErrInvalidPredictionInput) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid input parameter range for prediction features"))
		}
		log.Printf("[ERROR] [PredictionHandler.Create] user_id=%s failed to compute productivity prediction: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to compute productivity prediction"))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Productivity prediction generated successfully", pred))
}

// List retrieves paginated prediction history for the authenticated student.
// GET /api/v1/predictions or /api/predictions
func (h *PredictionHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	history, total, err := h.predictionService.GetHistory(c.Context(), userID, page, limit)
	if err != nil {
		log.Printf("[ERROR] [PredictionHandler.List] user_id=%s failed to fetch prediction history: %v", userID, err)
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to fetch prediction history"))
	}

	totalPages := int((total + int64(limit) - 1) / int64(limit))

	return c.Status(fiber.StatusOK).JSON(response.Success("Prediction history retrieved successfully", fiber.Map{
		"items":         history,
		"total":         total,
		"page":          page,
		"limit":         limit,
		"total_pages":   totalPages,
		"model_version": services.ModelVersionCurrent,
		"disclaimer":    services.ModelDisclaimerText,
	}))
}

// GetByID retrieves a specific prediction by UUID, enforcing strict user ownership.
// GET /api/v1/predictions/:id or /api/predictions/:id
func (h *PredictionHandler) GetByID(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	idParam := c.Params("id")
	predID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid prediction ID UUID format"))
	}

	record, err := h.predictionService.GetByID(c.Context(), predID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Error("Prediction not found or access unauthorized"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Prediction retrieved successfully", record))
}
