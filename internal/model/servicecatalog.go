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

// Request lifecycle status constants
const (
	RequestStatusSubmitted       = "submitted"
	RequestStatusPendingApproval = "pending_approval"
	RequestStatusApproved        = "approved"
	RequestStatusInProgress      = "in_progress"
	RequestStatusFulfilled       = "fulfilled"
	RequestStatusRejected        = "rejected"
	RequestStatusCancelled       = "cancelled"
)

// Request priority levels
const (
	RequestPriorityLow    = "low"
	RequestPriorityMedium = "medium"
	RequestPriorityHigh   = "high"
	RequestPriorityUrgent = "urgent"
)

// ServiceRequest represents an instance of a user request ordered through the service catalog
type ServiceRequest struct {
	ID                 int64      `json:"id"`
	WorkspaceID        int64      `json:"workspace_id"`
	CatalogItemID      int64      `json:"catalog_item_id"`
	RequestNumber      string     `json:"request_number"` // e.g. REQ-1001
	Title              string     `json:"title"`
	Description        string     `json:"description,omitempty"`
	Status             string     `json:"status"` // submitted, pending_approval, approved, in_progress, fulfilled, rejected, cancelled
	Priority           string     `json:"priority"` // low, medium, high, urgent
	RequesterID        string     `json:"requester_id"`
	AssignedToID       *string    `json:"assigned_to_id,omitempty"`
	FormDataJSON       string     `json:"form_data_json,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	SubmittedAt        time.Time  `json:"submitted_at"`
	ApprovedAt         *time.Time `json:"approved_at,omitempty"`
	FulfilledAt        *time.Time `json:"fulfilled_at,omitempty"`
	CancelledAt        *time.Time `json:"cancelled_at,omitempty"`
	CancellationReason string     `json:"cancellation_reason,omitempty"`
	RejectionReason    string     `json:"rejection_reason,omitempty"`
	TaskID             *int       `json:"task_id,omitempty"` // Link to task generated for fulfillment
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Enriched fields
	CatalogItemName string                 `json:"catalog_item_name,omitempty"`
	RequesterName   string                 `json:"requester_name,omitempty"`
	AssignedToName  string                 `json:"assigned_to_name,omitempty"`
	TaskTitle       string                 `json:"task_title,omitempty"`
	FormData        map[string]interface{} `json:"form_data,omitempty"`
	Approvals       []ServiceRequestApproval `json:"approvals,omitempty"`
	Comments        []ServiceRequestComment  `json:"comments,omitempty"`
}

// Approval status constants
const (
	ApprovalStatusPending  = "pending"
	ApprovalStatusApproved = "approved"
	ApprovalStatusRejected = "rejected"
)

// ServiceRequestApproval represents a governance review milestone for a service request
type ServiceRequestApproval struct {
	ID         int64      `json:"id"`
	RequestID  int64      `json:"request_id"`
	StepOrder  int        `json:"step_order"`
	ApproverID string     `json:"approver_id"`
	Status     string     `json:"status"` // pending, approved, rejected
	Comments   string     `json:"comments,omitempty"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`

	// Enriched fields
	ApproverName string `json:"approver_name,omitempty"`
}

// ServiceRequestComment represents an internal note or communication message on a request
type ServiceRequestComment struct {
	ID         int64     `json:"id"`
	RequestID  int64     `json:"request_id"`
	UserID     string    `json:"user_id"`
	IsInternal bool      `json:"is_internal"` // true for fulfillment team internal notes
	Comment    string    `json:"comment"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// Enriched fields
	UserName string `json:"user_name,omitempty"`
}

// CreateCatalogCategoryRequest payload for registering a category
type CreateCatalogCategoryRequest struct {
	WorkspaceID  int64  `json:"workspace_id"`
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	Description  string `json:"description,omitempty"`
	Icon         string `json:"icon,omitempty"`
	DisplayOrder int    `json:"display_order,omitempty"`
}

// UpdateCatalogCategoryRequest payload for updating a category
type UpdateCatalogCategoryRequest struct {
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Description  *string `json:"description,omitempty"`
	Icon         *string `json:"icon,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// CreateCatalogItemRequest payload for creating a catalog offering
type CreateCatalogItemRequest struct {
	WorkspaceID               int64              `json:"workspace_id"`
	CategoryID                int64              `json:"category_id"`
	Name                      string             `json:"name"`
	Slug                      string             `json:"slug,omitempty"`
	Description               string             `json:"description,omitempty"`
	ShortDescription          string             `json:"short_description,omitempty"`
	Icon                      string             `json:"icon,omitempty"`
	EstimatedFulfillmentHours float64            `json:"estimated_fulfillment_hours,omitempty"`
	CostEstimateCents         int64              `json:"cost_estimate_cents,omitempty"`
	Currency                  string             `json:"currency,omitempty"`
	RequiresApproval          bool               `json:"requires_approval,omitempty"`
	ApprovalRole              string             `json:"approval_role,omitempty"` // manager, admin, security_officer
	AutoCreateTask            bool               `json:"auto_create_task,omitempty"`
	TargetProjectID           *int64             `json:"target_project_id,omitempty"`
	CIID                      *int64             `json:"ci_id,omitempty"`
	FormSchema                *CatalogFormSchema `json:"form_schema,omitempty"`
}

// UpdateCatalogItemRequest payload for modifying a catalog item
type UpdateCatalogItemRequest struct {
	CategoryID                *int64             `json:"category_id,omitempty"`
	Name                      *string            `json:"name,omitempty"`
	Slug                      *string            `json:"slug,omitempty"`
	Description               *string            `json:"description,omitempty"`
	ShortDescription          *string            `json:"short_description,omitempty"`
	Icon                      *string            `json:"icon,omitempty"`
	EstimatedFulfillmentHours *float64           `json:"estimated_fulfillment_hours,omitempty"`
	CostEstimateCents         *int64             `json:"cost_estimate_cents,omitempty"`
	Currency                  *string            `json:"currency,omitempty"`
	RequiresApproval          *bool              `json:"requires_approval,omitempty"`
	ApprovalRole              *string            `json:"approval_role,omitempty"`
	AutoCreateTask            *bool              `json:"auto_create_task,omitempty"`
	TargetProjectID           *int64             `json:"target_project_id,omitempty"`
	CIID                      *int64             `json:"ci_id,omitempty"`
	IsActive                  *bool              `json:"is_active,omitempty"`
	FormSchema                *CatalogFormSchema `json:"form_schema,omitempty"`
}

// ValidateApprovalRole ensures the role matches an allowed approval persona
func ValidateApprovalRole(role string) error {
	switch role {
	case ApprovalRoleManager, ApprovalRoleAdmin, ApprovalRoleSecurity, "":
		return nil
	default:
		return ErrInvalidApprovalRole
	}
}

// ValidateRequestStatus checks if request status transition is recognized
func ValidateRequestStatus(status string) error {
	switch status {
	case RequestStatusSubmitted, RequestStatusPendingApproval, RequestStatusApproved,
		RequestStatusInProgress, RequestStatusFulfilled, RequestStatusRejected, RequestStatusCancelled:
		return nil
	default:
		return ErrInvalidRequestStatus
	}
}

// ValidateRequestPriority checks if priority is valid
func ValidateRequestPriority(priority string) error {
	switch priority {
	case RequestPriorityLow, RequestPriorityMedium, RequestPriorityHigh, RequestPriorityUrgent:
		return nil
	default:
		return ErrInvalidRequestPriority
	}
}

// SubmitServiceRequest payload for placing an order from a catalog item
type SubmitServiceRequest struct {
	CatalogItemID int64                  `json:"catalog_item_id"`
	Title         string                 `json:"title"`
	Description   string                 `json:"description,omitempty"`
	Priority      string                 `json:"priority,omitempty"` // low, medium, high, urgent
	FormData      map[string]interface{} `json:"form_data,omitempty"`
	DueAt         *time.Time             `json:"due_at,omitempty"`
}

// UpdateServiceRequest payload for transitioning or assigning a request
type UpdateServiceRequest struct {
	Status             *string    `json:"status,omitempty"`
	Priority           *string    `json:"priority,omitempty"`
	AssignedToID       *string    `json:"assigned_to_id,omitempty"`
	DueAt              *time.Time `json:"due_at,omitempty"`
	RejectionReason    *string    `json:"rejection_reason,omitempty"`
	CancellationReason *string    `json:"cancellation_reason,omitempty"`
}

// ReviewApprovalRequest payload for approving or rejecting a service request
type ReviewApprovalRequest struct {
	Status   string `json:"status"` // approved, rejected
	Comments string `json:"comments,omitempty"`
}

// AddRequestCommentRequest payload for adding an update or internal note to a request
type AddRequestCommentRequest struct {
	Comment    string `json:"comment"`
	IsInternal bool   `json:"is_internal,omitempty"`
}

// CatalogCategoryFilter defines search and filter options for catalog categories
type CatalogCategoryFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	IsActive    *bool  `json:"is_active,omitempty"`
	Search      string `json:"search,omitempty"`
}

// CatalogItemFilter defines search and filter options for catalog items
type CatalogItemFilter struct {
	WorkspaceID      int64  `json:"workspace_id"`
	CategoryID       *int64 `json:"category_id,omitempty"`
	RequiresApproval *bool  `json:"requires_approval,omitempty"`
	IsActive         *bool  `json:"is_active,omitempty"`
	Search           string `json:"search,omitempty"`
	Limit            int    `json:"limit,omitempty"`
	Offset           int    `json:"offset,omitempty"`
}

// ServiceRequestFilter defines query filters for listing user requests
type ServiceRequestFilter struct {
	WorkspaceID   int64  `json:"workspace_id"`
	CatalogItemID *int64 `json:"catalog_item_id,omitempty"`
	RequesterID   string `json:"requester_id,omitempty"`
	AssignedToID  string `json:"assigned_to_id,omitempty"`
	Status        string `json:"status,omitempty"`
	Priority      string `json:"priority,omitempty"`
	Search        string `json:"search,omitempty"`
	Limit         int    `json:"limit,omitempty"`
	Offset        int    `json:"offset,omitempty"`
}

// ItemRequestStat provides top-requested catalog item distribution
type ItemRequestStat struct {
	ItemID       int64  `json:"item_id"`
	ItemName     string `json:"item_name"`
	RequestCount int    `json:"request_count"`
}

// ServiceCatalogMetrics summarizes self-service portal performance and fulfillment velocity
type ServiceCatalogMetrics struct {
	TotalCategories         int               `json:"total_categories"`
	TotalItems              int               `json:"total_items"`
	TotalRequests           int               `json:"total_requests"`
	PendingApprovalRequests int               `json:"pending_approval_requests"`
	InProgressRequests      int               `json:"in_progress_requests"`
	FulfilledRequests       int               `json:"fulfilled_requests"`
	RejectedRequests        int               `json:"rejected_requests"`
	CancelledRequests       int               `json:"cancelled_requests"`
	AvgFulfillmentHours     float64           `json:"avg_fulfillment_hours"`
	TotalCostEstimateCents  int64             `json:"total_cost_estimate_cents"`
	ApprovalRate            float64           `json:"approval_rate"`
	MostRequestedItems      []ItemRequestStat `json:"most_requested_items"`
}
