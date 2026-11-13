package model

import (
	"errors"
	"regexp"
	"strings"
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

// Lifecycle status constants for status page incidents
const (
	StatusIncidentInvestigating = "investigating"
	StatusIncidentIdentified    = "identified"
	StatusIncidentMonitoring    = "monitoring"
	StatusIncidentResolved      = "resolved"
)

// Impact level constants for status incidents
const (
	ImpactNone     = "none"
	ImpactMinor    = "minor"
	ImpactMajor    = "major"
	ImpactCritical = "critical"
)

// StatusIncident represents a publicly or internally broadcasted service incident
type StatusIncident struct {
	ID                     int64      `json:"id"`
	PageID                 int64      `json:"page_id"`
	IncidentID             *int64     `json:"incident_id,omitempty"` // Link to internal /v2/incidents
	Title                  string     `json:"title"`
	Status                 string     `json:"status"` // investigating, identified, monitoring, resolved
	Impact                 string     `json:"impact"` // none, minor, major, critical
	StartedAt              time.Time  `json:"started_at"`
	ResolvedAt             *time.Time `json:"resolved_at,omitempty"`
	AffectedComponentsJSON string     `json:"affected_components_json,omitempty"`
	CreatedByID            string     `json:"created_by_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Enriched fields
	IncidentNumber         *int                   `json:"incident_number,omitempty"`
	Updates                []StatusIncidentUpdate `json:"updates,omitempty"`
	AffectedComponentIDs   []int64                `json:"affected_component_ids,omitempty"`
	AffectedComponentNames []string               `json:"affected_component_names,omitempty"`
}

// StatusIncidentUpdate represents a chronological message post on an active status incident
type StatusIncidentUpdate struct {
	ID               int64     `json:"id"`
	StatusIncidentID int64     `json:"status_incident_id"`
	Status           string    `json:"status"` // investigating, identified, monitoring, resolved
	Message          string    `json:"message"`
	PostedByID       string    `json:"posted_by_id"`
	CreatedAt        time.Time `json:"created_at"`

	// Enriched fields
	PosterName string `json:"poster_name,omitempty"`
}

// Maintenance window status constants
const (
	MaintenanceStatusScheduled  = "scheduled"
	MaintenanceStatusInProgress = "in_progress"
	MaintenanceStatusCompleted  = "completed"
	MaintenanceStatusCancelled  = "cancelled"
)

// StatusMaintenanceWindow represents a scheduled maintenance window published on a status page
type StatusMaintenanceWindow struct {
	ID                     int64      `json:"id"`
	PageID                 int64      `json:"page_id"`
	Title                  string     `json:"title"`
	Description            string     `json:"description,omitempty"`
	Status                 string     `json:"status"` // scheduled, in_progress, completed, cancelled
	ScheduledStartAt       time.Time  `json:"scheduled_start_at"`
	ScheduledEndAt         time.Time  `json:"scheduled_end_at"`
	ActualStartAt          *time.Time `json:"actual_start_at,omitempty"`
	ActualEndAt            *time.Time `json:"actual_end_at,omitempty"`
	AffectedComponentsJSON string     `json:"affected_components_json,omitempty"`
	CreatedByID            string     `json:"created_by_id"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`

	// Enriched fields
	AffectedComponentIDs   []int64  `json:"affected_component_ids,omitempty"`
	AffectedComponentNames []string `json:"affected_component_names,omitempty"`
}

// Notification channel types for status page subscribers
const (
	SubscriberTypeEmail   = "email"
	SubscriberTypeWebhook = "webhook"
	SubscriberTypeSlack   = "slack"
)

// StatusSubscriber represents an external user or webhook endpoint subscribed to status page notifications
type StatusSubscriber struct {
	ID                       int64     `json:"id"`
	PageID                   int64     `json:"page_id"`
	SubscriberType           string    `json:"subscriber_type"` // email, webhook, slack
	TargetEndpoint           string    `json:"target_endpoint"` // email address, webhook url, or slack incoming webhook
	VerificationToken        string    `json:"verification_token,omitempty"`
	IsVerified               bool      `json:"is_verified"`
	SubscribedComponentsJSON string    `json:"subscribed_components_json,omitempty"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`

	// Enriched fields
	SubscribedComponentIDs []int64 `json:"subscribed_component_ids,omitempty"`
}

var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)

// NormalizeSlug converts an arbitrary name into a clean, URL-safe slug
func NormalizeSlug(input string) string {
	s := strings.ToLower(strings.TrimSpace(input))
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	s = reg.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

// ValidateSlug validates that a slug consists only of lowercase alphanumerics and hyphens
func ValidateSlug(slug string) error {
	trimmed := strings.TrimSpace(slug)
	if trimmed == "" {
		return ErrStatusPageSlugRequired
	}
	if len(trimmed) < 2 || len(trimmed) > 80 || !slugRegex.MatchString(trimmed) {
		return ErrInvalidStatusPageSlug
	}
	return nil
}

// ValidateComponentStatus verifies that a status is one of the supported component status levels
func ValidateComponentStatus(st string) error {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case StatusOperational, StatusDegradedPerformance, StatusPartialOutage, StatusMajorOutage, StatusUnderMaintenance:
		return nil
	default:
		return ErrInvalidComponentStatus
	}
}

// ValidateStatusIncidentStatus verifies that an incident status is supported
func ValidateStatusIncidentStatus(st string) error {
	switch strings.ToLower(strings.TrimSpace(st)) {
	case StatusIncidentInvestigating, StatusIncidentIdentified, StatusIncidentMonitoring, StatusIncidentResolved:
		return nil
	default:
		return ErrInvalidStatusIncidentSt
	}
}

// ValidateIncidentImpact verifies that an incident impact level is recognized
func ValidateIncidentImpact(impact string) error {
	switch strings.ToLower(strings.TrimSpace(impact)) {
	case ImpactNone, ImpactMinor, ImpactMajor, ImpactCritical:
		return nil
	default:
		return ErrInvalidIncidentImpact
	}
}

// ValidateSubscriberType verifies that a subscriber channel is recognized
func ValidateSubscriberType(subType string) error {
	switch strings.ToLower(strings.TrimSpace(subType)) {
	case SubscriberTypeEmail, SubscriberTypeWebhook, SubscriberTypeSlack:
		return nil
	default:
		return ErrInvalidSubscriberType
	}
}

// CreateStatusPageRequest payload for creating a status page
type CreateStatusPageRequest struct {
	WorkspaceID  int64  `json:"workspace_id"`
	ProjectID    *int64 `json:"project_id,omitempty"`
	Name         string `json:"name"`
	Slug         string `json:"slug,omitempty"`
	Description  string `json:"description,omitempty"`
	Headline     string `json:"headline,omitempty"`
	Visibility   string `json:"visibility,omitempty"` // public, private
	CustomDomain string `json:"custom_domain,omitempty"`
	Theme        string `json:"theme,omitempty"`
}

// UpdateStatusPageRequest payload for modifying an existing status page
type UpdateStatusPageRequest struct {
	Name         *string `json:"name,omitempty"`
	Slug         *string `json:"slug,omitempty"`
	Description  *string `json:"description,omitempty"`
	Headline     *string `json:"headline,omitempty"`
	Visibility   *string `json:"visibility,omitempty"`
	CustomDomain *string `json:"custom_domain,omitempty"`
	Theme        *string `json:"theme,omitempty"`
	CurrentStatus *string `json:"current_status,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// CreateComponentGroupRequest payload for defining a component group
type CreateComponentGroupRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	DisplayOrder int    `json:"display_order,omitempty"`
	Collapsed    bool   `json:"collapsed,omitempty"`
}

// CreateComponentRequest payload for creating a status component
type CreateComponentRequest struct {
	GroupID      *int64 `json:"group_id,omitempty"`
	CIID         *int64 `json:"ci_id,omitempty"`
	Name         string `json:"name"`
	Description  string `json:"description,omitempty"`
	Status       string `json:"status,omitempty"`
	DisplayOrder int    `json:"display_order,omitempty"`
	IsMonitored  bool   `json:"is_monitored,omitempty"`
}

// UpdateComponentRequest payload for modifying a status component
type UpdateComponentRequest struct {
	GroupID      *int64  `json:"group_id,omitempty"`
	CIID         *int64  `json:"ci_id,omitempty"`
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Status       *string `json:"status,omitempty"`
	DisplayOrder *int    `json:"display_order,omitempty"`
	IsMonitored  *bool   `json:"is_monitored,omitempty"`
}

// StatusPageFilter defines querying criteria for listing status pages
type StatusPageFilter struct {
	WorkspaceID int64  `json:"workspace_id"`
	ProjectID   *int64 `json:"project_id,omitempty"`
	Visibility  string `json:"visibility,omitempty"`
	Search      string `json:"search,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}
