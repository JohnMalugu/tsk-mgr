package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrAuditActionRequired   = errors.New("audit action is required")
	ErrAuditCategoryRequired = errors.New("audit category is required")
	ErrInvalidAuditSeverity  = errors.New("invalid audit severity level")
)

// Audit severity levels
const (
	AuditSeverityInfo     = "info"
	AuditSeverityWarning  = "warning"
	AuditSeverityCritical = "critical"
)

// Audit category constants
const (
	AuditCategoryAuth          = "authentication"
	AuditCategoryAuthz         = "authorization"
	AuditCategoryDataAccess    = "data_access"
	AuditCategoryConfiguration = "configuration"
	AuditCategorySecurity      = "security"
	AuditCategoryBilling       = "billing"
	AuditCategorySystem        = "system"
)

// Standard audit action constants
const (
	AuditActionLogin             = "auth.login"
	AuditActionLogout            = "auth.logout"
	AuditActionFailedLogin       = "auth.failed_login"
	AuditActionPasswordReset     = "auth.password_reset"
	AuditActionPermissionGranted = "authz.permission_granted"
	AuditActionPermissionRevoked = "authz.permission_revoked"
	AuditActionRoleChanged       = "authz.role_changed"
	AuditActionDataExported      = "data.exported"
	AuditActionDataPurged        = "data.purged"
	AuditActionConfigUpdated     = "config.updated"
	AuditActionSecurityAlert     = "security.alert_triggered"
)

// Compliance framework identifiers
const (
	ComplianceSOC2     = "soc2"
	ComplianceGDPR     = "gdpr"
	ComplianceHIPAA    = "hipaa"
	ComplianceISO27001 = "iso27001"
	CompliancePCIDSS   = "pci_dss"
)

// ValidateAuditSeverity checks whether a severity string is valid
func ValidateAuditSeverity(sev string) error {
	switch strings.ToLower(strings.TrimSpace(sev)) {
	case AuditSeverityInfo, AuditSeverityWarning, AuditSeverityCritical:
		return nil
	default:
		return ErrInvalidAuditSeverity
	}
}

// AuditLog represents an immutable record of an administrative, security, or data mutation event
type AuditLog struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	ActorID     string    `json:"actor_id"`
	ActorName   string    `json:"actor_name"`
	ActorEmail  string    `json:"actor_email"`
	Action      string    `json:"action"`   // e.g. "auth.login", "task.create"
	Category    string    `json:"category"` // e.g. "authentication", "security"
	Severity    string    `json:"severity"` // "info", "warning", "critical"
	EntityType  string    `json:"entity_type,omitempty"`
	EntityID    string    `json:"entity_id,omitempty"`
	EntityName  string    `json:"entity_name,omitempty"`
	Description string    `json:"description"`
	BeforeState string    `json:"before_state,omitempty"` // JSON string snapshot
	AfterState  string    `json:"after_state,omitempty"`  // JSON string snapshot
	Metadata    string    `json:"metadata,omitempty"`     // JSON string
	IPAddress   string    `json:"ip_address,omitempty"`
	UserAgent   string    `json:"user_agent,omitempty"`
	Checksum    string    `json:"checksum"` // SHA256 integrity hash
	CreatedAt   time.Time `json:"created_at"`
}

// AuditRetentionPolicy governs how long audit records are preserved per workspace
type AuditRetentionPolicy struct {
	ID               int64      `json:"id"`
	WorkspaceID      int64      `json:"workspace_id"`
	RetentionDays    int        `json:"retention_days"` // e.g. 90, 365, 2555
	Frameworks       string     `json:"frameworks"`     // comma-separated compliance frameworks
	AutoPurgeEnabled bool       `json:"auto_purge_enabled"`
	LastPurgeAt      *time.Time `json:"last_purge_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// Alert status constants
const (
	AuditAlertStatusOpen         = "open"
	AuditAlertStatusAcknowledged = "acknowledged"
	AuditAlertStatusResolved     = "resolved"
	AuditAlertStatusDismissed    = "dismissed"
)

// AuditAlert represents a security anomaly or compliance policy violation
type AuditAlert struct {
	ID               int64      `json:"id"`
	WorkspaceID      int64      `json:"workspace_id"`
	Title            string     `json:"title"`
	Description      string     `json:"description"`
	Severity         string     `json:"severity"` // warning, critical
	Status           string     `json:"status"`   // open, acknowledged, resolved, dismissed
	TriggerEventID   *int64     `json:"trigger_event_id,omitempty"`
	ResolutionNotes  string     `json:"resolution_notes,omitempty"`
	ResolvedByUserID string     `json:"resolved_by_user_id,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Enriched fields
	TriggerAction      string `json:"trigger_action,omitempty"`
	TriggerActor       string `json:"trigger_actor,omitempty"`
	ResolvedByUsername string `json:"resolved_by_username,omitempty"`
}

// CreateAuditLogRequest payload for ingesting or recording an audit event
type CreateAuditLogRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	Action      string `json:"action"`
	Category    string `json:"category,omitempty"` // defaults to "system" if blank
	Severity    string `json:"severity,omitempty"` // defaults to "info" if blank
	EntityType  string `json:"entity_type,omitempty"`
	EntityID    string `json:"entity_id,omitempty"`
	EntityName  string `json:"entity_name,omitempty"`
	Description string `json:"description"`
	BeforeState string `json:"before_state,omitempty"` // JSON string
	AfterState  string `json:"after_state,omitempty"`  // JSON string
	Metadata    string `json:"metadata,omitempty"`     // JSON string
	IPAddress   string `json:"ip_address,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// AuditLogFilter provides querying, searching, and filtering options across audit logs
type AuditLogFilter struct {
	WorkspaceID int64   `json:"workspace_id"`
	ActorID     *string `json:"actor_id,omitempty"`
	Action      *string `json:"action,omitempty"`
	Category    *string `json:"category,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	EntityType  *string `json:"entity_type,omitempty"`
	EntityID    *string `json:"entity_id,omitempty"`
	Search      *string `json:"search,omitempty"`
	StartDate   *string `json:"start_date,omitempty"` // YYYY-MM-DD or RFC3339
	EndDate     *string `json:"end_date,omitempty"`   // YYYY-MM-DD or RFC3339
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}

// RetentionPolicyRequest payload for configuring workspace compliance data retention
type RetentionPolicyRequest struct {
	WorkspaceID      int64    `json:"workspace_id"`
	RetentionDays    int      `json:"retention_days"` // e.g. 90, 365, 2555
	Frameworks       []string `json:"frameworks,omitempty"`
	AutoPurgeEnabled bool     `json:"auto_purge_enabled"`
}

// AuditAlertFilter provides query filters for security and compliance anomaly alerts
type AuditAlertFilter struct {
	WorkspaceID int64   `json:"workspace_id"`
	Status      *string `json:"status,omitempty"`
	Severity    *string `json:"severity,omitempty"`
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}

// ResolveAlertRequest payload for acknowledging or resolving a security alert
type ResolveAlertRequest struct {
	Status          string `json:"status"` // acknowledged, resolved, dismissed
	ResolutionNotes string `json:"resolution_notes,omitempty"`
}

// AuditExport records an immutable export artifact for compliance submissions
type AuditExport struct {
	ID              int64     `json:"id"`
	WorkspaceID     int64     `json:"workspace_id"`
	Format          string    `json:"format"` // json, csv
	FilterSummary   string    `json:"filter_summary"`
	RecordCount     int       `json:"record_count"`
	Checksum        string    `json:"checksum"` // SHA256 of export contents
	CreatedByUserID string    `json:"created_by_user_id"`
	CreatedAt       time.Time `json:"created_at"`

	// Enriched fields
	CreatedByUsername string `json:"created_by_username,omitempty"`
}

// ExportAuditLogsRequest payload for generating an immutable compliance export
type ExportAuditLogsRequest struct {
	WorkspaceID int64          `json:"workspace_id"`
	Format      string         `json:"format,omitempty"` // json, csv (default: json)
	Filter      AuditLogFilter `json:"filter"`
}

// AuditMetrics aggregates log volume, severity distribution, and security events
type AuditMetrics struct {
	WorkspaceID         int64      `json:"workspace_id"`
	TotalLogs           int64      `json:"total_logs"`
	InfoCount           int64      `json:"info_count"`
	WarningCount        int64      `json:"warning_count"`
	CriticalCount       int64      `json:"critical_count"`
	OpenAlertsCount     int64      `json:"open_alerts_count"`
	AuthEventsCount     int64      `json:"auth_events_count"`
	SecurityEventsCount int64      `json:"security_events_count"`
	DataEventsCount     int64      `json:"data_events_count"`
	ConfigEventsCount   int64      `json:"config_events_count"`
	LastEventAt         *time.Time `json:"last_event_at,omitempty"`
}

// ComplianceSummary summarizes framework conformance and integrity state
type ComplianceSummary struct {
	WorkspaceID                 int64    `json:"workspace_id"`
	ActiveFrameworks            []string `json:"active_frameworks"`
	RetentionDays               int      `json:"retention_days"`
	AuditTrailIntegrityVerified bool     `json:"audit_trail_integrity_verified"`
	TotalEventsLogged           int64    `json:"total_events_logged"`
	TotalExportsCount           int      `json:"total_exports_count"`
	OpenCriticalAlerts          int      `json:"open_critical_alerts"`
	ComplianceScore             float64  `json:"compliance_score"` // 0.0 - 100.0
}
