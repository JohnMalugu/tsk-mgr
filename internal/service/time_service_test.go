package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTimeTrackingSvc(t *testing.T) (TimeTrackingService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	timeRepo := repository.NewTimeEntryRepository(sqliteDB)

	timeSvc := NewTimeTrackingService(timeRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)

	return timeSvc, taskSvc
}

func TestTimeTrackingService_TimerFlow(t *testing.T) {
	timeSvc, taskSvc := setupTestTimeTrackingSvc(t)
	userID := "user-timer-1"

	// 1. Create a task
	task, err := taskSvc.CreateTask(userID, "Feature Implementation", "Implement persistence", time.Now().Add(24*time.Hour), 60, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// 2. Unauthorized task
	_, err = timeSvc.StartTimer(task.ID, "user-other", "Starting work")
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 3. Non-existent task
	_, err = timeSvc.StartTimer(99999, userID, "Starting work")
	if err != ErrTaskNotFoundDB {
		t.Fatalf("expected ErrTaskNotFoundDB, got %v", err)
	}

	// 4. Start timer successfully
	entry, err := timeSvc.StartTimer(task.ID, userID, "Coding session 1")
	if err != nil {
		t.Fatalf("failed to start timer: %v", err)
	}
	if entry.ID == 0 || entry.TaskID != task.ID || entry.UserID != userID {
		t.Errorf("unexpected timer entry attributes: %+v", entry)
	}
	if entry.EndedAt != nil {
		t.Errorf("expected EndedAt to be nil for active timer")
	}

	// 5. Duplicate active timer should fail
	_, err = timeSvc.StartTimer(task.ID, userID, "Another timer")
	if err != ErrActiveTimerExists {
		t.Fatalf("expected ErrActiveTimerExists, got %v", err)
	}

	// 6. Get active timer
	active, err := timeSvc.GetActiveTimer(userID)
	if err != nil {
		t.Fatalf("failed to get active timer: %v", err)
	}
	if active == nil || active.ID != entry.ID {
		t.Fatalf("expected active timer with ID %d, got %+v", entry.ID, active)
	}

	// 7. Get active timer for user without active timer
	noActive, err := timeSvc.GetActiveTimer("user-other")
	if err != nil {
		t.Fatalf("unexpected error getting active timer for other user: %v", err)
	}
	if noActive != nil {
		t.Fatalf("expected nil active timer for user-other, got %+v", noActive)
	}

	// 8. Stop timer for user with no active timer
	_, err = timeSvc.StopTimer("user-other")
	if err != ErrNoActiveTimerFound {
		t.Fatalf("expected ErrNoActiveTimerFound, got %v", err)
	}

	// 9. Stop active timer
	stopped, err := timeSvc.StopTimer(userID)
	if err != nil {
		t.Fatalf("failed to stop timer: %v", err)
	}
	if stopped.EndedAt == nil {
		t.Fatalf("expected EndedAt to be set after stopping timer")
	}
	if stopped.DurationSeconds < 0 {
		t.Fatalf("expected duration >= 0, got %d", stopped.DurationSeconds)
	}

	// 10. Active timer should now be nil
	activeAfterStop, err := timeSvc.GetActiveTimer(userID)
	if err != nil {
		t.Fatalf("error checking active timer after stop: %v", err)
	}
	if activeAfterStop != nil {
		t.Fatalf("expected nil active timer after stop, got %+v", activeAfterStop)
	}

	// 11. Task entries list should contain the stopped entry
	entries, err := timeSvc.GetTaskTimeEntries(task.ID, userID)
	if err != nil {
		t.Fatalf("failed to get task time entries: %v", err)
	}
	if len(entries) != 1 || entries[0].ID != entry.ID {
		t.Fatalf("expected 1 entry with ID %d, got %d entries", entry.ID, len(entries))
	}
}
