package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestChangeSvc(t *testing.T) (ChangeService, repository.ChangeRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_3', 'charlie', 'charlie@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-chg', 'user_1', datetime('now'), datetime('now')),
		       (2, 'Product WS', 'prod-ws-chg', 'user_1', datetime('now'), datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Core Platform', 'Core Infrastructure', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Database Schema Migration', 'Execute flyway migrations', 'todo', 0, datetime('now'), 60, 'high', datetime('now'), datetime('now')),
		       (502, 'user_1', 101, 'Smoke Test Verification', 'Post-deploy check', 'todo', 0, datetime('now'), 30, 'medium', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	changeRepo := repository.NewChangeRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	userRepo := repository.NewUserRepository(sqliteDB)
	auditRepo := repository.NewAuditRepository(sqliteDB)
	auditSvc := NewAuditService(auditRepo, workspaceRepo, userRepo)

	svc := NewChangeService(changeRepo, workspaceRepo, projectRepo, taskRepo, userRepo, auditSvc)
	return svc, changeRepo, sqliteDB
}

func TestChangeService_CreateAndGet(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	// 1. Missing workspace
	_, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 0,
		Title:       "Deploy v2.4",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Non-existent workspace
	_, err = svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 999,
		Title:       "Deploy v2.4",
	})
	if !errors.Is(err, ErrWorkspaceNotFound) {
		t.Fatalf("expected ErrWorkspaceNotFound, got %v", err)
	}

	// 3. Missing title
	_, err = svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "   ",
	})
	if !errors.Is(err, ErrChangeTitleRequired) {
		t.Fatalf("expected ErrChangeTitleRequired, got %v", err)
	}

	// 4. Invalid change type
	_, err = svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Deploy v2.4",
		ChangeType:  "instant",
	})
	if !errors.Is(err, ErrInvalidChangeType) {
		t.Fatalf("expected ErrInvalidChangeType, got %v", err)
	}

	// 5. Invalid risk level
	_, err = svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Deploy v2.4",
		ChangeType:  "normal",
		RiskLevel:   "extreme",
	})
	if !errors.Is(err, ErrInvalidChangeRiskLevel) {
		t.Fatalf("expected ErrInvalidChangeRiskLevel, got %v", err)
	}

	// 6. Successful creation
	startTime := "2026-11-10T02:00:00Z"
	endTime := "2026-11-10T04:00:00Z"
	projID := int64(101)

	change, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID:        1,
		ProjectID:          &projID,
		Title:              "Deploy Payment Service v2.4",
		Description:        "Upgrades payment microservice",
		ChangeType:         "normal",
		RiskLevel:          "high",
		ImpactLevel:        "medium",
		ReasonForChange:    "Security patch",
		ImplementationPlan: "Apply DB migrations and helm release",
		RollbackPlan:       "Helm rollback to previous revision",
		TestPlan:           "Run automated synthetic checks",
		AssignedToUserID:   "user_2",
		PlannedStartAt:     &startTime,
		PlannedEndAt:       &endTime,
	})
	if err != nil {
		t.Fatalf("unexpected error creating change: %v", err)
	}
	if change.ID <= 0 {
		t.Fatalf("expected positive ID, got %d", change.ID)
	}
	if change.ChangeNumber != "CHG-1" {
		t.Fatalf("expected change number CHG-1, got %s", change.ChangeNumber)
	}
	if change.Status != model.ChangeStatusDraft {
		t.Fatalf("expected status draft, got %s", change.Status)
	}
	if change.RequesterName != "alice" {
		t.Fatalf("expected requester name alice, got %s", change.RequesterName)
	}
	if change.AssigneeName != "bob" {
		t.Fatalf("expected assignee name bob, got %s", change.AssigneeName)
	}
	if change.ProjectName != "Core Platform" {
		t.Fatalf("expected project name Core Platform, got %s", change.ProjectName)
	}

	// 7. Get by ID
	fetched, err := svc.GetChange(change.ID)
	if err != nil {
		t.Fatalf("failed to fetch change by ID: %v", err)
	}
	if fetched.Title != change.Title {
		t.Fatalf("expected title %s, got %s", change.Title, fetched.Title)
	}

	// 8. Get by Number
	byNum, err := svc.GetChangeByNumber(1, 1)
	if err != nil {
		t.Fatalf("failed to fetch change by number: %v", err)
	}
	if byNum.ID != change.ID {
		t.Fatalf("expected change ID %d, got %d", change.ID, byNum.ID)
	}
}

func TestChangeService_ListAndFilter(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	_, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Routine Certificate Renewal",
		ChangeType:  model.ChangeTypeStandard,
		RiskLevel:   model.ChangeRiskLow,
	})
	if err != nil {
		t.Fatalf("create standard change failed: %v", err)
	}

	_, err = svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Emergency Kernel Hotfix",
		ChangeType:  model.ChangeTypeEmergency,
		RiskLevel:   model.ChangeRiskCritical,
	})
	if err != nil {
		t.Fatalf("create emergency change failed: %v", err)
	}

	// 1. Filter by workspace
	list, err := svc.ListChanges(model.ChangeFilter{WorkspaceID: 1})
	if err != nil {
		t.Fatalf("list changes failed: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(list))
	}

	// 2. Filter by change type
	emType := model.ChangeTypeEmergency
	emList, err := svc.ListChanges(model.ChangeFilter{WorkspaceID: 1, ChangeType: &emType})
	if err != nil {
		t.Fatalf("filter by emergency failed: %v", err)
	}
	if len(emList) != 1 || emList[0].ChangeType != model.ChangeTypeEmergency {
		t.Fatalf("expected 1 emergency change, got %d", len(emList))
	}

	// 3. Search query
	q := "Kernel"
	searchList, err := svc.ListChanges(model.ChangeFilter{WorkspaceID: 1, Search: &q})
	if err != nil {
		t.Fatalf("search changes failed: %v", err)
	}
	if len(searchList) != 1 {
		t.Fatalf("expected 1 search result, got %d", len(searchList))
	}
}

func TestChangeService_UpdateAndDelete(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	change, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Initial Change",
		ChangeType:  model.ChangeTypeNormal,
		RiskLevel:   model.ChangeRiskMedium,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Update title and risk in draft
	newTitle := "Updated Production Change"
	newRisk := model.ChangeRiskHigh
	updated, err := svc.UpdateChange(change.ID, "user_1", model.UpdateChangeRequest{
		Title:     &newTitle,
		RiskLevel: &newRisk,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Title != newTitle || updated.RiskLevel != newRisk {
		t.Fatalf("expected title %s and risk %s, got %s and %s", newTitle, newRisk, updated.Title, updated.RiskLevel)
	}

	// 2. Transition to completed and verify update is blocked
	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCABReview})
	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	_, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCompleted})
	if err != nil {
		t.Fatalf("transition to completed failed: %v", err)
	}

	blockedTitle := "Attempted Post-Complete Edit"
	_, err = svc.UpdateChange(change.ID, "user_1", model.UpdateChangeRequest{Title: &blockedTitle})
	if !errors.Is(err, ErrChangeNotEditable) {
		t.Fatalf("expected ErrChangeNotEditable on completed change, got %v", err)
	}

	// 3. Deleting completed change should fail
	err = svc.DeleteChange(change.ID, "user_1")
	if !errors.Is(err, ErrChangeNotEditable) {
		t.Fatalf("expected ErrChangeNotEditable deleting completed change, got %v", err)
	}

	// 4. Create another draft change and delete it
	draftChange, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Scratchpad Change",
	})
	if err != nil {
		t.Fatalf("create draft failed: %v", err)
	}
	if err := svc.DeleteChange(draftChange.ID, "user_1"); err != nil {
		t.Fatalf("delete draft change failed: %v", err)
	}
	_, err = svc.GetChange(draftChange.ID)
	if !errors.Is(err, ErrChangeNotFound) {
		t.Fatalf("expected ErrChangeNotFound after deletion, got %v", err)
	}
}

func TestChangeService_LifecycleTransitions(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	change, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Full Lifecycle Flow",
		ChangeType:  model.ChangeTypeNormal,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// Invalid direct jump: draft -> implementing
	_, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	if !errors.Is(err, ErrInvalidChangeTransition) {
		t.Fatalf("expected ErrInvalidChangeTransition, got %v", err)
	}

	// draft -> submitted
	st, err := svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	if err != nil || st.Status != model.ChangeStatusSubmitted {
		t.Fatalf("failed transition to submitted: %v", err)
	}

	// submitted -> cab_review
	st, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCABReview})
	if err != nil || st.Status != model.ChangeStatusCABReview {
		t.Fatalf("failed transition to cab_review: %v", err)
	}

	// cab_review -> approved
	st, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	if err != nil || st.Status != model.ChangeStatusApproved {
		t.Fatalf("failed transition to approved: %v", err)
	}

	// approved -> scheduled
	st, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusScheduled})
	if err != nil || st.Status != model.ChangeStatusScheduled {
		t.Fatalf("failed transition to scheduled: %v", err)
	}

	// scheduled -> implementing (should record actual_start_at)
	st, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	if err != nil || st.Status != model.ChangeStatusImplementing {
		t.Fatalf("failed transition to implementing: %v", err)
	}
	if st.ActualStartAt == nil {
		t.Fatalf("expected actual_start_at to be populated")
	}

	// implementing -> completed (should record actual_end_at)
	st, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCompleted})
	if err != nil || st.Status != model.ChangeStatusCompleted {
		t.Fatalf("failed transition to completed: %v", err)
	}
	if st.ActualEndAt == nil {
		t.Fatalf("expected actual_end_at to be populated")
	}

	// completed -> draft is illegal
	_, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusDraft})
	if !errors.Is(err, ErrInvalidChangeTransition) {
		t.Fatalf("expected ErrInvalidChangeTransition from completed, got %v", err)
	}

	// Test Failure and Rollback path
	failChange, _ := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Failed Deployment",
		ChangeType:  model.ChangeTypeEmergency,
	})
	_, _ = svc.ChangeStatus(failChange.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	_, _ = svc.ChangeStatus(failChange.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	_, _ = svc.ChangeStatus(failChange.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	st, err = svc.ChangeStatus(failChange.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusFailed})
	if err != nil || st.Status != model.ChangeStatusFailed {
		t.Fatalf("transition to failed should succeed: %v", err)
	}
	st, err = svc.ChangeStatus(failChange.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusRolledBack})
	if err != nil || st.Status != model.ChangeStatusRolledBack {
		t.Fatalf("transition to rolled_back should succeed: %v", err)
	}
}

func TestChangeService_CABApprovalWorkflow(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	change, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "CAB Gated Release",
		ChangeType:  model.ChangeTypeNormal,
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	_, _ = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCABReview})

	// 1. Add reviewer
	appr1, err := svc.AddReviewer(change.ID, "user_1", model.AddReviewerRequest{ReviewerID: "user_2"})
	if err != nil {
		t.Fatalf("add reviewer failed: %v", err)
	}
	if appr1.ReviewerID != "user_2" || appr1.Decision != model.ApprovalDecisionPending {
		t.Fatalf("expected pending approval for user_2, got %+v", appr1)
	}

	// 2. Cast Approval
	decided, err := svc.ReviewChange(change.ID, "user_2", model.ChangeApprovalRequest{
		Decision: model.ApprovalDecisionApproved,
		Comments: "Architecture and test plan look solid",
	})
	if err != nil {
		t.Fatalf("review change failed: %v", err)
	}
	if decided.Decision != model.ApprovalDecisionApproved || decided.DecidedAt == nil {
		t.Fatalf("expected approved decision, got %+v", decided)
	}

	// 3. Add second reviewer who rejects
	_, err = svc.AddReviewer(change.ID, "user_1", model.AddReviewerRequest{ReviewerID: "user_3"})
	if err != nil {
		t.Fatalf("add user_3 reviewer failed: %v", err)
	}
	_, err = svc.ReviewChange(change.ID, "user_3", model.ChangeApprovalRequest{
		Decision: model.ApprovalDecisionRejected,
		Comments: "Rollback procedure does not cover message queue lag",
	})
	if err != nil {
		t.Fatalf("user_3 rejection review failed: %v", err)
	}

	// 4. Verifying transition to approved fails because of rejection
	_, err = svc.ChangeStatus(change.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	if err == nil {
		t.Fatalf("expected approval transition to be rejected due to dissenting vote")
	}

	// 5. List approvals
	approvals, err := svc.ListApprovals(change.ID)
	if err != nil {
		t.Fatalf("list approvals failed: %v", err)
	}
	if len(approvals) != 2 {
		t.Fatalf("expected 2 approvals, got %d", len(approvals))
	}

	// 6. Delete reviewer
	if err := svc.DeleteReviewer(approvals[1].ID, "user_1"); err != nil {
		t.Fatalf("delete reviewer failed: %v", err)
	}
	rem, _ := svc.ListApprovals(change.ID)
	if len(rem) != 1 {
		t.Fatalf("expected 1 approval remaining, got %d", len(rem))
	}
}

func TestChangeService_TaskLinkageAndPIR(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	change, err := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Task and PIR Integration",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	// 1. Link Task
	link, err := svc.LinkTask(change.ID, model.LinkChangeTaskRequest{
		TaskID: 501,
		Phase:  model.ChangePhaseImplementation,
	})
	if err != nil {
		t.Fatalf("link task failed: %v", err)
	}
	if link.TaskID != 501 || link.Phase != model.ChangePhaseImplementation {
		t.Fatalf("expected task 501 linked, got %+v", link)
	}

	// 2. Duplicate link fails
	_, err = svc.LinkTask(change.ID, model.LinkChangeTaskRequest{
		TaskID: 501,
	})
	if !errors.Is(err, ErrTaskAlreadyLinked) {
		t.Fatalf("expected ErrTaskAlreadyLinked, got %v", err)
	}

	// 3. List linked tasks
	tasks, err := svc.ListLinkedTasks(change.ID)
	if err != nil || len(tasks) != 1 {
		t.Fatalf("expected 1 linked task, got %d (err: %v)", len(tasks), err)
	}

	// 4. Unlink task
	if err := svc.UnlinkTask(change.ID, 501); err != nil {
		t.Fatalf("unlink task failed: %v", err)
	}
	tasksAfter, _ := svc.ListLinkedTasks(change.ID)
	if len(tasksAfter) != 0 {
		t.Fatalf("expected 0 linked tasks after unlink, got %d", len(tasksAfter))
	}

	// 5. Submit PIR
	pir, err := svc.SubmitPIR(change.ID, "user_2", model.ChangePIRRequest{
		Outcome:        model.PIROutcomeSuccessful,
		ReviewNotes:    "Deployment completed within scheduled change window without downtime",
		LessonsLearned: "Automated canary checks accelerated verification by 15 minutes",
	})
	if err != nil {
		t.Fatalf("submit PIR failed: %v", err)
	}
	if pir.Outcome != model.PIROutcomeSuccessful || pir.ReviewedByUsername != "bob" {
		t.Fatalf("expected successful outcome by bob, got %+v", pir)
	}

	// 6. Get PIR
	getPIR, err := svc.GetPIR(change.ID)
	if err != nil {
		t.Fatalf("get PIR failed: %v", err)
	}
	if getPIR.ReviewNotes != pir.ReviewNotes {
		t.Fatalf("expected matching review notes, got %s", getPIR.ReviewNotes)
	}
}

func TestChangeService_Metrics(t *testing.T) {
	svc, _, _ := setupTestChangeSvc(t)

	// Create 1 completed standard change
	c1, _ := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Change 1",
		ChangeType:  model.ChangeTypeStandard,
	})
	_, _ = svc.ChangeStatus(c1.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	_, _ = svc.ChangeStatus(c1.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	_, _ = svc.ChangeStatus(c1.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	_, _ = svc.ChangeStatus(c1.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusCompleted})

	// Create 1 failed emergency change
	c2, _ := svc.CreateChange("user_1", model.CreateChangeRequest{
		WorkspaceID: 1,
		Title:       "Change 2",
		ChangeType:  model.ChangeTypeEmergency,
	})
	_, _ = svc.ChangeStatus(c2.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusSubmitted})
	_, _ = svc.ChangeStatus(c2.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusApproved})
	_, _ = svc.ChangeStatus(c2.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusImplementing})
	_, _ = svc.ChangeStatus(c2.ID, "user_1", model.ChangeStatusRequest{Status: model.ChangeStatusFailed})

	// Calculate metrics
	metrics, err := svc.GetChangeMetrics(1)
	if err != nil {
		t.Fatalf("get metrics failed: %v", err)
	}
	if metrics.TotalChanges != 2 {
		t.Fatalf("expected 2 total changes, got %d", metrics.TotalChanges)
	}
	if metrics.CompletedCount != 1 || metrics.FailedCount != 1 {
		t.Fatalf("expected 1 completed and 1 failed, got completed=%d failed=%d", metrics.CompletedCount, metrics.FailedCount)
	}
	if metrics.EmergencyCount != 1 {
		t.Fatalf("expected 1 emergency change, got %d", metrics.EmergencyCount)
	}
	if metrics.SuccessRate != 50.0 {
		t.Fatalf("expected 50.0%% success rate, got %f", metrics.SuccessRate)
	}
	if metrics.EmergencyRatio != 50.0 {
		t.Fatalf("expected 50.0%% emergency ratio, got %f", metrics.EmergencyRatio)
	}
}
