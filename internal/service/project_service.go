package service

import (
	"errors"
	"strings"
	"time"

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

func (s *projectServiceImpl) CreateProject(userID, name, description, color string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrProjectNameRequired
	}

	if color == "" {
		color = "#6366f1"
	}

	now := time.Now().UTC()
	project := &model.Project{
		UserID:      userID,
		Name:        name,
		Description: strings.TrimSpace(description),
		Color:       color,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectServiceImpl) GetProject(id int, userID string) (*model.Project, error) {
	project, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if project == nil {
		return nil, ErrProjectNotFound
	}
	if project.UserID != userID {
		return nil, ErrUnauthorizedProject
	}
	return project, nil
}

func (s *projectServiceImpl) GetUserProjects(userID string) ([]*model.Project, error) {
	return s.repo.GetByUserID(userID)
}

func (s *projectServiceImpl) UpdateProject(id int, userID, name, description, color string) (*model.Project, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrProjectNameRequired
	}

	project, err := s.GetProject(id, userID)
	if err != nil {
		return nil, err
	}

	project.Name = name
	project.Description = strings.TrimSpace(description)
	if color != "" {
		project.Color = color
	}
	project.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *projectServiceImpl) DeleteProject(id int, userID string) error {
	_, err := s.GetProject(id, userID)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}
