package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrChangeNotFound   = errors.New("change request not found")
	ErrApprovalNotFound = errors.New("change approval not found")
	ErrPIRNotFound      = errors.New("change PIR not found")
)

// ChangeRepository manages database operations for ITIL change requests, CAB approvals, tasks, PIRs, and metrics.
type ChangeRepository interface {
	GetNextChangeNumber(workspaceID int64) (int, error)
	CreateChange(change *model.ChangeRequest) (int64, error)
	GetChangeByID(id int64) (*model.ChangeRequest, error)
	GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error)
	ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error)
	UpdateChange(change *model.ChangeRequest) error
	DeleteChange(id int64) error

	AddReviewer(approval *model.ChangeApproval) (int64, error)
	GetApprovalByID(id int64) (*model.ChangeApproval, error)
	GetApprovalByChangeAndReviewer(changeID int64, reviewerID string) (*model.ChangeApproval, error)
	CastDecision(changeID int64, reviewerID string, decision string, comments string) error
	ListApprovals(changeID int64) ([]model.ChangeApproval, error)
	DeleteReviewer(approvalID int64) error

	LinkTask(link *model.ChangeTaskLink) (int64, error)
	UnlinkTask(changeID int64, taskID int) error
	ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error)

	UpsertPIR(pir *model.ChangePIR) (int64, error)
	GetPIR(changeID int64) (*model.ChangePIR, error)

	GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error)
}

type changeRepositoryImpl struct {
	db *sql.DB
}

// NewChangeRepository creates a new instance of ChangeRepository.
func NewChangeRepository(db *sql.DB) ChangeRepository {
	return &changeRepositoryImpl{db: db}
}

const baseChangeSelect = `
	SELECT 
		c.id, c.workspace_id, c.project_id, c.change_number, c.title, c.description,
		c.change_type, c.risk_level, c.impact_level, c.status,
		c.reason_for_change, c.implementation_plan, c.rollback_plan, c.test_plan,
		c.requested_by_user_id, c.assigned_to_user_id,
		c.planned_start_at, c.planned_end_at, c.actual_start_at, c.actual_end_at,
		c.created_at, c.updated_at,
		COALESCE(u_req.username, u_req.email, '') AS requester_name,
		COALESCE(u_asg.username, u_asg.email, '') AS assignee_name,
		COALESCE(p.name, '') AS project_name,
		(SELECT COUNT(*) FROM change_approvals WHERE change_id = c.id) AS approvals_count,
		(SELECT COUNT(*) FROM change_approvals WHERE change_id = c.id AND decision = 'approved') AS approved_count,
		(SELECT COUNT(*) FROM change_task_links WHERE change_id = c.id) AS tasks_count
	FROM change_requests c
	LEFT JOIN users u_req ON c.requested_by_user_id = u_req.id
	LEFT JOIN users u_asg ON c.assigned_to_user_id = u_asg.id
	LEFT JOIN projects p ON c.project_id = p.id
`

func scanChange(scanner interface{ Scan(dest ...interface{}) error }) (*model.ChangeRequest, error) {
	var c model.ChangeRequest
	var projectID sql.NullInt64
	var assignedTo sql.NullString
	var plannedStart, plannedEnd, actualStart, actualEnd sql.NullTime
	var requesterName, assigneeName, projectName sql.NullString
	var approvalsCount, approvedCount, tasksCount sql.NullInt64

	err := scanner.Scan(
		&c.ID, &c.WorkspaceID, &projectID, &c.ChangeNumber, &c.Title, &c.Description,
		&c.ChangeType, &c.RiskLevel, &c.ImpactLevel, &c.Status,
		&c.ReasonForChange, &c.ImplementationPlan, &c.RollbackPlan, &c.TestPlan,
		&c.RequestedByUserID, &assignedTo,
		&plannedStart, &plannedEnd, &actualStart, &actualEnd,
		&c.CreatedAt, &c.UpdatedAt,
		&requesterName, &assigneeName, &projectName,
		&approvalsCount, &approvedCount, &tasksCount,
	)
	if err != nil {
		return nil, err
	}
	if projectID.Valid {
		c.ProjectID = &projectID.Int64
	}
	if assignedTo.Valid {
		c.AssignedToUserID = assignedTo.String
	}
	if plannedStart.Valid {
		t := plannedStart.Time
		c.PlannedStartAt = &t
	}
	if plannedEnd.Valid {
		t := plannedEnd.Time
		c.PlannedEndAt = &t
	}
	if actualStart.Valid {
		t := actualStart.Time
		c.ActualStartAt = &t
	}
	if actualEnd.Valid {
		t := actualEnd.Time
		c.ActualEndAt = &t
	}
	if requesterName.Valid {
		c.RequesterName = requesterName.String
	}
	if assigneeName.Valid {
		c.AssigneeName = assigneeName.String
	}
	if projectName.Valid {
		c.ProjectName = projectName.String
	}
	if approvalsCount.Valid {
		c.ApprovalsCount = int(approvalsCount.Int64)
	}
	if approvedCount.Valid {
		c.ApprovedCount = int(approvedCount.Int64)
	}
	if tasksCount.Valid {
		c.TasksCount = int(tasksCount.Int64)
	}
	return &c, nil
}

func (r *changeRepositoryImpl) GetNextChangeNumber(workspaceID int64) (int, error) {
	query := `SELECT COALESCE(MAX(CAST(SUBSTR(change_number, 5) AS INTEGER)), 0) + 1 FROM change_requests WHERE workspace_id = ?`
	var nextNum int
	err := r.db.QueryRow(query, workspaceID).Scan(&nextNum)
	if err != nil {
		return 1, err
	}
	return nextNum, nil
}

func (r *changeRepositoryImpl) CreateChange(change *model.ChangeRequest) (int64, error) {
	if change.ChangeNumber == "" {
		nextNum, err := r.GetNextChangeNumber(change.WorkspaceID)
		if err != nil {
			return 0, err
		}
		change.ChangeNumber = fmt.Sprintf("CHG-%d", nextNum)
	}
	now := time.Now().UTC()
	if change.CreatedAt.IsZero() {
		change.CreatedAt = now
	}
	if change.UpdatedAt.IsZero() {
		change.UpdatedAt = now
	}
	if change.Status == "" {
		change.Status = model.ChangeStatusDraft
	}
	if change.ChangeType == "" {
		change.ChangeType = model.ChangeTypeNormal
	}
	if change.RiskLevel == "" {
		change.RiskLevel = model.ChangeRiskMedium
	}
	if change.ImpactLevel == "" {
		change.ImpactLevel = model.ChangeImpactLow
	}

	query := `
	INSERT INTO change_requests (
		workspace_id, project_id, change_number, title, description,
		change_type, risk_level, impact_level, status,
		reason_for_change, implementation_plan, rollback_plan, test_plan,
		requested_by_user_id, assigned_to_user_id,
		planned_start_at, planned_end_at, actual_start_at, actual_end_at,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		change.WorkspaceID, change.ProjectID, change.ChangeNumber, change.Title, change.Description,
		change.ChangeType, change.RiskLevel, change.ImpactLevel, change.Status,
		change.ReasonForChange, change.ImplementationPlan, change.RollbackPlan, change.TestPlan,
		change.RequestedByUserID, change.AssignedToUserID,
		change.PlannedStartAt, change.PlannedEndAt, change.ActualStartAt, change.ActualEndAt,
		change.CreatedAt, change.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	change.ID = id
	return id, nil
}

func (r *changeRepositoryImpl) GetChangeByID(id int64) (*model.ChangeRequest, error) {
	query := baseChangeSelect + ` WHERE c.id = ?`
	row := r.db.QueryRow(query, id)
	change, err := scanChange(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}
	return change, nil
}

func (r *changeRepositoryImpl) GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error) {
	strNum := fmt.Sprintf("CHG-%d", changeNumber)
	query := baseChangeSelect + ` WHERE c.workspace_id = ? AND (c.change_number = ? OR c.change_number = ?)`
	row := r.db.QueryRow(query, workspaceID, strNum, fmt.Sprintf("%d", changeNumber))
	change, err := scanChange(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChangeNotFound
		}
		return nil, err
	}
	return change, nil
}

func (r *changeRepositoryImpl) ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error) {
	var conditions []string
	var args []interface{}

	if filter.WorkspaceID > 0 {
		conditions = append(conditions, "c.workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.ProjectID != nil {
		conditions = append(conditions, "c.project_id = ?")
		args = append(args, *filter.ProjectID)
	}
	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, "c.status = ?")
		args = append(args, *filter.Status)
	}
	if filter.ChangeType != nil && *filter.ChangeType != "" {
		conditions = append(conditions, "c.change_type = ?")
		args = append(args, *filter.ChangeType)
	}
	if filter.RiskLevel != nil && *filter.RiskLevel != "" {
		conditions = append(conditions, "c.risk_level = ?")
		args = append(args, *filter.RiskLevel)
	}
	if filter.RequestedBy != nil && *filter.RequestedBy != "" {
		conditions = append(conditions, "c.requested_by_user_id = ?")
		args = append(args, *filter.RequestedBy)
	}
	if filter.AssignedTo != nil && *filter.AssignedTo != "" {
		conditions = append(conditions, "c.assigned_to_user_id = ?")
		args = append(args, *filter.AssignedTo)
	}
	if filter.Search != nil && *filter.Search != "" {
		pattern := "%" + *filter.Search + "%"
		conditions = append(conditions, "(c.title LIKE ? OR c.description LIKE ? OR c.change_number LIKE ?)")
		args = append(args, pattern, pattern, pattern)
	}
	if filter.StartDate != nil && *filter.StartDate != "" {
		conditions = append(conditions, "DATE(c.created_at) >= ?")
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil && *filter.EndDate != "" {
		conditions = append(conditions, "DATE(c.created_at) <= ?")
		args = append(args, *filter.EndDate)
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	query := baseChangeSelect + whereClause + " ORDER BY c.created_at DESC"
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var changes []*model.ChangeRequest
	for rows.Next() {
		change, err := scanChange(rows)
		if err != nil {
			return nil, err
		}
		changes = append(changes, change)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return changes, nil
}

func (r *changeRepositoryImpl) UpdateChange(change *model.ChangeRequest) error {
	now := time.Now().UTC()
	change.UpdatedAt = now

	query := `
	UPDATE change_requests SET
		project_id = ?,
		title = ?,
		description = ?,
		change_type = ?,
		risk_level = ?,
		impact_level = ?,
		status = ?,
		reason_for_change = ?,
		implementation_plan = ?,
		rollback_plan = ?,
		test_plan = ?,
		assigned_to_user_id = ?,
		planned_start_at = ?,
		planned_end_at = ?,
		actual_start_at = ?,
		actual_end_at = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.Exec(query,
		change.ProjectID,
		change.Title,
		change.Description,
		change.ChangeType,
		change.RiskLevel,
		change.ImpactLevel,
		change.Status,
		change.ReasonForChange,
		change.ImplementationPlan,
		change.RollbackPlan,
		change.TestPlan,
		change.AssignedToUserID,
		change.PlannedStartAt,
		change.PlannedEndAt,
		change.ActualStartAt,
		change.ActualEndAt,
		change.UpdatedAt,
		change.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrChangeNotFound
	}
	return nil
}

func (r *changeRepositoryImpl) DeleteChange(id int64) error {
	res, err := r.db.Exec("DELETE FROM change_requests WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrChangeNotFound
	}
	return nil
}

const baseApprovalSelect = `
	SELECT 
		a.id, a.change_id, a.reviewer_id, a.decision, a.comments, a.decided_at, a.created_at,
		COALESCE(u.username, '') AS reviewer_name,
		COALESCE(u.email, '') AS reviewer_email
	FROM change_approvals a
	LEFT JOIN users u ON a.reviewer_id = u.id
`

func scanApproval(scanner interface{ Scan(dest ...interface{}) error }) (*model.ChangeApproval, error) {
	var a model.ChangeApproval
	var decidedAt sql.NullTime
	var name, email sql.NullString

	err := scanner.Scan(
		&a.ID, &a.ChangeID, &a.ReviewerID, &a.Decision, &a.Comments, &decidedAt, &a.CreatedAt,
		&name, &email,
	)
	if err != nil {
		return nil, err
	}
	if decidedAt.Valid {
		t := decidedAt.Time
		a.DecidedAt = &t
	}
	if name.Valid {
		a.ReviewerName = name.String
	}
	if email.Valid {
		a.ReviewerEmail = email.String
	}
	return &a, nil
}

func (r *changeRepositoryImpl) AddReviewer(approval *model.ChangeApproval) (int64, error) {
	if approval.CreatedAt.IsZero() {
		approval.CreatedAt = time.Now().UTC()
	}
	if approval.Decision == "" {
		approval.Decision = model.ApprovalDecisionPending
	}

	query := `
	INSERT INTO change_approvals (
		change_id, reviewer_id, decision, comments, decided_at, created_at
	) VALUES (?, ?, ?, ?, ?, ?)`

	res, err := r.db.Exec(query,
		approval.ChangeID, approval.ReviewerID, approval.Decision,
		approval.Comments, approval.DecidedAt, approval.CreatedAt,
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

func (r *changeRepositoryImpl) GetApprovalByID(id int64) (*model.ChangeApproval, error) {
	query := baseApprovalSelect + ` WHERE a.id = ?`
	row := r.db.QueryRow(query, id)
	approval, err := scanApproval(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrApprovalNotFound
		}
		return nil, err
	}
	return approval, nil
}

func (r *changeRepositoryImpl) GetApprovalByChangeAndReviewer(changeID int64, reviewerID string) (*model.ChangeApproval, error) {
	query := baseApprovalSelect + ` WHERE a.change_id = ? AND a.reviewer_id = ?`
	row := r.db.QueryRow(query, changeID, reviewerID)
	approval, err := scanApproval(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrApprovalNotFound
		}
		return nil, err
	}
	return approval, nil
}

func (r *changeRepositoryImpl) CastDecision(changeID int64, reviewerID string, decision string, comments string) error {
	now := time.Now().UTC()
	query := `
	UPDATE change_approvals SET
		decision = ?,
		comments = ?,
		decided_at = ?
	WHERE change_id = ? AND reviewer_id = ?`

	res, err := r.db.Exec(query, decision, comments, now, changeID, reviewerID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrApprovalNotFound
	}
	return nil
}

func (r *changeRepositoryImpl) ListApprovals(changeID int64) ([]model.ChangeApproval, error) {
	query := baseApprovalSelect + ` WHERE a.change_id = ? ORDER BY a.created_at ASC`
	rows, err := r.db.Query(query, changeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var approvals []model.ChangeApproval
	for rows.Next() {
		appr, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		approvals = append(approvals, *appr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return approvals, nil
}

func (r *changeRepositoryImpl) DeleteReviewer(approvalID int64) error {
	res, err := r.db.Exec("DELETE FROM change_approvals WHERE id = ?", approvalID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrApprovalNotFound
	}
	return nil
}

func (r *changeRepositoryImpl) LinkTask(link *model.ChangeTaskLink) (int64, error) {
	return 0, nil
}

func (r *changeRepositoryImpl) UnlinkTask(changeID int64, taskID int) error {
	return nil
}

func (r *changeRepositoryImpl) ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) UpsertPIR(pir *model.ChangePIR) (int64, error) {
	return 0, nil
}

func (r *changeRepositoryImpl) GetPIR(changeID int64) (*model.ChangePIR, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error) {
	return nil, nil
}
