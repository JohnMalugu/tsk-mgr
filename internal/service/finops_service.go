package service

import (
	"context"
	"strings"

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

// IT Budget Stubs
func (s *finopsServiceImpl) CreateITBudget(ctx context.Context, workspaceID string, req model.CreateITBudgetRequest) (*model.ITBudget, error) {
	return nil, nil
}
func (s *finopsServiceImpl) GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error) {
	return nil, ErrBudgetNotFound
}
func (s *finopsServiceImpl) UpdateITBudget(ctx context.Context, id string, req model.UpdateITBudgetRequest) (*model.ITBudget, error) {
	return nil, ErrBudgetNotFound
}
func (s *finopsServiceImpl) DeleteITBudget(ctx context.Context, id string) error {
	return nil
}
func (s *finopsServiceImpl) ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error) {
	return nil, 0, nil
}

// IT Expense Stubs
func (s *finopsServiceImpl) CreateITExpense(ctx context.Context, workspaceID string, req model.CreateITExpenseRequest) (*model.ITExpense, error) {
	return nil, nil
}
func (s *finopsServiceImpl) GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error) {
	return nil, ErrExpenseNotFound
}
func (s *finopsServiceImpl) UpdateITExpense(ctx context.Context, id string, req model.UpdateITExpenseRequest) (*model.ITExpense, error) {
	return nil, ErrExpenseNotFound
}
func (s *finopsServiceImpl) DeleteITExpense(ctx context.Context, id string) error {
	return nil
}
func (s *finopsServiceImpl) ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error) {
	return nil, 0, nil
}

// Chargeback Stubs
func (s *finopsServiceImpl) CreateChargebackRule(ctx context.Context, workspaceID string, req model.CreateChargebackRuleRequest) (*model.ChargebackRule, error) {
	return nil, nil
}
func (s *finopsServiceImpl) GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error) {
	return nil, ErrChargebackRuleNotFound
}
func (s *finopsServiceImpl) UpdateChargebackRule(ctx context.Context, id string, req model.UpdateChargebackRuleRequest) (*model.ChargebackRule, error) {
	return nil, ErrChargebackRuleNotFound
}
func (s *finopsServiceImpl) DeleteChargebackRule(ctx context.Context, id string) error {
	return nil
}
func (s *finopsServiceImpl) ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error) {
	return nil, 0, nil
}
func (s *finopsServiceImpl) CreateChargebackStatement(ctx context.Context, workspaceID string, req model.CreateChargebackStatementRequest) (*model.ChargebackStatement, error) {
	return nil, nil
}
func (s *finopsServiceImpl) GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error) {
	return nil, ErrStatementNotFound
}
func (s *finopsServiceImpl) UpdateChargebackStatement(ctx context.Context, id string, req model.UpdateChargebackStatementRequest) (*model.ChargebackStatement, error) {
	return nil, ErrStatementNotFound
}
func (s *finopsServiceImpl) DeleteChargebackStatement(ctx context.Context, id string) error {
	return nil
}
func (s *finopsServiceImpl) ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error) {
	return nil, 0, nil
}

// Analytics Stub
func (s *finopsServiceImpl) GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error) {
	return &model.FinOpsAnalyticsReport{WorkspaceID: workspaceID, FiscalYear: fiscalYear}, nil
}
