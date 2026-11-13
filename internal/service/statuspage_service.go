package service

import (
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrStatusPageNotFound        = errors.New("status page not found")
	ErrStatusPageNameRequired    = errors.New("status page name is required")
	ErrStatusPageSlugRequired    = errors.New("status page slug is required")
	ErrInvalidStatusPageSlug     = errors.New("invalid status page slug: alphanumeric and hyphens only")
	ErrDuplicateStatusPageSlug   = errors.New("status page slug already in use")
	ErrComponentNotFound         = errors.New("status component not found")
	ErrComponentNameRequired     = errors.New("status component name is required")
	ErrInvalidComponentStatus    = errors.New("invalid component status")
	ErrStatusIncidentNotFound    = errors.New("status incident not found")
	ErrStatusIncidentTitleReq    = errors.New("status incident title is required")
	ErrInvalidStatusIncidentSt   = errors.New("invalid status incident status")
	ErrInvalidIncidentImpact     = errors.New("invalid incident impact level")
	ErrStatusMaintenanceNotFound = errors.New("maintenance window not found")
	ErrMaintenanceTitleReq       = errors.New("maintenance window title is required")
	ErrInvalidMaintenanceTimes   = errors.New("maintenance start time must be before end time")
	ErrSubscriberNotFound        = errors.New("status subscriber not found")
	ErrSubscriberTargetRequired  = errors.New("subscriber target endpoint or email is required")
	ErrInvalidSubscriberType     = errors.New("invalid subscriber type: must be email, webhook, or slack")
)

// StatusPageService coordinates status pages, service component trees, incident broadcasts, and subscriber alerts
type StatusPageService interface {
	// Status Pages
	CreatePage(creatorID string, req model.CreateStatusPageRequest) (*model.StatusPage, error)
	GetPage(id int64) (*model.StatusPage, error)
	GetPageBySlug(slug string) (*model.StatusPage, error)
	ListPages(filter model.StatusPageFilter) ([]model.StatusPage, int, error)
	UpdatePage(id int64, req model.UpdateStatusPageRequest) (*model.StatusPage, error)
	DeletePage(id int64) error

	// Component Groups
	CreateComponentGroup(pageID int64, req model.CreateComponentGroupRequest) (*model.ComponentGroup, error)
	ListComponentGroups(pageID int64) ([]model.ComponentGroup, error)
	DeleteComponentGroup(id int64) error

	// Components
	CreateComponent(pageID int64, req model.CreateComponentRequest) (*model.StatusComponent, error)
	GetComponent(id int64) (*model.StatusComponent, error)
	ListComponents(pageID int64) ([]model.StatusComponent, error)
	UpdateComponent(id int64, req model.UpdateComponentRequest) (*model.StatusComponent, error)
	DeleteComponent(id int64) error
	SetComponentStatus(id int64, status string) error

	// Status Incidents & Timeline Updates
	CreateIncident(creatorID string, pageID int64, req model.CreateStatusIncidentRequest) (*model.StatusIncident, error)
	GetIncident(id int64) (*model.StatusIncident, error)
	ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error)
	UpdateIncident(id int64, req model.UpdateStatusIncidentRequest) (*model.StatusIncident, error)
	AddIncidentUpdate(postedByID string, incidentID int64, req model.AddStatusIncidentUpdateRequest) (*model.StatusIncidentUpdate, error)
	DeleteIncident(id int64) error

	// Maintenance Windows
	CreateMaintenance(creatorID string, pageID int64, req model.CreateMaintenanceRequest) (*model.StatusMaintenanceWindow, error)
	GetMaintenance(id int64) (*model.StatusMaintenanceWindow, error)
	ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error)
	UpdateMaintenance(id int64, req model.UpdateMaintenanceRequest) (*model.StatusMaintenanceWindow, error)
	DeleteMaintenance(id int64) error

	// Subscribers & Metrics
	CreateSubscriber(pageID int64, req model.CreateSubscriberRequest) (*model.StatusSubscriber, error)
	ListSubscribers(pageID int64) ([]model.StatusSubscriber, error)
	DeleteSubscriber(id int64) error
	GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error)
	GetPublicStatusOverview(slug string) (*model.StatusPage, []model.ComponentGroup, []model.StatusIncident, error)
}

type statusPageServiceImpl struct {
	repo         repository.StatusPageRepository
	projectRepo  repository.ProjectRepository
	cmdbRepo     repository.CMDBRepository
	incidentRepo repository.IncidentRepository
}

// NewStatusPageService creates a new StatusPageService instance
func NewStatusPageService(
	repo repository.StatusPageRepository,
	projectRepo repository.ProjectRepository,
	cmdbRepo repository.CMDBRepository,
	incidentRepo repository.IncidentRepository,
) StatusPageService {
	return &statusPageServiceImpl{
		repo:         repo,
		projectRepo:  projectRepo,
		cmdbRepo:     cmdbRepo,
		incidentRepo: incidentRepo,
	}
}

func (s *statusPageServiceImpl) CreatePage(creatorID string, req model.CreateStatusPageRequest) (*model.StatusPage, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrStatusPageNameRequired
	}

	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = model.NormalizeSlug(name)
	} else {
		slug = model.NormalizeSlug(slug)
	}

	if err := model.ValidateSlug(slug); err != nil {
		return nil, err
	}

	visibility := strings.ToLower(strings.TrimSpace(req.Visibility))
	if visibility == "" {
		visibility = model.VisibilityPublic
	}

	theme := strings.ToLower(strings.TrimSpace(req.Theme))
	if theme == "" {
		theme = model.ThemeSystem
	}

	page := &model.StatusPage{
		WorkspaceID:   req.WorkspaceID,
		ProjectID:     req.ProjectID,
		Name:          name,
		Slug:          slug,
		Description:   strings.TrimSpace(req.Description),
		Headline:      strings.TrimSpace(req.Headline),
		Visibility:    visibility,
		CustomDomain:  strings.TrimSpace(req.CustomDomain),
		Theme:         theme,
		CurrentStatus: model.StatusOperational,
		IsActive:      true,
		CreatedByID:   creatorID,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	id, err := s.repo.CreatePage(page)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateStatusPageSlug) {
			return nil, ErrDuplicateStatusPageSlug
		}
		return nil, err
	}

	return s.GetPage(id)
}

func (s *statusPageServiceImpl) GetPage(id int64) (*model.StatusPage, error) {
	page, err := s.repo.GetPageByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}
	return page, nil
}

func (s *statusPageServiceImpl) GetPageBySlug(slug string) (*model.StatusPage, error) {
	cleanSlug := strings.ToLower(strings.TrimSpace(slug))
	page, err := s.repo.GetPageBySlug(cleanSlug)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}
	return page, nil
}

func (s *statusPageServiceImpl) ListPages(filter model.StatusPageFilter) ([]model.StatusPage, int, error) {
	return s.repo.ListPages(filter)
}

func (s *statusPageServiceImpl) UpdatePage(id int64, req model.UpdateStatusPageRequest) (*model.StatusPage, error) {
	existing, err := s.repo.GetPageByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrStatusPageNameRequired
		}
		existing.Name = name
	}
	if req.Slug != nil {
		slug := model.NormalizeSlug(*req.Slug)
		if err := model.ValidateSlug(slug); err != nil {
			return nil, err
		}
		existing.Slug = slug
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.Headline != nil {
		existing.Headline = strings.TrimSpace(*req.Headline)
	}
	if req.Visibility != nil {
		vis := strings.ToLower(strings.TrimSpace(*req.Visibility))
		if vis == model.VisibilityPublic || vis == model.VisibilityPrivate {
			existing.Visibility = vis
		}
	}
	if req.CustomDomain != nil {
		existing.CustomDomain = strings.TrimSpace(*req.CustomDomain)
	}
	if req.Theme != nil {
		theme := strings.ToLower(strings.TrimSpace(*req.Theme))
		if theme == model.ThemeLight || theme == model.ThemeDark || theme == model.ThemeSystem {
			existing.Theme = theme
		}
	}
	if req.CurrentStatus != nil {
		st := strings.ToLower(strings.TrimSpace(*req.CurrentStatus))
		if err := model.ValidateComponentStatus(st); err != nil {
			return nil, err
		}
		existing.CurrentStatus = st
	}
	if req.IsActive != nil {
		existing.IsActive = *req.IsActive
	}

	if err := s.repo.UpdatePage(existing); err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	return s.GetPage(id)
}

func (s *statusPageServiceImpl) DeletePage(id int64) error {
	err := s.repo.DeletePage(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return ErrStatusPageNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) CreateComponentGroup(pageID int64, req model.CreateComponentGroupRequest) (*model.ComponentGroup, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) ListComponentGroups(pageID int64) ([]model.ComponentGroup, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) DeleteComponentGroup(id int64) error {
	return nil
}

func (s *statusPageServiceImpl) CreateComponent(pageID int64, req model.CreateComponentRequest) (*model.StatusComponent, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) GetComponent(id int64) (*model.StatusComponent, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) ListComponents(pageID int64) ([]model.StatusComponent, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) UpdateComponent(id int64, req model.UpdateComponentRequest) (*model.StatusComponent, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) DeleteComponent(id int64) error {
	return nil
}

func (s *statusPageServiceImpl) SetComponentStatus(id int64, status string) error {
	return nil
}

func (s *statusPageServiceImpl) CreateIncident(creatorID string, pageID int64, req model.CreateStatusIncidentRequest) (*model.StatusIncident, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) GetIncident(id int64) (*model.StatusIncident, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error) {
	return nil, 0, nil
}

func (s *statusPageServiceImpl) UpdateIncident(id int64, req model.UpdateStatusIncidentRequest) (*model.StatusIncident, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) AddIncidentUpdate(postedByID string, incidentID int64, req model.AddStatusIncidentUpdateRequest) (*model.StatusIncidentUpdate, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) DeleteIncident(id int64) error {
	return nil
}

func (s *statusPageServiceImpl) CreateMaintenance(creatorID string, pageID int64, req model.CreateMaintenanceRequest) (*model.StatusMaintenanceWindow, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) GetMaintenance(id int64) (*model.StatusMaintenanceWindow, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error) {
	return nil, 0, nil
}

func (s *statusPageServiceImpl) UpdateMaintenance(id int64, req model.UpdateMaintenanceRequest) (*model.StatusMaintenanceWindow, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) DeleteMaintenance(id int64) error {
	return nil
}

func (s *statusPageServiceImpl) CreateSubscriber(pageID int64, req model.CreateSubscriberRequest) (*model.StatusSubscriber, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) ListSubscribers(pageID int64) ([]model.StatusSubscriber, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) DeleteSubscriber(id int64) error {
	return nil
}

func (s *statusPageServiceImpl) GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error) {
	return nil, nil
}

func (s *statusPageServiceImpl) GetPublicStatusOverview(slug string) (*model.StatusPage, []model.ComponentGroup, []model.StatusIncident, error) {
	return nil, nil, nil, nil
}
