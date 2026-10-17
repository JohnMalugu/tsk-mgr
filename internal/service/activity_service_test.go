package service

import (
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestActivitySvc(t *testing.T) (ActivityService, TaskDBService) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	taskRepo := repository.NewTaskRepository(sqliteDB)
	activityRepo := repository.NewActivityRepository(sqliteDB)

	activitySvc := NewActivityService(activityRepo, taskRepo)
	taskSvc := NewTaskDBService(taskRepo)
	taskSvc.SetActivityService(activitySvc)

	return activitySvc, taskSvc
}

func TestActivityService_RecordAndQuery(t *testing.T) {
	activitySvc, taskSvc := setupTestActivitySvc(t)
	userID := "user-audit-1"
	otherUser := "user-audit-2"

	// 1. Validation tests
	_, err := activitySvc.RecordActivity(userID, nil, "   ", "Some summary", nil)
	if err != ErrActivityActionRequired {
		t.Fatalf("expected ErrActivityActionRequired, got %v", err)
	}

	_, err = activitySvc.RecordActivity(userID, nil, "custom_action", "   ", nil)
	if err != ErrActivitySummaryRequired {
		t.Fatalf("expected ErrActivitySummaryRequired, got %v", err)
	}

	// 2. Task-scoped activity validation
	nonExistentTaskID := 99999
	_, err = activitySvc.RecordActivity(userID, &nonExistentTaskID, "test_action", "summary", nil)
	if err != ErrTaskNotFoundDB {
		t.Fatalf("expected ErrTaskNotFoundDB, got %v", err)
	}

	// 3. Create task (lifecycle should auto-record activity)
	task, err := taskSvc.CreateTask(userID, "Launch Platform", "Prepare production rollout", time.Now().Add(24*time.Hour), 120, "high")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	// Cross-user access check
	_, err = activitySvc.RecordActivity(otherUser, &task.ID, "tamper_action", "summary", nil)
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 4. Manually record additional events
	meta := map[string]string{"version": "v2.0.0", "env": "prod"}
	record, err := activitySvc.RecordActivity(userID, &task.ID, "deployment_started", "Deployment to prod started", meta)
	if err != nil {
		t.Fatalf("failed to record activity: %v", err)
	}
	if record.ID == 0 || record.Action != "deployment_started" {
		t.Errorf("unexpected record: %+v", record)
	}

	// Complete task (should auto-record task_completed)
	if _, err := taskSvc.CompleteTask(task.ID, userID, true); err != nil {
		t.Fatalf("failed to complete task: %v", err)
	}

	// 5. Query user activity
	feed, err := activitySvc.GetUserActivity(userID, &model.ActivityFilter{})
	if err != nil {
		t.Fatalf("failed to get user activity: %v", err)
	}
	// Expected: task_created, deployment_started, task_completed -> total 3
	if feed.Total != 3 {
		t.Errorf("expected 3 activity events, got %d", feed.Total)
	}
	if len(feed.Items) != 3 {
		t.Errorf("expected 3 items in page, got %d", len(feed.Items))
	}

	// 6. Filter by action
	actionFeed, err := activitySvc.GetUserActivity(userID, &model.ActivityFilter{Action: "deployment_started"})
	if err != nil {
		t.Fatalf("failed to filter by action: %v", err)
	}
	if actionFeed.Total != 1 || len(actionFeed.Items) != 1 {
		t.Errorf("expected 1 deployment_started event, got %d", actionFeed.Total)
	}

	// 7. Query task activity
	taskFeed, err := activitySvc.GetTaskActivity(task.ID, userID, 0, 20)
	if err != nil {
		t.Fatalf("failed to get task activity: %v", err)
	}
	if len(taskFeed) != 3 {
		t.Errorf("expected 3 events for task, got %d", len(taskFeed))
	}

	// Task activity authorization check
	_, err = activitySvc.GetTaskActivity(task.ID, otherUser, 0, 20)
	if err != ErrUnauthorizedTask {
		t.Fatalf("expected ErrUnauthorizedTask, got %v", err)
	}

	// 8. Activity summary and metrics
	summary, err := activitySvc.GetActivitySummary(userID, nil, nil)
	if err != nil {
		t.Fatalf("failed to get activity summary: %v", err)
	}
	if summary.TotalEvents != 3 {
		t.Errorf("expected 3 total events in summary, got %d", summary.TotalEvents)
	}
	if len(summary.ActionBreakdown) != 3 {
		t.Errorf("expected 3 unique actions in breakdown, got %d", len(summary.ActionBreakdown))
	}

	// 9. Purge invalid cutoff
	_, err = activitySvc.PurgeOldActivities(0)
	if err != ErrInvalidRetentionCutoff {
		t.Fatalf("expected ErrInvalidRetentionCutoff, got %v", err)
	}

	// 10. Purge with future-relative cutoff (30 days ago should retain all current items)
	purged, err := activitySvc.PurgeOldActivities(30)
	if err != nil {
		t.Fatalf("failed to purge activities: %v", err)
	}
	if purged != 0 {
		t.Errorf("expected 0 purged events, got %d", purged)
	}
}
