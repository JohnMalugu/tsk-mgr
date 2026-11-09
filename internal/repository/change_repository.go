package repository

import (
	"database/sql"
	"errors"

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

func (r *changeRepositoryImpl) GetNextChangeNumber(workspaceID int64) (int, error) {
	query := `SELECT COALESCE(MAX(change_number), 0) + 1 FROM change_requests WHERE workspace_id = ?`
	var nextNum int
	err := r.db.QueryRow(query, workspaceID).Scan(&nextNum)
	if err != nil {
		return 1, err
	}
	return nextNum, nil
}

func (r *changeRepositoryImpl) CreateChange(change *model.ChangeRequest) (int64, error) {
	return 0, nil
}

func (r *changeRepositoryImpl) GetChangeByID(id int64) (*model.ChangeRequest, error) {
	return nil, nil
}

func (r *changeRepositoryImpl) GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error) {
	return nil, nil
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
