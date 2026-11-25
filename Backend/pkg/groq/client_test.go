package groq_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mindora/backend/pkg/groq"
)

func TestGroqClient_MissingAPIKey(t *testing.T) {
	client := groq.NewClient("", "openai/gpt-oss-20b", "https://api.groq.com/openai/v1", 5*time.Second)
	_, err := client.CreateChatCompletion(context.Background(), groq.ChatCompletionRequest{
		Messages: []groq.ChatMessage{{Role: "user", Content: "Hello"}},
	})
	if !errors.Is(err, groq.ErrMissingAPIKey) {
		t.Fatalf("expected ErrMissingAPIKey, got: %v", err)
	}
}

func TestGroqClient_EmptyMessages(t *testing.T) {
	client := groq.NewClient("test-key", "openai/gpt-oss-20b", "https://api.groq.com/openai/v1", 5*time.Second)
	_, err := client.CreateChatCompletion(context.Background(), groq.ChatCompletionRequest{
		Messages: []groq.ChatMessage{},
	})
	if !errors.Is(err, groq.ErrEmptyMessages) {
		t.Fatalf("expected ErrEmptyMessages, got: %v", err)
	}
}

func TestGroqClient_SuccessfulCompletion(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer valid-test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got: %s", r.Header.Get("Content-Type"))
		}

		var req groq.ChatCompletionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}

		if req.Model != "openai/gpt-oss-20b" {
			t.Errorf("expected default model openai/gpt-oss-20b, got: %s", req.Model)
		}

		resp := groq.ChatCompletionResponse{
			ID:    "chatcmpl-test-123",
			Model: req.Model,
			Choices: []groq.ChatChoice{
				{
					Index: 0,
					Message: groq.ChatMessage{
						Role:    "assistant",
						Content: "Tentu! Untuk meningkatkan fokus belajar, gunakan teknik Pomodoro 25 menit.",
					},
					FinishReason: "stop",
				},
			},
			Usage: groq.UsageStats{
				PromptTokens:     15,
				CompletionTokens: 20,
				TotalTokens:      35,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	client := groq.NewClient("valid-test-key", "openai/gpt-oss-20b", mockServer.URL, 5*time.Second)
	res, err := client.CreateChatCompletion(context.Background(), groq.ChatCompletionRequest{
		Messages: []groq.ChatMessage{
			{Role: "system", Content: "You are Mindora AI."},
			{Role: "user", Content: "Bagaimana cara fokus belajar?"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected completion error: %v", err)
	}

	if res.Choices[0].Message.Content != "Tentu! Untuk meningkatkan fokus belajar, gunakan teknik Pomodoro 25 menit." {
		t.Errorf("unexpected content: %s", res.Choices[0].Message.Content)
	}
	if res.Usage.TotalTokens != 35 {
		t.Errorf("expected 35 tokens, got: %d", res.Usage.TotalTokens)
	}
}

func TestGroqClient_ErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		expectedErr  error
	}{
		{
			name:         "Unauthorized 401",
			statusCode:   http.StatusUnauthorized,
			responseBody: `{"error":{"message":"Invalid API Key","type":"invalid_request_error"}}`,
			expectedErr:  groq.ErrUnauthorized,
		},
		{
			name:         "Rate Limit 429",
			statusCode:   http.StatusTooManyRequests,
			responseBody: `{"error":{"message":"Rate limit exceeded","type":"rate_limit_error"}}`,
			expectedErr:  groq.ErrRateLimitExceeded,
		},
		{
			name:         "Service Unavailable 503",
			statusCode:   http.StatusServiceUnavailable,
			responseBody: `Service Unavailable`,
			expectedErr:  groq.ErrServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer mockServer.Close()

			client := groq.NewClient("dummy-key", "openai/gpt-oss-20b", mockServer.URL, 5*time.Second)
			_, err := client.CreateChatCompletion(context.Background(), groq.ChatCompletionRequest{
				Messages: []groq.ChatMessage{{Role: "user", Content: "Ping"}},
			})

			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}

func TestGroqClient_ContextCancellation(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := groq.NewClient("dummy-key", "openai/gpt-oss-20b", mockServer.URL, 5*time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.CreateChatCompletion(ctx, groq.ChatCompletionRequest{
		Messages: []groq.ChatMessage{{Role: "user", Content: "Hi"}},
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestGroqClient_Timeout(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	client := groq.NewClient("dummy-key", "openai/gpt-oss-20b", mockServer.URL, 50*time.Millisecond)
	_, err := client.CreateChatCompletion(context.Background(), groq.ChatCompletionRequest{
		Messages: []groq.ChatMessage{{Role: "user", Content: "Hi"}},
	})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected context.DeadlineExceeded, got: %v", err)
	}
}
