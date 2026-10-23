package repository

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID != nil {
		conditions = append(conditions, "workspace_id = ?")
		args = append(args, *filter.WorkspaceID)
	}
	if filter.Category != nil {
		conditions = append(conditions, "category = ?")
		args = append(args, *filter.Category)
	}
	if filter.SearchQuery != nil && strings.TrimSpace(*filter.SearchQuery) != "" {
		conditions = append(conditions, "(title LIKE ? OR description LIKE ?)")
		pattern := "%" + strings.TrimSpace(*filter.SearchQuery) + "%"
		args = append(args, pattern, pattern)
	}
	if filter.IsArchived != nil {
		conditions = append(conditions, "is_archived = ?")
		args = append(args, *filter.IsArchived)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM task_templates %s", whereClause)
	var total int
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count templates: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, created_by_user_id, title, description, category,
		default_priority, estimated_minutes, tags, checklist_items, subtasks,
		usage_count, is_archived, created_at, updated_at, last_used_at
	FROM task_templates
	%s
	ORDER BY usage_count DESC, id DESC
	LIMIT ? OFFSET ?`, whereClause)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.Query(query, queryArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query templates: %w", err)
	}
	defer rows.Close()

	var templates []model.TaskTemplate
	for rows.Next() {
		var t model.TaskTemplate
		var tagsJSON, checklistJSON, subtasksJSON string
		var lastUsedAt sql.NullTime

		err := rows.Scan(
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
			return nil, 0, fmt.Errorf("failed to scan template row: %w", err)
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

		templates = append(templates, t)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error reading template rows: %w", err)
	}

	return templates, total, nil
}

func (r *sqlTemplateRepository) UpdateTemplate(t *model.TaskTemplate) error {
	t.UpdatedAt = time.Now()

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
	UPDATE task_templates
	SET title = ?, description = ?, category = ?, default_priority = ?,
		estimated_minutes = ?, tags = ?, checklist_items = ?, subtasks = ?,
		is_archived = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		t.Title,
		t.Description,
		t.Category,
		t.DefaultPriority,
		t.EstimatedMinutes,
		string(tagsJSON),
		string(checklistJSON),
		string(subtasksJSON),
		t.IsArchived,
		t.UpdatedAt,
		t.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update template %d: %w", t.ID, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

func (r *sqlTemplateRepository) DeleteTemplate(id int64) error {
	query := `DELETE FROM task_templates WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete template %d: %w", id, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

func (r *sqlTemplateRepository) IncrementTemplateUsage(id int64) error {
	query := `
	UPDATE task_templates
	SET usage_count = usage_count + 1, last_used_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to increment template usage: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrTemplateNotFound
	}

	return nil
}

func (r *sqlTemplateRepository) GetTemplateStats(workspaceID int64) (*model.TemplateStats, error) {
	return nil, nil
}
