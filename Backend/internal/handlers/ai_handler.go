package handlers

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"mindora/backend/internal/middleware"
	"mindora/backend/internal/services"
	"mindora/backend/pkg/groq"
	"mindora/backend/pkg/response"
)

// AIHandler exposes HTTP endpoints for Groq-powered academic productivity coaching.
type AIHandler struct {
	aiService services.AIService
}

// NewAIHandler instantiates a new AIHandler with the given service.
func NewAIHandler(aiService services.AIService) *AIHandler {
	return &AIHandler{
		aiService: aiService,
	}
}

// AIChatRequest defines the inbound payload for chat interactions.
type AIChatRequest struct {
	Prompt string `json:"prompt"`
}

// AIExplainRequest defines the inbound payload for explaining a specific prediction.
type AIExplainRequest struct {
	PredictionID *string `json:"prediction_id,omitempty"`
}

// Chat handles POST /api/v1/ai/chat and /api/ai/chat
func (h *AIHandler) Chat(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req AIChatRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Format JSON tidak valid"))
	}

	trimmedPrompt := strings.TrimSpace(req.Prompt)
	if trimmedPrompt == "" {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Pesan atau pertanyaan tidak boleh kosong"))
	}
	if len([]rune(trimmedPrompt)) > 1000 {
		return c.Status(fiber.StatusBadRequest).JSON(response.Error("Pesan melebihi batas maksimum 1000 karakter"))
	}

	resp, err := h.aiService.Chat(c.Context(), userID, trimmedPrompt)
	if err != nil {
		return h.handleAIError(c, err, "Gagal memproses percakapan asisten AI")
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Berhasil mendapatkan respons asisten AI", resp))
}

// ExplainPrediction handles POST /api/v1/ai/explain-prediction and /api/ai/explain-prediction
func (h *AIHandler) ExplainPrediction(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	var req AIExplainRequest
	var predIDPtr *uuid.UUID
	if len(c.Body()) > 0 {
		if err := c.BodyParser(&req); err == nil && req.PredictionID != nil && strings.TrimSpace(*req.PredictionID) != "" {
			parsedID, parseErr := uuid.Parse(strings.TrimSpace(*req.PredictionID))
			if parseErr != nil {
				return c.Status(fiber.StatusBadRequest).JSON(response.Error("ID prediksi tidak valid"))
			}
			predIDPtr = &parsedID
		}
	}

	resp, err := h.aiService.ExplainPrediction(c.Context(), userID, predIDPtr)
	if err != nil {
		if errors.Is(err, services.ErrAIPredictionNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(response.Error("Data prediksi produktivitas tidak ditemukan"))
		}
		return h.handleAIError(c, err, "Gagal memproses penjelasan prediksi")
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Berhasil menjelaskan estimasi produktivitas", resp))
}

// GenerateStudyPlan handles POST /api/v1/ai/study-plan and /api/ai/study-plan
func (h *AIHandler) GenerateStudyPlan(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(response.Error("Authentication required"))
	}

	resp, err := h.aiService.GenerateStudyPlan(c.Context(), userID)
	if err != nil {
		return h.handleAIError(c, err, "Gagal menyusun rencana belajar")
	}

	return c.Status(fiber.StatusOK).JSON(response.Success("Berhasil menyusun rencana belajar", resp))
}

func (h *AIHandler) handleAIError(c *fiber.Ctx, err error, defaultMsg string) error {
	if errors.Is(err, groq.ErrRateLimitExceeded) {
		return c.Status(fiber.StatusTooManyRequests).JSON(response.Error("Batas permintaan AI terlampaui. Silakan tunggu sesaat lagi."))
	}
	if errors.Is(err, groq.ErrUnauthorized) || errors.Is(err, services.ErrAIClientNotConfigured) || errors.Is(err, groq.ErrMissingAPIKey) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Error("Layanan asisten cerdas Mindora sedang tidak tersedia atau belum terkonfigurasi"))
	}
	if errors.Is(err, groq.ErrServiceUnavailable) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(response.Error("Layanan Groq AI sedang sibuk atau tidak dapat dijangkau. Silakan coba kembali beberapa saat lagi."))
	}
	if errors.Is(err, context.DeadlineExceeded) || strings.Contains(err.Error(), "timed out") {
		return c.Status(fiber.StatusGatewayTimeout).JSON(response.Error("Permintaan asisten AI membutuhkan waktu terlalu lama. Silakan coba kembali."))
	}
	log.Printf("[ERROR] [AIHandler] %v", err)
	return c.Status(fiber.StatusInternalServerError).JSON(response.Error(defaultMsg))
}
