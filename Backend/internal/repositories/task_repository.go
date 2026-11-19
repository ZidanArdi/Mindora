package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"mindora/backend/internal/models"
)

var (
	ErrTaskNotFound = errors.New("task not found or unauthorized")
)

// TaskFilter defines querying and sorting options for tasks.
type TaskFilter struct {
	Status   string
	Priority string
	Category string
	Search   string
	SortBy   string // "deadline", "created_at", "priority"
	SortOrder string // "asc", "desc"
	Limit    int
	Offset   int
}

type TaskRepository interface {
	Create(ctx context.Context, task *models.Task) error
	FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error)
	FindAll(ctx context.Context, userID uuid.UUID, filter TaskFilter) ([]models.Task, int64, error)
	Update(ctx context.Context, task *models.Task) error
	UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string, completedAt *time.Time) (*models.Task, error)
	SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	CountByStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error)
	CountCompletedByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error)
}

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *models.Task) error {
	if task.ID == uuid.Nil {
		task.ID = uuid.New()
	}
	now := time.Now().UTC()
	task.CreatedAt = now
	task.UpdatedAt = now
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) FindByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*models.Task, error) {
	var task models.Task
	err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		First(&task).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) FindAll(ctx context.Context, userID uuid.UUID, filter TaskFilter) ([]models.Task, int64, error) {
	query := r.db.WithContext(ctx).Model(&models.Task{}).Where("user_id = ?", userID)

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Priority != "" {
		query = query.Where("priority = ?", filter.Priority)
	}
	if filter.Category != "" {
		query = query.Where("LOWER(category) = ?", strings.ToLower(filter.Category))
	}
	if filter.Search != "" {
		searchPattern := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where("(LOWER(title) LIKE ? OR LOWER(description) LIKE ?)", searchPattern, searchPattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply Sorting
	orderClause := "created_at DESC"
	dir := "DESC"
	if strings.ToLower(filter.SortOrder) == "asc" {
		dir = "ASC"
	}

	switch filter.SortBy {
	case "deadline":
		orderClause = "deadline " + dir + " NULLS LAST, created_at DESC"
	case "priority":
		// High > Medium > Low
		if dir == "DESC" {
			orderClause = "CASE priority WHEN 'high' THEN 1 WHEN 'medium' THEN 2 WHEN 'low' THEN 3 ELSE 4 END ASC, created_at DESC"
		} else {
			orderClause = "CASE priority WHEN 'low' THEN 1 WHEN 'medium' THEN 2 WHEN 'high' THEN 3 ELSE 4 END ASC, created_at DESC"
		}
	case "title":
		orderClause = "title " + dir
	case "created_at":
		orderClause = "created_at " + dir
	}

	query = query.Order(orderClause)

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	var tasks []models.Task
	if err := query.Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

func (r *taskRepository) Update(ctx context.Context, task *models.Task) error {
	task.UpdatedAt = time.Now().UTC()
	result := r.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ? AND user_id = ?", task.ID, task.UserID).
		Updates(map[string]interface{}{
			"title":        task.Title,
			"description":  task.Description,
			"category":     task.Category,
			"priority":     task.Priority,
			"status":       task.Status,
			"deadline":     task.Deadline,
			"completed_at": task.CompletedAt,
			"updated_at":   task.UpdatedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id uuid.UUID, userID uuid.UUID, status string, completedAt *time.Time) (*models.Task, error) {
	now := time.Now().UTC()
	updates := map[string]interface{}{
		"status":       status,
		"completed_at": completedAt,
		"updated_at":   now,
	}

	result := r.db.WithContext(ctx).Model(&models.Task{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, ErrTaskNotFound
	}

	return r.FindByID(ctx, id, userID)
}

func (r *taskRepository) SoftDelete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", id, userID).
		Delete(&models.Task{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrTaskNotFound
	}
	return nil
}

func (r *taskRepository) CountByStatus(ctx context.Context, userID uuid.UUID) (map[string]int64, error) {
	type countResult struct {
		Status string
		Count  int64
	}
	var results []countResult
	err := r.db.WithContext(ctx).Model(&models.Task{}).
		Select("status, count(*) as count").
		Where("user_id = ?", userID).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := map[string]int64{
		models.TaskStatusPending:    0,
		models.TaskStatusInProgress: 0,
		models.TaskStatusCompleted:  0,
		models.TaskStatusCancelled:  0,
	}
	for _, res := range results {
		counts[res.Status] = res.Count
	}
	return counts, nil
}

func (r *taskRepository) CountCompletedByDateRange(ctx context.Context, userID uuid.UUID, start, end time.Time) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Task{}).
		Where("user_id = ? AND status = ? AND completed_at >= ? AND completed_at <= ?", userID, models.TaskStatusCompleted, start, end).
		Count(&count).Error
	return count, err
}
