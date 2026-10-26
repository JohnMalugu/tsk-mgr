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
	return nil, nil
}

func (r *releaseRepositoryImpl) UpdateRelease(rel *model.Release) error {
	return nil
}

func (r *releaseRepositoryImpl) UpdateReleaseStatus(id int64, status string, releasedAt *time.Time) error {
	return nil
}

func (r *releaseRepositoryImpl) DeleteRelease(id int64) error {
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
