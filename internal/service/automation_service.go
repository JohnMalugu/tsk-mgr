package service

import (
	"errors"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

// Domain errors for automation rules
var (
	ErrAutomationRuleNotFound   = errors.New("automation rule not found")
	ErrRuleNameRequired         = errors.New("rule name is required and must not exceed 100 characters")
	ErrInvalidTriggerType       = errors.New("invalid trigger type")
	ErrInvalidActionType        = errors.New("invalid action type")
	ErrInvalidConditionField    = errors.New("invalid condition field")
	ErrInvalidConditionOperator = errors.New("invalid condition operator")
	ErrEmptyActions             = errors.New("automation rule must have at least one action")
)

var allowedTriggers = map[string]bool{
	model.TriggerTaskCreated:         true,
	model.TriggerTaskStatusChanged:   true,
	model.TriggerTaskPriorityChanged: true,
	model.TriggerTaskAssigned:        true,
	model.TriggerTaskCompleted:       true,
}

var allowedConditionFields = map[string]bool{
	model.FieldStatus:      true,
	model.FieldPriority:    true,
	model.FieldTitle:       true,
	model.FieldDescription: true,
	model.FieldProjectID:   true,
	model.FieldAssignee:    true,
	model.FieldTags:        true,
}

var allowedConditionOperators = map[string]bool{
	model.OpEquals:    true,
	model.OpNotEquals: true,
	model.OpContains:  true,
	model.OpIn:        true,
}

var allowedActionTypes = map[string]bool{
	model.ActionSetStatus:     true,
	model.ActionSetPriority:   true,
	model.ActionAddTag:        true,
	model.ActionRemoveTag:     true,
	model.ActionAssignUser:    true,
	model.ActionCreateSubtask: true,
}

func validateRuleName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || len(trimmed) > 100 {
		return "", ErrRuleNameRequired
	}
	return trimmed, nil
}

func validateTriggerType(t string) (string, error) {
	trimmed := strings.ToLower(strings.TrimSpace(t))
	if !allowedTriggers[trimmed] {
		return "", ErrInvalidTriggerType
	}
	return trimmed, nil
}

func validateConditions(conditions []model.AutomationCondition) ([]model.AutomationCondition, error) {
	var validated []model.AutomationCondition
	for _, c := range conditions {
		field := strings.ToLower(strings.TrimSpace(c.Field))
		if !allowedConditionFields[field] {
			return nil, ErrInvalidConditionField
		}
		op := strings.ToLower(strings.TrimSpace(c.Operator))
		if !allowedConditionOperators[op] {
			return nil, ErrInvalidConditionOperator
		}
		val := strings.TrimSpace(c.Value)
		var values []string
		for _, v := range c.Values {
			vt := strings.TrimSpace(v)
			if vt != "" {
				values = append(values, vt)
			}
		}
		if op == model.OpIn && len(values) == 0 {
			return nil, ErrInvalidConditionOperator
		}
		if op != model.OpIn && val == "" {
			return nil, ErrInvalidConditionOperator
		}
		validated = append(validated, model.AutomationCondition{
			Field:    field,
			Operator: op,
			Value:    val,
			Values:   values,
		})
	}
	return validated, nil
}

func validateActions(actions []model.AutomationAction) ([]model.AutomationAction, error) {
	if len(actions) == 0 {
		return nil, ErrEmptyActions
	}
	var validated []model.AutomationAction
	for _, a := range actions {
		actType := strings.ToLower(strings.TrimSpace(a.Type))
		if !allowedActionTypes[actType] {
			return nil, ErrInvalidActionType
		}
		target := strings.TrimSpace(a.TargetValue)
		if target == "" {
			return nil, ErrInvalidActionType
		}
		validated = append(validated, model.AutomationAction{
			Type:        actType,
			TargetValue: target,
			ExtraInfo:   strings.TrimSpace(a.ExtraInfo),
		})
	}
	return validated, nil
}


// AutomationService coordinates workflow rules, condition evaluation, and action execution
type AutomationService interface {
	CreateRule(userID string, req model.CreateAutomationRuleRequest) (*model.AutomationRule, error)
	GetRule(id int64) (*model.AutomationRule, error)
	ListRules(workspaceID int64, activeOnly *bool, triggerType string) ([]*model.AutomationRule, error)
	UpdateRule(id int64, req model.UpdateAutomationRuleRequest) (*model.AutomationRule, error)
	DeleteRule(id int64) error
	ToggleRule(id int64) (bool, error)
	DuplicateRule(userID string, id int64, req model.DuplicateRuleRequest) (*model.AutomationRule, error)
	DryRunRule(id int64, req model.AutomationDryRunRequest) (*model.AutomationDryRunResult, error)
	DispatchEvent(event model.TriggerEvent) ([]*model.AutomationExecutionLog, error)
	ListLogs(filter model.AutomationLogFilter) ([]*model.AutomationExecutionLog, error)
	GetMetrics(workspaceID int64) (*model.AutomationMetrics, error)
}

type automationService struct {
	repo             repository.AutomationRepository
	taskRepo         repository.TaskRepository
	subtaskRepo      repository.SubtaskRepository
	notificationRepo repository.NotificationRepository
}

// NewAutomationService constructs a new AutomationService instance
func NewAutomationService(
	repo repository.AutomationRepository,
	taskRepo repository.TaskRepository,
	subtaskRepo repository.SubtaskRepository,
	notificationRepo repository.NotificationRepository,
) AutomationService {
	return &automationService{
		repo:             repo,
		taskRepo:         taskRepo,
		subtaskRepo:      subtaskRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *automationService) CreateRule(userID string, req model.CreateAutomationRuleRequest) (*model.AutomationRule, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	name, err := validateRuleName(req.Name)
	if err != nil {
		return nil, err
	}

	trigger, err := validateTriggerType(req.TriggerType)
	if err != nil {
		return nil, err
	}

	conditions, err := validateConditions(req.Conditions)
	if err != nil {
		return nil, err
	}

	actions, err := validateActions(req.Actions)
	if err != nil {
		return nil, err
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rule := &model.AutomationRule{
		WorkspaceID:     req.WorkspaceID,
		CreatedByUserID: userID,
		Name:            name,
		Description:     strings.TrimSpace(req.Description),
		TriggerType:     trigger,
		Conditions:      conditions,
		Actions:         actions,
		IsActive:        isActive,
	}

	id, err := s.repo.CreateRule(rule)
	if err != nil {
		return nil, err
	}

	return s.repo.GetRuleByID(id)
}

func (s *automationService) GetRule(id int64) (*model.AutomationRule, error) {
	rule, err := s.repo.GetRuleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAutomationRuleNotFound) {
			return nil, ErrAutomationRuleNotFound
		}
		return nil, err
	}
	return rule, nil
}

func (s *automationService) ListRules(workspaceID int64, activeOnly *bool, triggerType string) ([]*model.AutomationRule, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListRules(workspaceID, activeOnly, triggerType)
}


func (s *automationService) UpdateRule(id int64, req model.UpdateAutomationRuleRequest) (*model.AutomationRule, error) {
	return nil, nil
}

func (s *automationService) DeleteRule(id int64) error {
	return nil
}

func (s *automationService) ToggleRule(id int64) (bool, error) {
	return false, nil
}

func (s *automationService) DuplicateRule(userID string, id int64, req model.DuplicateRuleRequest) (*model.AutomationRule, error) {
	return nil, nil
}

func (s *automationService) DryRunRule(id int64, req model.AutomationDryRunRequest) (*model.AutomationDryRunResult, error) {
	return nil, nil
}

func (s *automationService) DispatchEvent(event model.TriggerEvent) ([]*model.AutomationExecutionLog, error) {
	return nil, nil
}

func (s *automationService) ListLogs(filter model.AutomationLogFilter) ([]*model.AutomationExecutionLog, error) {
	return nil, nil
}

func (s *automationService) GetMetrics(workspaceID int64) (*model.AutomationMetrics, error) {
	return nil, nil
}
