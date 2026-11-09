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
	ErrChangeNotFound          = repository.ErrChangeNotFound
	ErrApprovalNotFound        = repository.ErrApprovalNotFound
	ErrPIRNotFound             = repository.ErrPIRNotFound
	ErrChangeTitleRequired     = errors.New("change title is required")
	ErrInvalidChangeType       = model.ErrInvalidChangeType
	ErrInvalidChangeRiskLevel  = model.ErrInvalidChangeRiskLevel
	ErrInvalidChangeStatus     = model.ErrInvalidChangeStatus
	ErrInvalidChangeTransition = model.ErrInvalidChangeTransition
	ErrChangeNotEditable       = errors.New("cannot update change request in current status")
	ErrReviewerRequired        = errors.New("reviewer_id is required")
	ErrInvalidCABDecision      = errors.New("invalid CAB decision")
	ErrInvalidPIROutcome       = errors.New("invalid PIR outcome")
	ErrPIRNotesRequired        = errors.New("PIR review notes are required")
	ErrTaskAlreadyLinked       = errors.New("task is already linked to this change request")
	ErrInvalidChangePhase      = errors.New("invalid implementation phase")
)

// ChangeService manages ITIL change lifecycle, CAB reviews, tasks, PIR, and performance metrics
type ChangeService interface {
	CreateChange(requesterID string, req model.CreateChangeRequest) (*model.ChangeRequest, error)
	GetChange(id int64) (*model.ChangeRequest, error)
	GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error)
	ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error)
	UpdateChange(id int64, userID string, req model.UpdateChangeRequest) (*model.ChangeRequest, error)
	DeleteChange(id int64, userID string) error
	ChangeStatus(id int64, userID string, req model.ChangeStatusRequest) (*model.ChangeRequest, error)

	AddReviewer(changeID int64, userID string, req model.AddReviewerRequest) (*model.ChangeApproval, error)
	ReviewChange(changeID int64, reviewerID string, req model.ChangeApprovalRequest) (*model.ChangeApproval, error)
	ListApprovals(changeID int64) ([]model.ChangeApproval, error)
	DeleteReviewer(approvalID int64, userID string) error

	LinkTask(changeID int64, req model.LinkChangeTaskRequest) (*model.ChangeTaskLink, error)
	UnlinkTask(changeID int64, taskID int) error
	ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error)

	SubmitPIR(changeID int64, reviewerID string, req model.ChangePIRRequest) (*model.ChangePIR, error)
	GetPIR(changeID int64) (*model.ChangePIR, error)

	GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error)
}

type changeServiceImpl struct {
	repo          repository.ChangeRepository
	workspaceRepo repository.WorkspaceRepository
	projectRepo   repository.ProjectRepository
	taskRepo      repository.TaskRepository
	userRepo      repository.UserRepository
	auditService  AuditService
}

// NewChangeService creates a new instance of ChangeService
func NewChangeService(
	repo repository.ChangeRepository,
	workspaceRepo repository.WorkspaceRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	auditService AuditService,
) ChangeService {
	return &changeServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		auditService:  auditService,
	}
}

func parseOptionalTime(s *string) (*time.Time, error) {
	if s == nil || strings.TrimSpace(*s) == "" {
		return nil, nil
	}
	str := strings.TrimSpace(*s)
	t, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t, err = time.Parse("2006-01-02T15:04:05", str)
		if err != nil {
			t, err = time.Parse("2006-01-02", str)
			if err != nil {
				return nil, err
			}
		}
	}
	return &t, nil
}

func (s *changeServiceImpl) logAudit(actorID string, workspaceID int64, action, entityID, description string) {
	if s.auditService != nil {
		_, _ = s.auditService.RecordEvent(actorID, model.CreateAuditLogRequest{
			WorkspaceID: workspaceID,
			Action:      action,
			Category:    "change",
			Severity:    "info",
			EntityType:  "change_request",
			EntityID:    entityID,
			Description: description,
		})
	}
}

func (s *changeServiceImpl) CreateChange(requesterID string, req model.CreateChangeRequest) (*model.ChangeRequest, error) {
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrChangeTitleRequired
	}
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	ws, err := s.workspaceRepo.GetByID(int(req.WorkspaceID))
	if err != nil || ws == nil {
		return nil, ErrWorkspaceNotFound
	}

	changeType := strings.ToLower(strings.TrimSpace(req.ChangeType))
	if changeType == "" {
		changeType = model.ChangeTypeNormal
	} else if err := model.ValidateChangeType(changeType); err != nil {
		return nil, err
	}

	riskLevel := strings.ToLower(strings.TrimSpace(req.RiskLevel))
	if riskLevel == "" {
		riskLevel = model.ChangeRiskMedium
	} else if err := model.ValidateChangeRiskLevel(riskLevel); err != nil {
		return nil, err
	}

	impactLevel := strings.ToLower(strings.TrimSpace(req.ImpactLevel))
	if impactLevel == "" {
		impactLevel = model.ChangeImpactLow
	}

	plannedStart, err := parseOptionalTime(req.PlannedStartAt)
	if err != nil {
		return nil, errors.New("invalid planned_start_at timestamp")
	}
	plannedEnd, err := parseOptionalTime(req.PlannedEndAt)
	if err != nil {
		return nil, errors.New("invalid planned_end_at timestamp")
	}

	change := &model.ChangeRequest{
		WorkspaceID:        req.WorkspaceID,
		ProjectID:          req.ProjectID,
		Title:              title,
		Description:        strings.TrimSpace(req.Description),
		ChangeType:         changeType,
		RiskLevel:          riskLevel,
		ImpactLevel:        impactLevel,
		Status:             model.ChangeStatusDraft,
		ReasonForChange:    strings.TrimSpace(req.ReasonForChange),
		ImplementationPlan: strings.TrimSpace(req.ImplementationPlan),
		RollbackPlan:       strings.TrimSpace(req.RollbackPlan),
		TestPlan:           strings.TrimSpace(req.TestPlan),
		RequestedByUserID:  requesterID,
		AssignedToUserID:   strings.TrimSpace(req.AssignedToUserID),
		PlannedStartAt:     plannedStart,
		PlannedEndAt:       plannedEnd,
	}

	id, err := s.repo.CreateChange(change)
	if err != nil {
		return nil, err
	}

	s.logAudit(requesterID, req.WorkspaceID, "change_request.created", fmt.Sprintf("%d", id), fmt.Sprintf("Created change request %s (%s)", change.ChangeNumber, title))

	return s.repo.GetChangeByID(id)
}

func (s *changeServiceImpl) GetChange(id int64) (*model.ChangeRequest, error) {
	change, err := s.repo.GetChangeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}
	return change, nil
}

func (s *changeServiceImpl) GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error) {
	change, err := s.repo.GetChangeByNumber(workspaceID, changeNumber)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}
	return change, nil
}

func (s *changeServiceImpl) ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.repo.ListChanges(filter)
}

func (s *changeServiceImpl) UpdateChange(id int64, userID string, req model.UpdateChangeRequest) (*model.ChangeRequest, error) {
	change, err := s.repo.GetChangeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if change.Status != model.ChangeStatusDraft && change.Status != model.ChangeStatusSubmitted && change.Status != model.ChangeStatusCABReview {
		return nil, ErrChangeNotEditable
	}

	if req.Title != nil {
		t := strings.TrimSpace(*req.Title)
		if t == "" {
			return nil, ErrChangeTitleRequired
		}
		change.Title = t
	}
	if req.Description != nil {
		change.Description = strings.TrimSpace(*req.Description)
	}
	if req.ProjectID != nil {
		change.ProjectID = req.ProjectID
	}
	if req.ChangeType != nil {
		ct := strings.ToLower(strings.TrimSpace(*req.ChangeType))
		if err := model.ValidateChangeType(ct); err != nil {
			return nil, err
		}
		change.ChangeType = ct
	}
	if req.RiskLevel != nil {
		rl := strings.ToLower(strings.TrimSpace(*req.RiskLevel))
		if err := model.ValidateChangeRiskLevel(rl); err != nil {
			return nil, err
		}
		change.RiskLevel = rl
	}
	if req.ImpactLevel != nil {
		change.ImpactLevel = strings.ToLower(strings.TrimSpace(*req.ImpactLevel))
	}
	if req.ReasonForChange != nil {
		change.ReasonForChange = strings.TrimSpace(*req.ReasonForChange)
	}
	if req.ImplementationPlan != nil {
		change.ImplementationPlan = strings.TrimSpace(*req.ImplementationPlan)
	}
	if req.RollbackPlan != nil {
		change.RollbackPlan = strings.TrimSpace(*req.RollbackPlan)
	}
	if req.TestPlan != nil {
		change.TestPlan = strings.TrimSpace(*req.TestPlan)
	}
	if req.AssignedToUserID != nil {
		change.AssignedToUserID = strings.TrimSpace(*req.AssignedToUserID)
	}
	if req.PlannedStartAt != nil {
		t, err := parseOptionalTime(req.PlannedStartAt)
		if err != nil {
			return nil, errors.New("invalid planned_start_at timestamp")
		}
		change.PlannedStartAt = t
	}
	if req.PlannedEndAt != nil {
		t, err := parseOptionalTime(req.PlannedEndAt)
		if err != nil {
			return nil, errors.New("invalid planned_end_at timestamp")
		}
		change.PlannedEndAt = t
	}

	if err := s.repo.UpdateChange(change); err != nil {
		return nil, err
	}

	s.logAudit(userID, change.WorkspaceID, "change_request.updated", fmt.Sprintf("%d", id), fmt.Sprintf("Updated change request %s (%s)", change.ChangeNumber, change.Title))

	return s.repo.GetChangeByID(id)
}

func (s *changeServiceImpl) DeleteChange(id int64, userID string) error {
	change, err := s.repo.GetChangeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return ErrChangeNotFound
		}
		return err
	}

	if change.Status != model.ChangeStatusDraft && change.Status != model.ChangeStatusCancelled {
		return ErrChangeNotEditable
	}

	if err := s.repo.DeleteChange(id); err != nil {
		return err
	}

	s.logAudit(userID, change.WorkspaceID, "change_request.deleted", fmt.Sprintf("%d", id), fmt.Sprintf("Deleted change request %s (%s)", change.ChangeNumber, change.Title))
	return nil
}

func (s *changeServiceImpl) ChangeStatus(id int64, userID string, req model.ChangeStatusRequest) (*model.ChangeRequest, error) {
	targetStatus := strings.ToLower(strings.TrimSpace(req.Status))
	if err := model.ValidateChangeStatus(targetStatus); err != nil {
		return nil, ErrInvalidChangeStatus
	}

	change, err := s.repo.GetChangeByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	if err := model.ValidateChangeTransition(change.Status, targetStatus, change.ChangeType); err != nil {
		return nil, ErrInvalidChangeTransition
	}

	if targetStatus == model.ChangeStatusApproved {
		approvals, err := s.repo.ListApprovals(id)
		if err == nil {
			for _, a := range approvals {
				if a.Decision == model.ApprovalDecisionRejected {
					return nil, errors.New("cannot approve change request with rejected CAB review")
				}
			}
		}
	}

	if targetStatus == model.ChangeStatusImplementing {
		if change.ActualStartAt == nil {
			now := time.Now().UTC()
			if req.ActualStartAt != nil {
				parsed, err := parseOptionalTime(req.ActualStartAt)
				if err == nil && parsed != nil {
					now = *parsed
				}
			}
			change.ActualStartAt = &now
		}
	}

	if targetStatus == model.ChangeStatusCompleted || targetStatus == model.ChangeStatusFailed || targetStatus == model.ChangeStatusRolledBack {
		if change.ActualEndAt == nil {
			now := time.Now().UTC()
			if req.ActualEndAt != nil {
				parsed, err := parseOptionalTime(req.ActualEndAt)
				if err == nil && parsed != nil {
					now = *parsed
				}
			}
			change.ActualEndAt = &now
		}
	}

	change.Status = targetStatus

	if err := s.repo.UpdateChange(change); err != nil {
		return nil, err
	}

	s.logAudit(userID, change.WorkspaceID, "change_request.status_changed", fmt.Sprintf("%d", id), fmt.Sprintf("Changed status of %s to %s", change.ChangeNumber, targetStatus))

	return s.repo.GetChangeByID(id)
}

func (s *changeServiceImpl) AddReviewer(changeID int64, userID string, req model.AddReviewerRequest) (*model.ChangeApproval, error) {
	reviewerID := strings.TrimSpace(req.ReviewerID)
	if reviewerID == "" {
		return nil, ErrReviewerRequired
	}

	change, err := s.repo.GetChangeByID(changeID)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	existing, err := s.repo.GetApprovalByChangeAndReviewer(changeID, reviewerID)
	if err == nil && existing != nil {
		return existing, nil
	}

	appr := &model.ChangeApproval{
		ChangeID:   changeID,
		ReviewerID: reviewerID,
		Decision:   model.ApprovalDecisionPending,
		CreatedAt:  time.Now().UTC(),
	}

	id, err := s.repo.AddReviewer(appr)
	if err != nil {
		return nil, err
	}

	s.logAudit(userID, change.WorkspaceID, "change_approval.reviewer_added", fmt.Sprintf("%d", id), fmt.Sprintf("Added reviewer %s to change %s", reviewerID, change.ChangeNumber))

	return s.repo.GetApprovalByID(id)
}

func (s *changeServiceImpl) ReviewChange(changeID int64, reviewerID string, req model.ChangeApprovalRequest) (*model.ChangeApproval, error) {
	decision := strings.ToLower(strings.TrimSpace(req.Decision))
	if decision != model.ApprovalDecisionApproved && decision != model.ApprovalDecisionRejected && decision != model.ApprovalDecisionAbstained {
		return nil, ErrInvalidCABDecision
	}

	change, err := s.repo.GetChangeByID(changeID)
	if err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}

	existing, err := s.repo.GetApprovalByChangeAndReviewer(changeID, reviewerID)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			now := time.Now().UTC()
			appr := &model.ChangeApproval{
				ChangeID:   changeID,
				ReviewerID: reviewerID,
				Decision:   decision,
				Comments:   strings.TrimSpace(req.Comments),
				DecidedAt:  &now,
				CreatedAt:  now,
			}
			id, err := s.repo.AddReviewer(appr)
			if err != nil {
				return nil, err
			}
			s.logAudit(reviewerID, change.WorkspaceID, "change_approval.decision_cast", fmt.Sprintf("%d", id), fmt.Sprintf("Reviewer %s voted %s on change %s", reviewerID, decision, change.ChangeNumber))
			return s.repo.GetApprovalByID(id)
		}
		return nil, err
	}

	if err := s.repo.CastDecision(changeID, reviewerID, decision, strings.TrimSpace(req.Comments)); err != nil {
		return nil, err
	}

	s.logAudit(reviewerID, change.WorkspaceID, "change_approval.decision_cast", fmt.Sprintf("%d", existing.ID), fmt.Sprintf("Reviewer %s voted %s on change %s", reviewerID, decision, change.ChangeNumber))

	return s.repo.GetApprovalByID(existing.ID)
}

func (s *changeServiceImpl) ListApprovals(changeID int64) ([]model.ChangeApproval, error) {
	if _, err := s.repo.GetChangeByID(changeID); err != nil {
		if errors.Is(err, repository.ErrChangeNotFound) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}
	return s.repo.ListApprovals(changeID)
}

func (s *changeServiceImpl) DeleteReviewer(approvalID int64, userID string) error {
	approval, err := s.repo.GetApprovalByID(approvalID)
	if err != nil {
		if errors.Is(err, repository.ErrApprovalNotFound) {
			return ErrApprovalNotFound
		}
		return err
	}

	change, err := s.repo.GetChangeByID(approval.ChangeID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteReviewer(approvalID); err != nil {
		return err
	}

	s.logAudit(userID, change.WorkspaceID, "change_approval.reviewer_removed", fmt.Sprintf("%d", approvalID), fmt.Sprintf("Removed reviewer %s from change %s", approval.ReviewerID, change.ChangeNumber))
	return nil
}

func (s *changeServiceImpl) LinkTask(changeID int64, req model.LinkChangeTaskRequest) (*model.ChangeTaskLink, error) {
	return nil, nil
}

func (s *changeServiceImpl) UnlinkTask(changeID int64, taskID int) error {
	return nil
}

func (s *changeServiceImpl) ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error) {
	return nil, nil
}

func (s *changeServiceImpl) SubmitPIR(changeID int64, reviewerID string, req model.ChangePIRRequest) (*model.ChangePIR, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetPIR(changeID int64) (*model.ChangePIR, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error) {
	return nil, nil
}
