package service

import (
	"errors"
	"strings"
	"time"

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
		proj, err := s.projectRepo.GetByID(*req.ProjectID)
		if err != nil || proj == nil {
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
	risk, err := s.repo.GetRiskByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrRiskNotFound) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}
	return risk, nil
}

func (s *riskServiceImpl) GetRiskByCode(workspaceID int64, code string) (*model.Risk, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, ErrRiskNotFound
	}
	risk, err := s.repo.GetRiskByCode(workspaceID, code)
	if err != nil {
		if errors.Is(err, repository.ErrRiskNotFound) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}
	return risk, nil
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
		if errors.Is(err, repository.ErrRiskNotFound) {
			return nil, ErrRiskNotFound
		}
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
		if errors.Is(err, repository.ErrRiskNotFound) {
			return nil, ErrRiskNotFound
		}
		return nil, err
	}

	return s.repo.GetRiskByID(id)
}

func (s *riskServiceImpl) DeleteRisk(id int64) error {
	if id <= 0 {
		return ErrRiskNotFound
	}
	err := s.repo.DeleteRisk(id)
	if err != nil {
		if errors.Is(err, repository.ErrRiskNotFound) {
			return ErrRiskNotFound
		}
		return err
	}
	return nil
}

func (s *riskServiceImpl) AssessRisk(id int64, reviewerID string, req model.AssessRiskRequest) (*model.Risk, error) {
	if id <= 0 {
		return nil, ErrRiskNotFound
	}

	if req.Probability < 1 || req.Probability > 5 {
		return nil, ErrInvalidProbability
	}
	if req.Impact < 1 || req.Impact > 5 {
		return nil, ErrInvalidImpact
	}

	existing, err := s.repo.GetRiskByID(id)
	if err != nil {
		return nil, err
	}

	prevProbability := existing.Probability
	prevImpact := existing.Impact
	prevScore := existing.Score

	existing.Probability = req.Probability
	existing.Impact = req.Impact
	existing.Score, existing.Severity = model.CalculateRiskSeverity(existing.Probability, existing.Impact)

	if err := s.repo.UpdateRisk(existing); err != nil {
		return nil, err
	}

	// Record audit review log
	review := &model.RiskReview{
		RiskID:              id,
		ReviewerID:          reviewerID,
		PreviousProbability: prevProbability,
		PreviousImpact:      prevImpact,
		PreviousScore:       prevScore,
		NewProbability:      existing.Probability,
		NewImpact:           existing.Impact,
		NewScore:            existing.Score,
		PreviousStatus:      existing.Status,
		NewStatus:           existing.Status,
		Comments:            strings.TrimSpace(req.Comments),
	}
	_, _ = s.repo.CreateReview(review)

	return s.repo.GetRiskByID(id)
}

func (s *riskServiceImpl) ChangeStatus(id int64, reviewerID string, req model.ChangeRiskStatusRequest) (*model.Risk, error) {
	if id <= 0 {
		return nil, ErrRiskNotFound
	}

	status := strings.TrimSpace(req.Status)
	if !isValidRiskStatus(status) {
		return nil, ErrInvalidRiskStatus
	}

	existing, err := s.repo.GetRiskByID(id)
	if err != nil {
		return nil, err
	}

	prevStatus := existing.Status
	existing.Status = status

	if status == model.RiskStatusClosed {
		if existing.ClosedAt == nil {
			now := time.Now().UTC()
			existing.ClosedAt = &now
		}
	} else {
		existing.ClosedAt = nil
	}

	if err := s.repo.UpdateRisk(existing); err != nil {
		return nil, err
	}

	// Record review if comments provided or status changed
	comments := strings.TrimSpace(req.Comments)
	if comments != "" || prevStatus != status {
		review := &model.RiskReview{
			RiskID:              id,
			ReviewerID:          reviewerID,
			PreviousProbability: existing.Probability,
			PreviousImpact:      existing.Impact,
			PreviousScore:       existing.Score,
			NewProbability:      existing.Probability,
			NewImpact:           existing.Impact,
			NewScore:            existing.Score,
			PreviousStatus:      prevStatus,
			NewStatus:           status,
			Comments:            comments,
		}
		_, _ = s.repo.CreateReview(review)
	}

	return s.repo.GetRiskByID(id)
}

func isValidRiskAction(action string) bool {
	switch action {
	case model.RiskActionMitigation,
		model.RiskActionContingency,
		model.RiskActionInvestigation:
		return true
	default:
		return false
	}
}

func (s *riskServiceImpl) LinkTask(riskID int64, req model.LinkRiskTaskRequest) (*model.RiskTaskLink, error) {
	if riskID <= 0 {
		return nil, ErrRiskNotFound
	}
	if req.TaskID <= 0 {
		return nil, ErrTaskNotFound
	}

	if _, err := s.repo.GetRiskByID(riskID); err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(req.TaskID)
	if err != nil || task == nil {
		return nil, ErrTaskNotFound
	}

	actionType := strings.TrimSpace(req.ActionType)
	if actionType == "" {
		actionType = model.RiskActionMitigation
	} else if !isValidRiskAction(actionType) {
		return nil, ErrInvalidRiskAction
	}

	link := &model.RiskTaskLink{
		RiskID:     riskID,
		TaskID:     req.TaskID,
		ActionType: actionType,
		Notes:      strings.TrimSpace(req.Notes),
	}

	id, err := s.repo.LinkTask(link)
	if err != nil {
		return nil, err
	}
	link.ID = id
	link.TaskTitle = task.Title
	link.TaskStatus = task.Status
	link.TaskPriority = task.Priority
	link.TaskCompleted = task.Completed
	return link, nil
}

func (s *riskServiceImpl) UnlinkTask(riskID int64, taskID int) error {
	if riskID <= 0 {
		return ErrRiskNotFound
	}
	if taskID <= 0 {
		return ErrTaskNotFound
	}
	return s.repo.UnlinkTask(riskID, taskID)
}

func (s *riskServiceImpl) ListRiskTasks(riskID int64) ([]model.RiskTaskLink, error) {
	if riskID <= 0 {
		return nil, ErrRiskNotFound
	}
	if _, err := s.repo.GetRiskByID(riskID); err != nil {
		return nil, err
	}
	return s.repo.ListRiskTasks(riskID)
}

func (s *riskServiceImpl) CreateReview(riskID int64, reviewerID string, req model.CreateRiskReviewRequest) (*model.RiskReview, error) {
	if riskID <= 0 {
		return nil, ErrRiskNotFound
	}
	comments := strings.TrimSpace(req.Comments)
	if comments == "" {
		return nil, ErrReviewCommentReq
	}

	existing, err := s.repo.GetRiskByID(riskID)
	if err != nil {
		return nil, err
	}

	prevProbability := existing.Probability
	prevImpact := existing.Impact
	prevScore := existing.Score
	prevStatus := existing.Status

	modified := false
	if req.NewProbability != nil {
		if *req.NewProbability < 1 || *req.NewProbability > 5 {
			return nil, ErrInvalidProbability
		}
		existing.Probability = *req.NewProbability
		modified = true
	}
	if req.NewImpact != nil {
		if *req.NewImpact < 1 || *req.NewImpact > 5 {
			return nil, ErrInvalidImpact
		}
		existing.Impact = *req.NewImpact
		modified = true
	}
	if modified {
		existing.Score, existing.Severity = model.CalculateRiskSeverity(existing.Probability, existing.Impact)
	}

	targetStatus := strings.TrimSpace(req.NewStatus)
	if targetStatus != "" {
		if !isValidRiskStatus(targetStatus) {
			return nil, ErrInvalidRiskStatus
		}
		existing.Status = targetStatus
		if targetStatus == model.RiskStatusClosed {
			if existing.ClosedAt == nil {
				now := time.Now().UTC()
				existing.ClosedAt = &now
			}
		} else {
			existing.ClosedAt = nil
		}
		modified = true
	}

	if modified {
		if err := s.repo.UpdateRisk(existing); err != nil {
			return nil, err
		}
	}

	review := &model.RiskReview{
		RiskID:              riskID,
		ReviewerID:          reviewerID,
		PreviousProbability: prevProbability,
		PreviousImpact:      prevImpact,
		PreviousScore:       prevScore,
		NewProbability:      existing.Probability,
		NewImpact:           existing.Impact,
		NewScore:            existing.Score,
		PreviousStatus:      prevStatus,
		NewStatus:           existing.Status,
		Comments:            comments,
	}

	id, err := s.repo.CreateReview(review)
	if err != nil {
		return nil, err
	}
	review.ID = id
	return review, nil
}

func (s *riskServiceImpl) ListReviews(riskID int64) ([]model.RiskReview, error) {
	if riskID <= 0 {
		return nil, ErrRiskNotFound
	}
	if _, err := s.repo.GetRiskByID(riskID); err != nil {
		return nil, err
	}
	return s.repo.ListReviews(riskID)
}

func (s *riskServiceImpl) GetRiskMetrics(workspaceID int64) (*model.RiskMetrics, error) {
	if workspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}
	return s.repo.GetRiskMetrics(workspaceID)
}
