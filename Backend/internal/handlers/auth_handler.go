package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"mindora/backend/internal/config"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/oauth"
	"mindora/backend/pkg/response"
	"mindora/backend/pkg/turnstile"
)

type AuthHandler struct {
	authService   services.AuthService
	googleService *oauth.GoogleService
	turnstile     turnstile.Verifier
	cfg           *config.Config
}

func NewAuthHandler(
	authService services.AuthService,
	googleService *oauth.GoogleService,
	turnstileVerifier turnstile.Verifier,
	cfg *config.Config,
) *AuthHandler {
	return &AuthHandler{
		authService:   authService,
		googleService: googleService,
		turnstile:     turnstileVerifier,
		cfg:           cfg,
	}
}

// Register handles user registration via name, email, and password.
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req services.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid request payload format",
		})
	}

	remoteIP := c.IP()
	result, err := h.authService.Register(c.Context(), req, remoteIP)
	if err != nil {
		if errors.Is(err, services.ErrEmailAlreadyRegistered) {
			return c.Status(fiber.StatusConflict).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if errors.Is(err, turnstile.ErrVerificationFailed) || errors.Is(err, turnstile.ErrTokenRequired) {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if strings.HasPrefix(err.Error(), "validation error") {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to complete registration: " + err.Error(),
		})
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	// Don't expose refresh token in JSON body if delivered via HttpOnly cookie
	cleanResult := *result
	cleanResult.RefreshToken = ""

	return c.Status(fiber.StatusCreated).JSON(response.Response{
		Success: true,
		Message: "Registration completed successfully",
		Data:    cleanResult,
	})
}

// Login handles user login via email and password.
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req services.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Invalid request payload format",
		})
	}

	remoteIP := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	result, err := h.authService.Login(c.Context(), req, remoteIP, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrInvalidCredentials) {
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: "Invalid email or password",
			})
		}
		if errors.Is(err, services.ErrAccountNotActive) || errors.Is(err, services.ErrAccountSuspended) {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if errors.Is(err, turnstile.ErrVerificationFailed) || errors.Is(err, turnstile.ErrTokenRequired) {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if strings.HasPrefix(err.Error(), "validation error") {
			return c.Status(fiber.StatusUnprocessableEntity).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Login processing failed",
		})
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	cleanResult := *result
	cleanResult.RefreshToken = ""

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Login successful",
		Data:    cleanResult,
	})
}

// Refresh handles refresh token rotation and generates a replacement access token.
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	// Attempt to extract refresh token from HttpOnly cookie first, then fallback to JSON body or Header
	rawRefreshToken := c.Cookies("refresh_token")
	var body struct {
		RefreshToken string `json:"refresh_token"`
		OperationID  string `json:"operation_id"`
	}
	_ = c.BodyParser(&body)

	if rawRefreshToken == "" {
		rawRefreshToken = body.RefreshToken
	}

	if rawRefreshToken == "" {
		authHeader := c.Get(fiber.HeaderAuthorization)
		if strings.HasPrefix(authHeader, "Bearer ") {
			rawRefreshToken = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	operationID := c.Get("X-Operation-ID")
	if operationID == "" {
		operationID = body.OperationID
	}

	remoteIP := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	result, err := h.authService.Refresh(c.Context(), rawRefreshToken, operationID, remoteIP, userAgent)
	if err != nil {
		if errors.Is(err, services.ErrInvalidRefreshToken) || errors.Is(err, services.ErrTokenReuseDetected) {
			h.clearRefreshTokenCookie(c)
			return c.Status(fiber.StatusUnauthorized).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if errors.Is(err, services.ErrConcurrentConflict) {
			// Do NOT clear cookie on concurrent conflict; return 409 Conflict so client can retry with current token
			return c.Status(fiber.StatusConflict).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if errors.Is(err, services.ErrAccountNotActive) {
			h.clearRefreshTokenCookie(c)
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Token refresh failed",
		})
	}

	h.setRefreshTokenCookie(c, result.RefreshToken)

	cleanResult := *result
	cleanResult.RefreshToken = ""

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    cleanResult,
	})
}

// Logout revokes the current authentication session.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	rawRefreshToken := c.Cookies("refresh_token")
	if rawRefreshToken == "" {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = c.BodyParser(&body)
		rawRefreshToken = body.RefreshToken
	}

	_ = h.authService.Logout(c.Context(), rawRefreshToken)
	h.clearRefreshTokenCookie(c)

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Logged out successfully",
	})
}

// GoogleAuth initiates the Google OAuth 2.0 Authorization Code flow.
func (h *AuthHandler) GoogleAuth(c *fiber.Ctx) error {
	if !h.googleService.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Google OAuth is disabled or not configured in this environment",
		})
	}

	// Turnstile verification for Google initiation if token is provided
	turnstileToken := c.Query("turnstile_token")
	if turnstileToken != "" {
		if err := h.turnstile.Verify(c.Context(), turnstileToken, c.IP(), "google_auth"); err != nil {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: "Turnstile verification failed: " + err.Error(),
			})
		}
	}

	state, err := oauth.GenerateRandomString(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to generate oauth state",
		})
	}

	nonce, err := oauth.GenerateRandomString(32)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to generate oauth nonce",
		})
	}

	verifier, challenge, err := oauth.GeneratePKCE()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: "Failed to generate PKCE challenge",
		})
	}

	// Store state, nonce, and PKCE code_verifier in temporary HttpOnly cookies (10 minutes lifetime)
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/api/v1/auth/google",
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_nonce",
		Value:    nonce,
		Path:     "/api/v1/auth/google",
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "oauth_code_verifier",
		Value:    verifier,
		Path:     "/api/v1/auth/google",
		Expires:  time.Now().Add(10 * time.Minute),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})

	authURL, err := h.googleService.GetAuthURL(state, nonce, challenge)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(response.Response{
			Success: false,
			Message: err.Error(),
		})
	}

	// If request accepts JSON or explicitly asks for URL
	if strings.Contains(c.Get("Accept"), "application/json") || c.Query("redirect") == "false" {
		return c.Status(fiber.StatusOK).JSON(response.Response{
			Success: true,
			Message: "Google OAuth authorization URL generated",
			Data: map[string]string{
				"auth_url": authURL,
			},
		})
	}

	// Otherwise standard browser redirect
	return c.Redirect(authURL, http.StatusTemporaryRedirect)
}

// GoogleCallback processes the Google OAuth authorization code exchange.
func (h *AuthHandler) GoogleCallback(c *fiber.Ctx) error {
	if !h.googleService.IsEnabled() {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Response{
			Success: false,
			Message: "Google OAuth is disabled or not configured in this environment",
		})
	}

	code := c.Query("code")
	state := c.Query("state")
	expectedState := c.Cookies("oauth_state")
	expectedNonce := c.Cookies("oauth_nonce")
	codeVerifier := c.Cookies("oauth_code_verifier")

	// Clear temporary OAuth cookies immediately
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/api/v1/auth/google",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_nonce",
		Value:    "",
		Path:     "/api/v1/auth/google",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})
	c.Cookie(&fiber.Cookie{
		Name:     "oauth_code_verifier",
		Value:    "",
		Path:     "/api/v1/auth/google",
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		HTTPOnly: true,
		Secure:   h.cfg.Cookie.Secure,
		SameSite: "Lax",
	})

	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Authorization code missing from Google callback",
		})
	}

	remoteIP := c.IP()
	userAgent := c.Get(fiber.HeaderUserAgent)

	result, err := h.authService.HandleGoogleCallback(
		c.Context(),
		code,
		state,
		expectedState,
		expectedNonce,
		codeVerifier,
		remoteIP,
		userAgent,
	)

	if err != nil {
		if errors.Is(err, oauth.ErrInvalidState) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Response{
				Success: false,
				Message: "Invalid OAuth state parameter (possible CSRF attempt)",
			})
		}
		if errors.Is(err, oauth.ErrInvalidPKCE) {
			return c.Status(fiber.StatusBadRequest).JSON(response.Response{
				Success: false,
				Message: "Invalid or missing OAuth PKCE verifier (possible replay attempt)",
			})
		}
		if errors.Is(err, services.ErrEmailConflictExternalIdentity) {
			return c.Status(fiber.StatusConflict).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		if errors.Is(err, services.ErrAccountNotActive) || errors.Is(err, services.ErrAccountSuspended) {
			return c.Status(fiber.StatusForbidden).JSON(response.Response{
				Success: false,
				Message: err.Error(),
			})
		}
		return c.Status(fiber.StatusBadRequest).JSON(response.Response{
			Success: false,
			Message: "Google authentication failed: " + err.Error(),
		})
	}

	// Apply anti-caching and referrer security headers
	c.Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	c.Set("Pragma", "no-cache")
	c.Set("Referrer-Policy", "strict-origin-when-cross-origin")

	h.setRefreshTokenCookie(c, result.RefreshToken)

	// If browser redirect requested or HTML accepted
	if h.cfg.FrontendURL != "" && strings.Contains(c.Get("Accept"), "text/html") {
		// SECURE REDIRECT: Tokens are NEVER passed in URL parameters!
		// The browser receives the HttpOnly refresh token cookie and redirects cleanly.
		redirectURL := h.cfg.FrontendURL + "/login?auth=success"
		return c.Redirect(redirectURL, http.StatusTemporaryRedirect)
	}

	cleanResult := *result
	cleanResult.RefreshToken = ""

	return c.Status(fiber.StatusOK).JSON(response.Response{
		Success: true,
		Message: "Google authentication successful",
		Data:    cleanResult,
	})
}

func (h *AuthHandler) setRefreshTokenCookie(c *fiber.Ctx, rawRefreshToken string) {
	if rawRefreshToken == "" {
		return
	}
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    rawRefreshToken,
		Path:     "/api/v1/auth",
		Domain:   h.cfg.Cookie.Domain,
		MaxAge:   int(h.cfg.JWT.RefreshExpiry.Seconds()),
		Expires:  time.Now().Add(h.cfg.JWT.RefreshExpiry),
		Secure:   h.cfg.Cookie.Secure,
		HTTPOnly: true,
		SameSite: h.cfg.Cookie.SameSite,
	})
}

func (h *AuthHandler) clearRefreshTokenCookie(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/api/v1/auth",
		Domain:   h.cfg.Cookie.Domain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-1 * time.Hour),
		Secure:   h.cfg.Cookie.Secure,
		HTTPOnly: true,
		SameSite: h.cfg.Cookie.SameSite,
	})
}
