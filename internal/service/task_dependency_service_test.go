package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTaskDependencySvc(t *testing.T) (TaskDependencyService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	depRepo := repository.NewTaskDependencyRepository(sqliteDB)

	depSvc := NewTaskDependencyService(depRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)

	return depSvc, taskSvc
}

func TestTaskDependencyService_AddAndCycle(t *testing.T) {
	depSvc, taskSvc := setupTestTaskDependencySvc(t)
	userID := "user-dag-1"
	otherUser := "user-dag-2"

	// 1. Create tasks
	t1, err := taskSvc.CreateTask(userID, "Architecture Design", "Plan database and APIs", time.Now().Add(24*time.Hour), 60, "high")
	if err != nil {
		t.Fatalf("failed to create task 1: %v", err)
	}
	t2, err := taskSvc.CreateTask(userID, "Database Implementation", "Write tables and repos", time.Now().Add(24*time.Hour), 120, "high")
	if err != nil {
		t.Fatalf("failed to create task 2: %v", err)
	}
	t3, err := taskSvc.CreateTask(userID, "Service Implementation", "Write business logic", time.Now().Add(24*time.Hour), 90, "medium")
	if err != nil {
		t.Fatalf("failed to create task 3: %v", err)
	}
	otherTask, err := taskSvc.CreateTask(otherUser, "Other User Task", "Should not be linkable", time.Now().Add(24*time.Hour), 30, "low")
	if err != nil {
		t.Fatalf("failed to create other user task: %v", err)
	}

	// 2. Self reference should fail
	_, err = depSvc.AddDependency(t1.ID, t1.ID, userID)
	if err != ErrDependencySelfRefDB {
		t.Fatalf("expected ErrDependencySelfRefDB, got %v", err)
	}

	// 3. Cross-user dependency should fail
	_, err = depSvc.AddDependency(t1.ID, otherTask.ID, userID)
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}
	_, err = depSvc.AddDependency(otherTask.ID, t1.ID, userID)
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Non-existent task should fail
	_, err = depSvc.AddDependency(t1.ID, 99999, userID)
	if err != ErrTaskNotFoundDB {
		t.Fatalf("expected ErrTaskNotFoundDB, got %v", err)
	}

	// 5. Valid dependency: t2 depends on t1 (t1 is prerequisite for t2)
	info, err := depSvc.AddDependency(t2.ID, t1.ID, userID)
	if err != nil {
		t.Fatalf("failed to add dependency: %v", err)
	}
	if info.TaskID != t2.ID || info.DependsOnTaskID != t1.ID || info.Task == nil || info.Task.Title != "Architecture Design" {
		t.Errorf("unexpected dependency info: %+v", info)
	}

	// 6. Duplicate dependency should fail with ErrDependencyExistsDB
	_, err = depSvc.AddDependency(t2.ID, t1.ID, userID)
	if err != ErrDependencyExistsDB {
		t.Fatalf("expected ErrDependencyExistsDB, got %v", err)
	}

	// 7. Direct cycle: t1 depends on t2 should fail
	_, err = depSvc.AddDependency(t1.ID, t2.ID, userID)
	if err != ErrDependencyCycleDB {
		t.Fatalf("expected ErrDependencyCycleDB for direct cycle, got %v", err)
	}

	// 8. Transitive cycle:
	// t3 depends on t2
	_, err = depSvc.AddDependency(t3.ID, t2.ID, userID)
	if err != nil {
		t.Fatalf("failed to add t3 -> t2 dependency: %v", err)
	}
	// Chain is: t3 -> t2 -> t1
	// Now try adding t1 depends on t3 (cycle!)
	_, err = depSvc.AddDependency(t1.ID, t3.ID, userID)
	if err != ErrDependencyCycleDB {
		t.Fatalf("expected ErrDependencyCycleDB for transitive cycle, got %v", err)
	}

	// 9. Verify GetDependencies for t3
	deps, err := depSvc.GetDependencies(t3.ID, userID)
	if err != nil {
		t.Fatalf("failed to get dependencies: %v", err)
	}
	if len(deps) != 1 || deps[0].DependsOnTaskID != t2.ID {
		t.Fatalf("expected 1 dependency (t2) for t3, got %+v", deps)
	}
}
