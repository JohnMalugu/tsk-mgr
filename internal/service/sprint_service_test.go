package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestSprintSvc(t *testing.T) (SprintService, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	repo := repository.NewSprintRepository(sqliteDB)
	svc := NewSprintService(repo)
	return svc, sqliteDB
}

func TestSprintService_LifecycleAndEnforcement(t *testing.T) {
	svc, _ := setupTestSprintSvc(t)

	// 1. Validation: Name required
	_, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "   ",
	})
	if !errors.Is(err, ErrSprintNameRequired) {
		t.Fatalf("expected ErrSprintNameRequired, got: %v", err)
	}

	// 2. Create sprint 1 (planned)
	s1, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "Sprint 1",
		Goal:        "Deliver MVP milestone",
	})
	if err != nil {
		t.Fatalf("failed to create sprint 1: %v", err)
	}
	if s1.Status != model.SprintStatusPlanned {
		t.Fatalf("expected status planned, got: %v", s1.Status)
	}

	// 3. Create sprint 2 (planned) in same project
	s2, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 1,
		ProjectID:   1,
		Name:        "Sprint 2",
		Goal:        "Performance improvements",
	})
	if err != nil {
		t.Fatalf("failed to create sprint 2: %v", err)
	}

	// 4. Start sprint 1 -> active
	s1Started, err := svc.StartSprint(s1.ID)
	if err != nil {
		t.Fatalf("failed to start sprint 1: %v", err)
	}
	if s1Started.Status != model.SprintStatusActive {
		t.Fatalf("expected status active, got: %v", s1Started.Status)
	}
	if s1Started.StartDate == nil {
		t.Fatal("expected StartDate to be set upon starting sprint")
	}

	// 5. Enforce single active sprint: starting sprint 2 in same project should fail
	_, err = svc.StartSprint(s2.ID)
	if !errors.Is(err, ErrActiveSprintExists) {
		t.Fatalf("expected ErrActiveSprintExists, got: %v", err)
	}

	// 6. Starting already active sprint should fail
	_, err = svc.StartSprint(s1.ID)
	if !errors.Is(err, ErrSprintCannotStart) {
		t.Fatalf("expected ErrSprintCannotStart, got: %v", err)
	}

	// 7. Complete non-active sprint should fail
	_, err = svc.CompleteSprint(s2.ID)
	if !errors.Is(err, ErrSprintCannotComplete) {
		t.Fatalf("expected ErrSprintCannotComplete, got: %v", err)
	}

	// 8. Close non-completed sprint should fail
	_, err = svc.CloseSprint(s2.ID)
	if !errors.Is(err, ErrSprintCannotClose) {
		t.Fatalf("expected ErrSprintCannotClose, got: %v", err)
	}

	// 9. Complete active sprint 1 -> completed
	s1Completed, err := svc.CompleteSprint(s1.ID)
	if err != nil {
		t.Fatalf("failed to complete sprint 1: %v", err)
	}
	if s1Completed.Status != model.SprintStatusCompleted {
		t.Fatalf("expected status completed, got: %v", s1Completed.Status)
	}
	if s1Completed.EndDate == nil {
		t.Fatal("expected EndDate to be set upon completing sprint")
	}

	// 10. Close sprint 1 -> archived
	s1Closed, err := svc.CloseSprint(s1.ID)
	if err != nil {
		t.Fatalf("failed to close sprint 1: %v", err)
	}
	if s1Closed.Status != model.SprintStatusArchived {
		t.Fatalf("expected status archived, got: %v", s1Closed.Status)
	}

	// 11. Updating archived sprint should fail
	newName := "Renamed"
	_, err = svc.UpdateSprint(s1.ID, &model.UpdateSprintRequest{Name: &newName})
	if !errors.Is(err, ErrSprintClosed) {
		t.Fatalf("expected ErrSprintClosed, got: %v", err)
	}

	// 12. Now sprint 2 can be started since sprint 1 is no longer active
	s2Started, err := svc.StartSprint(s2.ID)
	if err != nil {
		t.Fatalf("expected sprint 2 to start successfully, got: %v", err)
	}
	if s2Started.Status != model.SprintStatusActive {
		t.Fatalf("expected status active, got: %v", s2Started.Status)
	}

	// 13. Cannot delete active sprint
	err = svc.DeleteSprint(s2.ID)
	if !errors.Is(err, ErrCannotDeleteActiveSprint) {
		t.Fatalf("expected ErrCannotDeleteActiveSprint, got: %v", err)
	}
}

func TestSprintService_ScopingAndMetrics(t *testing.T) {
	svc, sqliteDB := setupTestSprintSvc(t)

	// Create test tasks in DB
	now := time.Now()
	taskRepo := repository.NewTaskRepository(sqliteDB)

	projID := 10
	task1 := &model.Task{
		UserID:    "user-1",
		ProjectID: &projID,
		Title:     "API Endpoint",
		Status:    "completed",
		Completed: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := taskRepo.Create(task1); err != nil {
		t.Fatalf("failed to create task 1: %v", err)
	}

	task2 := &model.Task{
		UserID:    "user-1",
		ProjectID: &projID,
		Title:     "Unit Tests",
		Status:    "in_progress",
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := taskRepo.Create(task2); err != nil {
		t.Fatalf("failed to create task 2: %v", err)
	}

	task3 := &model.Task{
		UserID:    "user-1",
		ProjectID: &projID,
		Title:     "Documentation",
		Status:    "pending",
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := taskRepo.Create(task3); err != nil {
		t.Fatalf("failed to create task 3: %v", err)
	}

	// Create sprint
	sprint, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 1,
		ProjectID:   10,
		Name:        "Milestone Beta",
	})
	if err != nil {
		t.Fatalf("failed to create sprint: %v", err)
	}

	// Add tasks with story points
	err = svc.AddTaskToSprint(sprint.ID, &model.AddTaskToSprintRequest{
		TaskID:      int64(task1.ID),
		StoryPoints: 5,
	})
	if err != nil {
		t.Fatalf("failed to add task 1 to sprint: %v", err)
	}

	err = svc.AddTaskToSprint(sprint.ID, &model.AddTaskToSprintRequest{
		TaskID:      int64(task2.ID),
		StoryPoints: 3,
	})
	if err != nil {
		t.Fatalf("failed to add task 2 to sprint: %v", err)
	}

	err = svc.AddTaskToSprint(sprint.ID, &model.AddTaskToSprintRequest{
		TaskID:      int64(task3.ID),
		StoryPoints: 8,
	})
	if err != nil {
		t.Fatalf("failed to add task 3 to sprint: %v", err)
	}

	// Duplicate task constraint check
	err = svc.AddTaskToSprint(sprint.ID, &model.AddTaskToSprintRequest{
		TaskID:      int64(task1.ID),
		StoryPoints: 5,
	})
	if !errors.Is(err, ErrTaskAlreadyInSprint) {
		t.Fatalf("expected ErrTaskAlreadyInSprint, got: %v", err)
	}

	// Negative story points check
	err = svc.AddTaskToSprint(sprint.ID, &model.AddTaskToSprintRequest{
		TaskID:      999,
		StoryPoints: -2,
	})
	if !errors.Is(err, ErrInvalidStoryPoints) {
		t.Fatalf("expected ErrInvalidStoryPoints, got: %v", err)
	}

	// Update story points of task 2 from 3 to 5
	err = svc.UpdateTaskStoryPoints(sprint.ID, int64(task2.ID), 5)
	if err != nil {
		t.Fatalf("failed to update story points: %v", err)
	}

	// Invalid story points on update
	err = svc.UpdateTaskStoryPoints(sprint.ID, int64(task2.ID), -1)
	if !errors.Is(err, ErrInvalidStoryPoints) {
		t.Fatalf("expected ErrInvalidStoryPoints, got: %v", err)
	}

	// Verify tasks in sprint
	tasks, err := svc.GetSprintTasks(sprint.ID)
	if err != nil {
		t.Fatalf("failed to get sprint tasks: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("expected 3 tasks, got: %d", len(tasks))
	}

	// Remove task 3
	err = svc.RemoveTaskFromSprint(sprint.ID, int64(task3.ID))
	if err != nil {
		t.Fatalf("failed to remove task 3: %v", err)
	}

	tasksAfterRemove, err := svc.GetSprintTasks(sprint.ID)
	if err != nil {
		t.Fatalf("failed to get sprint tasks: %v", err)
	}
	if len(tasksAfterRemove) != 2 {
		t.Fatalf("expected 2 tasks after removal, got: %d", len(tasksAfterRemove))
	}

	// Check metrics: Task 1 (completed, 5 pts), Task 2 (in_progress, 5 pts)
	metrics, err := svc.GetSprintMetrics(sprint.ID)
	if err != nil {
		t.Fatalf("failed to get sprint metrics: %v", err)
	}

	if metrics.TotalTasks != 2 {
		t.Errorf("expected 2 total tasks, got: %d", metrics.TotalTasks)
	}
	if metrics.CompletedTasks != 1 {
		t.Errorf("expected 1 completed task, got: %d", metrics.CompletedTasks)
	}
	if metrics.InProgressTasks != 1 {
		t.Errorf("expected 1 in_progress task, got: %d", metrics.InProgressTasks)
	}
	if metrics.TotalStoryPoints != 10 {
		t.Errorf("expected 10 total story points, got: %d", metrics.TotalStoryPoints)
	}
	if metrics.CompletedStoryPoints != 5 {
		t.Errorf("expected 5 completed story points, got: %d", metrics.CompletedStoryPoints)
	}
	if metrics.RemainingStoryPoints != 5 {
		t.Errorf("expected 5 remaining story points, got: %d", metrics.RemainingStoryPoints)
	}
	if metrics.CompletionPercentage != 50.0 {
		t.Errorf("expected 50%% completion, got: %f", metrics.CompletionPercentage)
	}
}

func TestSprintService_CRUDAndFilters(t *testing.T) {
	svc, _ := setupTestSprintSvc(t)

	s1, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 100,
		ProjectID:   200,
		Name:        "Sprint Alpha",
	})
	if err != nil {
		t.Fatalf("failed to create sprint: %v", err)
	}

	s2, err := svc.CreateSprint(&model.CreateSprintRequest{
		WorkspaceID: 100,
		ProjectID:   201,
		Name:        "Sprint Beta",
	})
	if err != nil {
		t.Fatalf("failed to create sprint: %v", err)
	}

	// Get sprint by ID
	fetched, err := svc.GetSprint(s1.ID)
	if err != nil {
		t.Fatalf("failed to get sprint: %v", err)
	}
	if fetched.Name != "Sprint Alpha" {
		t.Errorf("expected name 'Sprint Alpha', got: %s", fetched.Name)
	}

	// Get non-existent sprint
	_, err = svc.GetSprint(99999)
	if !errors.Is(err, ErrSprintNotFound) {
		t.Fatalf("expected ErrSprintNotFound, got: %v", err)
	}

	// Update sprint
	updatedName := "Sprint Alpha Prime"
	updatedGoal := "New Goal"
	updated, err := svc.UpdateSprint(s1.ID, &model.UpdateSprintRequest{
		Name: &updatedName,
		Goal: &updatedGoal,
	})
	if err != nil {
		t.Fatalf("failed to update sprint: %v", err)
	}
	if updated.Name != "Sprint Alpha Prime" || updated.Goal != "New Goal" {
		t.Errorf("update did not persist correctly: %+v", updated)
	}

	// List with filter by project 201
	summary, err := svc.ListSprints(model.SprintFilter{
		WorkspaceID: 100,
		ProjectID:   201,
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("failed to list sprints: %v", err)
	}
	if summary.Total != 1 || len(summary.Sprints) != 1 || summary.Sprints[0].ID != s2.ID {
		t.Fatalf("filter by project failed, expected s2 only, got: %+v", summary)
	}

	// Delete planned sprint
	if err := svc.DeleteSprint(s2.ID); err != nil {
		t.Fatalf("failed to delete sprint: %v", err)
	}
	_, err = svc.GetSprint(s2.ID)
	if !errors.Is(err, ErrSprintNotFound) {
		t.Fatalf("expected ErrSprintNotFound after deletion, got: %v", err)
	}
}
