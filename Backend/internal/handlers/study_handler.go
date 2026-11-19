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

type StudyHandler struct {
	studyService services.StudyService
}

func NewStudyHandler(studyService services.StudyService) *StudyHandler {
	return &StudyHandler{studyService: studyService}
}

// Start handles starting a new study session.
func (h *StudyHandler) Start(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req services.StartStudyInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body"))
	}

	session, err := h.studyService.StartSession(c.Context(), userID, req)
	if err != nil {
		if err == services.ErrActiveStudySessionExists {
			return c.Status(fiber.StatusConflict).JSON(response.Error(err.Error()))
		}
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Study session started successfully", session))
}

// Pause handles pausing an active study session.
func (h *StudyHandler) Pause(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid session UUID"))
	}

	session, err := h.studyService.PauseSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study session paused successfully", session))
}

// Resume handles resuming a paused study session.
func (h *StudyHandler) Resume(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid session UUID"))
	}

	session, err := h.studyService.ResumeSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study session resumed successfully", session))
}

// Finish handles completing a study session.
func (h *StudyHandler) Finish(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid session UUID"))
	}

	session, err := h.studyService.FinishSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study session completed successfully", session))
}

// Cancel handles cancelling a study session.
func (h *StudyHandler) Cancel(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	sessionID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid session UUID"))
	}

	session, err := h.studyService.CancelSession(c.Context(), sessionID, userID)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study session cancelled", session))
}

// GetActive retrieves the current active or paused session for state restoration.
func (h *StudyHandler) GetActive(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	session, err := h.studyService.GetActiveSession(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to check active session"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Active study session retrieved", session))
}

// List retrieves paginated study session history.
func (h *StudyHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	var startDate, endDate *time.Time
	if startQuery := c.Query("start_date"); startQuery != "" {
		if t, err := time.Parse("2006-01-02", startQuery); err == nil {
			startDate = &t
		}
	}
	if endQuery := c.Query("end_date"); endQuery != "" {
		if t, err := time.Parse("2006-01-02", endQuery); err == nil {
			// end of day
			tEnd := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, time.UTC)
			endDate = &tEnd
		}
	}

	sessions, total, err := h.studyService.ListSessions(c.Context(), userID, startDate, endDate, page, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to list study sessions"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study sessions retrieved successfully", fiber.Map{
		"sessions": sessions,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}))
}

// GetDailyStats retrieves study statistics for a specific date.
func (h *StudyHandler) GetDailyStats(c *fiber.Ctx) error {
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

	stats, err := h.studyService.GetDailyStats(c.Context(), userID, targetDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to retrieve study statistics"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Study statistics retrieved successfully", stats))
}
