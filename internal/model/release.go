package model

import "time"

// Release status lifecycle constants
const (
	ReleaseStatusPlanning = "planning"
	ReleaseStatusActive   = "active"
	ReleaseStatusFreeze   = "code_freeze"
	ReleaseStatusReleased = "released"
	ReleaseStatusArchived = "archived"
)

// Release represents a scoped product version, milestone, or deployment target
type Release struct {
	ID              int64      `json:"id"`
	WorkspaceID     int64      `json:"workspace_id"`
	ProjectID       *int64     `json:"project_id,omitempty"`
	Name            string     `json:"name"`
	Version         string     `json:"version"` // e.g. "v1.2.0"
	Description     string     `json:"description,omitempty"`
	Status          string     `json:"status"` // planning, active, code_freeze, released, archived
	TargetDate      *time.Time `json:"target_date,omitempty"`
	ReleasedAt      *time.Time `json:"released_at,omitempty"`
	CreatedByUserID string     `json:"created_by_user_id"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
