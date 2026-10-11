package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrUnauthorizedProject = errors.New("unauthorized to access this project")
	ErrProjectNameRequired = errors.New("project name is required")
)

type ProjectService interface {
	CreateProject(userID, name, description, color string) (*model.Project, error)
	GetProject(id int, userID string) (*model.Project, error)
	GetUserProjects(userID string) ([]*model.Project, error)
	UpdateProject(id int, userID, name, description, color string) (*model.Project, error)
	DeleteProject(id int, userID string) error
}

type projectServiceImpl struct {
	repo repository.ProjectRepository
}

func NewProjectService(repo repository.ProjectRepository) ProjectService {
	return &projectServiceImpl{repo: repo}
}
