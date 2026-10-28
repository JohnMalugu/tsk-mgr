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
