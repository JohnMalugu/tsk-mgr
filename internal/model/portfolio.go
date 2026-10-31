package model

import "time"

// Portfolio health and lifecycle status constants
const (
	PortfolioStatusOnTrack   = "on_track"
	PortfolioStatusAtRisk    = "at_risk"
	PortfolioStatusOffTrack  = "off_track"
	PortfolioStatusOnHold    = "on_hold"
	PortfolioStatusCompleted = "completed"
)

// Portfolio represents a collection of projects and strategic initiatives
type Portfolio struct {
	ID             int64      `json:"id"`
	WorkspaceID    int64      `json:"workspace_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	Color          string     `json:"color"`
	OwnerUserID    string     `json:"owner_user_id"`
	Status         string     `json:"status"` // on_track, at_risk, off_track, on_hold, completed
	StartDate      *time.Time `json:"start_date,omitempty"`
	TargetDate     *time.Time `json:"target_date,omitempty"`
	BudgetCurrency string     `json:"budget_currency"` // e.g. "USD", "EUR"
	BudgetAmount   float64    `json:"budget_amount"`
	ActualCost     float64    `json:"actual_cost"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// Portfolio project priority constants
const (
	PortfolioProjectPriorityLow      = "low"
	PortfolioProjectPriorityMedium   = "medium"
	PortfolioProjectPriorityHigh     = "high"
	PortfolioProjectPriorityCritical = "critical"
)

// PortfolioProject represents a project mapped into a portfolio
type PortfolioProject struct {
	ID              int64     `json:"id"`
	PortfolioID     int64     `json:"portfolio_id"`
	ProjectID       int       `json:"project_id"`
	Priority        string    `json:"priority"`        // low, medium, high, critical
	StatusOverride  string    `json:"status_override"` // on_track, at_risk, off_track, completed
	AllocatedBudget float64   `json:"allocated_budget"`
	SpentAmount     float64   `json:"spent_amount"`
	OrderIndex      int       `json:"order_index"`
	CreatedAt       time.Time `json:"created_at"`

	// Enriched fields from projects & tasks
	ProjectName        string  `json:"project_name,omitempty"`
	ProjectColor       string  `json:"project_color,omitempty"`
	TaskCount          int     `json:"task_count,omitempty"`
	CompletedTaskCount int     `json:"completed_task_count,omitempty"`
	ProgressPercentage float64 `json:"progress_percentage,omitempty"`
}

// Portfolio milestone status constants
const (
	MilestoneStatusPending    = "pending"
	MilestoneStatusInProgress = "in_progress"
	MilestoneStatusCompleted  = "completed"
	MilestoneStatusMissed     = "missed"
)

// PortfolioMilestone represents a key milestone within a portfolio
type PortfolioMilestone struct {
	ID          int64      `json:"id"`
	PortfolioID int64      `json:"portfolio_id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	TargetDate  time.Time  `json:"target_date"`
	Status      string     `json:"status"` // pending, in_progress, completed, missed
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// PortfolioUpdate represents an executive progress report for a portfolio
type PortfolioUpdate struct {
	ID              int64     `json:"id"`
	PortfolioID     int64     `json:"portfolio_id"`
	AuthorUserID    string    `json:"author_user_id"`
	HealthStatus    string    `json:"health_status"` // on_track, at_risk, off_track, on_hold, completed
	Summary         string    `json:"summary"`
	Accomplishments string    `json:"accomplishments,omitempty"`
	Blockers        string    `json:"blockers,omitempty"`
	NextSteps       string    `json:"next_steps,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// CreatePortfolioRequest defines the payload for provisioning a new portfolio
type CreatePortfolioRequest struct {
	WorkspaceID    int64      `json:"workspace_id"`
	Name           string     `json:"name"`
	Description    string     `json:"description,omitempty"`
	Color          string     `json:"color,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	TargetDate     *time.Time `json:"target_date,omitempty"`
	BudgetCurrency string     `json:"budget_currency,omitempty"`
	BudgetAmount   *float64   `json:"budget_amount,omitempty"`
}

// UpdatePortfolioRequest defines the payload for updating portfolio attributes
type UpdatePortfolioRequest struct {
	Name           *string    `json:"name,omitempty"`
	Description    *string    `json:"description,omitempty"`
	Color          *string    `json:"color,omitempty"`
	OwnerUserID    *string    `json:"owner_user_id,omitempty"`
	Status         *string    `json:"status,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	TargetDate     *time.Time `json:"target_date,omitempty"`
	BudgetCurrency *string    `json:"budget_currency,omitempty"`
	BudgetAmount   *float64   `json:"budget_amount,omitempty"`
	ActualCost     *float64   `json:"actual_cost,omitempty"`
}

// AddPortfolioProjectRequest defines the payload for mapping a project into a portfolio
type AddPortfolioProjectRequest struct {
	ProjectID       int      `json:"project_id"`
	Priority        string   `json:"priority,omitempty"` // low, medium, high, critical
	StatusOverride  string   `json:"status_override,omitempty"`
	AllocatedBudget *float64 `json:"allocated_budget,omitempty"`
	SpentAmount     *float64 `json:"spent_amount,omitempty"`
	OrderIndex      *int     `json:"order_index,omitempty"`
}

// UpdatePortfolioProjectRequest defines the payload for updating a mapped project's settings
type UpdatePortfolioProjectRequest struct {
	Priority        *string  `json:"priority,omitempty"`
	StatusOverride  *string  `json:"status_override,omitempty"`
	AllocatedBudget *float64 `json:"allocated_budget,omitempty"`
	SpentAmount     *float64 `json:"spent_amount,omitempty"`
	OrderIndex      *int     `json:"order_index,omitempty"`
}

// CreatePortfolioMilestoneRequest defines the payload for creating a milestone
type CreatePortfolioMilestoneRequest struct {
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	TargetDate  time.Time `json:"target_date"`
}

// UpdatePortfolioMilestoneRequest defines the payload for modifying a milestone
type UpdatePortfolioMilestoneRequest struct {
	Title       *string    `json:"title,omitempty"`
	Description *string    `json:"description,omitempty"`
	TargetDate  *time.Time `json:"target_date,omitempty"`
	Status      *string    `json:"status,omitempty"` // pending, in_progress, completed, missed
}

// CreatePortfolioUpdateRequest defines the payload for logging an executive update
type CreatePortfolioUpdateRequest struct {
	HealthStatus    string `json:"health_status"` // on_track, at_risk, off_track, on_hold, completed
	Summary         string `json:"summary"`
	Accomplishments string `json:"accomplishments,omitempty"`
	Blockers        string `json:"blockers,omitempty"`
	NextSteps       string `json:"next_steps,omitempty"`
}

// PortfolioFilter specifies query filters for listing portfolios
type PortfolioFilter struct {
	WorkspaceID *int64 `json:"workspace_id,omitempty"`
	OwnerUserID string `json:"owner_user_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Search      string `json:"search,omitempty"`
	Limit       int    `json:"limit,omitempty"`
	Offset      int    `json:"offset,omitempty"`
}

// PortfolioDetail provides full portfolio view with linked projects, milestones, and latest update
type PortfolioDetail struct {
	Portfolio       *Portfolio           `json:"portfolio"`
	Projects        []PortfolioProject   `json:"projects"`
	Milestones      []PortfolioMilestone `json:"milestones"`
	LatestUpdate    *PortfolioUpdate     `json:"latest_update,omitempty"`
	OverallProgress float64              `json:"overall_progress"`
}

// PortfolioSummary represents a lightweight portfolio overview item
type PortfolioSummary struct {
	ID                  int64      `json:"id"`
	WorkspaceID         int64      `json:"workspace_id"`
	Name                string     `json:"name"`
	Color               string     `json:"color"`
	OwnerUserID         string     `json:"owner_user_id"`
	Status              string     `json:"status"`
	TargetDate          *time.Time `json:"target_date,omitempty"`
	BudgetCurrency      string     `json:"budget_currency"`
	BudgetAmount        float64    `json:"budget_amount"`
	ActualCost          float64    `json:"actual_cost"`
	ProjectCount        int        `json:"project_count"`
	MilestoneCount      int        `json:"milestone_count"`
	CompletedMilestones int        `json:"completed_milestones"`
	TotalTasks          int        `json:"total_tasks"`
	CompletedTasks      int        `json:"completed_tasks"`
	ProgressPercentage  float64    `json:"progress_percentage"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// PortfolioMetrics summarizes metrics for an individual portfolio
type PortfolioMetrics struct {
	PortfolioID             int64   `json:"portfolio_id"`
	TotalProjects           int     `json:"total_projects"`
	TotalTasks              int     `json:"total_tasks"`
	CompletedTasks          int     `json:"completed_tasks"`
	OverdueTasks            int     `json:"overdue_tasks"`
	ProgressPercentage      float64 `json:"progress_percentage"`
	TotalAllocatedBudget    float64 `json:"total_allocated_budget"`
	TotalSpentAmount        float64 `json:"total_spent_amount"`
	BudgetUtilizationRate   float64 `json:"budget_utilization_rate"` // percentage
	TotalMilestones         int     `json:"total_milestones"`
	CompletedMilestones     int     `json:"completed_milestones"`
	MilestoneCompletionRate float64 `json:"milestone_completion_rate"`
}

// WorkspacePortfolioMetrics aggregates portfolio health and financials across a workspace
type WorkspacePortfolioMetrics struct {
	WorkspaceID        int64          `json:"workspace_id"`
	TotalPortfolios    int            `json:"total_portfolios"`
	OnTrackCount       int            `json:"on_track_count"`
	AtRiskCount        int            `json:"at_risk_count"`
	OffTrackCount      int            `json:"off_track_count"`
	OnHoldCount        int            `json:"on_hold_count"`
	CompletedCount     int            `json:"completed_count"`
	TotalBudgetAmount  float64        `json:"total_budget_amount"`
	TotalActualCost    float64        `json:"total_actual_cost"`
	AverageProgress    float64        `json:"average_progress"`
	StatusDistribution map[string]int `json:"status_distribution"`
}
