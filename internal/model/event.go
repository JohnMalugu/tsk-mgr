package model

import (
	"time"
)

// Event severity constants
const (
	EventSeverityCritical = "critical"
	EventSeverityHigh     = "high"
	EventSeverityMedium   = "medium"
	EventSeverityLow      = "low"
	EventSeverityInfo     = "info"
)

// Event type classification constants
const (
	EventTypeInformational = "informational"
	EventTypeWarning       = "warning"
	EventTypeException     = "exception"
)

// Event monitoring source constants
const (
	EventSourceDatadog    = "datadog"
	EventSourcePrometheus = "prometheus"
	EventSourceCloudWatch = "cloudwatch"
	EventSourceGrafana    = "grafana"
	EventSourcePagerDuty  = "pagerduty"
	EventSourceCustom     = "custom"
)

// IsValidEventSeverity validates event severity
func IsValidEventSeverity(sev string) bool {
	switch sev {
	case EventSeverityCritical, EventSeverityHigh, EventSeverityMedium, EventSeverityLow, EventSeverityInfo:
		return true
	default:
		return false
	}
}

// IsValidEventType validates event classification type
func IsValidEventType(et string) bool {
	switch et {
	case EventTypeInformational, EventTypeWarning, EventTypeException:
		return true
	default:
		return false
	}
}

// IsValidEventSource validates monitoring source
func IsValidEventSource(src string) bool {
	switch src {
	case EventSourceDatadog, EventSourcePrometheus, EventSourceCloudWatch,
		EventSourceGrafana, EventSourcePagerDuty, EventSourceCustom:
		return true
	default:
		return false
	}
}

// EventRecord represents an individual ingested telemetry or monitoring event
type EventRecord struct {
	ID               string    `json:"id"`
	WorkspaceID      string    `json:"workspace_id"`
	Source           string    `json:"source"`
	EventType        string    `json:"event_type"`
	Severity         string    `json:"severity"`
	CIID             string    `json:"ci_id,omitempty"`
	Resource         string    `json:"resource"`
	Message          string    `json:"message"`
	Fingerprint      string    `json:"fingerprint"`
	DeduplicationKey string    `json:"deduplication_key"`
	RawPayload       string    `json:"raw_payload,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	CreatedAt        time.Time `json:"created_at"`
}

// Alert status lifecycle constants
const (
	AlertStatusOpen         = "open"
	AlertStatusAcknowledged = "acknowledged"
	AlertStatusSuppressed   = "suppressed"
	AlertStatusResolved     = "resolved"
	AlertStatusEscalated    = "escalated_to_incident"
)

// Alert priority constants
const (
	AlertPriorityP1 = "p1"
	AlertPriorityP2 = "p2"
	AlertPriorityP3 = "p3"
	AlertPriorityP4 = "p4"
)

// IsValidAlertStatus validates alert lifecycle status
func IsValidAlertStatus(status string) bool {
	switch status {
	case AlertStatusOpen, AlertStatusAcknowledged, AlertStatusSuppressed, AlertStatusResolved, AlertStatusEscalated:
		return true
	default:
		return false
	}
}

// IsValidAlertPriority validates alert priority rating
func IsValidAlertPriority(prio string) bool {
	switch prio {
	case AlertPriorityP1, AlertPriorityP2, AlertPriorityP3, AlertPriorityP4:
		return true
	default:
		return false
	}
}

// CorrelatedAlert represents an aggregated, deduplicated actionable alert clustered from raw events
type CorrelatedAlert struct {
	ID             string     `json:"id"`
	WorkspaceID    string     `json:"workspace_id"`
	AlertCode      string     `json:"alert_code"`
	Title          string     `json:"title"`
	Description    string     `json:"description,omitempty"`
	Status         string     `json:"status"`
	Severity       string     `json:"severity"`
	Priority       string     `json:"priority"`
	CIID           string     `json:"ci_id,omitempty"`
	Resource       string     `json:"resource"`
	Fingerprint    string     `json:"fingerprint"`
	EventCount     int        `json:"event_count"`
	FirstSeenAt    time.Time  `json:"first_seen_at"`
	LastSeenAt     time.Time  `json:"last_seen_at"`
	AcknowledgedBy string     `json:"acknowledged_by,omitempty"`
	AcknowledgedAt *time.Time `json:"acknowledged_at,omitempty"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	IncidentID     string     `json:"incident_id,omitempty"`
	RuleID         string     `json:"rule_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Correlation rule action constants
const (
	RuleActionCreateAlert       = "create_alert"
	RuleActionEscalateIncident = "escalate_to_incident"
	RuleActionPageOnCall        = "page_oncall"
)

// IsValidRuleAction validates rule action type
func IsValidRuleAction(action string) bool {
	switch action {
	case RuleActionCreateAlert, RuleActionEscalateIncident, RuleActionPageOnCall:
		return true
	default:
		return false
	}
}

// CorrelationRule defines event pattern matching, time window clustering, and auto-action policies
type CorrelationRule struct {
	ID                string    `json:"id"`
	WorkspaceID       string    `json:"workspace_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description,omitempty"`
	SourcePattern     string    `json:"source_pattern"`
	SeverityPattern   string    `json:"severity_pattern"`
	TimeWindowSeconds int       `json:"time_window_seconds"`
	ThresholdCount    int       `json:"threshold_count"`
	GroupingFields    string    `json:"grouping_fields"`
	ActionType        string    `json:"action_type"`
	IsActive          bool      `json:"is_active"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// EventSuppression defines noise-silencing criteria and scheduled maintenance blackout windows
type EventSuppression struct {
	ID              string    `json:"id"`
	WorkspaceID     string    `json:"workspace_id"`
	Name            string    `json:"name"`
	Reason          string    `json:"reason"`
	CIID            string    `json:"ci_id,omitempty"`
	ResourcePattern string    `json:"resource_pattern,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	IsActive        bool      `json:"is_active"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
