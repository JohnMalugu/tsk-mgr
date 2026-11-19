package service

import (
	"context"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

var _ = strings.TrimSpace

var (
	ErrCostCenterNotFound      = repository.ErrCostCenterNotFound
	ErrCostCenterCodeRequired  = repository.ErrCostCenterCodeRequired
	ErrCostCenterCodeConflict  = repository.ErrCostCenterCodeConflict
	ErrCostCenterNameRequired  = repository.ErrCostCenterNameRequired
	ErrInvalidCostCenterStatus = repository.ErrInvalidCostCenterStatus
	ErrInvalidDepartment       = repository.ErrInvalidDepartment
	ErrBudgetNotFound          = repository.ErrBudgetNotFound
	ErrBudgetNameRequired      = repository.ErrBudgetNameRequired
	ErrInvalidFiscalYear       = repository.ErrInvalidFiscalYear
	ErrInvalidBudgetPeriod     = repository.ErrInvalidBudgetPeriod
	ErrInvalidBudgetStatus     = repository.ErrInvalidBudgetStatus
	ErrInvalidAllocatedAmount  = repository.ErrInvalidAllocatedAmount
	ErrInvalidAlertThreshold   = repository.ErrInvalidAlertThreshold
	ErrExpenseNotFound         = repository.ErrExpenseNotFound
	ErrInvalidCostPool         = repository.ErrInvalidCostPool
	ErrInvalidExpenseType      = repository.ErrInvalidExpenseType
	ErrInvalidExpenseAmount    = repository.ErrInvalidExpenseAmount
	ErrInvalidPaymentStatus    = repository.ErrInvalidPaymentStatus
	ErrExpenseDateRequired     = repository.ErrExpenseDateRequired
	ErrChargebackRuleNotFound  = repository.ErrChargebackRuleNotFound
	ErrInvalidAllocationMethod = repository.ErrInvalidAllocationMethod
	ErrInvalidAllocationValue  = repository.ErrInvalidAllocationValue
	ErrBusinessUnitRequired    = repository.ErrBusinessUnitRequired
	ErrStatementNotFound       = repository.ErrStatementNotFound
	ErrInvalidStatementStatus  = repository.ErrInvalidStatementStatus
)

// FinOpsService defines business logic for IT Financial Management, Budgets, Expenses, and Chargeback
type FinOpsService interface {
	// Cost Centers
	CreateCostCenter(ctx context.Context, workspaceID string, req model.CreateCostCenterRequest) (*model.CostCenter, error)
	GetCostCenterByID(ctx context.Context, id string) (*model.CostCenter, error)
	GetCostCenterByCode(ctx context.Context, workspaceID string, code string) (*model.CostCenter, error)
	UpdateCostCenter(ctx context.Context, id string, req model.UpdateCostCenterRequest) (*model.CostCenter, error)
	DeleteCostCenter(ctx context.Context, id string) error
	ListCostCenters(ctx context.Context, filter model.CostCenterFilterCriteria) ([]model.CostCenter, int, error)

	// IT Budgets
	CreateITBudget(ctx context.Context, workspaceID string, req model.CreateITBudgetRequest) (*model.ITBudget, error)
	GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error)
	UpdateITBudget(ctx context.Context, id string, req model.UpdateITBudgetRequest) (*model.ITBudget, error)
	DeleteITBudget(ctx context.Context, id string) error
	ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error)

	// IT Expenses
	CreateITExpense(ctx context.Context, workspaceID string, req model.CreateITExpenseRequest) (*model.ITExpense, error)
	GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error)
	UpdateITExpense(ctx context.Context, id string, req model.UpdateITExpenseRequest) (*model.ITExpense, error)
	DeleteITExpense(ctx context.Context, id string) error
	ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error)

	// Chargeback
	CreateChargebackRule(ctx context.Context, workspaceID string, req model.CreateChargebackRuleRequest) (*model.ChargebackRule, error)
	GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error)
	UpdateChargebackRule(ctx context.Context, id string, req model.UpdateChargebackRuleRequest) (*model.ChargebackRule, error)
	DeleteChargebackRule(ctx context.Context, id string) error
	ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error)
	CreateChargebackStatement(ctx context.Context, workspaceID string, req model.CreateChargebackStatementRequest) (*model.ChargebackStatement, error)
	GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error)
	UpdateChargebackStatement(ctx context.Context, id string, req model.UpdateChargebackStatementRequest) (*model.ChargebackStatement, error)
	DeleteChargebackStatement(ctx context.Context, id string) error
	ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error)

	// Analytics
	GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error)
}

type finopsServiceImpl struct {
	repo repository.FinOpsRepository
}

// NewFinOpsService creates a new FinOpsService instance
func NewFinOpsService(repo repository.FinOpsRepository) FinOpsService {
	return &finopsServiceImpl{repo: repo}
}

// Cost Centers
func (s *finopsServiceImpl) CreateCostCenter(ctx context.Context, workspaceID string, req model.CreateCostCenterRequest) (*model.CostCenter, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	status := req.Status
	if status == "" {
		status = model.CostCenterStatusActive
	}

	cc := &model.CostCenter{
		WorkspaceID: workspaceID,
		Code:        strings.TrimSpace(req.Code),
		Name:        strings.TrimSpace(req.Name),
		Department:  req.Department,
		ManagerID:   req.ManagerID,
		Status:      status,
		Description: req.Description,
	}

	if err := s.repo.CreateCostCenter(ctx, cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *finopsServiceImpl) GetCostCenterByID(ctx context.Context, id string) (*model.CostCenter, error) {
	return s.repo.GetCostCenterByID(ctx, id)
}

func (s *finopsServiceImpl) GetCostCenterByCode(ctx context.Context, workspaceID string, code string) (*model.CostCenter, error) {
	return s.repo.GetCostCenterByCode(ctx, workspaceID, code)
}

func (s *finopsServiceImpl) UpdateCostCenter(ctx context.Context, id string, req model.UpdateCostCenterRequest) (*model.CostCenter, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	cc, err := s.repo.GetCostCenterByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		cc.Name = strings.TrimSpace(*req.Name)
	}
	if req.Department != nil {
		cc.Department = *req.Department
	}
	if req.ManagerID != nil {
		cc.ManagerID = *req.ManagerID
	}
	if req.Status != nil {
		cc.Status = *req.Status
	}
	if req.Description != nil {
		cc.Description = *req.Description
	}

	if err := s.repo.UpdateCostCenter(ctx, cc); err != nil {
		return nil, err
	}
	return cc, nil
}

func (s *finopsServiceImpl) DeleteCostCenter(ctx context.Context, id string) error {
	return s.repo.DeleteCostCenter(ctx, id)
}

func (s *finopsServiceImpl) ListCostCenters(ctx context.Context, filter model.CostCenterFilterCriteria) ([]model.CostCenter, int, error) {
	return s.repo.ListCostCenters(ctx, filter)
}

// IT Budgets
func (s *finopsServiceImpl) CreateITBudget(ctx context.Context, workspaceID string, req model.CreateITBudgetRequest) (*model.ITBudget, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify cost center exists
	if _, err := s.repo.GetCostCenterByID(ctx, req.CostCenterID); err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	alertThreshold := req.AlertThresholdPercent
	if alertThreshold <= 0 {
		alertThreshold = 80.0
	}
	status := req.Status
	if status == "" {
		status = model.BudgetStatusDraft
	}

	b := &model.ITBudget{
		WorkspaceID:           workspaceID,
		CostCenterID:          req.CostCenterID,
		Name:                  strings.TrimSpace(req.Name),
		FiscalYear:            req.FiscalYear,
		PeriodType:            req.PeriodType,
		AllocatedAmount:       req.AllocatedAmount,
		SpentAmount:           0.0,
		Currency:              currency,
		AlertThresholdPercent: alertThreshold,
		Status:                status,
	}

	if err := s.repo.CreateITBudget(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *finopsServiceImpl) GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error) {
	return s.repo.GetITBudgetByID(ctx, id)
}

func (s *finopsServiceImpl) UpdateITBudget(ctx context.Context, id string, req model.UpdateITBudgetRequest) (*model.ITBudget, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	b, err := s.repo.GetITBudgetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		b.Name = strings.TrimSpace(*req.Name)
	}
	if req.AllocatedAmount != nil {
		b.AllocatedAmount = *req.AllocatedAmount
	}
	if req.AlertThresholdPercent != nil {
		b.AlertThresholdPercent = *req.AlertThresholdPercent
	}
	if req.Status != nil {
		b.Status = *req.Status
	}

	// Update status to exceeded if spent > allocated and currently active
	if b.AllocatedAmount > 0 && b.SpentAmount > b.AllocatedAmount && b.Status == model.BudgetStatusActive {
		b.Status = model.BudgetStatusExceeded
	}

	if err := s.repo.UpdateITBudget(ctx, b); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *finopsServiceImpl) DeleteITBudget(ctx context.Context, id string) error {
	return s.repo.DeleteITBudget(ctx, id)
}

func (s *finopsServiceImpl) ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error) {
	return s.repo.ListITBudgets(ctx, filter)
}

// IT Expenses
func (s *finopsServiceImpl) CreateITExpense(ctx context.Context, workspaceID string, req model.CreateITExpenseRequest) (*model.ITExpense, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify cost center and budget exist
	if _, err := s.repo.GetCostCenterByID(ctx, req.CostCenterID); err != nil {
		return nil, err
	}
	budget, err := s.repo.GetITBudgetByID(ctx, req.BudgetID)
	if err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	paymentStatus := req.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = model.PaymentStatusPending
	}

	expense := &model.ITExpense{
		WorkspaceID:   workspaceID,
		BudgetID:      req.BudgetID,
		CostCenterID:  req.CostCenterID,
		VendorID:      req.VendorID,
		CIID:          req.CIID,
		InvoiceNumber: req.InvoiceNumber,
		Description:   strings.TrimSpace(req.Description),
		CostPool:      req.CostPool,
		ExpenseType:   req.ExpenseType,
		Amount:        req.Amount,
		Currency:      currency,
		ExpenseDate:   req.ExpenseDate,
		PaymentStatus: paymentStatus,
	}

	if err := s.repo.CreateITExpense(ctx, expense); err != nil {
		return nil, err
	}

	// Check if budget exceeded after expense
	newSpent := budget.SpentAmount + expense.Amount
	if budget.AllocatedAmount > 0 && newSpent > budget.AllocatedAmount && budget.Status == model.BudgetStatusActive {
		budget.Status = model.BudgetStatusExceeded
		_ = s.repo.UpdateITBudget(ctx, budget)
	}

	return expense, nil
}

func (s *finopsServiceImpl) GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error) {
	return s.repo.GetITExpenseByID(ctx, id)
}

func (s *finopsServiceImpl) UpdateITExpense(ctx context.Context, id string, req model.UpdateITExpenseRequest) (*model.ITExpense, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	expense, err := s.repo.GetITExpenseByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.VendorID != nil {
		expense.VendorID = *req.VendorID
	}
	if req.CIID != nil {
		expense.CIID = *req.CIID
	}
	if req.InvoiceNumber != nil {
		expense.InvoiceNumber = *req.InvoiceNumber
	}
	if req.Description != nil {
		expense.Description = strings.TrimSpace(*req.Description)
	}
	if req.CostPool != nil {
		expense.CostPool = *req.CostPool
	}
	if req.ExpenseType != nil {
		expense.ExpenseType = *req.ExpenseType
	}
	if req.Amount != nil {
		expense.Amount = *req.Amount
	}
	if req.PaymentStatus != nil {
		expense.PaymentStatus = *req.PaymentStatus
	}
	if req.ExpenseDate != nil {
		expense.ExpenseDate = *req.ExpenseDate
	}

	if err := s.repo.UpdateITExpense(ctx, expense); err != nil {
		return nil, err
	}
	return expense, nil
}

func (s *finopsServiceImpl) DeleteITExpense(ctx context.Context, id string) error {
	return s.repo.DeleteITExpense(ctx, id)
}

func (s *finopsServiceImpl) ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error) {
	return s.repo.ListITExpenses(ctx, filter)
}

// Chargeback Rules
func (s *finopsServiceImpl) CreateChargebackRule(ctx context.Context, workspaceID string, req model.CreateChargebackRuleRequest) (*model.ChargebackRule, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify cost center exists
	if _, err := s.repo.GetCostCenterByID(ctx, req.CostCenterID); err != nil {
		return nil, err
	}

	rule := &model.ChargebackRule{
		WorkspaceID:          workspaceID,
		CostCenterID:         req.CostCenterID,
		CIID:                 req.CIID,
		RuleName:             strings.TrimSpace(req.RuleName),
		TargetBusinessUnit:   strings.TrimSpace(req.TargetBusinessUnit),
		AllocationMethod:     req.AllocationMethod,
		AllocationPercentage: req.AllocationPercentage,
		RatePerUnit:          req.RatePerUnit,
		UnitOfMeasure:        req.UnitOfMeasure,
		IsActive:             true,
	}

	if err := s.repo.CreateChargebackRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *finopsServiceImpl) GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error) {
	return s.repo.GetChargebackRuleByID(ctx, id)
}

func (s *finopsServiceImpl) UpdateChargebackRule(ctx context.Context, id string, req model.UpdateChargebackRuleRequest) (*model.ChargebackRule, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	rule, err := s.repo.GetChargebackRuleByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.RuleName != nil {
		rule.RuleName = strings.TrimSpace(*req.RuleName)
	}
	if req.TargetBusinessUnit != nil {
		rule.TargetBusinessUnit = strings.TrimSpace(*req.TargetBusinessUnit)
	}
	if req.AllocationMethod != nil {
		rule.AllocationMethod = *req.AllocationMethod
	}
	if req.AllocationPercentage != nil {
		rule.AllocationPercentage = *req.AllocationPercentage
	}
	if req.RatePerUnit != nil {
		rule.RatePerUnit = *req.RatePerUnit
	}
	if req.UnitOfMeasure != nil {
		rule.UnitOfMeasure = *req.UnitOfMeasure
	}
	if req.IsActive != nil {
		rule.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateChargebackRule(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *finopsServiceImpl) DeleteChargebackRule(ctx context.Context, id string) error {
	return s.repo.DeleteChargebackRule(ctx, id)
}

func (s *finopsServiceImpl) ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error) {
	return s.repo.ListChargebackRules(ctx, filter)
}

// Chargeback Statements
func (s *finopsServiceImpl) CreateChargebackStatement(ctx context.Context, workspaceID string, req model.CreateChargebackStatementRequest) (*model.ChargebackStatement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	stmt := &model.ChargebackStatement{
		WorkspaceID:        workspaceID,
		TargetBusinessUnit: strings.TrimSpace(req.TargetBusinessUnit),
		StatementPeriod:    strings.TrimSpace(req.StatementPeriod),
		TotalCharge:        req.TotalCharge,
		Currency:           currency,
		Status:             model.StatementStatusDraft,
		Notes:              req.Notes,
	}

	if err := s.repo.CreateChargebackStatement(ctx, stmt); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (s *finopsServiceImpl) GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error) {
	return s.repo.GetChargebackStatementByID(ctx, id)
}

func (s *finopsServiceImpl) UpdateChargebackStatement(ctx context.Context, id string, req model.UpdateChargebackStatementRequest) (*model.ChargebackStatement, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	stmt, err := s.repo.GetChargebackStatementByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Status != nil {
		stmt.Status = *req.Status
		if *req.Status == model.StatementStatusSettled && stmt.SettledAt == nil {
			now := time.Now().UTC()
			stmt.SettledAt = &now
		}
	}
	if req.Notes != nil {
		stmt.Notes = *req.Notes
	}

	if err := s.repo.UpdateChargebackStatement(ctx, stmt); err != nil {
		return nil, err
	}
	return stmt, nil
}

func (s *finopsServiceImpl) DeleteChargebackStatement(ctx context.Context, id string) error {
	return s.repo.DeleteChargebackStatement(ctx, id)
}

func (s *finopsServiceImpl) ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error) {
	return s.repo.ListChargebackStatements(ctx, filter)
}

// Analytics
func (s *finopsServiceImpl) GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error) {
	if fiscalYear <= 0 {
		fiscalYear = time.Now().UTC().Year()
	}
	return s.repo.GetFinOpsAnalytics(ctx, workspaceID, fiscalYear)
}
