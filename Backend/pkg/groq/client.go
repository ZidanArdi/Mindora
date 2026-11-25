package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var (
	ErrMissingAPIKey      = errors.New("groq api key is required")
	ErrEmptyMessages      = errors.New("chat completion requires at least one message")
	ErrRateLimitExceeded  = errors.New("groq rate limit exceeded, please retry later")
	ErrUnauthorized       = errors.New("invalid or unauthorized groq api key")
	ErrServiceUnavailable = errors.New("groq service is temporarily unavailable")
)

// ChatMessage represents a single message in a chat conversation.
type ChatMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Reasoning string `json:"reasoning,omitempty"`
}

// ChatCompletionRequest encapsulates the payload sent to Groq's OpenAI-compatible chat API.
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
}

// ChatChoice represents an individual choice in the completion response.
type ChatChoice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// UsageStats records token consumption.
type UsageStats struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse models the response from Groq's /chat/completions endpoint.
type ChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []ChatChoice `json:"choices"`
	Usage   UsageStats   `json:"usage"`
}

// APIError models Groq's JSON error payload.
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// Client defines the contract for communicating with Groq LLM API.
type Client interface {
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)
	GetModel() string
}

type clientImpl struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a production-ready Groq API client using Go's standard net/http client.
func NewClient(apiKey, defaultModel, baseURL string, timeout time.Duration) Client {
	if defaultModel == "" {
		defaultModel = "openai/gpt-oss-20b"
	}
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &clientImpl{
		apiKey:  strings.TrimSpace(apiKey),
		model:   defaultModel,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// NewCustomClient allows injecting a custom HTTP transport/client (e.g. for testing).
func NewCustomClient(apiKey, defaultModel, baseURL string, httpClient *http.Client) Client {
	if defaultModel == "" {
		defaultModel = "openai/gpt-oss-20b"
	}
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &clientImpl{
		apiKey:     strings.TrimSpace(apiKey),
		model:      defaultModel,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

func (c *clientImpl) GetModel() string {
	return c.model
}

func (c *clientImpl) CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if c.apiKey == "" {
		return nil, ErrMissingAPIKey
	}
	if len(req.Messages) == 0 {
		return nil, ErrEmptyMessages
	}

	// Default model if not specified in request
	if req.Model == "" {
		req.Model = c.model
	}

	reqBodyBytes, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal completion request: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(reqBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "Mindora-Backend-LLM/1.0")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			return nil, ctx.Err()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") || strings.Contains(strings.ToLower(err.Error()), "deadline exceeded") {
			return nil, fmt.Errorf("groq request timed out: %w", context.DeadlineExceeded)
		}
		return nil, fmt.Errorf("network error communicating with groq: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read groq response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			return nil, ErrUnauthorized
		case http.StatusTooManyRequests:
			return nil, ErrRateLimitExceeded
		case http.StatusServiceUnavailable, http.StatusBadGateway, http.StatusGatewayTimeout:
			return nil, ErrServiceUnavailable
		default:
			var apiErr APIError
			if jsonErr := json.Unmarshal(respBytes, &apiErr); jsonErr == nil && apiErr.Error.Message != "" {
				return nil, fmt.Errorf("groq api error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
			}
			return nil, fmt.Errorf("groq api error with status %d", resp.StatusCode)
		}
	}

	var completionResp ChatCompletionResponse
	if err := json.Unmarshal(respBytes, &completionResp); err != nil {
		return nil, fmt.Errorf("failed to parse groq response json: %w", err)
	}

	if len(completionResp.Choices) == 0 {
		return nil, errors.New("groq returned completion with zero choices")
	}

	return &completionResp, nil
}
