package service

import (
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

func TestGetTaskByIDReturnsCopy(t *testing.T) {
	task := GetTaskByID(1)
	if task == nil {
		t.Fatal("expected seeded task")
	}

	originalTitle := task.Title
	task.Title = "locally changed"

	storedTask := GetTaskByID(1)
	if storedTask == nil {
		t.Fatal("expected seeded task")
	}
	if storedTask.Title != originalTitle {
		t.Fatalf("expected stored task title %q, got %q", originalTitle, storedTask.Title)
	}
}

func TestGetTasksReturnsEmptyPageBeyondResults(t *testing.T) {
	page := GetTasks(nil, nil, "", nil, nil, nil, nil, 100, 20, "id", false)
	if page == nil {
		t.Fatal("expected an empty slice, got nil")
	}
	if len(page) != 0 {
		t.Fatalf("expected empty page, got %d tasks", len(page))
	}
}

func TestBulkSetTaskCompletionUpdatesAllTasks(t *testing.T) {
	ResetTasks()

	result, ok := BulkSetTaskCompletion([]int{1, 2}, true)
	if !ok {
		t.Fatal("expected bulk update to succeed")
	}
	if result.Updated != 2 || len(result.Tasks) != 2 {
		t.Fatalf("expected two updated tasks, got %#v", result)
	}
	for _, task := range result.Tasks {
		if !task.Completed {
			t.Fatalf("expected task %d to be completed", task.ID)
		}
	}
}

func TestBulkSetTaskCompletionDoesNotPartiallyUpdate(t *testing.T) {
	ResetTasks()

	if _, ok := BulkSetTaskCompletion([]int{1, 999}, true); ok {
		t.Fatal("expected bulk update to reject missing task")
	}
	if task := GetTaskByID(1); task == nil || task.Completed {
		t.Fatal("expected task 1 to remain unchanged")
	}
}

func TestBulkDeleteTasksIsAtomic(t *testing.T) {
	ResetTasks()
	if _, ok := BulkDeleteTasks([]int{1, 999}); ok {
		t.Fatal("expected deletion to reject a missing task")
	}
	if GetTaskByID(1) == nil {
		t.Fatal("expected existing task to remain after rejected deletion")
	}
	result, ok := BulkDeleteTasks([]int{1, 2})
	if !ok || result.Count != 2 || len(result.Deleted) != 2 {
		t.Fatalf("expected both tasks deleted, got %#v, ok=%v", result, ok)
	}
	if len(GetAllTasks()) != 0 {
		t.Fatal("expected no tasks to remain")
	}
}

func TestCreateTaskNormalizesTags(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "Tagged", Tags: []string{" home ", "Home", "work"}})
	if len(created.Tags) != 2 || created.Tags[0] != "home" || created.Tags[1] != "work" {
		t.Fatalf("expected trimmed, unique tags, got %#v", created.Tags)
	}
}

func TestCreateTaskDefaultsPriority(t *testing.T) {
	ResetTasks()
	created := CreateTask(model.Task{Title: "No priority"})
	if created.Priority != "medium" {
		t.Fatalf("expected medium priority by default, got %q", created.Priority)
	}
}

func TestUpdateTaskPreservesCreatedAt(t *testing.T) {
	ResetTasks()
	original := GetTaskByID(1)
	if original.CreatedAt.IsZero() || original.UpdatedAt.IsZero() {
		t.Fatal("expected seeded task timestamps to be initialized")
	}
	updated := UpdateTask(1, model.Task{Title: "Updated", DueDate: original.DueDate})
	if updated == nil {
		t.Fatal("expected task to update")
	}
	if !updated.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("expected createdAt %v to be preserved, got %v", original.CreatedAt, updated.CreatedAt)
	}
	if updated.UpdatedAt.IsZero() {
		t.Fatal("expected updatedAt to be set")
	}
}
