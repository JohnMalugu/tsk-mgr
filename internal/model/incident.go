package model

import "time"

// Incident severity levels
const (
	IncidentSev0 = "sev0" // Critical: full outage, severe business impact
	IncidentSev1 = "sev1" // High: core functionality impaired, broad customer impact
	IncidentSev2 = "sev2" // Medium: non-critical service degraded, workaround available
	IncidentSev3 = "sev3" // Low: minor glitch, cosmetic issue, or low business impact
)

// Incident lifecycle statuses
const (
	IncidentStatusReported      = "reported"
	IncidentStatusInvestigating = "investigating"
	IncidentStatusIdentified    = "identified"
	IncidentStatusMonitoring    = "monitoring"
	IncidentStatusResolved      = "resolved"
	IncidentStatusClosed        = "closed"
)

// Postmortem documentation workflow statuses
const (
	PostmortemNotRequired = "not_required"
	PostmortemPending     = "pending"
	PostmortemInReview    = "in_review"
	PostmortemCompleted   = "completed"
	PostmortemWaived      = "waived"
)

// Incident represents an operational incident, outage, or critical production issue
type Incident struct {
	ID               int64      `json:"id"`
	WorkspaceID      int64      `json:"workspace_id"`
	ProjectID        *int       `json:"project_id,omitempty"`
	IncidentNumber   int        `json:"incident_number"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary"`
	Severity         string     `json:"severity"` // sev0, sev1, sev2, sev3
	Status           string     `json:"status"`   // reported, investigating, identified, monitoring, resolved, closed
	CommanderID      *string    `json:"commander_id,omitempty"`
	ScribeID         *string    `json:"scribe_id,omitempty"`
	ReportedByID     string     `json:"reported_by_id"`
	PostmortemStatus string     `json:"postmortem_status"` // not_required, pending, in_review, completed, waived
	PostmortemDocID  *int64     `json:"postmortem_doc_id,omitempty"`
	RootCause        string     `json:"root_cause,omitempty"`
	ImpactedServices []string   `json:"impacted_services,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at,omitempty"`
	MitigatedAt      *time.Time `json:"mitigated_at,omitempty"`
	ResolvedAt       *time.Time `json:"resolved_at,omitempty"`
	ClosedAt         *time.Time `json:"closed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`

	// Enriched fields
	CommanderUsername        string `json:"commander_username,omitempty"`
	ScribeUsername           string `json:"scribe_username,omitempty"`
	ReporterUsername         string `json:"reporter_username,omitempty"`
	ProjectName              string `json:"project_name,omitempty"`
	DurationMinutes          *int   `json:"duration_minutes,omitempty"`
	TimeToAcknowledgeMinutes *int   `json:"time_to_acknowledge_minutes,omitempty"`
	TimeToResolveMinutes     *int   `json:"time_to_resolve_minutes,omitempty"`
}

// Timeline event types for incident chronological log
const (
	TimelineEventStatusChange          = "status_change"
	TimelineEventSeverityChange        = "severity_change"
	TimelineEventLeadAssigned          = "lead_assigned"
	TimelineEventMitigation            = "mitigation"
	TimelineEventInvestigation         = "investigation"
	TimelineEventCustomerCommunication = "customer_communication"
)

// IncidentTimeline represents a single chronological event or log entry during an incident
type IncidentTimeline struct {
	ID         int64     `json:"id"`
	IncidentID int64     `json:"incident_id"`
	AuthorID   string    `json:"author_id"`
	EventType  string    `json:"event_type"` // status_change, severity_change, lead_assigned, mitigation, investigation, customer_communication
	Message    string    `json:"message"`
	Metadata   string    `json:"metadata,omitempty"` // JSON payload
	CreatedAt  time.Time `json:"created_at"`

	// Enriched fields
	AuthorUsername string `json:"author_username,omitempty"`
}

// Incident task classification types
const (
	IncidentTaskTypeMitigation         = "mitigation"
	IncidentTaskTypePreventative       = "preventative"
	IncidentTaskTypeDetection          = "detection"
	IncidentTaskTypeProcessImprovement = "process_improvement"
)

// IncidentTaskLink establishes a cross-reference between an incident and remediation tasks
type IncidentTaskLink struct {
	ID         int64     `json:"id"`
	IncidentID int64     `json:"incident_id"`
	TaskID     int       `json:"task_id"`
	TaskType   string    `json:"task_type"` // mitigation, preventative, detection, process_improvement
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`

	// Enriched fields
	TaskTitle    string `json:"task_title,omitempty"`
	TaskStatus   string `json:"task_status,omitempty"`
	TaskPriority string `json:"task_priority,omitempty"`
}

// CreateIncidentRequest defines payload to declare a new incident
type CreateIncidentRequest struct {
	WorkspaceID      int64      `json:"workspace_id"`
	ProjectID        *int       `json:"project_id,omitempty"`
	Title            string     `json:"title"`
	Summary          string     `json:"summary,omitempty"`
	Severity         string     `json:"severity"` // sev0, sev1, sev2, sev3
	CommanderID      *string    `json:"commander_id,omitempty"`
	ScribeID         *string    `json:"scribe_id,omitempty"`
	ImpactedServices []string   `json:"impacted_services,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
}

// UpdateIncidentRequest defines parameters to modify an active incident
type UpdateIncidentRequest struct {
	Title            *string   `json:"title,omitempty"`
	Summary          *string   `json:"summary,omitempty"`
	ProjectID        *int      `json:"project_id,omitempty"`
	CommanderID      *string   `json:"commander_id,omitempty"`
	ScribeID         *string   `json:"scribe_id,omitempty"`
	ImpactedServices *[]string `json:"impacted_services,omitempty"`
	RootCause        *string   `json:"root_cause,omitempty"`
}

// ChangeIncidentStatusRequest defines payload to transition incident lifecycle
type ChangeIncidentStatusRequest struct {
	Status  string `json:"status"` // reported, investigating, identified, monitoring, resolved, closed
	Message string `json:"message,omitempty"`
}

// ChangeIncidentSeverityRequest defines payload to escalate or de-escalate severity
type ChangeIncidentSeverityRequest struct {
	Severity string `json:"severity"` // sev0, sev1, sev2, sev3
	Reason   string `json:"reason,omitempty"`
}

// AssignIncidentRolesRequest defines payload to appoint commander and scribe
type AssignIncidentRolesRequest struct {
	CommanderID *string `json:"commander_id,omitempty"`
	ScribeID    *string `json:"scribe_id,omitempty"`
}

// IncidentFilter specifies search and filtering criteria for incident list queries
type IncidentFilter struct {
	WorkspaceID  int64      `json:"workspace_id"`
	ProjectID    *int       `json:"project_id,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	Status       string     `json:"status,omitempty"`
	CommanderID  string     `json:"commander_id,omitempty"`
	ReportedByID string     `json:"reported_by_id,omitempty"`
	Search       string     `json:"search,omitempty"`
	From         *time.Time `json:"from,omitempty"`
	To           *time.Time `json:"to,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// CreateIncidentTimelineRequest defines payload to record an operational update on the incident timeline
type CreateIncidentTimelineRequest struct {
	EventType string `json:"event_type"` // status_change, severity_change, lead_assigned, mitigation, investigation, customer_communication
	Message   string `json:"message"`
	Metadata  string `json:"metadata,omitempty"` // JSON string
}

// LinkIncidentTaskRequest defines payload to associate a task to an incident
type LinkIncidentTaskRequest struct {
	TaskID   int    `json:"task_id"`
	TaskType string `json:"task_type,omitempty"` // mitigation, preventative, detection, process_improvement
}




