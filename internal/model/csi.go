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

// CSI domain errors
var (
	ErrCSIInitiativeNotFound      = errors.New("continual improvement initiative not found")
	ErrCSIInitiativeTitleRequired = errors.New("initiative title is required")
	ErrInvalidCSIAlignment        = errors.New("invalid strategic alignment category")
	ErrInvalidCSIStatus           = errors.New("invalid CSI initiative status")
	ErrInvalidCSIPriority         = errors.New("invalid CSI initiative priority")
	ErrInvalidCSISource           = errors.New("invalid CSI initiative source")
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
