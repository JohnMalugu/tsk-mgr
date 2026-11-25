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

// Validation test case categories
const (
	TestCaseCategorySmoke       = "smoke"
	TestCaseCategoryIntegration = "integration"
	TestCaseCategorySecurity    = "security"
	TestCaseCategoryLoad        = "load"
	TestCaseCategoryFailover    = "failover"
	TestCaseCategoryE2E         = "e2e"
)

// Validation test case severities
const (
	TestCaseSeverityCritical = "critical"
	TestCaseSeverityHigh     = "high"
	TestCaseSeverityMedium   = "medium"
	TestCaseSeverityLow      = "low"
)

// Validation test case statuses
const (
	TestCaseStatusReady      = "ready"
	TestCaseStatusDraft      = "draft"
	TestCaseStatusDeprecated = "deprecated"
)

// IsValidTestCaseCategory validates test case category
func IsValidTestCaseCategory(cat string) bool {
	switch cat {
	case TestCaseCategorySmoke, TestCaseCategoryIntegration, TestCaseCategorySecurity,
		TestCaseCategoryLoad, TestCaseCategoryFailover, TestCaseCategoryE2E:
		return true
	default:
		return false
	}
}

// IsValidTestCaseSeverity validates test case severity rating
func IsValidTestCaseSeverity(sev string) bool {
	switch sev {
	case TestCaseSeverityCritical, TestCaseSeverityHigh, TestCaseSeverityMedium, TestCaseSeverityLow:
		return true
	default:
		return false
	}
}

// IsValidTestCaseStatus validates test case lifecycle status
func IsValidTestCaseStatus(status string) bool {
	switch status {
	case TestCaseStatusReady, TestCaseStatusDraft, TestCaseStatusDeprecated:
		return true
	default:
		return false
	}
}

// ValidationTestCase represents an individual verification scenario with steps and expectations
type ValidationTestCase struct {
	ID             string    `json:"id"`
	WorkspaceID    string    `json:"workspace_id"`
	PlanID         string    `json:"plan_id"`
	CaseCode       string    `json:"case_code"`
	Title          string    `json:"title"`
	Description    string    `json:"description,omitempty"`
	Preconditions  string    `json:"preconditions,omitempty"`
	TestSteps      string    `json:"test_steps"`
	ExpectedResult string    `json:"expected_result"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// Validation test run execution statuses
const (
	TestRunStatusPassed     = "passed"
	TestRunStatusFailed     = "failed"
	TestRunStatusBlocked    = "blocked"
	TestRunStatusSkipped    = "skipped"
	TestRunStatusInProgress = "in_progress"
)

// IsValidTestRunStatus validates execution outcome status
func IsValidTestRunStatus(status string) bool {
	switch status {
	case TestRunStatusPassed, TestRunStatusFailed, TestRunStatusBlocked,
		TestRunStatusSkipped, TestRunStatusInProgress:
		return true
	default:
		return false
	}
}

// ValidationTestRun records the execution instance of a test case against an environment
type ValidationTestRun struct {
	ID              string    `json:"id"`
	WorkspaceID     string    `json:"workspace_id"`
	PlanID          string    `json:"plan_id"`
	CaseID          string    `json:"case_id"`
	RunCode         string    `json:"run_code"`
	ExecutedBy      string    `json:"executed_by"`
	Status          string    `json:"status"`
	ActualResult    string    `json:"actual_result,omitempty"`
	ExecutionTimeMs int64     `json:"execution_time_ms"`
	ErrorMessage    string    `json:"error_message,omitempty"`
	StackTrace      string    `json:"stack_trace,omitempty"`
	EvidenceURL     string    `json:"evidence_url,omitempty"`
	DefectID        string    `json:"defect_id,omitempty"`
	ExecutedAt      time.Time `json:"executed_at"`
	CreatedAt       time.Time `json:"created_at"`
}

// Quality gate approval statuses
const (
	QualityGateStatusPending = "pending"
	QualityGateStatusPassed  = "passed"
	QualityGateStatusFailed  = "failed"
	QualityGateStatusWaived  = "waived"
)

// IsValidQualityGateStatus validates gate evaluation status
func IsValidQualityGateStatus(status string) bool {
	switch status {
	case QualityGateStatusPending, QualityGateStatusPassed, QualityGateStatusFailed, QualityGateStatusWaived:
		return true
	default:
		return false
	}
}

// QualityGate defines policy criteria and sign-off thresholds required before promoting releases
type QualityGate struct {
	ID                        string     `json:"id"`
	WorkspaceID               string     `json:"workspace_id"`
	Name                      string     `json:"name"`
	ReleaseID                 string     `json:"release_id,omitempty"`
	ChangeID                  string     `json:"change_id,omitempty"`
	MinPassRate               float64    `json:"min_pass_rate"`
	RequireSecuritySignoff    bool       `json:"require_security_signoff"`
	RequirePerformanceSignoff bool       `json:"require_performance_signoff"`
	Status                    string     `json:"status"`
	WaivedBy                  string     `json:"waived_by,omitempty"`
	WaivedReason              string     `json:"waived_reason,omitempty"`
	DecidedAt                 *time.Time `json:"decided_at,omitempty"`
	CreatedAt                 time.Time  `json:"created_at"`
	UpdatedAt                 time.Time  `json:"updated_at"`
}

// Validation defect lifecycle statuses
const (
	DefectStatusOpen          = "open"
	DefectStatusInvestigating = "investigating"
	DefectStatusResolved      = "resolved"
)

// IsValidDefectStatus validates defect status
func IsValidDefectStatus(status string) bool {
	switch status {
	case DefectStatusOpen, DefectStatusInvestigating, DefectStatusResolved:
		return true
	default:
		return false
	}
}

// ValidationDefect records defects identified during test executions and links to Incidents or Problems
type ValidationDefect struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	RunID       string    `json:"run_id"`
	IncidentID  string    `json:"incident_id,omitempty"`
	ProblemID   string    `json:"problem_id,omitempty"`
	Severity    string    `json:"severity"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	LoggedAt    time.Time `json:"logged_at"`
}
