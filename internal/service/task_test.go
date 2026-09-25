package service

import "testing"

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
	page := GetTasks(nil, "", nil, nil, 100, 20, "id", false)
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
