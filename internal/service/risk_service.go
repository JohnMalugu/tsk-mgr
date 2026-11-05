package service

import (
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrRiskNotFound        = errors.New("risk not found")
	ErrRiskTitleRequired   = errors.New("risk title is required")
	ErrInvalidRiskCategory = errors.New("invalid risk category")
	ErrInvalidRiskStatus   = errors.New("invalid risk status")
	ErrInvalidRiskStrategy = errors.New("invalid risk strategy")
	ErrInvalidProbability  = errors.New("invalid probability (must be between 1 and 5)")
	ErrInvalidImpact       = errors.New("invalid impact (must be between 1 and 5)")
	ErrInvalidRiskAction   = errors.New("invalid risk action type")
	ErrReviewCommentReq    = errors.New("review comment is required")
)

// RiskService manages enterprise risk registers, scoring, mitigation tasks, reviews, and metrics
type RiskService interface {
	CreateRisk(creatorID string, req model.CreateRiskRequest) (*model.Risk, error)
	GetRisk(id int64) (*model.Risk, error)
	GetRiskByCode(workspaceID int64, code string) (*model.Risk, error)
	ListRisks(filter model.RiskFilter) ([]*model.Risk, error)
	UpdateRisk(id int64, req model.UpdateRiskRequest) (*model.Risk, error)
	DeleteRisk(id int64) error

	AssessRisk(id int64, reviewerID string, req model.AssessRiskRequest) (*model.Risk, error)
	ChangeStatus(id int64, reviewerID string, req model.ChangeRiskStatusRequest) (*model.Risk, error)

	LinkTask(riskID int64, req model.LinkRiskTaskRequest) (*model.RiskTaskLink, error)
	UnlinkTask(riskID int64, taskID int) error
	ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error)

	CreateReview(riskID int64, reviewerID string, req model.CreateRiskReviewRequest) (*model.RiskReview, error)
	ListReviews(riskID int64) ([]model.RiskReview, error)

	GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error)
}

type riskServiceImpl struct {
	repo        repository.RiskRepository
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository
}

// NewRiskService creates a new instance of RiskService
func NewRiskService(
	repo repository.RiskRepository,
	projectRepo repository.ProjectRepository,
	taskRepo repository.TaskRepository,
) RiskService {
	return &riskServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
		taskRepo:    taskRepo,
	}
}

func (s *riskServiceImpl) CreateRisk(creatorID string, req model.CreateRiskRequest) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) GetRisk(id int64) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) GetRiskByCode(workspaceID int64, code string) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) ListRisks(filter model.RiskFilter) ([]*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) UpdateRisk(id int64, req model.UpdateRiskRequest) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) DeleteRisk(id int64) error {
	return errors.New("not implemented")
}

func (s *riskServiceImpl) AssessRisk(id int64, reviewerID string, req model.AssessRiskRequest) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) ChangeStatus(id int64, reviewerID string, req model.ChangeRiskStatusRequest) (*model.Risk, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) LinkTask(riskID int64, req model.LinkRiskTaskRequest) (*model.RiskTaskLink, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) UnlinkTask(riskID int64, taskID int) error {
	return errors.New("not implemented")
}

func (s *riskServiceImpl) ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) CreateReview(riskID int64, reviewerID string, req model.CreateRiskReviewRequest) (*model.RiskReview, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) ListReviews(riskID int64) ([]model.RiskReview, error) {
	return nil, errors.New("not implemented")
}

func (s *riskServiceImpl) GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error) {
	return nil, errors.New("not implemented")
}
