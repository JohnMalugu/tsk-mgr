package model

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrChangeTitleRequired      = errors.New("change request title is required")
	ErrInvalidChangeType        = errors.New("invalid change type: must be standard, normal, or emergency")
	ErrInvalidChangeRiskLevel   = errors.New("invalid change risk level: must be low, medium, high, or critical")
	ErrInvalidChangeStatus      = errors.New("invalid change request status")
	ErrInvalidChangeTransition  = errors.New("illegal change request status transition")
)

// Change status constants
const (
	ChangeStatusDraft        = "draft"
	ChangeStatusSubmitted    = "submitted"
	ChangeStatusCABReview    = "cab_review"
	ChangeStatusApproved     = "approved"
	ChangeStatusScheduled    = "scheduled"
	ChangeStatusImplementing = "implementing"
	ChangeStatusCompleted    = "completed"
	ChangeStatusFailed       = "failed"
	ChangeStatusRolledBack   = "rolled_back"
	ChangeStatusCancelled    = "cancelled"
)

// ValidateChangeStatus checks if a status string is valid
func ValidateChangeStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case ChangeStatusDraft, ChangeStatusSubmitted, ChangeStatusCABReview, ChangeStatusApproved,
		ChangeStatusScheduled, ChangeStatusImplementing, ChangeStatusCompleted,
		ChangeStatusFailed, ChangeStatusRolledBack, ChangeStatusCancelled:
		return nil
	default:
		return ErrInvalidChangeStatus
	}
}

// ValidateChangeTransition checks whether transition from current to target status is permitted
func ValidateChangeTransition(current, target, changeType string) error {
	current = strings.ToLower(strings.TrimSpace(current))
	target = strings.ToLower(strings.TrimSpace(target))
	changeType = strings.ToLower(strings.TrimSpace(changeType))

	if current == target {
		return nil
	}

	// Always allow cancelling unless already in a terminal state
	if target == ChangeStatusCancelled {
		if current == ChangeStatusCompleted || current == ChangeStatusRolledBack || current == ChangeStatusCancelled {
			return ErrInvalidChangeTransition
		}
		return nil
	}

	switch current {
	case ChangeStatusDraft:
		if target == ChangeStatusSubmitted {
			return nil
		}
	case ChangeStatusSubmitted:
		if changeType == ChangeTypeStandard && (target == ChangeStatusApproved || target == ChangeStatusScheduled) {
			return nil
		}
		if changeType == ChangeTypeEmergency && (target == ChangeStatusApproved || target == ChangeStatusImplementing) {
			return nil
		}
		if target == ChangeStatusCABReview {
			return nil
		}
	case ChangeStatusCABReview:
		if target == ChangeStatusApproved || target == ChangeStatusDraft {
			return nil
		}
	case ChangeStatusApproved:
		if target == ChangeStatusScheduled || target == ChangeStatusImplementing {
			return nil
		}
	case ChangeStatusScheduled:
		if target == ChangeStatusImplementing {
			return nil
		}
	case ChangeStatusImplementing:
		if target == ChangeStatusCompleted || target == ChangeStatusFailed || target == ChangeStatusRolledBack {
			return nil
		}
	case ChangeStatusFailed:
		if target == ChangeStatusRolledBack {
			return nil
		}
	}

	return ErrInvalidChangeTransition
}

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

// Approval decision constants
const (
	ApprovalDecisionPending   = "pending"
	ApprovalDecisionApproved  = "approved"
	ApprovalDecisionRejected  = "rejected"
	ApprovalDecisionAbstained = "abstained"
)

// ChangeApproval records a formal CAB voting review for a change request
type ChangeApproval struct {
	ID         int64      `json:"id"`
	ChangeID   int64      `json:"change_id"`
	ReviewerID string     `json:"reviewer_id"`
	Decision   string     `json:"decision"` // pending, approved, rejected, abstained
	Comments   string     `json:"comments,omitempty"`
	DecidedAt  *time.Time `json:"decided_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`

	// Enriched fields
	ReviewerName  string `json:"reviewer_name,omitempty"`
	ReviewerEmail string `json:"reviewer_email,omitempty"`
}

// Task implementation phases
const (
	ChangePhasePreImplementation  = "pre_implementation"
	ChangePhaseImplementation     = "implementation"
	ChangePhasePostImplementation = "post_implementation"
)

// ChangeTaskLink associates an execution or verification task with a change request
type ChangeTaskLink struct {
	ID        int64     `json:"id"`
	ChangeID  int64     `json:"change_id"`
	TaskID    int       `json:"task_id"`
	Phase     string    `json:"phase"` // pre_implementation, implementation, post_implementation
	CreatedAt time.Time `json:"created_at"`

	// Enriched fields
	TaskTitle  string `json:"task_title,omitempty"`
	TaskStatus string `json:"task_status,omitempty"`
}

// PIR outcome constants
const (
	PIROutcomeSuccessful           = "successful"
	PIROutcomeSuccessfulWithIssues = "successful_with_issues"
	PIROutcomeFailedRolledBack     = "failed_rolled_back"
	PIROutcomeCancelled            = "cancelled"
)

// ChangePIR represents a Post-Implementation Review examining deployment outcome and root cause
type ChangePIR struct {
	ID               int64     `json:"id"`
	ChangeID         int64     `json:"change_id"`
	Outcome          string    `json:"outcome"` // successful, successful_with_issues, failed_rolled_back, cancelled
	ReviewNotes      string    `json:"review_notes"`
	LessonsLearned   string    `json:"lessons_learned,omitempty"`
	ReviewedByUserID string    `json:"reviewed_by_user_id"`
	ReviewedAt       time.Time `json:"reviewed_at"`
	CreatedAt        time.Time `json:"created_at"`

	// Enriched fields
	ReviewedByUsername string `json:"reviewed_by_username,omitempty"`
}

// CreateChangeRequest payload for raising a new change request
type CreateChangeRequest struct {
	WorkspaceID        int64   `json:"workspace_id"`
	ProjectID          *int64  `json:"project_id,omitempty"`
	Title              string  `json:"title"`
	Description        string  `json:"description,omitempty"`
	ChangeType         string  `json:"change_type"` // standard, normal, emergency
	RiskLevel          string  `json:"risk_level,omitempty"` // low, medium, high, critical
	ImpactLevel        string  `json:"impact_level,omitempty"` // low, medium, high
	ReasonForChange    string  `json:"reason_for_change,omitempty"`
	ImplementationPlan string  `json:"implementation_plan,omitempty"`
	RollbackPlan       string  `json:"rollback_plan,omitempty"`
	TestPlan           string  `json:"test_plan,omitempty"`
	AssignedToUserID   string  `json:"assigned_to_user_id,omitempty"`
	PlannedStartAt     *string `json:"planned_start_at,omitempty"` // RFC3339
	PlannedEndAt       *string `json:"planned_end_at,omitempty"`   // RFC3339
}

// UpdateChangeRequest payload for editing a draft/pending change request
type UpdateChangeRequest struct {
	ProjectID          *int64  `json:"project_id,omitempty"`
	Title              *string `json:"title,omitempty"`
	Description        *string `json:"description,omitempty"`
	ChangeType         *string `json:"change_type,omitempty"`
	RiskLevel          *string `json:"risk_level,omitempty"`
	ImpactLevel        *string `json:"impact_level,omitempty"`
	ReasonForChange    *string `json:"reason_for_change,omitempty"`
	ImplementationPlan *string `json:"implementation_plan,omitempty"`
	RollbackPlan       *string `json:"rollback_plan,omitempty"`
	TestPlan           *string `json:"test_plan,omitempty"`
	AssignedToUserID   *string `json:"assigned_to_user_id,omitempty"`
	PlannedStartAt     *string `json:"planned_start_at,omitempty"`
	PlannedEndAt       *string `json:"planned_end_at,omitempty"`
}

// ChangeStatusRequest payload for transitioning change request lifecycle state
type ChangeStatusRequest struct {
	Status        string  `json:"status"`
	Comments      string  `json:"comments,omitempty"`
	ActualStartAt *string `json:"actual_start_at,omitempty"`
	ActualEndAt   *string `json:"actual_end_at,omitempty"`
}

// AddReviewerRequest payload for designating a CAB reviewer for a change
type AddReviewerRequest struct {
	ReviewerID string `json:"reviewer_id"`
}

// ChangeApprovalRequest payload for casting a CAB review vote
type ChangeApprovalRequest struct {
	Decision string `json:"decision"` // approved, rejected, abstained
	Comments string `json:"comments,omitempty"`
}

// ChangePIRRequest payload for submitting a Post-Implementation Review
type ChangePIRRequest struct {
	Outcome        string `json:"outcome"` // successful, successful_with_issues, failed_rolled_back, cancelled
	ReviewNotes    string `json:"review_notes"`
	LessonsLearned string `json:"lessons_learned,omitempty"`
}

// LinkChangeTaskRequest payload for associating an execution task to a change request
type LinkChangeTaskRequest struct {
	TaskID int    `json:"task_id"`
	Phase  string `json:"phase,omitempty"` // pre_implementation, implementation, post_implementation
}

// ChangeFilter provides search, filtering, and pagination parameters across change requests
type ChangeFilter struct {
	WorkspaceID int64   `json:"workspace_id"`
	ProjectID   *int64  `json:"project_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	ChangeType  *string `json:"change_type,omitempty"`
	RiskLevel   *string `json:"risk_level,omitempty"`
	RequestedBy *string `json:"requested_by,omitempty"`
	AssignedTo  *string `json:"assigned_to,omitempty"`
	Search      *string `json:"search,omitempty"`
	StartDate   *string `json:"start_date,omitempty"` // YYYY-MM-DD
	EndDate     *string `json:"end_date,omitempty"`   // YYYY-MM-DD
	Limit       int     `json:"limit,omitempty"`
	Offset      int     `json:"offset,omitempty"`
}
