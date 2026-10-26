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

func setupTestReleaseSvc(t *testing.T) (ReleaseService, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-rel', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Platform WS', 'plat-ws-rel', 'user_1', datetime('now'), datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert test workspaces: %v", err)
	}

	relRepo := repository.NewReleaseRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewReleaseService(relRepo, taskRepo)
	return svc, taskRepo, sqliteDB
}

func TestReleaseService_Validation(t *testing.T) {
	svc, _, _ := setupTestReleaseSvc(t)

	// 1. Missing workspace ID
	_, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 0,
		Name:        "Release v1.0",
		Version:     "v1.0.0",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// 2. Empty release name
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "   ",
		Version:     "v1.0.0",
	})
	if !errors.Is(err, ErrReleaseNameRequired) {
		t.Fatalf("expected ErrReleaseNameRequired, got: %v", err)
	}

	// 3. Name exceeds 100 runes
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        strings.Repeat("r", 101),
		Version:     "v1.0.0",
	})
	if err == nil || !strings.Contains(err.Error(), "cannot exceed 100 characters") {
		t.Fatalf("expected name length error, got: %v", err)
	}

	// 4. Empty release version
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Q1 Release",
		Version:     "   ",
	})
	if !errors.Is(err, ErrInvalidReleaseVersion) {
		t.Fatalf("expected ErrInvalidReleaseVersion, got: %v", err)
	}

	// 5. Invalid release version characters
	_, err = svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Q1 Release",
		Version:     "v1.0.0@#$",
	})
	if !errors.Is(err, ErrInvalidReleaseVersion) {
		t.Fatalf("expected ErrInvalidReleaseVersion for special chars, got: %v", err)
	}

	// 6. Invalid status transitions
	if err := validateReleaseStatusTransition(model.ReleaseStatusPlanning, "unknown_status"); !errors.Is(err, ErrInvalidReleaseStatus) {
		t.Fatalf("expected ErrInvalidReleaseStatus, got: %v", err)
	}
	if err := validateReleaseStatusTransition(model.ReleaseStatusPlanning, model.ReleaseStatusReleased); !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition from planning to released, got: %v", err)
	}
	if err := validateReleaseStatusTransition(model.ReleaseStatusReleased, model.ReleaseStatusPlanning); !errors.Is(err, ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition from released to planning, got: %v", err)
	}

	// 7. Invalid category
	if _, err := validateReleaseCategory("nonexistent"); !errors.Is(err, ErrInvalidReleaseCategory) {
		t.Fatalf("expected ErrInvalidReleaseCategory, got: %v", err)
	}
}

func TestReleaseService_CRUD(t *testing.T) {
	svc, _, _ := setupTestReleaseSvc(t)

	targetDate := time.Now().Add(14 * 24 * time.Hour)
	created, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Spring Platform Release",
		Version:     "v2.5.0",
		Description: "Major architecture upgrades and UI improvements",
		TargetDate:  &targetDate,
	})
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if created.Status != model.ReleaseStatusPlanning {
		t.Fatalf("expected initial status to be planning, got '%s'", created.Status)
	}

	// GetRelease
	fetched, err := svc.GetRelease(created.ID)
	if err != nil {
		t.Fatalf("failed to get release: %v", err)
	}
	if fetched.Name != "Spring Platform Release" || fetched.Version != "v2.5.0" {
		t.Fatalf("mismatched release data: %+v", fetched)
	}

	// UpdateRelease
	newName := "Spring Platform Enterprise Release"
	newVer := "v2.5.1"
	updated, err := svc.UpdateRelease(created.ID, model.UpdateReleaseRequest{
		Name:    &newName,
		Version: &newVer,
	})
	if err != nil {
		t.Fatalf("failed to update release: %v", err)
	}
	if updated.Name != newName || updated.Version != newVer {
		t.Fatalf("updated release mismatch: %+v", updated)
	}

	// ListReleases
	list, err := svc.ListReleases(model.ReleaseFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("failed to list releases: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 release in list, got %d", len(list))
	}
	if list[0].ID != created.ID {
		t.Fatalf("expected summary to match created release ID")
	}

	// DeleteRelease
	err = svc.DeleteRelease(created.ID)
	if err != nil {
		t.Fatalf("failed to delete release: %v", err)
	}

	_, err = svc.GetRelease(created.ID)
	if !errors.Is(err, ErrReleaseNotFound) {
		t.Fatalf("expected ErrReleaseNotFound for deleted release, got: %v", err)
	}
}

func TestReleaseService_StatusLifecycle(t *testing.T) {
	svc, taskRepo, _ := setupTestReleaseSvc(t)

	now := time.Now()
	task := &model.Task{
		UserID:    "user_1",
		Title:     "Setup Auth Middleware",
		Status:    model.TaskStatusTodo,
		Priority:  "high",
		DueDate:   now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := taskRepo.Create(task); err != nil {
		t.Fatalf("failed to seed task: %v", err)
	}

	rel, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Sprint Release v1",
		Version:     "v1.0.0",
	})
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// 1. Advance planning -> active
	activeRel, err := svc.UpdateReleaseStatus(rel.ID, model.UpdateReleaseStatusRequest{
		Status: model.ReleaseStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to activate release: %v", err)
	}
	if activeRel.Status != model.ReleaseStatusActive {
		t.Fatalf("expected active status, got '%s'", activeRel.Status)
	}

	// 2. Add task while active
	inc := true
	_, err = svc.AddTasksToRelease("user_1", rel.ID, model.AddReleaseTasksRequest{
		Tasks: []model.AddReleaseTaskItem{
			{TaskID: task.ID, Category: model.ReleaseCategoryFeature, IncludeInChangelog: &inc},
		},
	})
	if err != nil {
		t.Fatalf("failed to attach task while active: %v", err)
	}

	// 3. Advance active -> code_freeze
	freezeRel, err := svc.UpdateReleaseStatus(rel.ID, model.UpdateReleaseStatusRequest{
		Status: model.ReleaseStatusFreeze,
	})
	if err != nil {
		t.Fatalf("failed to freeze release: %v", err)
	}
	if freezeRel.Status != model.ReleaseStatusFreeze {
		t.Fatalf("expected code_freeze status, got '%s'", freezeRel.Status)
	}

	// Adding tasks during code_freeze must fail
	_, err = svc.AddTasksToRelease("user_1", rel.ID, model.AddReleaseTasksRequest{
		Tasks: []model.AddReleaseTaskItem{
			{TaskID: task.ID, Category: model.ReleaseCategoryBugfix},
		},
	})
	if !errors.Is(err, ErrReleaseCodeFrozen) {
		t.Fatalf("expected ErrReleaseCodeFrozen, got: %v", err)
	}

	// 4. Advance code_freeze -> released
	releasedRel, err := svc.UpdateReleaseStatus(rel.ID, model.UpdateReleaseStatusRequest{
		Status: model.ReleaseStatusReleased,
	})
	if err != nil {
		t.Fatalf("failed to mark release as released: %v", err)
	}
	if releasedRel.Status != model.ReleaseStatusReleased {
		t.Fatalf("expected released status, got '%s'", releasedRel.Status)
	}
	if releasedRel.ReleasedAt == nil {
		t.Fatalf("expected non-nil ReleasedAt timestamp")
	}

	// Modifying released release metadata or tasks must fail
	newName := "Modified After Release"
	_, err = svc.UpdateRelease(rel.ID, model.UpdateReleaseRequest{Name: &newName})
	if !errors.Is(err, ErrReleaseAlreadyReleased) {
		t.Fatalf("expected ErrReleaseAlreadyReleased on update, got: %v", err)
	}
}

func TestReleaseService_TasksAndChangelog(t *testing.T) {
	svc, taskRepo, _ := setupTestReleaseSvc(t)

	now := time.Now()
	// Task 1: Completed feature
	task1 := &model.Task{
		UserID:    "user_1",
		Title:     "Add OAuth2 authentication provider",
		Status:    model.TaskStatusCompleted,
		Completed: true,
		Priority:  "medium",
		DueDate:   now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(task1)

	// Task 2: Completed bugfix
	task2 := &model.Task{
		UserID:    "user_1",
		Title:     "Fix database deadlock on concurrent timers",
		Status:    model.TaskStatusCompleted,
		Completed: true,
		Priority:  "high",
		DueDate:   now.Add(24 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(task2)

	// Task 3: In-progress security patch
	task3 := &model.Task{
		UserID:    "user_1",
		Title:     "Upgrade argon2id salt entropy parameters",
		Status:    model.TaskStatusInProgress,
		Completed: false,
		Priority:  "urgent",
		DueDate:   now.Add(48 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(task3)

	rel, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Release Autumn Gold",
		Version:     "v3.0.0",
		Description: "Autumn 2026 major feature release",
	})
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	// Attach tasks
	inc := true
	_, err = svc.AddTasksToRelease("user_1", rel.ID, model.AddReleaseTasksRequest{
		Tasks: []model.AddReleaseTaskItem{
			{TaskID: task1.ID, Category: model.ReleaseCategoryFeature, IncludeInChangelog: &inc, CustomNotes: "Supports Google & GitHub"},
			{TaskID: task2.ID, Category: model.ReleaseCategoryBugfix, IncludeInChangelog: &inc},
			{TaskID: task3.ID, Category: model.ReleaseCategorySecurity, IncludeInChangelog: &inc},
		},
	})
	if err != nil {
		t.Fatalf("failed to add tasks to release: %v", err)
	}

	// Update release item
	newCat := model.ReleaseCategoryImprovement
	customNote := "Patched SQLite mutex lock"
	updatedItem, err := svc.UpdateReleaseItem(rel.ID, task2.ID, model.UpdateReleaseItemRequest{
		Category:    &newCat,
		CustomNotes: &customNote,
	})
	if err != nil {
		t.Fatalf("failed to update release item: %v", err)
	}
	if updatedItem.Category != model.ReleaseCategoryImprovement {
		t.Fatalf("expected category to be updated to improvement, got '%s'", updatedItem.Category)
	}

	// List release items
	items, err := svc.ListReleaseItems(rel.ID, "")
	if err != nil {
		t.Fatalf("failed to list items: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}

	// Generate Markdown release notes
	mdNotes, err := svc.GenerateReleaseNotes(rel.ID, model.ReleaseNotesOptions{
		Format:             model.ReleaseNotesFormatMarkdown,
		IncludeUncompleted: true,
	})
	if err != nil {
		t.Fatalf("failed to generate md release notes: %v", err)
	}
	if mdNotes.TotalItems != 3 {
		t.Fatalf("expected 3 items in changelog, got %d", mdNotes.TotalItems)
	}
	if !strings.Contains(mdNotes.Content, "### Features") || !strings.Contains(mdNotes.Content, "### Improvements") || !strings.Contains(mdNotes.Content, "### Security Patches") {
		t.Fatalf("expected markdown to contain category headers, got:\n%s", mdNotes.Content)
	}

	// Generate JSON release notes excluding uncompleted
	jsonNotes, err := svc.GenerateReleaseNotes(rel.ID, model.ReleaseNotesOptions{
		Format:             model.ReleaseNotesFormatJSON,
		IncludeUncompleted: false,
	})
	if err != nil {
		t.Fatalf("failed to generate json release notes: %v", err)
	}
	if jsonNotes.TotalItems != 2 {
		t.Fatalf("expected 2 completed items in changelog, got %d", jsonNotes.TotalItems)
	}

	// Check metrics
	metrics, err := svc.GetReleaseMetrics(rel.ID)
	if err != nil {
		t.Fatalf("failed to get release metrics: %v", err)
	}
	if metrics.TotalTasks != 3 || metrics.CompletedTasks != 2 || metrics.InProgressTasks != 1 {
		t.Fatalf("unexpected metrics task counts: %+v", metrics)
	}
	if !metrics.HasBlockers || metrics.BlockerCount != 1 {
		t.Fatalf("expected 1 blocker due to incomplete urgent task, got: %d", metrics.BlockerCount)
	}
	if metrics.HealthStatus != model.ReleaseHealthAtRisk {
		t.Fatalf("expected health status to be at_risk, got '%s'", metrics.HealthStatus)
	}
}

func TestReleaseService_CloneAndRollover(t *testing.T) {
	svc, taskRepo, _ := setupTestReleaseSvc(t)

	now := time.Now()
	doneTask := &model.Task{
		UserID:    "user_1",
		Title:     "Done Task in v1",
		Status:    model.TaskStatusCompleted,
		Completed: true,
		DueDate:   now,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(doneTask)

	openTask := &model.Task{
		UserID:    "user_1",
		Title:     "Incomplete Task Rolled Over",
		Status:    model.TaskStatusInProgress,
		Completed: false,
		DueDate:   now.Add(48 * time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(openTask)

	targetDate := now.Add(7 * 24 * time.Hour)
	rel, err := svc.CreateRelease("user_1", model.CreateReleaseRequest{
		WorkspaceID: 1,
		Name:        "Release v1.0",
		Version:     "v1.0.0",
		TargetDate:  &targetDate,
	})
	if err != nil {
		t.Fatalf("failed to create release: %v", err)
	}

	_, err = svc.AddTasksToRelease("user_1", rel.ID, model.AddReleaseTasksRequest{
		Tasks: []model.AddReleaseTaskItem{
			{TaskID: doneTask.ID, Category: model.ReleaseCategoryFeature},
			{TaskID: openTask.ID, Category: model.ReleaseCategoryBugfix},
		},
	})
	if err != nil {
		t.Fatalf("failed to add tasks: %v", err)
	}

	// Clone with rollover
	cloned, err := svc.CloneRelease("user_1", rel.ID, model.CloneReleaseRequest{
		NewName:            "Release v1.1",
		NewVersion:         "v1.1.0",
		RolloverIncomplete: true,
	})
	if err != nil {
		t.Fatalf("failed to clone release: %v", err)
	}
	if cloned.ID == rel.ID {
		t.Fatalf("cloned release must have a distinct ID")
	}
	if cloned.Name != "Release v1.1" || cloned.Version != "v1.1.0" {
		t.Fatalf("unexpected cloned metadata: %+v", cloned)
	}

	// Verify cloned items only have openTask
	clonedItems, err := svc.ListReleaseItems(cloned.ID, "")
	if err != nil {
		t.Fatalf("failed to list cloned items: %v", err)
	}
	if len(clonedItems) != 1 {
		t.Fatalf("expected exactly 1 incomplete rolled over item, got %d", len(clonedItems))
	}
	if clonedItems[0].TaskID != openTask.ID {
		t.Fatalf("expected rolled over task to be openTask (%d), got %d", openTask.ID, clonedItems[0].TaskID)
	}
}

