package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/response"
)

type TaskHandler struct {
	taskService services.TaskService
}

func NewTaskHandler(taskService services.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create handles creating a new productivity task.
func (h *TaskHandler) Create(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req services.CreateTaskInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body"))
	}

	task, err := h.taskService.CreateTask(c.Context(), userID, req)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(response.Success("Task created successfully", task))
}

// List handles listing tasks with filtering, search, sorting, and pagination.
func (h *TaskHandler) List(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	filter := repositories.TaskFilter{
		Status:    c.Query("status"),
		Priority:  c.Query("priority"),
		Category:  c.Query("category"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by", "created_at"),
		SortOrder: c.Query("sort_order", "desc"),
		Limit:     limit,
		Offset:    (page - 1) * limit,
	}

	tasks, total, err := h.taskService.ListTasks(c.Context(), userID, filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Error("Failed to query tasks: " + err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Tasks retrieved successfully", fiber.Map{
		"tasks": tasks,
		"meta": fiber.Map{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	}))
}

// GetByID handles retrieving a single task by ID.
func (h *TaskHandler) GetByID(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid task UUID format"))
	}

	task, err := h.taskService.GetTask(c.Context(), taskID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Error("Task not found"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Task retrieved successfully", task))
}

// Update handles updating task fields.
func (h *TaskHandler) Update(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid task UUID format"))
	}

	var req services.UpdateTaskInput
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body"))
	}

	task, err := h.taskService.UpdateTask(c.Context(), taskID, userID, req)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Task updated successfully", task))
}

// UpdateStatus handles updating task status.
func (h *TaskHandler) UpdateStatus(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid task UUID format"))
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid request body"))
	}

	task, err := h.taskService.UpdateTaskStatus(c.Context(), taskID, userID, req.Status)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Error(err.Error()))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Task status updated successfully", task))
}

// Delete handles soft-deleting a task.
func (h *TaskHandler) Delete(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	idParam := c.Params("id")
	taskID, err := uuid.Parse(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Invalid task UUID format"))
	}

	if err := h.taskService.DeleteTask(c.Context(), taskID, userID); err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Error("Task not found or unauthorized"))
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Task deleted successfully", fiber.Map{
		"deleted_id": taskID,
		"deleted_at": time.Now().UTC(),
	}))
}
