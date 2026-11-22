package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type LifestyleHandler struct {
	lifestyleService services.LifestyleService
}

func NewLifestyleHandler(lifestyleService services.LifestyleService) *LifestyleHandler {
	return &LifestyleHandler{lifestyleService: lifestyleService}
}

// GetRecord retrieves the daily lifestyle record for the specified or current date.
func (h *LifestyleHandler) GetRecord(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	targetDate := time.Now().UTC()
	if dateParam := c.Query("date"); dateParam != "" {
		if t, err := time.Parse("2006-01-02", dateParam); err == nil {
			targetDate = t
		} else {
			return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid date format (must be YYYY-MM-DD)"))
		}
	}

	record, err := h.lifestyleService.GetRecordByDate(c.Context(), userID, targetDate)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(response.Success("No lifestyle record found for date", nil))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Lifestyle record retrieved successfully", record))
}

// UpsertRecord creates or updates the daily lifestyle record for a calendar date.
func (h *LifestyleHandler) UpsertRecord(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req services.UpsertLifestyleInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body format"))
	}

	if req.RecordDate == "" {
		req.RecordDate = time.Now().UTC().Format("2006-01-02")
	}

	record, err := h.lifestyleService.UpsertRecord(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Lifestyle record saved successfully", record))
}

// ListHistory retrieves paginated lifestyle logs.
func (h *LifestyleHandler) ListHistory(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	records, total, err := h.lifestyleService.ListHistory(c.Context(), userID, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to query lifestyle history"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Lifestyle history retrieved successfully", fiber.Map{
		"records": records,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}))
}
