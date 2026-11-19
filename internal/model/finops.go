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
	ErrExpenseNotFound         = errors.New("it expense not found")
	ErrInvalidCostPool         = errors.New("invalid cost pool")
	ErrInvalidExpenseType      = errors.New("invalid expense type, must be capex or opex")
	ErrInvalidExpenseAmount    = errors.New("expense amount must be greater than zero")
	ErrInvalidPaymentStatus    = errors.New("invalid payment status")
	ErrExpenseDateRequired     = errors.New("expense date is required")
)

// Cost Pool constants
const (
	CostPoolCloudInfrastructure  = "cloud_infrastructure"
	CostPoolSoftwareLicenses    = "software_licenses"
	CostPoolHardwareEquipment    = "hardware_equipment"
	CostPoolTelecomNetwork       = "telecom_network"
	CostPoolPersonnelLabor       = "personnel_labor"
	CostPoolProfessionalServices = "professional_services"
)

// Expense Accounting Classification constants
const (
	ExpenseTypeCapex = "capex"
	ExpenseTypeOpex  = "opex"
)

// Payment Status constants
const (
	PaymentStatusPending  = "pending"
	PaymentStatusApproved = "approved"
	PaymentStatusPaid     = "paid"
	PaymentStatusRejected = "rejected"
)

// ValidateCostPool validates whether the given cost pool is valid
func ValidateCostPool(pool string) error {
	switch strings.ToLower(strings.TrimSpace(pool)) {
	case CostPoolCloudInfrastructure, CostPoolSoftwareLicenses, CostPoolHardwareEquipment,
		CostPoolTelecomNetwork, CostPoolPersonnelLabor, CostPoolProfessionalServices:
		return nil
	default:
		return ErrInvalidCostPool
	}
}

// ValidateExpenseType validates whether the given accounting type is capex or opex
func ValidateExpenseType(expType string) error {
	switch strings.ToLower(strings.TrimSpace(expType)) {
	case ExpenseTypeCapex, ExpenseTypeOpex:
		return nil
	default:
		return ErrInvalidExpenseType
	}
}

// ValidatePaymentStatus validates whether the payment status is valid
func ValidatePaymentStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case PaymentStatusPending, PaymentStatusApproved, PaymentStatusPaid, PaymentStatusRejected:
		return nil
	default:
		return ErrInvalidPaymentStatus
	}
}

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

// ITExpense represents an actual financial expenditure, invoice line item, or cloud cost charge
type ITExpense struct {
	ID            string    `json:"id"`
	WorkspaceID   string    `json:"workspace_id"`
	BudgetID      string    `json:"budget_id"`
	CostCenterID  string    `json:"cost_center_id"`
	VendorID      string    `json:"vendor_id,omitempty"`
	CIID          string    `json:"ci_id,omitempty"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
	Description   string    `json:"description"`
	CostPool      string    `json:"cost_pool"`
	ExpenseType   string    `json:"expense_type"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	ExpenseDate   time.Time `json:"expense_date"`
	PaymentStatus string    `json:"payment_status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Validate checks whether the ITExpense contains valid parameters
func (e *ITExpense) Validate() error {
	if strings.TrimSpace(e.Description) == "" {
		return errors.New("expense description is required")
	}
	if err := ValidateCostPool(e.CostPool); err != nil {
		return err
	}
	if err := ValidateExpenseType(e.ExpenseType); err != nil {
		return err
	}
	if e.Amount <= 0 {
		return ErrInvalidExpenseAmount
	}
	if e.ExpenseDate.IsZero() {
		return ErrExpenseDateRequired
	}
	if e.PaymentStatus != "" {
		if err := ValidatePaymentStatus(e.PaymentStatus); err != nil {
			return err
		}
	}
	return nil
}
