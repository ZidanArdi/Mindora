package handlers

import (
	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type AnalyticsHandler struct {
	analyticsService services.AnalyticsService
}

func NewAnalyticsHandler(analyticsService services.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{
		analyticsService: analyticsService,
	}
}

// GetDaily returns daily activity metrics, duration breakdowns, and lifestyle logs.
// GET /api/v1/analytics/daily or /api/analytics/daily
func (h *AnalyticsHandler) GetDaily(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var datePtr *string
	dateParam := c.Query("date")
	if dateParam != "" {
		datePtr = &dateParam
	}

	var tzPtr *string
	tzParam := c.Query("tz", c.Query("timezone"))
	if tzParam != "" {
		tzPtr = &tzParam
	}

	analytics, err := h.analyticsService.GetDailyAnalytics(c.Context(), userID, datePtr, tzPtr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid date format, expected YYYY-MM-DD: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Daily productivity analytics retrieved successfully", analytics))
}

// GetWeekly returns 7-day rolling or calendar weekly productivity metrics and trends.
// GET /api/v1/analytics/weekly or /api/analytics/weekly
func (h *AnalyticsHandler) GetWeekly(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var endDatePtr *string
	endParam := c.Query("end_date", c.Query("date"))
	if endParam != "" {
		endDatePtr = &endParam
	}

	var tzPtr *string
	tzParam := c.Query("tz", c.Query("timezone"))
	if tzParam != "" {
		tzPtr = &tzParam
	}

	analytics, err := h.analyticsService.GetWeeklyAnalytics(c.Context(), userID, endDatePtr, tzPtr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid end_date format, expected YYYY-MM-DD: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Weekly productivity analytics retrieved successfully", analytics))
}

// GetMonthly returns 30-day productivity trends, study consistency, and habit correlations.
// GET /api/v1/analytics/monthly or /api/analytics/monthly
func (h *AnalyticsHandler) GetMonthly(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var endDatePtr *string
	endParam := c.Query("end_date", c.Query("date"))
	if endParam != "" {
		endDatePtr = &endParam
	}

	var tzPtr *string
	tzParam := c.Query("tz", c.Query("timezone"))
	if tzParam != "" {
		tzPtr = &tzParam
	}

	analytics, err := h.analyticsService.GetMonthlyAnalytics(c.Context(), userID, endDatePtr, tzPtr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid end_date format, expected YYYY-MM-DD: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Monthly productivity analytics retrieved successfully", analytics))
}

// GetSummary returns high-level quick overview metrics.
// GET /api/v1/analytics/summary or /api/analytics/summary
func (h *AnalyticsHandler) GetSummary(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var tzPtr *string
	tzParam := c.Query("tz", c.Query("timezone"))
	if tzParam != "" {
		tzPtr = &tzParam
	}

	summary, err := h.analyticsService.GetAnalyticsSummary(c.Context(), userID, tzPtr)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to compute analytics summary: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Productivity analytics summary retrieved successfully", summary))
}
