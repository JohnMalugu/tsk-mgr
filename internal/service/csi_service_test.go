package service

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestCSISvc(t *testing.T) (CSIService, repository.CSIRepository, *sql.DB) {
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
		VALUES (1, 'CSI WS', 'csi-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewCSIRepository(sqliteDB)
	svc := NewCSIService(repo)
	return svc, repo, sqliteDB
}

func TestCSI_InitiativeLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing title
	_, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		StrategicAlignment: model.CSIAlignmentReliability,
	})
	if err != ErrCSIInitiativeTitleRequired {
		t.Fatalf("expected ErrCSIInitiativeTitleRequired, got %v", err)
	}

	// 2. Validation failure: invalid strategic alignment
	_, err = svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Invalid Alignment",
		StrategicAlignment: "invalid_category",
	})
	if err != model.ErrInvalidCSIAlignment {
		t.Fatalf("expected ErrInvalidCSIAlignment, got %v", err)
	}

	// 3. Create Initiative
	init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		InitiativeCode:       "CSI-2026-01",
		Title:                "Zero-Downtime Deployment Migration",
		Description:          "Adopt blue-green deployment pipelines to eliminate release downtime",
		StrategicAlignment:   model.CSIAlignmentReliability,
		Priority:             model.CSIPriorityHigh,
		Status:               model.CSIStatusApproved,
		Source:               model.CSISourceIncidentRCA,
		SourceReferenceID:    "inc-501",
		CIID:                 "ci-k8s-cluster",
		OwnerID:              "user_1",
		TargetCompletionDate: time.Now().UTC().AddDate(0, 2, 0),
	})
	if err != nil {
		t.Fatalf("failed to create initiative: %v", err)
	}
	if init.ID == "" || init.InitiativeCode != "CSI-2026-01" {
		t.Fatalf("unexpected initiative data: %+v", init)
	}

	// 4. Code conflict
	_, err = svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		InitiativeCode:     "CSI-2026-01",
		Title:              "Duplicate Code Initiative",
		StrategicAlignment: model.CSIAlignmentCostReduction,
	})
	if err != ErrCSIInitiativeCodeConflict {
		t.Fatalf("expected ErrCSIInitiativeCodeConflict, got %v", err)
	}

	// 5. Get by ID & Get by Code
	fetched, err := svc.GetInitiativeByID(ctx, init.ID)
	if err != nil {
		t.Fatalf("failed to get initiative by ID: %v", err)
	}
	if fetched.Title != init.Title {
		t.Fatalf("expected title %s, got %s", init.Title, fetched.Title)
	}

	byCode, err := svc.GetInitiativeByCode(ctx, "ws_1", "CSI-2026-01")
	if err != nil || byCode.ID != init.ID {
		t.Fatalf("failed to get initiative by code: %v", err)
	}

	// 6. Update Initiative
	newPriority := model.CSIPriorityUrgent
	updated, err := svc.UpdateInitiative(ctx, init.ID, model.UpdateCSIInitiativeRequest{
		Priority: &newPriority,
	})
	if err != nil {
		t.Fatalf("failed to update initiative: %v", err)
	}
	if updated.Priority != model.CSIPriorityUrgent {
		t.Fatalf("expected urgent priority, got %s", updated.Priority)
	}

	// 7. Complete Initiative
	completed, err := svc.CompleteInitiative(ctx, init.ID)
	if err != nil {
		t.Fatalf("failed to complete initiative: %v", err)
	}
	if completed.Status != model.CSIStatusCompleted || completed.CompletedAt == nil {
		t.Fatalf("initiative not completed properly: %+v", completed)
	}

	// 8. List Initiatives
	list, total, err := svc.ListInitiatives(ctx, model.CSIFilterCriteria{
		WorkspaceID: "ws_1",
		Status:      model.CSIStatusCompleted,
	})
	if err != nil {
		t.Fatalf("failed to list initiatives: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 completed initiative, got %d", total)
	}

	// 9. Delete Initiative
	if err := svc.DeleteInitiative(ctx, init.ID); err != nil {
		t.Fatalf("failed to delete initiative: %v", err)
	}
	_, err = svc.GetInitiativeByID(ctx, init.ID)
	if err != ErrCSIInitiativeNotFound {
		t.Fatalf("expected ErrCSIInitiativeNotFound after delete, got %v", err)
	}
}

func TestCSI_BenchmarkLifecycleAndTargetAchievement(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Database Latency Optimization",
		StrategicAlignment: model.CSIAlignmentReliability,
	})
	if err != nil {
		t.Fatalf("failed to create initiative: %v", err)
	}

	// 1. Add latency benchmark (lower is better: baseline 250ms, target 50ms)
	b1, err := svc.AddBenchmark(ctx, "ws_1", init.ID, model.CreateCSIBenchmarkRequest{
		MetricName:    "P99 API Latency",
		UnitOfMeasure: model.CSIMetricUnitMilliseconds,
		BaselineValue: 250.0,
		CurrentValue:  180.0,
		TargetValue:   50.0,
		Notes:         "Initial caching applied",
	})
	if err != nil {
		t.Fatalf("failed to add benchmark: %v", err)
	}

	// Has not achieved target yet (180 > 50)
	if b1.HasAchievedTarget() {
		t.Fatalf("benchmark should not be achieved yet (180ms > 50ms)")
	}

	// 2. Verify benchmark with new measured value (45ms, achieved!)
	verified, err := svc.VerifyBenchmark(ctx, b1.ID, 45.0, "Redis cluster deployed and warm")
	if err != nil {
		t.Fatalf("failed to verify benchmark: %v", err)
	}
	if verified.VerifiedAt == nil || verified.CurrentValue != 45.0 {
		t.Fatalf("verification timestamp or value missing")
	}
	if !verified.HasAchievedTarget() {
		t.Fatalf("benchmark should be achieved (45ms <= 50ms)")
	}

	// 3. Add availability benchmark (higher is better: baseline 99.5%, target 99.99%)
	b2, err := svc.AddBenchmark(ctx, "ws_1", init.ID, model.CreateCSIBenchmarkRequest{
		MetricName:    "Service Availability",
		UnitOfMeasure: model.CSIMetricUnitPercent,
		BaselineValue: 99.5,
		CurrentValue:  99.995,
		TargetValue:   99.99,
	})
	if err != nil {
		t.Fatalf("failed to add availability benchmark: %v", err)
	}
	if !b2.HasAchievedTarget() {
		t.Fatalf("availability benchmark should be achieved (99.995%% >= 99.99%%)")
	}

	// 4. List Benchmarks
	benchmarks, err := svc.ListBenchmarks(ctx, init.ID)
	if err != nil || len(benchmarks) != 2 {
		t.Fatalf("expected 2 benchmarks, got %d", len(benchmarks))
	}

	// 5. Delete Benchmark
	if err := svc.DeleteBenchmark(ctx, b1.ID); err != nil {
		t.Fatalf("failed to delete benchmark: %v", err)
	}
	_, err = svc.GetBenchmarkByID(ctx, b1.ID)
	if err != ErrCSIBenchmarkNotFound {
		t.Fatalf("expected ErrCSIBenchmarkNotFound, got %v", err)
	}
}

func TestCSI_BenefitRealizationAndROI(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Cloud Resource Right-Sizing",
		StrategicAlignment: model.CSIAlignmentCostReduction,
	})
	if err != nil {
		t.Fatalf("failed to create initiative: %v", err)
	}

	// 1. Record benefit projection
	benefit, err := svc.RecordBenefit(ctx, "ws_1", init.ID, model.RecordCSIBenefitRequest{
		ProjectedFinancialSavings: 45000.0,
		RealizedFinancialSavings:  15000.0,
		EffortHoursInvested:       80.0,
		HoursSavedPerMonth:        20.0,
		RiskReductionPercentage:   35.0,
		Status:                    model.CSIBenefitPartiallyRealized,
	})
	if err != nil {
		t.Fatalf("failed to record benefit: %v", err)
	}

	// Effort cost: 80 * $75 = $6,000
	// Annual benefit: $15,000 + (20 * 12 * $75 = $18,000) = $33,000
	// ROI = ($33,000 - $6,000) / $6,000 = 4.5
	roi := benefit.CalculateROI(75.0)
	if roi < 4.0 {
		t.Fatalf("expected ROI >= 4.0, got %f", roi)
	}

	// 2. Update to fully realized
	fullyRealizedStatus := model.CSIBenefitFullyRealized
	realizedSavings := 52000.0
	updated, err := svc.UpdateBenefit(ctx, init.ID, model.UpdateCSIBenefitRequest{
		RealizedFinancialSavings: &realizedSavings,
		Status:                   &fullyRealizedStatus,
	})
	if err != nil {
		t.Fatalf("failed to update benefit: %v", err)
	}
	if updated.Status != model.CSIBenefitFullyRealized || updated.RealizedAt == nil {
		t.Fatalf("expected fully realized benefit with timestamp: %+v", updated)
	}
}

func TestCSI_ActionPlanMilestones(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Automated Security Patch Pipeline",
		StrategicAlignment: model.CSIAlignmentSecurity,
	})
	if err != nil {
		t.Fatalf("failed to create initiative: %v", err)
	}

	// 1. Create sequential action steps
	a1, err := svc.CreateAction(ctx, "ws_1", init.ID, model.CreateCSIActionRequest{
		Description: "Configure vulnerability staging runner",
		AssigneeID:  "user_1",
		DueDate:     time.Now().UTC().AddDate(0, 0, 14),
	})
	if err != nil {
		t.Fatalf("failed to create action step 1: %v", err)
	}
	if a1.StepNumber != 1 {
		t.Fatalf("expected step number 1, got %d", a1.StepNumber)
	}

	a2, err := svc.CreateAction(ctx, "ws_1", init.ID, model.CreateCSIActionRequest{
		Description: "Run canary vulnerability patch in staging",
		AssigneeID:  "user_2",
		DueDate:     time.Now().UTC().AddDate(0, 0, 28),
	})
	if err != nil {
		t.Fatalf("failed to create action step 2: %v", err)
	}
	if a2.StepNumber != 2 {
		t.Fatalf("expected step number 2, got %d", a2.StepNumber)
	}

	// 2. Complete Step 1 with verification evidence
	completedA1, err := svc.CompleteAction(ctx, a1.ID, "Pipeline green in gitlab-ci #4021")
	if err != nil {
		t.Fatalf("failed to complete action 1: %v", err)
	}
	if completedA1.Status != model.CSIActionCompleted || completedA1.CompletedAt == nil {
		t.Fatalf("action 1 should be completed")
	}

	// 3. List Actions
	actions, err := svc.ListActions(ctx, init.ID)
	if err != nil || len(actions) != 2 {
		t.Fatalf("expected 2 action items, got %d", len(actions))
	}
	if actions[0].StepNumber != 1 || actions[1].StepNumber != 2 {
		t.Fatalf("action items should be ordered by step number")
	}
}

func TestCSI_GovernanceReviews(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Incident Response Runbook Refresh",
		StrategicAlignment: model.CSIAlignmentReliability,
	})
	if err != nil {
		t.Fatalf("failed to create initiative: %v", err)
	}

	// 1. Record Governance Review
	nextDate := time.Now().UTC().AddDate(0, 1, 0)
	rev, err := svc.RecordReview(ctx, "ws_1", init.ID, model.RecordCSIReviewRequest{
		ReviewDate:     time.Now().UTC(),
		ReviewerID:     "user_2",
		Outcome:        model.CSIReviewOutcomeOnTrack,
		Findings:       "All tier 1 runbooks tested during continuity drill successfully",
		LessonsLearned: "Automated alert paging reduced triage latency by 6 minutes",
		NextReviewDate: &nextDate,
	})
	if err != nil {
		t.Fatalf("failed to record review: %v", err)
	}
	if rev.Outcome != model.CSIReviewOutcomeOnTrack {
		t.Fatalf("expected on_track outcome, got %s", rev.Outcome)
	}

	// 2. List Reviews
	reviews, err := svc.ListReviews(ctx, init.ID)
	if err != nil || len(reviews) != 1 {
		t.Fatalf("expected 1 review, got %d", len(reviews))
	}
}

func TestCSI_AnalyticsAndMaturityPosture(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Create two initiatives: one active, one completed
	init1, _ := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "Reliability Drive",
		StrategicAlignment: model.CSIAlignmentReliability,
		Status:             model.CSIStatusInProgress,
	})
	init2, _ := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
		Title:              "FinOps Optimization",
		StrategicAlignment: model.CSIAlignmentCostReduction,
		Status:             model.CSIStatusCompleted,
	})

	// 2. Add Benefits
	_, _ = svc.RecordBenefit(ctx, "ws_1", init1.ID, model.RecordCSIBenefitRequest{
		ProjectedFinancialSavings: 20000.0,
		RealizedFinancialSavings:  5000.0,
		EffortHoursInvested:       20.0,
		HoursSavedPerMonth:        10.0,
	})
	_, _ = svc.RecordBenefit(ctx, "ws_1", init2.ID, model.RecordCSIBenefitRequest{
		ProjectedFinancialSavings: 30000.0,
		RealizedFinancialSavings:  30000.0,
		EffortHoursInvested:       40.0,
		HoursSavedPerMonth:        25.0,
		Status:                    model.CSIBenefitFullyRealized,
	})

	// 3. Get Analytics
	analytics, err := svc.GetCSIAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get CSI analytics: %v", err)
	}

	if analytics.TotalInitiatives != 2 {
		t.Fatalf("expected 2 total initiatives, got %d", analytics.TotalInitiatives)
	}
	if analytics.ActiveInitiatives != 1 || analytics.CompletedInitiatives != 1 {
		t.Fatalf("unexpected active/completed initiatives: %+v", analytics)
	}
	if analytics.CompletionRate != 50.0 {
		t.Fatalf("expected 50%% completion rate, got %f", analytics.CompletionRate)
	}
	if analytics.TotalRealizedSavings != 35000.0 {
		t.Fatalf("expected $35,000 realized savings, got %f", analytics.TotalRealizedSavings)
	}
	if analytics.MaturityScore <= 0 || analytics.MaturityScore > 100 {
		t.Fatalf("invalid maturity score: %f", analytics.MaturityScore)
	}
	if len(analytics.InitiativesByAlignment) != 2 {
		t.Fatalf("expected 2 alignment categories, got %d", len(analytics.InitiativesByAlignment))
	}
}

func TestCSI_ConcurrencySafe(t *testing.T) {
	svc, _, dbConn := setupTestCSISvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			init, err := svc.CreateInitiative(ctx, "ws_1", model.CreateCSIInitiativeRequest{
				Title:              "Worker Initiative",
				StrategicAlignment: model.CSIAlignmentVelocity,
			})
			if err == nil && init != nil {
				_, _ = svc.AddBenchmark(ctx, "ws_1", init.ID, model.CreateCSIBenchmarkRequest{
					MetricName:    "Throughput",
					UnitOfMeasure: model.CSIMetricUnitCount,
					BaselineValue: 100.0,
					TargetValue:   500.0,
				})
				_, _ = svc.RecordBenefit(ctx, "ws_1", init.ID, model.RecordCSIBenefitRequest{
					EffortHoursInvested: 10.0,
				})
			}
			_, _, _ = svc.ListInitiatives(ctx, model.CSIFilterCriteria{
				WorkspaceID: "ws_1",
			})
			_, _ = svc.GetCSIAnalytics(ctx, "ws_1")
		}(i)
	}

	wg.Wait()
}
