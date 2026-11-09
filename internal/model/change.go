package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrChangeTitleRequired    = errors.New("change request title is required")
	ErrInvalidChangeType      = errors.New("invalid change type: must be standard, normal, or emergency")
	ErrInvalidChangeRiskLevel = errors.New("invalid change risk level: must be low, medium, high, or critical")
)

// Change types according to ITIL
const (
	ChangeTypeStandard  = "standard"
	ChangeTypeNormal    = "normal"
	ChangeTypeEmergency = "emergency"
)

// Change risk levels
const (
	ChangeRiskLow      = "low"
	ChangeRiskMedium   = "medium"
	ChangeRiskHigh     = "high"
	ChangeRiskCritical = "critical"
)

// Change impact levels
const (
	ChangeImpactLow    = "low"
	ChangeImpactMedium = "medium"
	ChangeImpactHigh   = "high"
)

// ValidateChangeType checks if a change type string is valid
func ValidateChangeType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case ChangeTypeStandard, ChangeTypeNormal, ChangeTypeEmergency:
		return nil
	default:
		return ErrInvalidChangeType
	}
}

// ValidateChangeRiskLevel checks if a risk level is valid
func ValidateChangeRiskLevel(r string) error {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case ChangeRiskLow, ChangeRiskMedium, ChangeRiskHigh, ChangeRiskCritical:
		return nil
	default:
		return ErrInvalidChangeRiskLevel
	}
}

// ChangeRequest represents a formal ITIL change proposal in a workspace
type ChangeRequest struct {
	ID                 int64      `json:"id"`
	WorkspaceID        int64      `json:"workspace_id"`
	ProjectID          *int64     `json:"project_id,omitempty"`
	ChangeNumber       string     `json:"change_number"` // sequential e.g. "CHG-1"
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ChangeType         string     `json:"change_type"`  // standard, normal, emergency
	RiskLevel          string     `json:"risk_level"`   // low, medium, high, critical
	ImpactLevel        string     `json:"impact_level"` // low, medium, high
	Status             string     `json:"status"`
	ReasonForChange    string     `json:"reason_for_change"`
	ImplementationPlan string     `json:"implementation_plan"`
	RollbackPlan       string     `json:"rollback_plan"`
	TestPlan           string     `json:"test_plan"`
	RequestedByUserID  string     `json:"requested_by_user_id"`
	AssignedToUserID   string     `json:"assigned_to_user_id,omitempty"`
	PlannedStartAt     *time.Time `json:"planned_start_at,omitempty"`
	PlannedEndAt       *time.Time `json:"planned_end_at,omitempty"`
	ActualStartAt      *time.Time `json:"actual_start_at,omitempty"`
	ActualEndAt        *time.Time `json:"actual_end_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`

	// Enriched fields
	RequesterName  string `json:"requester_name,omitempty"`
	AssigneeName   string `json:"assignee_name,omitempty"`
	ProjectName    string `json:"project_name,omitempty"`
	ApprovalsCount int    `json:"approvals_count,omitempty"`
	ApprovedCount  int    `json:"approved_count,omitempty"`
	TasksCount     int    `json:"tasks_count,omitempty"`
}
