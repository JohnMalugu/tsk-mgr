package model

import (
	"errors"
	"time"
)

var (
	ErrStatusPageNotFound       = errors.New("status page not found")
	ErrStatusPageNameRequired   = errors.New("status page name is required")
	ErrStatusPageSlugRequired   = errors.New("status page slug is required")
	ErrInvalidStatusPageSlug    = errors.New("invalid status page slug: alphanumeric and hyphens only")
	ErrDuplicateStatusPageSlug  = errors.New("status page slug already in use")
	ErrComponentNotFound        = errors.New("status component not found")
	ErrComponentNameRequired    = errors.New("status component name is required")
	ErrInvalidComponentStatus   = errors.New("invalid component status")
	ErrStatusIncidentNotFound   = errors.New("status incident not found")
	ErrStatusIncidentTitleReq   = errors.New("status incident title is required")
	ErrInvalidStatusIncidentSt  = errors.New("invalid status incident status")
	ErrInvalidIncidentImpact    = errors.New("invalid incident impact level")
	ErrMaintenanceNotFound      = errors.New("maintenance window not found")
	ErrMaintenanceTitleReq      = errors.New("maintenance window title is required")
	ErrInvalidMaintenanceTimes  = errors.New("maintenance start time must be before end time")
	ErrSubscriberNotFound       = errors.New("status subscriber not found")
	ErrSubscriberTargetRequired = errors.New("subscriber target endpoint or email is required")
	ErrInvalidSubscriberType    = errors.New("invalid subscriber type: must be email, webhook, or slack")
)

// System-wide status levels for pages and components
const (
	StatusOperational         = "operational"
	StatusDegradedPerformance = "degraded_performance"
	StatusPartialOutage       = "partial_outage"
	StatusMajorOutage         = "major_outage"
	StatusUnderMaintenance    = "under_maintenance"
)

// Status page visibility modes
const (
	VisibilityPublic  = "public"
	VisibilityPrivate = "private"
)

// Status page color themes
const (
	ThemeLight  = "light"
	ThemeDark   = "dark"
	ThemeSystem = "system"
)

// StatusPage represents a public or internal service availability and status dashboard
type StatusPage struct {
	ID            int64     `json:"id"`
	WorkspaceID   int64     `json:"workspace_id"`
	ProjectID     *int64    `json:"project_id,omitempty"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	Description   string    `json:"description,omitempty"`
	Headline      string    `json:"headline,omitempty"`
	Visibility    string    `json:"visibility"` // public, private
	CustomDomain  string    `json:"custom_domain,omitempty"`
	Theme         string    `json:"theme"` // light, dark, system
	CurrentStatus string    `json:"current_status"` // operational, degraded_performance, partial_outage, major_outage, under_maintenance
	IsActive      bool      `json:"is_active"`
	CreatedByID   string    `json:"created_by_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Enriched fields
	ProjectName          string `json:"project_name,omitempty"`
	CreatorName          string `json:"creator_name,omitempty"`
	ComponentsCount      int    `json:"components_count"`
	ActiveIncidentsCount int    `json:"active_incidents_count"`
}

// ComponentGroup aggregates related components into logical sections (e.g. Core APIs, Third-Party Providers)
type ComponentGroup struct {
	ID           int64     `json:"id"`
	PageID       int64     `json:"page_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	DisplayOrder int       `json:"display_order"`
	Collapsed    bool      `json:"collapsed"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Enriched fields
	Components      []StatusComponent `json:"components,omitempty"`
	ComponentsCount int               `json:"components_count"`
}

// StatusComponent represents an individual tracked service, API, or infrastructure component
type StatusComponent struct {
	ID           int64     `json:"id"`
	PageID       int64     `json:"page_id"`
	GroupID      *int64    `json:"group_id,omitempty"`
	CIID         *int64    `json:"ci_id,omitempty"` // Optional link to CMDB Configuration Item
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	Status       string    `json:"status"` // operational, degraded_performance, partial_outage, major_outage, under_maintenance
	DisplayOrder int       `json:"display_order"`
	IsMonitored  bool      `json:"is_monitored"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Enriched fields
	GroupName string `json:"group_name,omitempty"`
	CIName    string `json:"ci_name,omitempty"`
	CICode    string `json:"ci_code,omitempty"`
}
