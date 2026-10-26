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

// CreateReleaseRequest represents payload for provisioning a new release milestone
type CreateReleaseRequest struct {
	WorkspaceID int64      `json:"workspace_id"`
	ProjectID   *int64     `json:"project_id,omitempty"`
	Name        string     `json:"name"`
	Version     string     `json:"version"` // semver-compliant version string
	Description string     `json:"description,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
}

// UpdateReleaseRequest represents payload for modifying release metadata
type UpdateReleaseRequest struct {
	Name        *string    `json:"name,omitempty"`
	Version     *string    `json:"version,omitempty"`
	Description *string    `json:"description,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
}

// UpdateReleaseStatusRequest represents payload for advancing release lifecycle
type UpdateReleaseStatusRequest struct {
	Status string `json:"status"` // planning, active, code_freeze, released, archived
}

// AddReleaseTaskItem specifies task ID and categorization when attaching a task to a release
type AddReleaseTaskItem struct {
	TaskID             int    `json:"task_id"`
	Category           string `json:"category"` // feature, bugfix, improvement, security, chore
	IncludeInChangelog *bool  `json:"include_in_changelog,omitempty"`
	CustomNotes        string `json:"custom_notes,omitempty"`
}

// AddReleaseTasksRequest allows batch associating tasks with a release
type AddReleaseTasksRequest struct {
	Tasks []AddReleaseTaskItem `json:"tasks"`
}

// UpdateReleaseItemRequest represents payload for modifying release task categorization or changelog text
type UpdateReleaseItemRequest struct {
	Category           *string `json:"category,omitempty"`
	IncludeInChangelog *bool   `json:"include_in_changelog,omitempty"`
	CustomNotes        *string `json:"custom_notes,omitempty"`
}

// Release notes formatting constants
const (
	ReleaseNotesFormatMarkdown = "markdown"
	ReleaseNotesFormatJSON     = "json"
)

// ReleaseNotesOptions specifies filtering and presentation flags for changelog generation
type ReleaseNotesOptions struct {
	Format             string `json:"format"`              // "markdown" or "json"
	IncludeUncompleted bool   `json:"include_uncompleted"` // whether to include incomplete tasks
	HeaderTitle        string `json:"header_title,omitempty"`
}

// ChangelogEntry represents a single bullet item in generated release notes
type ChangelogEntry struct {
	TaskID      int    `json:"task_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Completed   bool   `json:"completed"`
	CustomNotes string `json:"custom_notes,omitempty"`
}

// GeneratedReleaseNotes contains formatted changelog output and categorized items
type GeneratedReleaseNotes struct {
	ReleaseID   int64                       `json:"release_id"`
	Name        string                      `json:"name"`
	Version     string                      `json:"version"`
	Status      string                      `json:"status"`
	ReleasedAt  *time.Time                  `json:"released_at,omitempty"`
	Format      string                      `json:"format"`
	Content     string                      `json:"content"`
	Sections    map[string][]ChangelogEntry `json:"sections,omitempty"`
	TotalItems  int                         `json:"total_items"`
	GeneratedAt time.Time                   `json:"generated_at"`
}

// Release health status indicators
const (
	ReleaseHealthOnTrack   = "on_track"
	ReleaseHealthAtRisk    = "at_risk"
	ReleaseHealthDelayed   = "delayed"
	ReleaseHealthCompleted = "completed"
)

// ReleaseMetrics aggregates completion progress, category counts, and schedule health
type ReleaseMetrics struct {
	ReleaseID            int64          `json:"release_id"`
	Name                 string         `json:"name"`
	Version              string         `json:"version"`
	Status               string         `json:"status"`
	TotalTasks           int            `json:"total_tasks"`
	CompletedTasks       int            `json:"completed_tasks"`
	InProgressTasks      int            `json:"in_progress_tasks"`
	PendingTasks         int            `json:"pending_tasks"`
	CompletionPercentage float64        `json:"completion_percentage"`
	CategoryBreakdown    map[string]int `json:"category_breakdown"`
	HasBlockers          bool           `json:"has_blockers"`
	BlockerCount         int            `json:"blocker_count"`
	DaysRemaining        int            `json:"days_remaining"`
	HealthStatus         string         `json:"health_status"` // on_track, at_risk, delayed, completed
	CalculatedAt         time.Time      `json:"calculated_at"`
}

// CloneReleaseRequest specifies target version and optional task rollover rules when cloning a release
type CloneReleaseRequest struct {
	NewName            string     `json:"new_name"`
	NewVersion         string     `json:"new_version"`
	TargetDate         *time.Time `json:"target_date,omitempty"`
	RolloverIncomplete bool       `json:"rollover_incomplete"`
}

// ReleaseFilter defines query parameters for searching and listing releases
type ReleaseFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   *int64 `json:"project_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Query       string `json:"q,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// ReleaseSummary provides compact release overview for list catalogs and dashboards
type ReleaseSummary struct {
	ID                   int64      `json:"id"`
	WorkspaceID          int64      `json:"workspace_id"`
	ProjectID            *int64     `json:"project_id,omitempty"`
	Name                 string     `json:"name"`
	Version              string     `json:"version"`
	Status               string     `json:"status"`
	TaskCount            int        `json:"task_count"`
	CompletedCount       int        `json:"completed_count"`
	CompletionPercentage float64    `json:"completion_percentage"`
	TargetDate           *time.Time `json:"target_date,omitempty"`
	ReleasedAt           *time.Time `json:"released_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}
