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

func setupTestAvailabilitySvc(t *testing.T) (AvailabilityService, repository.AvailabilityRepository, *sql.DB) {
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
		VALUES (1, 'Avail WS', 'avail-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewAvailabilityRepository(sqliteDB)
	svc := NewAvailabilityService(repo)
	return svc, repo, sqliteDB
}

func TestAvailability_TargetLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing name
	_, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		TargetPercentage: 99.9,
	})
	if err != ErrAvailabilityTargetNameRequired {
		t.Fatalf("expected ErrAvailabilityTargetNameRequired, got %v", err)
	}

	// 2. Validation failure: invalid target percentage
	_, err = svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "API Gateway Target",
		TargetPercentage: 105.0,
	})
	if err != model.ErrInvalidTargetPercentage {
		t.Fatalf("expected ErrInvalidTargetPercentage, got %v", err)
	}

	// 3. Create Target
	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:               "Core Payment Gateway",
		TargetPercentage:   99.95,
		MeasurementWindow:  model.MeasurementWindowMonthly,
		ServiceHours:       model.ServiceHours24x7,
		CommittedMTBFHours: 720.0,
		TargetMTTRMinutes:  15.0,
		CIID:               "ci-payment-svc",
		Description:        "Production payments availability target",
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}
	if target.ID == "" || target.Name != "Core Payment Gateway" {
		t.Fatalf("unexpected target data: %+v", target)
	}

	// 4. Duplicate name conflict
	_, err = svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "Core Payment Gateway",
		TargetPercentage: 99.9,
	})
	if err != ErrAvailabilityTargetNameConflict {
		t.Fatalf("expected ErrAvailabilityTargetNameConflict, got %v", err)
	}

	// 5. Get by ID
	fetched, err := svc.GetTargetByID(ctx, target.ID)
	if err != nil {
		t.Fatalf("failed to get target by ID: %v", err)
	}
	if fetched.Name != target.Name {
		t.Fatalf("expected target name %s, got %s", target.Name, fetched.Name)
	}

	// 6. Update Target
	newPct := 99.99
	newDesc := "Updated SLA commitment"
	updated, err := svc.UpdateTarget(ctx, target.ID, model.UpdateAvailabilityTargetRequest{
		TargetPercentage: &newPct,
		Description:      &newDesc,
	})
	if err != nil {
		t.Fatalf("failed to update target: %v", err)
	}
	if updated.TargetPercentage != 99.99 || updated.Description != newDesc {
		t.Fatalf("update did not reflect: %+v", updated)
	}

	// 7. List Targets
	list, total, err := svc.ListTargets(ctx, model.AvailabilityFilterCriteria{
		WorkspaceID: "ws_1",
	})
	if err != nil {
		t.Fatalf("failed to list targets: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 target in list, got %d", total)
	}

	// 8. Delete Target
	if err := svc.DeleteTarget(ctx, target.ID); err != nil {
		t.Fatalf("failed to delete target: %v", err)
	}
	_, err = svc.GetTargetByID(ctx, target.ID)
	if err != ErrAvailabilityTargetNotFound {
		t.Fatalf("expected ErrAvailabilityTargetNotFound after delete, got %v", err)
	}
}

func TestAvailability_OutageLifecycleAndDuration(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "Database Cluster",
		TargetPercentage: 99.9,
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	started := time.Now().UTC().Add(-2 * time.Hour)
	ended := started.Add(45 * time.Minute)

	// 1. Record completed outage
	outage, err := svc.RecordOutage(ctx, "ws_1", model.CreateOutageEventRequest{
		TargetID:        target.ID,
		CIID:            "ci-db-cluster",
		Title:           "Primary Node Kernel Panic",
		OutageType:      model.OutageTypeUnplanned,
		Severity:        model.OutageSeverityCritical,
		StartedAt:       started,
		EndedAt:         &ended,
		RootCause:       "Memory allocation leak in hypervisor",
		IsSLAExempt:     false,
		ReportedBy:      "user_1",
	})
	if err != nil {
		t.Fatalf("failed to record outage: %v", err)
	}
	if outage.DurationMinutes != 45.0 {
		t.Fatalf("expected duration 45 minutes, got %f", outage.DurationMinutes)
	}

	// 2. Record ongoing outage and resolve it
	ongoingStart := time.Now().UTC().Add(-30 * time.Minute)
	ongoing, err := svc.RecordOutage(ctx, "ws_1", model.CreateOutageEventRequest{
		TargetID:   target.ID,
		CIID:       "ci-db-cluster",
		Title:      "Secondary Replica Sync Lag",
		OutageType: model.OutageTypeUnplanned,
		Severity:   model.OutageSeverityMajor,
		StartedAt:  ongoingStart,
	})
	if err != nil {
		t.Fatalf("failed to record ongoing outage: %v", err)
	}
	if ongoing.DurationMinutes != 0 || ongoing.EndedAt != nil {
		t.Fatalf("ongoing outage should have 0 duration and nil EndedAt")
	}

	// Resolve ongoing outage
	resolveTime := ongoingStart.Add(25 * time.Minute)
	resolved, err := svc.ResolveOutage(ctx, ongoing.ID, model.ResolveOutageRequest{
		EndedAt:   resolveTime,
		RootCause: "Re-established replication tunnel",
	})
	if err != nil {
		t.Fatalf("failed to resolve outage: %v", err)
	}
	if resolved.DurationMinutes != 25.0 {
		t.Fatalf("expected resolved duration 25 minutes, got %f", resolved.DurationMinutes)
	}

	// 3. List Outages
	outages, total, err := svc.ListOutages(ctx, model.AvailabilityFilterCriteria{
		WorkspaceID: "ws_1",
		TargetID:    target.ID,
	})
	if err != nil {
		t.Fatalf("failed to list outages: %v", err)
	}
	if total != 2 || len(outages) != 2 {
		t.Fatalf("expected 2 outages, got total %d", total)
	}

	// 4. Delete Outage
	if err := svc.DeleteOutage(ctx, ongoing.ID); err != nil {
		t.Fatalf("failed to delete outage: %v", err)
	}
	_, err = svc.GetOutageByID(ctx, ongoing.ID)
	if err != ErrOutageEventNotFound {
		t.Fatalf("expected ErrOutageEventNotFound, got %v", err)
	}
}

func TestAvailability_MeasurementCalculationAndSLABreach(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// High availability target: 99.9%
	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:               "Customer Portal",
		TargetPercentage:   99.9,
		MeasurementWindow:  model.MeasurementWindowMonthly,
		CommittedMTBFHours: 500.0,
		TargetMTTRMinutes:  30.0,
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Add an outage: 120 minutes non-exempt
	outageStart := time.Now().UTC().Add(-5 * time.Hour)
	outageEnd := outageStart.Add(120 * time.Minute)
	_, err = svc.RecordOutage(ctx, "ws_1", model.CreateOutageEventRequest{
		TargetID:    target.ID,
		Title:       "Data Center Power Transient",
		OutageType:  model.OutageTypeUnplanned,
		Severity:    model.OutageSeverityCritical,
		StartedAt:   outageStart,
		EndedAt:     &outageEnd,
		IsSLAExempt: false,
	})
	if err != nil {
		t.Fatalf("failed to record outage: %v", err)
	}

	// Monthly total operational minutes: 43200 (30 days * 24 * 60)
	// Non-exempt downtime: 120 minutes
	// Uptime %: (43200 - 120) / 43200 * 100 = 99.722%
	// SLA Breach: 99.722% < 99.9% -> true!
	measurement, err := svc.CalculatePeriodMeasurement(ctx, "ws_1", target.ID, "2026-M11", 43200.0)
	if err != nil {
		t.Fatalf("failed to calculate measurement: %v", err)
	}

	if !measurement.IsSLABreached {
		t.Fatalf("expected SLA to be breached (99.722%% < 99.9%%)")
	}
	if measurement.OutageCount != 1 {
		t.Fatalf("expected 1 outage count, got %d", measurement.OutageCount)
	}
	if measurement.TotalDowntimeMinutes != 120.0 {
		t.Fatalf("expected 120 minutes downtime, got %f", measurement.TotalDowntimeMinutes)
	}

	// Fetch measurement by ID
	fetched, err := svc.GetMeasurementByID(ctx, measurement.ID)
	if err != nil {
		t.Fatalf("failed to get measurement: %v", err)
	}
	if fetched.ActualUptimePercentage != measurement.ActualUptimePercentage {
		t.Fatalf("mismatched uptime percentage: %f vs %f", fetched.ActualUptimePercentage, measurement.ActualUptimePercentage)
	}
}

func TestAvailability_SPOFRegisterAndMitigation(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Register a critical SPOF
	spof, err := svc.RegisterSPOF(ctx, "ws_1", model.CreateSPOFRequest{
		ComponentName:       "Legacy Core Authentication LDAP",
		CIID:                "ci-ldap-auth",
		RiskLevel:           model.SPOFRiskCritical,
		RedundancyMechanism: model.RedundancyMechanismNone,
		MitigationPlan:      "Migrate to multi-region cloud identity provider",
		OwnerID:             "user_1",
	})
	if err != nil {
		t.Fatalf("failed to register SPOF: %v", err)
	}
	if spof.Status != model.SPOFStatusOpen {
		t.Fatalf("expected open status, got %s", spof.Status)
	}

	// 2. Update SPOF
	newMech := model.RedundancyMechanismActivePassive
	updated, err := svc.UpdateSPOF(ctx, spof.ID, model.UpdateSPOFRequest{
		RedundancyMechanism: &newMech,
	})
	if err != nil {
		t.Fatalf("failed to update SPOF: %v", err)
	}
	if updated.RedundancyMechanism != model.RedundancyMechanismActivePassive {
		t.Fatalf("expected active_passive mechanism, got %s", updated.RedundancyMechanism)
	}

	// 3. Mitigate SPOF
	mitigated, err := svc.MitigateSPOF(ctx, spof.ID, "Hot-standby LDAP replica deployed and DNS failover tested")
	if err != nil {
		t.Fatalf("failed to mitigate SPOF: %v", err)
	}
	if mitigated.Status != model.SPOFStatusMitigated {
		t.Fatalf("expected status mitigated, got %s", mitigated.Status)
	}

	// 4. List SPOFs
	list, total, err := svc.ListSPOFs(ctx, model.AvailabilityFilterCriteria{
		WorkspaceID: "ws_1",
		Status:      model.SPOFStatusMitigated,
	})
	if err != nil {
		t.Fatalf("failed to list SPOFs: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 mitigated SPOF, got %d", total)
	}
}

func TestAvailability_TestSchedulesAndResults(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "Payment Webhooks",
		TargetPercentage: 99.9,
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// 1. Schedule a failover test
	testRecord, err := svc.ScheduleTest(ctx, "ws_1", model.CreateAvailabilityTestRequest{
		TargetID:    target.ID,
		Name:        "AZ-East Failover Simulation",
		TestType:    model.TestTypeFailover,
		Frequency:   model.TestFrequencyQuarterly,
		NextTestDue: time.Now().UTC().AddDate(0, 1, 0),
		ConductedBy: "user_2",
	})
	if err != nil {
		t.Fatalf("failed to schedule test: %v", err)
	}
	if testRecord.Outcome != model.TestOutcomePending {
		t.Fatalf("expected pending outcome, got %s", testRecord.Outcome)
	}

	// 2. Record Test Result (Passed)
	testedTime := time.Now().UTC()
	nextDate := testedTime.AddDate(0, 3, 0)
	completed, err := svc.RecordTestResult(ctx, testRecord.ID, model.RecordTestResultRequest{
		TestedAt:    testedTime,
		Outcome:     model.TestOutcomePassed,
		Findings:    "DNS health probes rerouted traffic within 18 seconds with 0 dropped sessions",
		NextTestDue: nextDate,
		ConductedBy: "user_2",
	})
	if err != nil {
		t.Fatalf("failed to record test result: %v", err)
	}
	if completed.Outcome != model.TestOutcomePassed || completed.LastTestedAt == nil {
		t.Fatalf("expected passed outcome and non-nil LastTestedAt")
	}

	// 3. List Tests
	tests, total, err := svc.ListTests(ctx, model.AvailabilityFilterCriteria{
		WorkspaceID: "ws_1",
		Outcome:     model.TestOutcomePassed,
	})
	if err != nil {
		t.Fatalf("failed to list tests: %v", err)
	}
	if total != 1 || len(tests) != 1 {
		t.Fatalf("expected 1 test, got %d", total)
	}
}

func TestAvailability_AnalyticsAndResiliencePosture(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Setup Target
	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "Authentication Cluster",
		TargetPercentage: 99.9,
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	// Setup Outage
	s := time.Now().UTC().Add(-1 * time.Hour)
	e := s.Add(10 * time.Minute)
	_, err = svc.RecordOutage(ctx, "ws_1", model.CreateOutageEventRequest{
		TargetID:   target.ID,
		CIID:       "ci-auth",
		Title:      "Token cache eviction",
		OutageType: model.OutageTypeUnplanned,
		Severity:   model.OutageSeverityMajor,
		StartedAt:  s,
		EndedAt:    &e,
	})
	if err != nil {
		t.Fatalf("failed to record outage: %v", err)
	}

	// Setup Measurement
	_, err = svc.CalculatePeriodMeasurement(ctx, "ws_1", target.ID, "2026-M11", 43200.0)
	if err != nil {
		t.Fatalf("failed to calculate measurement: %v", err)
	}

	// Setup Critical SPOF
	_, err = svc.RegisterSPOF(ctx, "ws_1", model.CreateSPOFRequest{
		ComponentName: "Root Certificate Authority HSM",
		RiskLevel:     model.SPOFRiskCritical,
	})
	if err != nil {
		t.Fatalf("failed to register SPOF: %v", err)
	}

	// Get Analytics
	analytics, err := svc.GetAvailabilityAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to get analytics: %v", err)
	}

	if analytics.TotalTargets != 1 || analytics.ActiveTargets != 1 {
		t.Fatalf("unexpected targets count: %+v", analytics)
	}
	if analytics.TotalOutagesRecorded != 1 {
		t.Fatalf("expected 1 total outage, got %d", analytics.TotalOutagesRecorded)
	}
	if analytics.CriticalSPOFCount != 1 {
		t.Fatalf("expected 1 critical SPOF, got %d", analytics.CriticalSPOFCount)
	}
	if len(analytics.TopOutagesByCI) != 1 || analytics.TopOutagesByCI[0].CIID != "ci-auth" {
		t.Fatalf("expected top outage CI ci-auth, got %+v", analytics.TopOutagesByCI)
	}
	if analytics.HealthIndexScore <= 0 || analytics.HealthIndexScore > 100 {
		t.Fatalf("invalid health index score: %f", analytics.HealthIndexScore)
	}
}

func TestAvailability_ConcurrencySafe(t *testing.T) {
	svc, _, dbConn := setupTestAvailabilitySvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	target, err := svc.CreateTarget(ctx, "ws_1", model.CreateAvailabilityTargetRequest{
		Name:             "Shared Database Service",
		TargetPercentage: 99.9,
	})
	if err != nil {
		t.Fatalf("failed to create target: %v", err)
	}

	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			s := time.Now().UTC()
			e := s.Add(time.Duration(workerID+1) * time.Minute)
			_, _ = svc.RecordOutage(ctx, "ws_1", model.CreateOutageEventRequest{
				TargetID:   target.ID,
				CIID:       "ci-shared-db",
				Title:      "Worker query spike",
				OutageType: model.OutageTypeUnplanned,
				Severity:   model.OutageSeverityMinor,
				StartedAt:  s,
				EndedAt:    &e,
			})
			_, _, _ = svc.ListOutages(ctx, model.AvailabilityFilterCriteria{
				WorkspaceID: "ws_1",
			})
			_, _ = svc.GetAvailabilityAnalytics(ctx, "ws_1")
		}(i)
	}

	wg.Wait()
}
