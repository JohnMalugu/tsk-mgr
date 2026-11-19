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
	ErrChargebackRuleNotFound  = errors.New("chargeback rule not found")
	ErrInvalidAllocationMethod = errors.New("invalid allocation method")
	ErrInvalidAllocationValue  = errors.New("allocation value must be greater than zero")
	ErrBusinessUnitRequired    = errors.New("target business unit is required")
	ErrStatementNotFound       = errors.New("chargeback statement not found")
	ErrInvalidStatementStatus  = errors.New("invalid chargeback statement status")
)

// Chargeback Allocation Method constants
const (
	AllocMethodFixedPercent = "fixed_percent"
	AllocMethodUsageUnits   = "usage_units"
	AllocMethodEvenSplit    = "even_split"
)

// Chargeback Statement Status constants
const (
	StatementStatusDraft    = "draft"
	StatementStatusIssued   = "issued"
	StatementStatusSettled  = "settled"
	StatementStatusDisputed = "disputed"
)

// ValidateAllocationMethod validates whether the allocation method is recognized
func ValidateAllocationMethod(method string) error {
	switch strings.ToLower(strings.TrimSpace(method)) {
	case AllocMethodFixedPercent, AllocMethodUsageUnits, AllocMethodEvenSplit:
		return nil
	default:
		return ErrInvalidAllocationMethod
	}
}

// ValidateStatementStatus validates whether the statement status is recognized
func ValidateStatementStatus(status string) error {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case StatementStatusDraft, StatementStatusIssued, StatementStatusSettled, StatementStatusDisputed:
		return nil
	default:
		return ErrInvalidStatementStatus
	}
}

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

// ChargebackRule defines a policy for attributing IT costs to receiving business units
type ChargebackRule struct {
	ID                   string    `json:"id"`
	WorkspaceID          string    `json:"workspace_id"`
	CostCenterID         string    `json:"cost_center_id"`
	CIID                 string    `json:"ci_id,omitempty"`
	RuleName             string    `json:"rule_name"`
	TargetBusinessUnit   string    `json:"target_business_unit"`
	AllocationMethod     string    `json:"allocation_method"`
	AllocationPercentage float64   `json:"allocation_percentage"`
	RatePerUnit          float64   `json:"rate_per_unit"`
	UnitOfMeasure        string    `json:"unit_of_measure,omitempty"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Validate checks whether the ChargebackRule has sound attributes
func (r *ChargebackRule) Validate() error {
	if strings.TrimSpace(r.RuleName) == "" {
		return errors.New("chargeback rule_name is required")
	}
	if strings.TrimSpace(r.TargetBusinessUnit) == "" {
		return ErrBusinessUnitRequired
	}
	if err := ValidateAllocationMethod(r.AllocationMethod); err != nil {
		return err
	}
	if r.AllocationMethod == AllocMethodFixedPercent && (r.AllocationPercentage <= 0.0 || r.AllocationPercentage > 100.0) {
		return ErrInvalidAllocationValue
	}
	return nil
}

// ChargebackStatement represents an aggregated billing statement issued to a department or business unit
type ChargebackStatement struct {
	ID                 string     `json:"id"`
	WorkspaceID        string     `json:"workspace_id"`
	TargetBusinessUnit string     `json:"target_business_unit"`
	StatementPeriod    string     `json:"statement_period"`
	TotalCharge        float64    `json:"total_charge"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	Notes              string     `json:"notes,omitempty"`
	GeneratedAt        time.Time  `json:"generated_at"`
	SettledAt          *time.Time `json:"settled_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// Validate checks whether the ChargebackStatement parameters are valid
func (s *ChargebackStatement) Validate() error {
	if strings.TrimSpace(s.TargetBusinessUnit) == "" {
		return ErrBusinessUnitRequired
	}
	if strings.TrimSpace(s.StatementPeriod) == "" {
		return errors.New("statement_period is required")
	}
	if s.TotalCharge < 0.0 {
		return errors.New("total_charge cannot be negative")
	}
	if s.Status != "" {
		if err := ValidateStatementStatus(s.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateCostCenterRequest defines payload to create a new Cost Center
type CreateCostCenterRequest struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Department  string `json:"department"`
	ManagerID   string `json:"manager_id,omitempty"`
	Status      string `json:"status,omitempty"`
	Description string `json:"description,omitempty"`
}

// Validate validates CreateCostCenterRequest
func (r *CreateCostCenterRequest) Validate() error {
	if strings.TrimSpace(r.Code) == "" {
		return ErrCostCenterCodeRequired
	}
	if strings.TrimSpace(r.Name) == "" {
		return ErrCostCenterNameRequired
	}
	if err := ValidateDepartment(r.Department); err != nil {
		return err
	}
	if r.Status != "" {
		if err := ValidateCostCenterStatus(r.Status); err != nil {
			return err
		}
	}
	return nil
}

// UpdateCostCenterRequest defines payload to update an existing Cost Center
type UpdateCostCenterRequest struct {
	Name        *string `json:"name,omitempty"`
	Department  *string `json:"department,omitempty"`
	ManagerID   *string `json:"manager_id,omitempty"`
	Status      *string `json:"status,omitempty"`
	Description *string `json:"description,omitempty"`
}

// Validate validates UpdateCostCenterRequest
func (r *UpdateCostCenterRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrCostCenterNameRequired
	}
	if r.Department != nil {
		if err := ValidateDepartment(*r.Department); err != nil {
			return err
		}
	}
	if r.Status != nil {
		if err := ValidateCostCenterStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateITBudgetRequest defines payload to create a budget allocation
type CreateITBudgetRequest struct {
	CostCenterID          string  `json:"cost_center_id"`
	Name                  string  `json:"name"`
	FiscalYear            int     `json:"fiscal_year"`
	PeriodType            string  `json:"period_type"`
	AllocatedAmount       float64 `json:"allocated_amount"`
	Currency              string  `json:"currency,omitempty"`
	AlertThresholdPercent float64 `json:"alert_threshold_percent,omitempty"`
	Status                string  `json:"status,omitempty"`
}

// Validate validates CreateITBudgetRequest
func (r *CreateITBudgetRequest) Validate() error {
	if strings.TrimSpace(r.CostCenterID) == "" {
		return ErrCostCenterNotFound
	}
	if strings.TrimSpace(r.Name) == "" {
		return ErrBudgetNameRequired
	}
	if r.FiscalYear < 2000 || r.FiscalYear > 2100 {
		return ErrInvalidFiscalYear
	}
	if err := ValidateBudgetPeriod(r.PeriodType); err != nil {
		return err
	}
	if r.AllocatedAmount <= 0 {
		return ErrInvalidAllocatedAmount
	}
	if r.AlertThresholdPercent != 0.0 && (r.AlertThresholdPercent < 1.0 || r.AlertThresholdPercent > 100.0) {
		return ErrInvalidAlertThreshold
	}
	if r.Status != "" {
		if err := ValidateBudgetStatus(r.Status); err != nil {
			return err
		}
	}
	return nil
}

// UpdateITBudgetRequest defines payload to modify budget allocations and thresholds
type UpdateITBudgetRequest struct {
	Name                  *string  `json:"name,omitempty"`
	AllocatedAmount       *float64 `json:"allocated_amount,omitempty"`
	AlertThresholdPercent *float64 `json:"alert_threshold_percent,omitempty"`
	Status                *string  `json:"status,omitempty"`
}

// Validate validates UpdateITBudgetRequest
func (r *UpdateITBudgetRequest) Validate() error {
	if r.Name != nil && strings.TrimSpace(*r.Name) == "" {
		return ErrBudgetNameRequired
	}
	if r.AllocatedAmount != nil && *r.AllocatedAmount <= 0 {
		return ErrInvalidAllocatedAmount
	}
	if r.AlertThresholdPercent != nil && (*r.AlertThresholdPercent < 1.0 || *r.AlertThresholdPercent > 100.0) {
		return ErrInvalidAlertThreshold
	}
	if r.Status != nil {
		if err := ValidateBudgetStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}

// CreateITExpenseRequest defines payload to record an expense
type CreateITExpenseRequest struct {
	BudgetID      string    `json:"budget_id"`
	CostCenterID  string    `json:"cost_center_id"`
	VendorID      string    `json:"vendor_id,omitempty"`
	CIID          string    `json:"ci_id,omitempty"`
	InvoiceNumber string    `json:"invoice_number,omitempty"`
	Description   string    `json:"description"`
	CostPool      string    `json:"cost_pool"`
	ExpenseType   string    `json:"expense_type"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency,omitempty"`
	ExpenseDate   time.Time `json:"expense_date"`
	PaymentStatus string    `json:"payment_status,omitempty"`
}

// Validate validates CreateITExpenseRequest
func (r *CreateITExpenseRequest) Validate() error {
	if strings.TrimSpace(r.BudgetID) == "" {
		return ErrBudgetNotFound
	}
	if strings.TrimSpace(r.CostCenterID) == "" {
		return ErrCostCenterNotFound
	}
	if strings.TrimSpace(r.Description) == "" {
		return errors.New("expense description is required")
	}
	if err := ValidateCostPool(r.CostPool); err != nil {
		return err
	}
	if err := ValidateExpenseType(r.ExpenseType); err != nil {
		return err
	}
	if r.Amount <= 0 {
		return ErrInvalidExpenseAmount
	}
	if r.ExpenseDate.IsZero() {
		return ErrExpenseDateRequired
	}
	if r.PaymentStatus != "" {
		if err := ValidatePaymentStatus(r.PaymentStatus); err != nil {
			return err
		}
	}
	return nil
}

// UpdateITExpenseRequest defines payload to modify an expense
type UpdateITExpenseRequest struct {
	VendorID      *string    `json:"vendor_id,omitempty"`
	CIID          *string    `json:"ci_id,omitempty"`
	InvoiceNumber *string    `json:"invoice_number,omitempty"`
	Description   *string    `json:"description,omitempty"`
	CostPool      *string    `json:"cost_pool,omitempty"`
	ExpenseType   *string    `json:"expense_type,omitempty"`
	Amount        *float64   `json:"amount,omitempty"`
	PaymentStatus *string    `json:"payment_status,omitempty"`
	ExpenseDate   *time.Time `json:"expense_date,omitempty"`
}

// Validate validates UpdateITExpenseRequest
func (r *UpdateITExpenseRequest) Validate() error {
	if r.Description != nil && strings.TrimSpace(*r.Description) == "" {
		return errors.New("expense description is required")
	}
	if r.CostPool != nil {
		if err := ValidateCostPool(*r.CostPool); err != nil {
			return err
		}
	}
	if r.ExpenseType != nil {
		if err := ValidateExpenseType(*r.ExpenseType); err != nil {
			return err
		}
	}
	if r.Amount != nil && *r.Amount <= 0 {
		return ErrInvalidExpenseAmount
	}
	if r.PaymentStatus != nil {
		if err := ValidatePaymentStatus(*r.PaymentStatus); err != nil {
			return err
		}
	}
	return nil
}

// CreateChargebackRuleRequest defines payload to create a cost allocation rule
type CreateChargebackRuleRequest struct {
	CostCenterID         string  `json:"cost_center_id"`
	CIID                 string  `json:"ci_id,omitempty"`
	RuleName             string  `json:"rule_name"`
	TargetBusinessUnit   string  `json:"target_business_unit"`
	AllocationMethod     string  `json:"allocation_method"`
	AllocationPercentage float64 `json:"allocation_percentage,omitempty"`
	RatePerUnit          float64 `json:"rate_per_unit,omitempty"`
	UnitOfMeasure        string  `json:"unit_of_measure,omitempty"`
}

// Validate validates CreateChargebackRuleRequest
func (r *CreateChargebackRuleRequest) Validate() error {
	if strings.TrimSpace(r.CostCenterID) == "" {
		return ErrCostCenterNotFound
	}
	if strings.TrimSpace(r.RuleName) == "" {
		return errors.New("chargeback rule_name is required")
	}
	if strings.TrimSpace(r.TargetBusinessUnit) == "" {
		return ErrBusinessUnitRequired
	}
	if err := ValidateAllocationMethod(r.AllocationMethod); err != nil {
		return err
	}
	if r.AllocationMethod == AllocMethodFixedPercent && (r.AllocationPercentage <= 0.0 || r.AllocationPercentage > 100.0) {
		return ErrInvalidAllocationValue
	}
	return nil
}

// UpdateChargebackRuleRequest defines payload to update a chargeback rule
type UpdateChargebackRuleRequest struct {
	RuleName             *string  `json:"rule_name,omitempty"`
	TargetBusinessUnit   *string  `json:"target_business_unit,omitempty"`
	AllocationMethod     *string  `json:"allocation_method,omitempty"`
	AllocationPercentage *float64 `json:"allocation_percentage,omitempty"`
	RatePerUnit          *float64 `json:"rate_per_unit,omitempty"`
	UnitOfMeasure        *string  `json:"unit_of_measure,omitempty"`
	IsActive             *bool    `json:"is_active,omitempty"`
}

// Validate validates UpdateChargebackRuleRequest
func (r *UpdateChargebackRuleRequest) Validate() error {
	if r.RuleName != nil && strings.TrimSpace(*r.RuleName) == "" {
		return errors.New("chargeback rule_name is required")
	}
	if r.TargetBusinessUnit != nil && strings.TrimSpace(*r.TargetBusinessUnit) == "" {
		return ErrBusinessUnitRequired
	}
	if r.AllocationMethod != nil {
		if err := ValidateAllocationMethod(*r.AllocationMethod); err != nil {
			return err
		}
	}
	if r.AllocationPercentage != nil && (*r.AllocationPercentage <= 0.0 || *r.AllocationPercentage > 100.0) {
		return ErrInvalidAllocationValue
	}
	return nil
}

// CreateChargebackStatementRequest defines payload to generate a statement
type CreateChargebackStatementRequest struct {
	TargetBusinessUnit string  `json:"target_business_unit"`
	StatementPeriod    string  `json:"statement_period"`
	TotalCharge        float64 `json:"total_charge"`
	Currency           string  `json:"currency,omitempty"`
	Notes              string  `json:"notes,omitempty"`
}

// Validate validates CreateChargebackStatementRequest
func (r *CreateChargebackStatementRequest) Validate() error {
	if strings.TrimSpace(r.TargetBusinessUnit) == "" {
		return ErrBusinessUnitRequired
	}
	if strings.TrimSpace(r.StatementPeriod) == "" {
		return errors.New("statement_period is required")
	}
	if r.TotalCharge < 0.0 {
		return errors.New("total_charge cannot be negative")
	}
	return nil
}

// UpdateChargebackStatementRequest defines payload to update statement status
type UpdateChargebackStatementRequest struct {
	Status *string `json:"status,omitempty"`
	Notes  *string `json:"notes,omitempty"`
}

// Validate validates UpdateChargebackStatementRequest
func (r *UpdateChargebackStatementRequest) Validate() error {
	if r.Status != nil {
		if err := ValidateStatementStatus(*r.Status); err != nil {
			return err
		}
	}
	return nil
}
