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

func setupTestCustomFieldSvc(t *testing.T) (CustomFieldService, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	repo := repository.NewCustomFieldRepository(sqliteDB)
	svc := NewCustomFieldService(repo)
	return svc, sqliteDB
}

func TestCustomFieldService_FieldDefinitionLifecycle(t *testing.T) {
	svc, _ := setupTestCustomFieldSvc(t)

	// 1. Validation: Name required
	_, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "   ",
		Type:        model.CustomFieldTypeText,
	})
	if !errors.Is(err, ErrCustomFieldNameRequired) {
		t.Fatalf("expected ErrCustomFieldNameRequired, got: %v", err)
	}

	// 2. Validation: Invalid type
	_, err = svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Risk Level",
		Type:        "unsupported_type",
	})
	if !errors.Is(err, ErrInvalidCustomFieldType) {
		t.Fatalf("expected ErrInvalidCustomFieldType, got: %v", err)
	}

	// 3. Validation: Dropdown requires options
	_, err = svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Environment",
		Type:        model.CustomFieldTypeDropdown,
		Options:     []string{"   ", ""},
	})
	if !errors.Is(err, ErrDropdownOptionsRequired) {
		t.Fatalf("expected ErrDropdownOptionsRequired, got: %v", err)
	}

	// 4. Create text definition with auto-generated slug key
	textField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Release Version",
		Type:        model.CustomFieldTypeText,
		Description: "Target release version",
	})
	if err != nil {
		t.Fatalf("failed to create text field: %v", err)
	}
	if textField.FieldKey != "release_version" {
		t.Fatalf("expected auto-slugified key 'release_version', got: %s", textField.FieldKey)
	}

	// 5. Create dropdown definition with explicit key
	dropdownField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Environment",
		FieldKey:    "env",
		Type:        model.CustomFieldTypeDropdown,
		Options:     []string{"Production", "Staging", "Dev"},
	})
	if err != nil {
		t.Fatalf("failed to create dropdown field: %v", err)
	}
	if dropdownField.FieldKey != "env" {
		t.Fatalf("expected field key 'env', got: %s", dropdownField.FieldKey)
	}

	// 6. Duplicate key conflict check
	_, err = svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Environment Duplicate",
		FieldKey:    "env",
		Type:        model.CustomFieldTypeDropdown,
		Options:     []string{"Prod", "QA"},
	})
	if !errors.Is(err, ErrCustomFieldKeyTaken) {
		t.Fatalf("expected ErrCustomFieldKeyTaken, got: %v", err)
	}

	// 7. Get by ID and Get by Key
	byID, err := svc.GetFieldDefinition(dropdownField.ID)
	if err != nil {
		t.Fatalf("failed to get field by ID: %v", err)
	}
	if byID.Name != "Environment" || len(byID.Options) != 3 {
		t.Fatalf("unexpected fetched field: %+v", byID)
	}

	byKey, err := svc.GetFieldDefinitionByKey(1, "env")
	if err != nil {
		t.Fatalf("failed to get field by key: %v", err)
	}
	if byKey.ID != dropdownField.ID {
		t.Fatalf("expected ID %d, got: %d", dropdownField.ID, byKey.ID)
	}

	// 8. List definitions with filters and pagination
	summary, err := svc.ListFieldDefinitions(model.CustomFieldFilter{
		WorkspaceID: 1,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("failed to list field definitions: %v", err)
	}
	if summary.Total != 2 || len(summary.Definitions) != 2 {
		t.Fatalf("expected 2 definitions, got %d (total %d)", len(summary.Definitions), summary.Total)
	}

	// 9. Update definition
	newName := "Deployment Environment"
	newOptions := []string{"Production", "Staging", "Dev", "Sandbox"}
	updated, err := svc.UpdateFieldDefinition(dropdownField.ID, &model.UpdateCustomFieldRequest{
		Name:    &newName,
		Options: &newOptions,
	})
	if err != nil {
		t.Fatalf("failed to update field definition: %v", err)
	}
	if updated.Name != "Deployment Environment" || len(updated.Options) != 4 {
		t.Fatalf("unexpected updated field: %+v", updated)
	}

	// 10. Delete definition
	err = svc.DeleteFieldDefinition(textField.ID)
	if err != nil {
		t.Fatalf("failed to delete field definition: %v", err)
	}
	_, err = svc.GetFieldDefinition(textField.ID)
	if !errors.Is(err, ErrCustomFieldNotFound) {
		t.Fatalf("expected ErrCustomFieldNotFound after delete, got: %v", err)
	}
}

func TestCustomFieldService_TypeValidationAndValues(t *testing.T) {
	svc, sqliteDB := setupTestCustomFieldSvc(t)

	// Create test task
	taskRepo := repository.NewTaskRepository(sqliteDB)
	now := time.Now()
	testTask := &model.Task{
		UserID:    "user-dev",
		Title:     "Deploy Authentication Microservice",
		Status:    "in_progress",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := taskRepo.Create(testTask); err != nil {
		t.Fatalf("failed to create test task: %v", err)
	}
	taskID := int64(testTask.ID)

	// 1. Create fields of various types
	numField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Story Points",
		Type:        model.CustomFieldTypeNumber,
	})
	if err != nil {
		t.Fatalf("failed to create number field: %v", err)
	}

	dropdownField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Severity",
		Type:        model.CustomFieldTypeDropdown,
		Options:     []string{"Critical", "High", "Medium", "Low"},
	})
	if err != nil {
		t.Fatalf("failed to create dropdown field: %v", err)
	}

	boolField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Requires QA",
		Type:        model.CustomFieldTypeBoolean,
	})
	if err != nil {
		t.Fatalf("failed to create boolean field: %v", err)
	}

	dateField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Signoff Deadline",
		Type:        model.CustomFieldTypeDate,
	})
	if err != nil {
		t.Fatalf("failed to create date field: %v", err)
	}

	urlField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 1,
		Name:        "Figma Specs",
		Type:        model.CustomFieldTypeURL,
	})
	if err != nil {
		t.Fatalf("failed to create url field: %v", err)
	}

	// 2. Validate Number field
	_, err = svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: numField.ID,
		Value:   "not-a-number",
	})
	if !errors.Is(err, ErrInvalidNumberValue) {
		t.Fatalf("expected ErrInvalidNumberValue, got: %v", err)
	}

	numVal, err := svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: numField.ID,
		Value:   8.5,
	})
	if err != nil {
		t.Fatalf("failed to set number value: %v", err)
	}
	if numVal.ValueNumber == nil || *numVal.ValueNumber != 8.5 {
		t.Fatalf("expected number 8.5, got: %v", numVal.ValueNumber)
	}

	// 3. Validate Dropdown field
	_, err = svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: dropdownField.ID,
		Value:   "Urgent", // Not in options
	})
	if !errors.Is(err, ErrInvalidDropdownValue) {
		t.Fatalf("expected ErrInvalidDropdownValue, got: %v", err)
	}

	dropVal, err := svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: dropdownField.ID,
		Value:   "critical", // case-insensitive match -> Critical
	})
	if err != nil {
		t.Fatalf("failed to set dropdown value: %v", err)
	}
	if dropVal.ValueText == nil || *dropVal.ValueText != "Critical" {
		t.Fatalf("expected normalized 'Critical', got: %v", dropVal.ValueText)
	}

	// 4. Validate Boolean field
	_, err = svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: boolField.ID,
		Value:   "maybe",
	})
	if !errors.Is(err, ErrInvalidBooleanValue) {
		t.Fatalf("expected ErrInvalidBooleanValue, got: %v", err)
	}

	boolVal, err := svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: boolField.ID,
		Value:   "true",
	})
	if err != nil {
		t.Fatalf("failed to set boolean value: %v", err)
	}
	if boolVal.ValueBoolean == nil || !*boolVal.ValueBoolean {
		t.Fatalf("expected boolean true, got: %v", boolVal.ValueBoolean)
	}

	// 5. Validate Date field
	_, err = svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: dateField.ID,
		Value:   "invalid-date",
	})
	if !errors.Is(err, ErrInvalidDateValue) {
		t.Fatalf("expected ErrInvalidDateValue, got: %v", err)
	}

	dateVal, err := svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: dateField.ID,
		Value:   "2026-12-31",
	})
	if err != nil {
		t.Fatalf("failed to set date value: %v", err)
	}
	if dateVal.ValueDate == nil || dateVal.ValueDate.Year() != 2026 {
		t.Fatalf("expected year 2026, got: %v", dateVal.ValueDate)
	}

	// 6. Validate URL field
	_, err = svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: urlField.ID,
		Value:   "ht tp://bad-url",
	})
	if !errors.Is(err, ErrInvalidURLValue) {
		t.Fatalf("expected ErrInvalidURLValue, got: %v", err)
	}

	urlVal, err := svc.SetFieldValue(taskID, &model.SetCustomFieldValueRequest{
		FieldID: urlField.ID,
		Value:   "https://figma.com/file/12345/AppDesign",
	})
	if err != nil {
		t.Fatalf("failed to set url value: %v", err)
	}
	if urlVal.ValueText == nil || *urlVal.ValueText != "https://figma.com/file/12345/AppDesign" {
		t.Fatalf("expected figma url, got: %v", urlVal.ValueText)
	}

	// 7. Get enriched custom fields for task
	taskFields, err := svc.GetTaskCustomFields(taskID)
	if err != nil {
		t.Fatalf("failed to get task custom fields: %v", err)
	}
	if len(taskFields) != 5 {
		t.Fatalf("expected 5 enriched custom fields, got: %d", len(taskFields))
	}

	// 8. Delete a custom field value from task
	err = svc.DeleteFieldValue(taskID, numField.ID)
	if err != nil {
		t.Fatalf("failed to delete field value: %v", err)
	}
	taskFieldsAfterDelete, err := svc.GetTaskCustomFields(taskID)
	if err != nil {
		t.Fatalf("failed to get task custom fields after delete: %v", err)
	}
	if len(taskFieldsAfterDelete) != 4 {
		t.Fatalf("expected 4 fields after delete, got: %d", len(taskFieldsAfterDelete))
	}
}

func TestCustomFieldService_BulkAndSearch(t *testing.T) {
	svc, sqliteDB := setupTestCustomFieldSvc(t)

	taskRepo := repository.NewTaskRepository(sqliteDB)
	now := time.Now()
	t1 := &model.Task{UserID: "u1", Title: "Task 1", CreatedAt: now, UpdatedAt: now}
	t2 := &model.Task{UserID: "u1", Title: "Task 2", CreatedAt: now, UpdatedAt: now}
	_ = taskRepo.Create(t1)
	_ = taskRepo.Create(t2)

	// Create custom field
	envField, err := svc.CreateFieldDefinition(&model.CreateCustomFieldRequest{
		WorkspaceID: 10,
		Name:        "Target Environment",
		FieldKey:    "target_env",
		Type:        model.CustomFieldTypeDropdown,
		Options:     []string{"Production", "Staging", "Local"},
	})
	if err != nil {
		t.Fatalf("failed to create field: %v", err)
	}

	// Bulk assign custom fields
	bulkReq := &model.BulkSetCustomFieldRequest{
		Assignments: []model.BulkSetCustomFieldItem{
			{
				TaskID: int64(t1.ID),
				Fields: []model.SetCustomFieldValueRequest{
					{FieldID: envField.ID, Value: "Production"},
				},
			},
			{
				TaskID: int64(t2.ID),
				Fields: []model.SetCustomFieldValueRequest{
					{FieldID: envField.ID, Value: "Staging"},
				},
			},
		},
	}
	if err := svc.BulkSetCustomFields(bulkReq); err != nil {
		t.Fatalf("failed to execute bulk set: %v", err)
	}

	// Search tasks by custom field value "Production"
	matchingTasks, total, err := svc.FindTasksByFieldValue(model.CustomFieldValueFilter{
		WorkspaceID: 10,
		FieldID:     envField.ID,
		QueryValue:  "Production",
	})
	if err != nil {
		t.Fatalf("failed to search tasks by custom field: %v", err)
	}
	if total != 1 || len(matchingTasks) != 1 || matchingTasks[0] != int64(t1.ID) {
		t.Fatalf("expected task 1 only in search results, got: %+v (total %d)", matchingTasks, total)
	}
}
