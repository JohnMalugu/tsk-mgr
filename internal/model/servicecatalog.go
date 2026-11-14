package model

import (
	"errors"
	"time"
)

var (
	ErrCatalogCategoryNotFound    = errors.New("service catalog category not found")
	ErrCatalogCategoryNameReq     = errors.New("service catalog category name is required")
	ErrCatalogCategorySlugReq     = errors.New("service catalog category slug is required")
	ErrDuplicateCategorySlug      = errors.New("service catalog category slug already in use")
	ErrCatalogItemNotFound        = errors.New("service catalog item not found")
	ErrCatalogItemNameReq         = errors.New("service catalog item name is required")
	ErrCatalogItemSlugReq         = errors.New("service catalog item slug is required")
	ErrDuplicateItemSlug          = errors.New("service catalog item slug already in use")
	ErrInvalidApprovalRole        = errors.New("invalid approval role: must be manager, admin, or security_officer")
	ErrServiceRequestNotFound     = errors.New("service request not found")
	ErrServiceRequestTitleReq     = errors.New("service request title is required")
	ErrInvalidRequestStatus       = errors.New("invalid service request status")
	ErrInvalidRequestPriority     = errors.New("invalid service request priority")
	ErrRequestApprovalNotFound    = errors.New("service request approval not found")
	ErrInvalidApprovalStatus      = errors.New("invalid approval status: must be pending, approved, or rejected")
	ErrRequestAlreadyDecided      = errors.New("service request approval has already been decided")
	ErrRequestCommentEmpty        = errors.New("service request comment cannot be empty")
)

// ServiceCatalogCategory groups service catalog items into logical offerings
type ServiceCatalogCategory struct {
	ID           int64     `json:"id"`
	WorkspaceID  int64     `json:"workspace_id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  string    `json:"description,omitempty"`
	Icon         string    `json:"icon,omitempty"`
	DisplayOrder int       `json:"display_order"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Enriched fields
	ItemCount int `json:"item_count,omitempty"`
}

// Approval role constants for service catalog item workflows
const (
	ApprovalRoleManager  = "manager"
	ApprovalRoleAdmin    = "admin"
	ApprovalRoleSecurity = "security_officer"
)

// Form field type constants for dynamic catalog item questionnaires
const (
	FieldTypeText     = "text"
	FieldTypeTextArea = "textarea"
	FieldTypeNumber   = "number"
	FieldTypeSelect   = "select"
	FieldTypeBoolean  = "boolean"
)

// CatalogFormField represents a single input specification in a catalog item questionnaire
type CatalogFormField struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Label        string   `json:"label"`
	Type         string   `json:"type"` // text, textarea, number, select, boolean
	Required     bool     `json:"required"`
	DefaultValue string   `json:"default_value,omitempty"`
	Options      []string `json:"options,omitempty"` // For select dropdowns
	HelpText     string   `json:"help_text,omitempty"`
}

// CatalogFormSchema defines the schema for a catalog item intake form
type CatalogFormSchema struct {
	Fields []CatalogFormField `json:"fields"`
}

// ServiceCatalogItem represents an orderable IT service offering or resource
type ServiceCatalogItem struct {
	ID                        int64      `json:"id"`
	WorkspaceID               int64      `json:"workspace_id"`
	CategoryID                int64      `json:"category_id"`
	Name                      string     `json:"name"`
	Slug                      string     `json:"slug"`
	Description               string     `json:"description,omitempty"`
	ShortDescription          string     `json:"short_description,omitempty"`
	Icon                      string     `json:"icon,omitempty"`
	EstimatedFulfillmentHours float64    `json:"estimated_fulfillment_hours"`
	CostEstimateCents         int64      `json:"cost_estimate_cents"`
	Currency                  string     `json:"currency"`
	RequiresApproval          bool       `json:"requires_approval"`
	ApprovalRole              string     `json:"approval_role,omitempty"` // manager, admin, security_officer
	AutoCreateTask            bool       `json:"auto_create_task"`
	TargetProjectID           *int64     `json:"target_project_id,omitempty"`
	CIID                      *int64     `json:"ci_id,omitempty"` // Link to CMDB Configuration Item
	IsActive                  bool       `json:"is_active"`
	FormSchemaJSON            string     `json:"form_schema_json,omitempty"`
	CreatedByID               string     `json:"created_by_id"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`

	// Enriched fields
	CategoryName string             `json:"category_name,omitempty"`
	ProjectName  string             `json:"project_name,omitempty"`
	CIName       string             `json:"ci_name,omitempty"`
	RequestCount int                `json:"request_count,omitempty"`
	FormSchema   *CatalogFormSchema `json:"form_schema,omitempty"`
}
