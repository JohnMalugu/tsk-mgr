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
	return 0, nil
}

func (r *releaseRepositoryImpl) GetReleaseByID(id int64) (*model.Release, error) {
	return nil, nil
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
