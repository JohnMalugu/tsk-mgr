package repository

import (
	"database/sql"
	"errors"

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
	return 0, nil
}

func (r *sqlTemplateRepository) GetTemplateByID(id int64) (*model.TaskTemplate, error) {
	return nil, nil
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
