package service

import (
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

// Helper methods for business hours, deadline calculations, and rule validations

func isBusinessHour(t time.Time) bool {
	weekday := t.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	hour := t.Hour()
	return hour >= 9 && hour < 17
}

func addMinutes(startTime time.Time, minutes int, businessOnly bool) time.Time {
	if minutes <= 0 {
		return startTime
	}
	if !businessOnly {
		return startTime.Add(time.Duration(minutes) * time.Minute)
	}

	curr := startTime
	remaining := minutes
	for remaining > 0 {
		curr = curr.Add(time.Minute)
		if isBusinessHour(curr) {
			remaining--
		}
	}
	return curr
}

func calculateDeadlines(policy *model.SlaPolicy, createdAt time.Time) (time.Time, time.Time) {
	if policy == nil {
		return createdAt, createdAt
	}
	respDeadline := addMinutes(createdAt, policy.TargetResponseMinutes, policy.BusinessHoursOnly)
	resDeadline := addMinutes(createdAt, policy.TargetResolutionMinutes, policy.BusinessHoursOnly)
	return respDeadline, resDeadline
}

func validateEscalationRules(rules []model.EscalationRuleDefinition) error {
	for _, r := range rules {
		if r.TriggerEvent != model.SlaTriggerAtRisk && r.TriggerEvent != model.SlaTriggerBreached {
			return ErrInvalidTriggerEvent
		}
		if r.ActionType != model.SlaActionEscalatePriority &&
			r.ActionType != model.SlaActionReassign &&
			r.ActionType != model.SlaActionNotify {
			return ErrInvalidSlaActionType
		}
		if r.ThresholdPercentage < 0 || r.ThresholdPercentage > 100 {
			return ErrInvalidEscalationRule
		}
	}
	return nil
}

var (
	ErrSlaPolicyNotFound      = errors.New("sla policy not found")
	ErrSlaPolicyNameRequired  = errors.New("sla policy name is required")
	ErrSlaWorkspaceRequired   = errors.New("workspace ID is required")
	ErrInvalidTargetMinutes   = errors.New("target response and resolution minutes must be positive")
	ErrTaskSlaNotFound        = errors.New("task sla record not found")
	ErrSlaAlreadyPaused       = errors.New("sla timer is already paused")
	ErrSlaNotPaused           = errors.New("sla timer is not paused")
	ErrInvalidEscalationRule  = errors.New("invalid escalation rule configuration")
	ErrInvalidTriggerEvent    = errors.New("invalid escalation trigger event")
	ErrInvalidSlaActionType   = errors.New("invalid escalation action type")
)

// SlaService orchestrates SLA policy definitions, target deadlines, automated escalations, and compliance metrics
type SlaService interface {
	CreatePolicy(userID string, req model.CreateSlaPolicyRequest) (*model.SlaPolicy, error)
	GetPolicy(id int64) (*model.SlaPolicy, error)
	ListPolicies(filter model.SlaFilter) ([]*model.SlaSummary, error)
	UpdatePolicy(id int64, req model.UpdateSlaPolicyRequest) (*model.SlaPolicy, error)
	TogglePolicy(id int64, isActive bool) (*model.SlaPolicy, error)
	DeletePolicy(id int64) error

	SetEscalationRules(policyID int64, req model.SetEscalationRulesRequest) ([]model.EscalationRule, error)
	GetEscalationRules(policyID int64) ([]model.EscalationRule, error)
	DeleteEscalationRule(ruleID int64) error

	ApplySlaToTask(req model.ApplySlaRequest) (*model.TaskSlaRecord, error)
	GetTaskSla(taskID int) (*model.TaskSlaRecord, error)
	PauseSla(taskID int, req model.PauseSlaRequest) (*model.TaskSlaRecord, error)
	ResumeSla(taskID int) (*model.TaskSlaRecord, error)

	EvaluateTaskSla(taskID int) (*model.TaskSlaRecord, error)
	ListBreaches(filter model.BreachFilter) ([]*model.SlaBreachLog, error)
	GetSlaMetrics(workspaceID int64) (*model.SlaMetrics, error)
}

type slaServiceImpl struct {
	repo     repository.SlaRepository
	taskRepo repository.TaskRepository
}

// NewSlaService instantiates a new SlaService
func NewSlaService(repo repository.SlaRepository, taskRepo repository.TaskRepository) SlaService {
	return &slaServiceImpl{
		repo:     repo,
		taskRepo: taskRepo,
	}
}

func (s *slaServiceImpl) CreatePolicy(userID string, req model.CreateSlaPolicyRequest) (*model.SlaPolicy, error) {
	if req.Name == "" {
		return nil, ErrSlaPolicyNameRequired
	}
	if req.WorkspaceID <= 0 {
		return nil, ErrSlaWorkspaceRequired
	}
	if req.TargetResponseMinutes <= 0 || req.TargetResolutionMinutes <= 0 {
		return nil, ErrInvalidTargetMinutes
	}
	if len(req.EscalationRules) > 0 {
		if err := validateEscalationRules(req.EscalationRules); err != nil {
			return nil, err
		}
	}

	tiers := req.PriorityTiers
	if len(tiers) == 0 {
		tiers = []string{"urgent", "high", "medium", "low"}
	}
	bizHours := false
	if req.BusinessHoursOnly != nil {
		bizHours = *req.BusinessHoursOnly
	}
	isDef := false
	if req.IsDefault != nil {
		isDef = *req.IsDefault
	}

	policy := &model.SlaPolicy{
		WorkspaceID:             req.WorkspaceID,
		Name:                    req.Name,
		Description:             req.Description,
		PriorityTiers:           tiers,
		TargetResponseMinutes:   req.TargetResponseMinutes,
		TargetResolutionMinutes: req.TargetResolutionMinutes,
		BusinessHoursOnly:       bizHours,
		IsDefault:               isDef,
		IsActive:                true,
		CreatedByUserID:         userID,
	}

	id, err := s.repo.CreatePolicy(policy)
	if err != nil {
		return nil, err
	}
	policy.ID = id

	if len(req.EscalationRules) > 0 {
		rules := make([]model.EscalationRule, len(req.EscalationRules))
		for i, def := range req.EscalationRules {
			rules[i] = model.EscalationRule{
				PolicyID:            policy.ID,
				TriggerEvent:        def.TriggerEvent,
				ThresholdPercentage: def.ThresholdPercentage,
				ActionType:          def.ActionType,
				TargetUserID:        def.TargetUserID,
				TargetPriority:      def.TargetPriority,
			}
		}
		if err := s.repo.SetEscalationRules(policy.ID, rules); err != nil {
			return nil, err
		}
	}

	return policy, nil
}

func (s *slaServiceImpl) GetPolicy(id int64) (*model.SlaPolicy, error) {
	p, err := s.repo.GetPolicyByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrSlaPolicyNotFound) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}
	return p, nil
}

func (s *slaServiceImpl) ListPolicies(filter model.SlaFilter) ([]*model.SlaSummary, error) {
	return s.repo.ListPolicies(filter)
}

func (s *slaServiceImpl) UpdatePolicy(id int64, req model.UpdateSlaPolicyRequest) (*model.SlaPolicy, error) {
	return nil, nil
}

func (s *slaServiceImpl) TogglePolicy(id int64, isActive bool) (*model.SlaPolicy, error) {
	return nil, nil
}

func (s *slaServiceImpl) DeletePolicy(id int64) error {
	return nil
}

func (s *slaServiceImpl) SetEscalationRules(policyID int64, req model.SetEscalationRulesRequest) ([]model.EscalationRule, error) {
	return nil, nil
}

func (s *slaServiceImpl) GetEscalationRules(policyID int64) ([]model.EscalationRule, error) {
	return nil, nil
}

func (s *slaServiceImpl) DeleteEscalationRule(ruleID int64) error {
	return nil
}

func (s *slaServiceImpl) ApplySlaToTask(req model.ApplySlaRequest) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (s *slaServiceImpl) GetTaskSla(taskID int) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (s *slaServiceImpl) PauseSla(taskID int, req model.PauseSlaRequest) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (s *slaServiceImpl) ResumeSla(taskID int) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (s *slaServiceImpl) EvaluateTaskSla(taskID int) (*model.TaskSlaRecord, error) {
	return nil, nil
}

func (s *slaServiceImpl) ListBreaches(filter model.BreachFilter) ([]*model.SlaBreachLog, error) {
	return nil, nil
}

func (s *slaServiceImpl) GetSlaMetrics(workspaceID int64) (*model.SlaMetrics, error) {
	return nil, nil
}
