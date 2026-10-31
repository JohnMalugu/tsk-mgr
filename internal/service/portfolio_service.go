package service

import (
	"errors"
	"math"
	"strings"

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
	return nil, nil
}

func (s *portfolioServiceImpl) DeletePortfolio(id int64) error {
	return nil
}

func (s *portfolioServiceImpl) AddProject(portfolioID int64, req model.AddPortfolioProjectRequest) (*model.PortfolioProject, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) UpdateProject(portfolioID int64, projectID int, req model.UpdatePortfolioProjectRequest) (*model.PortfolioProject, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) RemoveProject(portfolioID int64, projectID int) error {
	return nil
}

func (s *portfolioServiceImpl) ListProjects(portfolioID int64) ([]model.PortfolioProject, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) CreateMilestone(portfolioID int64, req model.CreatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) UpdateMilestone(milestoneID int64, req model.UpdatePortfolioMilestoneRequest) (*model.PortfolioMilestone, error) {
	return nil, nil
}

func (s *portfolioServiceImpl) DeleteMilestone(milestoneID int64) error {
	return nil
}

func (s *portfolioServiceImpl) ListMilestones(portfolioID int64) ([]model.PortfolioMilestone, error) {
	return nil, nil
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
