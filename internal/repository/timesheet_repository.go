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

func (r *timesheetRepositoryImpl) AddTimesheetLine(line *model.TimesheetLine) (int64, error) {
	return 0, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetLineByID(id int64) (*model.TimesheetLine, error) {
	return nil, nil
}

func (r *timesheetRepositoryImpl) UpdateTimesheetLine(line *model.TimesheetLine) error {
	return nil
}

func (r *timesheetRepositoryImpl) DeleteTimesheetLine(id int64) error {
	return nil
}

func (r *timesheetRepositoryImpl) ListTimesheetLines(timesheetID int64) ([]model.TimesheetLine, error) {
	return nil, nil
}

func (r *timesheetRepositoryImpl) RecordApproval(approval *model.TimesheetApproval) (int64, error) {
	return 0, nil
}

func (r *timesheetRepositoryImpl) ListApprovals(timesheetID int64) ([]model.TimesheetApproval, error) {
	return nil, nil
}

func (r *timesheetRepositoryImpl) GetTimesheetMetrics(workspaceID int64) (*model.TimesheetMetrics, error) {
	return nil, nil
}

func (r *timesheetRepositoryImpl) RecalculateTotals(timesheetID int64) error {
	return nil
}
