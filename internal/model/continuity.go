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
	ErrDrillNotFound           = errors.New("continuity drill not found")
	ErrDrillFindingNotFound    = errors.New("continuity drill finding not found")
	ErrInvalidDrillType        = errors.New("invalid continuity drill type")
	ErrInvalidDrillStatus      = errors.New("invalid continuity drill status")
	ErrDrillScenarioRequired   = errors.New("drill scenario is required")
	ErrFindingDescRequired     = errors.New("drill finding description is required")
	ErrInvalidFindingSeverity  = errors.New("invalid drill finding severity")
	ErrInvalidFindingCategory  = errors.New("invalid drill finding category")
	ErrInvalidDrillFindingStatus = errors.New("invalid drill finding status")
)

// Continuity Drill Type constants
const (
	DrillTypeTabletop     = "tabletop"
	DrillTypeWalkthrough  = "walkthrough"
	DrillTypeSimulation   = "simulation"
	DrillTypeFullFailover = "full_failover"
)

// Continuity Drill Status constants
const (
	DrillStatusScheduled  = "scheduled"
	DrillStatusInProgress = "in_progress"
	DrillStatusCompleted  = "completed"
	DrillStatusFailed     = "failed"
	DrillStatusCancelled  = "cancelled"
)

// Drill Finding Severity constants
const (
	DrillFindingSeverityCritical = "critical"
	DrillFindingSeverityHigh     = "high"
	DrillFindingSeverityMedium   = "medium"
	DrillFindingSeverityLow      = "low"
)

// Drill Finding Category constants
const (
	DrillFindingCatProcessGap      = "process_gap"
	DrillFindingCatToolingFailure  = "tooling_failure"
	DrillFindingCatCommDelay       = "communication_delay"
	DrillFindingCatDocOutdated     = "documentation_outdated"
	DrillFindingCatPermissionIssue = "permission_issue"
)

// Drill Finding Status constants
const (
	DrillFindingStatusOpen       = "open"
	DrillFindingStatusInProgress = "in_progress"
	DrillFindingStatusResolved   = "resolved"
	DrillFindingStatusWaived     = "waived"
)

// ValidateDrillType checks whether drill type is valid
func ValidateDrillType(t string) error {
	switch strings.ToLower(strings.TrimSpace(t)) {
	case DrillTypeTabletop, DrillTypeWalkthrough, DrillTypeSimulation, DrillTypeFullFailover:
		return nil
	default:
		return ErrInvalidDrillType
	}
}

// ValidateDrillStatus checks whether drill status is valid
func ValidateDrillStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case DrillStatusScheduled, DrillStatusInProgress, DrillStatusCompleted, DrillStatusFailed, DrillStatusCancelled:
		return nil
	default:
		return ErrInvalidDrillStatus
	}
}

// ValidateFindingSeverity checks whether finding severity is valid
func ValidateFindingSeverity(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case DrillFindingSeverityCritical, DrillFindingSeverityHigh, DrillFindingSeverityMedium, DrillFindingSeverityLow:
		return nil
	default:
		return ErrInvalidFindingSeverity
	}
}

// ValidateFindingCategory checks whether finding category is valid
func ValidateFindingCategory(c string) error {
	switch strings.ToLower(strings.TrimSpace(c)) {
	case DrillFindingCatProcessGap, DrillFindingCatToolingFailure, DrillFindingCatCommDelay, DrillFindingCatDocOutdated, DrillFindingCatPermissionIssue:
		return nil
	default:
		return ErrInvalidFindingCategory
	}
}

// ValidateDrillFindingStatus checks whether finding status is valid
func ValidateDrillFindingStatus(s string) error {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case DrillFindingStatusOpen, DrillFindingStatusInProgress, DrillFindingStatusResolved, DrillFindingStatusWaived:
		return nil
	default:
		return ErrInvalidDrillFindingStatus
	}
}

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

// ContinuityDrill represents a planned or executed disaster recovery exercise
type ContinuityDrill struct {
	ID                            string         `json:"id"`
	WorkspaceID                   string         `json:"workspace_id"`
	PlanID                        string         `json:"plan_id"`
	DrillCode                     string         `json:"drill_code"`
	DrillType                     string         `json:"drill_type"`
	Scenario                      string         `json:"scenario"`
	ScheduledDate                 time.Time      `json:"scheduled_date"`
	Status                        string         `json:"status"`
	ActualRecoveryDurationMinutes int            `json:"actual_recovery_duration_minutes"`
	Success                       bool           `json:"success"`
	ConductedBy                   string         `json:"conducted_by"`
	PostMortemSummary             string         `json:"post_mortem_summary"`
	Findings                      []DrillFinding `json:"findings,omitempty"`
	CreatedAt                     time.Time      `json:"created_at"`
	UpdatedAt                     time.Time      `json:"updated_at"`
}

// Validate validates the continuity drill
func (d *ContinuityDrill) Validate() error {
	if strings.TrimSpace(d.Scenario) == "" {
		return ErrDrillScenarioRequired
	}
	if err := ValidateDrillType(d.DrillType); err != nil {
		return err
	}
	if d.Status != "" {
		if err := ValidateDrillStatus(d.Status); err != nil {
			return err
		}
	}
	return nil
}

// DrillFinding represents an actionable gap, observation, or remediation item uncovered during a DR drill
type DrillFinding struct {
	ID                string     `json:"id"`
	WorkspaceID       string     `json:"workspace_id"`
	DrillID           string     `json:"drill_id"`
	Severity          string     `json:"severity"`
	Category          string     `json:"category"`
	Description       string     `json:"description"`
	RemediationAction string     `json:"remediation_action"`
	AssignedTo        string     `json:"assigned_to"`
	Status            string     `json:"status"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// Validate validates the drill finding
func (f *DrillFinding) Validate() error {
	if strings.TrimSpace(f.Description) == "" {
		return ErrFindingDescRequired
	}
	if err := ValidateFindingSeverity(f.Severity); err != nil {
		return err
	}
	if err := ValidateFindingCategory(f.Category); err != nil {
		return err
	}
	if f.Status != "" {
		if err := ValidateDrillFindingStatus(f.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateDRPlanRequest defines payload to create a new disaster recovery plan
type CreateDRPlanRequest struct {
	PlanCode             string `json:"plan_code"`
	Name                 string `json:"name"`
	Scope                string `json:"scope"`
	RecoveryTier         string `json:"recovery_tier"`
	Status               string `json:"status,omitempty"`
	PrimaryCoordinatorID string `json:"primary_coordinator_id,omitempty"`
	BackupCoordinatorID  string `json:"backup_coordinator_id,omitempty"`
	InvocationCriteria   string `json:"invocation_criteria,omitempty"`
	DocumentURL          string `json:"document_url,omitempty"`
}

// Validate validates CreateDRPlanRequest
func (r *CreateDRPlanRequest) Validate() error {
	if strings.TrimSpace(r.PlanCode) == "" {
		return ErrDRPlanCodeRequired
	}
	if strings.TrimSpace(r.Name) == "" {
		return ErrDRPlanNameRequired
	}
	if err := ValidateRecoveryTier(r.RecoveryTier); err != nil {
		return err
	}
	if r.Status != "" {
		if err := ValidateDRPlanStatus(r.Status); err != nil {
			return err
		}
	}
	return nil
}

// UpdateDRPlanRequest defines payload to update an existing disaster recovery plan
type UpdateDRPlanRequest struct {
	Name                 *string `json:"name,omitempty"`
	Scope                *string `json:"scope,omitempty"`
	RecoveryTier         *string `json:"recovery_tier,omitempty"`
	Status               *string `json:"status,omitempty"`
	PrimaryCoordinatorID *string `json:"primary_coordinator_id,omitempty"`
	BackupCoordinatorID  *string `json:"backup_coordinator_id,omitempty"`
	InvocationCriteria   *string `json:"invocation_criteria,omitempty"`
	DocumentURL          *string `json:"document_url,omitempty"`
}

// Validate validates UpdateDRPlanRequest
func (r *UpdateDRPlanRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrDRPlanNameRequired
	}
	if r.RecoveryTier != nil {
		if err := ValidateRecoveryTier(*r.RecoveryTier); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateDRPlanStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateRecoveryObjectiveRequest defines payload to configure RTO/RPO for a service or CI
type CreateRecoveryObjectiveRequest struct {
	PlanID             string  `json:"plan_id"`
	CIID               string  `json:"ci_id,omitempty"`
	ServiceName        string  `json:"service_name"`
	RTOValueMinutes    int     `json:"rto_value_minutes"`
	RPOValueMinutes    int     `json:"rpo_value_minutes"`
	TargetAvailability float64 `json:"target_availability,omitempty"`
	DataLossTolerance  string  `json:"data_loss_tolerance,omitempty"`
	Notes              string  `json:"notes,omitempty"`
}

// Validate validates CreateRecoveryObjectiveRequest
func (r *CreateRecoveryObjectiveRequest) Validate() error {
	if strings.TrimSpace(r.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if r.RTOValueMinutes < 0 {
		return ErrInvalidRTOValue
	}
	if r.RPOValueMinutes < 0 {
		return ErrInvalidRPOValue
	}
	if r.TargetAvailability < 0.0 || r.TargetAvailability > 100.0 {
		return ErrInvalidAvailability
	}
	if r.DataLossTolerance != "" {
		if err := ValidateDataLossTolerance(r.DataLossTolerance); err != nil {
			return err
		}
	}
	return nil
}

// UpdateRecoveryObjectiveRequest defines payload to modify recovery objectives
type UpdateRecoveryObjectiveRequest struct {
	CIID               *string  `json:"ci_id,omitempty"`
	ServiceName        *string  `json:"service_name,omitempty"`
	RTOValueMinutes    *int     `json:"rto_value_minutes,omitempty"`
	RPOValueMinutes    *int     `json:"rpo_value_minutes,omitempty"`
	TargetAvailability *float64 `json:"target_availability,omitempty"`
	DataLossTolerance  *string  `json:"data_loss_tolerance,omitempty"`
	Notes              *string  `json:"notes,omitempty"`
}

// Validate validates UpdateRecoveryObjectiveRequest
func (r *UpdateRecoveryObjectiveRequest) Validate() error {
	if r.ServiceName != nil && strings.TrimSpace(*r.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if r.RTOValueMinutes != nil && *r.RTOValueMinutes < 0 {
		return ErrInvalidRTOValue
	}
	if r.RPOValueMinutes != nil && *r.RPOValueMinutes < 0 {
		return ErrInvalidRPOValue
	}
	if r.TargetAvailability != nil && (*r.TargetAvailability < 0.0 || *r.TargetAvailability > 100.0) {
		return ErrInvalidAvailability
	}
	if r.DataLossTolerance != nil && *r.DataLossTolerance != "" {
		if err := ValidateDataLossTolerance(*r.DataLossTolerance); err != nil {
			return err
		}
	}
	return nil
}

// CreateBIARequest defines payload to record a Business Impact Analysis
type CreateBIARequest struct {
	PlanID                   string  `json:"plan_id"`
	CIID                     string  `json:"ci_id,omitempty"`
	ServiceName              string  `json:"service_name"`
	MTDMinutes               int     `json:"mtd_minutes"`
	FinancialLossRatePerHour float64 `json:"financial_loss_rate_per_hour,omitempty"`
	OperationalImpact        string  `json:"operational_impact"`
	RegulatoryImpact         string  `json:"regulatory_impact"`
	CriticalRecoveryWindow   string  `json:"critical_recovery_window"`
	AssessedBy               string  `json:"assessed_by,omitempty"`
}

// Validate validates CreateBIARequest
func (r *CreateBIARequest) Validate() error {
	if strings.TrimSpace(r.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if r.MTDMinutes <= 0 {
		return ErrInvalidMTDValue
	}
	if err := ValidateOperationalImpact(r.OperationalImpact); err != nil {
		return err
	}
	if err := ValidateRegulatoryImpact(r.RegulatoryImpact); err != nil {
		return err
	}
	if err := ValidateCriticalRecoveryWindow(r.CriticalRecoveryWindow); err != nil {
		return err
	}
	return nil
}

// UpdateBIARequest defines payload to modify a Business Impact Analysis
type UpdateBIARequest struct {
	CIID                     *string  `json:"ci_id,omitempty"`
	ServiceName              *string  `json:"service_name,omitempty"`
	MTDMinutes               *int     `json:"mtd_minutes,omitempty"`
	FinancialLossRatePerHour *float64 `json:"financial_loss_rate_per_hour,omitempty"`
	OperationalImpact        *string  `json:"operational_impact,omitempty"`
	RegulatoryImpact         *string  `json:"regulatory_impact,omitempty"`
	CriticalRecoveryWindow   *string  `json:"critical_recovery_window,omitempty"`
	AssessedBy               *string  `json:"assessed_by,omitempty"`
}

// Validate validates UpdateBIARequest
func (r *UpdateBIARequest) Validate() error {
	if r.ServiceName != nil && strings.TrimSpace(*r.ServiceName) == "" {
		return ErrServiceNameRequired
	}
	if r.MTDMinutes != nil && *r.MTDMinutes <= 0 {
		return ErrInvalidMTDValue
	}
	if r.OperationalImpact != nil {
		if err := ValidateOperationalImpact(*r.OperationalImpact); err != nil {
			return err
		}
	}
	if r.RegulatoryImpact != nil {
		if err := ValidateRegulatoryImpact(*r.RegulatoryImpact); err != nil {
			return err
		}
	}
	if r.CriticalRecoveryWindow != nil {
		if err := ValidateCriticalRecoveryWindow(*r.CriticalRecoveryWindow); err != nil {
			return err
		}
	}
	return nil
}

// CreateRunbookRequest defines payload to create a recovery runbook
type CreateRunbookRequest struct {
	PlanID                   string `json:"plan_id"`
	RunbookCode              string `json:"runbook_code"`
	Title                    string `json:"title"`
	Description              string `json:"description,omitempty"`
	ExecutionMode            string `json:"execution_mode,omitempty"`
	EstimatedDurationMinutes int    `json:"estimated_duration_minutes,omitempty"`
}

// Validate validates CreateRunbookRequest
func (r *CreateRunbookRequest) Validate() error {
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

// UpdateRunbookRequest defines payload to modify a recovery runbook
type UpdateRunbookRequest struct {
	Title                    *string `json:"title,omitempty"`
	Description              *string `json:"description,omitempty"`
	ExecutionMode            *string `json:"execution_mode,omitempty"`
	EstimatedDurationMinutes *int    `json:"estimated_duration_minutes,omitempty"`
}

// Validate validates UpdateRunbookRequest
func (r *UpdateRunbookRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return ErrRunbookTitleRequired
	}
	if r.ExecutionMode != nil && *r.ExecutionMode != "" {
		if err := ValidateExecutionMode(*r.ExecutionMode); err != nil {
			return err
		}
	}
	return nil
}

// CreateRunbookStepRequest defines payload to append a step to a recovery runbook
type CreateRunbookStepRequest struct {
	StepOrder                int    `json:"step_order"`
	Name                     string `json:"name"`
	ActionType               string `json:"action_type"`
	ExecutionRole            string `json:"execution_role"`
	CommandOrScript          string `json:"command_or_script,omitempty"`
	EstimatedDurationMinutes int    `json:"estimated_duration_minutes,omitempty"`
	IsCritical               bool   `json:"is_critical,omitempty"`
	FallbackProcedure        string `json:"fallback_procedure,omitempty"`
}

// Validate validates CreateRunbookStepRequest
func (r *CreateRunbookStepRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return ErrStepNameRequired
	}
	if err := ValidateActionType(r.ActionType); err != nil {
		return err
	}
	if err := ValidateExecutionRole(r.ExecutionRole); err != nil {
		return err
	}
	return nil
}

// UpdateRunbookStepRequest defines payload to update a step in a runbook
type UpdateRunbookStepRequest struct {
	StepOrder                *int    `json:"step_order,omitempty"`
	Name                     *string `json:"name,omitempty"`
	ActionType               *string `json:"action_type,omitempty"`
	ExecutionRole            *string `json:"execution_role,omitempty"`
	CommandOrScript          *string `json:"command_or_script,omitempty"`
	EstimatedDurationMinutes *int    `json:"estimated_duration_minutes,omitempty"`
	IsCritical               *bool   `json:"is_critical,omitempty"`
	FallbackProcedure        *string `json:"fallback_procedure,omitempty"`
}

// Validate validates UpdateRunbookStepRequest
func (r *UpdateRunbookStepRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrStepNameRequired
	}
	if r.ActionType != nil {
		if err := ValidateActionType(*r.ActionType); err != nil {
			return err
		}
	}
	if r.ExecutionRole != nil {
		if err := ValidateExecutionRole(*r.ExecutionRole); err != nil {
			return err
		}
	}
	return nil
}

// CreateDrillRequest defines payload to schedule a continuity drill
type CreateDrillRequest struct {
	PlanID            string    `json:"plan_id"`
	DrillCode         string    `json:"drill_code"`
	DrillType         string    `json:"drill_type"`
	Scenario          string    `json:"scenario"`
	ScheduledDate     time.Time `json:"scheduled_date"`
	ConductedBy       string    `json:"conducted_by,omitempty"`
}

// Validate validates CreateDrillRequest
func (r *CreateDrillRequest) Validate() error {
	if strings.TrimSpace(r.Scenario) == "" {
		return ErrDrillScenarioRequired
	}
	if err := ValidateDrillType(r.DrillType); err != nil {
		return err
	}
	return nil
}

// UpdateDrillRequest defines payload to update a continuity drill status, outcome, and post-mortem
type UpdateDrillRequest struct {
	Scenario                      *string    `json:"scenario,omitempty"`
	ScheduledDate                 *time.Time `json:"scheduled_date,omitempty"`
	Status                        *string    `json:"status,omitempty"`
	ActualRecoveryDurationMinutes *int       `json:"actual_recovery_duration_minutes,omitempty"`
	Success                       *bool      `json:"success,omitempty"`
	ConductedBy                   *string    `json:"conducted_by,omitempty"`
	PostMortemSummary             *string    `json:"post_mortem_summary,omitempty"`
}

// Validate validates UpdateDrillRequest
func (r *UpdateDrillRequest) Validate() error {
	if r.Scenario != nil && strings.TrimSpace(*r.Scenario) == "" {
		return ErrDrillScenarioRequired
	}
	if r.Status != nil {
		if err := ValidateDrillStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateDrillFindingRequest defines payload to log a finding or gap during a drill
type CreateDrillFindingRequest struct {
	Severity          string `json:"severity"`
	Category          string `json:"category"`
	Description       string `json:"description"`
	RemediationAction string `json:"remediation_action,omitempty"`
	AssignedTo        string `json:"assigned_to,omitempty"`
}

// Validate validates CreateDrillFindingRequest
func (r *CreateDrillFindingRequest) Validate() error {
	if strings.TrimSpace(r.Description) == "" {
		return ErrFindingDescRequired
	}
	if err := ValidateFindingSeverity(r.Severity); err != nil {
		return err
	}
	if err := ValidateFindingCategory(r.Category); err != nil {
		return err
	}
	return nil
}

// UpdateDrillFindingRequest defines payload to update finding status or assignment
type UpdateDrillFindingRequest struct {
	Severity          *string `json:"severity,omitempty"`
	Category          *string `json:"category,omitempty"`
	Description       *string `json:"description,omitempty"`
	RemediationAction *string `json:"remediation_action,omitempty"`
	AssignedTo        *string `json:"assigned_to,omitempty"`
	Status            *string `json:"status,omitempty"`
}

// Validate validates UpdateDrillFindingRequest
func (r *UpdateDrillFindingRequest) Validate() error {
	if r.Description != nil && strings.TrimSpace(*r.Description) == "" {
		return ErrFindingDescRequired
	}
	if r.Severity != nil {
		if err := ValidateFindingSeverity(*r.Severity); err != nil {
			return err
		}
	}
	if r.Category != nil {
		if err := ValidateFindingCategory(*r.Category); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateDrillFindingStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// DRPlanFilterCriteria provides query parameters for listing disaster recovery plans
type DRPlanFilterCriteria struct {
	WorkspaceID  string
	RecoveryTier string
	Status       string
	Search       string
	Limit        int
	Offset       int
}

// DrillFilterCriteria provides query parameters for listing continuity drills
type DrillFilterCriteria struct {
	WorkspaceID string
	PlanID      string
	DrillType   string
	Status      string
	Limit       int
	Offset      int
}

// DrillFindingFilterCriteria provides query parameters for listing drill findings
type DrillFindingFilterCriteria struct {
	WorkspaceID string
	DrillID     string
	Severity    string
	Category    string
	Status      string
	Limit       int
	Offset      int
}

// ContinuityReadinessReport provides an aggregated executive posture of disaster recovery readiness and resilience
type ContinuityReadinessReport struct {
	WorkspaceID             string         `json:"workspace_id"`
	TotalPlans              int            `json:"total_plans"`
	ActivePlans             int            `json:"active_plans"`
	ApprovedPlans           int            `json:"approved_plans"`
	PlansByTier             map[string]int `json:"plans_by_tier"`
	TotalObjectives         int            `json:"total_objectives"`
	AverageRTOMinutes       float64        `json:"average_rto_minutes"`
	AverageRPOMinutes       float64        `json:"average_rpo_minutes"`
	TotalDrills             int            `json:"total_drills"`
	CompletedDrills         int            `json:"completed_drills"`
	DrillPassRate           float64        `json:"drill_pass_rate"`
	RTOComplianceRate       float64        `json:"rto_compliance_rate"`
	AverageRecoveryDuration float64        `json:"average_recovery_duration_minutes"`
	TotalFindings           int            `json:"total_findings"`
	OpenFindingsCount       int            `json:"open_findings_count"`
	FindingsBySeverity      map[string]int `json:"findings_by_severity"`
	ReadinessScore          float64        `json:"readiness_score"`
}

