package repository

import (
	"database/sql"
	"errors"
	"fmt"
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
	return nil, nil
}

func (r *changeRepositoryImpl) UpdateChange(change *model.ChangeRequest) error {
	return nil
}

func (r *changeRepositoryImpl) DeleteChange(id int64) error {
	return nil
}

func (r *changeRepositoryImpl) AddReviewer(approval *model.ChangeApproval) (int64, error) {
	return 0, nil
}

func (r *changeRepositoryImpl) GetApprovalByID(id int64) (*model.ChangeApproval, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) GetApprovalByChangeAndReviewer(changeID int64, reviewerID string) (*model.ChangeApproval, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) CastDecision(changeID int64, reviewerID string, decision string, comments string) error {
	return nil
}

func (r *changeRepositoryImpl) ListApprovals(changeID int64) ([]model.ChangeApproval, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) DeleteReviewer(approvalID int64) error {
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
