package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrIncidentNotFound          = errors.New("incident not found")
	ErrIncidentTitleRequired     = errors.New("incident title is required")
	ErrInvalidIncidentSeverity   = errors.New("invalid incident severity")
	ErrInvalidIncidentStatus     = errors.New("invalid incident status")
	ErrInvalidPostmortemStatus   = errors.New("invalid postmortem status")
	ErrInvalidTimelineEventType  = errors.New("invalid timeline event type")
	ErrTimelineMessageRequired   = errors.New("timeline message is required")
	ErrInvalidIncidentTaskType   = errors.New("invalid incident task type")
)

// IncidentService manages operational incident response, timelines, remediation tasks, and metrics
type IncidentService interface {
	CreateIncident(reporterID string, req model.CreateIncidentRequest) (*model.Incident, error)
	GetIncident(id int64) (*model.Incident, error)
	GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error)
	ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error)
	UpdateIncident(id int64, req model.UpdateIncidentRequest) (*model.Incident, error)
	DeleteIncident(id int64) error

	ChangeStatus(id int64, authorID string, req model.ChangeIncidentStatusRequest) (*model.Incident, error)
	ChangeSeverity(id int64, authorID string, req model.ChangeIncidentSeverityRequest) (*model.Incident, error)
	AssignRoles(id int64, authorID string, req model.AssignIncidentRolesRequest) (*model.Incident, error)

	AddTimelineEvent(incidentID int64, authorID string, req model.CreateIncidentTimelineRequest) (*model.IncidentTimeline, error)
	ListTimeline(incidentID int64) ([]model.IncidentTimeline, error)

	LinkTask(incidentID int64, createdBy string, req model.LinkIncidentTaskRequest) (*model.IncidentTaskLink, error)
	UnlinkTask(incidentID int64, taskID int) error
	ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error)

	UpdatePostmortem(id int64, req model.UpdatePostmortemRequest) (*model.Incident, error)
	GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error)
}

type incidentServiceImpl struct {
	repo        repository.IncidentRepository
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository
	docRepo     repository.DocumentRepository
}

// NewIncidentService creates a new IncidentService instance
func NewIncidentService(
	repo repository.IncidentRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	docRepo repository.DocumentRepository,
) IncidentService {
	return &incidentServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
		docRepo:     docRepo,
	}
}

func isValidIncidentSeverity(sev string) bool {
	switch sev {
	case model.IncidentSev0, model.IncidentSev1, model.IncidentSev2, model.IncidentSev3:
		return true
	default:
		return false
	}
}

func isValidIncidentStatus(status string) bool {
	switch status {
	case model.IncidentStatusReported, model.IncidentStatusInvestigating,
		model.IncidentStatusIdentified, model.IncidentStatusMonitoring,
		model.IncidentStatusResolved, model.IncidentStatusClosed:
		return true
	default:
		return false
	}
}

func (s *incidentServiceImpl) CreateIncident(reporterID string, req model.CreateIncidentRequest) (*model.Incident, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrIncidentTitleRequired
	}

	severity := req.Severity
	if severity == "" {
		severity = model.IncidentSev2
	}
	if !isValidIncidentSeverity(severity) {
		return nil, ErrInvalidIncidentSeverity
	}

	if req.ProjectID != nil && s.projectRepo != nil {
		p, err := s.projectRepo.GetByID(*req.ProjectID)
		if err != nil || p == nil {
			return nil, ErrProjectNotFound
		}
	}

	if reporterID == "" {
		reporterID = "user_1"
	}

	startedAt := time.Now().UTC()
	if req.StartedAt != nil && !req.StartedAt.IsZero() {
		startedAt = req.StartedAt.UTC()
	}

	pmStatus := model.PostmortemNotRequired
	if severity == model.IncidentSev0 || severity == model.IncidentSev1 {
		pmStatus = model.PostmortemPending
	}

	services := req.ImpactedServices
	if services == nil {
		services = []string{}
	}

	inc := &model.Incident{
		WorkspaceID:      req.WorkspaceID,
		ProjectID:        req.ProjectID,
		Title:            title,
		Summary:          strings.TrimSpace(req.Summary),
		Severity:         severity,
		Status:           model.IncidentStatusReported,
		CommanderID:      req.CommanderID,
		ScribeID:         req.ScribeID,
		ReportedByID:     reporterID,
		PostmortemStatus: pmStatus,
		ImpactedServices: services,
		StartedAt:        startedAt,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
	}

	id, err := s.repo.CreateIncident(inc)
	if err != nil {
		return nil, err
	}

	// Record initial timeline entry
	_, _ = s.repo.CreateTimelineEvent(&model.IncidentTimeline{
		IncidentID: id,
		AuthorID:   reporterID,
		EventType:  model.TimelineEventStatusChange,
		Message:    fmt.Sprintf("Incident declared with severity %s", severity),
		CreatedAt:  time.Now().UTC(),
	})

	return s.repo.GetIncidentByID(id)
}

func (s *incidentServiceImpl) GetIncident(id int64) (*model.Incident, error) {
	inc, err := s.repo.GetIncidentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrIncidentNotFound) {
			return nil, ErrIncidentNotFound
		}
		return nil, err
	}
	return inc, nil
}

func (s *incidentServiceImpl) GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	inc, err := s.repo.GetIncidentByNumber(workspaceID, incidentNumber)
	if err != nil {
		if errors.Is(err, repository.ErrIncidentNotFound) {
			return nil, ErrIncidentNotFound
		}
		return nil, err
	}
	return inc, nil
}

func (s *incidentServiceImpl) ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListIncidents(filter)
}

func (s *incidentServiceImpl) UpdateIncident(id int64, req model.UpdateIncidentRequest) (*model.Incident, error) {
	inc, err := s.repo.GetIncidentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrIncidentNotFound) {
			return nil, ErrIncidentNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, ErrIncidentTitleRequired
		}
		inc.Title = title
	}
	if req.Summary != nil {
		inc.Summary = strings.TrimSpace(*req.Summary)
	}
	if req.ProjectID != nil {
		if *req.ProjectID > 0 {
			if s.projectRepo != nil {
				p, err := s.projectRepo.GetByID(*req.ProjectID)
				if err != nil || p == nil {
					return nil, ErrProjectNotFound
				}
			}
			inc.ProjectID = req.ProjectID
		} else {
			inc.ProjectID = nil
		}
	}
	if req.CommanderID != nil {
		inc.CommanderID = req.CommanderID
	}
	if req.ScribeID != nil {
		inc.ScribeID = req.ScribeID
	}
	if req.ImpactedServices != nil {
		inc.ImpactedServices = *req.ImpactedServices
	}
	if req.RootCause != nil {
		inc.RootCause = strings.TrimSpace(*req.RootCause)
	}

	err = s.repo.UpdateIncident(inc)
	if err != nil {
		return nil, err
	}
	return s.repo.GetIncidentByID(id)
}

func (s *incidentServiceImpl) DeleteIncident(id int64) error {
	_, err := s.repo.GetIncidentByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrIncidentNotFound) {
			return ErrIncidentNotFound
		}
		return err
	}
	return s.repo.DeleteIncident(id)
}

func (s *incidentServiceImpl) ChangeStatus(id int64, authorID string, req model.ChangeIncidentStatusRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) ChangeSeverity(id int64, authorID string, req model.ChangeIncidentSeverityRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) AssignRoles(id int64, authorID string, req model.AssignIncidentRolesRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) AddTimelineEvent(incidentID int64, authorID string, req model.CreateIncidentTimelineRequest) (*model.IncidentTimeline, error) {
	return nil, nil
}

func (s *incidentServiceImpl) ListTimeline(incidentID int64) ([]model.IncidentTimeline, error) {
	return nil, nil
}

func (s *incidentServiceImpl) LinkTask(incidentID int64, createdBy string, req model.LinkIncidentTaskRequest) (*model.IncidentTaskLink, error) {
	return nil, nil
}

func (s *incidentServiceImpl) UnlinkTask(incidentID int64, taskID int) error {
	return nil
}

func (s *incidentServiceImpl) ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error) {
	return nil, nil
}

func (s *incidentServiceImpl) UpdatePostmortem(id int64, req model.UpdatePostmortemRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error) {
	return nil, nil
}
