package repository

import (
	"database/sql"
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrIncidentNotFound = errors.New("incident not found")
)

// IncidentRepository manages database operations for incidents, timelines, and remediation tasks
type IncidentRepository interface {
	CreateIncident(incident *model.Incident) (int64, error)
	GetIncidentByID(id int64) (*model.Incident, error)
	GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error)
	ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error)
	UpdateIncident(incident *model.Incident) error
	DeleteIncident(id int64) error
	GetNextIncidentNumber(workspaceID int64) (int, error)

	CreateTimelineEvent(event *model.IncidentTimeline) (int64, error)
	ListTimelineEvents(incidentID int64) ([]model.IncidentTimeline, error)

	LinkTask(link *model.IncidentTaskLink) (int64, error)
	UnlinkTask(incidentID int64, taskID int) error
	ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error)

	GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error)
}

type incidentRepositoryImpl struct {
	db *sql.DB
}

// NewIncidentRepository creates a new instance of IncidentRepository
func NewIncidentRepository(db *sql.DB) IncidentRepository {
	return &incidentRepositoryImpl{db: db}
}

func (r *incidentRepositoryImpl) CreateIncident(incident *model.Incident) (int64, error) {
	return 0, nil
}

func (r *incidentRepositoryImpl) GetIncidentByID(id int64) (*model.Incident, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) GetIncidentByNumber(workspaceID int64, incidentNumber int) (*model.Incident, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) ListIncidents(filter model.IncidentFilter) ([]*model.Incident, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) UpdateIncident(incident *model.Incident) error {
	return nil
}

func (r *incidentRepositoryImpl) DeleteIncident(id int64) error {
	return nil
}

func (r *incidentRepositoryImpl) GetNextIncidentNumber(workspaceID int64) (int, error) {
	return 1, nil
}

func (r *incidentRepositoryImpl) CreateTimelineEvent(event *model.IncidentTimeline) (int64, error) {
	return 0, nil
}

func (r *incidentRepositoryImpl) ListTimelineEvents(incidentID int64) ([]model.IncidentTimeline, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) LinkTask(link *model.IncidentTaskLink) (int64, error) {
	return 0, nil
}

func (r *incidentRepositoryImpl) UnlinkTask(incidentID int64, taskID int) error {
	return nil
}

func (r *incidentRepositoryImpl) ListIncidentTasks(incidentID int64) ([]model.IncidentTaskLink, error) {
	return nil, nil
}

func (r *incidentRepositoryImpl) GetIncidentMetrics(workspaceID int64) (*model.IncidentMetrics, error) {
	return nil, nil
}
