package service

import (
	"errors"
	"strings"

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

func isValidRiskCategory(cat string) bool {
	switch cat {
	case model.RiskCategoryTechnical,
		model.RiskCategorySchedule,
		model.RiskCategoryFinancial,
		model.RiskCategoryOperational,
		model.RiskCategorySecurity,
		model.RiskCategoryCompliance,
		model.RiskCategoryExternal:
		return true
	default:
		return false
	}
}

func isValidRiskStrategy(strat string) bool {
	switch strat {
	case model.RiskStrategyMitigate,
		model.RiskStrategyAvoid,
		model.RiskStrategyTransfer,
		model.RiskStrategyAccept:
		return true
	default:
		return false
	}
}

func isValidRiskStatus(st string) bool {
	switch st {
	case model.RiskStatusIdentified,
		model.RiskStatusAssessing,
		model.RiskStatusMitigating,
		model.RiskStatusMonitored,
		model.RiskStatusAccepted,
		model.RiskStatusClosed:
		return true
	default:
		return false
	}
}

func (s *riskServiceImpl) CreateRisk(creatorID string, req model.CreateRiskRequest) (*model.Risk, error) {
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrRiskTitleRequired
	}

	if req.ProjectID != nil {
		if _, err := s.projectRepo.GetByID(*req.ProjectID); err != nil {
			return nil, ErrProjectNotFound
		}
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = model.RiskCategoryOperational
	} else if !isValidRiskCategory(category) {
		return nil, ErrInvalidRiskCategory
	}

	probability := req.Probability
	if probability == 0 {
		probability = 3
	} else if probability < 1 || probability > 5 {
		return nil, ErrInvalidProbability
	}

	impact := req.Impact
	if impact == 0 {
		impact = 3
	} else if impact < 1 || impact > 5 {
		return nil, ErrInvalidImpact
	}

	strategy := strings.TrimSpace(req.Strategy)
	if strategy == "" {
		strategy = model.RiskStrategyMitigate
	} else if !isValidRiskStrategy(strategy) {
		return nil, ErrInvalidRiskStrategy
	}

	ownerID := req.OwnerID
	if ownerID == nil || strings.TrimSpace(*ownerID) == "" {
		ownerID = &creatorID
	}

	risk := &model.Risk{
		WorkspaceID:          req.WorkspaceID,
		ProjectID:            req.ProjectID,
		Title:                title,
		Description:          strings.TrimSpace(req.Description),
		Category:             category,
		Probability:          probability,
		Impact:               impact,
		Status:               model.RiskStatusIdentified,
		Strategy:             strategy,
		MitigationPlan:       strings.TrimSpace(req.MitigationPlan),
		ContingencyPlan:      strings.TrimSpace(req.ContingencyPlan),
		OwnerID:              ownerID,
		TargetResolutionDate: req.TargetResolutionDate,
	}

	id, err := s.repo.CreateRisk(risk)
	if err != nil {
		return nil, err
	}

	return s.repo.GetRiskByID(id)
}

func (s *riskServiceImpl) GetRisk(id int64) (*model.Risk, error) {
	if id <= 0 {
		return nil, ErrRiskNotFound
	}
	return s.repo.GetRiskByID(id)
}

func (s *riskServiceImpl) GetRiskByCode(workspaceID int64, code string) (*model.Risk, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrRiskNotFound
	}
	return s.repo.GetRiskByCode(workspaceID, code)
}

func (s *riskServiceImpl) ListRisks(filter model.RiskFilter) ([]*model.Risk, error) {
	if filter.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.ListRisks(filter)
}

func (s *riskServiceImpl) UpdateRisk(id int64, req model.UpdateRiskRequest) (*model.Risk, error) {
	if id <= 0 {
		return nil, ErrRiskNotFound
	}

	existing, err := s.repo.GetRiskByID(id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		title := strings.TrimSpace(*req.Title)
		if title == "" {
			return nil, ErrRiskTitleRequired
		}
		existing.Title = title
	}

	if req.Description != nil {
		existing.Description = strings.TrimSpace(*req.Description)
	}

	if req.Category != nil {
		cat := strings.TrimSpace(*req.Category)
		if !isValidRiskCategory(cat) {
			return nil, ErrInvalidRiskCategory
		}
		existing.Category = cat
	}

	if req.Strategy != nil {
		strat := strings.TrimSpace(*req.Strategy)
		if !isValidRiskStrategy(strat) {
			return nil, ErrInvalidRiskStrategy
		}
		existing.Strategy = strat
	}

	if req.MitigationPlan != nil {
		existing.MitigationPlan = strings.TrimSpace(*req.MitigationPlan)
	}

	if req.ContingencyPlan != nil {
		existing.ContingencyPlan = strings.TrimSpace(*req.ContingencyPlan)
	}

	if req.OwnerID != nil {
		existing.OwnerID = req.OwnerID
	}

	if req.TargetResolutionDate != nil {
		existing.TargetResolutionDate = req.TargetResolutionDate
	}

	if err := s.repo.UpdateRisk(existing); err != nil {
		return nil, err
	}

	return s.repo.GetRiskByID(id)
}

func (s *riskServiceImpl) DeleteRisk(id int64) error {
	if id <= 0 {
		return ErrRiskNotFound
	}
	return s.repo.DeleteRisk(id)
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
