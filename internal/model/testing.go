package model

import (
	"errors"
	"strings"
	"time"
)

var (
	_ = errors.New
	_ = strings.TrimSpace
	_ = time.Now
)

// Validation test types
const (
	TestTypeFunctional = "functional"
	TestTypePerformance = "performance"
	TestTypeSecurity    = "security"
	TestTypeRegression  = "regression"
	TestTypeUAT         = "uat"
	TestTypeResilience  = "resilience"
)

// Validation test plan statuses
const (
	TestPlanStatusDraft      = "draft"
	TestPlanStatusInProgress = "in_progress"
	TestPlanStatusCompleted  = "completed"
	TestPlanStatusAborted    = "aborted"
)

// Validation test environments
const (
	TestEnvStaging    = "staging"
	TestEnvPreProd    = "preprod"
	TestEnvProduction = "production"
	TestEnvSandbox    = "sandbox"
)

// IsValidTestType validates test strategy type
func IsValidTestType(tt string) bool {
	switch tt {
	case TestTypeFunctional, TestTypePerformance, TestTypeSecurity,
		TestTypeRegression, TestTypeUAT, TestTypeResilience:
		return true
	default:
		return false
	}
}

// IsValidTestPlanStatus validates test plan lifecycle status
func IsValidTestPlanStatus(ps string) bool {
	switch ps {
	case TestPlanStatusDraft, TestPlanStatusInProgress, TestPlanStatusCompleted, TestPlanStatusAborted:
		return true
	default:
		return false
	}
}

// IsValidTestEnvironment validates execution environment
func IsValidTestEnvironment(env string) bool {
	switch env {
	case TestEnvStaging, TestEnvPreProd, TestEnvProduction, TestEnvSandbox:
		return true
	default:
		return false
	}
}

// ValidationTestPlan represents an end-to-end service validation strategy or test suite
type ValidationTestPlan struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	PlanCode       string    `json:"plan_code"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	ReleaseID      string    `json:"release_id,omitempty"`
	ChangeID       string    `json:"change_id,omitempty"`
	CIID           string    `json:"ci_id,omitempty"`
	TestType       string    `json:"test_type"`
	Environment    string    `json:"environment"`
	Status         string    `json:"status"`
	TargetPassRate float64   `json:"target_pass_rate"`
	CreatedBy      string    `json:"created_by,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
