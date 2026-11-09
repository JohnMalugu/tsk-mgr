package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrChangeNotFound          = repository.ErrChangeNotFound
	ErrApprovalNotFound        = repository.ErrApprovalNotFound
	ErrPIRNotFound             = repository.ErrPIRNotFound
	ErrChangeTitleRequired     = errors.New("change title is required")
	ErrInvalidChangeType       = model.ErrInvalidChangeType
	ErrInvalidChangeRiskLevel  = model.ErrInvalidChangeRiskLevel
	ErrInvalidChangeStatus     = model.ErrInvalidChangeStatus
	ErrInvalidChangeTransition = model.ErrInvalidChangeTransition
	ErrChangeNotEditable       = errors.New("cannot update change request in current status")
	ErrReviewerRequired        = errors.New("reviewer_id is required")
	ErrInvalidCABDecision      = errors.New("invalid CAB decision")
	ErrInvalidPIROutcome       = errors.New("invalid PIR outcome")
	ErrPIRNotesRequired        = errors.New("PIR review notes are required")
	ErrTaskAlreadyLinked       = errors.New("task is already linked to this change request")
	ErrInvalidChangePhase      = errors.New("invalid implementation phase")
)

// ChangeService manages ITIL change lifecycle, CAB reviews, tasks, PIR, and performance metrics
type ChangeService interface {
	CreateChange(requesterID string, req model.CreateChangeRequest) (*model.ChangeRequest, error)
	GetChange(id int64) (*model.ChangeRequest, error)
	GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error)
	ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error)
	UpdateChange(id int64, userID string, req model.UpdateChangeRequest) (*model.ChangeRequest, error)
	DeleteChange(id int64, userID string) error
	ChangeStatus(id int64, userID string, req model.ChangeStatusRequest) (*model.ChangeRequest, error)

	AddReviewer(changeID int64, userID string, req model.AddReviewerRequest) (*model.ChangeApproval, error)
	ReviewChange(changeID int64, reviewerID string, req model.ChangeApprovalRequest) (*model.ChangeApproval, error)
	ListApprovals(changeID int64) ([]model.ChangeApproval, error)
	DeleteReviewer(approvalID int64, userID string) error

	LinkTask(changeID int64, req model.LinkChangeTaskRequest) (*model.ChangeTaskLink, error)
	UnlinkTask(changeID int64, taskID int) error
	ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error)

	SubmitPIR(changeID int64, reviewerID string, req model.ChangePIRRequest) (*model.ChangePIR, error)
	GetPIR(changeID int64) (*model.ChangePIR, error)

	GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error)
}

type changeServiceImpl struct {
	repo          repository.ChangeRepository
	workspaceRepo repository.WorkspaceRepository
	projectRepo   repository.ProjectRepository
	taskRepo      repository.TaskRepository
	userRepo      repository.UserRepository
	auditService  AuditService
}

// NewChangeService creates a new instance of ChangeService
func NewChangeService(
	repo repository.ChangeRepository,
	workspaceRepo repository.WorkspaceRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
	userRepo repository.UserRepository,
	auditService AuditService,
) ChangeService {
	return &changeServiceImpl{
		repo:          repo,
		workspaceRepo: workspaceRepo,
		projectRepo:   projectRepo,
		taskRepo:      taskRepo,
		userRepo:      userRepo,
		auditService:  auditService,
	}
}

func (s *changeServiceImpl) CreateChange(requesterID string, req model.CreateChangeRequest) (*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetChange(id int64) (*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetChangeByNumber(workspaceID int64, changeNumber int) (*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) ListChanges(filter model.ChangeFilter) ([]*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) UpdateChange(id int64, userID string, req model.UpdateChangeRequest) (*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) DeleteChange(id int64, userID string) error {
	return nil
}

func (s *changeServiceImpl) ChangeStatus(id int64, userID string, req model.ChangeStatusRequest) (*model.ChangeRequest, error) {
	return nil, nil
}

func (s *changeServiceImpl) AddReviewer(changeID int64, userID string, req model.AddReviewerRequest) (*model.ChangeApproval, error) {
	return nil, nil
}

func (s *changeServiceImpl) ReviewChange(changeID int64, reviewerID string, req model.ChangeApprovalRequest) (*model.ChangeApproval, error) {
	return nil, nil
}

func (s *changeServiceImpl) ListApprovals(changeID int64) ([]model.ChangeApproval, error) {
	return nil, nil
}

func (s *changeServiceImpl) DeleteReviewer(approvalID int64, userID string) error {
	return nil
}

func (s *changeServiceImpl) LinkTask(changeID int64, req model.LinkChangeTaskRequest) (*model.ChangeTaskLink, error) {
	return nil, nil
}

func (s *changeServiceImpl) UnlinkTask(changeID int64, taskID int) error {
	return nil
}

func (s *changeServiceImpl) ListLinkedTasks(changeID int64) ([]model.ChangeTaskLink, error) {
	return nil, nil
}

func (s *changeServiceImpl) SubmitPIR(changeID int64, reviewerID string, req model.ChangePIRRequest) (*model.ChangePIR, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetPIR(changeID int64) (*model.ChangePIR, error) {
	return nil, nil
}

func (s *changeServiceImpl) GetChangeMetrics(workspaceID int64) (*model.ChangeMetrics, error) {
	return nil, nil
}
