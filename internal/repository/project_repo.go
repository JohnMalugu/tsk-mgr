package repository

import (
	"database/sql"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

type ProjectRepository interface {
	Create(project *model.Project) error
	GetByID(id int) (*model.Project, error)
	GetByUserID(userID string) ([]*model.Project, error)
	Update(project *model.Project) error
	Delete(id int) error
}

type projectRepositoryImpl struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) ProjectRepository {
	return &projectRepositoryImpl{db: db}
}
