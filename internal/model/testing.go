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

// Request DTOs and validation rules for Service Validation and Testing API

// CreateTestPlanRequest payload for authoring a validation test plan
type CreateTestPlanRequest struct {
	WorkspaceID    string  `json:"workspace_id"`
	PlanCode       string  `json:"plan_code"`
	Title          string  `json:"title"`
	Description    string  `json:"description,omitempty"`
	ReleaseID      string  `json:"release_id,omitempty"`
	ChangeID       string  `json:"change_id,omitempty"`
	CIID           string  `json:"ci_id,omitempty"`
	TestType       string  `json:"test_type"`
	Environment    string  `json:"environment"`
	TargetPassRate float64 `json:"target_pass_rate"`
	CreatedBy      string  `json:"created_by,omitempty"`
}

func (r *CreateTestPlanRequest) Validate() error {
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if r.TestType != "" && !IsValidTestType(r.TestType) {
		return errors.New("invalid test_type")
	}
	if r.Environment != "" && !IsValidTestEnvironment(r.Environment) {
		return errors.New("invalid environment")
	}
	if r.TargetPassRate < 0 || r.TargetPassRate > 100 {
		return errors.New("target_pass_rate must be between 0 and 100")
	}
	return nil
}

// UpdateTestPlanRequest payload for modifying a test plan
type UpdateTestPlanRequest struct {
	Title          *string  `json:"title,omitempty"`
	Description    *string  `json:"description,omitempty"`
	ReleaseID      *string  `json:"release_id,omitempty"`
	ChangeID       *string  `json:"change_id,omitempty"`
	CIID           *string  `json:"ci_id,omitempty"`
	TestType       *string  `json:"test_type,omitempty"`
	Environment    *string  `json:"environment,omitempty"`
	Status         *string  `json:"status,omitempty"`
	TargetPassRate *float64 `json:"target_pass_rate,omitempty"`
}

func (r *UpdateTestPlanRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if r.TestType != nil && !IsValidTestType(*r.TestType) {
		return errors.New("invalid test_type")
	}
	if r.Environment != nil && !IsValidTestEnvironment(*r.Environment) {
		return errors.New("invalid environment")
	}
	if r.Status != nil && !IsValidTestPlanStatus(*r.Status) {
		return errors.New("invalid status")
	}
	if r.TargetPassRate != nil && (*r.TargetPassRate < 0 || *r.TargetPassRate > 100) {
		return errors.New("target_pass_rate must be between 0 and 100")
	}
	return nil
}

// CreateTestCaseRequest payload for authoring a test case
type CreateTestCaseRequest struct {
	WorkspaceID    string `json:"workspace_id"`
	PlanID         string `json:"plan_id"`
	CaseCode       string `json:"case_code"`
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	Preconditions  string `json:"preconditions,omitempty"`
	TestSteps      string `json:"test_steps"`
	ExpectedResult string `json:"expected_result"`
	Severity       string `json:"severity"`
	Category       string `json:"category"`
}

func (r *CreateTestCaseRequest) Validate() error {
	if strings.TrimSpace(r.PlanID) == "" {
		return errors.New("plan_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	if strings.TrimSpace(r.TestSteps) == "" {
		return errors.New("test_steps is required")
	}
	if strings.TrimSpace(r.ExpectedResult) == "" {
		return errors.New("expected_result is required")
	}
	if r.Severity != "" && !IsValidTestCaseSeverity(r.Severity) {
		return errors.New("invalid severity")
	}
	if r.Category != "" && !IsValidTestCaseCategory(r.Category) {
		return errors.New("invalid category")
	}
	return nil
}

// UpdateTestCaseRequest payload for updating a test case
type UpdateTestCaseRequest struct {
	Title          *string `json:"title,omitempty"`
	Description    *string `json:"description,omitempty"`
	Preconditions  *string `json:"preconditions,omitempty"`
	TestSteps      *string `json:"test_steps,omitempty"`
	ExpectedResult *string `json:"expected_result,omitempty"`
	Severity       *string `json:"severity,omitempty"`
	Category       *string `json:"category,omitempty"`
	Status         *string `json:"status,omitempty"`
}

func (r *UpdateTestCaseRequest) Validate() error {
	if r.Title != nil && strings.TrimSpace(*r.Title) == "" {
		return errors.New("title cannot be empty")
	}
	if r.Severity != nil && !IsValidTestCaseSeverity(*r.Severity) {
		return errors.New("invalid severity")
	}
	if r.Category != nil && !IsValidTestCaseCategory(*r.Category) {
		return errors.New("invalid category")
	}
	if r.Status != nil && !IsValidTestCaseStatus(*r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// ExecuteTestRunRequest payload for recording a test execution
type ExecuteTestRunRequest struct {
	WorkspaceID     string `json:"workspace_id"`
	PlanID          string `json:"plan_id"`
	CaseID          string `json:"case_id"`
	ExecutedBy      string `json:"executed_by"`
	Status          string `json:"status"`
	ActualResult    string `json:"actual_result,omitempty"`
	ExecutionTimeMs int64  `json:"execution_time_ms"`
	ErrorMessage    string `json:"error_message,omitempty"`
	StackTrace      string `json:"stack_trace,omitempty"`
	EvidenceURL     string `json:"evidence_url,omitempty"`
	DefectID        string `json:"defect_id,omitempty"`
}

func (r *ExecuteTestRunRequest) Validate() error {
	if strings.TrimSpace(r.PlanID) == "" {
		return errors.New("plan_id is required")
	}
	if strings.TrimSpace(r.CaseID) == "" {
		return errors.New("case_id is required")
	}
	if !IsValidTestRunStatus(r.Status) {
		return errors.New("invalid status")
	}
	return nil
}

// CreateQualityGateRequest payload for defining quality gate threshold
type CreateQualityGateRequest struct {
	WorkspaceID               string  `json:"workspace_id"`
	Name                      string  `json:"name"`
	ReleaseID                 string  `json:"release_id,omitempty"`
	ChangeID                  string  `json:"change_id,omitempty"`
	MinPassRate               float64 `json:"min_pass_rate"`
	RequireSecuritySignoff    bool    `json:"require_security_signoff"`
	RequirePerformanceSignoff bool    `json:"require_performance_signoff"`
}

func (r *CreateQualityGateRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if r.MinPassRate < 0 || r.MinPassRate > 100 {
		return errors.New("min_pass_rate must be between 0 and 100")
	}
	return nil
}

// WaiveQualityGateRequest payload for overriding quality gate
type WaiveQualityGateRequest struct {
	WaivedBy     string `json:"waived_by"`
	WaivedReason string `json:"waived_reason"`
}

func (r *WaiveQualityGateRequest) Validate() error {
	if strings.TrimSpace(r.WaivedBy) == "" {
		return errors.New("waived_by is required")
	}
	if strings.TrimSpace(r.WaivedReason) == "" {
		return errors.New("waived_reason is required")
	}
	return nil
}

// LinkDefectRequest payload for linking test run failure to defect
type LinkDefectRequest struct {
	WorkspaceID string `json:"workspace_id"`
	RunID       string `json:"run_id"`
	IncidentID  string `json:"incident_id,omitempty"`
	ProblemID   string `json:"problem_id,omitempty"`
	Severity    string `json:"severity"`
	Title       string `json:"title"`
}

func (r *LinkDefectRequest) Validate() error {
	if strings.TrimSpace(r.RunID) == "" {
		return errors.New("run_id is required")
	}
	if strings.TrimSpace(r.Title) == "" {
		return errors.New("title is required")
	}
	return nil
}

// Query filters and posture analytics models for Service Validation & Testing

// TestPlanFilter defines query criteria for browsing test plans
type TestPlanFilter struct {
	WorkspaceID string
	ReleaseID   string
	ChangeID    string
	TestType    string
	Environment string
	Status      string
	Search      string
	Limit       int
	Offset      int
}

// TestCaseFilter defines query criteria for browsing test cases
type TestCaseFilter struct {
	WorkspaceID string
	PlanID      string
	Category    string
	Severity    string
	Status      string
	Search      string
	Limit       int
	Offset      int
}

// TestRunFilter defines query criteria for test execution logs
type TestRunFilter struct {
	WorkspaceID string
	PlanID      string
	CaseID      string
	Status      string
	ExecutedBy  string
	Limit       int
	Offset      int
}

// QualityGateFilter defines query criteria for quality gates
type QualityGateFilter struct {
	WorkspaceID string
	ReleaseID   string
	ChangeID    string
	Status      string
	Limit       int
	Offset      int
}

// DefectLinkFilter defines query criteria for validation defects
type DefectLinkFilter struct {
	WorkspaceID string
	RunID       string
	IncidentID  string
	ProblemID   string
	Status      string
	Limit       int
	Offset      int
}

// TestingAnalytics represents executive Service Validation & Testing posture reporting
type TestingAnalytics struct {
	TotalPlans              int            `json:"total_plans"`
	TotalCases              int            `json:"total_cases"`
	TotalRuns               int            `json:"total_runs"`
	PassedRuns              int            `json:"passed_runs"`
	FailedRuns              int            `json:"failed_runs"`
	BlockedRuns             int            `json:"blocked_runs"`
	SkippedRuns             int            `json:"skipped_runs"`
	OverallPassRate         float64        `json:"overall_pass_rate"`
	DefectDensity           float64        `json:"defect_density"`
	QualityGateCompliance   float64        `json:"quality_gate_compliance"`
	TestTypeDistribution    map[string]int `json:"test_type_distribution"`
	EnvironmentDistribution map[string]int `json:"environment_distribution"`
	QualityHealthIndex      float64        `json:"quality_health_index"`
}
