package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTagSvc(t *testing.T) (TagService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	tagRepo := repository.NewTagRepository(sqliteDB)

	tagSvc := NewTagService(tagRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)

	return tagSvc, taskSvc
}

func TestTagService_CRUD(t *testing.T) {
	tagSvc, _ := setupTestTagSvc(t)
	userID := "user-123"

	// 1. Validation error on empty name
	_, err := tagSvc.CreateTag(userID, "   ", "#10b981")
	if err != ErrTagNameRequired {
		t.Fatalf("expected ErrTagNameRequired, got %v", err)
	}

	// 2. Create tag
	tag, err := tagSvc.CreateTag(userID, "Urgent", "#ef4444")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}
	if tag.ID == 0 {
		t.Errorf("expected positive tag ID")
	}
	if tag.Name != "Urgent" {
		t.Errorf("expected name Urgent, got %s", tag.Name)
	}

	// 3. Get tag
	got, err := tagSvc.GetTag(tag.ID, userID)
	if err != nil {
		t.Fatalf("failed to get tag: %v", err)
	}
	if got.ID != tag.ID {
		t.Errorf("expected ID %d, got %d", tag.ID, got.ID)
	}

	// 4. Unauthorized access
	_, err = tagSvc.GetTag(tag.ID, "user-999")
	if err != ErrUnauthorizedTag {
		t.Fatalf("expected ErrUnauthorizedTag, got %v", err)
	}

	// 5. Update tag
	updated, err := tagSvc.UpdateTag(tag.ID, userID, "Very Urgent", "#b91c1c")
	if err != nil {
		t.Fatalf("failed to update tag: %v", err)
	}
	if updated.Name != "Very Urgent" {
		t.Errorf("expected name 'Very Urgent', got %s", updated.Name)
	}

	// 6. List user tags
	list, err := tagSvc.GetUserTags(userID)
	if err != nil {
		t.Fatalf("failed to list user tags: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("expected 1 tag, got %d", len(list))
	}

	// 7. Delete tag
	err = tagSvc.DeleteTag(tag.ID, userID)
	if err != nil {
		t.Fatalf("failed to delete tag: %v", err)
	}

	// 8. Confirm deleted
	_, err = tagSvc.GetTag(tag.ID, userID)
	if err != ErrTagNotFound {
		t.Fatalf("expected ErrTagNotFound, got %v", err)
	}
}

func TestTagService_TaskTagging(t *testing.T) {
	tagSvc, taskSvc := setupTestTagSvc(t)
	userID := "user-123"

	// Create task
	task, err := taskSvc.CreateTask(userID, "Feature X", "Details", time.Now().Add(48*time.Hour), 60, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Create tags
	tag1, err := tagSvc.CreateTag(userID, "Backend", "#3b82f6")
	if err != nil {
		t.Fatalf("failed to create tag1: %v", err)
	}
	tag2, err := tagSvc.CreateTag(userID, "API", "#10b981")
	if err != nil {
		t.Fatalf("failed to create tag2: %v", err)
	}

	// Attach tags to task
	err = tagSvc.AddTagToTask(task.ID, tag1.ID, userID)
	if err != nil {
		t.Fatalf("failed to attach tag1 to task: %v", err)
	}
	err = tagSvc.AddTagToTask(task.ID, tag2.ID, userID)
	if err != nil {
		t.Fatalf("failed to attach tag2 to task: %v", err)
	}

	// Unauthorized attachment check
	err = tagSvc.AddTagToTask(task.ID, tag1.ID, "other-user")
	if err != ErrUnauthorizedTag {
		t.Fatalf("expected ErrUnauthorizedTag, got %v", err)
	}

	// Get tags for task
	taskTags, err := tagSvc.GetTaskTags(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get tags for task: %v", err)
	}
	if len(taskTags) != 2 {
		t.Fatalf("expected 2 tags on task, got %d", len(taskTags))
	}

	// Detach a tag
	err = tagSvc.RemoveTagFromTask(task.ID, tag1.ID, userID)
	if err != nil {
		t.Fatalf("failed to remove tag1 from task: %v", err)
	}

	// Verify remaining tag
	taskTags, err = tagSvc.GetTaskTags(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get tags for task: %v", err)
	}
	if len(taskTags) != 1 || taskTags[0].ID != tag2.ID {
		t.Fatalf("expected only tag2 remaining, got %#v", taskTags)
	}
}
