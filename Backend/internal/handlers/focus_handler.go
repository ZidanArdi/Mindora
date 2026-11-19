package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type FocusHandler struct {
	focusService services.FocusService
}

func NewFocusHandler(focusService services.FocusService) *FocusHandler {
	return &FocusHandler{focusService: focusService}
}

// Start handles beginning a new focus or break interval.
func (h *FocusHandler) Start(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req services.StartFocusInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body"))
	}

	interval, err := h.focusService.StartInterval(c.Context(), userID, req)
	if err != nil {
		if err == services.ErrActiveFocusIntervalExists {
			return c.Status(fiber.StatusConflict).JSON(response.Error(err.Error()))
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Focus interval started successfully", interval))
}

// RecordInterruption handles logging an interruption on the active interval.
func (h *FocusHandler) RecordInterruption(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	intervalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid interval UUID"))
	}

	interval, err := h.focusService.RecordInterruption(c.Context(), intervalID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Interruption logged", interval))
}

// Finish handles completing a focus or break interval.
func (h *FocusHandler) Finish(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	intervalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid interval UUID"))
	}

	interval, err := h.focusService.FinishInterval(c.Context(), intervalID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Interval completed successfully", interval))
}

// Abandon handles abandoning an active interval.
func (h *FocusHandler) Abandon(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	intervalID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid interval UUID"))
	}

	interval, err := h.focusService.AbandonInterval(c.Context(), intervalID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Interval marked as abandoned", interval))
}

// GetActive retrieves the current active focus or break interval.
func (h *FocusHandler) GetActive(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	interval, err := h.focusService.GetActiveInterval(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to check active interval"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Active focus interval retrieved", interval))
}

// List handles listing completed focus/break history.
func (h *FocusHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	sessionType := c.Query("session_type") // optional filter: "focus" or "break"

	intervals, total, err := h.focusService.ListHistory(c.Context(), userID, sessionType, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to list focus history"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Focus history retrieved successfully", fiber.Map{
		"intervals": intervals,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}))
}

// GetDailyStats retrieves focus statistics for a specific date.
func (h *FocusHandler) GetDailyStats(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	targetDate := time.Now().UTC()
	if dateParam := c.Query("date"); dateParam != "" {
		if t, err := time.Parse("2006-01-02", dateParam); err == nil {
			targetDate = t
		}
	}

	stats, err := h.focusService.GetDailyStats(c.Context(), userID, targetDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to retrieve focus statistics"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Focus statistics retrieved successfully", stats))
}
