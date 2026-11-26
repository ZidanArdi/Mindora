package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
	"mindora/backend/internal/services"
)

type mockDedicatedStudyRepo struct {
	sessions []*models.StudySession
}

func (m *mockDedicatedStudyRepo) CreateSession(ctx context.Context, s *models.StudySession) error {
	m.sessions = append(m.sessions, s)
	return nil
}

func (m *mockDedicatedStudyRepo) FindSessionByID(ctx context.Context, id uuid.UUID, uID uuid.UUID) (*models.StudySession, error) {
	for _, s := range m.sessions {
		if s.ID == id && s.UserID == uID {
			return s, nil
		}
	}
	return nil, repositories.ErrStudySessionNotFound
}

func (m *mockDedicatedStudyRepo) FindActiveOrPausedSession(ctx context.Context, uID uuid.UUID) (*models.StudySession, error) {
	for _, s := range m.sessions {
		if s.UserID == uID && (s.Status == models.StudyStatusActive || s.Status == models.StudyStatusPaused) {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockDedicatedStudyRepo) FindSessions(ctx context.Context, uID uuid.UUID, start, end *time.Time, limit, offset int) ([]models.StudySession, int64, error) {
	var list []models.StudySession
	for _, s := range m.sessions {
		if s.UserID == uID {
			if start != nil && s.StartedAt.Before(*start) {
				continue
			}
			if end != nil && s.StartedAt.After(*end) {
				continue
			}
			list = append(list, *s)
		}
	}
	return list, int64(len(list)), nil
}

func (m *mockDedicatedStudyRepo) FindCompletedSessions(ctx context.Context, uID uuid.UUID, start, end time.Time) ([]models.StudySession, error) {
	var list []models.StudySession
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			list = append(list, *s)
		}
	}
	return list, nil
}

func (m *mockDedicatedStudyRepo) UpdateSession(ctx context.Context, s *models.StudySession) error {
	for i, existing := range m.sessions {
		if existing.ID == s.ID && existing.UserID == s.UserID {
			m.sessions[i] = s
			return nil
		}
	}
	return repositories.ErrStudySessionNotFound
}

func (m *mockDedicatedStudyRepo) GetCompletedDuration(ctx context.Context, uID uuid.UUID, start, end time.Time) (int64, error) {
	var total int64
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			total += int64(s.DurationSeconds)
		}
	}
	return total, nil
}

func (m *mockDedicatedStudyRepo) GetDailyStatistics(ctx context.Context, uID uuid.UUID, start, end time.Time) (*repositories.StudyDailyStats, error) {
	var total int64
	var count int64
	for _, s := range m.sessions {
		if s.UserID == uID && s.Status == models.StudyStatusCompleted && !s.StartedAt.Before(start) && !s.StartedAt.After(end) {
			total += int64(s.DurationSeconds)
			count++
		}
	}
	return &repositories.StudyDailyStats{
		TotalCompletedSeconds:  total,
		CompletedSessionsCount: count,
	}, nil
}

type mockTaskRepoForStudy struct {
	tasks []*models.Task
}

func (m *mockTaskRepoForStudy) Create(ctx context.Context, t *models.Task) error {
	m.tasks = append(m.tasks, t)
	return nil
}
func (m *mockTaskRepoForStudy) FindByID(ctx context.Context, id, uID uuid.UUID) (*models.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id && t.UserID == uID {
			return t, nil
		}
	}
	return nil, repositories.ErrTaskNotFound
}
func (m *mockTaskRepoForStudy) FindAll(ctx context.Context, uID uuid.UUID, f repositories.TaskFilter) ([]models.Task, int64, error) {
	return nil, 0, nil
}
func (m *mockTaskRepoForStudy) Update(ctx context.Context, t *models.Task) error { return nil }
func (m *mockTaskRepoForStudy) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string, completedAt *time.Time) (*models.Task, error) {
	return nil, nil
}
func (m *mockTaskRepoForStudy) SoftDelete(ctx context.Context, id, uID uuid.UUID) error { return nil }
func (m *mockTaskRepoForStudy) CountByStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	return nil, nil
}
func (m *mockTaskRepoForStudy) CountCompletedByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	return 0, nil
}
func (m *mockTaskRepoForStudy) FindCompletedTasksByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) ([]models.Task, error) {
	return nil, nil
}

func TestStartStudySession_ValidationAndLinkedTask(t *testing.T) {
	studyRepo := &mockDedicatedStudyRepo{}
	taskRepo := &mockTaskRepoForStudy{}
	svc := services.NewStudyService(studyRepo, taskRepo)
	ctx := context.Background()
	userID := uuid.New()

	// 1. Empty subject should fail
	_, err := svc.StartSession(ctx, userID, services.StartStudyInput{Subject: "   "})
	if err != services.ErrStudySubjectRequired {
		t.Fatalf("expected ErrStudySubjectRequired, got: %v", err)
	}

	// 2. Start independent session (Belajar mandiri) -> task_id is nil
	sess1, err := svc.StartSession(ctx, userID, services.StartStudyInput{Subject: "Algoritma Terdistribusi"})
	if err != nil {
		t.Fatalf("unexpected error starting session: %v", err)
	}
	if sess1.TaskID != nil {
		t.Errorf("expected nil task_id for independent study, got: %v", sess1.TaskID)
	}
	if sess1.DurationSeconds != 0 {
		t.Errorf("expected 0 initial duration_seconds, got: %d", sess1.DurationSeconds)
	}
	if sess1.Status != models.StudyStatusActive {
		t.Errorf("expected status 'active', got: %s", sess1.Status)
	}

	// 3. Starting another session while one is active should fail
	_, err = svc.StartSession(ctx, userID, services.StartStudyInput{Subject: "Basis Data"})
	if err != services.ErrActiveStudySessionExists {
		t.Fatalf("expected ErrActiveStudySessionExists, got: %v", err)
	}

	// 4. Cancel active session
	_, err = svc.CancelSession(ctx, sess1.ID, userID)
	if err != nil {
		t.Fatalf("unexpected error cancelling session: %v", err)
	}

	// 5. Start with non-existent task should fail
	invalidTaskID := uuid.New()
	_, err = svc.StartSession(ctx, userID, services.StartStudyInput{
		Subject: "Jaringan Komputer",
		TaskID:  &invalidTaskID,
	})
	if err != services.ErrTaskOwnershipInvalid {
		t.Fatalf("expected ErrTaskOwnershipInvalid, got: %v", err)
	}

	// 6. Start with valid task ownership
	validTaskID := uuid.New()
	taskRepo.tasks = append(taskRepo.tasks, &models.Task{
		ID:     validTaskID,
		UserID: userID,
		Title:  "Tugas Praktikum 3",
	})
	sess2, err := svc.StartSession(ctx, userID, services.StartStudyInput{
		Subject: "Jaringan Komputer",
		TaskID:  &validTaskID,
	})
	if err != nil {
		t.Fatalf("unexpected error starting session with valid task: %v", err)
	}
	if sess2.TaskID == nil || *sess2.TaskID != validTaskID {
		t.Errorf("expected task_id %v, got %v", validTaskID, sess2.TaskID)
	}
}

func TestStudySession_ShortAndLongDurations(t *testing.T) {
	studyRepo := &mockDedicatedStudyRepo{}
	taskRepo := &mockTaskRepoForStudy{}
	svc := services.NewStudyService(studyRepo, taskRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Short session: 6 seconds
	sessShort := &models.StudySession{
		ID:              uuid.New(),
		UserID:          userID,
		Subject:         "Ngoding",
		DurationSeconds: 6,
		Status:          models.StudyStatusCompleted,
		StartedAt:       time.Date(2026, 11, 20, 18, 38, 20, 0, time.UTC),
	}
	studyRepo.sessions = append(studyRepo.sessions, sessShort)

	// Long session: 15349 seconds (4 hours 15 minutes 49 seconds)
	sessLong := &models.StudySession{
		ID:              uuid.New(),
		UserID:          userID,
		Subject:         "Manqulan",
		DurationSeconds: 15349,
		Status:          models.StudyStatusCompleted,
		StartedAt:       time.Date(2026, 11, 21, 12, 2, 8, 0, time.UTC),
	}
	studyRepo.sessions = append(studyRepo.sessions, sessLong)

	// Verify short session retrieved
	s1, _, err := svc.ListSessions(ctx, userID, nil, nil, 1, 10)
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	if len(s1) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(s1))
	}

	foundShort := false
	foundLong := false
	for _, s := range s1 {
		if s.DurationSeconds == 6 {
			foundShort = true
			if s.Subject != "Ngoding" {
				t.Errorf("unexpected subject for 6s session: %s", s.Subject)
			}
		}
		if s.DurationSeconds == 15349 {
			foundLong = true
			if s.Subject != "Manqulan" {
				t.Errorf("unexpected subject for 15349s session: %s", s.Subject)
			}
		}
	}
	if !foundShort || !foundLong {
		t.Errorf("expected to find both 6s and 15349s sessions, foundShort=%v, foundLong=%v", foundShort, foundLong)
	}
}

func TestGetDailyStats_TimezoneAndDateFiltering(t *testing.T) {
	studyRepo := &mockDedicatedStudyRepo{}
	taskRepo := &mockTaskRepoForStudy{}
	svc := services.NewStudyService(studyRepo, taskRepo)
	ctx := context.Background()
	userID := uuid.New()

	// Jakarta location (UTC+7)
	locJakarta, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		locJakarta = time.FixedZone("WIB", 7*3600)
	}

	// Session 1: November 21, 2026 23:30 WIB (which is 16:30 UTC on Nov 21)
	s1Time := time.Date(2026, 11, 21, 23, 30, 0, 0, locJakarta)
	studyRepo.sessions = append(studyRepo.sessions, &models.StudySession{
		ID:              uuid.New(),
		UserID:          userID,
		Subject:         "Sesi Malam 21 Nov",
		DurationSeconds: 1800, // 30 mins
		Status:          models.StudyStatusCompleted,
		StartedAt:       s1Time.UTC(),
	})

	// Session 2: November 22, 2026 01:00 WIB (which is 18:00 UTC on Nov 21)
	// In WIB this is Nov 22; in UTC this is Nov 21!
	s2Time := time.Date(2026, 11, 22, 1, 0, 0, 0, locJakarta)
	studyRepo.sessions = append(studyRepo.sessions, &models.StudySession{
		ID:              uuid.New(),
		UserID:          userID,
		Subject:         "Sesi Dini Hari 22 Nov",
		DurationSeconds: 3600, // 1 hour
		Status:          models.StudyStatusCompleted,
		StartedAt:       s2Time.UTC(),
	})

	// Query for Nov 21 in Asia/Jakarta: should ONLY include Session 1 (1800s, 1 session)
	statsNov21, err := svc.GetDailyStats(ctx, userID, time.Date(2026, 11, 21, 0, 0, 0, 0, locJakarta))
	if err != nil {
		t.Fatalf("failed to get Nov 21 stats: %v", err)
	}
	if statsNov21.CompletedSessionsCount != 1 {
		t.Errorf("expected 1 session on Nov 21 WIB, got %d", statsNov21.CompletedSessionsCount)
	}
	if statsNov21.TotalCompletedSeconds != 1800 {
		t.Errorf("expected 1800s on Nov 21 WIB, got %d", statsNov21.TotalCompletedSeconds)
	}

	// Query for Nov 22 in Asia/Jakarta: should ONLY include Session 2 (3600s, 1 session)
	statsNov22, err := svc.GetDailyStats(ctx, userID, time.Date(2026, 11, 22, 0, 0, 0, 0, locJakarta))
	if err != nil {
		t.Fatalf("failed to get Nov 22 stats: %v", err)
	}
	if statsNov22.CompletedSessionsCount != 1 {
		t.Errorf("expected 1 session on Nov 22 WIB, got %d", statsNov22.CompletedSessionsCount)
	}
	if statsNov22.TotalCompletedSeconds != 3600 {
		t.Errorf("expected 3600s on Nov 22 WIB, got %d", statsNov22.TotalCompletedSeconds)
	}

	// Query for Nov 23 (empty day): should return 0 sessions, 0 seconds
	statsNov23, err := svc.GetDailyStats(ctx, userID, time.Date(2026, 11, 23, 0, 0, 0, 0, locJakarta))
	if err != nil {
		t.Fatalf("failed to get Nov 23 stats: %v", err)
	}
	if statsNov23.CompletedSessionsCount != 0 {
		t.Errorf("expected 0 sessions on Nov 23 WIB, got %d", statsNov23.CompletedSessionsCount)
	}
	if statsNov23.TotalCompletedSeconds != 0 {
		t.Errorf("expected 0s on Nov 23 WIB, got %d", statsNov23.TotalCompletedSeconds)
	}
}
