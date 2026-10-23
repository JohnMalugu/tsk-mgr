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

func setupTestViewSvc(t *testing.T) (SavedViewService, repository.TaskRepository, *sql.DB) {
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

	viewRepo := repository.NewSavedViewRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewSavedViewService(viewRepo, taskRepo)
	return svc, taskRepo, sqliteDB
}

func TestSavedViewService_Validation(t *testing.T) {
	svc, _, _ := setupTestViewSvc(t)

	// Workspace ID required
	_, err := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 0,
		Name:        "Sprint Backlog",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got: %v", err)
	}

	// Empty name
	_, err = svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "   ",
	})
	if !errors.Is(err, ErrInvalidViewName) {
		t.Fatalf("expected ErrInvalidViewName, got: %v", err)
	}

	// Name too long
	_, err = svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        strings.Repeat("a", 101),
	})
	if !errors.Is(err, ErrInvalidViewName) {
		t.Fatalf("expected ErrInvalidViewName, got: %v", err)
	}

	// Invalid SortBy
	_, err = svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Valid Name",
		SortBy:      "invalid_field",
	})
	if !errors.Is(err, ErrInvalidSortField) {
		t.Fatalf("expected ErrInvalidSortField, got: %v", err)
	}

	// Invalid SortOrder
	_, err = svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Valid Name",
		SortOrder:   "sideways",
	})
	if !errors.Is(err, ErrInvalidSortOrder) {
		t.Fatalf("expected ErrInvalidSortOrder, got: %v", err)
	}

	// Invalid GroupBy
	_, err = svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Valid Name",
		GroupBy:     "random_group",
	})
	if !errors.Is(err, ErrInvalidGroupBy) {
		t.Fatalf("expected ErrInvalidGroupBy, got: %v", err)
	}
}

func TestSavedViewService_CRUD(t *testing.T) {
	svc, _, _ := setupTestViewSvc(t)

	isShared := false
	isPinned := true
	pos := 1

	// Create
	created, err := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "High Priority Bugs",
		Description: "Filtered bugs needing immediate attention",
		Icon:        "fire",
		Color:       "#EF4444",
		FilterCriteria: &model.ViewFilterCriteria{
			Status:   []string{"todo", "in_progress"},
			Priority: []string{"urgent", "high"},
			Search:   "bug",
		},
		SortBy:    model.SortByPriority,
		SortOrder: model.SortOrderDesc,
		GroupBy:   model.GroupByStatus,
		IsShared:  &isShared,
		IsPinned:  &isPinned,
		Position:  &pos,
	})
	if err != nil {
		t.Fatalf("unexpected error creating view: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected non-zero ID")
	}
	if created.Name != "High Priority Bugs" {
		t.Fatalf("expected name 'High Priority Bugs', got '%s'", created.Name)
	}
	if !created.IsPinned {
		t.Fatalf("expected view to be pinned")
	}

	// Get by owner
	fetched, err := svc.GetView("user_1", created.ID)
	if err != nil {
		t.Fatalf("failed to get view: %v", err)
	}
	if fetched.ID != created.ID {
		t.Fatalf("expected ID %d, got %d", created.ID, fetched.ID)
	}

	// Get by other user on private view -> Unauthorized
	_, err = svc.GetView("user_2", created.ID)
	if !errors.Is(err, ErrUnauthorizedViewAccess) {
		t.Fatalf("expected ErrUnauthorizedViewAccess, got: %v", err)
	}

	// Update view
	newName := "Urgent Bugs Only"
	newColor := "#DC2626"
	updated, err := svc.UpdateView("user_1", created.ID, model.UpdateViewRequest{
		Name:  &newName,
		Color: &newColor,
	})
	if err != nil {
		t.Fatalf("failed to update view: %v", err)
	}
	if updated.Name != "Urgent Bugs Only" {
		t.Fatalf("expected updated name 'Urgent Bugs Only', got '%s'", updated.Name)
	}
	if updated.Color != "#DC2626" {
		t.Fatalf("expected updated color '#DC2626', got '%s'", updated.Color)
	}

	// Non-owner cannot update
	_, err = svc.UpdateView("user_2", created.ID, model.UpdateViewRequest{
		Name: &newName,
	})
	if !errors.Is(err, ErrUnauthorizedViewAccess) {
		t.Fatalf("expected ErrUnauthorizedViewAccess on update, got: %v", err)
	}

	// Non-owner cannot delete
	err = svc.DeleteView("user_2", created.ID)
	if !errors.Is(err, ErrUnauthorizedViewAccess) {
		t.Fatalf("expected ErrUnauthorizedViewAccess on delete, got: %v", err)
	}

	// Owner delete
	err = svc.DeleteView("user_1", created.ID)
	if err != nil {
		t.Fatalf("failed to delete view: %v", err)
	}

	// Get deleted view
	_, err = svc.GetView("user_1", created.ID)
	if !errors.Is(err, ErrViewNotFound) {
		t.Fatalf("expected ErrViewNotFound for deleted view, got: %v", err)
	}
}

func TestSavedViewService_ListAndPin(t *testing.T) {
	svc, _, _ := setupTestViewSvc(t)

	pinnedTrue := true
	pinnedFalse := false
	sharedTrue := true

	v1, _ := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "View A",
		IsPinned:    &pinnedTrue,
	})
	v2, _ := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "View B",
		IsPinned:    &pinnedFalse,
	})
	// Shared view created by user_1 in workspace 1
	v3, _ := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Shared View",
		IsShared:    &sharedTrue,
		IsPinned:    &pinnedFalse,
	})

	// List for user_1
	list1, err := svc.ListViews("user_1", 1, false)
	if err != nil {
		t.Fatalf("failed to list views: %v", err)
	}
	if len(list1) != 3 {
		t.Fatalf("expected 3 views for user_1, got %d", len(list1))
	}

	// List only pinned for user_1
	listPinned, err := svc.ListViews("user_1", 1, true)
	if err != nil {
		t.Fatalf("failed to list pinned views: %v", err)
	}
	if len(listPinned) != 1 || listPinned[0].ID != v1.ID {
		t.Fatalf("expected 1 pinned view (v1), got %d", len(listPinned))
	}

	// List for user_2 (should only see shared view)
	list2, err := svc.ListViews("user_2", 1, false)
	if err != nil {
		t.Fatalf("failed to list views for user_2: %v", err)
	}
	if len(list2) != 1 || list2[0].ID != v3.ID {
		t.Fatalf("expected 1 shared view for user_2, got %d", len(list2))
	}

	// Toggle pin on v2
	newPin, err := svc.TogglePinView("user_1", v2.ID)
	if err != nil {
		t.Fatalf("failed to toggle pin on v2: %v", err)
	}
	if !newPin {
		t.Fatalf("expected v2 to become pinned (true)")
	}

	// Toggle back
	newPin, err = svc.TogglePinView("user_1", v2.ID)
	if err != nil {
		t.Fatalf("failed to toggle pin back on v2: %v", err)
	}
	if newPin {
		t.Fatalf("expected v2 to become unpinned (false)")
	}
}

func TestSavedViewService_Duplicate(t *testing.T) {
	svc, _, _ := setupTestViewSvc(t)

	original, err := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Original View",
		Description: "Original description",
		Icon:        "star",
		Color:       "#10B981",
		FilterCriteria: &model.ViewFilterCriteria{
			Status: []string{"todo"},
		},
		SortBy:    model.SortByTitle,
		SortOrder: model.SortOrderAsc,
		GroupBy:   model.GroupByPriority,
	})
	if err != nil {
		t.Fatalf("failed to create original view: %v", err)
	}

	// 1. Duplicate with default copy name
	dup1, err := svc.DuplicateView("user_1", original.ID, model.DuplicateViewRequest{})
	if err != nil {
		t.Fatalf("failed to duplicate view: %v", err)
	}
	if dup1.ID == original.ID {
		t.Fatalf("duplicated view must have a different ID")
	}
	if dup1.Name != "Original View (Copy)" {
		t.Fatalf("expected name 'Original View (Copy)', got '%s'", dup1.Name)
	}
	if dup1.WorkspaceID != 1 {
		t.Fatalf("expected same workspace ID 1, got %d", dup1.WorkspaceID)
	}
	if dup1.SortBy != model.SortByTitle {
		t.Fatalf("expected SortBy to match original, got '%s'", dup1.SortBy)
	}

	// 2. Duplicate with custom name and different target workspace
	targetWs := int64(2)
	dup2, err := svc.DuplicateView("user_1", original.ID, model.DuplicateViewRequest{
		NewName:           "Marketing Cloned View",
		TargetWorkspaceID: &targetWs,
	})
	if err != nil {
		t.Fatalf("failed to duplicate with custom overrides: %v", err)
	}
	if dup2.Name != "Marketing Cloned View" {
		t.Fatalf("expected 'Marketing Cloned View', got '%s'", dup2.Name)
	}
	if dup2.WorkspaceID != 2 {
		t.Fatalf("expected target workspace 2, got %d", dup2.WorkspaceID)
	}
}

func TestSavedViewService_Execution(t *testing.T) {
	svc, taskRepo, _ := setupTestViewSvc(t)

	now := time.Now()
	proj1 := 10
	proj2 := 20

	// Seed tasks for user_1
	t1 := &model.Task{
		UserID:      "user_1",
		ProjectID:   &proj1,
		Title:       "Fix login auth bug",
		Description: "Auth cookie header issue",
		Status:      "todo",
		Priority:    "urgent",
		DueDate:     now.Add(24 * time.Hour),
		CreatedAt:   now.Add(-2 * time.Hour),
		UpdatedAt:   now,
	}
	t2 := &model.Task{
		UserID:      "user_1",
		ProjectID:   &proj1,
		Title:       "Refactor API routes",
		Description: "Clean up router dispatch",
		Status:      "in_progress",
		Priority:    "high",
		DueDate:     now.Add(48 * time.Hour),
		CreatedAt:   now.Add(-1 * time.Hour),
		UpdatedAt:   now,
	}
	t3 := &model.Task{
		UserID:      "user_1",
		ProjectID:   &proj2,
		Title:       "Update docs",
		Description: "API specification update",
		Status:      "completed",
		Priority:    "low",
		DueDate:     time.Time{}, // no due date
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	_ = taskRepo.Create(t1)
	_ = taskRepo.Create(t2)
	_ = taskRepo.Create(t3)

	// Create view that filters for todo and in_progress, grouped by status
	view, err := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Active Tasks",
		FilterCriteria: &model.ViewFilterCriteria{
			Status: []string{"todo", "in_progress"},
		},
		SortBy:    model.SortByPriority,
		SortOrder: model.SortOrderDesc,
		GroupBy:   model.GroupByStatus,
	})
	if err != nil {
		t.Fatalf("failed to create view: %v", err)
	}

	// Execute view
	result, err := svc.ExecuteView("user_1", view.ID)
	if err != nil {
		t.Fatalf("failed to execute view: %v", err)
	}

	if result.TotalTasks != 2 {
		t.Fatalf("expected 2 active tasks, got %d", result.TotalTasks)
	}

	// Urgent should come before high in desc priority sort
	if result.Tasks[0].ID != t1.ID {
		t.Fatalf("expected first task to be t1 (urgent), got ID %d", result.Tasks[0].ID)
	}

	// Check grouping counts
	if len(result.GroupCounts) != 2 {
		t.Fatalf("expected 2 status groups, got %d", len(result.GroupCounts))
	}
	groupMap := make(map[string]int)
	for _, gc := range result.GroupCounts {
		groupMap[gc.GroupKey] = gc.Count
	}
	if groupMap["todo"] != 1 || groupMap["in_progress"] != 1 {
		t.Fatalf("unexpected group counts: %+v", groupMap)
	}

	// Check that view usage counter was incremented
	refreshed, _ := svc.GetView("user_1", view.ID)
	if refreshed.UsageCount != 1 {
		t.Fatalf("expected usage_count to be 1, got %d", refreshed.UsageCount)
	}
	if refreshed.LastUsedAt == nil {
		t.Fatalf("expected last_used_at to be non-nil")
	}

	// Test Search Filter
	searchView, err := svc.CreateView("user_1", model.CreateViewRequest{
		WorkspaceID: 1,
		Name:        "Bug Search View",
		FilterCriteria: &model.ViewFilterCriteria{
			Search: "auth",
		},
	})
	if err != nil {
		t.Fatalf("failed to create search view: %v", err)
	}
	searchRes, err := svc.ExecuteView("user_1", searchView.ID)
	if err != nil {
		t.Fatalf("failed to execute search view: %v", err)
	}
	if searchRes.TotalTasks != 1 || searchRes.Tasks[0].ID != t1.ID {
		t.Fatalf("expected search view to find 1 task (t1), got %d", searchRes.TotalTasks)
	}
}
