package model

import (
	"errors"
	"strings"
	"time"
)

// Department constants for IT Cost Centers
const (
	DeptEngineering    = "engineering"
	DeptInfrastructure = "infrastructure"
	DeptSecurity       = "security"
	DeptProduct        = "product"
	DeptOperations     = "operations"
	DeptFinance        = "finance"
)

// Cost Center Status constants
const (
	CostCenterStatusActive   = "active"
	CostCenterStatusInactive = "inactive"
	CostCenterStatusArchived = "archived"
)

// FinOps domain errors
var (
	ErrCostCenterNotFound     = errors.New("cost center not found")
	ErrCostCenterCodeRequired = errors.New("cost center code is required")
	ErrCostCenterCodeConflict = errors.New("cost center with this code already exists in workspace")
	ErrCostCenterNameRequired = errors.New("cost center name is required")
	ErrInvalidCostCenterStatus = errors.New("invalid cost center status")
	ErrInvalidDepartment       = errors.New("invalid department")
	ErrBudgetNotFound          = errors.New("it budget not found")
	ErrBudgetNameRequired      = errors.New("budget name is required")
	ErrInvalidFiscalYear       = errors.New("fiscal year must be a valid 4-digit year")
	ErrInvalidBudgetPeriod     = errors.New("invalid budget period")
	ErrInvalidBudgetStatus     = errors.New("invalid budget status")
	ErrInvalidAllocatedAmount  = errors.New("allocated amount must be greater than zero")
	ErrInvalidAlertThreshold   = errors.New("alert threshold must be between 1 and 100 percent")
)

// Fiscal Period constants
const (
	BudgetPeriodAnnual    = "annual"
	BudgetPeriodQuarterly = "quarterly"
	BudgetPeriodMonthly   = "monthly"
)

// Budget Status constants
const (
	BudgetStatusDraft    = "draft"
	BudgetStatusApproved = "approved"
	BudgetStatusActive   = "active"
	BudgetStatusExceeded = "exceeded"
	BudgetStatusClosed   = "closed"
)

// ValidateBudgetPeriod validates whether the given period is valid
func ValidateBudgetPeriod(period string) error {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case BudgetPeriodAnnual, BudgetPeriodQuarterly, BudgetPeriodMonthly:
		return nil
	default:
		return ErrInvalidBudgetPeriod
	}
}

// ValidateBudgetStatus validates whether the given budget status is valid
func ValidateBudgetStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case BudgetStatusDraft, BudgetStatusApproved, BudgetStatusActive, BudgetStatusExceeded, BudgetStatusClosed:
		return nil
	default:
		return ErrInvalidBudgetStatus
	}
}

// ValidateCostCenterStatus validates whether the given status is recognized
func ValidateCostCenterStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case CostCenterStatusActive, CostCenterStatusInactive, CostCenterStatusArchived:
		return nil
	default:
		return ErrInvalidCostCenterStatus
	}
}

// ValidateDepartment validates whether the given department is recognized
func ValidateDepartment(dept string) error {
	switch strings.ToLower(strings.TrimSpace(dept)) {
	case DeptEngineering, DeptInfrastructure, DeptSecurity, DeptProduct, DeptOperations, DeptFinance:
		return nil
	default:
		return ErrInvalidDepartment
	}
}

// CostCenter represents an accounting and operational unit for IT spending and cost attribution
type CostCenter struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Department  string    `json:"department"`
	ManagerID   string    `json:"manager_id"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ITBudget represents a dedicated financial budget allocation for an IT cost center or period
type ITBudget struct {
	ID                    string    `json:"id"`
	WorkspaceID           string    `json:"workspace_id"`
	CostCenterID          string    `json:"cost_center_id"`
	Name                  string    `json:"name"`
	FiscalYear            int       `json:"fiscal_year"`
	PeriodType            string    `json:"period_type"`
	AllocatedAmount       float64   `json:"allocated_amount"`
	SpentAmount           float64   `json:"spent_amount"`
	Currency              string    `json:"currency"`
	AlertThresholdPercent float64   `json:"alert_threshold_percent"`
	Status                string    `json:"status"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// Validate checks whether the ITBudget fields are logically sound
func (b *ITBudget) Validate() error {
	if strings.TrimSpace(b.Name) == "" {
		return ErrBudgetNameRequired
	}
	if b.FiscalYear < 2000 || b.FiscalYear > 2100 {
		return ErrInvalidFiscalYear
	}
	if err := ValidateBudgetPeriod(b.PeriodType); err != nil {
		return err
	}
	if b.AllocatedAmount <= 0 {
		return ErrInvalidAllocatedAmount
	}
	if b.AlertThresholdPercent < 1.0 || b.AlertThresholdPercent > 100.0 {
		return ErrInvalidAlertThreshold
	}
	if b.Status != "" {
		if err := ValidateBudgetStatus(b.Status); err != nil {
			return err
		}
	}
	return nil
}

// UtilizationPercent calculates the ratio of spent funds to total allocated budget
func (b *ITBudget) UtilizationPercent() float64 {
	if b.AllocatedAmount <= 0 {
		return 0
	}
	return (b.SpentAmount / b.AllocatedAmount) * 100.0
}

// IsThresholdBreached checks if the spent ratio has met or passed the early warning threshold
func (b *ITBudget) IsThresholdBreached() bool {
	return b.UtilizationPercent() >= b.AlertThresholdPercent
}
