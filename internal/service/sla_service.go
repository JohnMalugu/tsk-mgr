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
	policy, err := s.repo.GetPolicyByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrSlaPolicyNotFound) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		if *req.Name == "" {
			return nil, ErrSlaPolicyNameRequired
		}
		policy.Name = *req.Name
	}
	if req.Description != nil {
		policy.Description = *req.Description
	}
	if req.PriorityTiers != nil {
		policy.PriorityTiers = *req.PriorityTiers
	}
	if req.TargetResponseMinutes != nil {
		if *req.TargetResponseMinutes <= 0 {
			return nil, ErrInvalidTargetMinutes
		}
		policy.TargetResponseMinutes = *req.TargetResponseMinutes
	}
	if req.TargetResolutionMinutes != nil {
		if *req.TargetResolutionMinutes <= 0 {
			return nil, ErrInvalidTargetMinutes
		}
		policy.TargetResolutionMinutes = *req.TargetResolutionMinutes
	}
	if req.BusinessHoursOnly != nil {
		policy.BusinessHoursOnly = *req.BusinessHoursOnly
	}
	if req.IsDefault != nil {
		policy.IsDefault = *req.IsDefault
	}
	if req.IsActive != nil {
		policy.IsActive = *req.IsActive
	}

	if err := s.repo.UpdatePolicy(policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *slaServiceImpl) TogglePolicy(id int64, isActive bool) (*model.SlaPolicy, error) {
	if err := s.repo.TogglePolicyActive(id, isActive); err != nil {
		if errors.Is(err, repository.ErrSlaPolicyNotFound) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}
	return s.GetPolicy(id)
}

func (s *slaServiceImpl) DeletePolicy(id int64) error {
	err := s.repo.DeletePolicy(id)
	if err != nil && errors.Is(err, repository.ErrSlaPolicyNotFound) {
		return ErrSlaPolicyNotFound
	}
	return err
}

func (s *slaServiceImpl) SetEscalationRules(policyID int64, req model.SetEscalationRulesRequest) ([]model.EscalationRule, error) {
	if _, err := s.repo.GetPolicyByID(policyID); err != nil {
		if errors.Is(err, repository.ErrSlaPolicyNotFound) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}
	if err := validateEscalationRules(req.Rules); err != nil {
		return nil, err
	}

	rules := make([]model.EscalationRule, len(req.Rules))
	for i, def := range req.Rules {
		rules[i] = model.EscalationRule{
			PolicyID:            policyID,
			TriggerEvent:        def.TriggerEvent,
			ThresholdPercentage: def.ThresholdPercentage,
			ActionType:          def.ActionType,
			TargetUserID:        def.TargetUserID,
			TargetPriority:      def.TargetPriority,
		}
	}

	if err := s.repo.SetEscalationRules(policyID, rules); err != nil {
		return nil, err
	}
	return s.repo.GetEscalationRules(policyID)
}

func (s *slaServiceImpl) GetEscalationRules(policyID int64) ([]model.EscalationRule, error) {
	if _, err := s.repo.GetPolicyByID(policyID); err != nil {
		if errors.Is(err, repository.ErrSlaPolicyNotFound) {
			return nil, ErrSlaPolicyNotFound
		}
		return nil, err
	}
	return s.repo.GetEscalationRules(policyID)
}

func (s *slaServiceImpl) DeleteEscalationRule(ruleID int64) error {
	return s.repo.DeleteEscalationRule(ruleID)
}

func (s *slaServiceImpl) ApplySlaToTask(req model.ApplySlaRequest) (*model.TaskSlaRecord, error) {
	task, err := s.taskRepo.GetByID(req.TaskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrTaskNotFound
	}

	var policy *model.SlaPolicy
	if req.PolicyID != nil && *req.PolicyID > 0 {
		policy, err = s.repo.GetPolicyByID(*req.PolicyID)
		if err != nil {
			if errors.Is(err, repository.ErrSlaPolicyNotFound) {
				return nil, ErrSlaPolicyNotFound
			}
			return nil, err
		}
	} else {
		// Default to workspace 1 if not specified
		var wsID int64 = 1
		policy, err = s.repo.GetDefaultPolicy(wsID, task.Priority)
		if err != nil || policy == nil {
			return nil, ErrSlaPolicyNotFound
		}
	}

	now := time.Now()
	createdAt := task.CreatedAt
	if createdAt.IsZero() {
		createdAt = now
	}
	respDeadline, resDeadline := calculateDeadlines(policy, createdAt)

	rec := &model.TaskSlaRecord{
		TaskID:             task.ID,
		PolicyID:           policy.ID,
		WorkspaceID:        policy.WorkspaceID,
		ResponseDeadline:   respDeadline,
		ResolutionDeadline: resDeadline,
		ResponseStatus:     model.SlaStagePending,
		ResolutionStatus:   model.SlaStagePending,
		IsPaused:           false,
		TotalPausedMinutes: 0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if task.Completed || task.Status == model.TaskStatusCompleted {
		resolvedAt := now
		rec.ResolvedAt = &resolvedAt
		if now.Before(resDeadline) || now.Equal(resDeadline) {
			rec.ResolutionStatus = model.SlaStageMet
		} else {
			rec.ResolutionStatus = model.SlaStageBreached
		}
	}

	if task.Status != model.TaskStatusTodo {
		respondedAt := now
		rec.RespondedAt = &respondedAt
		rec.ResponseStatus = model.SlaStageMet
	}

	id, err := s.repo.CreateTaskSlaRecord(rec)
	if err != nil {
		return nil, err
	}
	rec.ID = id
	return rec, nil
}

func (s *slaServiceImpl) GetTaskSla(taskID int) (*model.TaskSlaRecord, error) {
	rec, err := s.repo.GetTaskSlaRecordByTaskID(taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskSlaNotFound) {
			return nil, ErrTaskSlaNotFound
		}
		return nil, err
	}
	return rec, nil
}

func (s *slaServiceImpl) PauseSla(taskID int, req model.PauseSlaRequest) (*model.TaskSlaRecord, error) {
	rec, err := s.repo.GetTaskSlaRecordByTaskID(taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskSlaNotFound) {
			return nil, ErrTaskSlaNotFound
		}
		return nil, err
	}
	if rec.IsPaused {
		return nil, ErrSlaAlreadyPaused
	}

	now := time.Now()
	rec.IsPaused = true
	rec.PausedAt = &now

	if err := s.repo.UpdateTaskSlaRecord(rec); err != nil {
		return nil, err
	}
	return rec, nil
}

func (s *slaServiceImpl) ResumeSla(taskID int) (*model.TaskSlaRecord, error) {
	rec, err := s.repo.GetTaskSlaRecordByTaskID(taskID)
	if err != nil {
		if errors.Is(err, repository.ErrTaskSlaNotFound) {
			return nil, ErrTaskSlaNotFound
		}
		return nil, err
	}
	if !rec.IsPaused {
		return nil, ErrSlaNotPaused
	}

	now := time.Now()
	pausedDuration := time.Duration(0)
	if rec.PausedAt != nil {
		pausedDuration = now.Sub(*rec.PausedAt)
		if pausedDuration < 0 {
			pausedDuration = 0
		}
	}

	mins := int(pausedDuration.Minutes())
	rec.TotalPausedMinutes += mins
	rec.ResponseDeadline = rec.ResponseDeadline.Add(pausedDuration)
	rec.ResolutionDeadline = rec.ResolutionDeadline.Add(pausedDuration)
	rec.IsPaused = false
	rec.PausedAt = nil

	if err := s.repo.UpdateTaskSlaRecord(rec); err != nil {
		return nil, err
	}
	return rec, nil
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
