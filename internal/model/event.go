package model

import (
	"errors"
	"strings"
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

// Event action type constants
const (
	ActionTriggerIncident   = "trigger_incident"
	ActionTriggerOnCallPage = "trigger_oncall_page"
	ActionRunRemediation    = "run_remediation"
	ActionSendWebhook       = "send_webhook"
)

// Event action execution status constants
const (
	ActionStatusPending  = "pending"
	ActionStatusExecuted = "executed"
	ActionStatusFailed   = "failed"
)

// IsValidActionType validates event action type
func IsValidActionType(at string) bool {
	switch at {
	case ActionTriggerIncident, ActionTriggerOnCallPage, ActionRunRemediation, ActionSendWebhook:
		return true
	default:
		return false
	}
}

// IsValidActionStatus validates event action execution status
func IsValidActionStatus(status string) bool {
	switch status {
	case ActionStatusPending, ActionStatusExecuted, ActionStatusFailed:
		return true
	default:
		return false
	}
}

// EventAction records automated downstream dispatching or remediation triggered by an alert
type EventAction struct {
	ID          string     `json:"id"`
	WorkspaceID string     `json:"workspace_id"`
	AlertID     string     `json:"alert_id"`
	ActionType  string     `json:"action_type"`
	Status      string     `json:"status"`
	OutputLogs  string     `json:"output_logs,omitempty"`
	ExecutedAt  *time.Time `json:"executed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// Request DTOs and validation rules for Event Management API

// IngestEventRequest represents payload for receiving a raw telemetry event
type IngestEventRequest struct {
	WorkspaceID      string     `json:"workspace_id"`
	Source           string     `json:"source"`
	EventType        string     `json:"event_type"`
	Severity         string     `json:"severity"`
	CIID             string     `json:"ci_id,omitempty"`
	Resource         string     `json:"resource"`
	Message          string     `json:"message"`
	Fingerprint      string     `json:"fingerprint,omitempty"`
	DeduplicationKey string     `json:"deduplication_key,omitempty"`
	RawPayload       string     `json:"raw_payload,omitempty"`
	Timestamp        *time.Time `json:"timestamp,omitempty"`
}

func (r *IngestEventRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.Source) == "" {
		return errors.New("source is required")
	}
	if strings.TrimSpace(r.Resource) == "" {
		return errors.New("resource is required")
	}
	if strings.TrimSpace(r.Message) == "" {
		return errors.New("message is required")
	}
	if r.Severity != "" && !IsValidEventSeverity(r.Severity) {
		return errors.New("invalid severity")
	}
	if r.EventType != "" && !IsValidEventType(r.EventType) {
		return errors.New("invalid event_type")
	}
	return nil
}

// UpdateAlertStatusRequest payload for updating an alert lifecycle state
type UpdateAlertStatusRequest struct {
	Status         string `json:"status"`
	AcknowledgedBy string `json:"acknowledged_by,omitempty"`
	IncidentID     string `json:"incident_id,omitempty"`
}

func (r *UpdateAlertStatusRequest) Validate() error {
	if !IsValidAlertStatus(r.Status) {
		return errors.New("invalid alert status")
	}
	return nil
}

// CreateCorrelationRuleRequest payload for configuring event correlation rules
type CreateCorrelationRuleRequest struct {
	WorkspaceID       string `json:"workspace_id"`
	Name              string `json:"name"`
	Description       string `json:"description,omitempty"`
	SourcePattern     string `json:"source_pattern"`
	SeverityPattern   string `json:"severity_pattern"`
	TimeWindowSeconds int    `json:"time_window_seconds"`
	ThresholdCount    int    `json:"threshold_count"`
	GroupingFields    string `json:"grouping_fields"`
	ActionType        string `json:"action_type"`
	IsActive          bool   `json:"is_active"`
}

func (r *CreateCorrelationRuleRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.TimeWindowSeconds <= 0 {
		return errors.New("time_window_seconds must be positive")
	}
	if r.ThresholdCount <= 0 {
		return errors.New("threshold_count must be positive")
	}
	if r.ActionType != "" && !IsValidRuleAction(r.ActionType) {
		return errors.New("invalid action_type")
	}
	return nil
}

// UpdateCorrelationRuleRequest payload for modifying a correlation rule
type UpdateCorrelationRuleRequest struct {
	Name              *string `json:"name,omitempty"`
	Description       *string `json:"description,omitempty"`
	SourcePattern     *string `json:"source_pattern,omitempty"`
	SeverityPattern   *string `json:"severity_pattern,omitempty"`
	TimeWindowSeconds *int    `json:"time_window_seconds,omitempty"`
	ThresholdCount    *int    `json:"threshold_count,omitempty"`
	GroupingFields    *string `json:"grouping_fields,omitempty"`
	ActionType        *string `json:"action_type,omitempty"`
	IsActive          *bool   `json:"is_active,omitempty"`
}

func (r *UpdateCorrelationRuleRequest) Validate() error {
	if r.TimeWindowSeconds != nil && *r.TimeWindowSeconds <= 0 {
		return errors.New("time_window_seconds must be positive")
	}
	if r.ThresholdCount != nil && *r.ThresholdCount <= 0 {
		return errors.New("threshold_count must be positive")
	}
	if r.ActionType != nil && !IsValidRuleAction(*r.ActionType) {
		return errors.New("invalid action_type")
	}
	return nil
}

// CreateEventSuppressionRequest payload for scheduling blackout / suppression windows
type CreateEventSuppressionRequest struct {
	WorkspaceID     string    `json:"workspace_id"`
	Name            string    `json:"name"`
	Reason          string    `json:"reason"`
	CIID            string    `json:"ci_id,omitempty"`
	ResourcePattern string    `json:"resource_pattern,omitempty"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
}

func (r *CreateEventSuppressionRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if strings.TrimSpace(r.Reason) == "" {
		return errors.New("reason is required")
	}
	if r.EndTime.Before(r.StartTime) {
		return errors.New("end_time cannot be before start_time")
	}
	return nil
}

// TriggerEventActionRequest payload for manually or automatically dispatching alert actions
type TriggerEventActionRequest struct {
	WorkspaceID string `json:"workspace_id"`
	AlertID     string `json:"alert_id"`
	ActionType  string `json:"action_type"`
	OutputLogs  string `json:"output_logs,omitempty"`
}

func (r *TriggerEventActionRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.AlertID) == "" {
		return errors.New("alert_id is required")
	}
	if !IsValidActionType(r.ActionType) {
		return errors.New("invalid action_type")
	}
	return nil
}
