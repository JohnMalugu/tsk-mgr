package repository

import (
	"database/sql"
	"errors"

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
	return 0, nil
}

func (r *sqlSavedViewRepository) GetViewByID(id int64) (*model.SavedView, error) {
	return nil, nil
}

func (r *sqlSavedViewRepository) ListViews(workspaceID int64, userID string, onlyPinned bool) ([]*model.SavedView, error) {
	return nil, nil
}

func (r *sqlSavedViewRepository) UpdateView(v *model.SavedView) error {
	return nil
}

func (r *sqlSavedViewRepository) DeleteView(id int64) error {
	return nil
}

func (r *sqlSavedViewRepository) TogglePinView(id int64) (bool, error) {
	return false, nil
}

func (r *sqlSavedViewRepository) IncrementViewUsage(id int64) error {
	return nil
}

