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

// Form field type constants
const (
	FormFieldTypeText      = "text"
	FormFieldTypeParagraph = "paragraph"
	FormFieldTypeDropdown  = "dropdown"
	FormFieldTypeDate      = "date"
	FormFieldTypePriority  = "priority"
	FormFieldTypeNumber    = "number"
	FormFieldTypeEmail     = "email"
)

// Form field target mapping constants
const (
	FieldTargetTitle       = "task_title"
	FieldTargetDescription = "task_description"
	FieldTargetPriority    = "task_priority"
	FieldTargetDueDate     = "task_due_date"
	FieldTargetCustom      = "custom_field"
)

// IsValidFormFieldType checks if a field type is supported by the form renderer
func IsValidFormFieldType(fieldType string) bool {
	switch fieldType {
	case FormFieldTypeText, FormFieldTypeParagraph, FormFieldTypeDropdown, FormFieldTypeDate, FormFieldTypePriority, FormFieldTypeNumber, FormFieldTypeEmail:
		return true
	default:
		return false
	}
}

// FormField defines an individual input field on an intake form
type FormField struct {
	ID           int64    `json:"id"`
	FormID       int64    `json:"form_id"`
	FieldKey     string   `json:"field_key"`
	FieldType    string   `json:"field_type"`   // text, paragraph, dropdown, date, priority, number, email
	TargetField  string   `json:"target_field"` // task_title, task_description, task_priority, task_due_date, custom_field
	Label        string   `json:"label"`
	Placeholder  string   `json:"placeholder,omitempty"`
	HelpText     string   `json:"help_text,omitempty"`
	IsRequired   bool     `json:"is_required"`
	Position     int      `json:"position"`
	Options      []string `json:"options,omitempty"` // choices for dropdown
	DefaultValue string   `json:"default_value,omitempty"`
}

// FormSubmission records each response submitted via public or internal forms
type FormSubmission struct {
	ID             int64                  `json:"id"`
	FormID         int64                  `json:"form_id"`
	TaskID         *int                   `json:"task_id,omitempty"`
	SubmitterEmail string                 `json:"submitter_email,omitempty"`
	SubmitterIP    string                 `json:"submitter_ip,omitempty"`
	UserAgent      string                 `json:"user_agent,omitempty"`
	Responses      map[string]interface{} `json:"responses"`
	SubmittedAt    time.Time              `json:"submitted_at"`
}

// PublicFormView provides public, unauthenticated metadata for rendering a web form
type PublicFormView struct {
	Title            string      `json:"title"`
	Description      string      `json:"description,omitempty"`
	Slug             string      `json:"slug"`
	SubmitButtonText string      `json:"submit_button_text"`
	AllowAnonymous   bool        `json:"allow_anonymous"`
	Fields           []FormField `json:"fields"`
}

// FormFieldDefinition defines an input field configuration when creating or updating a form
type FormFieldDefinition struct {
	FieldKey     string   `json:"field_key"`
	FieldType    string   `json:"field_type"`   // text, paragraph, dropdown, date, priority, number, email
	TargetField  string   `json:"target_field"` // task_title, task_description, task_priority, task_due_date, custom_field
	Label        string   `json:"label"`
	Placeholder  string   `json:"placeholder,omitempty"`
	HelpText     string   `json:"help_text,omitempty"`
	IsRequired   bool     `json:"is_required"`
	Position     int      `json:"position"`
	Options      []string `json:"options,omitempty"`
	DefaultValue string   `json:"default_value,omitempty"`
}

// CreateFormRequest represents payload for creating a new intake form
type CreateFormRequest struct {
	WorkspaceID       int64                 `json:"workspace_id"`
	ProjectID         *int64                `json:"project_id,omitempty"`
	Title             string                `json:"title"`
	Description       string                `json:"description,omitempty"`
	Slug              string                `json:"slug,omitempty"` // auto-derived from title if omitted
	AllowAnonymous    *bool                 `json:"allow_anonymous,omitempty"`
	SubmitButtonText  string                `json:"submit_button_text,omitempty"`
	SuccessMessage    string                `json:"success_message,omitempty"`
	RedirectURL       string                `json:"redirect_url,omitempty"`
	DefaultPriority   string                `json:"default_priority,omitempty"`
	DefaultTags       []string              `json:"default_tags,omitempty"`
	DefaultAssigneeID string                `json:"default_assignee_id,omitempty"`
	Fields            []FormFieldDefinition `json:"fields,omitempty"`
}

// UpdateFormRequest represents payload for modifying form settings and default task properties
type UpdateFormRequest struct {
	Title             *string   `json:"title,omitempty"`
	Description       *string   `json:"description,omitempty"`
	Slug              *string   `json:"slug,omitempty"`
	AllowAnonymous    *bool     `json:"allow_anonymous,omitempty"`
	SubmitButtonText  *string   `json:"submit_button_text,omitempty"`
	SuccessMessage    *string   `json:"success_message,omitempty"`
	RedirectURL       *string   `json:"redirect_url,omitempty"`
	DefaultPriority   *string   `json:"default_priority,omitempty"`
	DefaultTags       *[]string `json:"default_tags,omitempty"`
	DefaultAssigneeID *string   `json:"default_assignee_id,omitempty"`
	ProjectID         *int64    `json:"project_id,omitempty"`
}

// SetFormFieldsRequest represents payload for replacing or batch updating a form's field schema
type SetFormFieldsRequest struct {
	Fields []FormFieldDefinition `json:"fields"`
}

// UpdateFormFieldRequest represents payload for modifying an existing form field
type UpdateFormFieldRequest struct {
	Label        *string   `json:"label,omitempty"`
	Placeholder  *string   `json:"placeholder,omitempty"`
	HelpText     *string   `json:"help_text,omitempty"`
	IsRequired   *bool     `json:"is_required,omitempty"`
	Position     *int      `json:"position,omitempty"`
	Options      *[]string `json:"options,omitempty"`
	DefaultValue *string   `json:"default_value,omitempty"`
	TargetField  *string   `json:"target_field,omitempty"`
}
