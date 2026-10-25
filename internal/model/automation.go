package model

import "time"

// Supported automation triggers
const (
	TriggerTaskCreated         = "task_created"
	TriggerTaskStatusChanged   = "task_status_changed"
	TriggerTaskPriorityChanged = "task_priority_changed"
	TriggerTaskAssigned        = "task_assigned"
	TriggerTaskCompleted       = "task_completed"
)

// Supported condition operators
const (
	OpEquals    = "equals"
	OpNotEquals = "not_equals"
	OpContains  = "contains"
	OpIn        = "in"
)

// Supported condition field targets
const (
	FieldStatus      = "status"
	FieldPriority    = "priority"
	FieldTitle       = "title"
	FieldDescription = "description"
	FieldProjectID   = "project_id"
	FieldAssignee    = "assignee"
	FieldTags        = "tags"
)


// Supported automation actions
const (
	ActionSetStatus      = "set_status"
	ActionSetPriority    = "set_priority"
	ActionAddTag         = "add_tag"
	ActionRemoveTag      = "remove_tag"
	ActionAssignUser     = "assign_user"
	ActionCreateSubtask  = "create_subtask"
)

// AutomationCondition represents a rule predicate evaluated against task attributes
type AutomationCondition struct {
	Field    string   `json:"field"`              // e.g. "status", "priority", "title", "tags", "project_id"
	Operator string   `json:"operator"`           // equals, not_equals, contains, in
	Value    string   `json:"value,omitempty"`    // comparison value
	Values   []string `json:"values,omitempty"`   // comparison values for "in" operator
}

// IsValid validates whether the condition has supported fields and operators
func (c *AutomationCondition) IsValid() bool {
	if c.Field == "" || c.Operator == "" {
		return false
	}
	switch c.Operator {
	case OpEquals, OpNotEquals, OpContains:
		return c.Value != ""
	case OpIn:
		return len(c.Values) > 0
	default:
		return false
	}
}

// AutomationAction represents a mutation executed when conditions match
type AutomationAction struct {
	Type        string `json:"type"`                  // ActionSetStatus, ActionSetPriority, etc.
	TargetValue string `json:"target_value"`          // new status, new priority, tag name, assignee ID, etc.
	ExtraInfo   string `json:"extra_info,omitempty"`  // additional metadata e.g. subtask title
}

// IsValid checks if action contains valid type and non-empty target
func (a *AutomationAction) IsValid() bool {
	if a.Type == "" || a.TargetValue == "" {
		return false
	}
	switch a.Type {
	case ActionSetStatus, ActionSetPriority, ActionAddTag, ActionRemoveTag, ActionAssignUser, ActionCreateSubtask:
		return true
	default:
		return false
	}
}


// AutomationRule represents a configurable reactive workflow rule
type AutomationRule struct {
	ID              int64                 `json:"id"`
	WorkspaceID     int64                 `json:"workspace_id"`
	CreatedByUserID string                `json:"created_by_user_id"`
	Name            string                `json:"name"`
	Description     string                `json:"description"`
	TriggerType     string                `json:"trigger_type"`
	Conditions      []AutomationCondition `json:"conditions"`
	Actions         []AutomationAction    `json:"actions"`
	IsActive        bool                  `json:"is_active"`
	ExecutionCount  int                   `json:"execution_count"`
	LastTriggeredAt *time.Time            `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// Execution status constants
const (
	ExecutionStatusSuccess = "success"
	ExecutionStatusSkipped = "skipped"
	ExecutionStatusFailed  = "failed"
)

// TriggerEvent encapsulates an event payload dispatched to evaluate automation rules
type TriggerEvent struct {
	WorkspaceID int64     `json:"workspace_id"`
	UserID      string    `json:"user_id"`
	TaskID      int       `json:"task_id"`
	EventType   string    `json:"event_type"`
	TaskBefore  *Task     `json:"task_before,omitempty"`
	TaskAfter   *Task     `json:"task_after,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// AutomationExecutionLog records the execution trial, condition evaluation, and actions performed
type AutomationExecutionLog struct {
	ID              int64     `json:"id"`
	RuleID          int64     `json:"rule_id"`
	WorkspaceID     int64     `json:"workspace_id"`
	TriggerEvent    string    `json:"trigger_event"`
	Status          string    `json:"status"` // success, skipped, failed
	ActionsExecuted []string  `json:"actions_executed"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	DurationMs      int64     `json:"duration_ms"`
	ExecutedAt      time.Time `json:"executed_at"`
}

// CreateAutomationRuleRequest defines payloads for registering a new automation rule
type CreateAutomationRuleRequest struct {
	WorkspaceID int64                 `json:"workspace_id"`
	Name        string                `json:"name"`
	Description string                `json:"description,omitempty"`
	TriggerType string                `json:"trigger_type"`
	Conditions  []AutomationCondition `json:"conditions"`
	Actions     []AutomationAction    `json:"actions"`
	IsActive    *bool                 `json:"is_active,omitempty"`
}

// UpdateAutomationRuleRequest defines parameters for updating an existing automation rule
type UpdateAutomationRuleRequest struct {
	Name        *string                `json:"name,omitempty"`
	Description *string                `json:"description,omitempty"`
	TriggerType *string                `json:"trigger_type,omitempty"`
	Conditions  *[]AutomationCondition `json:"conditions,omitempty"`
	Actions     *[]AutomationAction    `json:"actions,omitempty"`
	IsActive    *bool                  `json:"is_active,omitempty"`
}

// AutomationDryRunRequest defines parameters for testing an automation rule without saving mutations
type AutomationDryRunRequest struct {
	TaskID        *int   `json:"task_id,omitempty"`
	MockTask      *Task  `json:"mock_task,omitempty"`
	EventOverride string `json:"event_override,omitempty"`
}

// ConditionEvaluationDetail describes single condition matching outcome during simulation
type ConditionEvaluationDetail struct {
	Condition   AutomationCondition `json:"condition"`
	Matched     bool                `json:"matched"`
	ActualValue string              `json:"actual_value"`
	Reason      string              `json:"reason"`
}

// ActionPreviewDetail describes action behavior during simulation
type ActionPreviewDetail struct {
	Action            AutomationAction `json:"action"`
	Description       string           `json:"description"`
	EstimatedMutation string           `json:"estimated_mutation"`
}

// AutomationDryRunResult encapsulates dry run simulation findings
type AutomationDryRunResult struct {
	RuleID               int64                       `json:"rule_id"`
	RuleName             string                      `json:"rule_name"`
	ConditionsMet        bool                        `json:"conditions_met"`
	ConditionEvaluations []ConditionEvaluationDetail `json:"condition_evaluations"`
	ActionsPreview       []ActionPreviewDetail       `json:"actions_preview"`
	EvaluatedAt          time.Time                   `json:"evaluated_at"`
}

// AutomationLogFilter contains criteria to filter execution logs
type AutomationLogFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	RuleID      *int64 `json:"rule_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// AutomationRuleStat aggregates execution numbers per rule
type AutomationRuleStat struct {
	RuleID         int64      `json:"rule_id"`
	RuleName       string     `json:"rule_name"`
	ExecutionCount int        `json:"execution_count"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
}

// AutomationMetrics summarizes rule performance and execution outcomes for a workspace
type AutomationMetrics struct {
	WorkspaceID          int64                `json:"workspace_id"`
	TotalRules           int                  `json:"total_rules"`
	ActiveRules          int                  `json:"active_rules"`
	TotalExecutions      int                  `json:"total_executions"`
	SuccessfulExecutions int                  `json:"successful_executions"`
	FailedExecutions     int                  `json:"failed_executions"`
	SkippedExecutions    int                  `json:"skipped_executions"`
	TopRules             []AutomationRuleStat `json:"top_rules,omitempty"`
}

// AutomationRuleSummary provides lightweight overview metadata for rule catalogs and sidebars
type AutomationRuleSummary struct {
	ID              int64      `json:"id"`
	WorkspaceID     int64      `json:"workspace_id"`
	Name            string     `json:"name"`
	TriggerType     string     `json:"trigger_type"`
	ConditionCount  int        `json:"condition_count"`
	ActionCount     int        `json:"action_count"`
	IsActive        bool       `json:"is_active"`
	ExecutionCount  int        `json:"execution_count"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// BatchToggleRuleRequest enables turning multiple automation rules on or off simultaneously
type BatchToggleRuleRequest struct {
	RuleIDs  []int64 `json:"rule_ids"`
	IsActive bool    `json:"is_active"`
}

// DuplicateRuleRequest defines optional overrides when cloning an automation rule
type DuplicateRuleRequest struct {
	NewName           string `json:"new_name,omitempty"`
	TargetWorkspaceID *int64 `json:"target_workspace_id,omitempty"`
}





