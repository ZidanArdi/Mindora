package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"mindora/backend/internal/models"
	"mindora/backend/internal/repositories"
)

var (
	ErrTaskTitleRequired  = errors.New("task title is required")
	ErrInvalidTaskStatus  = errors.New("invalid task status")
	ErrInvalidTaskPriority = errors.New("invalid task priority")
)

type CreateTaskInput struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Category    string     `json:"category"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

type UpdateTaskInput struct {
	Title       string     `json:"title"`
	Description *string    `json:"description,omitempty"`
	Category    string     `json:"category"`
	Priority    string     `json:"priority"`
	Deadline    *time.Time `json:"deadline,omitempty"`
}

type TaskService interface {
	CreateTask(ctx context.Context, userID uuid.UUID, input CreateTaskInput) (*models.Task, error)
	GetTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	ListTasks(ctx context.Context, userID uuid.UUID, filter repositories.TaskFilter) ([]models.Task, int64, error)
	UpdateTask(ctx context.Context, id uuid.UUID, userID uuid.UUID, input UpdateTaskInput) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string) (*models.Task, error)
	DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	GetTaskCounts(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
}

type taskService struct {
	repo repositories.TaskRepository
}

func NewTaskService(repo repositories.TaskRepository) TaskService {
	return &taskService{repo: repo}
}

func (s *taskService) CreateTask(ctx context.Context, userID uuid.UUID, input CreateTaskInput) (*models.Task, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	priority := strings.ToLower(strings.TrimSpace(input.Priority))
	if priority == "" {
		priority = models.TaskPriorityMedium
	} else if priority != models.TaskPriorityLow && priority != models.TaskPriorityMedium && priority != models.TaskPriorityHigh {
		return nil, ErrInvalidTaskPriority
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		category = "Study"
	}

	task := &models.Task{
		ID:          uuid.New(),
		UserID:      userID,
		Title:       title,
		Description: input.Description,
		Category:    category,
		Priority:    priority,
		Status:      models.TaskStatusPending,
		Deadline:    input.Deadline,
	}

	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	return s.repo.FindByID(ctx, id, userID)
}

func (s *taskService) ListTasks(ctx context.Context, userID uuid.UUID, filter repositories.TaskFilter) ([]models.Task, int64, error) {
	return s.repo.FindAll(ctx, userID, filter)
}

func (s *taskService) UpdateTask(ctx context.Context, id uuid.UUID, userID uuid.UUID, input UpdateTaskInput) (*models.Task, error) {
	task, err := s.repo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, ErrTaskTitleRequired
	}

	priority := strings.ToLower(strings.TrimSpace(input.Priority))
	if priority != "" && priority != models.TaskPriorityLow && priority != models.TaskPriorityMedium && priority != models.TaskPriorityHigh {
		return nil, ErrInvalidTaskPriority
	}

	task.Title = title
	task.Description = input.Description
	if input.Category != "" {
		task.Category = input.Category
	}
	if priority != "" {
		task.Priority = priority
	}
	task.Deadline = input.Deadline

	if err := s.repo.Update(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *taskService) UpdateTaskStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string) (*models.Task, error) {
	status = strings.ToLower(strings.TrimSpace(status))
	if status != models.TaskStatusPending &&
		status != models.TaskStatusInProgress &&
		status != models.TaskStatusCompleted &&
		status != models.TaskStatusCancelled {
		return nil, ErrInvalidTaskStatus
	}

	var completedAt *time.Time
	if status == models.TaskStatusCompleted {
		now := time.Now().UTC()
		completedAt = &now
	}

	return s.repo.UpdateStatus(ctx, id, userID, status, completedAt)
}

func (s *taskService) DeleteTask(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	return s.repo.SoftDelete(ctx, id, userID)
}

func (s *taskService) GetTaskCounts(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	return s.repo.CountByStatus(ctx, userID)
}
