package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
)

var (
	_ = uuid.New
	_ = fmt.Sprintf
	_ = strings.Join
	_ = time.Now
	_ = errors.New
)

var (
	ErrCostCenterNotFound      = model.ErrCostCenterNotFound
	ErrCostCenterCodeRequired  = model.ErrCostCenterCodeRequired
	ErrCostCenterCodeConflict  = model.ErrCostCenterCodeConflict
	ErrCostCenterNameRequired  = model.ErrCostCenterNameRequired
	ErrInvalidCostCenterStatus = model.ErrInvalidCostCenterStatus
	ErrInvalidDepartment       = model.ErrInvalidDepartment
	ErrBudgetNotFound          = model.ErrBudgetNotFound
	ErrBudgetNameRequired      = model.ErrBudgetNameRequired
	ErrInvalidFiscalYear       = model.ErrInvalidFiscalYear
	ErrInvalidBudgetPeriod     = model.ErrInvalidBudgetPeriod
	ErrInvalidBudgetStatus     = model.ErrInvalidBudgetStatus
	ErrInvalidAllocatedAmount  = model.ErrInvalidAllocatedAmount
	ErrInvalidAlertThreshold   = model.ErrInvalidAlertThreshold
	ErrExpenseNotFound         = model.ErrExpenseNotFound
	ErrInvalidCostPool         = model.ErrInvalidCostPool
	ErrInvalidExpenseType      = model.ErrInvalidExpenseType
	ErrInvalidExpenseAmount    = model.ErrInvalidExpenseAmount
	ErrInvalidPaymentStatus    = model.ErrInvalidPaymentStatus
	ErrExpenseDateRequired     = model.ErrExpenseDateRequired
	ErrChargebackRuleNotFound  = model.ErrChargebackRuleNotFound
	ErrInvalidAllocationMethod = model.ErrInvalidAllocationMethod
	ErrInvalidAllocationValue  = model.ErrInvalidAllocationValue
	ErrBusinessUnitRequired    = model.ErrBusinessUnitRequired
	ErrStatementNotFound       = model.ErrStatementNotFound
	ErrInvalidStatementStatus  = model.ErrInvalidStatementStatus
)

// FinOpsRepository defines the data access contract for IT Financial Management and FinOps
type FinOpsRepository interface {
	// Cost Centers
	CreateCostCenter(ctx context.Context, cc *model.CostCenter) error
	GetCostCenterByID(ctx context.Context, id string) (*model.CostCenter, error)
	GetCostCenterByCode(ctx context.Context, workspaceID string, code string) (*model.CostCenter, error)
	UpdateCostCenter(ctx context.Context, cc *model.CostCenter) error
	DeleteCostCenter(ctx context.Context, id string) error
	ListCostCenters(ctx context.Context, filter model.CostCenterFilterCriteria) ([]model.CostCenter, int, error)

	// IT Budgets
	CreateITBudget(ctx context.Context, b *model.ITBudget) error
	GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error)
	UpdateITBudget(ctx context.Context, b *model.ITBudget) error
	DeleteITBudget(ctx context.Context, id string) error
	ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error)

	// IT Expenses
	CreateITExpense(ctx context.Context, e *model.ITExpense) error
	GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error)
	UpdateITExpense(ctx context.Context, e *model.ITExpense) error
	DeleteITExpense(ctx context.Context, id string) error
	ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error)

	// Chargeback
	CreateChargebackRule(ctx context.Context, r *model.ChargebackRule) error
	GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error)
	UpdateChargebackRule(ctx context.Context, r *model.ChargebackRule) error
	DeleteChargebackRule(ctx context.Context, id string) error
	ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error)
	CreateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error
	GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error)
	UpdateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error
	DeleteChargebackStatement(ctx context.Context, id string) error
	ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error)

	// Analytics
	GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error)
}

type sqliteFinOpsRepository struct {
	db *sql.DB
}

// NewFinOpsRepository creates a new FinOpsRepository instance
func NewFinOpsRepository(db *sql.DB) FinOpsRepository {
	return &sqliteFinOpsRepository{db: db}
}

// Cost Centers
func (r *sqliteFinOpsRepository) CreateCostCenter(ctx context.Context, cc *model.CostCenter) error {
	if cc.ID == "" {
		cc.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if cc.CreatedAt.IsZero() {
		cc.CreatedAt = now
	}
	cc.UpdatedAt = now
	if cc.Status == "" {
		cc.Status = model.CostCenterStatusActive
	}

	query := `
	INSERT INTO cost_centers (
		id, workspace_id, code, name, department, manager_id,
		status, description, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		cc.ID, cc.WorkspaceID, cc.Code, cc.Name, cc.Department, cc.ManagerID,
		cc.Status, cc.Description, cc.CreatedAt, cc.UpdatedAt,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return ErrCostCenterCodeConflict
		}
		return err
	}
	return nil
}

func (r *sqliteFinOpsRepository) GetCostCenterByID(ctx context.Context, id string) (*model.CostCenter, error) {
	query := `
	SELECT id, workspace_id, code, name, department, manager_id,
	       status, description, created_at, updated_at
	FROM cost_centers WHERE id = ?`

	var cc model.CostCenter
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&cc.ID, &cc.WorkspaceID, &cc.Code, &cc.Name, &cc.Department, &cc.ManagerID,
		&cc.Status, &cc.Description, &cc.CreatedAt, &cc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCostCenterNotFound
		}
		return nil, err
	}
	return &cc, nil
}

func (r *sqliteFinOpsRepository) GetCostCenterByCode(ctx context.Context, workspaceID string, code string) (*model.CostCenter, error) {
	query := `
	SELECT id, workspace_id, code, name, department, manager_id,
	       status, description, created_at, updated_at
	FROM cost_centers WHERE workspace_id = ? AND code = ?`

	var cc model.CostCenter
	err := r.db.QueryRowContext(ctx, query, workspaceID, code).Scan(
		&cc.ID, &cc.WorkspaceID, &cc.Code, &cc.Name, &cc.Department, &cc.ManagerID,
		&cc.Status, &cc.Description, &cc.CreatedAt, &cc.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrCostCenterNotFound
		}
		return nil, err
	}
	return &cc, nil
}

func (r *sqliteFinOpsRepository) UpdateCostCenter(ctx context.Context, cc *model.CostCenter) error {
	cc.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE cost_centers SET
		name = ?, department = ?, manager_id = ?,
		status = ?, description = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		cc.Name, cc.Department, cc.ManagerID,
		cc.Status, cc.Description, cc.UpdatedAt,
		cc.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCostCenterNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) DeleteCostCenter(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM cost_centers WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrCostCenterNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) ListCostCenters(ctx context.Context, filter model.CostCenterFilterCriteria) ([]model.CostCenter, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.Department != "" {
		whereClauses = append(whereClauses, "department = ?")
		args = append(args, filter.Department)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		whereClauses = append(whereClauses, "(name LIKE ? OR code LIKE ?)")
		s := "%" + filter.Search + "%"
		args = append(args, s, s)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM cost_centers WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := fmt.Sprintf(`
	SELECT id, workspace_id, code, name, department, manager_id,
	       status, description, created_at, updated_at
	FROM cost_centers WHERE %s ORDER BY code ASC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.CostCenter{}
	for rows.Next() {
		var cc model.CostCenter
		if err := rows.Scan(
			&cc.ID, &cc.WorkspaceID, &cc.Code, &cc.Name, &cc.Department, &cc.ManagerID,
			&cc.Status, &cc.Description, &cc.CreatedAt, &cc.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, cc)
	}
	return list, total, rows.Err()
}

// IT Budget Stubs
func (r *sqliteFinOpsRepository) CreateITBudget(ctx context.Context, b *model.ITBudget) error {
	return nil
}
func (r *sqliteFinOpsRepository) GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error) {
	return nil, ErrBudgetNotFound
}
func (r *sqliteFinOpsRepository) UpdateITBudget(ctx context.Context, b *model.ITBudget) error {
	return nil
}
func (r *sqliteFinOpsRepository) DeleteITBudget(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteFinOpsRepository) ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error) {
	return nil, 0, nil
}

// IT Expense Stubs
func (r *sqliteFinOpsRepository) CreateITExpense(ctx context.Context, e *model.ITExpense) error {
	return nil
}
func (r *sqliteFinOpsRepository) GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error) {
	return nil, ErrExpenseNotFound
}
func (r *sqliteFinOpsRepository) UpdateITExpense(ctx context.Context, e *model.ITExpense) error {
	return nil
}
func (r *sqliteFinOpsRepository) DeleteITExpense(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteFinOpsRepository) ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error) {
	return nil, 0, nil
}

// Chargeback Stubs
func (r *sqliteFinOpsRepository) CreateChargebackRule(ctx context.Context, rule *model.ChargebackRule) error {
	return nil
}
func (r *sqliteFinOpsRepository) GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error) {
	return nil, ErrChargebackRuleNotFound
}
func (r *sqliteFinOpsRepository) UpdateChargebackRule(ctx context.Context, rule *model.ChargebackRule) error {
	return nil
}
func (r *sqliteFinOpsRepository) DeleteChargebackRule(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteFinOpsRepository) ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error) {
	return nil, 0, nil
}
func (r *sqliteFinOpsRepository) CreateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error {
	return nil
}
func (r *sqliteFinOpsRepository) GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error) {
	return nil, ErrStatementNotFound
}
func (r *sqliteFinOpsRepository) UpdateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error {
	return nil
}
func (r *sqliteFinOpsRepository) DeleteChargebackStatement(ctx context.Context, id string) error {
	return nil
}
func (r *sqliteFinOpsRepository) ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error) {
	return nil, 0, nil
}

// Analytics Stub
func (r *sqliteFinOpsRepository) GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error) {
	return &model.FinOpsAnalyticsReport{WorkspaceID: workspaceID, FiscalYear: fiscalYear}, nil
}
