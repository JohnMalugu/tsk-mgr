package model

import (
	"errors"
	"strings"
	"time"
)

// Target environment constants
const (
	DeploymentEnvDev     = "development"
	DeploymentEnvTesting = "testing"
	DeploymentEnvStaging = "staging"
	DeploymentEnvUAT     = "uat"
	DeploymentEnvProd    = "production"
)

// Deployment strategy constants
const (
	DeploymentStrategyRolling   = "rolling"
	DeploymentStrategyBlueGreen = "blue_green"
	DeploymentStrategyCanary    = "canary"
	DeploymentStrategyRecreate  = "recreate"
)

// Pipeline status constants
const (
	PipelineStatusDraft      = "draft"
	PipelineStatusActive     = "active"
	PipelineStatusDeprecated = "deprecated"
)

// IsValidEnvironment checks if the environment string is valid
func IsValidEnvironment(env string) bool {
	switch env {
	case DeploymentEnvDev, DeploymentEnvTesting, DeploymentEnvStaging, DeploymentEnvUAT, DeploymentEnvProd:
		return true
	default:
		return false
	}
}

// IsValidDeploymentStrategy checks if the strategy string is valid
func IsValidDeploymentStrategy(strategy string) bool {
	switch strategy {
	case DeploymentStrategyRolling, DeploymentStrategyBlueGreen, DeploymentStrategyCanary, DeploymentStrategyRecreate:
		return true
	default:
		return false
	}
}

// IsValidPipelineStatus checks if the pipeline status string is valid
func IsValidPipelineStatus(status string) bool {
	switch status {
	case PipelineStatusDraft, PipelineStatusActive, PipelineStatusDeprecated:
		return true
	default:
		return false
	}
}

// DeploymentPipeline represents a defined CI/CD delivery pipeline configuration
type DeploymentPipeline struct {
	ID                  string    `json:"id"`
	WorkspaceID         string    `json:"workspace_id"`
	PipelineCode        string    `json:"pipeline_code"`
	Name                string    `json:"name"`
	Description         string    `json:"description,omitempty"`
	TargetEnvironment   string    `json:"target_environment"`
	DeploymentStrategy  string    `json:"deployment_strategy"`
	AutoRollbackEnabled bool      `json:"auto_rollback_enabled"`
	LinkedCIID          string    `json:"linked_ci_id,omitempty"`
	Status              string    `json:"status"`
	CreatedByUserID     string    `json:"created_by_user_id,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// Trigger type constants
const (
	RunTriggerManual      = "manual"
	RunTriggerAutomatedCI = "automated_ci"
	RunTriggerScheduled   = "scheduled"
)

// Run status constants
const (
	RunStatusPending    = "pending"
	RunStatusInProgress = "in_progress"
	RunStatusSuccessful = "successful"
	RunStatusFailed     = "failed"
	RunStatusRolledBack = "rolled_back"
	RunStatusCancelled  = "cancelled"
)

// IsValidRunTrigger checks if the trigger type is valid
func IsValidRunTrigger(trigger string) bool {
	switch trigger {
	case RunTriggerManual, RunTriggerAutomatedCI, RunTriggerScheduled:
		return true
	default:
		return false
	}
}

// IsValidRunStatus checks if the run status is valid
func IsValidRunStatus(status string) bool {
	switch status {
	case RunStatusPending, RunStatusInProgress, RunStatusSuccessful, RunStatusFailed, RunStatusRolledBack, RunStatusCancelled:
		return true
	default:
		return false
	}
}

// DeploymentRun represents a release execution within a deployment pipeline
type DeploymentRun struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	PipelineID      string     `json:"pipeline_id"`
	RunNumber       int64      `json:"run_number"`
	ReleaseVersion  string     `json:"release_version"`
	Environment     string     `json:"environment"`
	TriggerType     string     `json:"trigger_type"`
	Status          string     `json:"status"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	DurationSeconds float64    `json:"duration_seconds"`
	DeployedByUserID string    `json:"deployed_by_user_id,omitempty"`
	RollbackReason  string     `json:"rollback_reason,omitempty"`
	CommitHash      string     `json:"commit_hash,omitempty"`
	Changelog       string     `json:"changelog,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Gate type constants
const (
	GateTypeAutomatedTest        = "automated_test"
	GateTypeSecurityScan         = "security_scan"
	GateTypePerformanceBenchmark = "performance_benchmark"
	GateTypeManualApproval       = "manual_approval"
	GateTypeChangeApproval       = "change_approval"
)

// Gate status constants
const (
	GateStatusPending = "pending"
	GateStatusPassed  = "passed"
	GateStatusFailed  = "failed"
	GateStatusWaived  = "waived"
)

// IsValidGateType checks if the gate type is valid
func IsValidGateType(gt string) bool {
	switch gt {
	case GateTypeAutomatedTest, GateTypeSecurityScan, GateTypePerformanceBenchmark, GateTypeManualApproval, GateTypeChangeApproval:
		return true
	default:
		return false
	}
}

// IsValidGateStatus checks if the gate status is valid
func IsValidGateStatus(status string) bool {
	switch status {
	case GateStatusPending, GateStatusPassed, GateStatusFailed, GateStatusWaived:
		return true
	default:
		return false
	}
}

// DeploymentGate represents a quality or approval check required for a deployment run
type DeploymentGate struct {
	ID              string     `json:"id"`
	WorkspaceID     string     `json:"workspace_id"`
	RunID           string     `json:"run_id"`
	GateType        string     `json:"gate_type"`
	GateName        string     `json:"gate_name"`
	Status          string     `json:"status"`
	EvaluatorID     string     `json:"evaluator_id,omitempty"`
	EvaluatedAt     *time.Time `json:"evaluated_at,omitempty"`
	WaiverReason    string     `json:"waiver_reason,omitempty"`
	EvidenceDetails string     `json:"evidence_details,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Deployment step status constants
const (
	StepStatusPending  = "pending"
	StepStatusSoaking  = "soaking"
	StepStatusPromoted = "promoted"
	StepStatusAborted  = "aborted"
)

// IsValidStepStatus checks if the step status is valid
func IsValidStepStatus(status string) bool {
	switch status {
	case StepStatusPending, StepStatusSoaking, StepStatusPromoted, StepStatusAborted:
		return true
	default:
		return false
	}
}

// DeploymentStep represents an incremental rollout phase (e.g. Canary traffic slice)
type DeploymentStep struct {
	ID                 string     `json:"id"`
	WorkspaceID        string     `json:"workspace_id"`
	RunID              string     `json:"run_id"`
	StepOrder          int        `json:"step_order"`
	StepName           string     `json:"step_name"`
	TrafficPercentage  float64    `json:"traffic_percentage"`
	SoakTimeMinutes    int        `json:"soak_time_minutes"`
	ErrorRateThreshold float64    `json:"error_rate_threshold"`
	ActualErrorRate    float64    `json:"actual_error_rate"`
	Status             string     `json:"status"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Rollback trigger condition constants
const (
	RollbackTriggerErrorRateSpike       = "error_rate_spike"
	RollbackTriggerLatencyP99Threshold  = "latency_p99_threshold"
	RollbackTriggerHealthCheckFailure   = "health_check_failure"
	RollbackTriggerManualAbort          = "manual_abort"
)

// Rollback execution status constants
const (
	RollbackStatusTriggered  = "triggered"
	RollbackStatusInProgress = "in_progress"
	RollbackStatusCompleted  = "completed"
	RollbackStatusFailed     = "failed"
)

// IsValidRollbackTrigger checks if the rollback trigger condition is valid
func IsValidRollbackTrigger(trigger string) bool {
	switch trigger {
	case RollbackTriggerErrorRateSpike, RollbackTriggerLatencyP99Threshold, RollbackTriggerHealthCheckFailure, RollbackTriggerManualAbort:
		return true
	default:
		return false
	}
}

// IsValidRollbackStatus checks if the rollback execution status is valid
func IsValidRollbackStatus(status string) bool {
	switch status {
	case RollbackStatusTriggered, RollbackStatusInProgress, RollbackStatusCompleted, RollbackStatusFailed:
		return true
	default:
		return false
	}
}

// RollbackPolicy represents an automated or manual safety policy to revert a failed release
type RollbackPolicy struct {
	ID                    string     `json:"id"`
	WorkspaceID           string     `json:"workspace_id"`
	PipelineID            string     `json:"pipeline_id"`
	RunID                 string     `json:"run_id,omitempty"`
	TriggerCondition      string     `json:"trigger_condition"`
	ThresholdValue        float64    `json:"threshold_value"`
	TargetRollbackVersion string     `json:"target_rollback_version"`
	ExecutionStatus       string     `json:"execution_status"`
	TriggeredAt           *time.Time `json:"triggered_at,omitempty"`
	CompletedAt           *time.Time `json:"completed_at,omitempty"`
	RollbackLogs          string     `json:"rollback_logs,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// DTOs and validation rules for Deployment API

// CreateDeploymentPipelineRequest payload for creating a pipeline
type CreateDeploymentPipelineRequest struct {
	WorkspaceID         string `json:"workspace_id"`
	PipelineCode        string `json:"pipeline_code"`
	Name                string `json:"name"`
	Description         string `json:"description,omitempty"`
	TargetEnvironment   string `json:"target_environment"`
	DeploymentStrategy  string `json:"deployment_strategy"`
	AutoRollbackEnabled bool   `json:"auto_rollback_enabled"`
	LinkedCIID          string `json:"linked_ci_id,omitempty"`
	CreatedByUserID     string `json:"created_by_user_id,omitempty"`
}

func (r *CreateDeploymentPipelineRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.PipelineCode) == "" {
		return errors.New("pipeline_code is required")
	}
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if !IsValidEnvironment(r.TargetEnvironment) {
		return errors.New("invalid target_environment")
	}
	if !IsValidDeploymentStrategy(r.DeploymentStrategy) {
		return errors.New("invalid deployment_strategy")
	}
	return nil
}

// UpdateDeploymentPipelineRequest payload for updating a pipeline
type UpdateDeploymentPipelineRequest struct {
	Name                *string `json:"name,omitempty"`
	Description         *string `json:"description,omitempty"`
	TargetEnvironment   *string `json:"target_environment,omitempty"`
	DeploymentStrategy  *string `json:"deployment_strategy,omitempty"`
	AutoRollbackEnabled *bool   `json:"auto_rollback_enabled,omitempty"`
	LinkedCIID          *string `json:"linked_ci_id,omitempty"`
	Status              *string `json:"status,omitempty"`
}

func (r *UpdateDeploymentPipelineRequest) Validate() error {
	if r.TargetEnvironment != nil && !IsValidEnvironment(*r.TargetEnvironment) {
		return errors.New("invalid target_environment")
	}
	if r.DeploymentStrategy != nil && !IsValidDeploymentStrategy(*r.DeploymentStrategy) {
		return errors.New("invalid deployment_strategy")
	}
	if r.Status != nil && !IsValidPipelineStatus(*r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// CreateDeploymentRunRequest payload for initiating a deployment run
type CreateDeploymentRunRequest struct {
	WorkspaceID      string `json:"workspace_id"`
	PipelineID       string `json:"pipeline_id"`
	ReleaseVersion   string `json:"release_version"`
	Environment      string `json:"environment"`
	TriggerType      string `json:"trigger_type"`
	DeployedByUserID string `json:"deployed_by_user_id,omitempty"`
	CommitHash       string `json:"commit_hash,omitempty"`
	Changelog        string `json:"changelog,omitempty"`
}

func (r *CreateDeploymentRunRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.PipelineID) == "" {
		return errors.New("pipeline_id is required")
	}
	if strings.TrimSpace(r.ReleaseVersion) == "" {
		return errors.New("release_version is required")
	}
	if !IsValidEnvironment(r.Environment) {
		return errors.New("invalid environment")
	}
	if !IsValidRunTrigger(r.TriggerType) {
		return errors.New("invalid trigger_type")
	}
	return nil
}

// UpdateDeploymentRunStatusRequest payload for transitioning run status
type UpdateDeploymentRunStatusRequest struct {
	Status          string  `json:"status"`
	DurationSeconds float64 `json:"duration_seconds,omitempty"`
	RollbackReason  string  `json:"rollback_reason,omitempty"`
}

func (r *UpdateDeploymentRunStatusRequest) Validate() error {
	if !IsValidRunStatus(r.Status) {
		return errors.New("invalid run status")
	}
	return nil
}

// CreateDeploymentGateRequest payload for registering a deployment gate
type CreateDeploymentGateRequest struct {
	WorkspaceID string `json:"workspace_id"`
	RunID       string `json:"run_id"`
	GateType    string `json:"gate_type"`
	GateName    string `json:"gate_name"`
}

func (r *CreateDeploymentGateRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.RunID) == "" {
		return errors.New("run_id is required")
	}
	if !IsValidGateType(r.GateType) {
		return errors.New("invalid gate_type")
	}
	if strings.TrimSpace(r.GateName) == "" {
		return errors.New("gate_name is required")
	}
	return nil
}

// EvaluateDeploymentGateRequest payload for evaluating or waiving a gate
type EvaluateDeploymentGateRequest struct {
	Status          string `json:"status"`
	EvaluatorID     string `json:"evaluator_id"`
	WaiverReason    string `json:"waiver_reason,omitempty"`
	EvidenceDetails string `json:"evidence_details,omitempty"`
}

func (r *EvaluateDeploymentGateRequest) Validate() error {
	if !IsValidGateStatus(r.Status) {
		return errors.New("invalid gate status")
	}
	if strings.TrimSpace(r.EvaluatorID) == "" {
		return errors.New("evaluator_id is required")
	}
	if r.Status == GateStatusWaived && strings.TrimSpace(r.WaiverReason) == "" {
		return errors.New("waiver_reason is required when status is waived")
	}
	return nil
}

// CreateDeploymentStepRequest payload for adding a rollout stage step
type CreateDeploymentStepRequest struct {
	WorkspaceID        string  `json:"workspace_id"`
	RunID              string  `json:"run_id"`
	StepOrder          int     `json:"step_order"`
	StepName           string  `json:"step_name"`
	TrafficPercentage  float64 `json:"traffic_percentage"`
	SoakTimeMinutes    int     `json:"soak_time_minutes"`
	ErrorRateThreshold float64 `json:"error_rate_threshold"`
}

func (r *CreateDeploymentStepRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.RunID) == "" {
		return errors.New("run_id is required")
	}
	if r.StepOrder <= 0 {
		return errors.New("step_order must be greater than 0")
	}
	if strings.TrimSpace(r.StepName) == "" {
		return errors.New("step_name is required")
	}
	if r.TrafficPercentage < 0 || r.TrafficPercentage > 100 {
		return errors.New("traffic_percentage must be between 0 and 100")
	}
	return nil
}

// PromoteDeploymentStepRequest payload for advancing or aborting a step
type PromoteDeploymentStepRequest struct {
	Status          string  `json:"status"`
	ActualErrorRate float64 `json:"actual_error_rate"`
}

func (r *PromoteDeploymentStepRequest) Validate() error {
	if !IsValidStepStatus(r.Status) {
		return errors.New("invalid step status")
	}
	return nil
}

// CreateRollbackPolicyRequest payload for registering a rollback policy
type CreateRollbackPolicyRequest struct {
	WorkspaceID           string  `json:"workspace_id"`
	PipelineID            string  `json:"pipeline_id"`
	RunID                 string  `json:"run_id,omitempty"`
	TriggerCondition      string  `json:"trigger_condition"`
	ThresholdValue        float64 `json:"threshold_value"`
	TargetRollbackVersion string  `json:"target_rollback_version"`
}

func (r *CreateRollbackPolicyRequest) Validate() error {
	if strings.TrimSpace(r.WorkspaceID) == "" {
		return errors.New("workspace_id is required")
	}
	if strings.TrimSpace(r.PipelineID) == "" {
		return errors.New("pipeline_id is required")
	}
	if !IsValidRollbackTrigger(r.TriggerCondition) {
		return errors.New("invalid trigger_condition")
	}
	if strings.TrimSpace(r.TargetRollbackVersion) == "" {
		return errors.New("target_rollback_version is required")
	}
	return nil
}

// ExecuteRollbackRequest payload for executing or updating a rollback
type ExecuteRollbackRequest struct {
	ExecutionStatus string `json:"execution_status"`
	RollbackLogs    string `json:"rollback_logs,omitempty"`
}

func (r *ExecuteRollbackRequest) Validate() error {
	if !IsValidRollbackStatus(r.ExecutionStatus) {
		return errors.New("invalid execution_status")
	}
	return nil
}
