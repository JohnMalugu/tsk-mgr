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

// AutomationAction represents a mutation executed when conditions match
type AutomationAction struct {
	Type        string `json:"type"`                  // ActionSetStatus, ActionSetPriority, etc.
	TargetValue string `json:"target_value"`          // new status, new priority, tag name, assignee ID, etc.
	ExtraInfo   string `json:"extra_info,omitempty"`  // additional metadata e.g. subtask title
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
