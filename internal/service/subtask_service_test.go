package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestSubtaskSvc(t *testing.T) (SubtaskService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	subtaskRepo := repository.NewSubtaskRepository(sqliteDB)

	subtaskSvc := NewSubtaskService(subtaskRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)

	return subtaskSvc, taskSvc
}

func TestSubtaskService_CRUD(t *testing.T) {
	subtaskSvc, taskSvc := setupTestSubtaskSvc(t)
	userID := "user-123"

	// 1. Create parent task
	task, err := taskSvc.CreateTask(userID, "Parent Task", "Description", time.Now().Add(24*time.Hour), 30, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 2. Title required validation
	_, err = subtaskSvc.CreateSubtask(task.ID, userID, "   ")
	if err != ErrSubtaskTitleRequired {
		t.Fatalf("expected ErrSubtaskTitleRequired, got %v", err)
	}

	// 3. Unauthorized task access
	_, err = subtaskSvc.CreateSubtask(task.ID, "user-456", "Subtask 1")
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Create subtasks
	s1, err := subtaskSvc.CreateSubtask(task.ID, userID, "Step 1")
	if err != nil {
		t.Fatalf("failed to create subtask 1: %v", err)
	}
	if s1.Position != 0 || s1.Completed {
		t.Errorf("expected position 0, not completed")
	}

	s2, err := subtaskSvc.CreateSubtask(task.ID, userID, "Step 2")
	if err != nil {
		t.Fatalf("failed to create subtask 2: %v", err)
	}
	if s2.Position != 1 {
		t.Errorf("expected position 1, got %d", s2.Position)
	}

	// 5. List subtasks
	list, err := subtaskSvc.GetSubtasks(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get subtasks: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 subtasks, got %d", len(list))
	}

	// 6. Toggle subtask
	toggled, err := subtaskSvc.ToggleSubtask(s1.ID, userID, true)
	if err != nil {
		t.Fatalf("failed to toggle subtask: %v", err)
	}
	if !toggled.Completed {
		t.Errorf("expected subtask to be completed")
	}

	// 7. Update subtask title
	updated, err := subtaskSvc.UpdateSubtask(s1.ID, userID, "Step 1: Revised")
	if err != nil {
		t.Fatalf("failed to update subtask: %v", err)
	}
	if updated.Title != "Step 1: Revised" {
		t.Errorf("expected updated title, got %s", updated.Title)
	}

	// 8. Reorder subtasks
	err = subtaskSvc.ReorderSubtasks(task.ID, userID, []int{s2.ID, s1.ID})
	if err != nil {
		t.Fatalf("failed to reorder subtasks: %v", err)
	}
	reorderedList, err := subtaskSvc.GetSubtasks(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get reordered subtasks: %v", err)
	}
	if reorderedList[0].ID != s2.ID || reorderedList[1].ID != s1.ID {
		t.Errorf("expected order [s2, s1], got [%d, %d]", reorderedList[0].ID, reorderedList[1].ID)
	}

	// 9. Delete subtask
	err = subtaskSvc.DeleteSubtask(s1.ID, userID)
	if err != nil {
		t.Fatalf("failed to delete subtask: %v", err)
	}
	remaining, err := subtaskSvc.GetSubtasks(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get remaining subtasks: %v", err)
	}
	if len(remaining) != 1 {
		t.Fatalf("expected 1 remaining subtask, got %d", len(remaining))
	}
}
