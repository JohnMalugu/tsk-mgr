package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrReleaseNotFound     = errors.New("release not found")
	ErrReleaseItemNotFound = errors.New("release item not found")
)

// ReleaseRepository defines data-access operations for releases and attached task items
type ReleaseRepository interface {
	CreateRelease(r *model.Release) (int64, error)
	GetReleaseByID(id int64) (*model.Release, error)
	ListReleases(filter model.ReleaseFilter) ([]*model.ReleaseSummary, error)
	UpdateRelease(r *model.Release) error
	UpdateReleaseStatus(id int64, status string, releasedAt *time.Time) error
	DeleteRelease(id int64) error

	AddReleaseItem(item *model.ReleaseItem) error
	RemoveReleaseItem(releaseID int64, taskID int) error
	UpdateReleaseItem(item *model.ReleaseItem) error
	GetReleaseItem(releaseID int64, taskID int) (*model.ReleaseItem, error)
	ListReleaseItems(releaseID int64, category string) ([]*model.ReleaseItem, error)
	GetIncompleteTasksForRelease(releaseID int64) ([]*model.ReleaseItem, error)
}

type releaseRepositoryImpl struct {
	db *sql.DB
}

// NewReleaseRepository creates a new SQLite-backed ReleaseRepository
func NewReleaseRepository(db *sql.DB) ReleaseRepository {
	return &releaseRepositoryImpl{db: db}
}

func (r *releaseRepositoryImpl) CreateRelease(rel *model.Release) (int64, error) {
	now := time.Now()
	rel.CreatedAt = now
	rel.UpdatedAt = now
	if rel.Status == "" {
		rel.Status = model.ReleaseStatusPlanning
	}

	query := `
	INSERT INTO releases (workspace_id, project_id, name, version, description, status, target_date, released_at, created_by_user_id, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		rel.WorkspaceID, rel.ProjectID, rel.Name, rel.Version, rel.Description,
		rel.Status, rel.TargetDate, rel.ReleasedAt, rel.CreatedByUserID,
		rel.CreatedAt, rel.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	rel.ID = id
	return id, nil
}

func (r *releaseRepositoryImpl) GetReleaseByID(id int64) (*model.Release, error) {
	query := `
	SELECT id, workspace_id, project_id, name, version, description, status, target_date, released_at, created_by_user_id, created_at, updated_at
	FROM releases
	WHERE id = ?`

	row := r.db.QueryRow(query, id)
	var rel model.Release
	var targetDate, releasedAt sql.NullTime
	var projectID sql.NullInt64

	err := row.Scan(
		&rel.ID, &rel.WorkspaceID, &projectID, &rel.Name, &rel.Version,
		&rel.Description, &rel.Status, &targetDate, &releasedAt,
		&rel.CreatedByUserID, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}
	if projectID.Valid {
		rel.ProjectID = &projectID.Int64
	}
	if targetDate.Valid {
		rel.TargetDate = &targetDate.Time
	}
	if releasedAt.Valid {
		rel.ReleasedAt = &releasedAt.Time
	}
	return &rel, nil
}

func (r *releaseRepositoryImpl) ListReleases(filter model.ReleaseFilter) ([]*model.ReleaseSummary, error) {
	query := `
	SELECT 
		r.id, r.workspace_id, r.project_id, r.name, r.version, r.status, r.target_date, r.released_at, r.created_at,
		COUNT(ri.task_id) AS task_count,
		COALESCE(SUM(CASE WHEN t.status = 'completed' OR t.completed = 1 THEN 1 ELSE 0 END), 0) AS completed_count
	FROM releases r
	LEFT JOIN release_items ri ON ri.release_id = r.id
	LEFT JOIN tasks t ON t.id = ri.task_id
	WHERE 1=1`

	var args []interface{}
	if filter.WorkspaceID > 0 {
		query += " AND r.workspace_id = ?"
		args = append(args, filter.WorkspaceID)
	}
	if filter.ProjectID != nil && *filter.ProjectID > 0 {
		query += " AND r.project_id = ?"
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != "" {
		query += " AND r.status = ?"
		args = append(args, filter.Status)
	}
	if filter.Query != "" {
		query += " AND (r.name LIKE ? OR r.version LIKE ?)"
		pattern := "%" + filter.Query + "%"
		args = append(args, pattern, pattern)
	}

	query += " GROUP BY r.id ORDER BY r.created_at DESC"

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query += " LIMIT ?"
	args = append(args, limit)

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*model.ReleaseSummary
	for rows.Next() {
		var s model.ReleaseSummary
		var projectID sql.NullInt64
		var targetDate, releasedAt sql.NullTime

		err := rows.Scan(
			&s.ID, &s.WorkspaceID, &projectID, &s.Name, &s.Version, &s.Status,
			&targetDate, &releasedAt, &s.CreatedAt,
			&s.TaskCount, &s.CompletedCount,
		)
		if err != nil {
			return nil, err
		}
		if projectID.Valid {
			s.ProjectID = &projectID.Int64
		}
		if targetDate.Valid {
			s.TargetDate = &targetDate.Time
		}
		if releasedAt.Valid {
			s.ReleasedAt = &releasedAt.Time
		}
		if s.TaskCount > 0 {
			s.CompletionPercentage = float64(s.CompletedCount) / float64(s.TaskCount) * 100.0
		}
		summaries = append(summaries, &s)
	}

	return summaries, rows.Err()
}

func (r *releaseRepositoryImpl) UpdateRelease(rel *model.Release) error {
	rel.UpdatedAt = time.Now()
	query := `
	UPDATE releases
	SET name = ?, version = ?, description = ?, target_date = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, rel.Name, rel.Version, rel.Description, rel.TargetDate, rel.UpdatedAt, rel.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrReleaseNotFound
	}
	return nil
}

func (r *releaseRepositoryImpl) UpdateReleaseStatus(id int64, status string, releasedAt *time.Time) error {
	now := time.Now()
	query := `
	UPDATE releases
	SET status = ?, released_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query, status, releasedAt, now, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrReleaseNotFound
	}
	return nil
}

func (r *releaseRepositoryImpl) DeleteRelease(id int64) error {
	query := `DELETE FROM releases WHERE id = ?`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrReleaseNotFound
	}
	return nil
}

func (r *releaseRepositoryImpl) AddReleaseItem(item *model.ReleaseItem) error {
	return nil
}

func (r *releaseRepositoryImpl) RemoveReleaseItem(releaseID int64, taskID int) error {
	return nil
}

func (r *releaseRepositoryImpl) UpdateReleaseItem(item *model.ReleaseItem) error {
	return nil
}

func (r *releaseRepositoryImpl) GetReleaseItem(releaseID int64, taskID int) (*model.ReleaseItem, error) {
	return nil, nil
}

func (r *releaseRepositoryImpl) ListReleaseItems(releaseID int64, category string) ([]*model.ReleaseItem, error) {
	return nil, nil
}

func (r *releaseRepositoryImpl) GetIncompleteTasksForRelease(releaseID int64) ([]*model.ReleaseItem, error) {
	return nil, nil
}
