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
	ErrCINotFound           = repository.ErrCINotFound
	ErrRelationshipNotFound = repository.ErrRelationshipNotFound
	ErrLinkNotFound         = repository.ErrLinkNotFound
	ErrMaintenanceNotFound  = repository.ErrMaintenanceNotFound
	ErrCINameRequired       = errors.New("configuration item name is required")
	ErrInvalidCIType        = model.ErrInvalidCIType
	ErrInvalidCICriticality = model.ErrInvalidCICriticality
	ErrInvalidCIStatus      = model.ErrInvalidCIStatus
	ErrInvalidCIEnvironment = model.ErrInvalidCIEnvironment
	ErrSelfRelationship     = errors.New("cannot create relationship to the same configuration item")
	ErrRelationshipExists   = errors.New("relationship already exists between these configuration items")
	ErrLinkAlreadyExists    = errors.New("entity is already linked to this configuration item")
	ErrMaintTitleRequired   = errors.New("maintenance schedule title is required")
	ErrInvalidMaintDates    = errors.New("maintenance window end time must be after start time")
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
	if ciID == req.TargetCIID {
		return nil, ErrSelfRelationship
	}

	src, err := s.repo.GetCIByID(ciID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	tgt, err := s.repo.GetCIByID(req.TargetCIID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	relType := strings.ToLower(strings.TrimSpace(req.RelationType))
	if relType == "" {
		relType = model.CIRelationDependsOn
	} else if err := model.ValidateCIRelationType(relType); err != nil {
		return nil, err
	}

	existing, err := s.repo.ListRelationships(ciID)
	if err == nil {
		for _, r := range existing {
			if r.SourceCIID == ciID && r.TargetCIID == req.TargetCIID && r.RelationType == relType {
				return nil, ErrRelationshipExists
			}
		}
	}

	rel := &model.CIRelationship{
		SourceCIID:   ciID,
		TargetCIID:   req.TargetCIID,
		RelationType: relType,
		Description:  strings.TrimSpace(req.Description),
	}

	id, err := s.repo.CreateRelationship(rel)
	if err != nil {
		return nil, err
	}

	s.logAudit(src.OwnerUserID, src.WorkspaceID, "cmdb.relationship_created", fmt.Sprintf("%d", id), fmt.Sprintf("Linked %s (%s) %s -> %s (%s)", src.CICode, src.Name, relType, tgt.CICode, tgt.Name))

	return s.repo.GetRelationship(id)
}

func (s *cmdbServiceImpl) ListRelationships(ciID int64) ([]model.CIRelationship, error) {
	if _, err := s.repo.GetCIByID(ciID); err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return s.repo.ListRelationships(ciID)
}

func (s *cmdbServiceImpl) DeleteRelationship(id int64) error {
	rel, err := s.repo.GetRelationship(id)
	if err != nil {
		if errors.Is(err, repository.ErrRelationshipNotFound) {
			return ErrRelationshipNotFound
		}
		return err
	}

	if err := s.repo.DeleteRelationship(id); err != nil {
		return err
	}

	s.logAudit("", 0, "cmdb.relationship_deleted", fmt.Sprintf("%d", id), fmt.Sprintf("Deleted relationship %s -> %s", rel.SourceCICode, rel.TargetCICode))
	return nil
}

func (s *cmdbServiceImpl) GetDependencyGraph(ciID int64, maxDepth int) (*model.CIDependencyGraph, error) {
	rootCI, err := s.repo.GetCIByID(ciID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	if maxDepth <= 0 || maxDepth > 10 {
		maxDepth = 5
	}

	upstream, err := s.repo.GetUpstreamDependencies(ciID, maxDepth)
	if err != nil {
		return nil, err
	}
	if upstream == nil {
		upstream = []model.CIDependencyNode{}
	}

	downstream, err := s.repo.GetDownstreamDependencies(ciID, maxDepth)
	if err != nil {
		return nil, err
	}
	if downstream == nil {
		downstream = []model.CIDependencyNode{}
	}

	computedMaxDepth := 0
	for _, n := range upstream {
		if n.Depth > computedMaxDepth {
			computedMaxDepth = n.Depth
		}
	}
	for _, n := range downstream {
		if n.Depth > computedMaxDepth {
			computedMaxDepth = n.Depth
		}
	}

	return &model.CIDependencyGraph{
		RootCI:     *rootCI,
		Upstream:   upstream,
		Downstream: downstream,
		TotalNodes: len(upstream) + len(downstream),
		MaxDepth:   computedMaxDepth,
	}, nil
}

func (s *cmdbServiceImpl) AnalyzeBlastRadius(ciID int64) (*model.CIImpactAnalysis, error) {
	targetCI, err := s.repo.GetCIByID(ciID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	downstream, err := s.repo.GetDownstreamDependencies(ciID, 10)
	if err != nil {
		return nil, err
	}
	if downstream == nil {
		downstream = []model.CIDependencyNode{}
	}

	var directCount, indirectCount, srvCount, dbCount int
	var hasTier1 bool

	for _, n := range downstream {
		if n.Depth == 1 {
			directCount++
		} else {
			indirectCount++
		}
		if n.CI.CIType == model.CITypeService {
			srvCount++
		} else if n.CI.CIType == model.CITypeDatabase {
			dbCount++
		}
		if n.CI.Criticality == model.CICriticalityTier1 {
			hasTier1 = true
		}
	}

	return &model.CIImpactAnalysis{
		TargetCI:               *targetCI,
		DirectImpactCount:      directCount,
		IndirectImpactCount:    indirectCount,
		ImpactedCIs:            downstream,
		ImpactedServicesCount:  srvCount,
		ImpactedDatabasesCount: dbCount,
		HasTier1Impact:         hasTier1,
	}, nil
}

func (s *cmdbServiceImpl) LinkEntity(ciID int64, req model.CreateCIEntityLinkRequest) (*model.CIEntityLink, error) {
	ci, err := s.repo.GetCIByID(ciID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	entityType := strings.ToLower(strings.TrimSpace(req.EntityType))
	if err := model.ValidateCILinkEntityType(entityType); err != nil {
		return nil, err
	}

	if req.EntityID <= 0 {
		return nil, errors.New("entity_id is required")
	}

	impact := strings.ToLower(strings.TrimSpace(req.ImpactLevel))
	if impact == "" {
		impact = model.CIImpactDirect
	} else if err := model.ValidateCIImpactLevel(impact); err != nil {
		return nil, err
	}

	existing, err := s.repo.ListEntityLinksByCI(ciID)
	if err == nil {
		for _, l := range existing {
			if l.EntityType == entityType && l.EntityID == req.EntityID {
				return nil, ErrLinkAlreadyExists
			}
		}
	}

	link := &model.CIEntityLink{
		CIID:        ciID,
		EntityType:  entityType,
		EntityID:    req.EntityID,
		ImpactLevel: impact,
		Notes:       strings.TrimSpace(req.Notes),
	}

	id, err := s.repo.CreateEntityLink(link)
	if err != nil {
		return nil, err
	}

	s.logAudit("", ci.WorkspaceID, "cmdb.entity_linked", fmt.Sprintf("%d", id), fmt.Sprintf("Linked %s to %s #%d (impact: %s)", ci.CICode, entityType, req.EntityID, impact))

	links, err := s.repo.ListEntityLinksByCI(ciID)
	if err == nil {
		for _, l := range links {
			if l.ID == id {
				return &l, nil
			}
		}
	}
	link.ID = id
	return link, nil
}

func (s *cmdbServiceImpl) ListEntityLinks(ciID int64) ([]model.CIEntityLink, error) {
	if _, err := s.repo.GetCIByID(ciID); err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return s.repo.ListEntityLinksByCI(ciID)
}

func (s *cmdbServiceImpl) ListCIsForEntity(entityType string, entityID int64) ([]model.CIEntityLink, error) {
	et := strings.ToLower(strings.TrimSpace(entityType))
	if err := model.ValidateCILinkEntityType(et); err != nil {
		return nil, err
	}
	return s.repo.ListCIsByEntity(et, entityID)
}

func (s *cmdbServiceImpl) UnlinkEntity(linkID int64) error {
	return s.repo.DeleteEntityLink(linkID)
}

func parseMaintTime(str string) (time.Time, error) {
	str = strings.TrimSpace(str)
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", str)
		if err != nil {
			t, err = time.Parse("2006-01-02", str)
			if err != nil {
				return time.Time{}, err
			}
		}
	}
	return t, nil
}

func (s *cmdbServiceImpl) CreateMaintenance(ciID int64, creatorID string, req model.CreateCIMaintenanceRequest) (*model.CIMaintenanceSchedule, error) {
	ci, err := s.repo.GetCIByID(ciID)
	if err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrMaintTitleRequired
	}

	start, err := parseMaintTime(req.WindowStartAt)
	if err != nil {
		return nil, errors.New("invalid window_start_at timestamp: must be RFC3339")
	}

	end, err := parseMaintTime(req.WindowEndAt)
	if err != nil {
		return nil, errors.New("invalid window_end_at timestamp: must be RFC3339")
	}

	if !end.After(start) {
		return nil, ErrInvalidMaintDates
	}

	maint := &model.CIMaintenanceSchedule{
		CIID:                    ciID,
		Title:                   title,
		Description:             strings.TrimSpace(req.Description),
		WindowStartAt:           start,
		WindowEndAt:             end,
		ExpectedDowntimeMinutes: req.ExpectedDowntimeMinutes,
		Status:                  model.CIMaintStatusScheduled,
		CreatedByUserID:         creatorID,
	}

	id, err := s.repo.CreateMaintenance(maint)
	if err != nil {
		return nil, err
	}

	s.logAudit(creatorID, ci.WorkspaceID, "cmdb.maintenance_scheduled", fmt.Sprintf("%d", id), fmt.Sprintf("Scheduled maintenance '%s' on %s", title, ci.CICode))

	return s.repo.GetMaintenanceByID(id)
}

func (s *cmdbServiceImpl) ListMaintenances(ciID int64) ([]model.CIMaintenanceSchedule, error) {
	if _, err := s.repo.GetCIByID(ciID); err != nil {
		if errors.Is(err, repository.ErrCINotFound) {
			return nil, ErrCINotFound
		}
		return nil, err
	}
	return s.repo.ListMaintenancesByCI(ciID)
}

func (s *cmdbServiceImpl) UpdateMaintenanceStatus(id int64, req model.UpdateCIMaintenanceStatusRequest) (*model.CIMaintenanceSchedule, error) {
	st := strings.ToLower(strings.TrimSpace(req.Status))
	if err := model.ValidateCIMaintStatus(st); err != nil {
		return nil, err
	}

	maint, err := s.repo.GetMaintenanceByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrMaintenanceNotFound) {
			return nil, ErrMaintenanceNotFound
		}
		return nil, err
	}

	if err := s.repo.UpdateMaintenanceStatus(id, st); err != nil {
		return nil, err
	}

	s.logAudit("", 0, "cmdb.maintenance_status_changed", fmt.Sprintf("%d", id), fmt.Sprintf("Updated maintenance %d status to %s", id, st))

	return s.repo.GetMaintenanceByID(maint.ID)
}

func (s *cmdbServiceImpl) DeleteMaintenance(id int64) error {
	return s.repo.DeleteMaintenance(id)
}

func (s *cmdbServiceImpl) GetCMDBMetrics(workspaceID int64) (*model.CMDBMetrics, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(workspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}
	return s.repo.GetCMDBMetrics(workspaceID)
}
