package service

import (
	"database/sql"
	"errors"
	"strings"
	"testing"

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
