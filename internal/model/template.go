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

// TemplateInstantiationRequest defines parameters to instantiate a new task from a template
type TemplateInstantiationRequest struct {
	Title          string     `json:"title,omitempty"`
	ProjectID      *int       `json:"project_id,omitempty"`
	AssigneeUserID string     `json:"assignee_user_id"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	Priority       string     `json:"priority,omitempty"`
	ExtraTags      []string   `json:"extra_tags,omitempty"`
}

// InstantiatedTaskResult encapsulates the output of a template instantiation
type InstantiatedTaskResult struct {
	Task                  *Task `json:"task"`
	SubtasksCreated       int   `json:"subtasks_created"`
	ChecklistItemsCreated int   `json:"checklist_items_created"`
	TemplateID            int64 `json:"template_id"`
}

// TemplateFilter specifies criteria for querying task templates
type TemplateFilter struct {
	WorkspaceID *int64  `json:"workspace_id,omitempty"`
	Category    *string `json:"category,omitempty"`
	SearchQuery *string `json:"search_query,omitempty"`
	IsArchived  *bool   `json:"is_archived,omitempty"`
	Offset      int     `json:"offset"`
	Limit       int     `json:"limit"`
}

// TemplateSummary provides paginated results of task templates
type TemplateSummary struct {
	Templates []TaskTemplate `json:"templates"`
	Total     int            `json:"total"`
	Offset    int            `json:"offset"`
	Limit     int            `json:"limit"`
}

// CreateTemplateRequest defines payload to create a new task template blueprint
type CreateTemplateRequest struct {
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
}

// UpdateTemplateRequest defines payload to modify existing template properties
type UpdateTemplateRequest struct {
	Title            *string               `json:"title,omitempty"`
	Description      *string               `json:"description,omitempty"`
	Category         *string               `json:"category,omitempty"`
	DefaultPriority  *string               `json:"default_priority,omitempty"`
	EstimatedMinutes *int                  `json:"estimated_minutes,omitempty"`
	Tags             *[]string             `json:"tags,omitempty"`
	ChecklistItems   *[]ChecklistBlueprint `json:"checklist_items,omitempty"`
	Subtasks         *[]SubtaskBlueprint   `json:"subtasks,omitempty"`
	IsArchived       *bool                 `json:"is_archived,omitempty"`
}
