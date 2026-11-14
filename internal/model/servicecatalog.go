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
