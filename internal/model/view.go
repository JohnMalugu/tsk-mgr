package model

import "time"

// Supported sort fields for saved views
const (
	SortByCreatedAt = "created_at"
	SortByDueDate   = "due_date"
	SortByPriority  = "priority"
	SortByTitle     = "title"
	SortByStatus    = "status"
	SortByUpdatedAt = "updated_at"
)

// Supported sort orders
const (
	SortOrderAsc  = "asc"
	SortOrderDesc = "desc"
)

// Supported group-by fields for aggregating view results
const (
	GroupByNone      = "none"
	GroupByStatus    = "status"
	GroupByPriority  = "priority"
	GroupByProjectID = "project_id"
	GroupByAssignee  = "assignee"
)

// ViewFilterCriteria represents the multi-dimensional filter configuration of a saved view
type ViewFilterCriteria struct {
	Status         []string   `json:"status,omitempty"`
	Priority       []string   `json:"priority,omitempty"`
	ProjectID      *int       `json:"project_id,omitempty"`
	WorkspaceID    *int64     `json:"workspace_id,omitempty"`
	AssigneeUserID *string    `json:"assignee_user_id,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	HasDueDate     *bool      `json:"has_due_date,omitempty"`
	DueBefore      *time.Time `json:"due_before,omitempty"`
	DueAfter       *time.Time `json:"due_after,omitempty"`
	Search         string     `json:"search,omitempty"`
}

// SavedView defines a user-curated or shared task view preset with custom filters and grouping
type SavedView struct {
	ID             int64              `json:"id"`
	WorkspaceID    int64              `json:"workspace_id"`
	UserID         string             `json:"user_id"`
	Name           string             `json:"name"`
	Description    string             `json:"description"`
	Icon           string             `json:"icon"`
	Color          string             `json:"color"`
	FilterCriteria ViewFilterCriteria `json:"filter_criteria"`
	SortBy         string             `json:"sort_by"`
	SortOrder      string             `json:"sort_order"`
	GroupBy        string             `json:"group_by"`
	IsShared       bool               `json:"is_shared"`
	IsPinned       bool               `json:"is_pinned"`
	Position       int                `json:"position"`
	UsageCount     int                `json:"usage_count"`
	LastUsedAt     *time.Time         `json:"last_used_at,omitempty"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}
