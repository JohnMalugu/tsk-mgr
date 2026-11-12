package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrScheduleNameRequired     = errors.New("on-call schedule name is required")
	ErrInvalidRotationType      = errors.New("invalid rotation type: must be daily, weekly, or custom")
	ErrInvalidTimezone          = errors.New("invalid timezone specification")
	ErrScheduleNotFound         = errors.New("on-call schedule not found")
	ErrEscalationPolicyNotFound = errors.New("escalation policy not found")
	ErrEscalationStepNotFound   = errors.New("escalation step not found")
	ErrOnCallOverrideNotFound   = errors.New("on-call override not found")
	ErrOnCallPageNotFound       = errors.New("on-call page not found")
	ErrNoOnCallResponders       = errors.New("no active on-call responders available")
	ErrInvalidOverrideTimeRange = errors.New("override start time must be before end time")
	ErrInvalidEscalationTarget  = errors.New("invalid escalation target type: must be user, schedule, or webhook")
)

// Rotation types for on-call schedules
const (
	RotationDaily  = "daily"
	RotationWeekly = "weekly"
	RotationCustom = "custom"
)

// OnCallSchedule represents an on-call rotation schedule defining shift rotations for responders
type OnCallSchedule struct {
	ID                   int64      `json:"id"`
	WorkspaceID          int64      `json:"workspace_id"`
	ProjectID            *int64     `json:"project_id,omitempty"`
	Name                 string     `json:"name"`
	Description          string     `json:"description,omitempty"`
	Timezone             string     `json:"timezone"`
	RotationType         string     `json:"rotation_type"` // daily, weekly, custom
	RotationIntervalDays int        `json:"rotation_interval_days"`
	ShiftStartTime       string     `json:"shift_start_time"` // HH:MM, e.g. "09:00"
	ShiftDurationHours   int        `json:"shift_duration_hours"`
	HandoverDayOfWeek    int        `json:"handover_day_of_week"` // 0=Sunday, 1=Monday...
	StartDate            time.Time  `json:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	IsActive             bool       `json:"is_active"`
	CreatedByID          string     `json:"created_by_id"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Enriched fields
	ProjectName           string  `json:"project_name,omitempty"`
	CreatorName           string  `json:"creator_name,omitempty"`
	ParticipantsCount     int     `json:"participants_count"`
	CurrentOnCallUserID   *string `json:"current_oncall_user_id,omitempty"`
	CurrentOnCallUserName string  `json:"current_oncall_user_name,omitempty"`
}

// OnCallParticipant represents a rotating participant within an on-call schedule
type OnCallParticipant struct {
	ID            int64     `json:"id"`
	ScheduleID    int64     `json:"schedule_id"`
	UserID        string    `json:"user_id"`
	PositionOrder int       `json:"position_order"`
	CreatedAt     time.Time `json:"created_at"`

	// Enriched fields
	UserName  string `json:"user_name,omitempty"`
	UserEmail string `json:"user_email,omitempty"`
}

// OnCallShift represents a concrete computed time window during which a user is on-call
type OnCallShift struct {
	ScheduleID  int64     `json:"schedule_id"`
	UserID      string    `json:"user_id"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	UserName    string    `json:"user_name,omitempty"`
	UserEmail   string    `json:"user_email,omitempty"`
	IsActiveNow bool      `json:"is_active_now"`
}

// Escalation step target types
const (
	TargetUser     = "user"
	TargetSchedule = "schedule"
	TargetWebhook  = "webhook"
)

// EscalationPolicy defines multi-tiered paging and alert escalation rules for incidents
type EscalationPolicy struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ProjectID   *int64    `json:"project_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	RepeatCount int       `json:"repeat_count"` // 0 = no loop, 1..9 = loop N times
	IsDefault   bool      `json:"is_default"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Enriched fields
	ProjectName string           `json:"project_name,omitempty"`
	StepsCount  int              `json:"steps_count"`
	Steps       []EscalationStep `json:"steps,omitempty"`
}

// EscalationStep represents a discrete notification level within an escalation policy
type EscalationStep struct {
	ID           int64     `json:"id"`
	PolicyID     int64     `json:"policy_id"`
	StepOrder    int       `json:"step_order"`
	TargetType   string    `json:"target_type"`   // user, schedule, webhook
	TargetID     string    `json:"target_id"`     // UserID, ScheduleID, or webhook URL
	DelayMinutes int       `json:"delay_minutes"` // minutes before escalating to next step
	CreatedAt    time.Time `json:"created_at"`

	// Enriched fields
	TargetName string `json:"target_name,omitempty"`
}

// OnCallOverride represents a temporary shift swap or coverage takeover on an on-call schedule
type OnCallOverride struct {
	ID             int64     `json:"id"`
	ScheduleID     int64     `json:"schedule_id"`
	OriginalUserID *string   `json:"original_user_id,omitempty"` // User being covered (optional)
	OverrideUserID string    `json:"override_user_id"`           // User covering the shift
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Reason         string    `json:"reason,omitempty"`
	CreatedByID    string    `json:"created_by_id"`
	CreatedAt      time.Time `json:"created_at"`

	// Enriched fields
	OriginalUserName string `json:"original_user_name,omitempty"`
	OverrideUserName string `json:"override_user_name,omitempty"`
	ScheduleName     string `json:"schedule_name,omitempty"`
	IsActiveNow      bool   `json:"is_active_now"`
}

// On-call page status constants
const (
	PageStatusTriggered    = "triggered"
	PageStatusAcknowledged = "acknowledged"
	PageStatusEscalated    = "escalated"
	PageStatusResolved     = "resolved"
)

// On-call page trigger sources
const (
	PageTriggerIncident = "incident"
	PageTriggerSLA      = "sla_breach"
	PageTriggerManual   = "manual"
	PageTriggerWebhook  = "webhook"
)

// OnCallPage represents an incident alert dispatched to an on-call escalation policy
type OnCallPage struct {
	ID               int64      `json:"id"`
	WorkspaceID      int64      `json:"workspace_id"`
	PolicyID         int64      `json:"policy_id"`
	IncidentID       *int64     `json:"incident_id,omitempty"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary,omitempty"`
	Status           string     `json:"status"` // triggered, acknowledged, escalated, resolved
	CurrentStepOrder int        `json:"current_step_order"`
	PagedUserID      *string    `json:"paged_user_id,omitempty"`
	AcknowledgedBy   *string    `json:"acknowledged_by,omitempty"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	TriggerSource    string     `json:"trigger_source"` // incident, sla_breach, manual, webhook
	CreatedByID      string     `json:"created_by_id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Enriched fields
	PolicyName           string `json:"policy_name,omitempty"`
	IncidentNumber       int    `json:"incident_number,omitempty"`
	IncidentTitle        string `json:"incident_title,omitempty"`
	PagedUserName        string `json:"paged_user_name,omitempty"`
	AcknowledgedUserName string `json:"acknowledged_user_name,omitempty"`
}

// ValidateRotationType validates whether a schedule rotation type is supported
func ValidateRotationType(rot string) error {
	switch strings.ToLower(strings.TrimSpace(rot)) {
	case RotationDaily, RotationWeekly, RotationCustom:
		return nil
	default:
		return ErrInvalidRotationType
	}
}

// ValidateTargetType validates whether an escalation target type is supported
func ValidateTargetType(target string) error {
	switch strings.ToLower(strings.TrimSpace(target)) {
	case TargetUser, TargetSchedule, TargetWebhook:
		return nil
	default:
		return ErrInvalidEscalationTarget
	}
}

// ValidatePageStatus validates whether an on-call page status is recognized
func ValidatePageStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case PageStatusTriggered, PageStatusAcknowledged, PageStatusEscalated, PageStatusResolved:
		return nil
	default:
		return errors.New("invalid on-call page status")
	}
}

// ValidateTimezone validates whether a timezone string can be loaded
func ValidateTimezone(tz string) error {
	trimmed := strings.TrimSpace(tz)
	if trimmed == "" {
		return nil
	}
	_, err := time.LoadLocation(trimmed)
	if err != nil {
		return ErrInvalidTimezone
	}
	return nil
}

// ActiveResponder represents the currently active on-call engineer resolved in real time
type ActiveResponder struct {
	UserID         string    `json:"user_id"`
	UserName       string    `json:"user_name,omitempty"`
	UserEmail      string    `json:"user_email,omitempty"`
	ScheduleID     int64     `json:"schedule_id"`
	ScheduleName   string    `json:"schedule_name,omitempty"`
	IsOverride     bool      `json:"is_override"`
	ShiftStartTime time.Time `json:"shift_start_time"`
	ShiftEndTime   time.Time `json:"shift_end_time"`
}

// CreateScheduleRequest represents payload for creating an on-call schedule
type CreateScheduleRequest struct {
	WorkspaceID          int64      `json:"workspace_id"`
	ProjectID            *int64     `json:"project_id,omitempty"`
	Name                 string     `json:"name"`
	Description          string     `json:"description,omitempty"`
	Timezone             string     `json:"timezone,omitempty"`
	RotationType         string     `json:"rotation_type"`
	RotationIntervalDays int        `json:"rotation_interval_days,omitempty"`
	ShiftStartTime       string     `json:"shift_start_time,omitempty"`
	ShiftDurationHours   int        `json:"shift_duration_hours,omitempty"`
	HandoverDayOfWeek    int        `json:"handover_day_of_week,omitempty"`
	StartDate            time.Time  `json:"start_date"`
	EndDate              *time.Time `json:"end_date,omitempty"`
	ParticipantUserIDs   []string   `json:"participant_user_ids,omitempty"`
}

// UpdateScheduleRequest represents payload for updating an on-call schedule
type UpdateScheduleRequest struct {
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	Timezone             *string `json:"timezone,omitempty"`
	RotationType         *string `json:"rotation_type,omitempty"`
	RotationIntervalDays *int    `json:"rotation_interval_days,omitempty"`
	ShiftStartTime       *string `json:"shift_start_time,omitempty"`
	ShiftDurationHours   *int    `json:"shift_duration_hours,omitempty"`
	HandoverDayOfWeek    *int    `json:"handover_day_of_week,omitempty"`
	IsActive             *bool   `json:"is_active,omitempty"`
}

// CreateEscalationPolicyRequest represents payload for defining an escalation policy
type CreateEscalationPolicyRequest struct {
	WorkspaceID int64                         `json:"workspace_id"`
	ProjectID   *int64                        `json:"project_id,omitempty"`
	Name        string                        `json:"name"`
	Description string                        `json:"description,omitempty"`
	RepeatCount int                           `json:"repeat_count,omitempty"`
	IsDefault   bool                          `json:"is_default,omitempty"`
	Steps       []CreateEscalationStepRequest `json:"steps,omitempty"`
}

// CreateEscalationStepRequest represents payload for an escalation policy step
type CreateEscalationStepRequest struct {
	TargetType   string `json:"target_type"`
	TargetID     string `json:"target_id"`
	DelayMinutes int    `json:"delay_minutes"`
}

// CreateOverrideRequest represents payload for scheduling an on-call shift override
type CreateOverrideRequest struct {
	OriginalUserID *string   `json:"original_user_id,omitempty"`
	OverrideUserID string    `json:"override_user_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	Reason         string    `json:"reason,omitempty"`
}

// TriggerPageRequest represents payload for triggering an on-call page
type TriggerPageRequest struct {
	WorkspaceID   int64  `json:"workspace_id"`
	PolicyID      int64  `json:"policy_id"`
	IncidentID    *int64 `json:"incident_id,omitempty"`
	Title         string `json:"title"`
	Summary       string `json:"summary,omitempty"`
	TriggerSource string `json:"trigger_source,omitempty"`
}

// AcknowledgePageRequest represents payload for acknowledging an active page
type AcknowledgePageRequest struct {
	Notes string `json:"notes,omitempty"`
}

// ResolvePageRequest represents payload for resolving an active page
type ResolvePageRequest struct {
	ResolutionNotes string `json:"resolution_notes,omitempty"`
}

// OnCallFilter defines search and filtering criteria for listing schedules
type OnCallFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   *int64 `json:"project_id,omitempty"`
	IsActive    *bool  `json:"is_active,omitempty"`
	Search      string `json:"search,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// OnCallMetrics provides aggregated operational statistics for on-call response
type OnCallMetrics struct {
	TotalSchedules               int            `json:"total_schedules"`
	ActiveSchedules              int            `json:"active_schedules"`
	TotalPolicies                int            `json:"total_policies"`
	ActiveOverrides              int            `json:"active_overrides"`
	TotalPagesTriggered          int            `json:"total_pages_triggered"`
	AcknowledgedPages            int            `json:"acknowledged_pages"`
	ResolvedPages                int            `json:"resolved_pages"`
	MeanTimeToAcknowledgeMinutes float64        `json:"mean_time_to_acknowledge_minutes"`
	MeanTimeToResolveMinutes     float64        `json:"mean_time_to_resolve_minutes"`
	PagesByStatus                map[string]int `json:"pages_by_status"`
}
