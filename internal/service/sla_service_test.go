package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestSlaSvc(t *testing.T) (SlaService, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize in-memory test db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at)
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_lead', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Operations WS', 'ops-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	slaRepo := repository.NewSlaRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewSlaService(slaRepo, taskRepo)

	return svc, taskRepo, sqliteDB
}

func TestSlaService_PolicyCRUD(t *testing.T) {
	svc, _, _ := setupTestSlaSvc(t)

	// Validation: missing name
	_, err := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "",
		TargetResponseMinutes:   60,
		TargetResolutionMinutes: 240,
	})
	if !errors.Is(err, ErrSlaPolicyNameRequired) {
		t.Fatalf("expected ErrSlaPolicyNameRequired, got: %v", err)
	}

	// Validation: invalid workspace
	_, err = svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             0,
		Name:                    "Valid Name",
		TargetResponseMinutes:   60,
		TargetResolutionMinutes: 240,
	})
	if !errors.Is(err, ErrSlaWorkspaceRequired) {
		t.Fatalf("expected ErrSlaWorkspaceRequired, got: %v", err)
	}

	// Validation: non-positive minutes
	_, err = svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "Valid Name",
		TargetResponseMinutes:   0,
		TargetResolutionMinutes: 240,
	})
	if !errors.Is(err, ErrInvalidTargetMinutes) {
		t.Fatalf("expected ErrInvalidTargetMinutes, got: %v", err)
	}

	// Create valid policy
	bizHours := true
	isDef := true
	policy, err := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "Urgent Incidents SLA",
		Description:             "SLA for critical production issues",
		PriorityTiers:           []string{"urgent", "high"},
		TargetResponseMinutes:   30,
		TargetResolutionMinutes: 120,
		BusinessHoursOnly:       &bizHours,
		IsDefault:               &isDef,
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}
	if policy.ID <= 0 {
		t.Fatalf("expected positive policy ID, got %d", policy.ID)
	}
	if policy.TargetResponseMinutes != 30 || policy.TargetResolutionMinutes != 120 {
		t.Fatalf("unexpected targets: %d, %d", policy.TargetResponseMinutes, policy.TargetResolutionMinutes)
	}

	// Get Policy
	fetched, err := svc.GetPolicy(policy.ID)
	if err != nil {
		t.Fatalf("failed to get policy: %v", err)
	}
	if fetched.Name != "Urgent Incidents SLA" {
		t.Fatalf("expected name 'Urgent Incidents SLA', got %s", fetched.Name)
	}

	// Get Non-existent policy
	_, err = svc.GetPolicy(99999)
	if !errors.Is(err, ErrSlaPolicyNotFound) {
		t.Fatalf("expected ErrSlaPolicyNotFound, got: %v", err)
	}

	// List Policies
	wsID := int64(1)
	active := true
	list, err := svc.ListPolicies(model.SlaFilter{
		WorkspaceID: &wsID,
		IsActive:    &active,
	})
	if err != nil {
		t.Fatalf("failed to list policies: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 policy in list, got %d", len(list))
	}

	// Update Policy
	newName := "Urgent & Critical Incidents SLA"
	newResMinutes := 90
	updated, err := svc.UpdatePolicy(policy.ID, model.UpdateSlaPolicyRequest{
		Name:                    &newName,
		TargetResolutionMinutes: &newResMinutes,
	})
	if err != nil {
		t.Fatalf("failed to update policy: %v", err)
	}
	if updated.Name != newName || updated.TargetResolutionMinutes != 90 {
		t.Fatalf("updated policy mismatch: name=%s, resMin=%d", updated.Name, updated.TargetResolutionMinutes)
	}

	// Toggle Policy
	toggled, err := svc.TogglePolicy(policy.ID, false)
	if err != nil {
		t.Fatalf("failed to toggle policy: %v", err)
	}
	if toggled.IsActive {
		t.Fatalf("expected policy to be inactive")
	}

	// Delete Policy
	if err := svc.DeletePolicy(policy.ID); err != nil {
		t.Fatalf("failed to delete policy: %v", err)
	}
	_, err = svc.GetPolicy(policy.ID)
	if !errors.Is(err, ErrSlaPolicyNotFound) {
		t.Fatalf("expected ErrSlaPolicyNotFound after delete, got %v", err)
	}
}

func TestSlaService_EscalationRules(t *testing.T) {
	svc, _, _ := setupTestSlaSvc(t)

	policy, err := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "Standard SLA",
		TargetResponseMinutes:   60,
		TargetResolutionMinutes: 240,
	})
	if err != nil {
		t.Fatalf("failed to create policy: %v", err)
	}

	// Test rule validation: invalid trigger
	_, err = svc.SetEscalationRules(policy.ID, model.SetEscalationRulesRequest{
		Rules: []model.EscalationRuleDefinition{
			{TriggerEvent: "unknown_event", ActionType: model.SlaActionNotify},
		},
	})
	if !errors.Is(err, ErrInvalidTriggerEvent) {
		t.Fatalf("expected ErrInvalidTriggerEvent, got: %v", err)
	}

	// Test rule validation: invalid action
	_, err = svc.SetEscalationRules(policy.ID, model.SetEscalationRulesRequest{
		Rules: []model.EscalationRuleDefinition{
			{TriggerEvent: model.SlaTriggerBreached, ActionType: "unknown_action"},
		},
	})
	if !errors.Is(err, ErrInvalidSlaActionType) {
		t.Fatalf("expected ErrInvalidSlaActionType, got: %v", err)
	}

	// Valid rules
	rules, err := svc.SetEscalationRules(policy.ID, model.SetEscalationRulesRequest{
		Rules: []model.EscalationRuleDefinition{
			{
				TriggerEvent:        model.SlaTriggerAtRisk,
				ThresholdPercentage: 80,
				ActionType:          model.SlaActionEscalatePriority,
				TargetPriority:      "urgent",
			},
			{
				TriggerEvent: model.SlaTriggerBreached,
				ActionType:   model.SlaActionReassign,
				TargetUserID: "user_lead",
			},
		},
	})
	if err != nil {
		t.Fatalf("failed to set escalation rules: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 escalation rules, got %d", len(rules))
	}

	// Fetch rules
	fetchedRules, err := svc.GetEscalationRules(policy.ID)
	if err != nil {
		t.Fatalf("failed to get escalation rules: %v", err)
	}
	if len(fetchedRules) != 2 {
		t.Fatalf("expected 2 fetched rules, got %d", len(fetchedRules))
	}

	// Delete rule
	if err := svc.DeleteEscalationRule(fetchedRules[0].ID); err != nil {
		t.Fatalf("failed to delete escalation rule: %v", err)
	}
	remainingRules, _ := svc.GetEscalationRules(policy.ID)
	if len(remainingRules) != 1 {
		t.Fatalf("expected 1 rule remaining, got %d", len(remainingRules))
	}
}

func TestSlaService_BusinessHoursAndDeadlines(t *testing.T) {
	// Wednesday at 10:00 AM (during business hours)
	wed10AM := time.Date(2026, 10, 28, 10, 0, 0, 0, time.UTC)
	if !isBusinessHour(wed10AM) {
		t.Fatalf("expected Wednesday 10 AM to be a business hour")
	}

	// Saturday at 10:00 AM (weekend)
	sat10AM := time.Date(2026, 10, 31, 10, 0, 0, 0, time.UTC)
	if isBusinessHour(sat10AM) {
		t.Fatalf("expected Saturday 10 AM NOT to be a business hour")
	}

	// Wednesday at 8:00 PM (outside business hours)
	wed8PM := time.Date(2026, 10, 28, 20, 0, 0, 0, time.UTC)
	if isBusinessHour(wed8PM) {
		t.Fatalf("expected Wednesday 8 PM NOT to be a business hour")
	}

	// Test addMinutes without business hours
	resStandard := addMinutes(wed10AM, 120, false)
	if resStandard != wed10AM.Add(2*time.Hour) {
		t.Fatalf("unexpected standard deadline calculation: %v", resStandard)
	}

	// Test addMinutes with business hours: 60 minutes from 10am should be 11am
	resBiz := addMinutes(wed10AM, 60, true)
	if resBiz.Hour() != 11 {
		t.Fatalf("expected 11 AM, got %d", resBiz.Hour())
	}
}

func TestSlaService_ApplyAndTracking(t *testing.T) {
	svc, taskRepo, _ := setupTestSlaSvc(t)

	// Create task
	task := &model.Task{
		UserID:    "user_1",
		Title:     "Production Outage Investigation",
		Priority:  "urgent",
		Status:    model.TaskStatusTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Create SLA policy
	isDef := true
	policy, err := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "Urgent Task SLA",
		PriorityTiers:           []string{"urgent"},
		TargetResponseMinutes:   15,
		TargetResolutionMinutes: 60,
		IsDefault:               &isDef,
	})
	if err != nil {
		t.Fatalf("failed to create SLA policy: %v", err)
	}

	// Apply SLA to Task explicitly
	rec, err := svc.ApplySlaToTask(model.ApplySlaRequest{
		TaskID:   task.ID,
		PolicyID: &policy.ID,
	})
	if err != nil {
		t.Fatalf("failed to apply SLA to task: %v", err)
	}
	if rec.TaskID != task.ID || rec.PolicyID != policy.ID {
		t.Fatalf("mismatched task SLA record: %+v", rec)
	}
	if rec.ResponseStatus != model.SlaStagePending || rec.ResolutionStatus != model.SlaStagePending {
		t.Fatalf("expected pending status, got resp=%s res=%s", rec.ResponseStatus, rec.ResolutionStatus)
	}

	// Get Task SLA
	fetchedRec, err := svc.GetTaskSla(task.ID)
	if err != nil {
		t.Fatalf("failed to get task SLA: %v", err)
	}
	if fetchedRec.ID != rec.ID {
		t.Fatalf("expected ID %d, got %d", rec.ID, fetchedRec.ID)
	}
}

func TestSlaService_PauseAndResume(t *testing.T) {
	svc, taskRepo, _ := setupTestSlaSvc(t)

	task := &model.Task{
		UserID:    "user_1",
		Title:     "API Latency Spike",
		Priority:  "high",
		Status:    model.TaskStatusTodo,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = taskRepo.Create(task)

	policy, _ := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "High Priority SLA",
		TargetResponseMinutes:   60,
		TargetResolutionMinutes: 240,
	})

	rec, _ := svc.ApplySlaToTask(model.ApplySlaRequest{
		TaskID:   task.ID,
		PolicyID: &policy.ID,
	})

	initialRespDeadline := rec.ResponseDeadline

	// Pause SLA
	paused, err := svc.PauseSla(task.ID, model.PauseSlaRequest{Reason: "Waiting on customer reproduction"})
	if err != nil {
		t.Fatalf("failed to pause SLA: %v", err)
	}
	if !paused.IsPaused || paused.PausedAt == nil {
		t.Fatalf("expected SLA to be paused")
	}

	// Pause already paused
	_, err = svc.PauseSla(task.ID, model.PauseSlaRequest{})
	if !errors.Is(err, ErrSlaAlreadyPaused) {
		t.Fatalf("expected ErrSlaAlreadyPaused, got %v", err)
	}

	// Resume SLA
	resumed, err := svc.ResumeSla(task.ID)
	if err != nil {
		t.Fatalf("failed to resume SLA: %v", err)
	}
	if resumed.IsPaused || resumed.PausedAt != nil {
		t.Fatalf("expected SLA to be resumed")
	}
	if resumed.ResponseDeadline.Before(initialRespDeadline) {
		t.Fatalf("response deadline should not decrease after pause")
	}

	// Resume already active SLA
	_, err = svc.ResumeSla(task.ID)
	if !errors.Is(err, ErrSlaNotPaused) {
		t.Fatalf("expected ErrSlaNotPaused, got %v", err)
	}
}

func TestSlaService_EvaluationAndEscalation(t *testing.T) {
	svc, taskRepo, _ := setupTestSlaSvc(t)

	// Past created time so deadline is immediately breached
	pastTime := time.Now().Add(-2 * time.Hour)
	task := &model.Task{
		UserID:    "user_1",
		Title:     "Critical Database Deadlock",
		Priority:  "medium",
		Status:    model.TaskStatusTodo,
		CreatedAt: pastTime,
		UpdatedAt: pastTime,
	}
	_ = taskRepo.Create(task)

	// Create policy with 10 min response & 30 min resolution
	policy, _ := svc.CreatePolicy("user_1", model.CreateSlaPolicyRequest{
		WorkspaceID:             1,
		Name:                    "Incident Policy",
		TargetResponseMinutes:   10,
		TargetResolutionMinutes: 30,
	})

	// Configure escalation rule: on breach, escalate to urgent and reassign to user_lead
	_, _ = svc.SetEscalationRules(policy.ID, model.SetEscalationRulesRequest{
		Rules: []model.EscalationRuleDefinition{
			{
				TriggerEvent:   model.SlaTriggerBreached,
				ActionType:     model.SlaActionEscalatePriority,
				TargetPriority: "urgent",
			},
			{
				TriggerEvent: model.SlaTriggerBreached,
				ActionType:   model.SlaActionReassign,
				TargetUserID: "user_lead",
			},
		},
	})

	// Apply SLA
	_, err := svc.ApplySlaToTask(model.ApplySlaRequest{
		TaskID:   task.ID,
		PolicyID: &policy.ID,
	})
	if err != nil {
		t.Fatalf("failed to apply SLA: %v", err)
	}

	// Evaluate SLA
	evaluated, err := svc.EvaluateTaskSla(task.ID)
	if err != nil {
		t.Fatalf("failed to evaluate SLA: %v", err)
	}
	if evaluated.ResponseStatus != model.SlaStageBreached || evaluated.ResolutionStatus != model.SlaStageBreached {
		t.Fatalf("expected breached statuses, got resp=%s res=%s", evaluated.ResponseStatus, evaluated.ResolutionStatus)
	}

	// Verify task priority was escalated to urgent and reassigned to user_lead
	updatedTask, _ := taskRepo.GetByID(task.ID)
	if updatedTask.Priority != "urgent" {
		t.Fatalf("expected priority urgent, got %s", updatedTask.Priority)
	}
	if updatedTask.UserID != "user_lead" {
		t.Fatalf("expected assigned user user_lead, got %s", updatedTask.UserID)
	}

	// Verify breach logs were generated
	breaches, err := svc.ListBreaches(model.BreachFilter{TaskID: &task.ID})
	if err != nil {
		t.Fatalf("failed to list breaches: %v", err)
	}
	if len(breaches) == 0 {
		t.Fatalf("expected breach logs to be created")
	}

	// Verify SLA metrics
	metrics, err := svc.GetSlaMetrics(1)
	if err != nil {
		t.Fatalf("failed to get SLA metrics: %v", err)
	}
	if metrics.TotalTrackedTasks != 1 {
		t.Fatalf("expected 1 total tracked task, got %d", metrics.TotalTrackedTasks)
	}
	if metrics.BreachedResponseCount != 1 || metrics.BreachedResolutionCount != 1 {
		t.Fatalf("expected 1 breached response and resolution, got resp=%d res=%d",
			metrics.BreachedResponseCount, metrics.BreachedResolutionCount)
	}
}
