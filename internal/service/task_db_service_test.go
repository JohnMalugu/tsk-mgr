package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTaskDBSvc(t *testing.T) TaskDBService {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory db: %v", err)
	}

	// Create user for foreign key constraint
	userRepo := repository.NewUserRepository(sqliteDB)
	authSvc := NewAuthService(userRepo)
	_, err = authSvc.Register("tester", "tester@example.com", "pass123")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	return NewTaskDBService(taskRepo)
}

func TestTaskDBService_CRUD(t *testing.T) {
	svc := setupTestTaskDBSvc(t)
	userID := "user123"

	// 1. Create task
	task, err := svc.CreateTask(userID, "Test Task", "Description", time.Now().Add(24*time.Hour), 30, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if task.ID == 0 {
		t.Errorf("expected non-zero task ID")
	}
	if task.Title != "Test Task" {
		t.Errorf("expected title 'Test Task', got %s", task.Title)
	}

	// 2. Get task
	retrieved, err := svc.GetTask(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get task: %v", err)
	}
	if retrieved.ID != task.ID {
		t.Errorf("expected ID %d, got %d", task.ID, retrieved.ID)
	}

	// 3. Unauthorized access check
	_, err = svc.GetTask(task.ID, "other-user")
	if err != ErrUnauthorizedTask {
		t.Errorf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Update task
	updated, err := svc.UpdateTask(task.ID, userID, "Updated Title", "New Desc", task.DueDate, 45, false, "in_progress", "medium")
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected updated title, got %s", updated.Title)
	}

	// 5. Complete task
	completed, err := svc.CompleteTask(task.ID, userID, true)
	if err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}
	if !completed.Completed {
		t.Errorf("expected task to be completed")
	}

	// 6. List user tasks
	tasks, err := svc.GetUserTasks(userID)
	if err != nil {
		t.Fatalf("failed to list user tasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	// 7. Delete task
	err = svc.DeleteTask(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to delete task: %v", err)
	}

	// 8. Verify deletion
	_, err = svc.GetTask(task.ID, userID)
	if err != ErrTaskNotFoundDB {
		t.Errorf("expected ErrTaskNotFoundDB, got %v", err)
	}
}
