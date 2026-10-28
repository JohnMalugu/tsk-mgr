package model

import "time"

// SLA policy status constants
const (
	SlaStatusActive   = "active"
	SlaStatusInactive = "inactive"
)

// SlaPolicy defines service level targets for response and resolution times by workspace and priority
type SlaPolicy struct {
	ID                     int64     `json:"id"`
	WorkspaceID            int64     `json:"workspace_id"`
	Name                   string    `json:"name"`
	Description            string    `json:"description,omitempty"`
	PriorityTiers          []string  `json:"priority_tiers"`           // e.g. ["urgent", "high"]
	TargetResponseMinutes  int       `json:"target_response_minutes"`   // time to first response/acknowledgment
	TargetResolutionMinutes int      `json:"target_resolution_minutes"` // time to task completion
	BusinessHoursOnly      bool      `json:"business_hours_only"`       // count only business hours
	IsDefault              bool      `json:"is_default"`
	IsActive               bool      `json:"is_active"`
	CreatedByUserID        string    `json:"created_by_user_id"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Escalation trigger event constants
const (
	SlaTriggerAtRisk   = "at_risk"   // deadline threshold reached
	SlaTriggerBreached = "breached"  // deadline exceeded
)

// Escalation action type constants
const (
	SlaActionEscalatePriority = "escalate_priority"
	SlaActionReassign         = "reassign"
	SlaActionNotify           = "notify"
)

// EscalationRule defines automated mitigation when an SLA target is endangered or missed
type EscalationRule struct {
	ID                  int64     `json:"id"`
	PolicyID            int64     `json:"policy_id"`
	TriggerEvent        string    `json:"trigger_event"`        // at_risk, breached
	ThresholdPercentage int       `json:"threshold_percentage"` // e.g. 75, 90 (0 for breached)
	ActionType          string    `json:"action_type"`          // escalate_priority, reassign, notify
	TargetUserID        string    `json:"target_user_id,omitempty"`
	TargetPriority      string    `json:"target_priority,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
}

// SLA deadline status lifecycle constants
const (
	SlaStagePending  = "pending"
	SlaStageMet      = "met"
	SlaStageBreached = "breached"
	SlaStagePaused   = "paused"
)

// TaskSlaRecord tracks live SLA targets, elapsed durations, and fulfillment status for a task
type TaskSlaRecord struct {
	ID                 int64      `json:"id"`
	TaskID             int        `json:"task_id"`
	PolicyID           int64      `json:"policy_id"`
	WorkspaceID        int64      `json:"workspace_id"`
	ResponseDeadline   time.Time  `json:"response_deadline"`
	ResolutionDeadline time.Time  `json:"resolution_deadline"`
	RespondedAt        *time.Time `json:"responded_at,omitempty"`
	ResolvedAt         *time.Time `json:"resolved_at,omitempty"`
	ResponseStatus     string     `json:"response_status"`   // pending, met, breached
	ResolutionStatus   string     `json:"resolution_status"` // pending, met, breached
	IsPaused           bool       `json:"is_paused"`
	PausedAt           *time.Time `json:"paused_at,omitempty"`
	TotalPausedMinutes int        `json:"total_paused_minutes"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Breach type constants
const (
	BreachTypeResponse   = "response"
	BreachTypeResolution = "resolution"
)

// SlaBreachLog records historical breach events and automated escalation actions taken
type SlaBreachLog struct {
	ID                  int64     `json:"id"`
	TaskID              int       `json:"task_id"`
	PolicyID            int64     `json:"policy_id"`
	WorkspaceID         int64     `json:"workspace_id"`
	BreachType          string    `json:"breach_type"` // response, resolution
	BreachedAt          time.Time `json:"breached_at"`
	MinutesOverdue      int       `json:"minutes_overdue"`
	EscalationTriggered bool      `json:"escalation_triggered"`
	EscalationAction    string    `json:"escalation_action,omitempty"`
	Details             string    `json:"details,omitempty"`
}

// EscalationRuleDefinition defines an escalation rule blueprint when configuring policies
type EscalationRuleDefinition struct {
	TriggerEvent        string `json:"trigger_event"`        // at_risk, breached
	ThresholdPercentage int    `json:"threshold_percentage"` // e.g. 75, 90 (0 for breached)
	ActionType          string `json:"action_type"`          // escalate_priority, reassign, notify
	TargetUserID        string `json:"target_user_id,omitempty"`
	TargetPriority      string `json:"target_priority,omitempty"`
}

// CreateSlaPolicyRequest represents payload for creating a new SLA policy
type CreateSlaPolicyRequest struct {
	WorkspaceID             int64                      `json:"workspace_id"`
	Name                    string                     `json:"name"`
	Description             string                     `json:"description,omitempty"`
	PriorityTiers           []string                   `json:"priority_tiers,omitempty"`
	TargetResponseMinutes   int                        `json:"target_response_minutes"`
	TargetResolutionMinutes int                        `json:"target_resolution_minutes"`
	BusinessHoursOnly       *bool                      `json:"business_hours_only,omitempty"`
	IsDefault               *bool                      `json:"is_default,omitempty"`
	EscalationRules         []EscalationRuleDefinition `json:"escalation_rules,omitempty"`
}

// UpdateSlaPolicyRequest represents payload for modifying an existing SLA policy
type UpdateSlaPolicyRequest struct {
	Name                    *string   `json:"name,omitempty"`
	Description             *string   `json:"description,omitempty"`
	PriorityTiers           *[]string `json:"priority_tiers,omitempty"`
	TargetResponseMinutes   *int      `json:"target_response_minutes,omitempty"`
	TargetResolutionMinutes *int      `json:"target_resolution_minutes,omitempty"`
	BusinessHoursOnly       *bool     `json:"business_hours_only,omitempty"`
	IsDefault               *bool     `json:"is_default,omitempty"`
	IsActive                *bool     `json:"is_active,omitempty"`
}

// SetEscalationRulesRequest represents payload for replacing escalation rules of a policy
type SetEscalationRulesRequest struct {
	Rules []EscalationRuleDefinition `json:"rules"`
}

// ApplySlaRequest represents payload for attaching or calculating SLA targets on a task
type ApplySlaRequest struct {
	TaskID   int    `json:"task_id"`
	PolicyID *int64 `json:"policy_id,omitempty"` // if nil, uses default policy for priority
}

// PauseSlaRequest represents payload for pausing an active SLA countdown
type PauseSlaRequest struct {
	Reason string `json:"reason,omitempty"`
}

// SlaMetrics aggregates SLA compliance, breach statistics, and speed metrics
type SlaMetrics struct {
	WorkspaceID             int64          `json:"workspace_id"`
	TotalTrackedTasks       int            `json:"total_tracked_tasks"`
	MetResponseCount        int            `json:"met_response_count"`
	BreachedResponseCount   int            `json:"breached_response_count"`
	MetResolutionCount      int            `json:"met_resolution_count"`
	BreachedResolutionCount int            `json:"breached_resolution_count"`
	ActiveSlaCount          int            `json:"active_sla_count"`
	ComplianceRate          float64        `json:"compliance_rate"` // percentage (0.0 to 100.0)
	AverageResponseMin      float64        `json:"average_response_min"`
	AverageResolutionMin    float64        `json:"average_resolution_min"`
	BreachesByPriority      map[string]int `json:"breaches_by_priority"`
}

// SlaSummary provides lightweight policy details with rule count and attached task count
type SlaSummary struct {
	ID                      int64     `json:"id"`
	WorkspaceID             int64     `json:"workspace_id"`
	Name                    string    `json:"name"`
	PriorityTiers           []string  `json:"priority_tiers"`
	TargetResponseMinutes   int       `json:"target_response_minutes"`
	TargetResolutionMinutes int       `json:"target_resolution_minutes"`
	BusinessHoursOnly       bool      `json:"business_hours_only"`
	IsDefault               bool      `json:"is_default"`
	IsActive                bool      `json:"is_active"`
	RuleCount               int       `json:"rule_count"`
	TrackedTaskCount        int       `json:"tracked_task_count"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
}

// SlaFilter defines search and filter options when querying SLA policies
type SlaFilter struct {
	WorkspaceID *int64 `json:"workspace_id,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	Priority    string `json:"priority,omitempty"`
	Search      string `json:"search,omitempty"`
}

// BreachFilter defines search and filter options when inspecting SLA breaches
type BreachFilter struct {
	WorkspaceID *int64 `json:"workspace_id,omitempty"`
	PolicyID    *int64 `json:"policy_id,omitempty"`
	TaskID      *int   `json:"task_id,omitempty"`
	BreachType  string `json:"breach_type,omitempty"` // response, resolution
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

