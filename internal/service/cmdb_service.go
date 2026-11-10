package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrCINotFound          = repository.ErrCINotFound
	ErrRelationshipNotFound= repository.ErrRelationshipNotFound
	ErrLinkNotFound        = repository.ErrLinkNotFound
	ErrMaintenanceNotFound = repository.ErrMaintenanceNotFound
	ErrCINameRequired      = errors.New("configuration item name is required")
	ErrSelfRelationship    = errors.New("cannot create relationship to the same configuration item")
	ErrRelationshipExists  = errors.New("relationship already exists between these configuration items")
	ErrLinkAlreadyExists   = errors.New("entity is already linked to this configuration item")
	ErrMaintTitleRequired  = errors.New("maintenance schedule title is required")
	ErrInvalidMaintDates   = errors.New("maintenance window end time must be after start time")
)

// CMDBService manages ITIL Configuration Items, dependencies, blast-radius analysis, and maintenance schedules
type CMDBService interface {
	CreateCI(creatorID string, req model.CreateCIRequest) (*model.ConfigurationItem, error)
	GetCI(id int64) (*model.ConfigurationItem, error)
	GetCIByCode(workspaceID int64, code string) (*model.ConfigurationItem, error)
	ListCIs(filter model.CIFilter) ([]*model.ConfigurationItem, error)
	UpdateCI(id int64, updaterID string, req model.UpdateCIRequest) (*model.ConfigurationItem, error)
	DeleteCI(id int64, deleterID string) error
	UpdateCIStatus(id int64, updaterID string, req model.UpdateCIStatusRequest) (*model.ConfigurationItem, error)

	CreateRelationship(ciID int64, req model.CreateCIRelationshipRequest) (*model.CIRelationship, error)
	ListRelationships(ciID int64) ([]model.CIRelationship, error)
	DeleteRelationship(id int64) error
	GetDependencyGraph(ciID int64, maxDepth int) (*model.CIDependencyGraph, error)
	AnalyzeBlastRadius(ciID int64) (*model.CIImpactAnalysis, error)

	LinkEntity(ciID int64, req model.CreateCIEntityLinkRequest) (*model.CIEntityLink, error)
	ListEntityLinks(ciID int64) ([]model.CIEntityLink, error)
	ListCIsForEntity(entityType string, entityID int64) ([]model.CIEntityLink, error)
	UnlinkEntity(linkID int64) error

	CreateMaintenance(ciID int64, creatorID string, req model.CreateCIMaintenanceRequest) (*model.CIMaintenanceSchedule, error)
	ListMaintenances(ciID int64) ([]model.CIMaintenanceSchedule, error)
	UpdateMaintenanceStatus(id int64, req model.UpdateCIMaintenanceStatusRequest) (*model.CIMaintenanceSchedule, error)
	DeleteMaintenance(id int64) error

	GetCMDBMetrics(workspaceID int64) (*model.CMDBMetrics, error)
}

type cmdbServiceImpl struct {
	repo          repository.CMDBRepository
	workspaceRepo repository.WorkspaceRepository
	projectRepo   repository.ProjectRepository
	taskRepo      repository.TaskRepository
	auditService  AuditService
}

// NewCMDBService creates a new instance of CMDBService
func NewCMDBService(
	repo repository.CMDBRepository,
	workspaceRepo repository.WorkspaceRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	auditService AuditService,
) CMDBService {
	return &cmdbServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		taskRepo:      taskRepo,
		auditService:  auditService,
	}
}

func (s *cmdbServiceImpl) logAudit(actorID string, workspaceID int64, action, entityID, description string) {
	if s.auditService != nil {
		_, _ = s.auditService.RecordEvent(actorID, model.CreateAuditLogRequest{
			WorkspaceID: workspaceID,
			Action:      action,
			Category:    "cmdb",
			Severity:    "info",
			EntityType:  "configuration_item",
			EntityID:    entityID,
			Description: description,
		})
	}
}

func (s *cmdbServiceImpl) CreateCI(creatorID string, req model.CreateCIRequest) (*model.ConfigurationItem, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrCINameRequired
	}

	ciType := strings.ToLower(strings.TrimSpace(req.CIType))
	if ciType == "" {
		ciType = model.CITypeService
	} else if err := model.ValidateCIType(ciType); err != nil {
		return nil, err
	}

	criticality := strings.ToLower(strings.TrimSpace(req.Criticality))
	if criticality == "" {
		criticality = model.CICriticalityTier2
	} else if err := model.ValidateCICriticality(criticality); err != nil {
		return nil, err
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))
	if status == "" {
		status = model.CIStatusActive
	} else if err := model.ValidateCIStatus(status); err != nil {
		return nil, err
	}

	env := strings.ToLower(strings.TrimSpace(req.Environment))
	if env == "" {
		env = model.CIEnvProduction
	} else if err := model.ValidateCIEnvironment(env); err != nil {
		return nil, err
	}

	owner := strings.TrimSpace(req.OwnerUserID)
	if owner == "" {
		owner = creatorID
	}

	ci := &model.ConfigurationItem{
		WorkspaceID:  req.WorkspaceID,
		ProjectID:    req.ProjectID,
		Name:         name,
		CIType:       ciType,
		Status:       status,
		Criticality:  criticality,
		Environment:  env,
		OwnerUserID:  owner,
		IPAddress:    strings.TrimSpace(req.IPAddress),
		Hostname:     strings.TrimSpace(req.Hostname),
		Version:      strings.TrimSpace(req.Version),
		CostPerMonth: req.CostPerMonth,
		Attributes:   strings.TrimSpace(req.Attributes),
	}

	id, err := s.repo.CreateCI(ci)
	if err != nil {
		return nil, err
	}

	s.logAudit(creatorID, req.WorkspaceID, "cmdb.ci_created", fmt.Sprintf("%d", id), fmt.Sprintf("Created configuration item %s (%s)", ci.CICode, name))

	return s.repo.GetCIByID(id)
}

func (s *cmdbServiceImpl) GetCI(id int64) (*model.ConfigurationItem, error) {
	ci, err := s.repo.GetCIByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return ci, nil
}

func (s *cmdbServiceImpl) GetCIByCode(workspaceID int64, code string) (*model.ConfigurationItem, error) {
	ci, err := s.repo.GetCIByCode(workspaceID, code)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return ci, nil
}

func (s *cmdbServiceImpl) ListCIs(filter model.CIFilter) ([]*model.ConfigurationItem, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListCIs(filter)
}

func (s *cmdbServiceImpl) UpdateCI(id int64, updaterID string, req model.UpdateCIRequest) (*model.ConfigurationItem, error) {
	ci, err := s.repo.GetCIByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name := strings.TrimSpace(*req.Name)
		if name == "" {
			return nil, ErrCINameRequired
		}
		ci.Name = name
	}
	if req.ProjectID != nil {
		ci.ProjectID = req.ProjectID
	}
	if req.CIType != nil {
		t := strings.ToLower(strings.TrimSpace(*req.CIType))
		if err := model.ValidateCIType(t); err != nil {
			return nil, err
		}
		ci.CIType = t
	}
	if req.Status != nil {
		st := strings.ToLower(strings.TrimSpace(*req.Status))
		if err := model.ValidateCIStatus(st); err != nil {
			return nil, err
		}
		ci.Status = st
	}
	if req.Criticality != nil {
		cr := strings.ToLower(strings.TrimSpace(*req.Criticality))
		if err := model.ValidateCICriticality(cr); err != nil {
			return nil, err
		}
		ci.Criticality = cr
	}
	if req.Environment != nil {
		env := strings.ToLower(strings.TrimSpace(*req.Environment))
		if err := model.ValidateCIEnvironment(env); err != nil {
			return nil, err
		}
		ci.Environment = env
	}
	if req.OwnerUserID != nil {
		ci.OwnerUserID = strings.TrimSpace(*req.OwnerUserID)
	}
	if req.IPAddress != nil {
		ci.IPAddress = strings.TrimSpace(*req.IPAddress)
	}
	if req.Hostname != nil {
		ci.Hostname = strings.TrimSpace(*req.Hostname)
	}
	if req.Version != nil {
		ci.Version = strings.TrimSpace(*req.Version)
	}
	if req.CostPerMonth != nil {
		ci.CostPerMonth = *req.CostPerMonth
	}
	if req.Attributes != nil {
		ci.Attributes = strings.TrimSpace(*req.Attributes)
	}

	if err := s.repo.UpdateCI(ci); err != nil {
		return nil, err
	}

	s.logAudit(updaterID, ci.WorkspaceID, "cmdb.ci_updated", fmt.Sprintf("%d", id), fmt.Sprintf("Updated configuration item %s (%s)", ci.CICode, ci.Name))

	return s.repo.GetCIByID(id)
}

func (s *cmdbServiceImpl) DeleteCI(id int64, deleterID string) error {
	ci, err := s.repo.GetCIByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return ErrCINotFound
		}
		return err
	}

	if err := s.repo.DeleteCI(id); err != nil {
		return err
	}

	s.logAudit(deleterID, ci.WorkspaceID, "cmdb.ci_deleted", fmt.Sprintf("%d", id), fmt.Sprintf("Deleted configuration item %s (%s)", ci.CICode, ci.Name))
	return nil
}

func (s *cmdbServiceImpl) UpdateCIStatus(id int64, updaterID string, req model.UpdateCIStatusRequest) (*model.ConfigurationItem, error) {
	st := strings.ToLower(strings.TrimSpace(req.Status))
	if err := model.ValidateCIStatus(st); err != nil {
		return nil, err
	}

	ci, err := s.repo.GetCIByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	ci.Status = st
	if err := s.repo.UpdateCI(ci); err != nil {
		return nil, err
	}

	s.logAudit(updaterID, ci.WorkspaceID, "cmdb.ci_status_changed", fmt.Sprintf("%d", id), fmt.Sprintf("Changed status of %s to %s (reason: %s)", ci.CICode, st, req.Reason))

	return s.repo.GetCIByID(id)
}

func (s *cmdbServiceImpl) CreateRelationship(ciID int64, req model.CreateCIRelationshipRequest) (*model.CIRelationship, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) ListRelationships(ciID int64) ([]model.CIRelationship, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) DeleteRelationship(id int64) error {
	return nil
}

func (s *cmdbServiceImpl) GetDependencyGraph(ciID int64, maxDepth int) (*model.CIDependencyGraph, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) AnalyzeBlastRadius(ciID int64) (*model.CIImpactAnalysis, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) LinkEntity(ciID int64, req model.CreateCIEntityLinkRequest) (*model.CIEntityLink, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) ListEntityLinks(ciID int64) ([]model.CIEntityLink, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) ListCIsForEntity(entityType string, entityID int64) ([]model.CIEntityLink, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) UnlinkEntity(linkID int64) error {
	return nil
}

func (s *cmdbServiceImpl) CreateMaintenance(ciID int64, creatorID string, req model.CreateCIMaintenanceRequest) (*model.CIMaintenanceSchedule, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) ListMaintenances(ciID int64) ([]model.CIMaintenanceSchedule, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) UpdateMaintenanceStatus(id int64, req model.UpdateCIMaintenanceStatusRequest) (*model.CIMaintenanceSchedule, error) {
	return nil, nil
}

func (s *cmdbServiceImpl) DeleteMaintenance(id int64) error {
	return nil
}

func (s *cmdbServiceImpl) GetCMDBMetrics(workspaceID int64) (*model.CMDBMetrics, error) {
	return nil, nil
}
