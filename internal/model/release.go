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

// Release item category constants for changelog grouping and metrics
const (
	ReleaseCategoryFeature     = "feature"
	ReleaseCategoryBugfix      = "bugfix"
	ReleaseCategoryImprovement = "improvement"
	ReleaseCategorySecurity    = "security"
	ReleaseCategoryChore       = "chore"
)

// IsValidReleaseCategory validates if the category is supported
func IsValidReleaseCategory(category string) bool {
	switch category {
	case ReleaseCategoryFeature, ReleaseCategoryBugfix, ReleaseCategoryImprovement, ReleaseCategorySecurity, ReleaseCategoryChore:
		return true
	default:
		return false
	}
}

// ReleaseItem maps a task to a release milestone with categorization and changelog directives
type ReleaseItem struct {
	ID                 int64     `json:"id"`
	ReleaseID          int64     `json:"release_id"`
	TaskID             int       `json:"task_id"`
	Category           string    `json:"category"` // feature, bugfix, improvement, security, chore
	IncludeInChangelog bool      `json:"include_in_changelog"`
	CustomNotes        string    `json:"custom_notes,omitempty"`
	AddedByUserID      string    `json:"added_by_user_id"`
	AddedAt            time.Time `json:"added_at"`

	// Enriched task metadata populated on read
	TaskTitle     string `json:"task_title,omitempty"`
	TaskStatus    string `json:"task_status,omitempty"`
	TaskPriority  string `json:"task_priority,omitempty"`
	TaskUserID    string `json:"task_user_id,omitempty"`
	TaskCompleted bool   `json:"task_completed,omitempty"`
}

