package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	_ = fmt.Sprintf
	_ = strings.TrimSpace
	_ = time.Now
	_ = errors.New
)

var (
	ErrTestPlanNotFound        = repository.ErrTestPlanNotFound
	ErrTestPlanCodeConflict    = repository.ErrTestPlanCodeConflict
	ErrTestCaseNotFound        = repository.ErrTestCaseNotFound
	ErrTestCaseCodeConflict    = repository.ErrTestCaseCodeConflict
	ErrTestRunNotFound         = repository.ErrTestRunNotFound
	ErrQualityGateNotFound     = repository.ErrQualityGateNotFound
	ErrQualityGateNameConflict = repository.ErrQualityGateNameConflict
	ErrDefectLinkNotFound      = repository.ErrDefectLinkNotFound
	ErrInvalidGateState        = errors.New("invalid quality gate state")
)

// TestingService defines business logic and orchestration for ITIL Service Validation & Testing
type TestingService interface {
	// Test Plans
	CreatePlan(ctx context.Context, workspaceID string, req model.CreateTestPlanRequest) (*model.ValidationTestPlan, error)
	GetPlanByID(ctx context.Context, id string) (*model.ValidationTestPlan, error)
	GetPlanByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestPlan, error)
	UpdatePlan(ctx context.Context, id string, req model.UpdateTestPlanRequest) (*model.ValidationTestPlan, error)
	DeletePlan(ctx context.Context, id string) error
	ListPlans(ctx context.Context, filter model.TestPlanFilter) ([]model.ValidationTestPlan, int, error)

	// Test Cases
	CreateCase(ctx context.Context, workspaceID string, req model.CreateTestCaseRequest) (*model.ValidationTestCase, error)
	GetCaseByID(ctx context.Context, id string) (*model.ValidationTestCase, error)
	GetCaseByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestCase, error)
	UpdateCase(ctx context.Context, id string, req model.UpdateTestCaseRequest) (*model.ValidationTestCase, error)
	DeleteCase(ctx context.Context, id string) error
	ListCases(ctx context.Context, filter model.TestCaseFilter) ([]model.ValidationTestCase, int, error)

	// Test Runs
	ExecuteRun(ctx context.Context, workspaceID string, req model.ExecuteTestRunRequest) (*model.ValidationTestRun, error)
	GetRunByID(ctx context.Context, id string) (*model.ValidationTestRun, error)
	ListRuns(ctx context.Context, filter model.TestRunFilter) ([]model.ValidationTestRun, int, error)

	// Quality Gates
	CreateQualityGate(ctx context.Context, workspaceID string, req model.CreateQualityGateRequest) (*model.QualityGate, error)
	GetQualityGateByID(ctx context.Context, id string) (*model.QualityGate, error)
	EvaluateQualityGate(ctx context.Context, id string) (*model.QualityGate, error)
	WaiveQualityGate(ctx context.Context, id string, req model.WaiveQualityGateRequest) (*model.QualityGate, error)
	DeleteQualityGate(ctx context.Context, id string) error
	ListQualityGates(ctx context.Context, filter model.QualityGateFilter) ([]model.QualityGate, int, error)

	// Defect Links
	LinkDefect(ctx context.Context, workspaceID string, req model.LinkDefectRequest) (*model.ValidationDefect, error)
	ListDefects(ctx context.Context, filter model.DefectLinkFilter) ([]model.ValidationDefect, int, error)

	// Posture Analytics
	GetTestingAnalytics(ctx context.Context, workspaceID string) (*model.TestingAnalytics, error)
}

type testingService struct {
	repo repository.TestingRepository
}

// NewTestingService constructs an ITIL Service Validation & Testing service
func NewTestingService(repo repository.TestingRepository) TestingService {
	return &testingService{repo: repo}
}

// CreatePlan validates and persists a new validation test plan
func (s *testingService) CreatePlan(ctx context.Context, workspaceID string, req model.CreateTestPlanRequest) (*model.ValidationTestPlan, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	planCode := strings.TrimSpace(req.PlanCode)
	if planCode == "" {
		planCode = fmt.Sprintf("TP-%04d", time.Now().UnixNano()%10000)
	}

	testType := req.TestType
	if testType == "" {
		testType = model.TestTypeFunctional
	}
	env := req.Environment
	if env == "" {
		env = model.TestEnvStaging
	}
	passRate := req.TargetPassRate
	if passRate <= 0 {
		passRate = 95.0
	}

	plan := &model.ValidationTestPlan{
		WorkspaceID:    req.WorkspaceID,
		PlanCode:       planCode,
		Title:          req.Title,
		Description:    req.Description,
		ReleaseID:      req.ReleaseID,
		ChangeID:       req.ChangeID,
		CIID:           req.CIID,
		TestType:       testType,
		Environment:    env,
		Status:         model.TestPlanStatusDraft,
		TargetPassRate: passRate,
		CreatedBy:      req.CreatedBy,
	}

	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// GetPlanByID retrieves a test plan by ID
func (s *testingService) GetPlanByID(ctx context.Context, id string) (*model.ValidationTestPlan, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTestPlanNotFound
	}
	return s.repo.GetPlanByID(ctx, id)
}

// GetPlanByCode retrieves a test plan by code within a workspace
func (s *testingService) GetPlanByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestPlan, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrTestPlanNotFound
	}
	return s.repo.GetPlanByCode(ctx, workspaceID, code)
}

// UpdatePlan modifies an existing validation test plan
func (s *testingService) UpdatePlan(ctx context.Context, id string, req model.UpdateTestPlanRequest) (*model.ValidationTestPlan, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTestPlanNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	plan, err := s.repo.GetPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		plan.Title = *req.Title
	}
	if req.Description != nil {
		plan.Description = *req.Description
	}
	if req.ReleaseID != nil {
		plan.ReleaseID = *req.ReleaseID
	}
	if req.ChangeID != nil {
		plan.ChangeID = *req.ChangeID
	}
	if req.CIID != nil {
		plan.CIID = *req.CIID
	}
	if req.TestType != nil {
		plan.TestType = *req.TestType
	}
	if req.Environment != nil {
		plan.Environment = *req.Environment
	}
	if req.Status != nil {
		plan.Status = *req.Status
	}
	if req.TargetPassRate != nil {
		plan.TargetPassRate = *req.TargetPassRate
	}

	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

// DeletePlan deletes a validation test plan by ID
func (s *testingService) DeletePlan(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrTestPlanNotFound
	}
	return s.repo.DeletePlan(ctx, id)
}

// ListPlans queries test plans based on filter criteria
func (s *testingService) ListPlans(ctx context.Context, filter model.TestPlanFilter) ([]model.ValidationTestPlan, int, error) {
	return s.repo.ListPlans(ctx, filter)
}

// CreateCase validates and creates a new test case linked to a test plan
func (s *testingService) CreateCase(ctx context.Context, workspaceID string, req model.CreateTestCaseRequest) (*model.ValidationTestCase, error) {
	if req.WorkspaceID == "" {
		req.WorkspaceID = workspaceID
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify plan exists
	_, err := s.repo.GetPlanByID(ctx, req.PlanID)
	if err != nil {
		return nil, err
	}

	caseCode := strings.TrimSpace(req.CaseCode)
	if caseCode == "" {
		caseCode = fmt.Sprintf("TC-%04d", time.Now().UnixNano()%10000)
	}

	severity := req.Severity
	if severity == "" {
		severity = model.TestCaseSeverityMedium
	}
	category := req.Category
	if category == "" {
		category = model.TestCaseCategoryIntegration
	}

	testCase := &model.ValidationTestCase{
		WorkspaceID:    req.WorkspaceID,
		PlanID:         req.PlanID,
		CaseCode:       caseCode,
		Title:          req.Title,
		Description:    req.Description,
		Preconditions:  req.Preconditions,
		TestSteps:      req.TestSteps,
		ExpectedResult: req.ExpectedResult,
		Severity:       severity,
		Category:       category,
		Status:         model.TestCaseStatusReady,
	}

	if err := s.repo.CreateCase(ctx, testCase); err != nil {
		return nil, err
	}
	return testCase, nil
}

// GetCaseByID retrieves a test case by ID
func (s *testingService) GetCaseByID(ctx context.Context, id string) (*model.ValidationTestCase, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTestCaseNotFound
	}
	return s.repo.GetCaseByID(ctx, id)
}

// GetCaseByCode retrieves a test case by code within a workspace
func (s *testingService) GetCaseByCode(ctx context.Context, workspaceID, code string) (*model.ValidationTestCase, error) {
	if strings.TrimSpace(code) == "" {
		return nil, ErrTestCaseNotFound
	}
	return s.repo.GetCaseByCode(ctx, workspaceID, code)
}

// UpdateCase modifies an existing test case
func (s *testingService) UpdateCase(ctx context.Context, id string, req model.UpdateTestCaseRequest) (*model.ValidationTestCase, error) {
	if strings.TrimSpace(id) == "" {
		return nil, ErrTestCaseNotFound
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}

	testCase, err := s.repo.GetCaseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		testCase.Title = *req.Title
	}
	if req.Description != nil {
		testCase.Description = *req.Description
	}
	if req.Preconditions != nil {
		testCase.Preconditions = *req.Preconditions
	}
	if req.TestSteps != nil {
		testCase.TestSteps = *req.TestSteps
	}
	if req.ExpectedResult != nil {
		testCase.ExpectedResult = *req.ExpectedResult
	}
	if req.Severity != nil {
		testCase.Severity = *req.Severity
	}
	if req.Category != nil {
		testCase.Category = *req.Category
	}
	if req.Status != nil {
		testCase.Status = *req.Status
	}

	if err := s.repo.UpdateCase(ctx, testCase); err != nil {
		return nil, err
	}
	return testCase, nil
}

// DeleteCase removes a test case by ID
func (s *testingService) DeleteCase(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrTestCaseNotFound
	}
	return s.repo.DeleteCase(ctx, id)
}

// ListCases queries test cases based on filter criteria
func (s *testingService) ListCases(ctx context.Context, filter model.TestCaseFilter) ([]model.ValidationTestCase, int, error) {
	return s.repo.ListCases(ctx, filter)
}

// Runs stubs
func (s *testingService) ExecuteRun(ctx context.Context, workspaceID string, req model.ExecuteTestRunRequest) (*model.ValidationTestRun, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) GetRunByID(ctx context.Context, id string) (*model.ValidationTestRun, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) ListRuns(ctx context.Context, filter model.TestRunFilter) ([]model.ValidationTestRun, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Quality Gates stubs
func (s *testingService) CreateQualityGate(ctx context.Context, workspaceID string, req model.CreateQualityGateRequest) (*model.QualityGate, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) GetQualityGateByID(ctx context.Context, id string) (*model.QualityGate, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) EvaluateQualityGate(ctx context.Context, id string) (*model.QualityGate, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) WaiveQualityGate(ctx context.Context, id string, req model.WaiveQualityGateRequest) (*model.QualityGate, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) DeleteQualityGate(ctx context.Context, id string) error {
	return errors.New("not implemented")
}
func (s *testingService) ListQualityGates(ctx context.Context, filter model.QualityGateFilter) ([]model.QualityGate, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Defects stubs
func (s *testingService) LinkDefect(ctx context.Context, workspaceID string, req model.LinkDefectRequest) (*model.ValidationDefect, error) {
	return nil, errors.New("not implemented")
}
func (s *testingService) ListDefects(ctx context.Context, filter model.DefectLinkFilter) ([]model.ValidationDefect, int, error) {
	return nil, 0, errors.New("not implemented")
}

// Posture Analytics stub
func (s *testingService) GetTestingAnalytics(ctx context.Context, workspaceID string) (*model.TestingAnalytics, error) {
	return nil, errors.New("not implemented")
}
