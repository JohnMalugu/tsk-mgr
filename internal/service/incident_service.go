package service

import (
	"errors"

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

func (s *incidentServiceImpl) CreateIncident(reporterID string, req model.CreateIncidentRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) GetIncident(id int64) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) UpdateIncident(id int64, req model.UpdateIncidentRequest) (*model.Incident, error) {
	return nil, nil
}

func (s *incidentServiceImpl) DeleteIncident(id int64) error {
	return nil
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
