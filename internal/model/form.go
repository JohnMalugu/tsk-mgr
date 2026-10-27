package model

import "time"

// Form publication and lifecycle status constants
const (
	FormStatusDraft     = "draft"
	FormStatusPublished = "published"
	FormStatusArchived  = "archived"
)

// IntakeForm defines a customizable task intake portal with vanity slug and task blueprints
type IntakeForm struct {
	ID                int64     `json:"id"`
	WorkspaceID       int64     `json:"workspace_id"`
	ProjectID         *int64    `json:"project_id,omitempty"`
	Title             string    `json:"title"`
	Description       string    `json:"description,omitempty"`
	Slug              string    `json:"slug"` // unique vanity URL slug
	Status            string    `json:"status"` // draft, published, archived
	IsPublished       bool      `json:"is_published"`
	AllowAnonymous    bool      `json:"allow_anonymous"`
	SubmitButtonText  string    `json:"submit_button_text"`
	SuccessMessage    string    `json:"success_message"`
	RedirectURL       string    `json:"redirect_url,omitempty"`
	DefaultPriority   string    `json:"default_priority"`
	DefaultTags       []string  `json:"default_tags"`
	DefaultAssigneeID string    `json:"default_assignee_id,omitempty"`
	SubmissionCount   int       `json:"submission_count"`
	CreatedByUserID   string    `json:"created_by_user_id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
