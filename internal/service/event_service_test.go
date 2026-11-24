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

func setupTestEventSvc(t *testing.T) (EventService, repository.EventRepository, *sql.DB) {
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
		VALUES (1, 'Monitoring WS', 'mon-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewEventRepository(sqliteDB)
	svc := NewEventService(repo)
	return svc, repo, sqliteDB
}

func TestEvent_IngestionAndDeduplication(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing source
	_, _, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Resource: "api-gateway",
		Message:  "504 Gateway Timeout",
	})
	if err == nil {
		t.Fatal("expected error for missing source, got nil")
	}

	// 2. Ingest first event -> should create a new alert
	event1, alert1, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourceDatadog,
		Severity: model.EventSeverityHigh,
		Resource: "api-gateway",
		Message:  "High 504 Gateway Timeout rate",
	})
	if err != nil {
		t.Fatalf("failed to ingest event 1: %v", err)
	}
	if event1.ID == "" || alert1.ID == "" {
		t.Fatalf("expected non-empty IDs, got event ID: %s, alert ID: %s", event1.ID, alert1.ID)
	}
	if alert1.EventCount != 1 {
		t.Fatalf("expected alert event count 1, got %d", alert1.EventCount)
	}
	if alert1.Status != model.AlertStatusOpen {
		t.Fatalf("expected alert status 'open', got '%s'", alert1.Status)
	}

	// 3. Ingest duplicate event with same fingerprint -> should cluster/aggregate into existing alert
	event2, alert2, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:      model.EventSourceDatadog,
		Severity:    model.EventSeverityHigh,
		Resource:    "api-gateway",
		Message:     "High 504 Gateway Timeout rate",
		Fingerprint: event1.Fingerprint,
	})
	if err != nil {
		t.Fatalf("failed to ingest event 2: %v", err)
	}
	if alert2.ID != alert1.ID {
		t.Fatalf("expected clustered alert ID %s, got %s", alert1.ID, alert2.ID)
	}
	if alert2.EventCount != 2 {
		t.Fatalf("expected event count 2, got %d", alert2.EventCount)
	}
	if event2.ID == event1.ID {
		t.Fatalf("expected separate raw event record IDs")
	}

	// 4. Ingest third event with higher severity (critical) -> should upgrade alert severity and priority
	_, alert3, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:      model.EventSourceDatadog,
		Severity:    model.EventSeverityCritical,
		Resource:    "api-gateway",
		Message:     "Critical outage on api-gateway",
		Fingerprint: event1.Fingerprint,
	})
	if err != nil {
		t.Fatalf("failed to ingest event 3: %v", err)
	}
	if alert3.EventCount != 3 {
		t.Fatalf("expected event count 3, got %d", alert3.EventCount)
	}
	if alert3.Severity != model.EventSeverityCritical {
		t.Fatalf("expected alert severity 'critical', got '%s'", alert3.Severity)
	}
	if alert3.Priority != model.AlertPriorityP1 {
		t.Fatalf("expected alert priority 'p1', got '%s'", alert3.Priority)
	}

	// 5. GetEventByID and ListEvents
	gotEvent, err := svc.GetEventByID(ctx, event1.ID)
	if err != nil {
		t.Fatalf("failed to get event by ID: %v", err)
	}
	if gotEvent.Resource != "api-gateway" {
		t.Fatalf("expected resource 'api-gateway', got '%s'", gotEvent.Resource)
	}

	events, total, err := svc.ListEvents(ctx, model.EventRecordFilter{
		WorkspaceID: "ws_1",
		Source:      model.EventSourceDatadog,
	})
	if err != nil {
		t.Fatalf("failed to list events: %v", err)
	}
	if total != 3 || len(events) != 3 {
		t.Fatalf("expected 3 events, got total %d, count %d", total, len(events))
	}
}

func TestEvent_SuppressionAndBlackout(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	now := time.Now().UTC()
	// Create suppression window for CI-9001
	suppression, err := svc.CreateSuppression(ctx, "ws_1", model.CreateEventSuppressionRequest{
		Name:            "Scheduled Database Maintenance",
		Reason:          "Upgrading PostgreSQL Aurora cluster",
		CIID:            "CI-9001",
		ResourcePattern: "db-primary",
		StartTime:       now.Add(-10 * time.Minute),
		EndTime:         now.Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("failed to create suppression: %v", err)
	}
	if suppression.ID == "" || !suppression.IsActive {
		t.Fatalf("invalid suppression state: %+v", suppression)
	}

	// Ingest event matching suppression window and CI
	_, alertSuppressed, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourcePrometheus,
		Severity: model.EventSeverityHigh,
		CIID:     "CI-9001",
		Resource: "db-primary",
		Message:  "Database connection pool exhausted",
	})
	if err != nil {
		t.Fatalf("failed to ingest suppressed event: %v", err)
	}
	if alertSuppressed.Status != model.AlertStatusSuppressed {
		t.Fatalf("expected alert status 'suppressed', got '%s'", alertSuppressed.Status)
	}

	// Ingest event for different CI outside suppression
	_, alertActive, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourcePrometheus,
		Severity: model.EventSeverityMedium,
		CIID:     "CI-9999",
		Resource: "redis-cache",
		Message:  "High memory usage on redis-cache",
	})
	if err != nil {
		t.Fatalf("failed to ingest unsuppressed event: %v", err)
	}
	if alertActive.Status != model.AlertStatusOpen {
		t.Fatalf("expected alert status 'open', got '%s'", alertActive.Status)
	}

	// Suppression CRUD verification
	gotSup, err := svc.GetSuppressionByID(ctx, suppression.ID)
	if err != nil {
		t.Fatalf("failed to get suppression by ID: %v", err)
	}
	if gotSup.Name != "Scheduled Database Maintenance" {
		t.Fatalf("unexpected suppression name: %s", gotSup.Name)
	}

	sups, total, err := svc.ListSuppressions(ctx, model.EventSuppressionFilter{
		WorkspaceID: "ws_1",
	})
	if err != nil || total != 1 || len(sups) != 1 {
		t.Fatalf("expected 1 suppression, got total %d, err %v", total, err)
	}

	if err := svc.DeleteSuppression(ctx, suppression.ID); err != nil {
		t.Fatalf("failed to delete suppression: %v", err)
	}
}

func TestEvent_CorrelationRulesAndAutoEscalation(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create correlation rule: datadog critical -> escalate to incident
	rule, err := svc.CreateRule(ctx, "ws_1", model.CreateCorrelationRuleRequest{
		Name:              "Auto-Escalate Critical Datadog",
		SourcePattern:     model.EventSourceDatadog,
		SeverityPattern:   model.EventSeverityCritical,
		TimeWindowSeconds: 300,
		ThresholdCount:    1,
		ActionType:        model.RuleActionEscalateIncident,
		IsActive:          true,
	})
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}
	if rule.ID == "" {
		t.Fatalf("expected non-empty rule ID")
	}

	// Ingest matching critical event
	_, alert, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourceDatadog,
		Severity: model.EventSeverityCritical,
		Resource: "payment-service",
		Message:  "Payment gateway 500 error spike",
	})
	if err != nil {
		t.Fatalf("failed to ingest event: %v", err)
	}
	if alert.RuleID != rule.ID {
		t.Fatalf("expected rule_id %s, got %s", rule.ID, alert.RuleID)
	}
	if alert.Status != model.AlertStatusEscalated {
		t.Fatalf("expected alert status 'escalated_to_incident', got '%s'", alert.Status)
	}
	if alert.IncidentID == "" {
		t.Fatalf("expected linked incident_id to be populated")
	}

	// Verify automated action was recorded
	actions, total, err := svc.ListActions(ctx, model.EventActionFilter{
		WorkspaceID: "ws_1",
		AlertID:     alert.ID,
	})
	if err != nil || total != 1 || len(actions) != 1 {
		t.Fatalf("expected 1 automated action, got total %d, err %v", total, err)
	}
	if actions[0].ActionType != model.ActionTriggerIncident {
		t.Fatalf("expected action type trigger_incident, got %s", actions[0].ActionType)
	}

	// Rule CRUD verification
	ruleDesc := "Updated rule description"
	updatedRule, err := svc.UpdateRule(ctx, rule.ID, model.UpdateCorrelationRuleRequest{
		Description: &ruleDesc,
	})
	if err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}
	if updatedRule.Description != ruleDesc {
		t.Fatalf("expected updated description, got %s", updatedRule.Description)
	}

	rules, totalRules, err := svc.ListRules(ctx, model.CorrelationRuleFilter{WorkspaceID: "ws_1"})
	if err != nil || totalRules != 1 || len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", totalRules)
	}

	if err := svc.DeleteRule(ctx, rule.ID); err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}
}

func TestEvent_AlertLifecycleAndActions(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Ingest event to generate alert
	_, alert, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourceCloudWatch,
		Severity: model.EventSeverityMedium,
		Resource: "lambda-auth",
		Message:  "Lambda throttling detected",
	})
	if err != nil {
		t.Fatalf("failed to ingest event: %v", err)
	}

	// 1. Acknowledge alert
	ackAlert, err := svc.UpdateAlertStatus(ctx, alert.ID, model.UpdateAlertStatusRequest{
		Status:         model.AlertStatusAcknowledged,
		AcknowledgedBy: "alice",
	})
	if err != nil {
		t.Fatalf("failed to acknowledge alert: %v", err)
	}
	if ackAlert.Status != model.AlertStatusAcknowledged {
		t.Fatalf("expected status 'acknowledged', got '%s'", ackAlert.Status)
	}
	if ackAlert.AcknowledgedBy != "alice" || ackAlert.AcknowledgedAt == nil {
		t.Fatalf("expected acknowledged_by and acknowledged_at to be set")
	}

	// 2. Dispatch manual action / remediation
	action, err := svc.TriggerAction(ctx, "ws_1", model.TriggerEventActionRequest{
		AlertID:    alert.ID,
		ActionType: model.ActionRunRemediation,
		OutputLogs: "Increased Lambda provisioned concurrency to 50",
	})
	if err != nil {
		t.Fatalf("failed to trigger action: %v", err)
	}
	if action.ID == "" || action.Status != model.ActionStatusExecuted {
		t.Fatalf("unexpected action state: %+v", action)
	}

	// 3. Resolve alert
	resAlert, err := svc.UpdateAlertStatus(ctx, alert.ID, model.UpdateAlertStatusRequest{
		Status: model.AlertStatusResolved,
	})
	if err != nil {
		t.Fatalf("failed to resolve alert: %v", err)
	}
	if resAlert.Status != model.AlertStatusResolved || resAlert.ResolvedAt == nil {
		t.Fatalf("expected resolved status with timestamp")
	}

	// 4. List alerts with filters
	alerts, total, err := svc.ListAlerts(ctx, model.CorrelatedAlertFilter{
		WorkspaceID: "ws_1",
		Status:      model.AlertStatusResolved,
	})
	if err != nil || total != 1 || len(alerts) != 1 {
		t.Fatalf("expected 1 resolved alert, got total %d", total)
	}
}

func TestEvent_AnalyticsAndNoiseReduction(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Ingest 5 events on same fingerprint -> creates 1 alert with count 5
	for i := 0; i < 5; i++ {
		_, _, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
			Source:      model.EventSourcePrometheus,
			Severity:    model.EventSeverityHigh,
			Resource:    "k8s-ingress",
			Message:     "502 Bad Gateway from upstream pods",
			Fingerprint: "prom:k8s-ingress:high",
		})
		if err != nil {
			t.Fatalf("failed to ingest repeated event: %v", err)
		}
	}

	// Ingest 1 separate event -> creates another alert
	_, _, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
		Source:   model.EventSourceGrafana,
		Severity: model.EventSeverityLow,
		Resource: "grafana-dashboards",
		Message:  "Dashboard rendering slow",
	})
	if err != nil {
		t.Fatalf("failed to ingest single event: %v", err)
	}

	analytics, err := svc.GetEventAnalytics(ctx, "ws_1")
	if err != nil {
		t.Fatalf("failed to compute event analytics: %v", err)
	}

	if analytics.TotalEventsIngested != 6 {
		t.Fatalf("expected 6 total events, got %d", analytics.TotalEventsIngested)
	}
	if analytics.TotalAlertsCreated != 2 {
		t.Fatalf("expected 2 total alerts, got %d", analytics.TotalAlertsCreated)
	}
	if analytics.NoiseReductionRatio <= 50.0 {
		t.Fatalf("expected noise reduction ratio > 50%%, got %.2f%%", analytics.NoiseReductionRatio)
	}
	if analytics.EventHealthIndex <= 0.0 || analytics.EventHealthIndex > 100.0 {
		t.Fatalf("expected EventHealthIndex between 0 and 100, got %.2f", analytics.EventHealthIndex)
	}
	if analytics.SourceDistribution[model.EventSourcePrometheus] != 5 {
		t.Fatalf("expected 5 prometheus events, got %d", analytics.SourceDistribution[model.EventSourcePrometheus])
	}
}

func TestEvent_Concurrency(t *testing.T) {
	svc, _, dbConn := setupTestEventSvc(t)
	defer dbConn.Close()

	ctx := context.Background()
	concurrency := 15
	var wg sync.WaitGroup
	errCh := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, _, err := svc.IngestEvent(ctx, "ws_1", model.IngestEventRequest{
				Source:      model.EventSourcePrometheus,
				Severity:    model.EventSeverityMedium,
				Resource:    fmt.Sprintf("node-%d", idx%3),
				Message:     fmt.Sprintf("CPU spike on node-%d", idx%3),
				Fingerprint: fmt.Sprintf("prom:node-%d:medium", idx%3),
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
			t.Fatalf("concurrent event ingestion error: %v", err)
		}
	}

	events, total, err := svc.ListEvents(ctx, model.EventRecordFilter{WorkspaceID: "ws_1"})
	if err != nil {
		t.Fatalf("failed to list events after concurrency: %v", err)
	}
	if total != concurrency || len(events) != concurrency {
		t.Fatalf("expected %d events, got %d", concurrency, total)
	}
}
