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

func setupTestGoalSvc(t *testing.T) (GoalService, repository.TaskRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'user1', 'user1@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-goal', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-goal', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	goalRepo := repository.NewGoalRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	svc := NewGoalService(goalRepo, taskRepo)
	return svc, taskRepo, sqliteDB
}

func TestGoalService_GoalCRUD(t *testing.T) {
	svc, _, _ := setupTestGoalSvc(t)

	// 1. Validation: Title required
	_, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "",
		Timeframe:   "Q4 2026",
	})
	if !errors.Is(err, ErrGoalTitleRequired) {
		t.Fatalf("expected ErrGoalTitleRequired, got: %v", err)
	}

	// 2. Validation: Workspace required
	_, err = svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 0,
		Title:       "Ship v2 API",
		Timeframe:   "Q4 2026",
	})
	if !errors.Is(err, ErrGoalWorkspaceRequired) {
		t.Fatalf("expected ErrGoalWorkspaceRequired, got: %v", err)
	}

	// 3. Validation: Timeframe required
	_, err = svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Ship v2 API",
		Timeframe:   "",
	})
	if !errors.Is(err, ErrGoalTimeframeRequired) {
		t.Fatalf("expected ErrGoalTimeframeRequired, got: %v", err)
	}

	// 4. Create valid Goal
	endDate := time.Now().Add(30 * 24 * time.Hour)
	desc := "Modernize the core task management backend services"
	created, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Accelerate Core Backend Performance",
		Description: desc,
		Timeframe:   "Q4 2026",
		EndDate:     &endDate,
	})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}
	if created.ID == 0 {
		t.Fatalf("expected positive goal ID, got 0")
	}
	if created.Status != model.GoalStatusDraft {
		t.Fatalf("expected initial status draft, got %s", created.Status)
	}

	// 5. Get Goal Details
	detail, err := svc.GetGoal(created.ID)
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}
	if detail.Goal.Title != created.Title {
		t.Fatalf("expected title %s, got %s", created.Title, detail.Goal.Title)
	}

	// 6. Update Goal
	newTitle := "Accelerate Core Platform & Infrastructure"
	newStatus := model.GoalStatusActive
	newProgress := 25.0
	updated, err := svc.UpdateGoal(created.ID, model.UpdateGoalRequest{
		Title:              &newTitle,
		Status:             &newStatus,
		ProgressPercentage: &newProgress,
	})
	if err != nil {
		t.Fatalf("failed to update goal: %v", err)
	}
	if updated.Title != newTitle || updated.Status != newStatus || updated.ProgressPercentage != 25.0 {
		t.Fatalf("updated goal fields do not match: %+v", updated)
	}

	// 7. List Goals with Filter
	wsID := int64(1)
	goals, err := svc.ListGoals(model.GoalFilter{
		WorkspaceID: &wsID,
		Status:      model.GoalStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to list goals: %v", err)
	}
	if len(goals) != 1 {
		t.Fatalf("expected 1 goal, got %d", len(goals))
	}

	// 8. Delete Goal
	if err := svc.DeleteGoal(created.ID); err != nil {
		t.Fatalf("failed to delete goal: %v", err)
	}
	_, err = svc.GetGoal(created.ID)
	if !errors.Is(err, ErrGoalNotFound) {
		t.Fatalf("expected ErrGoalNotFound after deletion, got: %v", err)
	}
}

func TestGoalService_KeyResults(t *testing.T) {
	svc, _, _ := setupTestGoalSvc(t)

	goal, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Increase Application Reliability",
		Timeframe:   "Q4 2026",
	})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	// 1. Validation: Key result title required
	_, err = svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:       "",
		MetricType:  model.KeyResultMetricNumerical,
		TargetValue: 100,
	})
	if !errors.Is(err, ErrKeyResultTitleRequired) {
		t.Fatalf("expected ErrKeyResultTitleRequired, got: %v", err)
	}

	// 2. Validation: Invalid metric type
	_, err = svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:       "Achieve uptime SLA",
		MetricType:  "unsupported_metric",
		TargetValue: 100,
	})
	if !errors.Is(err, ErrInvalidMetricType) {
		t.Fatalf("expected ErrInvalidMetricType, got: %v", err)
	}

	// 3. Validation: StartValue equals TargetValue for non-boolean
	_, err = svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:       "Achieve uptime SLA",
		MetricType:  model.KeyResultMetricPercentage,
		StartValue:  99.9,
		TargetValue: 99.9,
	})
	if !errors.Is(err, ErrInvalidTargetValue) {
		t.Fatalf("expected ErrInvalidTargetValue, got: %v", err)
	}

	// Activate goal first so status inference works dynamically
	statAct := model.GoalStatusActive
	_, _ = svc.UpdateGoal(goal.ID, model.UpdateGoalRequest{Status: &statAct})

	// 4. Create Key Result 1 (Percentage metric)
	curVal1 := 95.0
	weight1 := 2.0
	kr1, err := svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:        "System Uptime Availability",
		MetricType:   model.KeyResultMetricPercentage,
		Unit:         "%",
		StartValue:   90.0,
		TargetValue:  100.0,
		CurrentValue: &curVal1,
		Weight:       &weight1,
	})
	if err != nil {
		t.Fatalf("failed to create key result 1: %v", err)
	}
	// (95-90)/(100-90) * 100 = 50%
	if kr1.ProgressPercentage != 50.0 {
		t.Fatalf("expected 50%% progress on KR1, got %f", kr1.ProgressPercentage)
	}

	// 5. Create Key Result 2 (Currency metric)
	curVal2 := 25000.0
	weight2 := 1.0
	kr2, err := svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:        "Reduce Cloud Infrastructure Spend",
		MetricType:   model.KeyResultMetricCurrency,
		Unit:         "USD",
		StartValue:   0.0,
		TargetValue:  50000.0,
		CurrentValue: &curVal2,
		Weight:       &weight2,
	})
	if err != nil {
		t.Fatalf("failed to create key result 2: %v", err)
	}
	// 25000/50000 = 50%
	if kr2.ProgressPercentage != 50.0 {
		t.Fatalf("expected 50%% progress on KR2, got %f", kr2.ProgressPercentage)
	}

	// 6. Verify Goal aggregated progress
	// Weighted: (50*2 + 50*1) / 3 = 50%
	goalDetail, err := svc.GetGoal(goal.ID)
	if err != nil {
		t.Fatalf("failed to fetch goal detail: %v", err)
	}
	if goalDetail.Goal.ProgressPercentage != 50.0 {
		t.Fatalf("expected aggregated goal progress 50%%, got %f", goalDetail.Goal.ProgressPercentage)
	}
	if goalDetail.Goal.Status != model.GoalStatusBehind {
		t.Fatalf("expected status behind, got %s", goalDetail.Goal.Status)
	}

	// 7. Update Key Result 1 to complete (CurrentValue = 100)
	completeVal := 100.0
	updatedKr1, err := svc.UpdateKeyResult(kr1.ID, model.UpdateKeyResultRequest{
		CurrentValue: &completeVal,
	})
	if err != nil {
		t.Fatalf("failed to update key result 1: %v", err)
	}
	if updatedKr1.ProgressPercentage != 100.0 {
		t.Fatalf("expected 100%% progress, got %f", updatedKr1.ProgressPercentage)
	}

	// Goal aggregated progress: (100*2 + 50*1) / 3 = 83.3%
	goalDetail, _ = svc.GetGoal(goal.ID)
	if goalDetail.Goal.ProgressPercentage < 83.0 || goalDetail.Goal.ProgressPercentage > 84.0 {
		t.Fatalf("expected approx 83.3%% goal progress, got %f", goalDetail.Goal.ProgressPercentage)
	}
	if goalDetail.Goal.Status != model.GoalStatusOnTrack {
		t.Fatalf("expected status on_track, got %s", goalDetail.Goal.Status)
	}

	// 8. Delete Key Result 2
	if err := svc.DeleteKeyResult(kr2.ID); err != nil {
		t.Fatalf("failed to delete key result 2: %v", err)
	}
	// Goal aggregated progress now only based on KR1: 100%
	goalDetail, _ = svc.GetGoal(goal.ID)
	if goalDetail.Goal.ProgressPercentage != 100.0 {
		t.Fatalf("expected 100%% goal progress after deleting KR2, got %f", goalDetail.Goal.ProgressPercentage)
	}
	if goalDetail.Goal.Status != model.GoalStatusAchieved {
		t.Fatalf("expected goal status achieved, got %s", goalDetail.Goal.Status)
	}
}

func TestGoalService_TaskLinking(t *testing.T) {
	svc, taskRepo, _ := setupTestGoalSvc(t)

	goal, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Ship Self-Service Onboarding",
		Timeframe:   "Q4 2026",
	})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	kr, err := svc.CreateKeyResult(goal.ID, model.CreateKeyResultRequest{
		Title:       "Deliver 3 onboarding modules",
		MetricType:  model.KeyResultMetricNumerical,
		StartValue:  0,
		TargetValue: 3,
	})
	if err != nil {
		t.Fatalf("failed to create key result: %v", err)
	}

	now := time.Now()
	// Insert 2 tasks
	task1 := &model.Task{
		UserID:    "user_1",
		Title:     "Module 1: User Signup",
		Status:    model.TaskStatusCompleted,
		Completed: true,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(task1)

	task2 := &model.Task{
		UserID:    "user_1",
		Title:     "Module 2: Workspace Setup",
		Status:    model.TaskStatusTodo,
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_ = taskRepo.Create(task2)

	// Link tasks
	if err := svc.LinkTask(kr.ID, task1.ID); err != nil {
		t.Fatalf("failed to link task 1: %v", err)
	}
	if err := svc.LinkTask(kr.ID, task2.ID); err != nil {
		t.Fatalf("failed to link task 2: %v", err)
	}

	// Linking already linked task should return error
	if err := svc.LinkTask(kr.ID, task1.ID); err == nil {
		t.Fatalf("expected error when linking duplicate task, got nil")
	}

	// List linked tasks
	linked, err := svc.ListLinkedTasks(kr.ID)
	if err != nil {
		t.Fatalf("failed to list linked tasks: %v", err)
	}
	if len(linked) != 2 {
		t.Fatalf("expected 2 linked tasks, got %d", len(linked))
	}

	// Key Result progress should be auto-synced to 1/2 = 50%
	updatedKr, err := svc.GetKeyResult(kr.ID)
	if err != nil {
		t.Fatalf("failed to get key result: %v", err)
	}
	if updatedKr.ProgressPercentage != 50.0 {
		t.Fatalf("expected 50%% task-synced progress, got %f", updatedKr.ProgressPercentage)
	}

	// Unlink task 2 (leaving only task1 which is completed)
	if err := svc.UnlinkTask(kr.ID, task2.ID); err != nil {
		t.Fatalf("failed to unlink task 2: %v", err)
	}

	// Now 1/1 completed = 100%
	updatedKr, _ = svc.GetKeyResult(kr.ID)
	if updatedKr.ProgressPercentage != 100.0 {
		t.Fatalf("expected 100%% progress after unlinking incomplete task, got %f", updatedKr.ProgressPercentage)
	}
}

func TestGoalService_CheckIns(t *testing.T) {
	svc, _, _ := setupTestGoalSvc(t)

	goal, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Enterprise Single Sign-On",
		Timeframe:   "Q4 2026",
	})
	if err != nil {
		t.Fatalf("failed to create goal: %v", err)
	}

	// Validation: Invalid confidence rating
	_, err = svc.CreateCheckIn("user_1", goal.ID, model.GoalCheckInRequest{
		ConfidenceRating: "super_confident",
	})
	if !errors.Is(err, ErrInvalidConfidenceRating) {
		t.Fatalf("expected ErrInvalidConfidenceRating, got: %v", err)
	}

	// Record valid check-in
	notes := "SAML 2.0 and OIDC integrations completed in staging"
	prog := 70.0
	status := model.GoalStatusOnTrack
	checkIn, err := svc.CreateCheckIn("user_1", goal.ID, model.GoalCheckInRequest{
		ConfidenceRating:   model.GoalConfidenceHigh,
		ProgressPercentage: &prog,
		Note:               notes,
		Status:             status,
	})
	if err != nil {
		t.Fatalf("failed to record check in: %v", err)
	}
	if checkIn.ID == 0 {
		t.Fatalf("expected positive check-in ID")
	}

	// Verify check-ins list
	checkIns, err := svc.ListCheckIns(goal.ID)
	if err != nil {
		t.Fatalf("failed to list check-ins: %v", err)
	}
	if len(checkIns) != 1 {
		t.Fatalf("expected 1 check-in, got %d", len(checkIns))
	}
	if checkIns[0].Note != notes {
		t.Fatalf("expected notes %s, got %s", notes, checkIns[0].Note)
	}

	// Verify goal status updated
	goalDetail, _ := svc.GetGoal(goal.ID)
	if goalDetail.Goal.Status != model.GoalStatusOnTrack {
		t.Fatalf("expected goal status on_track, got %s", goalDetail.Goal.Status)
	}
}

func TestGoalService_GoalTree(t *testing.T) {
	svc, _, _ := setupTestGoalSvc(t)

	// Parent strategic goal
	parent, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Expand Enterprise Market Share",
		Timeframe:   "FY 2026",
	})
	if err != nil {
		t.Fatalf("failed to create parent goal: %v", err)
	}

	// Child sub-goal 1
	child1, err := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID:  1,
		Title:        "SOC-2 Type II Compliance",
		Timeframe:    "Q4 2026",
		ParentGoalID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("failed to create child goal 1: %v", err)
	}

	// Child sub-goal 2
	_, err = svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID:  1,
		Title:        "Dedicated Enterprise VPC Deployment",
		Timeframe:    "Q4 2026",
		ParentGoalID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("failed to create child goal 2: %v", err)
	}

	// Sub-child under child 1
	_, err = svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID:  1,
		Title:        "Complete Third-Party Security Audit",
		Timeframe:    "Q4 2026",
		ParentGoalID: &child1.ID,
	})
	if err != nil {
		t.Fatalf("failed to create sub-child goal: %v", err)
	}

	// Get Tree
	tree, err := svc.GetGoalTree(parent.ID)
	if err != nil {
		t.Fatalf("failed to get goal tree: %v", err)
	}
	if tree.Goal.ID != parent.ID {
		t.Fatalf("expected tree root ID %d, got %d", parent.ID, tree.Goal.ID)
	}
	if len(tree.SubGoals) != 2 {
		t.Fatalf("expected 2 direct children, got %d", len(tree.SubGoals))
	}
	if len(tree.SubGoals[0].SubGoals) != 1 {
		t.Fatalf("expected 1 grandchild under child 1, got %d", len(tree.SubGoals[0].SubGoals))
	}
}

func TestGoalService_Metrics(t *testing.T) {
	svc, _, _ := setupTestGoalSvc(t)

	// Create goals in workspace 1
	g1, _ := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Goal 1: Completed",
		Timeframe:   "Q4 2026",
	})
	statComp := model.GoalStatusAchieved
	prog100 := 100.0
	_, _ = svc.UpdateGoal(g1.ID, model.UpdateGoalRequest{
		Status:             &statComp,
		ProgressPercentage: &prog100,
	})

	g2, _ := svc.CreateGoal("user_1", model.CreateGoalRequest{
		WorkspaceID: 1,
		Title:       "Goal 2: In Progress",
		Timeframe:   "Q4 2026",
	})
	statProg := model.GoalStatusOnTrack
	prog50 := 50.0
	_, _ = svc.UpdateGoal(g2.ID, model.UpdateGoalRequest{
		Status:             &statProg,
		ProgressPercentage: &prog50,
	})

	// Fetch metrics
	metrics, err := svc.GetGoalMetrics(1)
	if err != nil {
		t.Fatalf("failed to get goal metrics: %v", err)
	}
	if metrics.TotalGoals != 2 {
		t.Fatalf("expected 2 total goals, got %d", metrics.TotalGoals)
	}
	if metrics.AchievedGoals != 1 {
		t.Fatalf("expected 1 achieved goal, got %d", metrics.AchievedGoals)
	}
	if metrics.OnTrackGoals != 1 {
		t.Fatalf("expected 1 on_track goal, got %d", metrics.OnTrackGoals)
	}
	if metrics.AverageProgress != 75.0 {
		t.Fatalf("expected 75%% average progress, got %f", metrics.AverageProgress)
	}
}
