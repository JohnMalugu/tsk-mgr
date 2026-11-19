package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
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

// IT Budgets
func (r *sqliteFinOpsRepository) CreateITBudget(ctx context.Context, b *model.ITBudget) error {
	if b.ID == "" {
		b.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if b.CreatedAt.IsZero() {
		b.CreatedAt = now
	}
	b.UpdatedAt = now
	if b.Currency == "" {
		b.Currency = "USD"
	}
	if b.PeriodType == "" {
		b.PeriodType = model.BudgetPeriodAnnual
	}
	if b.Status == "" {
		b.Status = model.BudgetStatusDraft
	}
	if b.AlertThresholdPercent <= 0 {
		b.AlertThresholdPercent = 80.0
	}

	query := `
	INSERT INTO it_budgets (
		id, workspace_id, cost_center_id, name, fiscal_year,
		period_type, allocated_amount, spent_amount, currency,
		alert_threshold_percent, status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		b.ID, b.WorkspaceID, b.CostCenterID, b.Name, b.FiscalYear,
		b.PeriodType, b.AllocatedAmount, b.SpentAmount, b.Currency,
		b.AlertThresholdPercent, b.Status, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *sqliteFinOpsRepository) GetITBudgetByID(ctx context.Context, id string) (*model.ITBudget, error) {
	query := `
	SELECT id, workspace_id, cost_center_id, name, fiscal_year,
	       period_type, allocated_amount, spent_amount, currency,
	       alert_threshold_percent, status, created_at, updated_at
	FROM it_budgets WHERE id = ?`

	var b model.ITBudget
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&b.ID, &b.WorkspaceID, &b.CostCenterID, &b.Name, &b.FiscalYear,
		&b.PeriodType, &b.AllocatedAmount, &b.SpentAmount, &b.Currency,
		&b.AlertThresholdPercent, &b.Status, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBudgetNotFound
		}
		return nil, err
	}
	return &b, nil
}

func (r *sqliteFinOpsRepository) UpdateITBudget(ctx context.Context, b *model.ITBudget) error {
	b.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE it_budgets SET
		cost_center_id = ?, name = ?, fiscal_year = ?,
		period_type = ?, allocated_amount = ?, spent_amount = ?,
		currency = ?, alert_threshold_percent = ?, status = ?,
		updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		b.CostCenterID, b.Name, b.FiscalYear,
		b.PeriodType, b.AllocatedAmount, b.SpentAmount,
		b.Currency, b.AlertThresholdPercent, b.Status,
		b.UpdatedAt, b.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBudgetNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) DeleteITBudget(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM it_budgets WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBudgetNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) ListITBudgets(ctx context.Context, filter model.ITBudgetFilterCriteria) ([]model.ITBudget, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.CostCenterID != "" {
		whereClauses = append(whereClauses, "cost_center_id = ?")
		args = append(args, filter.CostCenterID)
	}
	if filter.FiscalYear > 0 {
		whereClauses = append(whereClauses, "fiscal_year = ?")
		args = append(args, filter.FiscalYear)
	}
	if filter.PeriodType != "" {
		whereClauses = append(whereClauses, "period_type = ?")
		args = append(args, filter.PeriodType)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM it_budgets WHERE %s", whereSQL)
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
	SELECT id, workspace_id, cost_center_id, name, fiscal_year,
	       period_type, allocated_amount, spent_amount, currency,
	       alert_threshold_percent, status, created_at, updated_at
	FROM it_budgets WHERE %s ORDER BY fiscal_year DESC, name ASC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.ITBudget{}
	for rows.Next() {
		var b model.ITBudget
		if err := rows.Scan(
			&b.ID, &b.WorkspaceID, &b.CostCenterID, &b.Name, &b.FiscalYear,
			&b.PeriodType, &b.AllocatedAmount, &b.SpentAmount, &b.Currency,
			&b.AlertThresholdPercent, &b.Status, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, b)
	}
	return list, total, rows.Err()
}

// IT Expenses
func (r *sqliteFinOpsRepository) CreateITExpense(ctx context.Context, e *model.ITExpense) error {
	if e.ID == "" {
		e.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if e.CreatedAt.IsZero() {
		e.CreatedAt = now
	}
	e.UpdatedAt = now
	if e.ExpenseDate.IsZero() {
		e.ExpenseDate = now
	}
	if e.Currency == "" {
		e.Currency = "USD"
	}
	if e.ExpenseType == "" {
		e.ExpenseType = model.ExpenseTypeOpex
	}
	if e.PaymentStatus == "" {
		e.PaymentStatus = model.PaymentStatusPending
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
	INSERT INTO it_expenses (
		id, workspace_id, budget_id, cost_center_id, vendor_id,
		ci_id, invoice_number, description, cost_pool, expense_type,
		amount, currency, expense_date, payment_status, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err = tx.ExecContext(ctx, query,
		e.ID, e.WorkspaceID, e.BudgetID, e.CostCenterID, e.VendorID,
		e.CIID, e.InvoiceNumber, e.Description, e.CostPool, e.ExpenseType,
		e.Amount, e.Currency, e.ExpenseDate, e.PaymentStatus, e.CreatedAt, e.UpdatedAt,
	)
	if err != nil {
		return err
	}

	// Update budget spent_amount
	if e.BudgetID != "" {
		_, err = tx.ExecContext(ctx, `
			UPDATE it_budgets
			SET spent_amount = spent_amount + ?, updated_at = ?
			WHERE id = ?`,
			e.Amount, now, e.BudgetID,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteFinOpsRepository) GetITExpenseByID(ctx context.Context, id string) (*model.ITExpense, error) {
	query := `
	SELECT id, workspace_id, budget_id, cost_center_id, vendor_id,
	       ci_id, invoice_number, description, cost_pool, expense_type,
	       amount, currency, expense_date, payment_status, created_at, updated_at
	FROM it_expenses WHERE id = ?`

	var e model.ITExpense
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&e.ID, &e.WorkspaceID, &e.BudgetID, &e.CostCenterID, &e.VendorID,
		&e.CIID, &e.InvoiceNumber, &e.Description, &e.CostPool, &e.ExpenseType,
		&e.Amount, &e.Currency, &e.ExpenseDate, &e.PaymentStatus, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrExpenseNotFound
		}
		return nil, err
	}
	return &e, nil
}

func (r *sqliteFinOpsRepository) UpdateITExpense(ctx context.Context, e *model.ITExpense) error {
	e.UpdatedAt = time.Now().UTC()

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Fetch old expense to adjust budget spent_amount if amount or budget changed
	var oldAmount float64
	var oldBudgetID string
	err = tx.QueryRowContext(ctx, "SELECT amount, budget_id FROM it_expenses WHERE id = ?", e.ID).Scan(&oldAmount, &oldBudgetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrExpenseNotFound
		}
		return err
	}

	query := `
	UPDATE it_expenses SET
		budget_id = ?, cost_center_id = ?, vendor_id = ?,
		ci_id = ?, invoice_number = ?, description = ?,
		cost_pool = ?, expense_type = ?, amount = ?,
		currency = ?, expense_date = ?, payment_status = ?,
		updated_at = ?
	WHERE id = ?`

	_, err = tx.ExecContext(ctx, query,
		e.BudgetID, e.CostCenterID, e.VendorID,
		e.CIID, e.InvoiceNumber, e.Description,
		e.CostPool, e.ExpenseType, e.Amount,
		e.Currency, e.ExpenseDate, e.PaymentStatus,
		e.UpdatedAt, e.ID,
	)
	if err != nil {
		return err
	}

	// Adjust old budget if budget changed or amount changed
	if oldBudgetID == e.BudgetID {
		diff := e.Amount - oldAmount
		if diff != 0 && e.BudgetID != "" {
			_, err = tx.ExecContext(ctx, "UPDATE it_budgets SET spent_amount = spent_amount + ?, updated_at = ? WHERE id = ?", diff, e.UpdatedAt, e.BudgetID)
			if err != nil {
				return err
			}
		}
	} else {
		if oldBudgetID != "" {
			_, err = tx.ExecContext(ctx, "UPDATE it_budgets SET spent_amount = spent_amount - ?, updated_at = ? WHERE id = ?", oldAmount, e.UpdatedAt, oldBudgetID)
			if err != nil {
				return err
			}
		}
		if e.BudgetID != "" {
			_, err = tx.ExecContext(ctx, "UPDATE it_budgets SET spent_amount = spent_amount + ?, updated_at = ? WHERE id = ?", e.Amount, e.UpdatedAt, e.BudgetID)
			if err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *sqliteFinOpsRepository) DeleteITExpense(ctx context.Context, id string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var amount float64
	var budgetID string
	err = tx.QueryRowContext(ctx, "SELECT amount, budget_id FROM it_expenses WHERE id = ?", id).Scan(&amount, &budgetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrExpenseNotFound
		}
		return err
	}

	res, err := tx.ExecContext(ctx, "DELETE FROM it_expenses WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrExpenseNotFound
	}

	if budgetID != "" {
		_, err = tx.ExecContext(ctx, "UPDATE it_budgets SET spent_amount = spent_amount - ?, updated_at = ? WHERE id = ?", amount, time.Now().UTC(), budgetID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *sqliteFinOpsRepository) ListITExpenses(ctx context.Context, filter model.ITExpenseFilterCriteria) ([]model.ITExpense, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.BudgetID != "" {
		whereClauses = append(whereClauses, "budget_id = ?")
		args = append(args, filter.BudgetID)
	}
	if filter.CostCenterID != "" {
		whereClauses = append(whereClauses, "cost_center_id = ?")
		args = append(args, filter.CostCenterID)
	}
	if filter.VendorID != "" {
		whereClauses = append(whereClauses, "vendor_id = ?")
		args = append(args, filter.VendorID)
	}
	if filter.CIID != "" {
		whereClauses = append(whereClauses, "ci_id = ?")
		args = append(args, filter.CIID)
	}
	if filter.CostPool != "" {
		whereClauses = append(whereClauses, "cost_pool = ?")
		args = append(args, filter.CostPool)
	}
	if filter.ExpenseType != "" {
		whereClauses = append(whereClauses, "expense_type = ?")
		args = append(args, filter.ExpenseType)
	}
	if filter.PaymentStatus != "" {
		whereClauses = append(whereClauses, "payment_status = ?")
		args = append(args, filter.PaymentStatus)
	}
	if filter.FromDate != nil {
		whereClauses = append(whereClauses, "expense_date >= ?")
		args = append(args, *filter.FromDate)
	}
	if filter.ToDate != nil {
		whereClauses = append(whereClauses, "expense_date <= ?")
		args = append(args, *filter.ToDate)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM it_expenses WHERE %s", whereSQL)
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
	SELECT id, workspace_id, budget_id, cost_center_id, vendor_id,
	       ci_id, invoice_number, description, cost_pool, expense_type,
	       amount, currency, expense_date, payment_status, created_at, updated_at
	FROM it_expenses WHERE %s ORDER BY expense_date DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.ITExpense{}
	for rows.Next() {
		var e model.ITExpense
		if err := rows.Scan(
			&e.ID, &e.WorkspaceID, &e.BudgetID, &e.CostCenterID, &e.VendorID,
			&e.CIID, &e.InvoiceNumber, &e.Description, &e.CostPool, &e.ExpenseType,
			&e.Amount, &e.Currency, &e.ExpenseDate, &e.PaymentStatus, &e.CreatedAt, &e.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, e)
	}
	return list, total, rows.Err()
}

// Chargeback Rules
func (r *sqliteFinOpsRepository) CreateChargebackRule(ctx context.Context, rule *model.ChargebackRule) error {
	if rule.ID == "" {
		rule.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = now
	}
	rule.UpdatedAt = now
	if rule.AllocationMethod == "" {
		rule.AllocationMethod = model.AllocMethodFixedPercent
	}

	query := `
	INSERT INTO chargeback_rules (
		id, workspace_id, cost_center_id, ci_id, rule_name,
		target_business_unit, allocation_method, allocation_percentage,
		rate_per_unit, unit_of_measure, is_active, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	isActiveInt := 0
	if rule.IsActive {
		isActiveInt = 1
	}

	_, err := r.db.ExecContext(ctx, query,
		rule.ID, rule.WorkspaceID, rule.CostCenterID, rule.CIID, rule.RuleName,
		rule.TargetBusinessUnit, rule.AllocationMethod, rule.AllocationPercentage,
		rule.RatePerUnit, rule.UnitOfMeasure, isActiveInt, rule.CreatedAt, rule.UpdatedAt,
	)
	return err
}

func (r *sqliteFinOpsRepository) GetChargebackRuleByID(ctx context.Context, id string) (*model.ChargebackRule, error) {
	query := `
	SELECT id, workspace_id, cost_center_id, ci_id, rule_name,
	       target_business_unit, allocation_method, allocation_percentage,
	       rate_per_unit, unit_of_measure, is_active, created_at, updated_at
	FROM chargeback_rules WHERE id = ?`

	var rule model.ChargebackRule
	var isActiveInt int
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rule.ID, &rule.WorkspaceID, &rule.CostCenterID, &rule.CIID, &rule.RuleName,
		&rule.TargetBusinessUnit, &rule.AllocationMethod, &rule.AllocationPercentage,
		&rule.RatePerUnit, &rule.UnitOfMeasure, &isActiveInt, &rule.CreatedAt, &rule.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrChargebackRuleNotFound
		}
		return nil, err
	}
	rule.IsActive = (isActiveInt == 1)
	return &rule, nil
}

func (r *sqliteFinOpsRepository) UpdateChargebackRule(ctx context.Context, rule *model.ChargebackRule) error {
	rule.UpdatedAt = time.Now().UTC()
	isActiveInt := 0
	if rule.IsActive {
		isActiveInt = 1
	}

	query := `
	UPDATE chargeback_rules SET
		cost_center_id = ?, ci_id = ?, rule_name = ?,
		target_business_unit = ?, allocation_method = ?, allocation_percentage = ?,
		rate_per_unit = ?, unit_of_measure = ?, is_active = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		rule.CostCenterID, rule.CIID, rule.RuleName,
		rule.TargetBusinessUnit, rule.AllocationMethod, rule.AllocationPercentage,
		rule.RatePerUnit, rule.UnitOfMeasure, isActiveInt, rule.UpdatedAt,
		rule.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrChargebackRuleNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) DeleteChargebackRule(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM chargeback_rules WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrChargebackRuleNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) ListChargebackRules(ctx context.Context, filter model.ChargebackRuleFilterCriteria) ([]model.ChargebackRule, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.CostCenterID != "" {
		whereClauses = append(whereClauses, "cost_center_id = ?")
		args = append(args, filter.CostCenterID)
	}
	if filter.TargetBusinessUnit != "" {
		whereClauses = append(whereClauses, "target_business_unit = ?")
		args = append(args, filter.TargetBusinessUnit)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM chargeback_rules WHERE %s", whereSQL)
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
	SELECT id, workspace_id, cost_center_id, ci_id, rule_name,
	       target_business_unit, allocation_method, allocation_percentage,
	       rate_per_unit, unit_of_measure, is_active, created_at, updated_at
	FROM chargeback_rules WHERE %s ORDER BY rule_name ASC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.ChargebackRule{}
	for rows.Next() {
		var rule model.ChargebackRule
		var isActiveInt int
		if err := rows.Scan(
			&rule.ID, &rule.WorkspaceID, &rule.CostCenterID, &rule.CIID, &rule.RuleName,
			&rule.TargetBusinessUnit, &rule.AllocationMethod, &rule.AllocationPercentage,
			&rule.RatePerUnit, &rule.UnitOfMeasure, &isActiveInt, &rule.CreatedAt, &rule.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		rule.IsActive = (isActiveInt == 1)
		list = append(list, rule)
	}
	return list, total, rows.Err()
}

// Chargeback Statements
func (r *sqliteFinOpsRepository) CreateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	now := time.Now().UTC()
	if s.CreatedAt.IsZero() {
		s.CreatedAt = now
	}
	s.UpdatedAt = now
	if s.GeneratedAt.IsZero() {
		s.GeneratedAt = now
	}
	if s.Currency == "" {
		s.Currency = "USD"
	}
	if s.Status == "" {
		s.Status = model.StatementStatusDraft
	}

	query := `
	INSERT INTO chargeback_statements (
		id, workspace_id, target_business_unit, statement_period,
		total_charge, currency, status, notes, generated_at,
		settled_at, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.WorkspaceID, s.TargetBusinessUnit, s.StatementPeriod,
		s.TotalCharge, s.Currency, s.Status, s.Notes, s.GeneratedAt,
		s.SettledAt, s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *sqliteFinOpsRepository) GetChargebackStatementByID(ctx context.Context, id string) (*model.ChargebackStatement, error) {
	query := `
	SELECT id, workspace_id, target_business_unit, statement_period,
	       total_charge, currency, status, notes, generated_at,
	       settled_at, created_at, updated_at
	FROM chargeback_statements WHERE id = ?`

	var s model.ChargebackStatement
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.WorkspaceID, &s.TargetBusinessUnit, &s.StatementPeriod,
		&s.TotalCharge, &s.Currency, &s.Status, &s.Notes, &s.GeneratedAt,
		&s.SettledAt, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrStatementNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *sqliteFinOpsRepository) UpdateChargebackStatement(ctx context.Context, s *model.ChargebackStatement) error {
	s.UpdatedAt = time.Now().UTC()

	query := `
	UPDATE chargeback_statements SET
		target_business_unit = ?, statement_period = ?, total_charge = ?,
		currency = ?, status = ?, notes = ?, settled_at = ?, updated_at = ?
	WHERE id = ?`

	res, err := r.db.ExecContext(ctx, query,
		s.TargetBusinessUnit, s.StatementPeriod, s.TotalCharge,
		s.Currency, s.Status, s.Notes, s.SettledAt, s.UpdatedAt,
		s.ID,
	)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrStatementNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) DeleteChargebackStatement(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM chargeback_statements WHERE id = ?", id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrStatementNotFound
	}
	return nil
}

func (r *sqliteFinOpsRepository) ListChargebackStatements(ctx context.Context, filter model.ChargebackStatementFilterCriteria) ([]model.ChargebackStatement, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}

	if filter.WorkspaceID != "" {
		whereClauses = append(whereClauses, "workspace_id = ?")
		args = append(args, filter.WorkspaceID)
	}
	if filter.TargetBusinessUnit != "" {
		whereClauses = append(whereClauses, "target_business_unit = ?")
		args = append(args, filter.TargetBusinessUnit)
	}
	if filter.StatementPeriod != "" {
		whereClauses = append(whereClauses, "statement_period = ?")
		args = append(args, filter.StatementPeriod)
	}
	if filter.Status != "" {
		whereClauses = append(whereClauses, "status = ?")
		args = append(args, filter.Status)
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM chargeback_statements WHERE %s", whereSQL)
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
	SELECT id, workspace_id, target_business_unit, statement_period,
	       total_charge, currency, status, notes, generated_at,
	       settled_at, created_at, updated_at
	FROM chargeback_statements WHERE %s ORDER BY generated_at DESC LIMIT ? OFFSET ?`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []model.ChargebackStatement{}
	for rows.Next() {
		var s model.ChargebackStatement
		if err := rows.Scan(
			&s.ID, &s.WorkspaceID, &s.TargetBusinessUnit, &s.StatementPeriod,
			&s.TotalCharge, &s.Currency, &s.Status, &s.Notes, &s.GeneratedAt,
			&s.SettledAt, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}
	return list, total, rows.Err()
}

// FinOps Analytics
func (r *sqliteFinOpsRepository) GetFinOpsAnalytics(ctx context.Context, workspaceID string, fiscalYear int) (*model.FinOpsAnalyticsReport, error) {
	if fiscalYear <= 0 {
		fiscalYear = time.Now().UTC().Year()
	}

	report := &model.FinOpsAnalyticsReport{
		WorkspaceID:          workspaceID,
		FiscalYear:           fiscalYear,
		SpendByCostPool:      make(map[string]float64),
		SpendByExpenseType:   make(map[string]float64),
		SpendByDepartment:    make(map[string]float64),
		TopCostDriversCI:     make(map[string]float64),
		TopCostDriversVendor: make(map[string]float64),
	}

	// 1. Total Cost Centers
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM cost_centers WHERE workspace_id = ?", workspaceID).Scan(&report.TotalCostCenters)
	if err != nil {
		return nil, err
	}

	// 2. Budget Allocation and Spend
	budgetQuery := `
	SELECT COALESCE(SUM(allocated_amount), 0.0), COALESCE(SUM(spent_amount), 0.0)
	FROM it_budgets
	WHERE workspace_id = ? AND fiscal_year = ?`
	err = r.db.QueryRowContext(ctx, budgetQuery, workspaceID, fiscalYear).Scan(&report.TotalBudgetAllocated, &report.TotalActualSpend)
	if err != nil {
		return nil, err
	}

	report.BudgetVariance = report.TotalBudgetAllocated - report.TotalActualSpend
	if report.TotalBudgetAllocated > 0 {
		report.BudgetUtilizationPercent = math.Round((report.TotalActualSpend/report.TotalBudgetAllocated)*10000) / 100
	}

	yearStr := fmt.Sprintf("%04d", fiscalYear)

	// 3. Spend by Cost Pool
	poolRows, err := r.db.QueryContext(ctx, `
		SELECT cost_pool, COALESCE(SUM(amount), 0.0)
		FROM it_expenses
		WHERE workspace_id = ? AND strftime('%Y', expense_date) = ?
		GROUP BY cost_pool`,
		workspaceID, yearStr,
	)
	if err == nil {
		defer poolRows.Close()
		for poolRows.Next() {
			var pool string
			var amt float64
			if err := poolRows.Scan(&pool, &amt); err == nil {
				report.SpendByCostPool[pool] = amt
			}
		}
	}

	// 4. Spend by Expense Type
	typeRows, err := r.db.QueryContext(ctx, `
		SELECT expense_type, COALESCE(SUM(amount), 0.0)
		FROM it_expenses
		WHERE workspace_id = ? AND strftime('%Y', expense_date) = ?
		GROUP BY expense_type`,
		workspaceID, yearStr,
	)
	if err == nil {
		defer typeRows.Close()
		for typeRows.Next() {
			var etype string
			var amt float64
			if err := typeRows.Scan(&etype, &amt); err == nil {
				report.SpendByExpenseType[etype] = amt
			}
		}
	}

	// 5. Spend by Department
	deptRows, err := r.db.QueryContext(ctx, `
		SELECT cc.department, COALESCE(SUM(e.amount), 0.0)
		FROM it_expenses e
		JOIN cost_centers cc ON e.cost_center_id = cc.id
		WHERE e.workspace_id = ? AND strftime('%Y', e.expense_date) = ?
		GROUP BY cc.department`,
		workspaceID, yearStr,
	)
	if err == nil {
		defer deptRows.Close()
		for deptRows.Next() {
			var dept string
			var amt float64
			if err := deptRows.Scan(&dept, &amt); err == nil {
				report.SpendByDepartment[dept] = amt
			}
		}
	}

	// 6. Top Cost Drivers CI
	ciRows, err := r.db.QueryContext(ctx, `
		SELECT ci_id, COALESCE(SUM(amount), 0.0)
		FROM it_expenses
		WHERE workspace_id = ? AND strftime('%Y', expense_date) = ? AND ci_id != ''
		GROUP BY ci_id
		ORDER BY SUM(amount) DESC
		LIMIT 5`,
		workspaceID, yearStr,
	)
	if err == nil {
		defer ciRows.Close()
		for ciRows.Next() {
			var ci string
			var amt float64
			if err := ciRows.Scan(&ci, &amt); err == nil {
				report.TopCostDriversCI[ci] = amt
			}
		}
	}

	// 7. Top Cost Drivers Vendor
	vendorRows, err := r.db.QueryContext(ctx, `
		SELECT vendor_id, COALESCE(SUM(amount), 0.0)
		FROM it_expenses
		WHERE workspace_id = ? AND strftime('%Y', expense_date) = ? AND vendor_id != ''
		GROUP BY vendor_id
		ORDER BY SUM(amount) DESC
		LIMIT 5`,
		workspaceID, yearStr,
	)
	if err == nil {
		defer vendorRows.Close()
		for vendorRows.Next() {
			var vendor string
			var amt float64
			if err := vendorRows.Scan(&vendor, &amt); err == nil {
				report.TopCostDriversVendor[vendor] = amt
			}
		}
	}

	// 8. Chargeback Statements
	chargebackQuery := `
	SELECT
		COALESCE(SUM(total_charge), 0.0),
		COALESCE(SUM(CASE WHEN status = 'settled' THEN total_charge ELSE 0.0 END), 0.0)
	FROM chargeback_statements
	WHERE workspace_id = ? AND strftime('%Y', generated_at) = ?`
	_ = r.db.QueryRowContext(ctx, chargebackQuery, workspaceID, yearStr).Scan(
		&report.TotalChargebackIssued,
		&report.TotalChargebackSettled,
	)

	// 9. Cost Efficiency Score (0-100)
	score := 100.0
	if report.TotalBudgetAllocated > 0 {
		if report.TotalActualSpend > report.TotalBudgetAllocated {
			overrunPct := ((report.TotalActualSpend - report.TotalBudgetAllocated) / report.TotalBudgetAllocated) * 100.0
			score -= overrunPct
		}
	}
	if report.TotalChargebackIssued > 0 {
		unsettledRatio := (report.TotalChargebackIssued - report.TotalChargebackSettled) / report.TotalChargebackIssued
		score -= unsettledRatio * 15.0
	}
	if score < 0.0 {
		score = 0.0
	}
	if score > 100.0 {
		score = 100.0
	}
	report.CostEfficiencyScore = math.Round(score*10) / 10

	return report, nil
}
