package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var allowedCategories = map[string]bool{
	model.TemplateCategoryGeneral:    true,
	model.TemplateCategoryBugReport:  true,
	model.TemplateCategoryFeature:    true,
	model.TemplateCategoryOnboarding: true,
	model.TemplateCategorySprintTask: true,
	model.TemplateCategoryReview:     true,
}

var allowedPriorities = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
	"urgent": true,
}

func normalizeCategory(category string) (string, error) {
	c := strings.ToLower(strings.TrimSpace(category))
	if c == "" {
		return model.TemplateCategoryGeneral, nil
	}
	if !allowedCategories[c] {
		return "", ErrInvalidCategory
	}
	return c, nil
}

func normalizePriority(priority string) (string, error) {
	p := strings.ToLower(strings.TrimSpace(priority))
	if p == "" {
		return "medium", nil
	}
	if !allowedPriorities[p] {
		return "", ErrInvalidPriority
	}
	return p, nil
}

func validateTemplateFields(title string, wsID int64, userID string, estimatedMinutes int) error {
	if strings.TrimSpace(title) == "" {
		return ErrTemplateTitleRequired
	}
	if wsID <= 0 {
		return ErrInvalidWorkspaceID
	}
	if strings.TrimSpace(userID) == "" {
		return ErrInvalidUserID
	}
	if estimatedMinutes < 0 {
		return ErrNegativeEstimate
	}
	return nil
}

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
	if err := validateTemplateFields(req.Title, req.WorkspaceID, req.CreatedByUserID, req.EstimatedMinutes); err != nil {
		return nil, err
	}

	category, err := normalizeCategory(req.Category)
	if err != nil {
		return nil, err
	}

	priority, err := normalizePriority(req.DefaultPriority)
	if err != nil {
		return nil, err
	}

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}
	checklists := req.ChecklistItems
	if checklists == nil {
		checklists = []model.ChecklistBlueprint{}
	}
	subtasks := req.Subtasks
	if subtasks == nil {
		subtasks = []model.SubtaskBlueprint{}
	}

	template := &model.TaskTemplate{
		WorkspaceID:      req.WorkspaceID,
		CreatedByUserID:  strings.TrimSpace(req.CreatedByUserID),
		Title:            strings.TrimSpace(req.Title),
		Description:      strings.TrimSpace(req.Description),
		Category:         category,
		DefaultPriority:  priority,
		EstimatedMinutes: req.EstimatedMinutes,
		Tags:             tags,
		ChecklistItems:   checklists,
		Subtasks:         subtasks,
		UsageCount:       0,
		IsArchived:       false,
	}

	id, err := s.templateRepo.CreateTemplate(template)
	if err != nil {
		return nil, err
	}

	template.ID = id
	return template, nil
}

func (s *templateService) GetTemplate(id int64) (*model.TaskTemplate, error) {
	if id <= 0 {
		return nil, ErrTemplateNotFound
	}

	t, err := s.templateRepo.GetTemplateByID(id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, ErrTemplateNotFound
	}

	return t, nil
}

func (s *templateService) ListTemplates(filter model.TemplateFilter) (*model.TemplateSummary, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	} else if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	if filter.Category != nil {
		normalized, err := normalizeCategory(*filter.Category)
		if err == nil {
			filter.Category = &normalized
		}
	}

	list, total, err := s.templateRepo.ListTemplates(filter)
	if err != nil {
		return nil, err
	}

	return &model.TemplateSummary{
		Templates: list,
		Total:     total,
		Offset:    filter.Offset,
		Limit:     filter.Limit,
	}, nil
}

func (s *templateService) UpdateTemplate(id int64, req *model.UpdateTemplateRequest) (*model.TaskTemplate, error) {
	t, err := s.GetTemplate(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return nil, ErrTemplateTitleRequired
		}
		t.Title = trimmed
	}
	if req.Description != nil {
		t.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		cat, err := normalizeCategory(*req.Category)
		if err != nil {
			return nil, err
		}
		t.Category = cat
	}
	if req.DefaultPriority != nil {
		prio, err := normalizePriority(*req.DefaultPriority)
		if err != nil {
			return nil, err
		}
		t.DefaultPriority = prio
	}
	if req.EstimatedMinutes != nil {
		if *req.EstimatedMinutes < 0 {
			return nil, ErrNegativeEstimate
		}
		t.EstimatedMinutes = *req.EstimatedMinutes
	}
	if req.Tags != nil {
		t.Tags = *req.Tags
	}
	if req.ChecklistItems != nil {
		t.ChecklistItems = *req.ChecklistItems
	}
	if req.Subtasks != nil {
		t.Subtasks = *req.Subtasks
	}
	if req.IsArchived != nil {
		t.IsArchived = *req.IsArchived
	}

	if err := s.templateRepo.UpdateTemplate(t); err != nil {
		return nil, err
	}

	return t, nil
}

func (s *templateService) DeleteTemplate(id int64) error {
	if _, err := s.GetTemplate(id); err != nil {
		return err
	}
	return s.templateRepo.DeleteTemplate(id)
}

func (s *templateService) DuplicateTemplate(id int64) (*model.TaskTemplate, error) {
	orig, err := s.GetTemplate(id)
	if err != nil {
		return nil, err
	}

	tagsCopy := make([]string, len(orig.Tags))
	copy(tagsCopy, orig.Tags)

	checklistCopy := make([]model.ChecklistBlueprint, len(orig.ChecklistItems))
	copy(checklistCopy, orig.ChecklistItems)

	subtasksCopy := make([]model.SubtaskBlueprint, len(orig.Subtasks))
	copy(subtasksCopy, orig.Subtasks)

	cloned := &model.TaskTemplate{
		WorkspaceID:      orig.WorkspaceID,
		CreatedByUserID:  orig.CreatedByUserID,
		Title:            fmt.Sprintf("Copy of %s", orig.Title),
		Description:      orig.Description,
		Category:         orig.Category,
		DefaultPriority:  orig.DefaultPriority,
		EstimatedMinutes: orig.EstimatedMinutes,
		Tags:             tagsCopy,
		ChecklistItems:   checklistCopy,
		Subtasks:         subtasksCopy,
		UsageCount:       0,
		IsArchived:       false,
	}

	newID, err := s.templateRepo.CreateTemplate(cloned)
	if err != nil {
		return nil, err
	}

	cloned.ID = newID
	return cloned, nil
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
