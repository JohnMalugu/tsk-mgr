package repository

import (
	"database/sql"
	"errors"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	ErrPortfolioNotFound      = errors.New("portfolio not found")
	ErrPortfolioProjectExists = errors.New("project is already in this portfolio")
	ErrPortfolioProjectNF     = errors.New("project not found in portfolio")
	ErrMilestoneNotFound      = errors.New("milestone not found")
)

// PortfolioRepository defines persistent storage operations for portfolios and strategic initiatives
type PortfolioRepository interface {
	CreatePortfolio(portfolio *model.Portfolio) (int64, error)
	GetPortfolioByID(id int64) (*model.Portfolio, error)
	ListPortfolios(filter model.PortfolioFilter) ([]*model.PortfolioSummary, error)
	UpdatePortfolio(portfolio *model.Portfolio) error
	DeletePortfolio(id int64) error

	AddProjectToPortfolio(item *model.PortfolioProject) (int64, error)
	GetPortfolioProject(portfolioID int64, projectID int) (*model.PortfolioProject, error)
	ListPortfolioProjects(portfolioID int64) ([]model.PortfolioProject, error)
	UpdatePortfolioProject(item *model.PortfolioProject) error
	RemoveProjectFromPortfolio(portfolioID int64, projectID int) error

	CreateMilestone(milestone *model.PortfolioMilestone) (int64, error)
	GetMilestoneByID(id int64) (*model.PortfolioMilestone, error)
	ListMilestones(portfolioID int64) ([]model.PortfolioMilestone, error)
	UpdateMilestone(milestone *model.PortfolioMilestone) error
	DeleteMilestone(id int64) error

	CreateUpdate(update *model.PortfolioUpdate) (int64, error)
	ListUpdates(portfolioID int64) ([]model.PortfolioUpdate, error)

	GetPortfolioMetrics(portfolioID int64) (*model.PortfolioMetrics, error)
	GetWorkspacePortfolioMetrics(workspaceID int64) (*model.WorkspacePortfolioMetrics, error)
}

type portfolioRepositoryImpl struct {
	db *sql.DB
}

// NewPortfolioRepository constructs a SQLite-backed PortfolioRepository instance
func NewPortfolioRepository(db *sql.DB) PortfolioRepository {
	return &portfolioRepositoryImpl{db: db}
}

func (r *portfolioRepositoryImpl) CreatePortfolio(portfolio *model.Portfolio) (int64, error) {
	return 0, nil
}

func (r *portfolioRepositoryImpl) GetPortfolioByID(id int64) (*model.Portfolio, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) ListPortfolios(filter model.PortfolioFilter) ([]*model.PortfolioSummary, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) UpdatePortfolio(portfolio *model.Portfolio) error {
	return nil
}

func (r *portfolioRepositoryImpl) DeletePortfolio(id int64) error {
	return nil
}

func (r *portfolioRepositoryImpl) AddProjectToPortfolio(item *model.PortfolioProject) (int64, error) {
	return 0, nil
}

func (r *portfolioRepositoryImpl) GetPortfolioProject(portfolioID int64, projectID int) (*model.PortfolioProject, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) ListPortfolioProjects(portfolioID int64) ([]model.PortfolioProject, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) UpdatePortfolioProject(item *model.PortfolioProject) error {
	return nil
}

func (r *portfolioRepositoryImpl) RemoveProjectFromPortfolio(portfolioID int64, projectID int) error {
	return nil
}

func (r *portfolioRepositoryImpl) CreateMilestone(milestone *model.PortfolioMilestone) (int64, error) {
	return 0, nil
}

func (r *portfolioRepositoryImpl) GetMilestoneByID(id int64) (*model.PortfolioMilestone, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) ListMilestones(portfolioID int64) ([]model.PortfolioMilestone, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) UpdateMilestone(milestone *model.PortfolioMilestone) error {
	return nil
}

func (r *portfolioRepositoryImpl) DeleteMilestone(id int64) error {
	return nil
}

func (r *portfolioRepositoryImpl) CreateUpdate(update *model.PortfolioUpdate) (int64, error) {
	return 0, nil
}

func (r *portfolioRepositoryImpl) ListUpdates(portfolioID int64) ([]model.PortfolioUpdate, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) GetPortfolioMetrics(portfolioID int64) (*model.PortfolioMetrics, error) {
	return nil, nil
}

func (r *portfolioRepositoryImpl) GetWorkspacePortfolioMetrics(workspaceID int64) (*model.WorkspacePortfolioMetrics, error) {
	return nil, nil
}

