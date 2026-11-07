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
