package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/pkg/groq"
)

var (
	ErrAIPromptRequired      = errors.New("pesan atau pertanyaan tidak boleh kosong")
	ErrAIPromptTooLong       = errors.New("pesan melebihi batas maksimum 1000 karakter")
	ErrAIClientNotConfigured = errors.New("layanan asisten cerdas Mindora belum dikonfigurasi")
	ErrAIPredictionNotFound  = errors.New("data prediksi produktivitas tidak ditemukan")
)

// AIResponse models the reply sent back to the handler and client.
type AIResponse struct {
	Reply      string `json:"reply"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used,omitempty"`
}

// AIService defines operations for intelligent academic coaching via Groq LLM.
type AIService interface {
	Chat(ctx context.Context, userID uuid.UUID, prompt string) (*AIResponse, error)
	ExplainPrediction(ctx context.Context, userID uuid.UUID, predictionID *uuid.UUID) (*AIResponse, error)
	GenerateStudyPlan(ctx context.Context, userID uuid.UUID) (*AIResponse, error)
}

type aiService struct {
	client         groq.Client
	taskRepo       repositories.TaskRepository
	studyRepo      repositories.StudyRepository
	focusRepo      repositories.FocusRepository
	lifestyleRepo  repositories.LifestyleRepository
	predictionRepo repositories.PredictionRepository
	recRepo        repositories.RecommendationRepository
}

// NewAIService creates an AIService instance with required dependencies.
func NewAIService(
	client groq.Client,
	taskRepo repositories.TaskRepository,
	studyRepo repositories.StudyRepository,
	focusRepo repositories.FocusRepository,
	lifestyleRepo repositories.LifestyleRepository,
	predictionRepo repositories.PredictionRepository,
	recRepo repositories.RecommendationRepository,
) AIService {
	return &aiService{
		client:         client,
		taskRepo:       taskRepo,
		studyRepo:      studyRepo,
		focusRepo:      focusRepo,
		lifestyleRepo:  lifestyleRepo,
		predictionRepo: predictionRepo,
		recRepo:        recRepo,
	}
}

const systemPromptCoach = `Kamu adalah teman belajar santai dan suportif di Mindora. Bicaralah seperti teman dekat yang memberi saran praktis dan hangat lewat pesan chat, menggunakan Bahasa Indonesia kasual yang natural (gunakan sapaan akrab "kamu", jangan "Anda" atau "mahasiswa").

Prinsip Utama Gaya Percakapan:
1. Mengobrol Mengalir, Bukan Artikel Instruksi:
   - Sampaikan jawaban dalam 2–3 paragraf pendek yang mengalir alami, seperti seseorang yang sedang membalas chat teman.
   - HINDARI bullet points (- atau *) dan daftar nomor untuk obrolan biasa. Jangan ubah saran menjadi daftar poin terpisah.
   - HINDARI pola robotik seperti "- **Judul**: Penjelasan". Ini sangat kaku dan terasa seperti artikel AI generik.
2. Langsung ke Inti Tanpa Basa-Basi:
   - HINDARI kalimat pembuka klise seperti "Berikut beberapa tips...", "Biar nggak gampang ..., coba pakai trik simpel berikut:", "Tentu, ini solusinya:". Mulailah langsung dari inti tanggapanmu.
   - HINDARI penutup klise seperti "Selamat belajar!", "Semoga bermanfaat!", "Semoga membantu!", atau "Semoga sukses!". Cukup akhiri dengan kalimat penutup yang tenang dan wajar.
   - Jangan mengulang-ulang kata "coba", "berikut", "trik", atau "solusi" di setiap kalimat.
3. Proporsional & Bersahabat:
   - Sesuaikan panjang respon: pertanyaan santai/pendek dijawab singkat (1-2 paragraf).
   - Nada bicara santai, empatik, praktis, dan tidak menggurui.
4. Integritas Skor & Data (Sangat Penting):
   - Kamu adalah asisten percakapan, BUKAN penentu atau pengubah skor. Skor produktivitas di Mindora dihitung murni secara objektif oleh model Machine Learning berdasarkan data riil sesi belajar dan kebiasaan harian.
   - JANGAN PERNAH mengubah, memalsukan, atau menjanjikan perubahan skor produktivitas jika pengguna memintanya. Jelaskan secara ramah bahwa skor tidak bisa diubah lewat chat dan akan naik dengan konsistensi belajar nyata.
   - JANGAN PERNAH mengarang data aktivitas atau skor jika tidak ada di konteks. Jika data belum tercatat, akui apa adanya secara wajar.
5. Privasi & Keamanan Pengguna:
   - Kamu hanya mendampingi pengguna yang sedang mengobrol denganmu secara privat. Kamu sama sekali tidak memiliki akses ke data, akun, atau skor pengguna lain. Jika pengguna menanyakan data orang lain (misalnya skor atau nama teman), tolak dengan ramah bahwa kamu menjaga privasi setiap orang dan hanya fokus mendampingi proses belajarnya.
6. Rencana Belajar yang Manusiawi:
   - Jika ditanya jadwal belajar, jelaskan alurnya secara santai dan realistis dalam paragraf biasa (mulai dari materi apa, berapa lama fokusnya, kapan jeda santai).
   - JANGAN PERNAH membuat tabel Markdown atau jadwal kaku jam-per-jam dari pagi sampai malam.
7. Kontekstual & Alami:
   - Jika ada konteks pengguna (seperti tidur, tugas, atau waktu belajar), manfaatkan secara halus dalam saranmu tanpa menyebutkan istilah teknis, rumus ML, atau mengomentari data yang belum lengkap.
8. Non-Klinis: Fokus pada kebiasaan belajar dan manajemen waktu, bukan diagnosis medis.`

// Chat handles interactive conversation with context awareness.
func (s *aiService) Chat(ctx context.Context, userID uuid.UUID, prompt string) (*AIResponse, error) {
	if s.client == nil {
		return nil, ErrAIClientNotConfigured
	}

	trimmedPrompt := strings.TrimSpace(prompt)
	if trimmedPrompt == "" {
		return nil, ErrAIPromptRequired
	}
	if len([]rune(trimmedPrompt)) > 1000 {
		return nil, ErrAIPromptTooLong
	}

	contextSummary := s.gatherStudentContext(ctx, userID)

	var userMessage string
	if contextSummary != "" {
		userMessage = fmt.Sprintf("%s\n\nPesan pengguna: %s", contextSummary, trimmedPrompt)
	} else {
		userMessage = trimmedPrompt
	}

	temp := 0.7
	maxTokens := 1024
	req := groq.ChatCompletionRequest{
		Model: s.client.GetModel(),
		Messages: []groq.ChatMessage{
			{Role: "system", Content: systemPromptCoach},
			{Role: "user", Content: userMessage},
		},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return &AIResponse{
			Reply:      "Maaf, asisten Mindora belum dapat menghasilkan respons saat ini. Silakan coba kembali sesaat lagi.",
			Model:      resp.Model,
			TokensUsed: resp.Usage.TotalTokens,
		}, nil
	}

	return &AIResponse{
		Reply:      strings.TrimSpace(resp.Choices[0].Message.Content),
		Model:      resp.Model,
		TokensUsed: resp.Usage.TotalTokens,
	}, nil
}

// ExplainPrediction breaks down the Ridge Regression prediction into actionable insights.
func (s *aiService) ExplainPrediction(ctx context.Context, userID uuid.UUID, predictionID *uuid.UUID) (*AIResponse, error) {
	if s.client == nil {
		return nil, ErrAIClientNotConfigured
	}

	var pred *models.ProductivityPrediction
	var err error

	if predictionID != nil && *predictionID != uuid.Nil {
		pred, err = s.predictionRepo.FindByID(ctx, *predictionID, userID)
	} else {
		pred, err = s.predictionRepo.FindLatest(ctx, userID)
	}

	if err != nil || pred == nil {
		return nil, ErrAIPredictionNotFound
	}

	sleepText := "belum tercatat"
	if pred.SleepHours != nil {
		sleepText = fmt.Sprintf("%.1f jam", *pred.SleepHours)
	}

	phoneText := "belum tercatat"
	if pred.PhoneUsageHours != nil {
		phoneText = fmt.Sprintf("%.1f jam", *pred.PhoneUsageHours)
	}

	stressText := "belum tercatat"
	if pred.StressLevel != nil {
		stressText = fmt.Sprintf("%d/5", *pred.StressLevel)
	}

	stateLabel := pred.StateLabel
	if stateLabel == "" {
		if pred.IsBaseline {
			stateLabel = "Estimasi Awal"
		} else {
			stateLabel = "Estimasi Personal"
		}
	}

	var baselineNotice string
	if pred.IsBaseline {
		baselineNotice = "\n- CATATAN KHUSUS: Skor ini adalah estimasi awal berdasarkan baseline rata-rata karena data sesi belajar dan gaya hidup belum mencukupi. Jelaskan secara ramah bahwa ini adalah acuan awal dan skor personal yang lebih akurat akan terbentuk setelah kamu mencatat sesi belajar dan kebiasaan harian secara konsisten."
	}

	explanationPrompt := fmt.Sprintf(`Jelaskan hasil estimasi produktivitas ini kepada temanmu secara santai, hangat, dan mudah dipahami:

Informasi:
- Skor Produktivitas: %.1f / 100 (%s)
- Waktu belajar: %.1f jam
- Jam tidur: %s
- Layar HP: %s
- Tingkat stres: %s%s

Aturan penyampaian:
- Sampaikan dalam 2-3 paragraf naratif pendek yang mengalir (JANGAN gunakan bullet points, JANGAN gunakan tabel, dan JANGAN gunakan bab bernomor).
- Jelaskan arti skor %.1f secara wajar dan positif.
- Ceritakan secara natural kaitan antara istirahat/tidur, layar HP, waktu belajar, dan stres terhadap energimu.
- Jika ini adalah estimasi awal (baseline), jelaskan dengan tenang bahwa skor ini acuan permulaan dan akan menyesuaikan ritme aslimu seiring waktu.
- Berikan saran santai untuk menjaga ritme hari ini tanpa penutup klise.`,
		pred.PredictedScore,
		stateLabel,
		pred.StudyHours,
		sleepText,
		phoneText,
		stressText,
		baselineNotice,
		pred.PredictedScore,
	)

	temp := 0.6
	maxTokens := 1024
	req := groq.ChatCompletionRequest{
		Model: s.client.GetModel(),
		Messages: []groq.ChatMessage{
			{Role: "system", Content: systemPromptCoach},
			{Role: "user", Content: explanationPrompt},
		},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return &AIResponse{
			Reply:      "Maaf, saat ini sistem belum dapat merinci penjelasan prediksi. Silakan coba kembali.",
			Model:      resp.Model,
			TokensUsed: resp.Usage.TotalTokens,
		}, nil
	}

	return &AIResponse{
		Reply:      strings.TrimSpace(resp.Choices[0].Message.Content),
		Model:      resp.Model,
		TokensUsed: resp.Usage.TotalTokens,
	}, nil
}

// GenerateStudyPlan crafts an optimized daily schedule based on tasks and stamina.
func (s *aiService) GenerateStudyPlan(ctx context.Context, userID uuid.UUID) (*AIResponse, error) {
	if s.client == nil {
		return nil, ErrAIClientNotConfigured
	}

	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	// Fetch pending and in_progress tasks
	tasks, _, _ := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Limit:     10,
		SortBy:    "priority",
		SortOrder: "desc",
	})

	studyStats, _ := s.studyRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	focusStats, _ := s.focusRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay)
	lifestyle, _ := s.lifestyleRepo.FindByDate(ctx, userID, now)

	var taskLines []string
	for _, t := range tasks {
		if t.Status == models.TaskStatusCompleted || t.Status == models.TaskStatusCancelled {
			continue
		}
		deadlineStr := "tanpa tenggat"
		if t.Deadline != nil {
			deadlineStr = t.Deadline.Format("02 Jan 15:04")
		}
		taskLines = append(taskLines, fmt.Sprintf("- [%s] %s (Kategori: %s, Prioritas: %s, Tenggat: %s)",
			strings.ToUpper(t.Status), t.Title, t.Category, t.Priority, deadlineStr))
	}

	tasksDescription := "Tidak ada tugas tertunda saat ini. Mahasiswa dapat fokus pada pengulangan materi atau eksplorasi mandiri."
	if len(taskLines) > 0 {
		tasksDescription = strings.Join(taskLines, "\n")
	}

	studyMinutes := int64(0)
	if studyStats != nil {
		studyMinutes = studyStats.TotalCompletedSeconds / 60
	}

	focusMinutes := int64(0)
	interruptions := int64(0)
	if focusStats != nil {
		focusMinutes = focusStats.CompletedFocusMinutes
		interruptions = focusStats.TotalInterruptions
	}

	lifestyleText := "Belum tercatat hari ini"
	if lifestyle != nil {
		var parts []string
		if lifestyle.SleepHours != nil {
			parts = append(parts, fmt.Sprintf("Tidur: %.1f jam", *lifestyle.SleepHours))
		}
		if lifestyle.PhoneUsageHours != nil {
			parts = append(parts, fmt.Sprintf("Layar HP: %.1f jam", *lifestyle.PhoneUsageHours))
		}
		if lifestyle.StressLevel != nil {
			parts = append(parts, fmt.Sprintf("Tingkat Stres: %d/5", *lifestyle.StressLevel))
		}
		if len(parts) > 0 {
			lifestyleText = strings.Join(parts, ", ")
		}
	}

	prompt := fmt.Sprintf(`Bantu susun alur belajar hari ini yang santai, realistis, dan fleksibel untuk temanmu.

Kondisi saat ini:
- Sudah belajar: %d menit
- Fokus & distraksi: %d menit fokus, %d distraksi
- Kondisi fisik: %s

Tugas yang ada:
%s

Aturan penyusunan:
- Tuliskan dalam 2-3 paragraf naratif mengalir yang enak dibaca (JANGAN gunakan tabel Markdown, JANGAN gunakan bullet point atau daftar nomor, dan JANGAN buat judul-judul tebal).
- Jelaskan tugas apa yang sebaiknya dicicil lebih dulu, saran durasi fokus yang nyaman (misal 25-40 menit), dan kapan sebaiknya istirahat santai sejenak.
- Tekankan bahwa alur ini santai dan bisa disesuaikan dengan kondisi energinya hari ini.`,
		studyMinutes,
		focusMinutes,
		interruptions,
		lifestyleText,
		tasksDescription,
	)

	temp := 0.7
	maxTokens := 1024
	req := groq.ChatCompletionRequest{
		Model: s.client.GetModel(),
		Messages: []groq.ChatMessage{
			{Role: "system", Content: systemPromptCoach},
			{Role: "user", Content: prompt},
		},
		Temperature: &temp,
		MaxTokens:   &maxTokens,
	}

	resp, err := s.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return &AIResponse{
			Reply:      "Maaf, saat ini sistem belum dapat menyusun rencana belajar. Silakan coba kembali.",
			Model:      resp.Model,
			TokensUsed: resp.Usage.TotalTokens,
		}, nil
	}

	return &AIResponse{
		Reply:      strings.TrimSpace(resp.Choices[0].Message.Content),
		Model:      resp.Model,
		TokensUsed: resp.Usage.TotalTokens,
	}, nil
}

// gatherStudentContext assembles non-sensitive student metrics into a context block.
func (s *aiService) gatherStudentContext(ctx context.Context, userID uuid.UUID) string {
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	endOfDay := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, time.UTC)

	var items []string

	// 1. Latest Prediction
	if pred, err := s.predictionRepo.FindLatest(ctx, userID); err == nil && pred != nil {
		items = append(items, fmt.Sprintf("- Skor estimasi produktivitas: %.1f/100 (%s)",
			pred.PredictedScore, pred.StateLabel))
	}

	// 2. Study Session Stats
	if studyStats, err := s.studyRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay); err == nil && studyStats != nil && studyStats.TotalCompletedSeconds > 0 {
		studyMin := studyStats.TotalCompletedSeconds / 60
		items = append(items, fmt.Sprintf("- Belajar hari ini: %d menit (%d sesi)",
			studyMin, studyStats.CompletedSessionsCount))
	}

	// 3. Focus Interval Stats
	if focusStats, err := s.focusRepo.GetDailyStatistics(ctx, userID, startOfDay, endOfDay); err == nil && focusStats != nil && (focusStats.CompletedFocusMinutes > 0 || focusStats.TotalInterruptions > 0) {
		items = append(items, fmt.Sprintf("- Fokus Pomodoro hari ini: %d menit, %d distraksi",
			focusStats.CompletedFocusMinutes, focusStats.TotalInterruptions))
	}

	// 4. Lifestyle Record
	if lifestyle, err := s.lifestyleRepo.FindByDate(ctx, userID, now); err == nil && lifestyle != nil {
		var parts []string
		if lifestyle.SleepHours != nil {
			parts = append(parts, fmt.Sprintf("tidur %.1f jam", *lifestyle.SleepHours))
		}
		if lifestyle.PhoneUsageHours != nil {
			parts = append(parts, fmt.Sprintf("layar HP %.1f jam", *lifestyle.PhoneUsageHours))
		}
		if lifestyle.StressLevel != nil {
			parts = append(parts, fmt.Sprintf("stres %d/5", *lifestyle.StressLevel))
		}
		if len(parts) > 0 {
			items = append(items, fmt.Sprintf("- Kondisi fisik/kebiasaan: %s", strings.Join(parts, ", ")))
		}
	}

	// 5. Active Pending Tasks
	if tasks, _, err := s.taskRepo.FindAll(ctx, userID, repositories.TaskFilter{
		Status: models.TaskStatusPending,
		Limit:  3,
	}); err == nil && len(tasks) > 0 {
		var titles []string
		for _, t := range tasks {
			titles = append(titles, fmt.Sprintf("%s (%s)", t.Title, t.Priority))
		}
		items = append(items, fmt.Sprintf("- Tugas tertunda: %s", strings.Join(titles, ", ")))
	}

	if len(items) == 0 {
		return ""
	}

	return fmt.Sprintf("[Konteks aktivitas pengguna saat ini (gunakan secara alami hanya jika relevan dengan pertanyaannya, jangan sebutkan jika tidak berkaitan)]:\n%s", strings.Join(items, "\n"))
}
