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
