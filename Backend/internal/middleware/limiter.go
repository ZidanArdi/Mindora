package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"mindora/backend/pkg/response"
)

// AuthRateLimiter returns an IP-based rate limiting middleware for sensitive authentication endpoints.
func AuthRateLimiter(maxRequests int, window time.Duration) fiber.Handler {
	if maxRequests <= 0 {
		maxRequests = 10
	}
	if window <= 0 {
		window = 1 * time.Minute
	}

	return limiter.New(limiter.Config{
		Max:        maxRequests,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(response.Response{
				Success: false,
				Message: "Too many authentication attempts; please wait before retrying",
			})
		},
	})
}
