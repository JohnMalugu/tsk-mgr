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

// ErrViewNotFound is returned when a requested saved view does not exist
var ErrViewNotFound = errors.New("saved view not found")

// SavedViewRepository defines persistence operations for saved filter views
type SavedViewRepository interface {
	CreateView(v *model.SavedView) (int64, error)
	GetViewByID(id int64) (*model.SavedView, error)
	ListViews(workspaceID int64, userID string, onlyPinned bool) ([]*model.SavedView, error)
	UpdateView(v *model.SavedView) error
	DeleteView(id int64) error
	TogglePinView(id int64) (bool, error)
	IncrementViewUsage(id int64) error
}

type sqlSavedViewRepository struct {
	db *sql.DB
}

// NewSavedViewRepository instantiates a new SQLite-backed SavedViewRepository
func NewSavedViewRepository(db *sql.DB) SavedViewRepository {
	return &sqlSavedViewRepository{db: db}
}

func (r *sqlSavedViewRepository) CreateView(v *model.SavedView) (int64, error) {
	now := time.Now()
	if v.CreatedAt.IsZero() {
		v.CreatedAt = now
	}
	if v.UpdatedAt.IsZero() {
		v.UpdatedAt = now
	}
	if v.Icon == "" {
		v.Icon = "bookmark"
	}
	if v.Color == "" {
		v.Color = "#4F46E5"
	}
	if v.SortBy == "" {
		v.SortBy = model.SortByCreatedAt
	}
	if v.SortOrder == "" {
		v.SortOrder = model.SortOrderDesc
	}
	if v.GroupBy == "" {
		v.GroupBy = model.GroupByNone
	}

	filterJSON, err := json.Marshal(v.FilterCriteria)
	if err != nil {
		filterJSON = []byte("{}")
	}

	query := `
	INSERT INTO saved_views (
		workspace_id, user_id, name, description, icon, color,
		filter_criteria, sort_by, sort_order, group_by,
		is_shared, is_pinned, position, usage_count, last_used_at,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	sharedInt := 0
	if v.IsShared {
		sharedInt = 1
	}
	pinnedInt := 0
	if v.IsPinned {
		pinnedInt = 1
	}

	res, err := r.db.Exec(query,
		v.WorkspaceID, v.UserID, v.Name, v.Description, v.Icon, v.Color,
		string(filterJSON), v.SortBy, v.SortOrder, v.GroupBy,
		sharedInt, pinnedInt, v.Position, v.UsageCount, v.LastUsedAt,
		v.CreatedAt, v.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	v.ID = id
	return id, nil
}

func (r *sqlSavedViewRepository) GetViewByID(id int64) (*model.SavedView, error) {
	query := `
	SELECT id, workspace_id, user_id, name, description, icon, color,
	       filter_criteria, sort_by, sort_order, group_by,
	       is_shared, is_pinned, position, usage_count, last_used_at,
	       created_at, updated_at
	FROM saved_views
	WHERE id = ?`

	row := r.db.QueryRow(query, id)

	var v model.SavedView
	var filterJSON string
	var sharedInt, pinnedInt int

	err := row.Scan(
		&v.ID, &v.WorkspaceID, &v.UserID, &v.Name, &v.Description, &v.Icon, &v.Color,
		&filterJSON, &v.SortBy, &v.SortOrder, &v.GroupBy,
		&sharedInt, &pinnedInt, &v.Position, &v.UsageCount, &v.LastUsedAt,
		&v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrViewNotFound
		}
		return nil, err
	}

	v.IsShared = sharedInt == 1
	v.IsPinned = pinnedInt == 1

	if filterJSON != "" {
		_ = json.Unmarshal([]byte(filterJSON), &v.FilterCriteria)
	}

	return &v, nil
}


func (r *sqlSavedViewRepository) ListViews(workspaceID int64, userID string, onlyPinned bool) ([]*model.SavedView, error) {
	var conditions []string
	var args []interface{}

	conditions = append(conditions, "workspace_id = ?")
	args = append(args, workspaceID)

	if userID != "" {
		conditions = append(conditions, "(is_shared = 1 OR user_id = ?)")
		args = append(args, userID)
	}

	if onlyPinned {
		conditions = append(conditions, "is_pinned = 1")
	}

	whereClause := strings.Join(conditions, " AND ")
	query := fmt.Sprintf(`
	SELECT id, workspace_id, user_id, name, description, icon, color,
	       filter_criteria, sort_by, sort_order, group_by,
	       is_shared, is_pinned, position, usage_count, last_used_at,
	       created_at, updated_at
	FROM saved_views
	WHERE %s
	ORDER BY is_pinned DESC, position ASC, name ASC`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var views []*model.SavedView
	for rows.Next() {
		var v model.SavedView
		var filterJSON string
		var sharedInt, pinnedInt int

		err := rows.Scan(
			&v.ID, &v.WorkspaceID, &v.UserID, &v.Name, &v.Description, &v.Icon, &v.Color,
			&filterJSON, &v.SortBy, &v.SortOrder, &v.GroupBy,
			&sharedInt, &pinnedInt, &v.Position, &v.UsageCount, &v.LastUsedAt,
			&v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		v.IsShared = sharedInt == 1
		v.IsPinned = pinnedInt == 1

		if filterJSON != "" {
			_ = json.Unmarshal([]byte(filterJSON), &v.FilterCriteria)
		}

		views = append(views, &v)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return views, nil
}


func (r *sqlSavedViewRepository) UpdateView(v *model.SavedView) error {
	v.UpdatedAt = time.Now()

	filterJSON, err := json.Marshal(v.FilterCriteria)
	if err != nil {
		filterJSON = []byte("{}")
	}

	sharedInt := 0
	if v.IsShared {
		sharedInt = 1
	}
	pinnedInt := 0
	if v.IsPinned {
		pinnedInt = 1
	}

	query := `
	UPDATE saved_views SET
		name = ?,
		description = ?,
		icon = ?,
		color = ?,
		filter_criteria = ?,
		sort_by = ?,
		sort_order = ?,
		group_by = ?,
		is_shared = ?,
		is_pinned = ?,
		position = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		v.Name, v.Description, v.Icon, v.Color,
		string(filterJSON), v.SortBy, v.SortOrder, v.GroupBy,
		sharedInt, pinnedInt, v.Position, v.UpdatedAt,
		v.ID,
	)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrViewNotFound
	}

	return nil
}

func (r *sqlSavedViewRepository) DeleteView(id int64) error {
	query := `DELETE FROM saved_views WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrViewNotFound
	}

	return nil
}


func (r *sqlSavedViewRepository) TogglePinView(id int64) (bool, error) {
	return false, nil
}

func (r *sqlSavedViewRepository) IncrementViewUsage(id int64) error {
	return nil
}

