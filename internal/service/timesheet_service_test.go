package service

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestTimesheetSvc(t *testing.T) (TimesheetService, repository.TimesheetRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_mgr', 'charlie', 'charlie@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'Engineering WS', 'eng-ws-ts', 'user_mgr', datetime('now'), datetime('now')),
		       (2, 'Operations WS', 'ops-ws-ts', 'user_mgr', datetime('now'), datetime('now'));

		INSERT INTO workspace_members (workspace_id, user_id, role, joined_at)
		VALUES (1, 'user_1', 'member', datetime('now')),
		       (1, 'user_mgr', 'admin', datetime('now')),
		       (2, 'user_mgr', 'admin', datetime('now'));

		INSERT INTO projects (id, user_id, name, description, color, created_at, updated_at)
		VALUES (101, 'user_1', 'Backend API', 'Golang Service', '#4f46e5', datetime('now'), datetime('now'));

		INSERT INTO tasks (id, user_id, project_id, title, description, status, completed, due_date, estimate_minutes, priority, created_at, updated_at)
		VALUES (501, 'user_1', 101, 'Build Timesheet API', 'Implement timesheet service', 'in_progress', 0, datetime('now'), 120, 'high', datetime('now'), datetime('now')),
		       (502, 'user_1', 101, 'Unit Tests', 'Write unit tests', 'todo', 0, datetime('now'), 60, 'medium', datetime('now'), datetime('now'));

		INSERT INTO time_entries (id, task_id, user_id, started_at, ended_at, duration_seconds, note, created_at, updated_at)
		VALUES (901, 501, 'user_1', '2026-11-02 09:00:00', '2026-11-02 11:00:00', 7200, 'Developed models and schema', '2026-11-02 11:00:00', '2026-11-02 11:00:00'),
		       (902, 502, 'user_1', '2026-11-03 14:00:00', '2026-11-03 15:30:00', 5400, 'Wrote unit tests', '2026-11-03 15:30:00', '2026-11-03 15:30:00');
	`)
	if err != nil {
		t.Fatalf("failed to seed test database: %v", err)
	}

	timesheetRepo := repository.NewTimesheetRepository(sqliteDB)
	timeEntryRepo := repository.NewTimeEntryRepository(sqliteDB)
	taskRepo := repository.NewTaskRepository(sqliteDB)
	projectRepo := repository.NewProjectRepository(sqliteDB)
	workspaceRepo := repository.NewWorkspaceRepository(sqliteDB)

	svc := NewTimesheetService(timesheetRepo, timeEntryRepo, taskRepo, projectRepo, workspaceRepo)
	return svc, timesheetRepo, sqliteDB
}

func TestTimesheetService_CreateAndGet(t *testing.T) {
	svc, _, dbConn := setupTestTimesheetSvc(t)
	defer dbConn.Close()

	// 1. Validation: workspace required
	_, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     0,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
	})
	if !errors.Is(err, ErrWorkspaceRequired) {
		t.Fatalf("expected ErrWorkspaceRequired, got %v", err)
	}

	// 2. Validation: invalid period dates (end before start)
	_, err = svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-08",
		PeriodEndDate:   "2026-11-02",
	})
	if !errors.Is(err, ErrInvalidPeriodDates) {
		t.Fatalf("expected ErrInvalidPeriodDates, got %v", err)
	}

	// 3. Authorization: user not in workspace
	_, err = svc.CreateTimesheet("user_2", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
	})
	if !errors.Is(err, ErrUnauthorizedTimesheet) {
		t.Fatalf("expected ErrUnauthorizedTimesheet, got %v", err)
	}

	// 4. Success creation
	ts, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
		Notes:           "Weekly timesheet for sprint 14",
	})
	if err != nil {
		t.Fatalf("unexpected error creating timesheet: %v", err)
	}
	if ts.ID <= 0 || ts.Status != model.TimesheetStatusDraft {
		t.Fatalf("expected draft timesheet with ID, got %+v", ts)
	}

	// 5. Duplicate timesheet rejection
	_, err = svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
	})
	if !errors.Is(err, ErrTimesheetAlreadyExists) {
		t.Fatalf("expected ErrTimesheetAlreadyExists, got %v", err)
	}

	// 6. Get timesheet by ID
	fetched, err := svc.GetTimesheet(ts.ID)
	if err != nil {
		t.Fatalf("failed to get timesheet: %v", err)
	}
	if fetched.ID != ts.ID || fetched.Notes != "Weekly timesheet for sprint 14" {
		t.Fatalf("fetched timesheet mismatch: %+v", fetched)
	}
}

func TestTimesheetService_LineManagementAndTotals(t *testing.T) {
	svc, _, dbConn := setupTestTimesheetSvc(t)
	defer dbConn.Close()

	ts, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
	})
	if err != nil {
		t.Fatalf("failed to create timesheet: %v", err)
	}

	taskID := int64(501)
	isBillableTrue := true
	isBillableFalse := false

	// Validation: duration <= 0
	_, err = svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-02",
		DurationMinutes: 0,
	})
	if !errors.Is(err, ErrInvalidLineDuration) {
		t.Fatalf("expected ErrInvalidLineDuration, got %v", err)
	}

	// Validation: date outside period
	_, err = svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-09",
		DurationMinutes: 60,
	})
	if !errors.Is(err, ErrInvalidWorkDate) {
		t.Fatalf("expected ErrInvalidWorkDate, got %v", err)
	}

	// Add Line 1: 120 mins billable
	line1, err := svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-02",
		DurationMinutes: 120,
		IsBillable:      &isBillableTrue,
		Description:     "Core engine implementation",
	})
	if err != nil {
		t.Fatalf("failed to add line 1: %v", err)
	}
	if line1.ID <= 0 || line1.DurationMinutes != 120 {
		t.Fatalf("unexpected line 1: %+v", line1)
	}

	// Add Line 2: 60 mins non-billable
	line2, err := svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-03",
		DurationMinutes: 60,
		IsBillable:      &isBillableFalse,
		Description:     "Team standup and internal review",
	})
	if err != nil {
		t.Fatalf("failed to add line 2: %v", err)
	}

	// Verify totals on timesheet
	tsUpdated, err := svc.GetTimesheet(ts.ID)
	if err != nil {
		t.Fatalf("failed to get updated timesheet: %v", err)
	}
	if tsUpdated.TotalMinutes != 180 || tsUpdated.BillableMinutes != 120 {
		t.Fatalf("expected 180 total mins and 120 billable mins, got total=%d billable=%d", tsUpdated.TotalMinutes, tsUpdated.BillableMinutes)
	}

	// Update line 2: increase duration to 90 mins
	line2Updated, err := svc.UpdateLine(ts.ID, line2.ID, "user_1", model.TimesheetLineRequest{
		DurationMinutes: 90,
		Description:     "Extended internal review",
	})
	if err != nil {
		t.Fatalf("failed to update line 2: %v", err)
	}
	if line2Updated.DurationMinutes != 90 {
		t.Fatalf("expected 90 mins, got %d", line2Updated.DurationMinutes)
	}

	// Check totals after update
	tsUpdated, _ = svc.GetTimesheet(ts.ID)
	if tsUpdated.TotalMinutes != 210 || tsUpdated.BillableMinutes != 120 {
		t.Fatalf("expected 210 total mins after line update, got total=%d", tsUpdated.TotalMinutes)
	}

	// Delete line 1
	if err := svc.DeleteLine(ts.ID, line1.ID, "user_1"); err != nil {
		t.Fatalf("failed to delete line 1: %v", err)
	}

	// Check totals after deletion
	tsUpdated, _ = svc.GetTimesheet(ts.ID)
	if tsUpdated.TotalMinutes != 90 || tsUpdated.BillableMinutes != 0 {
		t.Fatalf("expected 90 total mins and 0 billable mins, got total=%d billable=%d", tsUpdated.TotalMinutes, tsUpdated.BillableMinutes)
	}
}

func TestTimesheetService_AutoPopulateFromTimeEntries(t *testing.T) {
	svc, _, dbConn := setupTestTimesheetSvc(t)
	defer dbConn.Close()

	// Create timesheet with AutoPopulate=true
	ts, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
		AutoPopulate:    true,
	})
	if err != nil {
		t.Fatalf("failed to create auto-populated timesheet: %v", err)
	}

	lines, err := svc.ListLines(ts.ID)
	if err != nil {
		t.Fatalf("failed to list lines: %v", err)
	}
	if len(lines) != 2 {
		t.Fatalf("expected 2 populated lines from time_entries, got %d", len(lines))
	}

	// 7200s = 120 mins, 5400s = 90 mins -> Total = 210 mins
	if ts.TotalMinutes != 210 {
		t.Fatalf("expected 210 total minutes, got %d", ts.TotalMinutes)
	}
}

func TestTimesheetService_SubmissionAndApprovalWorkflow(t *testing.T) {
	svc, _, dbConn := setupTestTimesheetSvc(t)
	defer dbConn.Close()

	ts, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
	})
	if err != nil {
		t.Fatalf("failed to create timesheet: %v", err)
	}

	taskID := int64(501)
	_, err = svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-02",
		DurationMinutes: 120,
	})
	if err != nil {
		t.Fatalf("failed to add line: %v", err)
	}

	// 1. Unauthorized submit attempt
	_, err = svc.SubmitTimesheet(ts.ID, "user_2", model.SubmitTimesheetRequest{
		Comments: "Submitting someone else's sheet",
	})
	if !errors.Is(err, ErrUnauthorizedTimesheet) {
		t.Fatalf("expected ErrUnauthorizedTimesheet, got %v", err)
	}

	// 2. Successful submit by owner
	submitted, err := svc.SubmitTimesheet(ts.ID, "user_1", model.SubmitTimesheetRequest{
		Comments: "Ready for review",
	})
	if err != nil {
		t.Fatalf("failed to submit timesheet: %v", err)
	}
	if submitted.Status != model.TimesheetStatusSubmitted || submitted.SubmittedAt == nil {
		t.Fatalf("expected submitted status and timestamp, got %+v", submitted)
	}

	// 3. Modifying submitted timesheet is forbidden
	_, err = svc.AddLine(ts.ID, "user_1", model.TimesheetLineRequest{
		TaskID:          &taskID,
		WorkDate:        "2026-11-03",
		DurationMinutes: 60,
	})
	if !errors.Is(err, ErrTimesheetNotEditable) {
		t.Fatalf("expected ErrTimesheetNotEditable, got %v", err)
	}

	// 4. Reject without reason is forbidden
	_, err = svc.RejectTimesheet(ts.ID, "user_mgr", model.ReviewTimesheetRequest{
		RejectionReason: "",
		Comments:        "",
	})
	if !errors.Is(err, ErrRejectionReasonRequired) {
		t.Fatalf("expected ErrRejectionReasonRequired, got %v", err)
	}

	// 5. Reject with reason
	rejected, err := svc.RejectTimesheet(ts.ID, "user_mgr", model.ReviewTimesheetRequest{
		RejectionReason: "Missing documentation for sprint items",
	})
	if err != nil {
		t.Fatalf("failed to reject timesheet: %v", err)
	}
	if rejected.Status != model.TimesheetStatusRejected || rejected.RejectionReason == "" {
		t.Fatalf("expected rejected status with reason, got %+v", rejected)
	}

	// 6. Reopen timesheet back to draft
	reopened, err := svc.ReopenTimesheet(ts.ID, "user_mgr", "Reopened for updates")
	if err != nil {
		t.Fatalf("failed to reopen timesheet: %v", err)
	}
	if reopened.Status != model.TimesheetStatusDraft {
		t.Fatalf("expected draft status after reopening, got %s", reopened.Status)
	}

	// 7. Submit again and approve
	_, err = svc.SubmitTimesheet(ts.ID, "user_1", model.SubmitTimesheetRequest{})
	if err != nil {
		t.Fatalf("failed to resubmit: %v", err)
	}

	approved, err := svc.ApproveTimesheet(ts.ID, "user_mgr", model.ReviewTimesheetRequest{
		Comments: "Looks great, approved!",
	})
	if err != nil {
		t.Fatalf("failed to approve timesheet: %v", err)
	}
	if approved.Status != model.TimesheetStatusApproved || approved.ReviewedAt == nil {
		t.Fatalf("expected approved status and timestamp, got %+v", approved)
	}

	// 8. Verify audit history
	approvals, err := svc.ListApprovals(ts.ID)
	if err != nil {
		t.Fatalf("failed to list approvals: %v", err)
	}
	if len(approvals) != 5 { // submit -> reject -> reopen -> submit -> approve
		t.Fatalf("expected 5 approval events, got %d", len(approvals))
	}
}

func TestTimesheetService_ListAndMetrics(t *testing.T) {
	svc, _, dbConn := setupTestTimesheetSvc(t)
	defer dbConn.Close()

	// Create 2 timesheets
	ts1, err := svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-02",
		PeriodEndDate:   "2026-11-08",
		AutoPopulate:    true,
	})
	if err != nil {
		t.Fatalf("failed to create timesheet 1: %v", err)
	}
	_, _ = svc.SubmitTimesheet(ts1.ID, "user_1", model.SubmitTimesheetRequest{})
	_, _ = svc.ApproveTimesheet(ts1.ID, "user_mgr", model.ReviewTimesheetRequest{})

	_, err = svc.CreateTimesheet("user_1", model.CreateTimesheetRequest{
		WorkspaceID:     1,
		PeriodStartDate: "2026-11-09",
		PeriodEndDate:   "2026-11-15",
	})
	if err != nil {
		t.Fatalf("failed to create timesheet 2: %v", err)
	}

	// Filter by workspace
	list, err := svc.ListTimesheets(model.TimesheetFilter{
		WorkspaceID: 1,
	})
	if err != nil {
		t.Fatalf("failed to list timesheets: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 timesheets, got %d", len(list))
	}

	// Filter by status=approved
	approvedStatus := model.TimesheetStatusApproved
	approvedList, err := svc.ListTimesheets(model.TimesheetFilter{
		WorkspaceID: 1,
		Status:      &approvedStatus,
	})
	if err != nil {
		t.Fatalf("failed to list approved timesheets: %v", err)
	}
	if len(approvedList) != 1 {
		t.Fatalf("expected 1 approved timesheet, got %d", len(approvedList))
	}

	// Metrics
	metrics, err := svc.GetTimesheetMetrics(1)
	if err != nil {
		t.Fatalf("failed to get metrics: %v", err)
	}
	if metrics.TotalTimesheets != 2 || metrics.ApprovedCount != 1 || metrics.DraftCount != 1 {
		t.Fatalf("metrics mismatch: %+v", metrics)
	}
	if metrics.TotalLoggedMinutes != 210 {
		t.Fatalf("expected 210 logged minutes in metrics, got %d", metrics.TotalLoggedMinutes)
	}
}
