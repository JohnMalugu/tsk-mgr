package model

import (
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
