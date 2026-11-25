package turnstile

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	ErrTokenRequired      = errors.New("turnstile token is required")
	ErrVerificationFailed = errors.New("turnstile verification failed")
	ErrHostnameMismatch   = errors.New("turnstile hostname mismatch")
	ErrActionMismatch     = errors.New("turnstile action mismatch")
)

// IsTurnstileError checks if an error was caused by a Turnstile verification failure.
func IsTurnstileError(err error) bool {
	return errors.Is(err, ErrTokenRequired) ||
		errors.Is(err, ErrVerificationFailed) ||
		errors.Is(err, ErrHostnameMismatch) ||
		errors.Is(err, ErrActionMismatch)
}

// Verifier defines the interface for verifying Cloudflare Turnstile tokens.
type Verifier interface {
	Verify(ctx context.Context, token, remoteIP, action string) error
	IsEnabled() bool
}

// Config holds Cloudflare Turnstile verification settings.
type Config struct {
	Enabled          bool
	SecretKey        string
	ExpectedHostname string
	ExpectedAction   string
	Timeout          time.Duration
	AllowTestTokens  bool
}

// SiteverifyResponse represents Cloudflare's verification response.
type SiteverifyResponse struct {
	Success     bool      `json:"success"`
	ChallengeTS time.Time `json:"challenge_ts"`
	Hostname    string    `json:"hostname"`
	ErrorCodes  []string  `json:"error-codes"`
	Action      string    `json:"action"`
	CData       string    `json:"cdata"`
}

// Client implements Verifier using Cloudflare Siteverify API.
type Client struct {
	cfg        Config
	httpClient *http.Client
	verifyURL  string
}

// NewClient initializes a Turnstile verification client.
func NewClient(cfg Config) *Client {
	return NewClientWithURL(cfg, "https://challenges.cloudflare.com/turnstile/v0/siteverify")
}

// NewClientWithURL initializes a client with a custom siteverify URL (useful for testing).
func NewClientWithURL(cfg Config, url string) *Client {
	if url == "" {
		url = "https://challenges.cloudflare.com/turnstile/v0/siteverify"
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		verifyURL: url,
	}
}

// IsEnabled indicates if Turnstile enforcement is enabled.
func (c *Client) IsEnabled() bool {
	return c.cfg.Enabled
}

// Verify verifies the Turnstile challenge response token against Cloudflare.
func (c *Client) Verify(ctx context.Context, token, remoteIP, action string) error {
	// If Turnstile is explicitly disabled (e.g. dev/test mode without credentials)
	if !c.cfg.Enabled {
		return nil
	}

	if strings.TrimSpace(token) == "" {
		return ErrTokenRequired
	}

	// Cloudflare standard test dummy tokens pass strictly in non-production when explicitly allowed
	if c.cfg.AllowTestTokens && (token == "1x0000000000000000000000000000000AA" || token == "dummy-test-token") {
		return nil
	}

	formData := url.Values{}
	formData.Set("secret", c.cfg.SecretKey)
	formData.Set("response", token)
	if remoteIP != "" {
		formData.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.verifyURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return fmt.Errorf("failed to build turnstile request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: failed to contact turnstile service: %v", ErrVerificationFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w: status %d", ErrVerificationFailed, resp.StatusCode)
	}

	var siteResp SiteverifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&siteResp); err != nil {
		return fmt.Errorf("%w: failed to decode turnstile response: %v", ErrVerificationFailed, err)
	}

	if !siteResp.Success {
		return fmt.Errorf("%w: %s", ErrVerificationFailed, strings.Join(siteResp.ErrorCodes, ", "))
	}

	if c.cfg.ExpectedHostname != "" && siteResp.Hostname != "" {
		if !strings.EqualFold(siteResp.Hostname, c.cfg.ExpectedHostname) {
			return fmt.Errorf("%w: expected %s, got %s", ErrHostnameMismatch, c.cfg.ExpectedHostname, siteResp.Hostname)
		}
	}

	if action != "" && siteResp.Action != "" {
		if siteResp.Action != action {
			return fmt.Errorf("%w: expected %s, got %s", ErrActionMismatch, action, siteResp.Action)
		}
	}

	return nil
}
