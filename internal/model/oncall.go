package model

import (
	"errors"
	"time"
)

var (
	ErrScheduleNameRequired      = errors.New("on-call schedule name is required")
	ErrInvalidRotationType       = errors.New("invalid rotation type: must be daily, weekly, or custom")
	ErrInvalidTimezone           = errors.New("invalid timezone specification")
	ErrScheduleNotFound          = errors.New("on-call schedule not found")
	ErrEscalationPolicyNotFound  = errors.New("escalation policy not found")
	ErrEscalationStepNotFound    = errors.New("escalation step not found")
	ErrOnCallOverrideNotFound    = errors.New("on-call override not found")
	ErrOnCallPageNotFound        = errors.New("on-call page not found")
	ErrNoOnCallResponders        = errors.New("no active on-call responders available")
	ErrInvalidOverrideTimeRange  = errors.New("override start time must be before end time")
	ErrInvalidEscalationTarget   = errors.New("invalid escalation target type: must be user, schedule, or webhook")
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
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   *int64     `json:"project_id,omitempty"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	RepeatCount int        `json:"repeat_count"` // 0 = no loop, 1..9 = loop N times
	IsDefault   bool       `json:"is_default"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

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
	TargetType   string    `json:"target_type"` // user, schedule, webhook
	TargetID     string    `json:"target_id"`   // UserID, ScheduleID, or webhook URL
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
