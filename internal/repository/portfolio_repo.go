package repository

import (
	"database/sql"
	"errors"
	"math"
	"time"

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

func scanPortfolio(scanner interface{ Scan(dest ...interface{}) error }) (*model.Portfolio, error) {
	var (
		p          model.Portfolio
		desc       sql.NullString
		startDate  sql.NullTime
		targetDate sql.NullTime
	)

	err := scanner.Scan(
		&p.ID,
		&p.WorkspaceID,
		&p.Name,
		&desc,
		&p.Color,
		&p.OwnerUserID,
		&p.Status,
		&startDate,
		&targetDate,
		&p.BudgetCurrency,
		&p.BudgetAmount,
		&p.ActualCost,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrPortfolioNotFound
		}
		return nil, err
	}

	if desc.Valid {
		p.Description = desc.String
	}
	if startDate.Valid {
		p.StartDate = &startDate.Time
	}
	if targetDate.Valid {
		p.TargetDate = &targetDate.Time
	}

	return &p, nil
}

func (r *portfolioRepositoryImpl) CreatePortfolio(portfolio *model.Portfolio) (int64, error) {
	query := `
	INSERT INTO portfolios (
		workspace_id, name, description, color,
		owner_user_id, status, start_date, target_date,
		budget_currency, budget_amount, actual_cost,
		created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if portfolio.CreatedAt.IsZero() {
		portfolio.CreatedAt = now
	}
	if portfolio.UpdatedAt.IsZero() {
		portfolio.UpdatedAt = now
	}
	if portfolio.Color == "" {
		portfolio.Color = "#4f46e5"
	}
	if portfolio.Status == "" {
		portfolio.Status = model.PortfolioStatusOnTrack
	}
	if portfolio.BudgetCurrency == "" {
		portfolio.BudgetCurrency = "USD"
	}

	res, err := r.db.Exec(query,
		portfolio.WorkspaceID,
		portfolio.Name,
		portfolio.Description,
		portfolio.Color,
		portfolio.OwnerUserID,
		portfolio.Status,
		portfolio.StartDate,
		portfolio.TargetDate,
		portfolio.BudgetCurrency,
		portfolio.BudgetAmount,
		portfolio.ActualCost,
		portfolio.CreatedAt,
		portfolio.UpdatedAt,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	portfolio.ID = id
	return id, nil
}

func (r *portfolioRepositoryImpl) GetPortfolioByID(id int64) (*model.Portfolio, error) {
	query := `
	SELECT id, workspace_id, name, description, color,
		owner_user_id, status, start_date, target_date,
		budget_currency, budget_amount, actual_cost,
		created_at, updated_at
	FROM portfolios
	WHERE id = ?`

	return scanPortfolio(r.db.QueryRow(query, id))
}

func (r *portfolioRepositoryImpl) ListPortfolios(filter model.PortfolioFilter) ([]*model.PortfolioSummary, error) {
	baseQuery := `
	SELECT 
		p.id, p.workspace_id, p.name, p.color, p.owner_user_id, p.status, p.target_date,
		p.budget_currency, p.budget_amount, p.actual_cost, p.created_at, p.updated_at,
		COUNT(DISTINCT pp.id) AS project_count,
		COUNT(DISTINCT pm.id) AS milestone_count,
		COUNT(DISTINCT CASE WHEN pm.status = 'completed' THEN pm.id END) AS completed_milestones
	FROM portfolios p
	LEFT JOIN portfolio_projects pp ON p.id = pp.portfolio_id
	LEFT JOIN portfolio_milestones pm ON p.id = pm.portfolio_id
	WHERE 1=1`

	var args []interface{}

	if filter.WorkspaceID != nil && *filter.WorkspaceID > 0 {
		baseQuery += ` AND p.workspace_id = ?`
		args = append(args, *filter.WorkspaceID)
	}
	if filter.OwnerUserID != "" {
		baseQuery += ` AND p.owner_user_id = ?`
		args = append(args, filter.OwnerUserID)
	}
	if filter.Status != "" {
		baseQuery += ` AND p.status = ?`
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		baseQuery += ` AND (p.name LIKE ? OR p.description LIKE ?)`
		searchTerm := "%" + filter.Search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	baseQuery += ` GROUP BY p.id ORDER BY p.updated_at DESC`

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	baseQuery += ` LIMIT ?`
	args = append(args, limit)

	if filter.Offset > 0 {
		baseQuery += ` OFFSET ?`
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []*model.PortfolioSummary
	for rows.Next() {
		var (
			s          model.PortfolioSummary
			targetDate sql.NullTime
		)
		err := rows.Scan(
			&s.ID,
			&s.WorkspaceID,
			&s.Name,
			&s.Color,
			&s.OwnerUserID,
			&s.Status,
			&targetDate,
			&s.BudgetCurrency,
			&s.BudgetAmount,
			&s.ActualCost,
			&s.CreatedAt,
			&s.UpdatedAt,
			&s.ProjectCount,
			&s.MilestoneCount,
			&s.CompletedMilestones,
		)
		if err != nil {
			return nil, err
		}
		if targetDate.Valid {
			s.TargetDate = &targetDate.Time
		}

		var totalTasks, completedTasks int
		taskQuery := `
		SELECT 
			COUNT(t.id),
			COALESCE(SUM(CASE WHEN t.status = 'completed' OR t.completed = 1 THEN 1 ELSE 0 END), 0)
		FROM tasks t
		JOIN portfolio_projects pp ON t.project_id = pp.project_id
		WHERE pp.portfolio_id = ?`
		_ = r.db.QueryRow(taskQuery, s.ID).Scan(&totalTasks, &completedTasks)
		s.TotalTasks = totalTasks
		s.CompletedTasks = completedTasks
		if totalTasks > 0 {
			s.ProgressPercentage = math.Round((float64(completedTasks)/float64(totalTasks))*1000) / 10
		}

		summaries = append(summaries, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if summaries == nil {
		summaries = []*model.PortfolioSummary{}
	}

	return summaries, nil
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

