package service

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestContinuitySvc(t *testing.T) (ContinuityService, repository.ContinuityRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-continuity', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewContinuityRepository(sqliteDB)
	svc := NewContinuityService(repo)
	return svc, repo, sqliteDB
}

func TestContinuity_DRPlanLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing plan code
	_, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		Name:         "Core API Failover",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != ErrDRPlanCodeRequired {
		t.Fatalf("expected ErrDRPlanCodeRequired, got %v", err)
	}

	// 2. Validation failure: missing plan name
	_, err = svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-001",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != ErrDRPlanNameRequired {
		t.Fatalf("expected ErrDRPlanNameRequired, got %v", err)
	}

	// 3. Validation failure: invalid tier
	_, err = svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-001",
		Name:         "Core API Failover",
		RecoveryTier: "invalid_tier",
	})
	if err != ErrInvalidRecoveryTier {
		t.Fatalf("expected ErrInvalidRecoveryTier, got %v", err)
	}

	// 4. Successful creation
	plan, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:             "DRP-001",
		Name:                 "Core API Multi-Region Failover",
		Scope:                "Payment & Order Ingestion Services",
		RecoveryTier:         model.RecoveryTier1MissionCritical,
		PrimaryCoordinatorID: "user_1",
		InvocationCriteria:   "Complete primary region outage > 15 minutes",
		DocumentURL:          "https://docs.corp.internal/drp/001",
	})
	if err != nil {
		t.Fatalf("failed to create DR plan: %v", err)
	}
	if plan.ID == "" || plan.Status != model.DRPlanStatusDraft {
		t.Fatalf("unexpected plan attributes: %+v", plan)
	}

	// 5. Conflict error on duplicate plan code
	_, err = svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-001",
		Name:         "Duplicate DRP",
		RecoveryTier: model.RecoveryTier2BusinessCritical,
	})
	if err != ErrDRPlanCodeConflict {
		t.Fatalf("expected ErrDRPlanCodeConflict, got %v", err)
	}

	// 6. Get by ID and Code
	fetched, err := svc.GetDRPlanByID(ctx, plan.ID)
	if err != nil || fetched.PlanCode != "DRP-001" {
		t.Fatalf("GetDRPlanByID failed: %v, fetched: %+v", err, fetched)
	}
	byCode, err := svc.GetDRPlanByCode(ctx, "ws_1", "DRP-001")
	if err != nil || byCode.ID != plan.ID {
		t.Fatalf("GetDRPlanByCode failed: %v, byCode: %+v", err, byCode)
	}

	// 7. Update Plan
	updatedScope := "Global Payment and Order Ingestion Infrastructure"
	updated, err := svc.UpdateDRPlan(ctx, plan.ID, model.UpdateDRPlanRequest{
		Scope: &updatedScope,
	})
	if err != nil || updated.Scope != updatedScope || updated.Version != 2 {
		t.Fatalf("UpdateDRPlan failed: %v, updated: %+v", err, updated)
	}

	// 8. Governance approval and activation
	approved, err := svc.ApproveDRPlan(ctx, plan.ID, "user_2")
	if err != nil || approved.Status != model.DRPlanStatusApproved {
		t.Fatalf("ApproveDRPlan failed: %v, approved: %+v", err, approved)
	}
	activated, err := svc.ActivateDRPlan(ctx, plan.ID)
	if err != nil || activated.Status != model.DRPlanStatusActive {
		t.Fatalf("ActivateDRPlan failed: %v, activated: %+v", err, activated)
	}

	// 9. List plans with filter
	list, total, err := svc.ListDRPlans(ctx, model.DRPlanFilterCriteria{
		WorkspaceID:  "ws_1",
		RecoveryTier: model.RecoveryTier1MissionCritical,
		Status:       model.DRPlanStatusActive,
		Search:       "Multi-Region",
		Limit:        10,
	})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("ListDRPlans failed: %v, total: %d, len: %d", err, total, len(list))
	}

	// 10. Delete Plan
	if err := svc.DeleteDRPlan(ctx, plan.ID); err != nil {
		t.Fatalf("DeleteDRPlan failed: %v", err)
	}
	_, err = svc.GetDRPlanByID(ctx, plan.ID)
	if err != ErrDRPlanNotFound {
		t.Fatalf("expected ErrDRPlanNotFound after deletion, got %v", err)
	}
}

func TestContinuity_RecoveryObjectives(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-002",
		Name:         "Customer Portal DR Plan",
		RecoveryTier: model.RecoveryTier2BusinessCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// 1. Validation failure: missing service name
	_, err = svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:          plan.ID,
		RTOValueMinutes: 60,
		RPOValueMinutes: 15,
	})
	if err != ErrServiceNameRequired {
		t.Fatalf("expected ErrServiceNameRequired, got %v", err)
	}

	// 2. Validation failure: negative RTO
	_, err = svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:          plan.ID,
		ServiceName:     "User Auth Service",
		RTOValueMinutes: -5,
	})
	if err != ErrInvalidRTOValue {
		t.Fatalf("expected ErrInvalidRTOValue, got %v", err)
	}

	// 3. Validation failure: plan not found
	_, err = svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:          "non_existent_plan",
		ServiceName:     "User Auth Service",
		RTOValueMinutes: 60,
		RPOValueMinutes: 15,
	})
	if err != ErrDRPlanNotFound {
		t.Fatalf("expected ErrDRPlanNotFound, got %v", err)
	}

	// 4. Successful creation
	obj, err := svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:             plan.ID,
		CIID:               "CI-0010",
		ServiceName:        "User Authentication Gateway",
		RTOValueMinutes:    30,
		RPOValueMinutes:    5,
		TargetAvailability: 99.95,
		DataLossTolerance:  model.DataLossToleranceZeroLoss,
		Notes:              "Synchronous replication to secondary region",
	})
	if err != nil {
		t.Fatalf("failed to create recovery objective: %v", err)
	}
	if obj.ID == "" || obj.TargetAvailability != 99.95 {
		t.Fatalf("unexpected objective: %+v", obj)
	}

	// 5. Update recovery objective
	newRTO := 20
	updated, err := svc.UpdateRecoveryObjective(ctx, obj.ID, model.UpdateRecoveryObjectiveRequest{
		RTOValueMinutes: &newRTO,
	})
	if err != nil || updated.RTOValueMinutes != 20 {
		t.Fatalf("UpdateRecoveryObjective failed: %v, updated: %+v", err, updated)
	}

	// 6. List by plan and workspace
	planObjs, err := svc.ListRecoveryObjectivesByPlan(ctx, plan.ID)
	if err != nil || len(planObjs) != 1 {
		t.Fatalf("ListRecoveryObjectivesByPlan failed: %v, len: %d", err, len(planObjs))
	}
	wsObjs, err := svc.ListRecoveryObjectives(ctx, "ws_1")
	if err != nil || len(wsObjs) != 1 {
		t.Fatalf("ListRecoveryObjectives failed: %v, len: %d", err, len(wsObjs))
	}

	// 7. Delete recovery objective
	if err := svc.DeleteRecoveryObjective(ctx, obj.ID); err != nil {
		t.Fatalf("DeleteRecoveryObjective failed: %v", err)
	}
	_, err = svc.GetRecoveryObjectiveByID(ctx, obj.ID)
	if err != ErrRecoveryObjectiveNotFound {
		t.Fatalf("expected ErrRecoveryObjectiveNotFound, got %v", err)
	}
}

func TestContinuity_BusinessImpactAnalysis(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-003",
		Name:         "Payment Processing Continuity",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// 1. Validation failure: non-positive MTD
	_, err = svc.CreateBIA(ctx, "ws_1", model.CreateBIARequest{
		PlanID:                 plan.ID,
		ServiceName:            "Test Service",
		MTDMinutes:             0,
		OperationalImpact:      model.OpImpactSevere,
		RegulatoryImpact:       model.RegImpactSanctionsPossible,
		CriticalRecoveryWindow: model.RecoveryWindowImmediate,
	})
	if err != ErrInvalidMTDValue {
		t.Fatalf("expected ErrInvalidMTDValue, got %v", err)
	}

	// 2. Successful BIA creation
	bia, err := svc.CreateBIA(ctx, "ws_1", model.CreateBIARequest{
		PlanID:                   plan.ID,
		CIID:                     "CI-0010",
		ServiceName:              "Checkout & Billing Service",
		MTDMinutes:               120,
		FinancialLossRatePerHour: 50000.0,
		OperationalImpact:        model.OpImpactCatastrophic,
		RegulatoryImpact:         model.RegImpactCriticalBreach,
		CriticalRecoveryWindow:   model.RecoveryWindowImmediate,
		AssessedBy:               "alice",
	})
	if err != nil {
		t.Fatalf("failed to create BIA: %v", err)
	}
	if bia.ID == "" || bia.FinancialLossRatePerHour != 50000.0 {
		t.Fatalf("unexpected BIA: %+v", bia)
	}

	// 3. Update BIA
	newLossRate := 75000.0
	updated, err := svc.UpdateBIA(ctx, bia.ID, model.UpdateBIARequest{
		FinancialLossRatePerHour: &newLossRate,
	})
	if err != nil || updated.FinancialLossRatePerHour != 75000.0 {
		t.Fatalf("UpdateBIA failed: %v, updated: %+v", err, updated)
	}

	// 4. List BIAs
	bias, err := svc.ListBIAsByPlan(ctx, plan.ID)
	if err != nil || len(bias) != 1 {
		t.Fatalf("ListBIAsByPlan failed: %v, len: %d", err, len(bias))
	}
	allBIAs, err := svc.ListBIAs(ctx, "ws_1")
	if err != nil || len(allBIAs) != 1 {
		t.Fatalf("ListBIAs failed: %v, len: %d", err, len(allBIAs))
	}

	// 5. Delete BIA
	if err := svc.DeleteBIA(ctx, bia.ID); err != nil {
		t.Fatalf("DeleteBIA failed: %v", err)
	}
	_, err = svc.GetBIAByID(ctx, bia.ID)
	if err != ErrBIANotFound {
		t.Fatalf("expected ErrBIANotFound, got %v", err)
	}
}

func TestContinuity_RunbooksAndSteps(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-004",
		Name:         "Database Cluster Failover Plan",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// 1. Validation failure: missing title
	_, err = svc.CreateRunbook(ctx, "ws_1", model.CreateRunbookRequest{
		PlanID:      plan.ID,
		RunbookCode: "RB-001",
	})
	if err != ErrRunbookTitleRequired {
		t.Fatalf("expected ErrRunbookTitleRequired, got %v", err)
	}

	// 2. Successful runbook creation
	rb, err := svc.CreateRunbook(ctx, "ws_1", model.CreateRunbookRequest{
		PlanID:                   plan.ID,
		RunbookCode:              "RB-001",
		Title:                    "Primary DB Failover & Replica Promotion",
		Description:              "Step-by-step procedure to promote secondary database replica to primary",
		ExecutionMode:            model.RunbookModeAutomated,
		EstimatedDurationMinutes: 15,
	})
	if err != nil {
		t.Fatalf("failed to create runbook: %v", err)
	}
	if rb.ID == "" || rb.ExecutionMode != model.RunbookModeAutomated {
		t.Fatalf("unexpected runbook: %+v", rb)
	}

	// 3. Validation failure: invalid step role
	_, err = svc.CreateRunbookStep(ctx, rb.ID, model.CreateRunbookStepRequest{
		Name:          "Promote Database Replica",
		ActionType:    model.ActionInfraFailover,
		ExecutionRole: "unrecognized_role",
	})
	if err != ErrInvalidExecutionRole {
		t.Fatalf("expected ErrInvalidExecutionRole, got %v", err)
	}

	// 4. Create sequential runbook steps
	step1, err := svc.CreateRunbookStep(ctx, rb.ID, model.CreateRunbookStepRequest{
		StepOrder:                1,
		Name:                     "Verify Primary Unreachable",
		ActionType:               model.ActionHealthCheck,
		ExecutionRole:            model.RoleIncidentCommander,
		CommandOrScript:          "curl -s https://db-primary.internal/health",
		EstimatedDurationMinutes: 2,
		IsCritical:               true,
		FallbackProcedure:        "Double check DNS propagation",
	})
	if err != nil {
		t.Fatalf("failed to create step 1: %v", err)
	}

	step2, err := svc.CreateRunbookStep(ctx, rb.ID, model.CreateRunbookStepRequest{
		StepOrder:                2,
		Name:                     "Promote Secondary Replica",
		ActionType:               model.ActionInfraFailover,
		ExecutionRole:            model.RoleDatabaseAdmin,
		CommandOrScript:          "./scripts/promote_replica.sh --force",
		EstimatedDurationMinutes: 5,
		IsCritical:               true,
		FallbackProcedure:        "Activate read-only degraded mode",
	})
	if err != nil {
		t.Fatalf("failed to create step 2: %v", err)
	}
	if step1.ID == "" || step2.ID == "" {
		t.Fatalf("steps missing IDs")
	}

	// 5. Get Runbook by ID (confirms steps are included)
	fetchedRB, err := svc.GetRunbookByID(ctx, rb.ID)
	if err != nil || len(fetchedRB.Steps) != 2 {
		t.Fatalf("GetRunbookByID failed: %v, steps count: %d", err, len(fetchedRB.Steps))
	}
	if fetchedRB.Steps[0].Name != "Verify Primary Unreachable" {
		t.Fatalf("unexpected step ordering or name: %+v", fetchedRB.Steps)
	}

	// 6. Update step
	newDuration := 3
	updatedStep, err := svc.UpdateRunbookStep(ctx, step1.ID, model.UpdateRunbookStepRequest{
		EstimatedDurationMinutes: &newDuration,
	})
	if err != nil || updatedStep.EstimatedDurationMinutes != 3 {
		t.Fatalf("UpdateRunbookStep failed: %v, updated: %+v", err, updatedStep)
	}

	// 7. Delete step and verify count
	if err := svc.DeleteRunbookStep(ctx, step2.ID); err != nil {
		t.Fatalf("DeleteRunbookStep failed: %v", err)
	}
	remainingSteps, err := svc.ListRunbookSteps(ctx, rb.ID)
	if err != nil || len(remainingSteps) != 1 {
		t.Fatalf("ListRunbookSteps failed: %v, len: %d", err, len(remainingSteps))
	}

	// 8. Delete runbook
	if err := svc.DeleteRunbook(ctx, rb.ID); err != nil {
		t.Fatalf("DeleteRunbook failed: %v", err)
	}
	_, err = svc.GetRunbookByID(ctx, rb.ID)
	if err != ErrRunbookNotFound {
		t.Fatalf("expected ErrRunbookNotFound, got %v", err)
	}
}

func TestContinuity_DrillsAndFindings(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	plan, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-005",
		Name:         "Annual Datacenter Outage Simulation",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan: %v", err)
	}

	// 1. Validation failure: missing scenario
	_, err = svc.CreateDrill(ctx, "ws_1", model.CreateDrillRequest{
		PlanID:    plan.ID,
		DrillCode: "DRILL-2026-Q1",
		DrillType: model.DrillTypeFullFailover,
	})
	if err != ErrDrillScenarioRequired {
		t.Fatalf("expected ErrDrillScenarioRequired, got %v", err)
	}

	// 2. Schedule Drill
	scheduledDate := time.Now().Add(24 * time.Hour)
	drill, err := svc.CreateDrill(ctx, "ws_1", model.CreateDrillRequest{
		PlanID:        plan.ID,
		DrillCode:     "DRILL-2026-Q1",
		DrillType:     model.DrillTypeFullFailover,
		Scenario:      "Simulated fiber cut severing primary AWS us-east-1 VPC",
		ScheduledDate: scheduledDate,
		ConductedBy:   "alice",
	})
	if err != nil {
		t.Fatalf("failed to create drill: %v", err)
	}
	if drill.ID == "" || drill.Status != model.DrillStatusScheduled {
		t.Fatalf("unexpected drill: %+v", drill)
	}

	// 3. Execute and complete drill
	statusCompleted := model.DrillStatusCompleted
	actualDuration := 25
	successTrue := true
	postMortem := "Failover completed within target RTO. Secondary region picked up 100% of production traffic."
	updatedDrill, err := svc.UpdateDrill(ctx, drill.ID, model.UpdateDrillRequest{
		Status:                        &statusCompleted,
		ActualRecoveryDurationMinutes: &actualDuration,
		Success:                       &successTrue,
		PostMortemSummary:             &postMortem,
	})
	if err != nil || !updatedDrill.Success || updatedDrill.ActualRecoveryDurationMinutes != 25 {
		t.Fatalf("UpdateDrill failed: %v, updated: %+v", err, updatedDrill)
	}

	// 4. Log drill findings
	finding1, err := svc.CreateDrillFinding(ctx, "ws_1", drill.ID, model.CreateDrillFindingRequest{
		Severity:          model.DrillFindingSeverityHigh,
		Category:          model.DrillFindingCatToolingFailure,
		Description:       "Automated DNS TTL propagation took 8 minutes longer than estimated",
		RemediationAction: "Reduce Route53 record TTL from 300s to 60s",
		AssignedTo:        "bob",
	})
	if err != nil {
		t.Fatalf("failed to create finding 1: %v", err)
	}

	finding2, err := svc.CreateDrillFinding(ctx, "ws_1", drill.ID, model.CreateDrillFindingRequest{
		Severity:          model.DrillFindingSeverityLow,
		Category:          model.DrillFindingCatDocOutdated,
		Description:       "Runbook link in incident dashboard pointed to archived wiki page",
		RemediationAction: "Update confluence link in monitoring runbook",
		AssignedTo:        "alice",
	})
	if err != nil {
		t.Fatalf("failed to create finding 2: %v", err)
	}

	// 5. Get Drill by ID (confirms findings are loaded)
	fetchedDrill, err := svc.GetDrillByID(ctx, drill.ID)
	if err != nil || len(fetchedDrill.Findings) != 2 {
		t.Fatalf("GetDrillByID failed: %v, findings len: %d", err, len(fetchedDrill.Findings))
	}

	// 6. Resolve finding 1
	statusResolved := model.DrillFindingStatusResolved
	resolvedFinding, err := svc.UpdateDrillFinding(ctx, finding1.ID, model.UpdateDrillFindingRequest{
		Status: &statusResolved,
	})
	if err != nil || resolvedFinding.Status != model.DrillFindingStatusResolved || resolvedFinding.ResolvedAt == nil {
		t.Fatalf("UpdateDrillFinding failed: %v, resolved: %+v", err, resolvedFinding)
	}

	// 7. List findings with filter
	openFindings, totalOpen, err := svc.ListDrillFindings(ctx, model.DrillFindingFilterCriteria{
		WorkspaceID: "ws_1",
		DrillID:     drill.ID,
		Status:      model.DrillFindingStatusOpen,
	})
	if err != nil || totalOpen != 1 || len(openFindings) != 1 {
		t.Fatalf("ListDrillFindings failed: %v, total: %d, len: %d", err, totalOpen, len(openFindings))
	}

	// 8. Delete finding 2
	if err := svc.DeleteDrillFinding(ctx, finding2.ID); err != nil {
		t.Fatalf("DeleteDrillFinding failed: %v", err)
	}
	_, err = svc.GetDrillFindingByID(ctx, finding2.ID)
	if err != ErrDrillFindingNotFound {
		t.Fatalf("expected ErrDrillFindingNotFound, got %v", err)
	}
}

func TestContinuity_ReadinessAnalytics(t *testing.T) {
	svc, _, dbConn := setupTestContinuitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create 2 DR plans (one mission-critical active, one business-critical approved)
	plan1, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-AN-1",
		Name:         "Mission Critical Core Cloud DR",
		RecoveryTier: model.RecoveryTier1MissionCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan 1: %v", err)
	}
	_, err = svc.ApproveDRPlan(ctx, plan1.ID, "user_1")
	if err != nil {
		t.Fatalf("failed to approve plan 1: %v", err)
	}
	_, err = svc.ActivateDRPlan(ctx, plan1.ID)
	if err != nil {
		t.Fatalf("failed to activate plan 1: %v", err)
	}

	plan2, err := svc.CreateDRPlan(ctx, "ws_1", model.CreateDRPlanRequest{
		PlanCode:     "DRP-AN-2",
		Name:         "Analytics & Reporting Cluster DR",
		RecoveryTier: model.RecoveryTier2BusinessCritical,
	})
	if err != nil {
		t.Fatalf("failed to create plan 2: %v", err)
	}
	_, err = svc.ApproveDRPlan(ctx, plan2.ID, "user_1")
	if err != nil {
		t.Fatalf("failed to approve plan 2: %v", err)
	}

	// Recovery objectives
	_, err = svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:             plan1.ID,
		ServiceName:        "Payment Gateway",
		RTOValueMinutes:    30,
		RPOValueMinutes:    5,
		TargetAvailability: 99.99,
	})
	if err != nil {
		t.Fatalf("failed to create obj 1: %v", err)
	}
	_, err = svc.CreateRecoveryObjective(ctx, "ws_1", model.CreateRecoveryObjectiveRequest{
		PlanID:             plan2.ID,
		ServiceName:        "Warehouse Reporting",
		RTOValueMinutes:    120,
		RPOValueMinutes:    30,
		TargetAvailability: 99.5,
	})
	if err != nil {
		t.Fatalf("failed to create obj 2: %v", err)
	}

	// Drills (1 completed & successful, 1 scheduled)
	drill1, err := svc.CreateDrill(ctx, "ws_1", model.CreateDrillRequest{
		PlanID:        plan1.ID,
		DrillCode:     "DRILL-AN-1",
		DrillType:     model.DrillTypeFullFailover,
		Scenario:      "Regional power outage simulation",
		ScheduledDate: time.Now().Add(-48 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create drill 1: %v", err)
	}
	statusDone := model.DrillStatusCompleted
	dur25 := 25
	successTrue := true
	_, err = svc.UpdateDrill(ctx, drill1.ID, model.UpdateDrillRequest{
		Status:                        &statusDone,
		ActualRecoveryDurationMinutes: &dur25,
		Success:                       &successTrue,
	})
	if err != nil {
		t.Fatalf("failed to complete drill 1: %v", err)
	}

	_, err = svc.CreateDrill(ctx, "ws_1", model.CreateDrillRequest{
		PlanID:        plan2.ID,
		DrillCode:     "DRILL-AN-2",
		DrillType:     model.DrillTypeTabletop,
		Scenario:      "Storage corruption recovery drill",
		ScheduledDate: time.Now().Add(72 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to schedule drill 2: %v", err)
	}

	// Drill finding (resolved)
	finding, err := svc.CreateDrillFinding(ctx, "ws_1", drill1.ID, model.CreateDrillFindingRequest{
		Severity:    model.DrillFindingSeverityMedium,
		Category:    model.DrillFindingCatProcessGap,
		Description: "Call roster had outdated cell number for DBA lead",
	})
	if err != nil {
		t.Fatalf("failed to create finding: %v", err)
	}
	stResolved := model.DrillFindingStatusResolved
	_, err = svc.UpdateDrillFinding(ctx, finding.ID, model.UpdateDrillFindingRequest{
		Status: &stResolved,
	})
	if err != nil {
		t.Fatalf("failed to resolve finding: %v", err)
	}

	// Generate Continuity Readiness Report
	report, err := svc.GetContinuityReadinessReport(ctx, "ws_1")
	if err != nil {
		t.Fatalf("GetContinuityReadinessReport failed: %v", err)
	}

	if report.TotalPlans != 2 || report.ActivePlans != 1 || report.ApprovedPlans != 1 {
		t.Fatalf("unexpected plan counts: %+v", report)
	}
	if report.TotalObjectives != 2 || report.AverageRTOMinutes != 75.0 || report.AverageRPOMinutes != 17.5 {
		t.Fatalf("unexpected objective metrics: %+v", report)
	}
	if report.TotalDrills != 2 || report.CompletedDrills != 1 || report.DrillPassRate != 100.0 {
		t.Fatalf("unexpected drill metrics: %+v", report)
	}
	if report.TotalFindings != 1 || report.OpenFindingsCount != 0 {
		t.Fatalf("unexpected finding counts: %+v", report)
	}
	if report.ReadinessScore <= 50.0 {
		t.Fatalf("expected readiness score to be boosted (> 50), got %f", report.ReadinessScore)
	}
}
