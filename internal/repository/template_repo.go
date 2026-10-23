package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

// ErrTemplateNotFound is returned when a requested task template does not exist
var ErrTemplateNotFound = errors.New("template not found")

// TemplateRepository defines persistence operations for task templates and usage analytics
type TemplateRepository interface {
	CreateTemplate(t *model.TaskTemplate) (int64, error)
	GetTemplateByID(id int64) (*model.TaskTemplate, error)
	ListTemplates(filter model.TemplateFilter) ([]model.TaskTemplate, int, error)
	UpdateTemplate(t *model.TaskTemplate) error
	DeleteTemplate(id int64) error
	IncrementTemplateUsage(id int64) error
	GetTemplateStats(workspaceID int64) (*model.TemplateStats, error)
}

type sqlTemplateRepository struct {
	db *sql.DB
}

// NewTemplateRepository creates a new TemplateRepository instance backed by SQLite
func NewTemplateRepository(db *sql.DB) TemplateRepository {
	return &sqlTemplateRepository{db: db}
}

func (r *sqlTemplateRepository) CreateTemplate(t *model.TaskTemplate) (int64, error) {
	now := time.Now()
	if t.CreatedAt.IsZero() {
		t.CreatedAt = now
	}
	if t.UpdatedAt.IsZero() {
		t.UpdatedAt = now
	}

	tagsJSON, err := json.Marshal(t.Tags)
	if err != nil {
		tagsJSON = []byte("[]")
	}

	checklistJSON, err := json.Marshal(t.ChecklistItems)
	if err != nil {
		checklistJSON = []byte("[]")
	}

	subtasksJSON, err := json.Marshal(t.Subtasks)
	if err != nil {
		subtasksJSON = []byte("[]")
	}

	query := `
	INSERT INTO task_templates (
		workspace_id, created_by_user_id, title, description, category,
		default_priority, estimated_minutes, tags, checklist_items, subtasks,
		usage_count, is_archived, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		t.WorkspaceID,
		t.CreatedByUserID,
		t.Title,
		t.Description,
		t.Category,
		t.DefaultPriority,
		t.EstimatedMinutes,
		string(tagsJSON),
		string(checklistJSON),
		string(subtasksJSON),
		t.UsageCount,
		t.IsArchived,
		t.CreatedAt,
		t.UpdatedAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert task template: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last insert id: %w", err)
	}

	t.ID = id
	return id, nil
}

func (r *sqlTemplateRepository) GetTemplateByID(id int64) (*model.TaskTemplate, error) {
	query := `
	SELECT id, workspace_id, created_by_user_id, title, description, category,
		default_priority, estimated_minutes, tags, checklist_items, subtasks,
		usage_count, is_archived, created_at, updated_at, last_used_at
	FROM task_templates
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var t model.TaskTemplate
	var tagsJSON, checklistJSON, subtasksJSON string
	var lastUsedAt sql.NullTime

	err := row.Scan(
		&t.ID,
		&t.WorkspaceID,
		&t.CreatedByUserID,
		&t.Title,
		&t.Description,
		&t.Category,
		&t.DefaultPriority,
		&t.EstimatedMinutes,
		&tagsJSON,
		&checklistJSON,
		&subtasksJSON,
		&t.UsageCount,
		&t.IsArchived,
		&t.CreatedAt,
		&t.UpdatedAt,
		&lastUsedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get template %d: %w", id, err)
	}

	if lastUsedAt.Valid {
		t.LastUsedAt = &lastUsedAt.Time
	}

	if err := json.Unmarshal([]byte(tagsJSON), &t.Tags); err != nil {
		t.Tags = []string{}
	}
	if err := json.Unmarshal([]byte(checklistJSON), &t.ChecklistItems); err != nil {
		t.ChecklistItems = []model.ChecklistBlueprint{}
	}
	if err := json.Unmarshal([]byte(subtasksJSON), &t.Subtasks); err != nil {
		t.Subtasks = []model.SubtaskBlueprint{}
	}

	return &t, nil
}

func (r *sqlTemplateRepository) ListTemplates(filter model.TemplateFilter) ([]model.TaskTemplate, int, error) {
	return nil, 0, nil
}

func (r *sqlTemplateRepository) UpdateTemplate(t *model.TaskTemplate) error {
	return nil
}

func (r *sqlTemplateRepository) DeleteTemplate(id int64) error {
	return nil
}

func (r *sqlTemplateRepository) IncrementTemplateUsage(id int64) error {
	return nil
}

func (r *sqlTemplateRepository) GetTemplateStats(workspaceID int64) (*model.TemplateStats, error) {
	return nil, nil
}
