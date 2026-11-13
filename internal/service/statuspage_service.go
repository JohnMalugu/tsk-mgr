package service

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrStatusPageNotFound        = model.ErrStatusPageNotFound
	ErrStatusPageNameRequired    = model.ErrStatusPageNameRequired
	ErrStatusPageSlugRequired    = model.ErrStatusPageSlugRequired
	ErrInvalidStatusPageSlug     = model.ErrInvalidStatusPageSlug
	ErrDuplicateStatusPageSlug   = model.ErrDuplicateStatusPageSlug
	ErrComponentNotFound         = model.ErrComponentNotFound
	ErrComponentNameRequired     = model.ErrComponentNameRequired
	ErrInvalidComponentStatus    = model.ErrInvalidComponentStatus
	ErrStatusIncidentNotFound    = model.ErrStatusIncidentNotFound
	ErrStatusIncidentTitleReq    = model.ErrStatusIncidentTitleReq
	ErrInvalidStatusIncidentSt   = model.ErrInvalidStatusIncidentSt
	ErrInvalidIncidentImpact     = model.ErrInvalidIncidentImpact
	ErrStatusMaintenanceNotFound = model.ErrMaintenanceNotFound
	ErrMaintenanceTitleReq       = model.ErrMaintenanceTitleReq
	ErrInvalidMaintenanceTimes   = model.ErrInvalidMaintenanceTimes
	ErrSubscriberNotFound        = model.ErrSubscriberNotFound
	ErrSubscriberTargetRequired  = model.ErrSubscriberTargetRequired
	ErrInvalidSubscriberType     = model.ErrInvalidSubscriberType
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
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("component group name is required")
	}

	group := &model.ComponentGroup{
		PageID:       pageID,
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		DisplayOrder: req.DisplayOrder,
		Collapsed:    req.Collapsed,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.repo.CreateComponentGroup(group)
	if err != nil {
		return nil, err
	}

	return s.repo.GetComponentGroupByID(id)
}

func (s *statusPageServiceImpl) ListComponentGroups(pageID int64) ([]model.ComponentGroup, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	groups, err := s.repo.ListComponentGroups(pageID)
	if err != nil {
		return nil, err
	}

	allComps, err := s.repo.ListComponents(pageID)
	if err == nil {
		for i := range groups {
			for _, comp := range allComps {
				if comp.GroupID != nil && *comp.GroupID == groups[i].ID {
					groups[i].Components = append(groups[i].Components, comp)
				}
			}
			groups[i].ComponentsCount = len(groups[i].Components)
		}
	}

	return groups, nil
}

func (s *statusPageServiceImpl) DeleteComponentGroup(id int64) error {
	err := s.repo.DeleteComponentGroup(id)
	if err != nil {
		if errors.Is(err, repository.ErrComponentGroupNotFound) {
			return errors.New("component group not found")
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) CreateComponent(pageID int64, req model.CreateComponentRequest) (*model.StatusComponent, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrComponentNameRequired
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.StatusOperational
	} else if err := model.ValidateComponentStatus(status); err != nil {
		return nil, err
	}

	if req.GroupID != nil {
		if _, err := s.repo.GetComponentGroupByID(*req.GroupID); err != nil {
			return nil, errors.New("specified component group does not exist")
		}
	}

	comp := &model.StatusComponent{
		PageID:       pageID,
		GroupID:      req.GroupID,
		CIID:         req.CIID,
		Name:         name,
		Description:  strings.TrimSpace(req.Description),
		Status:       status,
		DisplayOrder: req.DisplayOrder,
		IsMonitored:  req.IsMonitored,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	id, err := s.repo.CreateComponent(comp)
	if err != nil {
		return nil, err
	}

	return s.GetComponent(id)
}

func (s *statusPageServiceImpl) GetComponent(id int64) (*model.StatusComponent, error) {
	comp, err := s.repo.GetComponentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrComponentNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}
	return comp, nil
}

func (s *statusPageServiceImpl) ListComponents(pageID int64) ([]model.StatusComponent, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}
	return s.repo.ListComponents(pageID)
}

func (s *statusPageServiceImpl) UpdateComponent(id int64, req model.UpdateComponentRequest) (*model.StatusComponent, error) {
	existing, err := s.repo.GetComponentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrComponentNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrComponentNameRequired
		}
		existing.Name = name
	}
	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := model.ValidateComponentStatus(st); err != nil {
			return nil, err
		}
		existing.Status = st
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}
	if req.GroupID != nil {
		existing.GroupID = req.GroupID
	}
	if req.CIID != nil {
		existing.CIID = req.CIID
	}
	if req.DisplayOrder != nil {
		existing.DisplayOrder = *req.DisplayOrder
	}
	if req.IsMonitored != nil {
		existing.IsMonitored = *req.IsMonitored
	}

	if err := s.repo.UpdateComponent(existing); err != nil {
		if errors.Is(err, repository.ErrComponentNotFound) {
			return nil, ErrComponentNotFound
		}
		return nil, err
	}

	return s.GetComponent(id)
}

func (s *statusPageServiceImpl) DeleteComponent(id int64) error {
	err := s.repo.DeleteComponent(id)
	if err != nil {
		if errors.Is(err, repository.ErrComponentNotFound) {
			return ErrComponentNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) SetComponentStatus(id int64, status string) error {
	st := strings.ToLower(strings.TrimSpace(status))
	if err := model.ValidateComponentStatus(st); err != nil {
		return err
	}
	err := s.repo.UpdateComponentStatus(id, st)
	if err != nil {
		if errors.Is(err, repository.ErrComponentNotFound) {
			return ErrComponentNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) CreateIncident(creatorID string, pageID int64, req model.CreateStatusIncidentRequest) (*model.StatusIncident, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrStatusIncidentTitleReq
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.StatusIncidentInvestigating
	} else if err := model.ValidateStatusIncidentStatus(status); err != nil {
		return nil, err
	}

	impact := strings.ToLower(strings.TrimSpace(req.Impact))
	if impact == "" {
		impact = model.ImpactMinor
	} else if err := model.ValidateIncidentImpact(impact); err != nil {
		return nil, err
	}

	affectedBytes, err := json.Marshal(req.AffectedComponentIDs)
	if err != nil {
		affectedBytes = []byte("[]")
	}

	inc := &model.StatusIncident{
		PageID:                 pageID,
		IncidentID:             req.IncidentID,
		Title:                  title,
		Status:                 status,
		Impact:                 impact,
		StartedAt:              time.Now(),
		AffectedComponentsJSON: string(affectedBytes),
		CreatedByID:            creatorID,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	if status == model.StatusIncidentResolved {
		now := time.Now()
		inc.ResolvedAt = &now
	}

	id, err := s.repo.CreateIncident(inc)
	if err != nil {
		return nil, err
	}

	// Post initial timeline message if provided
	msg := strings.TrimSpace(req.Message)
	if msg != "" {
		_, _ = s.repo.AddIncidentUpdate(&model.StatusIncidentUpdate{
			StatusIncidentID: id,
			Status:           status,
			Message:          msg,
			PostedByID:       creatorID,
			CreatedAt:        time.Now(),
		})
	}

	// Update affected component statuses if map provided
	for compID, compSt := range req.ComponentStatusMap {
		_ = s.SetComponentStatus(compID, compSt)
	}

	return s.GetIncident(id)
}

func (s *statusPageServiceImpl) GetIncident(id int64) (*model.StatusIncident, error) {
	inc, err := s.repo.GetIncidentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusIncidentNotFound) {
			return nil, ErrStatusIncidentNotFound
		}
		return nil, err
	}

	if inc.AffectedComponentsJSON != "" {
		var ids []int64
		if err := json.Unmarshal([]byte(inc.AffectedComponentsJSON), &ids); err == nil {
			inc.AffectedComponentIDs = ids
		}
	}

	return inc, nil
}

func (s *statusPageServiceImpl) ListIncidents(pageID int64, limit, offset int) ([]model.StatusIncident, int, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, 0, ErrStatusPageNotFound
		}
		return nil, 0, err
	}

	incidents, total, err := s.repo.ListIncidents(pageID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range incidents {
		if incidents[i].AffectedComponentsJSON != "" {
			var ids []int64
			if err := json.Unmarshal([]byte(incidents[i].AffectedComponentsJSON), &ids); err == nil {
				incidents[i].AffectedComponentIDs = ids
			}
		}
	}

	return incidents, total, nil
}

func (s *statusPageServiceImpl) UpdateIncident(id int64, req model.UpdateStatusIncidentRequest) (*model.StatusIncident, error) {
	existing, err := s.repo.GetIncidentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusIncidentNotFound) {
			return nil, ErrStatusIncidentNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrStatusIncidentTitleReq
		}
		existing.Title = t
	}
	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := model.ValidateStatusIncidentStatus(st); err != nil {
			return nil, err
		}
		existing.Status = st
		if st == model.StatusIncidentResolved && existing.ResolvedAt == nil {
			now := time.Now()
			existing.ResolvedAt = &now
		}
	}
	if req.Impact != nil {
		imp := strings.ToLower(strings.TrimSpace(*req.Impact))
		if err := model.ValidateIncidentImpact(imp); err != nil {
			return nil, err
		}
		existing.Impact = imp
	}
	if req.Resolved != nil && *req.Resolved {
		existing.Status = model.StatusIncidentResolved
		if existing.ResolvedAt == nil {
			now := time.Now()
			existing.ResolvedAt = &now
		}
	}

	if err := s.repo.UpdateIncident(existing); err != nil {
		if errors.Is(err, repository.ErrStatusIncidentNotFound) {
			return nil, ErrStatusIncidentNotFound
		}
		return nil, err
	}

	return s.GetIncident(id)
}

func (s *statusPageServiceImpl) AddIncidentUpdate(postedByID string, incidentID int64, req model.AddStatusIncidentUpdateRequest) (*model.StatusIncidentUpdate, error) {
	inc, err := s.repo.GetIncidentByID(incidentID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusIncidentNotFound) {
			return nil, ErrStatusIncidentNotFound
		}
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = inc.Status
	} else if err := model.ValidateStatusIncidentStatus(status); err != nil {
		return nil, err
	}

	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		return nil, errors.New("incident update message is required")
	}

	update := &model.StatusIncidentUpdate{
		StatusIncidentID: incidentID,
		Status:           status,
		Message:          msg,
		PostedByID:       postedByID,
		CreatedAt:        time.Now(),
	}

	id, err := s.repo.AddIncidentUpdate(update)
	if err != nil {
		return nil, err
	}

	// Update affected component statuses if map provided
	for compID, compSt := range req.ComponentStatusMap {
		_ = s.SetComponentStatus(compID, compSt)
	}

	update.ID = id
	return update, nil
}

func (s *statusPageServiceImpl) DeleteIncident(id int64) error {
	err := s.repo.DeleteIncident(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusIncidentNotFound) {
			return ErrStatusIncidentNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) CreateMaintenance(creatorID string, pageID int64, req model.CreateMaintenanceRequest) (*model.StatusMaintenanceWindow, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrMaintenanceTitleReq
	}

	if req.ScheduledStartAt.IsZero() || req.ScheduledEndAt.IsZero() || !req.ScheduledStartAt.Before(req.ScheduledEndAt) {
		return nil, ErrInvalidMaintenanceTimes
	}

	affectedBytes, err := json.Marshal(req.AffectedComponentIDs)
	if err != nil {
		affectedBytes = []byte("[]")
	}

	maint := &model.StatusMaintenanceWindow{
		PageID:                 pageID,
		Title:                  title,
		Description:            strings.TrimSpace(req.Description),
		Status:                 model.MaintenanceStatusScheduled,
		ScheduledStartAt:       req.ScheduledStartAt,
		ScheduledEndAt:         req.ScheduledEndAt,
		AffectedComponentsJSON: string(affectedBytes),
		CreatedByID:            creatorID,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	id, err := s.repo.CreateMaintenance(maint)
	if err != nil {
		return nil, err
	}

	return s.GetMaintenance(id)
}

func (s *statusPageServiceImpl) GetMaintenance(id int64) (*model.StatusMaintenanceWindow, error) {
	m, err := s.repo.GetMaintenanceByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusMaintenanceNotFound) {
			return nil, ErrStatusMaintenanceNotFound
		}
		return nil, err
	}

	if m.AffectedComponentsJSON != "" {
		var ids []int64
		if err := json.Unmarshal([]byte(m.AffectedComponentsJSON), &ids); err == nil {
			m.AffectedComponentIDs = ids
		}
	}

	return m, nil
}

func (s *statusPageServiceImpl) ListMaintenances(pageID int64, limit, offset int) ([]model.StatusMaintenanceWindow, int, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, 0, ErrStatusPageNotFound
		}
		return nil, 0, err
	}

	maintenances, total, err := s.repo.ListMaintenances(pageID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	for i := range maintenances {
		if maintenances[i].AffectedComponentsJSON != "" {
			var ids []int64
			if err := json.Unmarshal([]byte(maintenances[i].AffectedComponentsJSON), &ids); err == nil {
				maintenances[i].AffectedComponentIDs = ids
			}
		}
	}

	return maintenances, total, nil
}

func (s *statusPageServiceImpl) UpdateMaintenance(id int64, req model.UpdateMaintenanceRequest) (*model.StatusMaintenanceWindow, error) {
	existing, err := s.repo.GetMaintenanceByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusMaintenanceNotFound) {
			return nil, ErrStatusMaintenanceNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrMaintenanceTitleReq
		}
		existing.Title = t
	}
	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}

	newStart := existing.ScheduledStartAt
	newEnd := existing.ScheduledEndAt
	if req.ScheduledStartAt != nil {
		newStart = *req.ScheduledStartAt
	}
	if req.ScheduledEndAt != nil {
		newEnd = *req.ScheduledEndAt
	}
	if !newStart.Before(newEnd) {
		return nil, ErrInvalidMaintenanceTimes
	}
	existing.ScheduledStartAt = newStart
	existing.ScheduledEndAt = newEnd

	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		switch st {
		case model.MaintenanceStatusScheduled, model.MaintenanceStatusInProgress,
			model.MaintenanceStatusCompleted, model.MaintenanceStatusCancelled:
			existing.Status = st
		default:
			return nil, errors.New("invalid maintenance status")
		}

		now := time.Now()
		if st == model.MaintenanceStatusInProgress && existing.ActualStartAt == nil {
			existing.ActualStartAt = &now
			// Mark affected components as under maintenance
			if existing.AffectedComponentsJSON != "" {
				var ids []int64
				if err := json.Unmarshal([]byte(existing.AffectedComponentsJSON), &ids); err == nil {
					for _, compID := range ids {
						_ = s.SetComponentStatus(compID, model.StatusUnderMaintenance)
					}
				}
			}
		} else if (st == model.MaintenanceStatusCompleted || st == model.MaintenanceStatusCancelled) && existing.ActualEndAt == nil {
			existing.ActualEndAt = &now
			// Return affected components to operational
			if existing.AffectedComponentsJSON != "" {
				var ids []int64
				if err := json.Unmarshal([]byte(existing.AffectedComponentsJSON), &ids); err == nil {
					for _, compID := range ids {
						_ = s.SetComponentStatus(compID, model.StatusOperational)
					}
				}
			}
		}
	}

	if err := s.repo.UpdateMaintenance(existing); err != nil {
		if errors.Is(err, repository.ErrStatusMaintenanceNotFound) {
			return nil, ErrStatusMaintenanceNotFound
		}
		return nil, err
	}

	return s.GetMaintenance(id)
}

func (s *statusPageServiceImpl) DeleteMaintenance(id int64) error {
	err := s.repo.DeleteMaintenance(id)
	if err != nil {
		if errors.Is(err, repository.ErrStatusMaintenanceNotFound) {
			return ErrStatusMaintenanceNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) CreateSubscriber(pageID int64, req model.CreateSubscriberRequest) (*model.StatusSubscriber, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	target := strings.TrimSpace(req.TargetEndpoint)
	if target == "" {
		return nil, ErrSubscriberTargetRequired
	}

	subType := strings.ToLower(strings.TrimSpace(req.SubscriberType))
	if subType == "" {
		subType = model.SubscriberTypeEmail
	} else if err := model.ValidateSubscriberType(subType); err != nil {
		return nil, err
	}

	componentsBytes, err := json.Marshal(req.SubscribedComponentIDs)
	if err != nil {
		componentsBytes = []byte("[]")
	}

	sub := &model.StatusSubscriber{
		PageID:                   pageID,
		SubscriberType:           subType,
		TargetEndpoint:           target,
		IsVerified:               true,
		SubscribedComponentsJSON: string(componentsBytes),
		CreatedAt:                time.Now(),
		UpdatedAt:                time.Now(),
	}

	id, err := s.repo.CreateSubscriber(sub)
	if err != nil {
		return nil, err
	}

	created, err := s.repo.GetSubscriberByID(id)
	if err != nil {
		return nil, err
	}

	if created.SubscribedComponentsJSON != "" {
		var ids []int64
		if err := json.Unmarshal([]byte(created.SubscribedComponentsJSON), &ids); err == nil {
			created.SubscribedComponentIDs = ids
		}
	}

	return created, nil
}

func (s *statusPageServiceImpl) ListSubscribers(pageID int64) ([]model.StatusSubscriber, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	subs, err := s.repo.ListSubscribers(pageID)
	if err != nil {
		return nil, err
	}

	for i := range subs {
		if subs[i].SubscribedComponentsJSON != "" {
			var ids []int64
			if err := json.Unmarshal([]byte(subs[i].SubscribedComponentsJSON), &ids); err == nil {
				subs[i].SubscribedComponentIDs = ids
			}
		}
	}

	return subs, nil
}

func (s *statusPageServiceImpl) DeleteSubscriber(id int64) error {
	err := s.repo.DeleteSubscriber(id)
	if err != nil {
		if errors.Is(err, repository.ErrSubscriberNotFound) {
			return ErrSubscriberNotFound
		}
		return err
	}
	return nil
}

func (s *statusPageServiceImpl) GetPageMetrics(pageID int64) (*model.StatusPageMetrics, error) {
	_, err := s.repo.GetPageByID(pageID)
	if err != nil {
		if errors.Is(err, repository.ErrStatusPageNotFound) {
			return nil, ErrStatusPageNotFound
		}
		return nil, err
	}

	return s.repo.GetPageMetrics(pageID)
}

func (s *statusPageServiceImpl) GetPublicStatusOverview(slug string) (*model.StatusPage, []model.ComponentGroup, []model.StatusIncident, error) {
	page, err := s.GetPageBySlug(slug)
	if err != nil {
		return nil, nil, nil, err
	}

	if !page.IsActive {
		return nil, nil, nil, ErrStatusPageNotFound
	}

	groups, err := s.ListComponentGroups(page.ID)
	if err != nil {
		return nil, nil, nil, err
	}
	if groups == nil {
		groups = []model.ComponentGroup{}
	}

	incidents, _, err := s.ListIncidents(page.ID, 10, 0)
	if err != nil {
		return nil, nil, nil, err
	}
	if incidents == nil {
		incidents = []model.StatusIncident{}
	}

	return page, groups, incidents, nil
}
