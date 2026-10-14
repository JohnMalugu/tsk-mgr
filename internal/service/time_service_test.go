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

func TestTimeTrackingService_ManualAndReporting(t *testing.T) {
	timeSvc, taskSvc := setupTestTimeTrackingSvc(t)
	userID := "user-reporting-1"

	// 1. Create two tasks
	task1, err := taskSvc.CreateTask(userID, "Frontend Dashboard", "Build UI components", time.Now().Add(48*time.Hour), 60, "high")
	if err != nil {
		t.Fatalf("failed to create task 1: %v", err)
	}
	task2, err := taskSvc.CreateTask(userID, "Backend API", "Build endpoints", time.Now().Add(48*time.Hour), 30, "medium")
	if err != nil {
		t.Fatalf("failed to create task 2: %v", err)
	}

	now := time.Now().UTC()

	// 2. Test Invalid Time Range (ended before started)
	_, err = timeSvc.LogManualTime(task1.ID, userID, now, now.Add(-10*time.Minute), "Invalid")
	if err != ErrInvalidTimeRange {
		t.Fatalf("expected ErrInvalidTimeRange, got %v", err)
	}

	// 3. Test Unauthorized Manual Time
	_, err = timeSvc.LogManualTime(task1.ID, "user-other", now.Add(-60*time.Minute), now.Add(-30*time.Minute), "Other")
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Log valid manual entries
	// Task 1: 30 minutes (1800s)
	e1, err := timeSvc.LogManualTime(task1.ID, userID, now.Add(-60*time.Minute), now.Add(-30*time.Minute), "Wireframing")
	if err != nil {
		t.Fatalf("failed to log manual time 1: %v", err)
	}
	if e1.DurationSeconds != 1800 {
		t.Errorf("expected 1800 seconds, got %d", e1.DurationSeconds)
	}

	// Task 1: 20 minutes (1200s)
	e2, err := timeSvc.LogManualTime(task1.ID, userID, now.Add(-30*time.Minute), now.Add(-10*time.Minute), "Component layout")
	if err != nil {
		t.Fatalf("failed to log manual time 2: %v", err)
	}
	if e2.DurationSeconds != 1200 {
		t.Errorf("expected 1200 seconds, got %d", e2.DurationSeconds)
	}

	// Task 2: 10 minutes (600s)
	_, err = timeSvc.LogManualTime(task2.ID, userID, now.Add(-10*time.Minute), now, "Route setup")
	if err != nil {
		t.Fatalf("failed to log manual time 3: %v", err)
	}

	// 5. Test GetTaskTimeSummary
	summary1, err := timeSvc.GetTaskTimeSummary(task1.ID, userID)
	if err != nil {
		t.Fatalf("failed to get task 1 time summary: %v", err)
	}
	// Estimated: 60 min * 60 = 3600s, Actual: 1800 + 1200 = 3000s, Variance: 3000 - 3600 = -600s
	if summary1.EstimatedSeconds != 3600 {
		t.Errorf("expected EstimatedSeconds 3600, got %d", summary1.EstimatedSeconds)
	}
	if summary1.ActualSeconds != 3000 {
		t.Errorf("expected ActualSeconds 3000, got %d", summary1.ActualSeconds)
	}
	if summary1.VarianceSeconds != -600 {
		t.Errorf("expected VarianceSeconds -600, got %d", summary1.VarianceSeconds)
	}

	// 6. Delete time entry
	// Unauthorized delete
	if err := timeSvc.DeleteTimeEntry(e2.ID, "user-other"); err != ErrUnauthorizedTimeEntry {
		t.Fatalf("expected ErrUnauthorizedTimeEntry, got %v", err)
	}
	// Non-existent delete
	if err := timeSvc.DeleteTimeEntry(99999, userID); err != ErrTimeEntryNotFoundDB {
		t.Fatalf("expected ErrTimeEntryNotFoundDB, got %v", err)
	}
	// Valid delete
	if err := timeSvc.DeleteTimeEntry(e2.ID, userID); err != nil {
		t.Fatalf("failed to delete time entry: %v", err)
	}

	// Verify updated summary after delete
	summary1Updated, err := timeSvc.GetTaskTimeSummary(task1.ID, userID)
	if err != nil {
		t.Fatalf("failed to get updated summary: %v", err)
	}
	if summary1Updated.ActualSeconds != 1800 {
		t.Errorf("expected ActualSeconds 1800 after deletion, got %d", summary1Updated.ActualSeconds)
	}

	// 7. GetTimeReport across all tasks
	report, err := timeSvc.GetTimeReport(userID, nil, nil)
	if err != nil {
		t.Fatalf("failed to get time report: %v", err)
	}
	// Task 1: 1800s, Task 2: 600s -> Total: 2400s
	if report.TotalSeconds != 2400 {
		t.Errorf("expected total 2400s, got %d", report.TotalSeconds)
	}
	if report.EntryCount != 2 {
		t.Errorf("expected 2 completed entries, got %d", report.EntryCount)
	}
	if len(report.ByTask) != 2 {
		t.Errorf("expected 2 tasks in breakdown, got %d", len(report.ByTask))
	}

	// 8. Filtered report (future date window should return 0)
	futureFrom := now.Add(24 * time.Hour)
	futureTo := now.Add(48 * time.Hour)
	emptyReport, err := timeSvc.GetTimeReport(userID, &futureFrom, &futureTo)
	if err != nil {
		t.Fatalf("failed to get filtered time report: %v", err)
	}
	if emptyReport.TotalSeconds != 0 || emptyReport.EntryCount != 0 {
		t.Errorf("expected empty report for future filter, got %+v", emptyReport)
	}
}
