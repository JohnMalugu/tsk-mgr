package service

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var _ = time.Now

func setupTestTestingSvc(t *testing.T) (TestingService, repository.TestingRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}
	sqliteDB.SetMaxOpenConns(1)

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Testing WS', 'test-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewTestingRepository(sqliteDB)
	svc := NewTestingService(repo)
	return svc, repo, sqliteDB
}

func TestTesting_PlanLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing title
	_, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		TestType: model.TestTypeFunctional,
	})
	if err == nil {
		t.Fatal("expected error for missing title, got nil")
	}

	// 2. Validation failure: invalid test type
	_, err = svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		Title:    "Core Payment Validation",
		TestType: "invalid_type",
	})
	if err == nil {
		t.Fatal("expected error for invalid test type, got nil")
	}

	// 3. Create plan successfully
	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode:       "TP-1001",
		Title:          "Core Payment API Regression",
		Description:    "Full functional and regression test suite for checkout and billing",
		TestType:       model.TestTypeRegression,
		Environment:    model.TestEnvPreProd,
		TargetPassRate: 98.0,
		CreatedBy:      "user_1",
	})
	if err != nil {
		t.Fatalf("failed to create test plan: %v", err)
	}
	if plan.ID == "" || plan.Status != model.TestPlanStatusDraft {
		t.Fatalf("expected plan with ID and draft status, got %+v", plan)
	}

	// 4. Duplicate plan_code within workspace conflict
	_, err = svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode: "TP-1001",
		Title:    "Duplicate plan code test",
	})
	if err == nil || err != ErrTestPlanCodeConflict {
		t.Fatalf("expected ErrTestPlanCodeConflict, got %v", err)
	}

	// 5. GetPlanByID and GetPlanByCode
	gotPlan, err := svc.GetPlanByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("failed to get plan by ID: %v", err)
	}
	if gotPlan.Title != "Core Payment API Regression" {
		t.Fatalf("expected title 'Core Payment API Regression', got '%s'", gotPlan.Title)
	}

	byCode, err := svc.GetPlanByCode(ctx, "ws_1", "TP-1001")
	if err != nil {
		t.Fatalf("failed to get plan by code: %v", err)
	}
	if byCode.ID != plan.ID {
		t.Fatalf("expected ID %s, got %s", plan.ID, byCode.ID)
	}

	// 6. Update plan
	newTitle := "Updated Core Payment Regression"
	updatedPlan, err := svc.UpdatePlan(ctx, plan.ID, model.UpdateTestPlanRequest{
		Title: &newTitle,
	})
	if err != nil {
		t.Fatalf("failed to update plan: %v", err)
	}
	if updatedPlan.Title != newTitle {
		t.Fatalf("expected updated title '%s', got '%s'", newTitle, updatedPlan.Title)
	}

	// 7. List plans
	plans, total, err := svc.ListPlans(ctx, model.TestPlanFilter{
		WorkspaceID: "ws_1",
		TestType:    model.TestTypeRegression,
	})
	if err != nil || total != 1 || len(plans) != 1 {
		t.Fatalf("expected 1 plan in list, got total %d, count %d", total, len(plans))
	}

	// 8. Delete plan
	if err := svc.DeletePlan(ctx, plan.ID); err != nil {
		t.Fatalf("failed to delete plan: %v", err)
	}
	_, err = svc.GetPlanByID(ctx, plan.ID)
	if err == nil {
		t.Fatal("expected error getting deleted plan, got nil")
	}
}

func TestTesting_CaseManagement(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create test plan first
	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode: "TP-2001",
		Title:    "Identity & Auth Suite",
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// 1. Validation failure: missing steps
	_, err = svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID: plan.ID,
		Title:  "Login with valid JWT",
	})
	if err == nil {
		t.Fatal("expected error for missing test steps, got nil")
	}

	// 2. Create test case successfully
	tc, err := svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-101",
		Title:          "Authenticate with OAuth2 Bearer Token",
		Description:    "Verify that valid tokens return HTTP 200 with claims",
		Preconditions:  "User account exists and is active",
		TestSteps:      "1. Send POST /auth/token\n2. Verify 200 OK\n3. Call /v2/me with token",
		ExpectedResult: "HTTP 200 with valid user claims",
		Severity:       model.TestCaseSeverityCritical,
		Category:       model.TestCaseCategorySecurity,
	})
	if err != nil {
		t.Fatalf("failed to create test case: %v", err)
	}
	if tc.ID == "" || tc.Status != model.TestCaseStatusReady {
		t.Fatalf("expected valid test case with ready status, got %+v", tc)
	}

	// 3. Duplicate case_code conflict within workspace
	_, err = svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-101",
		Title:          "Duplicate code test case",
		TestSteps:      "1. Steps",
		ExpectedResult: "Pass",
	})
	if err == nil || err != ErrTestCaseCodeConflict {
		t.Fatalf("expected ErrTestCaseCodeConflict, got %v", err)
	}

	// 4. GetCaseByID and GetCaseByCode
	gotCase, err := svc.GetCaseByID(ctx, tc.ID)
	if err != nil {
		t.Fatalf("failed to get case by ID: %v", err)
	}
	if gotCase.Severity != model.TestCaseSeverityCritical {
		t.Fatalf("expected severity 'critical', got '%s'", gotCase.Severity)
	}

	byCode, err := svc.GetCaseByCode(ctx, "ws_1", "TC-101")
	if err != nil {
		t.Fatalf("failed to get case by code: %v", err)
	}
	if byCode.ID != tc.ID {
		t.Fatalf("expected case ID %s, got %s", tc.ID, byCode.ID)
	}

	// 5. Update case
	newDesc := "Updated case description"
	updatedCase, err := svc.UpdateCase(ctx, tc.ID, model.UpdateTestCaseRequest{
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("failed to update case: %v", err)
	}
	if updatedCase.Description != newDesc {
		t.Fatalf("expected description '%s', got '%s'", newDesc, updatedCase.Description)
	}

	// 6. List cases
	cases, total, err := svc.ListCases(ctx, model.TestCaseFilter{
		WorkspaceID: "ws_1",
		PlanID:      plan.ID,
	})
	if err != nil || total != 1 || len(cases) != 1 {
		t.Fatalf("expected 1 case in list, got total %d", total)
	}

	// 7. Delete case
	if err := svc.DeleteCase(ctx, tc.ID); err != nil {
		t.Fatalf("failed to delete case: %v", err)
	}
	_, err = svc.GetCaseByID(ctx, tc.ID)
	if err == nil {
		t.Fatal("expected error getting deleted case, got nil")
	}
}

func TestTesting_RunExecutionAndAutoDefect(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode: "TP-3001",
		Title:    "Payment Processing Suite",
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	testCase, err := svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-301",
		Title:          "Process Stripe Charge",
		TestSteps:      "1. POST /payments/charge",
		ExpectedResult: "HTTP 200 and charge ID returned",
		Severity:       model.TestCaseSeverityCritical,
	})
	if err != nil {
		t.Fatalf("failed to create case: %v", err)
	}

	// 1. Execute successful test run
	run1, err := svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
		PlanID:          plan.ID,
		CaseID:          testCase.ID,
		ExecutedBy:      "alice",
		Status:          model.TestRunStatusPassed,
		ActualResult:    "Charge processed in 120ms with ID ch_test123",
		ExecutionTimeMs: 120,
	})
	if err != nil {
		t.Fatalf("failed to execute run 1: %v", err)
	}
	if run1.ID == "" || run1.Status != model.TestRunStatusPassed {
		t.Fatalf("expected passed run, got %+v", run1)
	}

	// Verify plan was advanced from draft to in_progress upon first run
	refreshedPlan, err := svc.GetPlanByID(ctx, plan.ID)
	if err != nil {
		t.Fatalf("failed to refresh plan: %v", err)
	}
	if refreshedPlan.Status != model.TestPlanStatusInProgress {
		t.Fatalf("expected plan status 'in_progress', got '%s'", refreshedPlan.Status)
	}

	// 2. Execute failing test run -> should auto-create defect link
	run2, err := svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
		PlanID:          plan.ID,
		CaseID:          testCase.ID,
		ExecutedBy:      "bob",
		Status:          model.TestRunStatusFailed,
		ActualResult:    "Received HTTP 500 Gateway Timeout from Stripe webhook",
		ExecutionTimeMs: 5000,
		ErrorMessage:    "Upstream payment gateway timeout",
	})
	if err != nil {
		t.Fatalf("failed to execute run 2: %v", err)
	}
	if run2.DefectID == "" {
		t.Fatalf("expected run2 to have an auto-linked defect_id, got empty")
	}

	// Verify defect record exists
	defects, totalDefects, err := svc.ListDefects(ctx, model.DefectLinkFilter{
		WorkspaceID: "ws_1",
		RunID:       run2.ID,
	})
	if err != nil || totalDefects != 1 || len(defects) != 1 {
		t.Fatalf("expected 1 defect link, got total %d, count %d", totalDefects, len(defects))
	}
	if defects[0].Severity != model.TestCaseSeverityCritical {
		t.Fatalf("expected defect severity critical, got %s", defects[0].Severity)
	}

	// 3. List runs
	runs, totalRuns, err := svc.ListRuns(ctx, model.TestRunFilter{
		WorkspaceID: "ws_1",
		PlanID:      plan.ID,
	})
	if err != nil || totalRuns != 2 || len(runs) != 2 {
		t.Fatalf("expected 2 runs, got total %d", totalRuns)
	}
}

func TestTesting_QualityGateLifecycleAndWaiver(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create quality gate with 95% threshold
	gate, err := svc.CreateQualityGate(ctx, "ws_1", model.CreateQualityGateRequest{
		Name:                      "Production Deployment Gate",
		ReleaseID:                 "REL-101",
		MinPassRate:               95.0,
		RequireSecuritySignoff:    true,
		RequirePerformanceSignoff: true,
	})
	if err != nil {
		t.Fatalf("failed to create quality gate: %v", err)
	}
	if gate.ID == "" || gate.Status != model.QualityGateStatusPending {
		t.Fatalf("expected pending gate, got %+v", gate)
	}

	// Create plan and failing run to simulate low pass rate
	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode: "TP-4001",
		Title:    "Canary Suite",
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	testCase, err := svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-401",
		Title:          "Canary Health Check",
		TestSteps:      "1. Ping health",
		ExpectedResult: "200 OK",
	})
	if err != nil {
		t.Fatalf("failed to create case: %v", err)
	}

	// 1 passed, 1 failed run -> 50% pass rate (below 95%)
	_, _ = svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
		PlanID: plan.ID, CaseID: testCase.ID, Status: model.TestRunStatusPassed,
	})
	_, _ = svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
		PlanID: plan.ID, CaseID: testCase.ID, Status: model.TestRunStatusFailed, ErrorMessage: "Canary unhealthy",
	})

	// Evaluate quality gate -> should fail
	evaluatedGate, err := svc.EvaluateQualityGate(ctx, gate.ID)
	if err != nil {
		t.Fatalf("failed to evaluate gate: %v", err)
	}
	if evaluatedGate.Status != model.QualityGateStatusFailed {
		t.Fatalf("expected gate status 'failed', got '%s'", evaluatedGate.Status)
	}

	// Waive quality gate with justification
	waivedGate, err := svc.WaiveQualityGate(ctx, gate.ID, model.WaiveQualityGateRequest{
		WaivedBy:     "alice_vp_eng",
		WaivedReason: "Known non-blocking edge case covered by hotfix HF-992",
	})
	if err != nil {
		t.Fatalf("failed to waive quality gate: %v", err)
	}
	if waivedGate.Status != model.QualityGateStatusWaived {
		t.Fatalf("expected gate status 'waived', got '%s'", waivedGate.Status)
	}
	if waivedGate.WaivedBy != "alice_vp_eng" || waivedGate.DecidedAt == nil {
		t.Fatalf("expected waived_by and decided_at to be populated")
	}

	// List gates
	gates, total, err := svc.ListQualityGates(ctx, model.QualityGateFilter{
		WorkspaceID: "ws_1",
	})
	if err != nil || total != 1 || len(gates) != 1 {
		t.Fatalf("expected 1 gate, got total %d", total)
	}
}

func TestTesting_Analytics(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode:    "TP-5001",
		Title:       "Analytics Suite",
		TestType:    model.TestTypeFunctional,
		Environment: model.TestEnvStaging,
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	testCase, err := svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-501",
		Title:          "Basic Assert",
		TestSteps:      "Assert true",
		ExpectedResult: "Pass",
	})
	if err != nil {
		t.Fatalf("failed to create case: %v", err)
	}

	// Execute 3 passed and 1 failed run
	for i := 0; i < 3; i++ {
		_, _ = svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
			PlanID: plan.ID, CaseID: testCase.ID, Status: model.TestRunStatusPassed,
		})
	}
	_, _ = svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
		PlanID: plan.ID, CaseID: testCase.ID, Status: model.TestRunStatusFailed, ErrorMessage: "Error",
	})

	analytics, err := svc.GetTestingAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get testing analytics: %v", err)
	}

	if analytics.TotalPlans != 1 {
		t.Fatalf("expected 1 plan, got %d", analytics.TotalPlans)
	}
	if analytics.TotalRuns != 4 {
		t.Fatalf("expected 4 total runs, got %d", analytics.TotalRuns)
	}
	if analytics.PassedRuns != 3 || analytics.FailedRuns != 1 {
		t.Fatalf("expected 3 passed, 1 failed, got passed=%d, failed=%d", analytics.PassedRuns, analytics.FailedRuns)
	}
	if analytics.OverallPassRate != 75.0 {
		t.Fatalf("expected 75%% pass rate, got %.2f%%", analytics.OverallPassRate)
	}
	if analytics.QualityHealthIndex <= 0.0 || analytics.QualityHealthIndex > 100.0 {
		t.Fatalf("expected QualityHealthIndex between 0 and 100, got %.2f", analytics.QualityHealthIndex)
	}
	if analytics.TestTypeDistribution[model.TestTypeFunctional] != 1 {
		t.Fatalf("expected 1 functional plan in distribution")
	}
}

func TestTesting_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestTestingSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreatePlan(ctx, "ws_1", model.CreateTestPlanRequest{
		PlanCode: "TP-CONC-1",
		Title:    "Concurrent Test Suite",
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	testCase, err := svc.CreateCase(ctx, "ws_1", model.CreateTestCaseRequest{
		PlanID:         plan.ID,
		CaseCode:       "TC-CONC-1",
		Title:          "Thread Safety Verification",
		TestSteps:      "Execute concurrent runs",
		ExpectedResult: "Consistent DB state without races",
	})
	if err != nil {
		t.Fatalf("failed to create case: %v", err)
	}

	concurrency := 15
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			status := model.TestRunStatusPassed
			if idx%4 == 0 {
				status = model.TestRunStatusFailed
			}
			_, err := svc.ExecuteRun(ctx, "ws_1", model.ExecuteTestRunRequest{
				PlanID:          plan.ID,
				CaseID:          testCase.ID,
				ExecutedBy:      fmt.Sprintf("worker-%d", idx),
				Status:          status,
				ActualResult:    fmt.Sprintf("Result from worker %d", idx),
				ExecutionTimeMs: int64(100 + idx),
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("concurrent test run execution error: %v", err)
		}
	}

	runs, total, err := svc.ListRuns(ctx, model.TestRunFilter{WorkspaceID: "ws_1"})
	if err != nil {
		t.Fatalf("failed to list runs after concurrency: %v", err)
	}
	if total != concurrency || len(runs) != concurrency {
		t.Fatalf("expected %d runs, got total %d", concurrency, total)
	}
}
