package model

import (
	"errors"
	"strings"
	"time"
)

// Strategic alignment constants
const (
	CSIAlignmentReliability    = "reliability"
	CSIAlignmentCostReduction  = "cost_reduction"
	CSIAlignmentSecurity       = "security"
	CSIAlignmentVelocity       = "velocity"
	CSIAlignmentCompliance     = "compliance"
	CSIAlignmentUserExperience = "user_experience"
)

// Initiative status constants
const (
	CSIStatusIdentified = "identified"
	CSIStatusApproved   = "approved"
	CSIStatusInProgress = "in_progress"
	CSIStatusOnHold     = "on_hold"
	CSIStatusCompleted  = "completed"
	CSIStatusCancelled  = "cancelled"
)

// Initiative priority constants
const (
	CSIPriorityLow    = "low"
	CSIPriorityMedium = "medium"
	CSIPriorityHigh   = "high"
	CSIPriorityUrgent = "urgent"
)

// Initiative source constants
const (
	CSISourceIncidentRCA         = "incident_rca"
	CSISourceProblemKEDB         = "problem_kedb"
	CSISourceCapacityBottleneck  = "capacity_bottleneck"
	CSISourceSPOFRemediation     = "spof_remediation"
	CSISourceAuditFinding        = "audit_finding"
	CSISourceUserFeedback        = "user_feedback"
	CSISourceFinOpsReview        = "finops_review"
	CSISourceOther               = "other"
)

// Metric unit constants
const (
	CSIMetricUnitPercent      = "percent"
	CSIMetricUnitMilliseconds = "milliseconds"
	CSIMetricUnitSeconds      = "seconds"
	CSIMetricUnitDollars      = "dollars"
	CSIMetricUnitHours        = "hours"
	CSIMetricUnitCount        = "count"
)

// CSI domain errors
var (
	ErrCSIInitiativeNotFound          = errors.New("continual improvement initiative not found")
	ErrCSIInitiativeTitleRequired     = errors.New("initiative title is required")
	ErrInvalidCSIAlignment            = errors.New("invalid strategic alignment category")
	ErrInvalidCSIStatus               = errors.New("invalid CSI initiative status")
	ErrInvalidCSIPriority             = errors.New("invalid CSI initiative priority")
	ErrInvalidCSISource               = errors.New("invalid CSI initiative source")
	ErrCSIBenchmarkNotFound           = errors.New("CSI benchmark metric not found")
	ErrCSIBenchmarkMetricNameRequired = errors.New("benchmark metric name is required")
	ErrInvalidCSIMetricUnit           = errors.New("invalid CSI benchmark metric unit")
)

// ValidateCSIAlignment verifies if an alignment category is recognized
func ValidateCSIAlignment(a string) error {
	switch a {
	case CSIAlignmentReliability, CSIAlignmentCostReduction, CSIAlignmentSecurity,
		CSIAlignmentVelocity, CSIAlignmentCompliance, CSIAlignmentUserExperience:
		return nil
	default:
		return ErrInvalidCSIAlignment
	}
}

// ValidateCSIStatus verifies if status is valid
func ValidateCSIStatus(s string) error {
	switch s {
	case CSIStatusIdentified, CSIStatusApproved, CSIStatusInProgress,
		CSIStatusOnHold, CSIStatusCompleted, CSIStatusCancelled:
		return nil
	default:
		return ErrInvalidCSIStatus
	}
}

// ValidateCSIPriority verifies if priority is recognized
func ValidateCSIPriority(p string) error {
	switch p {
	case CSIPriorityLow, CSIPriorityMedium, CSIPriorityHigh, CSIPriorityUrgent:
		return nil
	default:
		return ErrInvalidCSIPriority
	}
}

// ValidateCSISource verifies if source is recognized
func ValidateCSISource(s string) error {
	switch s {
	case CSISourceIncidentRCA, CSISourceProblemKEDB, CSISourceCapacityBottleneck,
		CSISourceSPOFRemediation, CSISourceAuditFinding, CSISourceUserFeedback,
		CSISourceFinOpsReview, CSISourceOther:
		return nil
	default:
		return ErrInvalidCSISource
	}
}

// ValidateCSIMetricUnit verifies if a unit of measure is supported
func ValidateCSIMetricUnit(u string) error {
	switch u {
	case CSIMetricUnitPercent, CSIMetricUnitMilliseconds, CSIMetricUnitSeconds,
		CSIMetricUnitDollars, CSIMetricUnitHours, CSIMetricUnitCount:
		return nil
	default:
		return ErrInvalidCSIMetricUnit
	}
}

// CSIInitiative represents an improvement entry in the ITIL Continual Improvement Register (CIR)
type CSIInitiative struct {
	ID                   string     `json:"id"`
	WorkspaceID          string     `json:"workspace_id"`
	InitiativeCode       string     `json:"initiative_code"`
	Title                string     `json:"title"`
	Description          string     `json:"description,omitempty"`
	StrategicAlignment   string     `json:"strategic_alignment"`
	Priority             string     `json:"priority"`
	Status               string     `json:"status"`
	Source               string     `json:"source"`
	SourceReferenceID    string     `json:"source_reference_id,omitempty"`
	CIID                 string     `json:"ci_id,omitempty"`
	OwnerID              string     `json:"owner_id,omitempty"`
	TargetCompletionDate time.Time  `json:"target_completion_date"`
	CompletedAt          *time.Time `json:"completed_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// Validate validates the CSIInitiative entity
func (i *CSIInitiative) Validate() error {
	if strings.TrimSpace(i.Title) == "" {
		return ErrCSIInitiativeTitleRequired
	}
	if err := ValidateCSIAlignment(i.StrategicAlignment); err != nil {
		return err
	}
	if err := ValidateCSIPriority(i.Priority); err != nil {
		return err
	}
	if err := ValidateCSIStatus(i.Status); err != nil {
		return err
	}
	if err := ValidateCSISource(i.Source); err != nil {
		return err
	}
	return nil
}

// CSIBenchmark tracks baseline vs current vs target metrics for an improvement initiative
type CSIBenchmark struct {
	ID            string     `json:"id"`
	WorkspaceID   string     `json:"workspace_id"`
	InitiativeID  string     `json:"initiative_id"`
	MetricName    string     `json:"metric_name"`
	UnitOfMeasure string     `json:"unit_of_measure"`
	BaselineValue float64    `json:"baseline_value"`
	CurrentValue  float64    `json:"current_value"`
	TargetValue   float64    `json:"target_value"`
	VerifiedAt    *time.Time `json:"verified_at,omitempty"`
	Notes         string     `json:"notes,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Validate validates the CSIBenchmark entity
func (b *CSIBenchmark) Validate() error {
	if strings.TrimSpace(b.MetricName) == "" {
		return ErrCSIBenchmarkMetricNameRequired
	}
	if err := ValidateCSIMetricUnit(b.UnitOfMeasure); err != nil {
		return err
	}
	return nil
}

// HasAchievedTarget evaluates whether CurrentValue meets or surpasses TargetValue
func (b *CSIBenchmark) HasAchievedTarget() bool {
	// If baseline was higher than target (e.g. latency, error rate, cost), lower or equal is better
	if b.BaselineValue > b.TargetValue {
		return b.CurrentValue <= b.TargetValue
	}
	// If baseline was lower than target (e.g. availability, throughput, coverage), higher or equal is better
	return b.CurrentValue >= b.TargetValue
}
