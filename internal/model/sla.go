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
