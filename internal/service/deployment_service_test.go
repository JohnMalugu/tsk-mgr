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

func setupTestDeploymentSvc(t *testing.T) (DeploymentService, repository.DeploymentRepository, *sql.DB) {
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
		VALUES (1, 'Deployment WS', 'deploy-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewDeploymentRepository(sqliteDB)
	svc := NewDeploymentService(repo)
	return svc, repo, sqliteDB
}

func TestDeployment_PipelineLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing pipeline_code
	_, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		Name:               "Production Payments Pipeline",
		TargetEnvironment:  model.DeploymentEnvProd,
		DeploymentStrategy: model.DeploymentStrategyCanary,
	})
	if err == nil {
		t.Fatal("expected error for missing pipeline_code, got nil")
	}

	// 2. Validation failure: invalid environment
	_, err = svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:       "PIPE-001",
		Name:               "Invalid Env Pipeline",
		TargetEnvironment:  "invalid_env",
		DeploymentStrategy: model.DeploymentStrategyRolling,
	})
	if err == nil {
		t.Fatal("expected error for invalid environment, got nil")
	}

	// 3. Create pipeline successfully
	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-PAY-001",
		Name:                "Payments Service Pipeline",
		Description:         "Continuous delivery pipeline for core payment processing",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyCanary,
		AutoRollbackEnabled: true,
		LinkedCIID:          "ci-srv-payments",
		CreatedByUserID:     "user_1",
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}
	if pipe.ID == "" || pipe.Status != model.PipelineStatusActive {
		t.Fatalf("unexpected pipeline created state: %+v", pipe)
	}

	// 4. Unique pipeline_code constraint
	_, err = svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-PAY-001",
		Name:                "Duplicate Code Pipeline",
		TargetEnvironment:   model.DeploymentEnvStaging,
		DeploymentStrategy:  model.DeploymentStrategyBlueGreen,
		AutoRollbackEnabled: false,
	})
	if err != ErrDeploymentPipelineCodeConflict {
		t.Fatalf("expected ErrDeploymentPipelineCodeConflict, got %v", err)
	}

	// 5. Get pipeline by ID and by code
	fetched, err := svc.GetPipelineByID(ctx, pipe.ID)
	if err != nil {
		t.Fatalf("failed to get pipeline by ID: %v", err)
	}
	if fetched.Name != "Payments Service Pipeline" {
		t.Fatalf("unexpected pipeline name: %s", fetched.Name)
	}

	byCode, err := svc.GetPipelineByCode(ctx, "ws_1", "PIPE-PAY-001")
	if err != nil {
		t.Fatalf("failed to get pipeline by code: %v", err)
	}
	if byCode.ID != pipe.ID {
		t.Fatalf("pipeline ID mismatch: %s vs %s", byCode.ID, pipe.ID)
	}

	// 6. Update pipeline
	newDesc := "Updated pipeline description"
	newStrategy := model.DeploymentStrategyBlueGreen
	updated, err := svc.UpdatePipeline(ctx, pipe.ID, model.UpdateDeploymentPipelineRequest{
		Description:        &newDesc,
		DeploymentStrategy: &newStrategy,
	})
	if err != nil {
		t.Fatalf("failed to update pipeline: %v", err)
	}
	if updated.Description != newDesc || updated.DeploymentStrategy != newStrategy {
		t.Fatalf("unexpected updated fields: %+v", updated)
	}

	// 7. List pipelines with filters
	list, total, err := svc.ListPipelines(ctx, model.DeploymentPipelineFilter{
		WorkspaceID:        "ws_1",
		DeploymentStrategy: model.DeploymentStrategyBlueGreen,
	})
	if err != nil {
		t.Fatalf("failed to list pipelines: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 pipeline, got %d (total %d)", len(list), total)
	}

	// 8. Delete pipeline
	if err := svc.DeletePipeline(ctx, pipe.ID); err != nil {
		t.Fatalf("failed to delete pipeline: %v", err)
	}
	_, err = svc.GetPipelineByID(ctx, pipe.ID)
	if err != ErrDeploymentPipelineNotFound {
		t.Fatalf("expected ErrDeploymentPipelineNotFound, got %v", err)
	}
}

func TestDeployment_RunLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-ORDERS-01",
		Name:                "Orders Service Pipeline",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyRolling,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// 1. Validation failure: missing release version
	_, err = svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:  pipe.ID,
		Environment: model.DeploymentEnvProd,
		TriggerType: model.RunTriggerAutomatedCI,
	})
	if err == nil {
		t.Fatal("expected error for missing release_version, got nil")
	}

	// 2. Create Run 1
	run1, err := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:       pipe.ID,
		ReleaseVersion:   "v1.0.0",
		Environment:      model.DeploymentEnvProd,
		TriggerType:      model.RunTriggerAutomatedCI,
		DeployedByUserID: "user_1",
		CommitHash:       "a1b2c3d4",
		Changelog:        "Initial production release",
	})
	if err != nil {
		t.Fatalf("failed to create run 1: %v", err)
	}
	if run1.RunNumber != 1 || run1.Status != model.RunStatusPending {
		t.Fatalf("unexpected run1 state: %+v", run1)
	}

	// 3. Create Run 2 (auto-increments run_number)
	run2, err := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:       pipe.ID,
		ReleaseVersion:   "v1.0.1",
		Environment:      model.DeploymentEnvProd,
		TriggerType:      model.RunTriggerManual,
		DeployedByUserID: "user_2",
	})
	if err != nil {
		t.Fatalf("failed to create run 2: %v", err)
	}
	if run2.RunNumber != 2 {
		t.Fatalf("expected run number 2, got %d", run2.RunNumber)
	}

	// 4. Update Run 1 to successful
	run1Up, err := svc.UpdateRunStatus(ctx, run1.ID, model.UpdateDeploymentRunStatusRequest{
		Status:          model.RunStatusSuccessful,
		DurationSeconds: 120.5,
	})
	if err != nil {
		t.Fatalf("failed to update run1 status: %v", err)
	}
	if run1Up.Status != model.RunStatusSuccessful || run1Up.FinishedAt == nil || run1Up.DurationSeconds != 120.5 {
		t.Fatalf("unexpected run1 updated state: %+v", run1Up)
	}

	// 5. Update Run 2 to rolled_back
	run2Up, err := svc.UpdateRunStatus(ctx, run2.ID, model.UpdateDeploymentRunStatusRequest{
		Status:         model.RunStatusRolledBack,
		RollbackReason: "High latency detected on checkout endpoint",
	})
	if err != nil {
		t.Fatalf("failed to update run2 status: %v", err)
	}
	if run2Up.Status != model.RunStatusRolledBack || run2Up.RollbackReason != "High latency detected on checkout endpoint" {
		t.Fatalf("unexpected run2 updated state: %+v", run2Up)
	}

	// 6. List runs
	runs, total, err := svc.ListRuns(ctx, model.DeploymentRunFilter{
		WorkspaceID: "ws_1",
		PipelineID:  pipe.ID,
	})
	if err != nil {
		t.Fatalf("failed to list runs: %v", err)
	}
	if total != 2 || len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d (total %d)", len(runs), total)
	}
}

func TestDeployment_GatesAndApproval(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-AUTH-01",
		Name:                "Auth Service Pipeline",
		TargetEnvironment:   model.DeploymentEnvStaging,
		DeploymentStrategy:  model.DeploymentStrategyBlueGreen,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	run, err := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:     pipe.ID,
		ReleaseVersion: "v2.0.0",
		Environment:    model.DeploymentEnvStaging,
		TriggerType:    model.RunTriggerAutomatedCI,
	})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// 1. Validation failure: missing gate_name
	_, err = svc.CreateGate(ctx, "ws_1", model.CreateDeploymentGateRequest{
		RunID:    run.ID,
		GateType: model.GateTypeAutomatedTest,
	})
	if err == nil {
		t.Fatal("expected error for missing gate_name, got nil")
	}

	// 2. Create Gate 1: automated test
	gate1, err := svc.CreateGate(ctx, "ws_1", model.CreateDeploymentGateRequest{
		RunID:    run.ID,
		GateType: model.GateTypeAutomatedTest,
		GateName: "Integration Test Suite",
	})
	if err != nil {
		t.Fatalf("failed to create gate 1: %v", err)
	}
	if gate1.Status != model.GateStatusPending {
		t.Fatalf("expected pending status, got %s", gate1.Status)
	}

	// 3. Create Gate 2: manual approval
	gate2, err := svc.CreateGate(ctx, "ws_1", model.CreateDeploymentGateRequest{
		RunID:    run.ID,
		GateType: model.GateTypeManualApproval,
		GateName: "CAB Release Approval",
	})
	if err != nil {
		t.Fatalf("failed to create gate 2: %v", err)
	}

	// 4. Evaluate Gate 1: Passed
	gate1Eval, err := svc.EvaluateGate(ctx, gate1.ID, model.EvaluateDeploymentGateRequest{
		Status:          model.GateStatusPassed,
		EvaluatorID:     "ci_worker_1",
		EvidenceDetails: "148 tests executed, 0 failures, 100% pass rate",
	})
	if err != nil {
		t.Fatalf("failed to evaluate gate 1: %v", err)
	}
	if gate1Eval.Status != model.GateStatusPassed || gate1Eval.EvaluatedAt == nil {
		t.Fatalf("unexpected gate1 evaluated state: %+v", gate1Eval)
	}

	// 5. Evaluate Gate 2: Waived with reason
	gate2Eval, err := svc.EvaluateGate(ctx, gate2.ID, model.EvaluateDeploymentGateRequest{
		Status:       model.GateStatusWaived,
		EvaluatorID:  "user_1",
		WaiverReason: "Emergency hotfix pre-approved under Incident INC-9001",
	})
	if err != nil {
		t.Fatalf("failed to evaluate gate 2: %v", err)
	}
	if gate2Eval.Status != model.GateStatusWaived || gate2Eval.WaiverReason == "" {
		t.Fatalf("unexpected gate2 evaluated state: %+v", gate2Eval)
	}

	// 6. List gates
	gates, total, err := svc.ListGates(ctx, model.DeploymentGateFilter{
		WorkspaceID: "ws_1",
		RunID:       run.ID,
	})
	if err != nil {
		t.Fatalf("failed to list gates: %v", err)
	}
	if total != 2 || len(gates) != 2 {
		t.Fatalf("expected 2 gates, got %d (total %d)", len(gates), total)
	}
}

func TestDeployment_CanarySteps(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-CANARY-01",
		Name:                "Canary Pipeline",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyCanary,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	run, err := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:     pipe.ID,
		ReleaseVersion: "v3.1.0",
		Environment:    model.DeploymentEnvProd,
		TriggerType:    model.RunTriggerAutomatedCI,
	})
	if err != nil {
		t.Fatalf("failed to create run: %v", err)
	}

	// 1. Validation failure: invalid traffic percentage
	_, err = svc.CreateStep(ctx, "ws_1", model.CreateDeploymentStepRequest{
		RunID:             run.ID,
		StepOrder:         1,
		StepName:          "Invalid Step",
		TrafficPercentage: 150.0,
	})
	if err == nil {
		t.Fatal("expected error for traffic_percentage > 100, got nil")
	}

	// 2. Create Step 1: 10% canary traffic
	step1, err := svc.CreateStep(ctx, "ws_1", model.CreateDeploymentStepRequest{
		RunID:              run.ID,
		StepOrder:          1,
		StepName:           "Canary 10%",
		TrafficPercentage:  10.0,
		SoakTimeMinutes:    15,
		ErrorRateThreshold: 1.0,
	})
	if err != nil {
		t.Fatalf("failed to create step 1: %v", err)
	}
	if step1.Status != model.StepStatusPending {
		t.Fatalf("expected pending status, got %s", step1.Status)
	}

	// 3. Promote Step 1 to soaking then promoted
	soakingStep, err := svc.PromoteStep(ctx, step1.ID, model.PromoteDeploymentStepRequest{
		Status:          model.StepStatusSoaking,
		ActualErrorRate: 0.12,
	})
	if err != nil {
		t.Fatalf("failed to transition step to soaking: %v", err)
	}
	if soakingStep.StartedAt == nil {
		t.Fatal("expected non-nil started_at for soaking step")
	}

	promotedStep, err := svc.PromoteStep(ctx, step1.ID, model.PromoteDeploymentStepRequest{
		Status:          model.StepStatusPromoted,
		ActualErrorRate: 0.15,
	})
	if err != nil {
		t.Fatalf("failed to promote step: %v", err)
	}
	if promotedStep.CompletedAt == nil || promotedStep.Status != model.StepStatusPromoted {
		t.Fatalf("unexpected promoted step state: %+v", promotedStep)
	}

	// 4. List steps
	steps, total, err := svc.ListSteps(ctx, model.DeploymentStepFilter{
		WorkspaceID: "ws_1",
		RunID:       run.ID,
	})
	if err != nil {
		t.Fatalf("failed to list steps: %v", err)
	}
	if total != 1 || len(steps) != 1 {
		t.Fatalf("expected 1 step, got %d (total %d)", len(steps), total)
	}
}

func TestDeployment_RollbackPolicy(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-RB-01",
		Name:                "Rollback Test Pipeline",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyCanary,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// 1. Validation failure: missing target_rollback_version
	_, err = svc.CreateRollbackPolicy(ctx, "ws_1", model.CreateRollbackPolicyRequest{
		PipelineID:       pipe.ID,
		TriggerCondition: model.RollbackTriggerErrorRateSpike,
	})
	if err == nil {
		t.Fatal("expected error for missing target_rollback_version, got nil")
	}

	// 2. Create Rollback Policy
	policy, err := svc.CreateRollbackPolicy(ctx, "ws_1", model.CreateRollbackPolicyRequest{
		PipelineID:            pipe.ID,
		TriggerCondition:      model.RollbackTriggerErrorRateSpike,
		ThresholdValue:        2.5,
		TargetRollbackVersion: "v1.9.9",
	})
	if err != nil {
		t.Fatalf("failed to create rollback policy: %v", err)
	}
	if policy.ExecutionStatus != model.RollbackStatusTriggered || policy.TriggeredAt == nil {
		t.Fatalf("unexpected policy state: %+v", policy)
	}

	// 3. Execute Rollback to Completed
	executed, err := svc.ExecuteRollback(ctx, policy.ID, model.ExecuteRollbackRequest{
		ExecutionStatus: model.RollbackStatusCompleted,
		RollbackLogs:    "Automated traffic rerouted to v1.9.9 replica set successfully",
	})
	if err != nil {
		t.Fatalf("failed to execute rollback: %v", err)
	}
	if executed.ExecutionStatus != model.RollbackStatusCompleted || executed.CompletedAt == nil {
		t.Fatalf("unexpected completed rollback state: %+v", executed)
	}

	// 4. List Rollbacks
	policies, total, err := svc.ListRollbackPolicies(ctx, model.RollbackPolicyFilter{
		WorkspaceID: "ws_1",
		PipelineID:  pipe.ID,
	})
	if err != nil {
		t.Fatalf("failed to list policies: %v", err)
	}
	if total != 1 || len(policies) != 1 {
		t.Fatalf("expected 1 policy, got %d (total %d)", len(policies), total)
	}
}

func TestDeployment_AnalyticsAndDORAMetrics(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-DORA-01",
		Name:                "DORA Analytics Pipeline",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyRolling,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create pipeline: %v", err)
	}

	// Create 3 runs: 2 successful, 1 rolled back
	r1, _ := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:     pipe.ID,
		ReleaseVersion: "v1.0.0",
		Environment:    model.DeploymentEnvProd,
		TriggerType:    model.RunTriggerAutomatedCI,
	})
	_, _ = svc.UpdateRunStatus(ctx, r1.ID, model.UpdateDeploymentRunStatusRequest{
		Status:          model.RunStatusSuccessful,
		DurationSeconds: 1800, // 0.5 hours
	})

	r2, _ := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:     pipe.ID,
		ReleaseVersion: "v1.1.0",
		Environment:    model.DeploymentEnvProd,
		TriggerType:    model.RunTriggerAutomatedCI,
	})
	_, _ = svc.UpdateRunStatus(ctx, r2.ID, model.UpdateDeploymentRunStatusRequest{
		Status:          model.RunStatusSuccessful,
		DurationSeconds: 3600, // 1.0 hour
	})

	r3, _ := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
		PipelineID:     pipe.ID,
		ReleaseVersion: "v1.2.0",
		Environment:    model.DeploymentEnvStaging,
		TriggerType:    model.RunTriggerManual,
	})
	_, _ = svc.UpdateRunStatus(ctx, r3.ID, model.UpdateDeploymentRunStatusRequest{
		Status:          model.RunStatusRolledBack,
		DurationSeconds: 600, // 10 minutes recovery
		RollbackReason:  "Unit test regression in payment connector",
	})

	analytics, err := svc.GetDeploymentAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get deployment analytics: %v", err)
	}

	if analytics.TotalPipelines != 1 || analytics.ActivePipelines != 1 {
		t.Fatalf("unexpected pipeline metrics: %+v", analytics)
	}
	if analytics.TotalRuns != 3 || analytics.SuccessfulRuns != 2 || analytics.RolledBackRuns != 1 {
		t.Fatalf("unexpected run counts: %+v", analytics)
	}
	// Change failure rate should be 1/3 * 100 = 33.33%
	if analytics.ChangeFailureRate < 33.0 || analytics.ChangeFailureRate > 34.0 {
		t.Fatalf("unexpected change failure rate: %f", analytics.ChangeFailureRate)
	}
	// Average lead time: (1800 + 3600)/2 / 3600 = 0.75 hours
	if analytics.LeadTimeForChangesHours < 0.7 || analytics.LeadTimeForChangesHours > 0.8 {
		t.Fatalf("unexpected lead time hours: %f", analytics.LeadTimeForChangesHours)
	}
	// MTTR: 600 / 60 = 10.0 minutes
	if analytics.MTTRMinutes < 9.9 || analytics.MTTRMinutes > 10.1 {
		t.Fatalf("unexpected MTTR: %f", analytics.MTTRMinutes)
	}
	if analytics.DeploymentHealthIndex <= 0.0 || analytics.DeploymentHealthIndex > 100.0 {
		t.Fatalf("unexpected health index: %f", analytics.DeploymentHealthIndex)
	}
	if analytics.EnvironmentDistribution[model.DeploymentEnvProd] != 2 || analytics.EnvironmentDistribution[model.DeploymentEnvStaging] != 1 {
		t.Fatalf("unexpected env distribution: %+v", analytics.EnvironmentDistribution)
	}
}

func TestDeployment_ConcurrentRunsAndPipelines(t *testing.T) {
	svc, _, dbConn := setupTestDeploymentSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	const concurrency = 10
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency*2)

	// Create a shared pipeline
	pipe, err := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
		PipelineCode:        "PIPE-CONC-01",
		Name:                "Concurrent Delivery Pipeline",
		TargetEnvironment:   model.DeploymentEnvProd,
		DeploymentStrategy:  model.DeploymentStrategyCanary,
		AutoRollbackEnabled: true,
	})
	if err != nil {
		t.Fatalf("failed to create base pipeline: %v", err)
	}

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Create pipeline
			_, pErr := svc.CreatePipeline(ctx, "ws_1", model.CreateDeploymentPipelineRequest{
				PipelineCode:        fmt.Sprintf("PIPE-PARALLEL-%02d", idx),
				Name:                fmt.Sprintf("Parallel Pipeline %d", idx),
				TargetEnvironment:   model.DeploymentEnvDev,
				DeploymentStrategy:  model.DeploymentStrategyRolling,
				AutoRollbackEnabled: true,
			})
			if pErr != nil {
				errCh <- pErr
				return
			}

			// Create run on shared pipeline
			run, rErr := svc.CreateRun(ctx, "ws_1", model.CreateDeploymentRunRequest{
				PipelineID:     pipe.ID,
				ReleaseVersion: fmt.Sprintf("v4.%d.0", idx),
				Environment:    model.DeploymentEnvProd,
				TriggerType:    model.RunTriggerAutomatedCI,
			})
			if rErr != nil {
				errCh <- rErr
				return
			}

			// Update run
			_, uErr := svc.UpdateRunStatus(ctx, run.ID, model.UpdateDeploymentRunStatusRequest{
				Status:          model.RunStatusSuccessful,
				DurationSeconds: float64(idx * 10),
			})
			if uErr != nil {
				errCh <- uErr
				return
			}
		}(i)
	}

	wg.Wait()
	close(errCh)

	for cErr := range errCh {
		if cErr != nil {
			t.Fatalf("concurrency error encountered: %v", cErr)
		}
	}

	analytics, err := svc.GetDeploymentAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get analytics after concurrent runs: %v", err)
	}
	if analytics.TotalRuns < concurrency {
		t.Fatalf("expected at least %d runs, got %d", concurrency, analytics.TotalRuns)
	}
}
