package model

import (
	"errors"
	"strings"
	"time"
)

// Disaster Recovery Tier constants
const (
	RecoveryTier1MissionCritical = "tier_1_mission_critical"
	RecoveryTier2BusinessCritical = "tier_2_business_critical"
	RecoveryTier3Operational     = "tier_3_operational"
)

// Disaster Recovery Plan Status constants
const (
	DRPlanStatusDraft       = "draft"
	DRPlanStatusUnderReview = "under_review"
	DRPlanStatusApproved    = "approved"
	DRPlanStatusActive      = "active"
	DRPlanStatusArchived    = "archived"
)

// Continuity domain errors
var (
	ErrDRPlanNotFound          = errors.New("disaster recovery plan not found")
	ErrDRPlanCodeRequired      = errors.New("plan_code is required")
	ErrDRPlanCodeConflict      = errors.New("disaster recovery plan with this code already exists in workspace")
	ErrDRPlanNameRequired      = errors.New("plan name is required")
	ErrInvalidRecoveryTier     = errors.New("invalid recovery tier")
	ErrInvalidDRPlanStatus     = errors.New("invalid disaster recovery plan status")
)

// ValidateRecoveryTier validates whether the given tier string is recognized
func ValidateRecoveryTier(tier string) error {
	switch strings.ToLower(strings.TrimSpace(tier)) {
	case RecoveryTier1MissionCritical, RecoveryTier2BusinessCritical, RecoveryTier3Operational:
		return nil
	default:
		return ErrInvalidRecoveryTier
	}
}

// ValidateDRPlanStatus validates whether the given plan status is recognized
func ValidateDRPlanStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case DRPlanStatusDraft, DRPlanStatusUnderReview, DRPlanStatusApproved, DRPlanStatusActive, DRPlanStatusArchived:
		return nil
	default:
		return ErrInvalidDRPlanStatus
	}
}

// DisasterRecoveryPlan represents a formally defined IT service continuity and disaster recovery plan
type DisasterRecoveryPlan struct {
	ID                   string    `json:"id"`
	WorkspaceID          string    `json:"workspace_id"`
	PlanCode             string    `json:"plan_code"`
	Name                 string    `json:"name"`
	Scope                string    `json:"scope"`
	RecoveryTier         string    `json:"recovery_tier"`
	Status               string    `json:"status"`
	PrimaryCoordinatorID string    `json:"primary_coordinator_id"`
	BackupCoordinatorID  string    `json:"backup_coordinator_id"`
	InvocationCriteria   string    `json:"invocation_criteria"`
	DocumentURL          string    `json:"document_url"`
	Version              int       `json:"version"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}
