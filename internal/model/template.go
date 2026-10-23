package model

import "time"

// Template categories
const (
	TemplateCategoryGeneral    = "general"
	TemplateCategoryBugReport  = "bug_report"
	TemplateCategoryFeature    = "feature"
	TemplateCategoryOnboarding = "onboarding"
	TemplateCategorySprintTask = "sprint_task"
	TemplateCategoryReview     = "review"
)

// SubtaskBlueprint represents a blueprint for creating a subtask upon template instantiation
type SubtaskBlueprint struct {
	Title    string `json:"title"`
	Position int    `json:"position"`
}

// ChecklistBlueprint represents a blueprint for a checklist item upon template instantiation
type ChecklistBlueprint struct {
	Text string `json:"text"`
}

// TaskTemplate defines a reusable blueprint for instantiating tasks with subtasks, checklists, and tags
type TaskTemplate struct {
	ID               int64                `json:"id"`
	WorkspaceID      int64                `json:"workspace_id"`
	CreatedByUserID  string               `json:"created_by_user_id"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	Category         string               `json:"category"`
	DefaultPriority  string               `json:"default_priority"`
	EstimatedMinutes int                  `json:"estimated_minutes"`
	Tags             []string             `json:"tags"`
	ChecklistItems   []ChecklistBlueprint `json:"checklist_items"`
	Subtasks         []SubtaskBlueprint   `json:"subtasks"`
	UsageCount       int                  `json:"usage_count"`
	IsArchived       bool                 `json:"is_archived"`
	CreatedAt        time.Time            `json:"created_at"`
	UpdatedAt        time.Time            `json:"updated_at"`
	LastUsedAt       *time.Time           `json:"last_used_at,omitempty"`
}
