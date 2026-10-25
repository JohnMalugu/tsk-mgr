package service

import (
	"errors"
	"fmt"
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

func evaluateCondition(task *model.Task, c model.AutomationCondition) model.ConditionEvaluationDetail {
	if task == nil {
		return model.ConditionEvaluationDetail{
			Condition: c,
			Matched:   false,
			Reason:    "task is nil",
		}
	}

	var actualStr string
	var actualList []string

	switch c.Field {
	case model.FieldStatus:
		actualStr = task.Status
		actualList = []string{task.Status}
	case model.FieldPriority:
		actualStr = task.Priority
		actualList = []string{task.Priority}
	case model.FieldTitle:
		actualStr = task.Title
		actualList = []string{task.Title}
	case model.FieldDescription:
		actualStr = task.Description
		actualList = []string{task.Description}
	case model.FieldProjectID:
		if task.ProjectID != nil {
			actualStr = fmt.Sprintf("%d", *task.ProjectID)
			actualList = []string{actualStr}
		}
	case model.FieldAssignee:
		actualStr = task.UserID
		actualList = []string{task.UserID}
	case model.FieldTags:
		actualList = task.Tags
		actualStr = strings.Join(task.Tags, ", ")
	}

	matched := false
	reason := ""

	switch c.Operator {
	case model.OpEquals:
		target := strings.ToLower(c.Value)
		if c.Field == model.FieldTags {
			for _, tag := range actualList {
				if strings.ToLower(tag) == target {
					matched = true
					break
				}
			}
		} else {
			matched = strings.ToLower(actualStr) == target
		}
		if matched {
			reason = fmt.Sprintf("field '%s' matches '%s'", c.Field, c.Value)
		} else {
			reason = fmt.Sprintf("field '%s' value '%s' does not equal '%s'", c.Field, actualStr, c.Value)
		}

	case model.OpNotEquals:
		target := strings.ToLower(c.Value)
		if c.Field == model.FieldTags {
			hasTag := false
			for _, tag := range actualList {
				if strings.ToLower(tag) == target {
					hasTag = true
					break
				}
			}
			matched = !hasTag
		} else {
			matched = strings.ToLower(actualStr) != target
		}
		if matched {
			reason = fmt.Sprintf("field '%s' value '%s' is not '%s'", c.Field, actualStr, c.Value)
		} else {
			reason = fmt.Sprintf("field '%s' equals '%s'", c.Field, c.Value)
		}

	case model.OpContains:
		target := strings.ToLower(c.Value)
		if c.Field == model.FieldTags {
			for _, tag := range actualList {
				if strings.Contains(strings.ToLower(tag), target) {
					matched = true
					break
				}
			}
		} else {
			matched = strings.Contains(strings.ToLower(actualStr), target)
		}
		if matched {
			reason = fmt.Sprintf("field '%s' contains '%s'", c.Field, c.Value)
		} else {
			reason = fmt.Sprintf("field '%s' ('%s') does not contain '%s'", c.Field, actualStr, c.Value)
		}

	case model.OpIn:
		targetMap := make(map[string]bool)
		for _, v := range c.Values {
			targetMap[strings.ToLower(v)] = true
		}
		if c.Field == model.FieldTags {
			for _, tag := range actualList {
				if targetMap[strings.ToLower(tag)] {
					matched = true
					break
				}
			}
		} else {
			matched = targetMap[strings.ToLower(actualStr)]
		}
		if matched {
			reason = fmt.Sprintf("field '%s' value '%s' is present in %v", c.Field, actualStr, c.Values)
		} else {
			reason = fmt.Sprintf("field '%s' value '%s' is not in %v", c.Field, actualStr, c.Values)
		}
	}

	return model.ConditionEvaluationDetail{
		Condition:   c,
		Matched:     matched,
		ActualValue: actualStr,
		Reason:      reason,
	}
}

func evaluateAllConditions(task *model.Task, conditions []model.AutomationCondition) (bool, []model.ConditionEvaluationDetail) {
	if len(conditions) == 0 {
		return true, nil
	}
	var details []model.ConditionEvaluationDetail
	allMatched := true
	for _, c := range conditions {
		detail := evaluateCondition(task, c)
		details = append(details, detail)
		if !detail.Matched {
			allMatched = false
		}
	}
	return allMatched, details
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
	rule, err := s.repo.GetRuleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAutomationRuleNotFound) {
			return nil, ErrAutomationRuleNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		name, err := validateRuleName(*req.Name)
		if err != nil {
			return nil, err
		}
		rule.Name = name
	}

	if req.Description != nil {
		rule.Description = strings.TrimSpace(*req.Description)
	}

	if req.TriggerType != nil {
		trigger, err := validateTriggerType(*req.TriggerType)
		if err != nil {
			return nil, err
		}
		rule.TriggerType = trigger
	}

	if req.Conditions != nil {
		conditions, err := validateConditions(*req.Conditions)
		if err != nil {
			return nil, err
		}
		rule.Conditions = conditions
	}

	if req.Actions != nil {
		actions, err := validateActions(*req.Actions)
		if err != nil {
			return nil, err
		}
		rule.Actions = actions
	}

	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateRule(rule); err != nil {
		return nil, err
	}

	return s.repo.GetRuleByID(id)
}

func (s *automationService) DeleteRule(id int64) error {
	_, err := s.repo.GetRuleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAutomationRuleNotFound) {
			return ErrAutomationRuleNotFound
		}
		return err
	}
	return s.repo.DeleteRule(id)
}

func (s *automationService) ToggleRule(id int64) (bool, error) {
	_, err := s.repo.GetRuleByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrAutomationRuleNotFound) {
			return false, ErrAutomationRuleNotFound
		}
		return false, err
	}
	return s.repo.ToggleRuleActive(id)
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
