package service

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var (
	ErrPortfolioNotFound      = errors.New("portfolio not found")
	ErrPortfolioNameRequired  = errors.New("portfolio name is required")
	ErrInvalidPortfolioStatus = errors.New("invalid portfolio status")
	ErrPortfolioProjectExists = errors.New("project is already in this portfolio")
	ErrPortfolioProjectNF     = errors.New("project not found in portfolio")
	ErrMilestoneNotFound      = errors.New("milestone not found")
	ErrMilestoneTitleRequired = errors.New("milestone title is required")
	ErrMilestoneDateRequired  = errors.New("milestone target date is required")
	ErrUpdateSummaryRequired  = errors.New("executive update summary is required")
	ErrInvalidHealthStatus    = errors.New("invalid health status")
)

// PortfolioService orchestrates portfolios, cross-project tracking, milestones, and executive reporting
type PortfolioService interface {
	CreatePortfolio(userID string, req model.CreatePortfolioRequest) (*model.Portfolio, error)
	GetPortfolio(id int64) (*model.PortfolioDetail, error)
	ListPortfolios(filter model.PortfolioFilter) ([]*model.PortfolioSummary, error)
	UpdatePortfolio(id int64, req model.UpdatePortfolioRequest) (*model.Portfolio, error)
	DeletePortfolio(id int64) error

	AddProject(portfolioID int64, req model.AddPortfolioProjectRequest) (*model.PortfolioProject, error)
	UpdateProject(portfolioID int64, projectID int, req model.UpdatePortfolioProjectRequest) (*model.PortfolioProject, error)
	RemoveProject(portfolioID int64, projectID int) error
	ListProjects(portfolioID int64) ([]model.PortfolioProject, error)

	CreateMilestone(portfolioID int64, req model.CreatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error)
	UpdateMilestone(milestoneID int64, req model.UpdatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error)
	DeleteMilestone(milestoneID int64) error
	ListMilestones(portfolioID int64) ([]model.PortfolioMilestone, error)

	CreateUpdate(authorUserID string, portfolioID int64, req model.CreatePortfolioUpdateRequest) (*model.PortfolioUpdate, error)
	ListUpdates(portfolioID int64) ([]model.PortfolioUpdate, error)

	GetPortfolioMetrics(portfolioID int64) (*model.PortfolioMetrics, error)
	GetWorkspaceMetrics(workspaceID int64) (*model.WorkspacePortfolioMetrics, error)
}

type portfolioServiceImpl struct {
	repo        repository.PortfolioRepository
	projectRepo repository.ProjectRepository
}

// NewPortfolioService constructs a new PortfolioService instance
func NewPortfolioService(repo repository.PortfolioRepository, projectRepo repository.ProjectRepository) PortfolioService {
	return &portfolioServiceImpl{
		repo:        repo,
		projectRepo: projectRepo,
	}
}

func (s *portfolioServiceImpl) CreatePortfolio(userID string, req model.CreatePortfolioRequest) (*model.Portfolio, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, ErrPortfolioNameRequired
	}
	if req.WorkspaceID <= 0 {
		return nil, ErrWorkspaceRequired
	}

	color := strings.TrimSpace(req.Color)
	if color == "" {
		color = "#4f46e5"
	}

	curr := strings.TrimSpace(req.BudgetCurrency)
	if curr == "" {
		curr = "USD"
	}

	var budget float64
	if req.BudgetAmount != nil && *req.BudgetAmount >= 0 {
		budget = *req.BudgetAmount
	}

	portfolio := &model.Portfolio{
		WorkspaceID:    req.WorkspaceID,
		Name:           name,
		Description:    req.Description,
		Color:          color,
		OwnerUserID:    userID,
		Status:         model.PortfolioStatusOnTrack,
		StartDate:      req.StartDate,
		TargetDate:     req.TargetDate,
		BudgetCurrency: curr,
		BudgetAmount:   budget,
		ActualCost:     0.0,
	}

	id, err := s.repo.CreatePortfolio(portfolio)
	if err != nil {
		return nil, err
	}
	portfolio.ID = id
	return portfolio, nil
}

func (s *portfolioServiceImpl) GetPortfolio(id int64) (*model.PortfolioDetail, error) {
	portfolio, err := s.repo.GetPortfolioByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	projects, err := s.repo.ListPortfolioProjects(id)
	if err != nil {
		return nil, err
	}

	milestones, err := s.repo.ListMilestones(id)
	if err != nil {
		return nil, err
	}

	updates, err := s.repo.ListUpdates(id)
	if err != nil {
		return nil, err
	}

	var latestUpdate *model.PortfolioUpdate
	if len(updates) > 0 {
		latestUpdate = &updates[0]
	}

	var totalTasks, completedTasks int
	for _, pr := range projects {
		totalTasks += pr.TaskCount
		completedTasks += pr.CompletedTaskCount
	}

	var overallProgress float64
	if totalTasks > 0 {
		overallProgress = math.Round((float64(completedTasks)/float64(totalTasks))*1000) / 10
	}

	return &model.PortfolioDetail{
		Portfolio:       portfolio,
		Projects:        projects,
		Milestones:      milestones,
		LatestUpdate:    latestUpdate,
		OverallProgress: overallProgress,
	}, nil
}

func (s *portfolioServiceImpl) ListPortfolios(filter model.PortfolioFilter) ([]*model.PortfolioSummary, error) {
	return s.repo.ListPortfolios(filter)
}

func (s *portfolioServiceImpl) UpdatePortfolio(id int64, req model.UpdatePortfolioRequest) (*model.Portfolio, error) {
	portfolio, err := s.repo.GetPortfolioByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if trimmed == "" {
			return nil, ErrPortfolioNameRequired
		}
		portfolio.Name = trimmed
	}
	if req.Description != nil {
		portfolio.Description = *req.Description
	}
	if req.Color != nil {
		portfolio.Color = strings.TrimSpace(*req.Color)
	}
	if req.OwnerUserID != nil && strings.TrimSpace(*req.OwnerUserID) != "" {
		portfolio.OwnerUserID = strings.TrimSpace(*req.OwnerUserID)
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st != model.PortfolioStatusOnTrack &&
			st != model.PortfolioStatusAtRisk &&
			st != model.PortfolioStatusOffTrack &&
			st != model.PortfolioStatusOnHold &&
			st != model.PortfolioStatusCompleted {
			return nil, ErrInvalidPortfolioStatus
		}
		portfolio.Status = st
	}
	if req.StartDate != nil {
		portfolio.StartDate = req.StartDate
	}
	if req.TargetDate != nil {
		portfolio.TargetDate = req.TargetDate
	}
	if req.BudgetCurrency != nil && strings.TrimSpace(*req.BudgetCurrency) != "" {
		portfolio.BudgetCurrency = strings.TrimSpace(*req.BudgetCurrency)
	}
	if req.BudgetAmount != nil && *req.BudgetAmount >= 0 {
		portfolio.BudgetAmount = *req.BudgetAmount
	}
	if req.ActualCost != nil && *req.ActualCost >= 0 {
		portfolio.ActualCost = *req.ActualCost
	}

	if err := s.repo.UpdatePortfolio(portfolio); err != nil {
		return nil, err
	}
	return portfolio, nil
}

func (s *portfolioServiceImpl) DeletePortfolio(id int64) error {
	err := s.repo.DeletePortfolio(id)
	if err != nil && errors.Is(err, repository.ErrPortfolioNotFound) {
		return ErrPortfolioNotFound
	}
	return err
}

func (s *portfolioServiceImpl) AddProject(portfolioID int64, req model.AddPortfolioProjectRequest) (*model.PortfolioProject, error) {
	if _, err := s.repo.GetPortfolioByID(portfolioID); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	if req.ProjectID <= 0 {
		return nil, ErrProjectNotFound
	}

	if s.projectRepo != nil {
		p, err := s.projectRepo.GetByID(req.ProjectID)
		if err != nil || p == nil {
			return nil, ErrProjectNotFound
		}
	}

	priority := req.Priority
	if priority == "" {
		priority = model.PortfolioProjectPriorityMedium
	}
	if priority != model.PortfolioProjectPriorityLow &&
		priority != model.PortfolioProjectPriorityMedium &&
		priority != model.PortfolioProjectPriorityHigh &&
		priority != model.PortfolioProjectPriorityCritical {
		priority = model.PortfolioProjectPriorityMedium
	}

	var allocBudget, spentAmount float64
	if req.AllocatedBudget != nil && *req.AllocatedBudget >= 0 {
		allocBudget = *req.AllocatedBudget
	}
	if req.SpentAmount != nil && *req.SpentAmount >= 0 {
		spentAmount = *req.SpentAmount
	}

	orderIdx := 0
	if req.OrderIndex != nil {
		orderIdx = *req.OrderIndex
	}

	item := &model.PortfolioProject{
		PortfolioID:     portfolioID,
		ProjectID:       req.ProjectID,
		Priority:        priority,
		StatusOverride:  req.StatusOverride,
		AllocatedBudget: allocBudget,
		SpentAmount:     spentAmount,
		OrderIndex:      orderIdx,
	}

	if _, err := s.repo.AddProjectToPortfolio(item); err != nil {
		if errors.Is(err, repository.ErrPortfolioProjectExists) {
			return nil, ErrPortfolioProjectExists
		}
		return nil, err
	}

	return s.repo.GetPortfolioProject(portfolioID, req.ProjectID)
}

func (s *portfolioServiceImpl) UpdateProject(portfolioID int64, projectID int, req model.UpdatePortfolioProjectRequest) (*model.PortfolioProject, error) {
	item, err := s.repo.GetPortfolioProject(portfolioID, projectID)
	if err != nil {
		if errors.Is(err, repository.ErrPortfolioProjectNF) {
			return nil, ErrPortfolioProjectNF
		}
		return nil, err
	}

	if req.Priority != nil {
		pr := *req.Priority
		if pr == model.PortfolioProjectPriorityLow ||
			pr == model.PortfolioProjectPriorityMedium ||
			pr == model.PortfolioProjectPriorityHigh ||
			pr == model.PortfolioProjectPriorityCritical {
			item.Priority = pr
		}
	}
	if req.StatusOverride != nil {
		item.StatusOverride = *req.StatusOverride
	}
	if req.AllocatedBudget != nil && *req.AllocatedBudget >= 0 {
		item.AllocatedBudget = *req.AllocatedBudget
	}
	if req.SpentAmount != nil && *req.SpentAmount >= 0 {
		item.SpentAmount = *req.SpentAmount
	}
	if req.OrderIndex != nil {
		item.OrderIndex = *req.OrderIndex
	}

	if err := s.repo.UpdatePortfolioProject(item); err != nil {
		return nil, err
	}

	return s.repo.GetPortfolioProject(portfolioID, projectID)
}

func (s *portfolioServiceImpl) RemoveProject(portfolioID int64, projectID int) error {
	err := s.repo.RemoveProjectFromPortfolio(portfolioID, projectID)
	if err != nil && errors.Is(err, repository.ErrPortfolioProjectNF) {
		return ErrPortfolioProjectNF
	}
	return err
}

func (s *portfolioServiceImpl) ListProjects(portfolioID int64) ([]model.PortfolioProject, error) {
	if _, err := s.repo.GetPortfolioByID(portfolioID); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}
	return s.repo.ListPortfolioProjects(portfolioID)
}

func (s *portfolioServiceImpl) CreateMilestone(portfolioID int64, req model.CreatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error) {
	if _, err := s.repo.GetPortfolioByID(portfolioID); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, ErrMilestoneTitleRequired
	}
	if req.TargetDate.IsZero() {
		return nil, ErrMilestoneDateRequired
	}

	milestone := &model.PortfolioMilestone{
		PortfolioID: portfolioID,
		Title:       title,
		Description: req.Description,
		TargetDate:  req.TargetDate,
		Status:      model.MilestoneStatusPending,
	}

	id, err := s.repo.CreateMilestone(milestone)
	if err != nil {
		return nil, err
	}
	milestone.ID = id
	return milestone, nil
}

func (s *portfolioServiceImpl) UpdateMilestone(milestoneID int64, req model.UpdatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error) {
	milestone, err := s.repo.GetMilestoneByID(milestoneID)
	if err != nil {
		if errors.Is(err, repository.ErrMilestoneNotFound) {
			return nil, ErrMilestoneNotFound
		}
		return nil, err
	}

	if req.Title != nil {
		trimmed := strings.TrimSpace(*req.Title)
		if trimmed == "" {
			return nil, ErrMilestoneTitleRequired
		}
		milestone.Title = trimmed
	}
	if req.Description != nil {
		milestone.Description = *req.Description
	}
	if req.TargetDate != nil {
		milestone.TargetDate = *req.TargetDate
	}
	if req.Status != nil {
		st := strings.TrimSpace(*req.Status)
		if st == model.MilestoneStatusPending ||
			st == model.MilestoneStatusInProgress ||
			st == model.MilestoneStatusCompleted ||
			st == model.MilestoneStatusMissed {
			milestone.Status = st
			if st == model.MilestoneStatusCompleted {
				now := time.Now()
				milestone.CompletedAt = &now
			} else {
				milestone.CompletedAt = nil
			}
		}
	}

	if err := s.repo.UpdateMilestone(milestone); err != nil {
		return nil, err
	}
	return milestone, nil
}

func (s *portfolioServiceImpl) DeleteMilestone(milestoneID int64) error {
	err := s.repo.DeleteMilestone(milestoneID)
	if err != nil && errors.Is(err, repository.ErrMilestoneNotFound) {
		return ErrMilestoneNotFound
	}
	return err
}

func (s *portfolioServiceImpl) ListMilestones(portfolioID int64) ([]model.PortfolioMilestone, error) {
	if _, err := s.repo.GetPortfolioByID(portfolioID); err != nil {
		if errors.Is(err, repository.ErrPortfolioNotFound) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}
	return s.repo.ListMilestones(portfolioID)
}

func (s *portfolioServiceImpl) CreateUpdate(authorUserID string, portfolioID int64, req model.CreatePortfolioUpdateRequest) (*model.PortfolioUpdate, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) ListUpdates(portfolioID int64) ([]model.PortfolioUpdate, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) GetPortfolioMetrics(portfolioID int64) (*model.PortfolioMetrics, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) GetWorkspaceMetrics(workspaceID int64) (*model.WorkspacePortfolioMetrics, error) {
	return nil, nil
}
