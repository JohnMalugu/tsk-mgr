package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestAutomationSvc(t *testing.T) (AutomationService, repository.TaskRepository, repository.SubtaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize in-memory test db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Marketing WS', 'mkt-ws', 'user_1', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test workspaces: %v", err)
	}

	autoRepo := repository.NewAutomationRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	subtaskRepo := repository.NewSubtaskRepository(sqliteDB)
	notifRepo := repository.NewNotificationRepository(sqliteDB)

	svc := NewAutomationService(autoRepo, taskRepo, subtaskRepo, notifRepo)
	return svc, taskRepo, subtaskRepo, sqliteDB
}

func TestAutomationService_Validation(t *testing.T) {
	svc, _, _, _ := setupTestAutomationSvc(t)

	// 1. Missing workspace ID
	_, err := svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 0,
		Name:        "Auto Triage",
		TriggerType: model.TriggerTaskCreated,
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// 2. Empty rule name
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "   ",
		TriggerType: model.TriggerTaskCreated,
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrRuleNameRequired) {
		t.Fatalf("expected ErrRuleNameRequired, got: %v", err)
	}

	// 3. Name exceeds 100 runes
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        strings.Repeat("r", 101),
		TriggerType: model.TriggerTaskCreated,
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrRuleNameRequired) {
		t.Fatalf("expected ErrRuleNameRequired, got: %v", err)
	}

	// 4. Invalid trigger type
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage",
		TriggerType: "unsupported_trigger",
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrInvalidTriggerType) {
		t.Fatalf("expected ErrInvalidTriggerType, got: %v", err)
	}

	// 5. Invalid condition field
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage",
		TriggerType: model.TriggerTaskCreated,
		Conditions: []model.AutomationCondition{
			{Field: "unknown_field", Operator: model.OpEquals, Value: "foo"},
		},
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrInvalidConditionField) {
		t.Fatalf("expected ErrInvalidConditionField, got: %v", err)
	}

	// 6. Invalid condition operator
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage",
		TriggerType: model.TriggerTaskCreated,
		Conditions: []model.AutomationCondition{
			{Field: model.FieldPriority, Operator: "greater_than", Value: "high"},
		},
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "high"},
		},
	})
	if !errors.Is(err, ErrInvalidConditionOperator) {
		t.Fatalf("expected ErrInvalidConditionOperator, got: %v", err)
	}

	// 7. Empty actions
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage",
		TriggerType: model.TriggerTaskCreated,
		Actions:     []model.AutomationAction{},
	})
	if !errors.Is(err, ErrEmptyActions) {
		t.Fatalf("expected ErrEmptyActions, got: %v", err)
	}

	// 8. Invalid action type
	_, err = svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage",
		TriggerType: model.TriggerTaskCreated,
		Actions: []model.AutomationAction{
			{Type: "drop_database", TargetValue: "production"},
		},
	})
	if !errors.Is(err, ErrInvalidActionType) {
		t.Fatalf("expected ErrInvalidActionType, got: %v", err)
	}
}

func TestAutomationService_CRUD(t *testing.T) {
	svc, _, _, _ := setupTestAutomationSvc(t)

	active := true
	created, err := svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Escalate Urgent Bugs",
		Description: "Automatically sets priority to urgent if title has bug",
		TriggerType: model.TriggerTaskCreated,
		Conditions: []model.AutomationCondition{
			{Field: model.FieldTitle, Operator: model.OpContains, Value: "bug"},
		},
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "urgent"},
			{Type: model.ActionAddTag, TargetValue: "triage"},
		},
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if !created.IsActive {
		t.Fatalf("expected rule to be active")
	}

	// GetRule
	fetched, err := svc.GetRule(created.ID)
	if err != nil {
		t.Fatalf("failed to get rule: %v", err)
	}
	if fetched.Name != "Auto Escalate Urgent Bugs" {
		t.Fatalf("expected rule name match, got '%s'", fetched.Name)
	}
	if len(fetched.Actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(fetched.Actions))
	}

	// UpdateRule
	newName := "Auto Escalate Critical Issues"
	updated, err := svc.UpdateRule(created.ID, model.UpdateAutomationRuleRequest{
		Name: &newName,
	})
	if err != nil {
		t.Fatalf("failed to update rule: %v", err)
	}
	if updated.Name != "Auto Escalate Critical Issues" {
		t.Fatalf("expected updated name, got '%s'", updated.Name)
	}

	// ToggleRule
	toggledActive, err := svc.ToggleRule(created.ID)
	if err != nil {
		t.Fatalf("failed to toggle rule: %v", err)
	}
	if toggledActive {
		t.Fatalf("expected rule to be deactivated (false)")
	}

	// Toggle back
	toggledActive, err = svc.ToggleRule(created.ID)
	if err != nil {
		t.Fatalf("failed to toggle rule back: %v", err)
	}
	if !toggledActive {
		t.Fatalf("expected rule to be reactivated (true)")
	}

	// DuplicateRule
	dup, err := svc.DuplicateRule("user_2", created.ID, model.DuplicateRuleRequest{
		NewName: "Cloned Automation Rule",
	})
	if err != nil {
		t.Fatalf("failed to duplicate rule: %v", err)
	}
	if dup.ID == created.ID {
		t.Fatalf("duplicated rule must have a distinct ID")
	}
	if dup.Name != "Cloned Automation Rule" {
		t.Fatalf("expected cloned name, got '%s'", dup.Name)
	}
	if dup.IsActive {
		t.Fatalf("newly duplicated rule must start in inactive state")
	}

	// ListRules
	listAll, err := svc.ListRules(1, nil, "")
	if err != nil {
		t.Fatalf("failed to list rules: %v", err)
	}
	if len(listAll) != 2 {
		t.Fatalf("expected 2 rules in workspace 1, got %d", len(listAll))
	}

	// List active only
	activeFilter := true
	listActive, err := svc.ListRules(1, &activeFilter, "")
	if err != nil {
		t.Fatalf("failed to list active rules: %v", err)
	}
	if len(listActive) != 1 || listActive[0].ID != created.ID {
		t.Fatalf("expected 1 active rule (created.ID), got %d", len(listActive))
	}

	// DeleteRule
	err = svc.DeleteRule(created.ID)
	if err != nil {
		t.Fatalf("failed to delete rule: %v", err)
	}

	_, err = svc.GetRule(created.ID)
	if !errors.Is(err, ErrAutomationRuleNotFound) {
		t.Fatalf("expected ErrAutomationRuleNotFound for deleted rule, got: %v", err)
	}
}

func TestAutomationService_ConditionEvaluator(t *testing.T) {
	projID := 10
	task := &model.Task{
		ID:          1,
		UserID:      "alice",
		Title:       "Critical Security Bug in Auth",
		Description: "Reproduce via OAuth callback flow",
		Status:      model.TaskStatusTodo,
		Priority:    "urgent",
		ProjectID:   &projID,
		Tags:        []string{"security", "backend", "auth"},
	}

	// 1. Nil task
	nilDetail := evaluateCondition(nil, model.AutomationCondition{
		Field:    model.FieldStatus,
		Operator: model.OpEquals,
		Value:    model.TaskStatusTodo,
	})
	if nilDetail.Matched {
		t.Fatalf("expected nil task to never match condition")
	}

	// 2. Empty conditions slice matches any task
	matched, details := evaluateAllConditions(task, nil)
	if !matched || len(details) != 0 {
		t.Fatalf("expected empty conditions to match with zero details")
	}

	// 3. Status equals & not equals
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldStatus, Operator: model.OpEquals, Value: model.TaskStatusTodo}); !d.Matched {
		t.Fatalf("expected status equals todo to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldStatus, Operator: model.OpEquals, Value: model.TaskStatusCompleted}); d.Matched {
		t.Fatalf("expected status equals completed to not match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldStatus, Operator: model.OpNotEquals, Value: model.TaskStatusCompleted}); !d.Matched {
		t.Fatalf("expected status not equals completed to match")
	}

	// 4. Status in
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldStatus, Operator: model.OpIn, Values: []string{model.TaskStatusInProgress, model.TaskStatusTodo}}); !d.Matched {
		t.Fatalf("expected status in [in_progress, todo] to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldStatus, Operator: model.OpIn, Values: []string{model.TaskStatusCompleted, "blocked"}}); d.Matched {
		t.Fatalf("expected status in [completed, blocked] to not match")
	}

	// 5. Priority equals & in
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldPriority, Operator: model.OpEquals, Value: "urgent"}); !d.Matched {
		t.Fatalf("expected priority equals urgent to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldPriority, Operator: model.OpIn, Values: []string{"low", "medium"}}); d.Matched {
		t.Fatalf("expected priority in [low, medium] to not match")
	}

	// 6. Title contains (case insensitive)
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTitle, Operator: model.OpContains, Value: "security"}); !d.Matched {
		t.Fatalf("expected title contains 'security' to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTitle, Operator: model.OpContains, Value: "SECURITY"}); !d.Matched {
		t.Fatalf("expected case-insensitive title match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTitle, Operator: model.OpContains, Value: "nonexistent"}); d.Matched {
		t.Fatalf("expected non-matching title to fail")
	}

	// 7. Description contains
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldDescription, Operator: model.OpContains, Value: "oauth"}); !d.Matched {
		t.Fatalf("expected description contains 'oauth' to match")
	}

	// 8. Assignee equals & not equals
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldAssignee, Operator: model.OpEquals, Value: "alice"}); !d.Matched {
		t.Fatalf("expected assignee equals alice to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldAssignee, Operator: model.OpNotEquals, Value: "bob"}); !d.Matched {
		t.Fatalf("expected assignee not equals bob to match")
	}

	// 9. ProjectID equals
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldProjectID, Operator: model.OpEquals, Value: "10"}); !d.Matched {
		t.Fatalf("expected project_id equals '10' to match")
	}

	// 10. Tags equals, not equals, contains, in
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTags, Operator: model.OpEquals, Value: "auth"}); !d.Matched {
		t.Fatalf("expected tags equals 'auth' to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTags, Operator: model.OpNotEquals, Value: "frontend"}); !d.Matched {
		t.Fatalf("expected tags not equals 'frontend' to match")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTags, Operator: model.OpContains, Value: "sec"}); !d.Matched {
		t.Fatalf("expected tags contains 'sec' to match 'security'")
	}
	if d := evaluateCondition(task, model.AutomationCondition{Field: model.FieldTags, Operator: model.OpIn, Values: []string{"devops", "auth"}}); !d.Matched {
		t.Fatalf("expected tags in ['devops', 'auth'] to match")
	}

	// 11. Compound conditions
	allConditions := []model.AutomationCondition{
		{Field: model.FieldPriority, Operator: model.OpEquals, Value: "urgent"},
		{Field: model.FieldTitle, Operator: model.OpContains, Value: "auth"},
		{Field: model.FieldTags, Operator: model.OpEquals, Value: "security"},
	}
	compMatched, compDetails := evaluateAllConditions(task, allConditions)
	if !compMatched || len(compDetails) != 3 {
		t.Fatalf("expected compound conditions to match, got %v (%d details)", compMatched, len(compDetails))
	}
}

func TestAutomationService_ActionExecutorAndDispatch(t *testing.T) {
	svc, taskRepo, subtaskRepo, _ := setupTestAutomationSvc(t)

	now := time.Now()
	testTask := &model.Task{
		UserID:      "initial_owner",
		Title:       "Production Payment Crash",
		Description: "Failure in Stripe webhook gateway",
		Status:      model.TaskStatusTodo,
		Priority:    "low",
		DueDate:     now.Add(24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
		Tags:        []string{"finance", "webhook"},
	}
	if err := taskRepo.Create(testTask); err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	active := true
	rule, err := svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Auto Triage Payment Issues",
		Description: "Escalate payment tasks automatically",
		TriggerType: model.TriggerTaskCreated,
		Conditions: []model.AutomationCondition{
			{Field: model.FieldTitle, Operator: model.OpContains, Value: "Payment"},
			{Field: model.FieldStatus, Operator: model.OpEquals, Value: model.TaskStatusTodo},
		},
		Actions: []model.AutomationAction{
			{Type: model.ActionSetPriority, TargetValue: "urgent"},
			{Type: model.ActionSetStatus, TargetValue: model.TaskStatusInProgress},
			{Type: model.ActionAssignUser, TargetValue: "lead_oncall"},
			{Type: model.ActionAddTag, TargetValue: "p0"},
			{Type: model.ActionCreateSubtask, TargetValue: "Inspect payment provider logs"},
		},
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("failed to create automation rule: %v", err)
	}

	// Dispatch event matching the rule
	event := model.TriggerEvent{
		WorkspaceID: 1,
		UserID:      "ci_webhook",
		EventType:   model.TriggerTaskCreated,
		TaskID:      testTask.ID,
		TaskAfter:   testTask,
		Timestamp:   time.Now(),
	}

	logs, err := svc.DispatchEvent(event)
	if err != nil {
		t.Fatalf("failed to dispatch trigger event: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("expected 1 execution log, got %d", len(logs))
	}
	if logs[0].Status != model.ExecutionStatusSuccess {
		t.Fatalf("expected execution status 'success', got '%s'", logs[0].Status)
	}
	if len(logs[0].ActionsExecuted) != 5 {
		t.Fatalf("expected 5 actions executed, got %d", len(logs[0].ActionsExecuted))
	}

	// Verify task in database has updated priority, status, and assignee
	persisted, err := taskRepo.GetByID(testTask.ID)
	if err != nil {
		t.Fatalf("failed to get updated task: %v", err)
	}
	if persisted.Priority != "urgent" {
		t.Fatalf("expected priority 'urgent', got '%s'", persisted.Priority)
	}
	if persisted.Status != model.TaskStatusInProgress {
		t.Fatalf("expected status 'in_progress', got '%s'", persisted.Status)
	}
	if persisted.UserID != "lead_oncall" {
		t.Fatalf("expected assignee 'lead_oncall', got '%s'", persisted.UserID)
	}

	// Verify subtask created in DB
	subtasks, err := subtaskRepo.GetByTaskID(testTask.ID)
	if err != nil {
		t.Fatalf("failed to get subtasks: %v", err)
	}
	if len(subtasks) != 1 {
		t.Fatalf("expected 1 subtask created, got %d", len(subtasks))
	}
	if subtasks[0].Title != "Inspect payment provider logs" {
		t.Fatalf("unexpected subtask title: '%s'", subtasks[0].Title)
	}

	// Verify metrics
	metrics, err := svc.GetMetrics(1)
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	if metrics.TotalExecutions != 1 || metrics.SuccessfulExecutions != 1 {
		t.Fatalf("expected 1 total and 1 successful execution, got %d and %d", metrics.TotalExecutions, metrics.SuccessfulExecutions)
	}
	if metrics.ActiveRules != 1 {
		t.Fatalf("expected 1 active rule, got %d", metrics.ActiveRules)
	}

	// Dispatch an event where conditions do NOT match
	unmatchedTask := &model.Task{
		UserID:      "dev_1",
		Title:       "Update Documentation",
		Status:      model.TaskStatusCompleted,
		Priority:    "low",
		DueDate:     now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	unmatchedEvent := model.TriggerEvent{
		WorkspaceID: 1,
		UserID:      "user_1",
		EventType:   model.TriggerTaskCreated,
		TaskID:      2,
		TaskAfter:   unmatchedTask,
		Timestamp:   time.Now(),
	}
	unmatchedLogs, err := svc.DispatchEvent(unmatchedEvent)
	if err != nil {
		t.Fatalf("failed to dispatch unmatched event: %v", err)
	}
	if len(unmatchedLogs) != 1 {
		t.Fatalf("expected 1 log for unmatched rule, got %d", len(unmatchedLogs))
	}
	if unmatchedLogs[0].Status != model.ExecutionStatusSkipped {
		t.Fatalf("expected execution status 'skipped', got '%s'", unmatchedLogs[0].Status)
	}

	// Verify log filtering by workspace
	allLogs, err := svc.ListLogs(model.AutomationLogFilter{
		WorkspaceID: 1,
		RuleID:      &rule.ID,
	})
	if err != nil {
		t.Fatalf("failed to list logs: %v", err)
	}
	if len(allLogs) != 2 {
		t.Fatalf("expected 2 logs recorded, got %d", len(allLogs))
	}
}

func TestAutomationService_DryRun(t *testing.T) {
	svc, _, _, _ := setupTestAutomationSvc(t)

	active := true
	rule, err := svc.CreateRule("user_1", model.CreateAutomationRuleRequest{
		WorkspaceID: 1,
		Name:        "Simulated Escalation",
		TriggerType: model.TriggerTaskCreated,
		Conditions: []model.AutomationCondition{
			{Field: model.FieldPriority, Operator: model.OpEquals, Value: "urgent"},
		},
		Actions: []model.AutomationAction{
			{Type: model.ActionSetStatus, TargetValue: model.TaskStatusInProgress},
			{Type: model.ActionAddTag, TargetValue: "simulated"},
		},
		IsActive: &active,
	})
	if err != nil {
		t.Fatalf("failed to create rule: %v", err)
	}

	// Dry run with matching mock task
	res, err := svc.DryRunRule(rule.ID, model.AutomationDryRunRequest{
		MockTask: &model.Task{
			Title:    "Simulated Incident",
			Priority: "urgent",
			Status:   model.TaskStatusTodo,
		},
	})
	if err != nil {
		t.Fatalf("failed to dry run rule: %v", err)
	}
	if !res.ConditionsMet {
		t.Fatalf("expected dry-run conditions to be met")
	}
	if len(res.ActionsPreview) != 2 {
		t.Fatalf("expected 2 action previews, got %d", len(res.ActionsPreview))
	}
	if res.RuleName != "Simulated Escalation" {
		t.Fatalf("unexpected rule name in result: '%s'", res.RuleName)
	}
}
