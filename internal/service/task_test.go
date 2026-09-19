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
