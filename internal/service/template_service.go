package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrTemplateNotFound      = errors.New("task template not found")
	ErrTemplateTitleRequired = errors.New("template title is required")
	ErrInvalidCategory       = errors.New("invalid template category")
	ErrInvalidPriority       = errors.New("invalid default priority")
	ErrNegativeEstimate      = errors.New("estimated minutes cannot be negative")
	ErrTemplateArchived      = errors.New("cannot instantiate an archived template")
	ErrInvalidBundle         = errors.New("invalid template bundle format")
)

// TemplateService orchestrates business logic for task templates, instantiations, and bundles
type TemplateService interface {
	CreateTemplate(req *model.CreateTemplateRequest) (*model.TaskTemplate, error)
	GetTemplate(id int64) (*model.TaskTemplate, error)
	ListTemplates(filter model.TemplateFilter) (*model.TemplateSummary, error)
	UpdateTemplate(id int64, req *model.UpdateTemplateRequest) (*model.TaskTemplate, error)
	DeleteTemplate(id int64) error
	DuplicateTemplate(id int64) (*model.TaskTemplate, error)
	InstantiateTask(templateID int64, req *model.TemplateInstantiationRequest) (*model.InstantiatedTaskResult, error)
	ExportTemplate(id int64) (*model.TemplateBundle, error)
	ImportTemplate(req *model.ImportTemplateRequest) (*model.TaskTemplate, error)
	GetTemplateStats(workspaceID int64) (*model.TemplateStats, error)
}

type templateService struct {
	templateRepo repository.TemplateRepository
	taskRepo     repository.TaskRepository
	subtaskRepo  repository.SubtaskRepository
}

// NewTemplateService creates a new TemplateService instance
func NewTemplateService(
	templateRepo repository.TemplateRepository,
	taskRepo repository.TaskRepository,
	subtaskRepo repository.SubtaskRepository,
) TemplateService {
	return &templateService{
		templateRepo: templateRepo,
		taskRepo:     taskRepo,
		subtaskRepo:  subtaskRepo,
	}
}

func (s *templateService) CreateTemplate(req *model.CreateTemplateRequest) (*model.TaskTemplate, error) {
	return nil, nil
}

func (s *templateService) GetTemplate(id int64) (*model.TaskTemplate, error) {
	return nil, nil
}

func (s *templateService) ListTemplates(filter model.TemplateFilter) (*model.TemplateSummary, error) {
	return nil, nil
}

func (s *templateService) UpdateTemplate(id int64, req *model.UpdateTemplateRequest) (*model.TaskTemplate, error) {
	return nil, nil
}

func (s *templateService) DeleteTemplate(id int64) error {
	return nil
}

func (s *templateService) DuplicateTemplate(id int64) (*model.TaskTemplate, error) {
	return nil, nil
}

func (s *templateService) InstantiateTask(templateID int64, req *model.TemplateInstantiationRequest) (*model.InstantiatedTaskResult, error) {
	return nil, nil
}

func (s *templateService) ExportTemplate(id int64) (*model.TemplateBundle, error) {
	return nil, nil
}

func (s *templateService) ImportTemplate(req *model.ImportTemplateRequest) (*model.TaskTemplate, error) {
	return nil, nil
}

func (s *templateService) GetTemplateStats(workspaceID int64) (*model.TemplateStats, error) {
	return nil, nil
}
