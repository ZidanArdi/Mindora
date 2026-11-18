package middleware

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/pkg/jwt"
	"mindora/backend/pkg/response"
)

const (
	ContextUserID    = "userID"
	ContextUserEmail = "userEmail"
	ContextUserRole  = "userRole"
)

var (
	ErrUnauthorized       = errors.New("unauthorized: missing or invalid token")
	ErrForbidden          = errors.New("forbidden: insufficient permissions")
	ErrInvalidUserContext = errors.New("invalid user context")
)

// JWTAuthMiddleware verifies the Bearer JWT token from the Authorization header
// and populates the authenticated user's ID, email, and role into Fiber's context.
func JWTAuthMiddleware(tokenManager *jwt.TokenManager) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Authorization header is required",
			})
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Authorization format must be Bearer <token>",
			})
		}

		tokenStr := strings.TrimSpace(parts[1])
		if tokenStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Access token is empty",
			})
		}

		claims, err := tokenManager.ValidateAccessToken(tokenStr)
		if err != nil {
			if errors.Is(err, jwt.ErrExpiredToken) {
				return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
					Success: false,
					Message: "Access token has expired",
				})
			}
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Invalid or malformed access token",
			})
		}

		// Inject authenticated user context
		c.Locals(ContextUserID, claims.UserID)
		c.Locals(ContextUserEmail, claims.Email)
		c.Locals(ContextUserRole, claims.Role)

		return c.Next()
	}
}

// GetUserID extracts the authenticated user's UUID from Fiber context.
func GetUserID(c *fiber.Ctx) (uuid.UUID, error) {
	val := c.Locals(ContextUserID)
	if val == nil {
		return uuid.Nil, ErrInvalidUserContext
	}
	id, ok := val.(uuid.UUID)
	if !ok {
		return uuid.Nil, ErrInvalidUserContext
	}
	return id, nil
}

// GetUserEmail extracts the authenticated user's email string from Fiber context.
func GetUserEmail(c *fiber.Ctx) string {
	val := c.Locals(ContextUserEmail)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetUserRole extracts the authenticated user's role name from Fiber context.
func GetUserRole(c *fiber.Ctx) string {
	val := c.Locals(ContextUserRole)
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// RequireRole ensures that the authenticated user possesses at least one of the specified roles.
func RequireRole(allowedRoles ...string) fiber.Handler {
	roleSet := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		roleSet[strings.ToLower(r)] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		userRole := strings.ToLower(GetUserRole(c))
		if userRole == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Authentication required",
			})
		}

		if _, ok := roleSet[userRole]; !ok {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: "Access forbidden: insufficient role permissions",
			})
		}

		return c.Next()
	}
}

// RequirePermission verifies whether the user's role possesses the required (resource, action) permission
// directly against the database permissions model or pre-defined role catalog.
func RequirePermission(db *gorm.DB, resource, action string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userRole := strings.ToLower(GetUserRole(c))
		if userRole == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Authentication required",
			})
		}

		// Query database role_permissions join table
		var count int64
		err := db.WithContext(c.Context()).
			Table("role_permissions").
			Joins("JOIN roles ON roles.id = role_permissions.role_id").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("roles.name = ? AND permissions.resource = ? AND permissions.action = ?", userRole, resource, action).
			Count(&count).Error

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
				Success: false,
				Message: "Failed to evaluate permission: " + err.Error(),
			})
		}

		if count == 0 {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: "Access forbidden: missing required permission '" + resource + "." + action + "'",
			})
		}

		return c.Next()
	}
}

// RequireOwnership ensures a user can only access their own resources.
// In accordance with Mindora privacy rules, Admin cannot freely read private student productivity records.
func RequireOwnership(paramKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		currentUserID, err := GetUserID(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Authentication required",
			})
		}

		paramVal := c.Params(paramKey)
		if paramVal == "" {
			return c.Status(fiber.StatusBadRequest).JSON(response.Response{
				Success: false,
				Message: "Resource identifier parameter is required",
			})
		}

		targetUserID, err := uuid.Parse(paramVal)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(response.Response{
				Success: false,
				Message: "Invalid user UUID parameter format",
			})
		}

		// Strict ownership check: target user ID must match the token's authenticated subject
		if currentUserID != targetUserID {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: "Access forbidden: you do not own this resource",
			})
		}

		return c.Next()
	}
}
