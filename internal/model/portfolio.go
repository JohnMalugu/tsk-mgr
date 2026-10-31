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

