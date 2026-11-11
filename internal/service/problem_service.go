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
	ErrProblemTitleRequired         = errors.New("problem title is required")
	ErrProblemNotFound              = errors.New("problem not found")
	ErrInvalidProblemStatus         = errors.New("invalid problem status")
	ErrInvalidProblemCategory       = errors.New("invalid problem category")
	ErrInvalidProblemPriority       = errors.New("invalid problem priority")
	ErrInvalidProblemUrgency        = errors.New("invalid problem urgency")
	ErrInvalidProblemImpact         = errors.New("invalid problem impact")
	ErrInvalidRCAMethod             = errors.New("invalid root cause analysis method")
	ErrInvalidProblemTransition     = errors.New("illegal problem status transition")
	ErrKnownErrorRequiresWorkaround = errors.New("known error requires an approved workaround or root cause description")
	ErrProblemLinkExists            = errors.New("problem link already exists")
	ErrProblemLinkNotFound          = errors.New("problem link not found")
	ErrRCAStepNotFound              = errors.New("root cause step not found")
)

// ProblemService provides ITIL Problem Management, 5-Whys RCA, and KEDB business operations
type ProblemService interface {
	// Problem Lifecycle
	CreateProblem(creatorID string, req model.CreateProblemRequest) (*model.Problem, error)
	GetProblem(id int64) (*model.Problem, error)
	GetProblemByCode(workspaceID int64, code string) (*model.Problem, error)
	ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error)
	UpdateProblem(id int64, updaterID string, req model.UpdateProblemRequest) (*model.Problem, error)
	DeleteProblem(id int64, deleterID string) error
	UpdateProblemStatus(id int64, updaterID string, req model.UpdateProblemStatusRequest) (*model.Problem, error)

	// Associations
	LinkIncident(problemID int64, linkerID string, req model.LinkIncidentRequest) (*model.ProblemIncidentLink, error)
	UnlinkIncident(problemID, incidentID int64, unlinkerID string) error
	ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error)

	LinkCI(problemID int64, linkerID string, req model.LinkCIRequest) (*model.ProblemCILink, error)
	UnlinkCI(problemID, linkID int64, unlinkerID string) error
	ListProblemCIs(problemID int64) ([]model.ProblemCILink, error)

	LinkChange(problemID int64, linkerID string, req model.LinkChangeRequest) (*model.ProblemChangeLink, error)
	UnlinkChange(problemID, linkID int64, unlinkerID string) error
	ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error)

	LinkTask(problemID int64, linkerID string, req model.LinkTaskRequest) (*model.ProblemTaskLink, error)
	UnlinkTask(problemID, linkID int64, unlinkerID string) error
	ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error)

	// 5-Whys Root Cause Analysis
	AddRCAStep(problemID int64, authorID string, req model.CreateRCAStepRequest) (*model.ProblemRCAStep, error)
	GetRCAStep(id int64) (*model.ProblemRCAStep, error)
	UpdateRCAStep(id int64, updaterID string, req model.UpdateRCAStepRequest) (*model.ProblemRCAStep, error)
	DeleteRCAStep(id int64, deleterID string) error
	ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error)

	// Known Error Database (KEDB)
	PublishKnownError(problemID int64, publisherID string, req model.PublishKnownErrorRequest) (*model.Problem, error)
	RetireKnownError(problemID int64, actorID string) (*model.Problem, error)
	ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error)

	// Metrics
	GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error)
}

type problemServiceImpl struct {
	repo          repository.ProblemRepository
	workspaceRepo repository.WorkspaceRepository
	projectRepo   repository.ProjectRepository
	taskRepo      repository.TaskRepository
	incidentRepo  repository.IncidentRepository
	cmdbRepo      repository.CMDBRepository
	changeRepo    repository.ChangeRepository
	auditService  AuditService
}

// NewProblemService creates a new ProblemService instance
func NewProblemService(
	repo repository.ProblemRepository,
	workspaceRepo repository.WorkspaceRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	incidentRepo repository.IncidentRepository,
	cmdbRepo repository.CMDBRepository,
	changeRepo repository.ChangeRepository,
	auditService AuditService,
) ProblemService {
	return &problemServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		taskRepo:      taskRepo,
		incidentRepo:  incidentRepo,
		cmdbRepo:      cmdbRepo,
		changeRepo:    changeRepo,
		auditService:  auditService,
	}
}

func (s *problemServiceImpl) logAudit(actorID string, workspaceID int64, action, entityID, description string) {
	if s.auditService != nil {
		_, _ = s.auditService.RecordEvent(actorID, model.CreateAuditLogRequest{
			WorkspaceID: workspaceID,
			Action:      action,
			Category:    "problem_management",
			Severity:    "info",
			EntityType:  "problem",
			EntityID:    entityID,
			Description: description,
		})
	}
}

func (s *problemServiceImpl) CreateProblem(creatorID string, req model.CreateProblemRequest) (*model.Problem, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	if req.ProjectID != nil && *req.ProjectID > 0 {
		proj, err := s.projectRepo.GetByID(int(*req.ProjectID))
		if err != nil || proj == nil {
			return nil, ErrProjectNotFound
		}
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrProblemTitleRequired
	}

	cat := strings.ToLower(strings.TrimSpace(req.Category))
	if cat == "" {
		cat = model.ProblemCategorySoftware
	} else if err := model.ValidateProblemCategory(cat); err != nil {
		return nil, err
	}

	pri := strings.ToLower(strings.TrimSpace(req.Priority))
	if pri == "" {
		pri = model.ProblemPriorityMedium
	} else if err := model.ValidateProblemPriority(pri); err != nil {
		return nil, err
	}

	urg := strings.ToLower(strings.TrimSpace(req.Urgency))
	if urg == "" {
		urg = model.ProblemUrgencyMedium
	} else if err := model.ValidateProblemUrgency(urg); err != nil {
		return nil, err
	}

	imp := strings.ToLower(strings.TrimSpace(req.Impact))
	if imp == "" {
		imp = model.ProblemImpactMedium
	} else if err := model.ValidateProblemImpact(imp); err != nil {
		return nil, err
	}

	rcaMethod := strings.ToLower(strings.TrimSpace(req.RCAMethod))
	if err := model.ValidateRCAMethod(rcaMethod); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	prob := &model.Problem{
		WorkspaceID:  req.WorkspaceID,
		ProjectID:    req.ProjectID,
		ProblemCode:  strings.ToUpper(strings.TrimSpace(req.ProblemCode)),
		Title:        title,
		Description:  strings.TrimSpace(req.Description),
		Status:       model.ProblemStatusLogged,
		Category:     cat,
		Priority:     pri,
		Urgency:      urg,
		Impact:       imp,
		AssignedTo:   req.AssignedTo,
		AssignedTeam: strings.TrimSpace(req.AssignedTeam),
		RCAMethod:    rcaMethod,
		CreatedByID:  creatorID,
		StartedAt:    &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	id, err := s.repo.CreateProblem(prob)
	if err != nil {
		return nil, err
	}

	// Link initial incidents if provided
	for _, incID := range req.IncidentIDs {
		_, _ = s.repo.LinkIncident(&model.ProblemIncidentLink{
			ProblemID:  id,
			IncidentID: incID,
			Notes:      "Initial incident association upon problem registration",
			LinkedAt:   now,
		})
	}

	// Link initial CIs if provided
	for _, ciID := range req.CIIDs {
		_, _ = s.repo.LinkCI(&model.ProblemCILink{
			ProblemID:    id,
			CIID:         ciID,
			Relationship: model.ProblemCIRelImpacted,
			Notes:        "Initial configuration item association upon problem registration",
			LinkedAt:     now,
		})
	}

	s.logAudit(creatorID, req.WorkspaceID, "problem.created", fmt.Sprintf("%d", id), fmt.Sprintf("Registered problem %s (%s)", prob.ProblemCode, title))

	return s.repo.GetProblemByID(id)
}

func (s *problemServiceImpl) GetProblem(id int64) (*model.Problem, error) {
	p, err := s.repo.GetProblemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *problemServiceImpl) GetProblemByCode(workspaceID int64, code string) (*model.Problem, error) {
	p, err := s.repo.GetProblemByCode(workspaceID, code)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *problemServiceImpl) ListProblems(filter model.ProblemFilter) ([]model.Problem, int, error) {
	if filter.WorkspaceID <= 0 {
		return nil, 0, ErrWorkspaceRequired
	}
	return s.repo.ListProblems(filter)
}

func (s *problemServiceImpl) UpdateProblem(id int64, updaterID string, req model.UpdateProblemRequest) (*model.Problem, error) {
	prob, err := s.repo.GetProblemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrProblemTitleRequired
		}
		prob.Title = t
	}
	if req.Description != nil {
		prob.Description = strings.TrimSpace(*req.Description)
	}
	if req.Category != nil {
		cat := strings.ToLower(strings.TrimSpace(*req.Category))
		if err := model.ValidateProblemCategory(cat); err != nil {
			return nil, err
		}
		prob.Category = cat
	}
	if req.Priority != nil {
		pri := strings.ToLower(strings.TrimSpace(*req.Priority))
		if err := model.ValidateProblemPriority(pri); err != nil {
			return nil, err
		}
		prob.Priority = pri
	}
	if req.Urgency != nil {
		urg := strings.ToLower(strings.TrimSpace(*req.Urgency))
		if err := model.ValidateProblemUrgency(urg); err != nil {
			return nil, err
		}
		prob.Urgency = urg
	}
	if req.Impact != nil {
		imp := strings.ToLower(strings.TrimSpace(*req.Impact))
		if err := model.ValidateProblemImpact(imp); err != nil {
			return nil, err
		}
		prob.Impact = imp
	}
	if req.AssignedTo != nil {
		prob.AssignedTo = req.AssignedTo
	}
	if req.AssignedTeam != nil {
		prob.AssignedTeam = strings.TrimSpace(*req.AssignedTeam)
	}
	if req.RootCauseCategory != nil {
		prob.RootCauseCategory = strings.TrimSpace(*req.RootCauseCategory)
	}
	if req.RootCauseSummary != nil {
		prob.RootCauseSummary = strings.TrimSpace(*req.RootCauseSummary)
	}
	if req.Workaround != nil {
		prob.Workaround = strings.TrimSpace(*req.Workaround)
	}
	if req.PermanentSolution != nil {
		prob.PermanentSolution = strings.TrimSpace(*req.PermanentSolution)
	}
	if req.RCAMethod != nil {
		m := strings.ToLower(strings.TrimSpace(*req.RCAMethod))
		if err := model.ValidateRCAMethod(m); err != nil {
			return nil, err
		}
		prob.RCAMethod = m
	}

	if err := s.repo.UpdateProblem(prob); err != nil {
		return nil, err
	}

	s.logAudit(updaterID, prob.WorkspaceID, "problem.updated", fmt.Sprintf("%d", id), fmt.Sprintf("Updated problem %s (%s)", prob.ProblemCode, prob.Title))

	return s.repo.GetProblemByID(id)
}

func (s *problemServiceImpl) DeleteProblem(id int64, deleterID string) error {
	prob, err := s.repo.GetProblemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return ErrProblemNotFound
		}
		return err
	}

	if err := s.repo.DeleteProblem(id); err != nil {
		return err
	}

	s.logAudit(deleterID, prob.WorkspaceID, "problem.deleted", fmt.Sprintf("%d", id), fmt.Sprintf("Deleted problem %s (%s)", prob.ProblemCode, prob.Title))
	return nil
}

func (s *problemServiceImpl) UpdateProblemStatus(id int64, updaterID string, req model.UpdateProblemStatusRequest) (*model.Problem, error) {
	prob, err := s.repo.GetProblemByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrProblemNotFound) {
			return nil, ErrProblemNotFound
		}
		return nil, err
	}

	targetStatus := strings.ToLower(strings.TrimSpace(req.Status))
	if err := model.ValidateProblemTransition(prob.Status, targetStatus); err != nil {
		return nil, ErrInvalidProblemTransition
	}

	if req.Workaround != "" {
		prob.Workaround = strings.TrimSpace(req.Workaround)
	}
	if req.PermanentSolution != "" {
		prob.PermanentSolution = strings.TrimSpace(req.PermanentSolution)
	}

	now := time.Now().UTC()
	switch targetStatus {
	case model.ProblemStatusKnownError:
		if prob.Workaround == "" && prob.RootCauseSummary == "" {
			return nil, ErrKnownErrorRequiresWorkaround
		}
		prob.IsKnownError = true
		prob.KnownErrorPublishedAt = &now
	case model.ProblemStatusResolved:
		prob.ResolvedAt = &now
	case model.ProblemStatusClosed:
		prob.ClosedAt = &now
	case model.ProblemStatusInvestigating:
		// Re-opening clears resolved and closed timestamps
		prob.ResolvedAt = nil
		prob.ClosedAt = nil
	}

	prob.Status = targetStatus
	if err := s.repo.UpdateProblem(prob); err != nil {
		return nil, err
	}

	s.logAudit(updaterID, prob.WorkspaceID, "problem.status_changed", fmt.Sprintf("%d", id), fmt.Sprintf("Transitioned problem %s status to %s (reason: %s)", prob.ProblemCode, targetStatus, req.Reason))

	return s.repo.GetProblemByID(id)
}

func (s *problemServiceImpl) LinkIncident(problemID int64, linkerID string, req model.LinkIncidentRequest) (*model.ProblemIncidentLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) UnlinkIncident(problemID, incidentID int64, unlinkerID string) error {
	return nil
}

func (s *problemServiceImpl) ListProblemIncidents(problemID int64) ([]model.ProblemIncidentLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) LinkCI(problemID int64, linkerID string, req model.LinkCIRequest) (*model.ProblemCILink, error) {
	return nil, nil
}

func (s *problemServiceImpl) UnlinkCI(problemID, linkID int64, unlinkerID string) error {
	return nil
}

func (s *problemServiceImpl) ListProblemCIs(problemID int64) ([]model.ProblemCILink, error) {
	return nil, nil
}

func (s *problemServiceImpl) LinkChange(problemID int64, linkerID string, req model.LinkChangeRequest) (*model.ProblemChangeLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) UnlinkChange(problemID, linkID int64, unlinkerID string) error {
	return nil
}

func (s *problemServiceImpl) ListProblemChanges(problemID int64) ([]model.ProblemChangeLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) LinkTask(problemID int64, linkerID string, req model.LinkTaskRequest) (*model.ProblemTaskLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) UnlinkTask(problemID, linkID int64, unlinkerID string) error {
	return nil
}

func (s *problemServiceImpl) ListProblemTasks(problemID int64) ([]model.ProblemTaskLink, error) {
	return nil, nil
}

func (s *problemServiceImpl) AddRCAStep(problemID int64, authorID string, req model.CreateRCAStepRequest) (*model.ProblemRCAStep, error) {
	return nil, nil
}

func (s *problemServiceImpl) GetRCAStep(id int64) (*model.ProblemRCAStep, error) {
	return nil, nil
}

func (s *problemServiceImpl) UpdateRCAStep(id int64, updaterID string, req model.UpdateRCAStepRequest) (*model.ProblemRCAStep, error) {
	return nil, nil
}

func (s *problemServiceImpl) DeleteRCAStep(id int64, deleterID string) error {
	return nil
}

func (s *problemServiceImpl) ListRCASteps(problemID int64) ([]model.ProblemRCAStep, error) {
	return nil, nil
}

func (s *problemServiceImpl) PublishKnownError(problemID int64, publisherID string, req model.PublishKnownErrorRequest) (*model.Problem, error) {
	return nil, nil
}

func (s *problemServiceImpl) RetireKnownError(problemID int64, actorID string) (*model.Problem, error) {
	return nil, nil
}

func (s *problemServiceImpl) ListKnownErrors(filter model.KnownErrorFilter) ([]model.Problem, int, error) {
	return nil, 0, nil
}

func (s *problemServiceImpl) GetProblemMetrics(workspaceID int64) (*model.ProblemMetrics, error) {
	return nil, nil
}
