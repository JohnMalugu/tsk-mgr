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
	ErrRecoveryObjectiveNotFound = errors.New("recovery objective not found")
	ErrServiceNameRequired     = errors.New("service name is required")
	ErrInvalidRTOValue         = errors.New("rto_value_minutes must be non-negative")
	ErrInvalidRPOValue         = errors.New("rpo_value_minutes must be non-negative")
	ErrInvalidAvailability     = errors.New("target availability must be between 0.0 and 100.0 percent")
	ErrInvalidDataLossTolerance = errors.New("invalid data loss tolerance")
	ErrBIANotFound             = errors.New("business impact analysis not found")
	ErrInvalidMTDValue         = errors.New("mtd_minutes must be greater than zero")
	ErrInvalidOperationalImpact = errors.New("invalid operational impact")
	ErrInvalidRegulatoryImpact = errors.New("invalid regulatory impact")
	ErrInvalidCriticalWindow   = errors.New("invalid critical recovery window")
	ErrRunbookNotFound         = errors.New("recovery runbook not found")
	ErrRunbookStepNotFound     = errors.New("runbook step not found")
	ErrRunbookTitleRequired    = errors.New("runbook title is required")
	ErrStepNameRequired        = errors.New("step name is required")
	ErrInvalidExecutionMode    = errors.New("invalid runbook execution mode")
	ErrInvalidActionType       = errors.New("invalid runbook step action type")
	ErrInvalidExecutionRole    = errors.New("invalid runbook step execution role")
)

// Runbook Execution Mode constants
const (
	RunbookModeAutomated = "automated"
	RunbookModeManual    = "manual"
	RunbookModeHybrid    = "hybrid"
)

// Runbook Step Action Type constants
const (
	ActionInfraFailover       = "infrastructure_failover"
	ActionDNSCutover          = "dns_cutover"
	ActionDataRestore         = "data_restore"
	ActionHealthCheck         = "health_check"
	ActionNotification        = "notification"
	ActionManualIntervention  = "manual_intervention"
)

// Runbook Step Execution Role constants
const (
	RoleIncidentCommander      = "incident_commander"
	RoleDatabaseAdmin          = "database_admin"
	RoleNetworkEngineer        = "network_engineer"
	RoleSiteReliabilityEngineer = "site_reliability_engineer"
	RoleCloudOperator          = "cloud_operator"
)

// ValidateExecutionMode checks whether execution mode is valid
func ValidateExecutionMode(m string) error {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case RunbookModeAutomated, RunbookModeManual, RunbookModeHybrid:
		return nil
	default:
		return ErrInvalidExecutionMode
	}
}

// ValidateActionType checks whether step action type is valid
func ValidateActionType(a string) error {
	switch strings.ToLower(strings.TrimSpace(a)) {
	case ActionInfraFailover, ActionDNSCutover, ActionDataRestore, ActionHealthCheck, ActionNotification, ActionManualIntervention:
		return nil
	default:
		return ErrInvalidActionType
	}
}

// ValidateExecutionRole checks whether step execution role is valid
func ValidateExecutionRole(r string) error {
	switch strings.ToLower(strings.TrimSpace(r)) {
	case RoleIncidentCommander, RoleDatabaseAdmin, RoleNetworkEngineer, RoleSiteReliabilityEngineer, RoleCloudOperator:
		return nil
	default:
		return ErrInvalidExecutionRole
	}
}

// Operational Impact constants
const (
	OpImpactNegligible   = "negligible"
	OpImpactMinor        = "minor"
	OpImpactModerate     = "moderate"
	OpImpactSevere       = "severe"
	OpImpactCatastrophic = "catastrophic"
)

// Regulatory Impact constants
const (
	RegImpactNone              = "none"
	RegImpactReportingRequired = "reporting_required"
	RegImpactSanctionsPossible = "sanctions_possible"
	RegImpactCriticalBreach    = "critical_breach"
)

// Critical Recovery Window constants
const (
	RecoveryWindowImmediate = "immediate_under_1h"
	RecoveryWindowSameDay   = "same_day_under_8h"
	RecoveryWindowNextDay   = "next_day_under_24h"
	RecoveryWindowExtended  = "extended_over_24h"
)

// ValidateOperationalImpact checks whether operational impact is recognized
func ValidateOperationalImpact(imp string) error {
	switch strings.ToLower(strings.TrimSpace(imp)) {
	case OpImpactNegligible, OpImpactMinor, OpImpactModerate, OpImpactSevere, OpImpactCatastrophic:
		return nil
	default:
		return ErrInvalidOperationalImpact
	}
}

// ValidateRegulatoryImpact checks whether regulatory impact is recognized
func ValidateRegulatoryImpact(imp string) error {
	switch strings.ToLower(strings.TrimSpace(imp)) {
	case RegImpactNone, RegImpactReportingRequired, RegImpactSanctionsPossible, RegImpactCriticalBreach:
		return nil
	default:
		return ErrInvalidRegulatoryImpact
	}
}

// ValidateCriticalRecoveryWindow checks whether recovery window is recognized
func ValidateCriticalRecoveryWindow(w string) error {
	switch strings.ToLower(strings.TrimSpace(w)) {
	case RecoveryWindowImmediate, RecoveryWindowSameDay, RecoveryWindowNextDay, RecoveryWindowExtended:
		return nil
	default:
		return ErrInvalidCriticalWindow
	}
}

// Data Loss Tolerance constants
const (
	DataLossToleranceZeroLoss    = "zero_loss"
	DataLossToleranceMinimalLoss = "minimal_loss"
	DataLossToleranceRelaxed     = "relaxed"
)

// ValidateDataLossTolerance checks whether data loss tolerance is recognized
func ValidateDataLossTolerance(tol string) error {
	switch strings.ToLower(strings.TrimSpace(tol)) {
	case DataLossToleranceZeroLoss, DataLossToleranceMinimalLoss, DataLossToleranceRelaxed:
		return nil
	default:
		return ErrInvalidDataLossTolerance
	}
}

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

// RecoveryObjective represents specific RTO/RPO targets mapped to a business service or CMDB CI
type RecoveryObjective struct {
	ID                 string    `json:"id"`
	WorkspaceID        string    `json:"workspace_id"`
	PlanID             string    `json:"plan_id"`
	CIID               string    `json:"ci_id,omitempty"`
	ServiceName        string    `json:"service_name"`
	RTOValueMinutes    int       `json:"rto_value_minutes"`
	RPOValueMinutes    int       `json:"rpo_value_minutes"`
	TargetAvailability float64   `json:"target_availability"`
	DataLossTolerance  string    `json:"data_loss_tolerance"`
	Notes              string    `json:"notes"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// Validate validates the recovery objective fields
func (ro *RecoveryObjective) Validate() error {
	if strings.TrimSpace(ro.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if ro.RTOValueMinutes < 0 {
		return ErrInvalidRTOValue
	}
	if ro.RPOValueMinutes < 0 {
		return ErrInvalidRPOValue
	}
	if ro.TargetAvailability < 0.0 || ro.TargetAvailability > 100.0 {
		return ErrInvalidAvailability
	}
	if ro.DataLossTolerance != "" {
		if err := ValidateDataLossTolerance(ro.DataLossTolerance); err != nil {
			return err
		}
	}
	return nil
}

// BusinessImpactAnalysis represents a comprehensive BIA record evaluating outage losses and tolerances
type BusinessImpactAnalysis struct {
	ID                       string    `json:"id"`
	WorkspaceID              string    `json:"workspace_id"`
	PlanID                   string    `json:"plan_id"`
	CIID                     string    `json:"ci_id,omitempty"`
	ServiceName              string    `json:"service_name"`
	MTDMinutes               int       `json:"mtd_minutes"`
	FinancialLossRatePerHour float64   `json:"financial_loss_rate_per_hour"`
	OperationalImpact        string    `json:"operational_impact"`
	RegulatoryImpact         string    `json:"regulatory_impact"`
	CriticalRecoveryWindow   string    `json:"critical_recovery_window"`
	AssessmentDate           time.Time `json:"assessment_date"`
	AssessedBy               string    `json:"assessed_by"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// Validate validates the business impact analysis fields
func (b *BusinessImpactAnalysis) Validate() error {
	if strings.TrimSpace(b.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if b.MTDMinutes <= 0 {
		return ErrInvalidMTDValue
	}
	if err := ValidateOperationalImpact(b.OperationalImpact); err != nil {
		return err
	}
	if err := ValidateRegulatoryImpact(b.RegulatoryImpact); err != nil {
		return err
	}
	if err := ValidateCriticalRecoveryWindow(b.CriticalRecoveryWindow); err != nil {
		return err
	}
	return nil
}

// RecoveryRunbook defines an actionable execution procedure for disaster recovery
type RecoveryRunbook struct {
	ID                       string        `json:"id"`
	WorkspaceID              string        `json:"workspace_id"`
	PlanID                   string        `json:"plan_id"`
	RunbookCode              string        `json:"runbook_code"`
	Title                    string        `json:"title"`
	Description              string        `json:"description"`
	ExecutionMode            string        `json:"execution_mode"`
	EstimatedDurationMinutes int           `json:"estimated_duration_minutes"`
	Steps                    []RunbookStep `json:"steps,omitempty"`
	CreatedAt                time.Time     `json:"created_at"`
	UpdatedAt                time.Time     `json:"updated_at"`
}

// Validate validates the recovery runbook
func (r *RecoveryRunbook) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return ErrRunbookTitleRequired
	}
	if r.ExecutionMode != "" {
		if err := ValidateExecutionMode(r.ExecutionMode); err != nil {
			return err
		}
	}
	return nil
}

// RunbookStep represents an individual sequential procedure within a disaster recovery runbook
type RunbookStep struct {
	ID                       string    `json:"id"`
	RunbookID                string    `json:"runbook_id"`
	StepOrder                int       `json:"step_order"`
	Name                     string    `json:"name"`
	ActionType               string    `json:"action_type"`
	ExecutionRole            string    `json:"execution_role"`
	CommandOrScript          string    `json:"command_or_script"`
	EstimatedDurationMinutes int       `json:"estimated_duration_minutes"`
	IsCritical               bool      `json:"is_critical"`
	FallbackProcedure        string    `json:"fallback_procedure"`
	CreatedAt                time.Time `json:"created_at"`
	UpdatedAt                time.Time `json:"updated_at"`
}

// Validate validates the runbook step
func (s *RunbookStep) Validate() error {
	if strings.TrimSpace(s.Name) == "" {
		return ErrStepNameRequired
	}
	if err := ValidateActionType(s.ActionType); err != nil {
		return err
	}
	if err := ValidateExecutionRole(s.ExecutionRole); err != nil {
		return err
	}
	return nil
}
