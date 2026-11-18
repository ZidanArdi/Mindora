package handlers

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/repositories"
	"mindora/backend/pkg/response"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}

// GetProfile retrieves the authenticated user's profile and granted permissions.
func (h *UserHandler) GetProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
			Success: false,
			Message: "Authentication required",
		})
	}

	user, err := h.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Response{
			Success: false,
			Message: "User profile not found",
		})
	}

	// Fetch permissions
	perms, err := h.userRepo.GetUserPermissions(c.Context(), user.RoleID)
	if err != nil {
		perms = []string{}
	}

	roleName := "user"
	if user.Role.Name != "" {
		roleName = user.Role.Name
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Profile retrieved successfully",
		Data: fiber.Map{
			"id":          user.ID,
			"name":        user.Name,
			"email":       user.Email,
			"role":        roleName,
			"status":      user.Status,
			"permissions": perms,
			"created_at":  user.CreatedAt,
		},
	})
}

// UpdateProfile updates the current user's profile information.
func (h *UserHandler) UpdateProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
			Success: false,
			Message: "Authentication required",
		})
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid request payload format",
		})
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Response{
			Success: false,
			Message: "Name cannot be empty",
		})
	}

	user, err := h.userRepo.FindByID(c.Context(), userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(response.Response{
			Success: false,
			Message: "User profile not found",
		})
	}

	user.Name = name
	if err := h.userRepo.Update(c.Context(), user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to update profile: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Profile updated successfully",
		Data: fiber.Map{
			"id":         user.ID,
			"name":       user.Name,
			"email":      user.Email,
			"role":       user.Role.Name,
			"status":     user.Status,
			"updated_at": user.UpdatedAt,
		},
	})
}

// GetTasks returns personal productivity tasks for the specified user ID.
// Ownership is enforced by RequireOwnership middleware to guarantee privacy.
func (h *UserHandler) GetTasks(c *fiber.Ctx) error {
	paramUserID := c.Params("userId")
	targetID, err := uuid.Parse(paramUserID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid user UUID format",
		})
	}

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Tasks retrieved successfully",
		Data: fiber.Map{
			"owner_id": targetID,
			"tasks": []fiber.Map{
				{
					"id":       "task-1",
					"title":    "Review Discrete Mathematics Lecture 4",
					"category": "Study",
					"status":   "in_progress",
				},
				{
					"id":       "task-2",
					"title":    "Complete Database System Lab Assignment",
					"category": "Lab",
					"status":   "pending",
				},
			},
		},
	})
}
