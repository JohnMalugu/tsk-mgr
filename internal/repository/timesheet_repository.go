package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrTimesheetNotFound     = errors.New("timesheet not found")
	ErrTimesheetLineNotFound = errors.New("timesheet line not found")
)

const timesheetBaseSelect = `
	SELECT 
		ts.id,
		ts.workspace_id,
		ts.user_id,
		ts.period_start_date,
		ts.period_end_date,
		ts.status,
		ts.total_minutes,
		ts.billable_minutes,
		ts.submitted_at,
		ts.reviewed_at,
		ts.reviewer_id,
		ts.rejection_reason,
		ts.notes,
		ts.created_at,
		ts.updated_at,
		u.username AS username,
		rev.username AS reviewer_username,
		(SELECT COUNT(*) FROM timesheet_lines tl WHERE tl.timesheet_id = ts.id) AS lines_count
	FROM timesheets ts
	LEFT JOIN users u ON ts.user_id = u.id
	LEFT JOIN users rev ON ts.reviewer_id = rev.id
`

func scanTimesheet(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.Timesheet, error) {
	var ts model.Timesheet
	var submittedAt, reviewedAt sql.NullTime
	var reviewerID sql.NullString
	var username, reviewerUsername sql.NullString

	err := scanner.Scan(
		&ts.ID,
		&ts.WorkspaceID,
		&ts.UserID,
		&ts.PeriodStartDate,
		&ts.PeriodEndDate,
		&ts.Status,
		&ts.TotalMinutes,
		&ts.BillableMinutes,
		&submittedAt,
		&reviewedAt,
		&reviewerID,
		&ts.RejectionReason,
		&ts.Notes,
		&ts.CreatedAt,
		&ts.UpdatedAt,
		&username,
		&reviewerUsername,
		&ts.LinesCount,
	)
	if err != nil {
		return nil, err
	}
	if submittedAt.Valid {
		ts.SubmittedAt = &submittedAt.Time
	}
	if reviewedAt.Valid {
		ts.ReviewedAt = &reviewedAt.Time
	}
	if reviewerID.Valid {
		ts.ReviewerID = &reviewerID.String
	}
	if username.Valid {
		ts.Username = username.String
	}
	if reviewerUsername.Valid {
		ts.ReviewerUsername = reviewerUsername.String
	}

	return &ts, nil
}

// TimesheetRepository manages database operations for timesheets, line entries, and approvals
type TimesheetRepository interface {
	CreateTimesheet(ts *model.Timesheet) (int64, error)
	GetTimesheetByID(id int64) (*model.Timesheet, error)
	GetTimesheetByPeriod(workspaceID int64, userID string, startDate string) (*model.Timesheet, error)
	ListTimesheets(filter model.TimesheetFilter) ([]*model.Timesheet, error)
	UpdateTimesheet(ts *model.Timesheet) error
	DeleteTimesheet(id int64) error

	AddTimesheetLine(line *model.TimesheetLine) (int64, error)
	GetTimesheetLineByID(id int64) (*model.TimesheetLine, error)
	UpdateTimesheetLine(line *model.TimesheetLine) error
	DeleteTimesheetLine(id int64) error
	ListTimesheetLines(timesheetID int64) ([]model.TimesheetLine, error)

	RecordApproval(approval *model.TimesheetApproval) (int64, error)
	ListApprovals(timesheetID int64) ([]model.TimesheetApproval, error)

	GetTimesheetMetrics(workspaceID int64) (*model.TimesheetMetrics, error)
	RecalculateTotals(timesheetID int64) error
}

type timesheetRepositoryImpl struct {
	db *sql.DB
}

// NewTimesheetRepository creates a new instance of TimesheetRepository
func NewTimesheetRepository(db *sql.DB) TimesheetRepository {
	return &timesheetRepositoryImpl{db: db}
}

func (r *timesheetRepositoryImpl) CreateTimesheet(ts *model.Timesheet) (int64, error) {
	now := time.Now()
	ts.CreatedAt = now
	ts.UpdatedAt = now
	if ts.Status == "" {
		ts.Status = model.TimesheetStatusDraft
	}

	query := `
	INSERT INTO timesheets (
		workspace_id, user_id, period_start_date, period_end_date, status,
		total_minutes, billable_minutes, submitted_at, reviewed_at, reviewer_id,
		rejection_reason, notes, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		ts.WorkspaceID,
		ts.UserID,
		ts.PeriodStartDate,
		ts.PeriodEndDate,
		ts.Status,
		ts.TotalMinutes,
		ts.BillableMinutes,
		ts.SubmittedAt,
		ts.ReviewedAt,
		ts.ReviewerID,
		ts.RejectionReason,
		ts.Notes,
		ts.CreatedAt,
		ts.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	ts.ID = id
	return id, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetByID(id int64) (*model.Timesheet, error) {
	query := timesheetBaseSelect + ` WHERE ts.id = ?`
	row := r.db.QueryRow(query, id)
	ts, err := scanTimesheet(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTimesheetNotFound
		}
		return nil, err
	}
	return ts, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetByPeriod(workspaceID int64, userID string, startDate string) (*model.Timesheet, error) {
	query := timesheetBaseSelect + ` WHERE ts.workspace_id = ? AND ts.user_id = ? AND ts.period_start_date = ?`
	row := r.db.QueryRow(query, workspaceID, userID, startDate)
	ts, err := scanTimesheet(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTimesheetNotFound
		}
		return nil, err
	}
	return ts, nil
}

func (r *timesheetRepositoryImpl) ListTimesheets(filter model.TimesheetFilter) ([]*model.Timesheet, error) {
	query := timesheetBaseSelect + ` WHERE 1=1`
	var args []interface{}

	if filter.WorkspaceID > 0 {
		query += ` AND ts.workspace_id = ?`
		args = append(args, filter.WorkspaceID)
	}
	if filter.UserID != nil && *filter.UserID != "" {
		query += ` AND ts.user_id = ?`
		args = append(args, *filter.UserID)
	}
	if filter.Status != nil && *filter.Status != "" {
		query += ` AND ts.status = ?`
		args = append(args, *filter.Status)
	}
	if filter.StartDate != nil && *filter.StartDate != "" {
		query += ` AND ts.period_start_date >= ?`
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		query += ` AND ts.period_end_date <= ?`
		args = append(args, *filter.EndDate)
	}

	query += ` ORDER BY ts.period_start_date DESC, ts.created_at DESC`

	if filter.Limit > 0 {
		query += ` LIMIT ?`
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += ` OFFSET ?`
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var timesheets []*model.Timesheet
	for rows.Next() {
		ts, err := scanTimesheet(rows)
		if err != nil {
			return nil, err
		}
		timesheets = append(timesheets, ts)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return timesheets, nil
}

func (r *timesheetRepositoryImpl) UpdateTimesheet(ts *model.Timesheet) error {
	ts.UpdatedAt = time.Now()
	query := `
	UPDATE timesheets
	SET status = ?,
		total_minutes = ?,
		billable_minutes = ?,
		submitted_at = ?,
		reviewed_at = ?,
		reviewer_id = ?,
		rejection_reason = ?,
		notes = ?,
		updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(
		query,
		ts.Status,
		ts.TotalMinutes,
		ts.BillableMinutes,
		ts.SubmittedAt,
		ts.ReviewedAt,
		ts.ReviewerID,
		ts.RejectionReason,
		ts.Notes,
		ts.UpdatedAt,
		ts.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTimesheetNotFound
	}
	return nil
}

func (r *timesheetRepositoryImpl) DeleteTimesheet(id int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM timesheet_approvals WHERE timesheet_id = ?`, id); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM timesheet_lines WHERE timesheet_id = ?`, id); err != nil {
		return err
	}

	res, err := tx.Exec(`DELETE FROM timesheets WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTimesheetNotFound
	}

	return tx.Commit()
}

const timesheetLineBaseSelect = `
	SELECT 
		tl.id,
		tl.timesheet_id,
		tl.task_id,
		tl.project_id,
		tl.work_date,
		tl.duration_minutes,
		tl.is_billable,
		tl.description,
		tl.created_at,
		tl.updated_at,
		t.title AS task_title,
		p.name AS project_name
	FROM timesheet_lines tl
	LEFT JOIN tasks t ON tl.task_id = t.id
	LEFT JOIN projects p ON tl.project_id = p.id
`

func scanTimesheetLine(scanner interface {
	Scan(dest ...interface{}) error
}) (*model.TimesheetLine, error) {
	var line model.TimesheetLine
	var taskID, projectID sql.NullInt64
	var taskTitle, projectName sql.NullString

	err := scanner.Scan(
		&line.ID,
		&line.TimesheetID,
		&taskID,
		&projectID,
		&line.WorkDate,
		&line.DurationMinutes,
		&line.IsBillable,
		&line.Description,
		&line.CreatedAt,
		&line.UpdatedAt,
		&taskTitle,
		&projectName,
	)
	if err != nil {
		return nil, err
	}
	if taskID.Valid {
		tid := int(taskID.Int64)
		line.TaskID = &tid
	}
	if projectID.Valid {
		pid := int(projectID.Int64)
		line.ProjectID = &pid
	}
	if taskTitle.Valid {
		line.TaskTitle = taskTitle.String
	}
	if projectName.Valid {
		line.ProjectName = projectName.String
	}
	return &line, nil
}

func (r *timesheetRepositoryImpl) AddTimesheetLine(line *model.TimesheetLine) (int64, error) {
	now := time.Now()
	line.CreatedAt = now
	line.UpdatedAt = now

	query := `
	INSERT INTO timesheet_lines (
		timesheet_id, task_id, project_id, work_date, duration_minutes,
		is_billable, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		line.TimesheetID,
		line.TaskID,
		line.ProjectID,
		line.WorkDate,
		line.DurationMinutes,
		line.IsBillable,
		line.Description,
		line.CreatedAt,
		line.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	line.ID = id
	return id, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetLineByID(id int64) (*model.TimesheetLine, error) {
	query := timesheetLineBaseSelect + ` WHERE tl.id = ?`
	row := r.db.QueryRow(query, id)
	line, err := scanTimesheetLine(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrTimesheetLineNotFound
		}
		return nil, err
	}
	return line, nil
}

func (r *timesheetRepositoryImpl) UpdateTimesheetLine(line *model.TimesheetLine) error {
	line.UpdatedAt = time.Now()
	query := `
	UPDATE timesheet_lines
	SET task_id = ?,
		project_id = ?,
		work_date = ?,
		duration_minutes = ?,
		is_billable = ?,
		description = ?,
		updated_at = ?
	WHERE id = ?
	`
	res, err := r.db.Exec(
		query,
		line.TaskID,
		line.ProjectID,
		line.WorkDate,
		line.DurationMinutes,
		line.IsBillable,
		line.Description,
		line.UpdatedAt,
		line.ID,
	)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTimesheetLineNotFound
	}
	return nil
}

func (r *timesheetRepositoryImpl) DeleteTimesheetLine(id int64) error {
	res, err := r.db.Exec(`DELETE FROM timesheet_lines WHERE id = ?`, id)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrTimesheetLineNotFound
	}
	return nil
}

func (r *timesheetRepositoryImpl) ListTimesheetLines(timesheetID int64) ([]model.TimesheetLine, error) {
	query := timesheetLineBaseSelect + ` WHERE tl.timesheet_id = ? ORDER BY tl.work_date ASC, tl.id ASC`
	rows, err := r.db.Query(query, timesheetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lines []model.TimesheetLine
	for rows.Next() {
		line, err := scanTimesheetLine(rows)
		if err != nil {
			return nil, err
		}
		lines = append(lines, *line)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func (r *timesheetRepositoryImpl) RecordApproval(approval *model.TimesheetApproval) (int64, error) {
	if approval.CreatedAt.IsZero() {
		approval.CreatedAt = time.Now()
	}

	query := `
	INSERT INTO timesheet_approvals (
		timesheet_id, actor_id, action, previous_status, new_status, comments, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	res, err := r.db.Exec(
		query,
		approval.TimesheetID,
		approval.ActorID,
		approval.Action,
		approval.PreviousStatus,
		approval.NewStatus,
		approval.Comments,
		approval.CreatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	approval.ID = id
	return id, nil
}

func (r *timesheetRepositoryImpl) ListApprovals(timesheetID int64) ([]model.TimesheetApproval, error) {
	query := `
	SELECT 
		ta.id,
		ta.timesheet_id,
		ta.actor_id,
		ta.action,
		ta.previous_status,
		ta.new_status,
		ta.comments,
		ta.created_at,
		u.username AS actor_username
	FROM timesheet_approvals ta
	LEFT JOIN users u ON ta.actor_id = u.id
	WHERE ta.timesheet_id = ?
	ORDER BY ta.created_at ASC, ta.id ASC
	`
	rows, err := r.db.Query(query, timesheetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []model.TimesheetApproval
	for rows.Next() {
		var a model.TimesheetApproval
		var username sql.NullString
		err := rows.Scan(
			&a.ID,
			&a.TimesheetID,
			&a.ActorID,
			&a.Action,
			&a.PreviousStatus,
			&a.NewStatus,
			&a.Comments,
			&a.CreatedAt,
			&username,
		)
		if err != nil {
			return nil, err
		}
		if username.Valid {
			a.ActorUsername = username.String
		}
		approvals = append(approvals, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetMetrics(workspaceID int64) (*model.TimesheetMetrics, error) {
	metrics := &model.TimesheetMetrics{
		WorkspaceID: workspaceID,
	}

	query := `
	SELECT 
		COUNT(*),
		COALESCE(SUM(CASE WHEN status = 'draft' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'submitted' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'approved' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'rejected' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(total_minutes), 0),
		COALESCE(SUM(billable_minutes), 0)
	FROM timesheets
	WHERE workspace_id = ?
	`
	err := r.db.QueryRow(query, workspaceID).Scan(
		&metrics.TotalTimesheets,
		&metrics.DraftCount,
		&metrics.SubmittedCount,
		&metrics.ApprovedCount,
		&metrics.RejectedCount,
		&metrics.TotalLoggedMinutes,
		&metrics.TotalBillableMinutes,
	)
	if err != nil {
		return nil, err
	}

	metrics.TotalLoggedHours = float64(metrics.TotalLoggedMinutes) / 60.0
	metrics.TotalBillableHours = float64(metrics.TotalBillableMinutes) / 60.0
	if metrics.TotalLoggedMinutes > 0 {
		metrics.BillablePercentage = float64(metrics.TotalBillableMinutes) / float64(metrics.TotalLoggedMinutes) * 100.0
	}

	return metrics, nil
}

func (r *timesheetRepositoryImpl) RecalculateTotals(timesheetID int64) error {
	query := `
	SELECT 
		COALESCE(SUM(duration_minutes), 0),
		COALESCE(SUM(CASE WHEN is_billable = 1 THEN duration_minutes ELSE 0 END), 0)
	FROM timesheet_lines
	WHERE timesheet_id = ?
	`
	var total, billable int
	if err := r.db.QueryRow(query, timesheetID).Scan(&total, &billable); err != nil {
		return err
	}

	updateQuery := `
	UPDATE timesheets
	SET total_minutes = ?, billable_minutes = ?, updated_at = ?
	WHERE id = ?
	`
	_, err := r.db.Exec(updateQuery, total, billable, time.Now(), timesheetID)
	return err
}
