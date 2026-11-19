package service

import (
	"context"
	"database/sql"
	"sync"
	"testing"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/db"
	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
)

func setupTestFinOpsSvc(t *testing.T) (FinOpsService, repository.FinOpsRepository, *sql.DB) {
	sqliteDB, err := db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to initialize test in-memory db: %v", err)
	}
	sqliteDB.SetMaxOpenConns(1)

	_, err = sqliteDB.Exec(`
		INSERT INTO users (id, username, email, password_hash, created_at, updated_at) 
		VALUES ('user_1', 'alice', 'alice@example.com', 'hash', datetime('now'), datetime('now')),
		       ('user_2', 'bob', 'bob@example.com', 'hash', datetime('now'), datetime('now'));

		INSERT INTO workspaces (id, name, slug, owner_id, created_at, updated_at) 
		VALUES (1, 'FinOps WS', 'finops-ws', 'user_1', datetime('now'), datetime('now'));
	`)
	if err != nil {
		t.Fatalf("failed to insert test seed data: %v", err)
	}

	repo := repository.NewFinOpsRepository(sqliteDB)
	svc := NewFinOpsService(repo)
	return svc, repo, sqliteDB
}

func TestFinOps_CostCenterLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// 1. Validation failure: missing code
	_, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Name:       "Core Engineering",
		Department: model.DeptEngineering,
	})
	if err != ErrCostCenterCodeRequired {
		t.Fatalf("expected ErrCostCenterCodeRequired, got %v", err)
	}

	// 2. Validation failure: missing name
	_, err = svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-ENG-001",
		Department: model.DeptEngineering,
	})
	if err != ErrCostCenterNameRequired {
		t.Fatalf("expected ErrCostCenterNameRequired, got %v", err)
	}

	// 3. Validation failure: invalid department
	_, err = svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-ENG-001",
		Name:       "Core Engineering",
		Department: "invalid_dept",
	})
	if err != ErrInvalidDepartment {
		t.Fatalf("expected ErrInvalidDepartment, got %v", err)
	}

	// 4. Successful creation
	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:        "CC-ENG-101",
		Name:        "Platform Engineering",
		Department:  model.DeptEngineering,
		ManagerID:   "user_1",
		Description: "Platform infrastructure and cloud engineering cost center",
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}
	if cc.ID == "" || cc.Code != "CC-ENG-101" || cc.Status != model.CostCenterStatusActive {
		t.Fatalf("unexpected cost center fields: %+v", cc)
	}

	// 5. Unique code constraint failure
	_, err = svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-ENG-101",
		Name:       "Duplicate Code Cost Center",
		Department: model.DeptInfrastructure,
	})
	if err != ErrCostCenterCodeConflict {
		t.Fatalf("expected ErrCostCenterCodeConflict, got %v", err)
	}

	// 6. GetByID & GetByCode
	fetchedByID, err := svc.GetCostCenterByID(ctx, cc.ID)
	if err != nil {
		t.Fatalf("failed to get cost center by id: %v", err)
	}
	if fetchedByID.Name != "Platform Engineering" {
		t.Fatalf("expected name Platform Engineering, got %s", fetchedByID.Name)
	}

	fetchedByCode, err := svc.GetCostCenterByCode(ctx, "ws_1", "CC-ENG-101")
	if err != nil {
		t.Fatalf("failed to get cost center by code: %v", err)
	}
	if fetchedByCode.ID != cc.ID {
		t.Fatalf("mismatched id from get by code")
	}

	// 7. Update cost center
	newName := "Cloud Platform Engineering"
	newDesc := "Updated cost center description"
	updated, err := svc.UpdateCostCenter(ctx, cc.ID, model.UpdateCostCenterRequest{
		Name:        &newName,
		Description: &newDesc,
	})
	if err != nil {
		t.Fatalf("failed to update cost center: %v", err)
	}
	if updated.Name != newName || updated.Description != newDesc {
		t.Fatalf("updated cost center fields mismatch: %+v", updated)
	}

	// 8. List cost centers
	list, total, err := svc.ListCostCenters(ctx, model.CostCenterFilterCriteria{
		WorkspaceID: "ws_1",
		Department:  model.DeptEngineering,
	})
	if err != nil {
		t.Fatalf("failed to list cost centers: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 cost center, got total=%d len=%d", total, len(list))
	}

	// 9. Delete cost center
	if err := svc.DeleteCostCenter(ctx, cc.ID); err != nil {
		t.Fatalf("failed to delete cost center: %v", err)
	}
	_, err = svc.GetCostCenterByID(ctx, cc.ID)
	if err != ErrCostCenterNotFound {
		t.Fatalf("expected ErrCostCenterNotFound, got %v", err)
	}
}

func TestFinOps_ITBudgetLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Create Cost Center
	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-SEC-201",
		Name:       "Cybersecurity Ops",
		Department: model.DeptSecurity,
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}

	// 1. Validation failure: missing name
	_, err = svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:    cc.ID,
		FiscalYear:      2026,
		PeriodType:      model.BudgetPeriodAnnual,
		AllocatedAmount: 150000,
	})
	if err != ErrBudgetNameRequired {
		t.Fatalf("expected ErrBudgetNameRequired, got %v", err)
	}

	// 2. Validation failure: invalid fiscal year
	_, err = svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:    cc.ID,
		Name:            "FY1990 Security Budget",
		FiscalYear:      1990,
		PeriodType:      model.BudgetPeriodAnnual,
		AllocatedAmount: 150000,
	})
	if err != ErrInvalidFiscalYear {
		t.Fatalf("expected ErrInvalidFiscalYear, got %v", err)
	}

	// 3. Validation failure: non-existent cost center
	_, err = svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:    "non_existent_cc",
		Name:            "FY2026 Sec Budget",
		FiscalYear:      2026,
		PeriodType:      model.BudgetPeriodAnnual,
		AllocatedAmount: 150000,
	})
	if err != ErrCostCenterNotFound {
		t.Fatalf("expected ErrCostCenterNotFound, got %v", err)
	}

	// 4. Successful creation
	budget, err := svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:          cc.ID,
		Name:                  "FY2026 Cybersecurity Defense",
		FiscalYear:            2026,
		PeriodType:            model.BudgetPeriodAnnual,
		AllocatedAmount:       250000.0,
		Currency:              "USD",
		AlertThresholdPercent: 85.0,
		Status:                model.BudgetStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create IT budget: %v", err)
	}
	if budget.ID == "" || budget.AllocatedAmount != 250000.0 || budget.AlertThresholdPercent != 85.0 {
		t.Fatalf("unexpected budget fields: %+v", budget)
	}

	// 5. GetByID
	fetched, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to get budget by id: %v", err)
	}
	if fetched.Name != "FY2026 Cybersecurity Defense" {
		t.Fatalf("budget name mismatch: %s", fetched.Name)
	}

	// 6. Update budget
	newAlloc := 300000.0
	newThreshold := 90.0
	updated, err := svc.UpdateITBudget(ctx, budget.ID, model.UpdateITBudgetRequest{
		AllocatedAmount:       &newAlloc,
		AlertThresholdPercent: &newThreshold,
	})
	if err != nil {
		t.Fatalf("failed to update IT budget: %v", err)
	}
	if updated.AllocatedAmount != 300000.0 || updated.AlertThresholdPercent != 90.0 {
		t.Fatalf("updated budget mismatch: %+v", updated)
	}

	// 7. List budgets
	list, total, err := svc.ListITBudgets(ctx, model.ITBudgetFilterCriteria{
		WorkspaceID:  "ws_1",
		CostCenterID: cc.ID,
		FiscalYear:   2026,
	})
	if err != nil {
		t.Fatalf("failed to list budgets: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("expected 1 budget, got total=%d len=%d", total, len(list))
	}

	// 8. Delete budget
	if err := svc.DeleteITBudget(ctx, budget.ID); err != nil {
		t.Fatalf("failed to delete budget: %v", err)
	}
	_, err = svc.GetITBudgetByID(ctx, budget.ID)
	if err != ErrBudgetNotFound {
		t.Fatalf("expected ErrBudgetNotFound, got %v", err)
	}
}

func TestFinOps_ITExpenseLifecycleAndThresholds(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Cost Center & Budget Setup
	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-INF-301",
		Name:       "Core Infrastructure",
		Department: model.DeptInfrastructure,
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}

	budget, err := svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:          cc.ID,
		Name:                  "FY2026 Cloud Spend",
		FiscalYear:            2026,
		PeriodType:            model.BudgetPeriodAnnual,
		AllocatedAmount:       50000.0,
		AlertThresholdPercent: 80.0,
		Status:                model.BudgetStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	// 1. Validation failure: missing description
	now := time.Now().UTC()
	_, err = svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:     budget.ID,
		CostCenterID: cc.ID,
		CostPool:     model.CostPoolCloudInfrastructure,
		ExpenseType:  model.ExpenseTypeOpex,
		Amount:       12000.0,
		ExpenseDate:  now,
	})
	if err == nil {
		t.Fatalf("expected validation error for empty description")
	}

	// 2. Validation failure: invalid cost pool
	_, err = svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:     budget.ID,
		CostCenterID: cc.ID,
		Description:  "AWS EKS Cluster Nodes",
		CostPool:     "invalid_pool",
		ExpenseType:  model.ExpenseTypeOpex,
		Amount:       12000.0,
		ExpenseDate:  now,
	})
	if err != ErrInvalidCostPool {
		t.Fatalf("expected ErrInvalidCostPool, got %v", err)
	}

	// 3. Create First Expense ($20,000)
	expense1, err := svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:      budget.ID,
		CostCenterID:  cc.ID,
		VendorID:      "vendor_aws",
		CIID:          "ci_k8s_prod",
		InvoiceNumber: "INV-2026-001",
		Description:   "AWS Production Compute & Network",
		CostPool:      model.CostPoolCloudInfrastructure,
		ExpenseType:   model.ExpenseTypeOpex,
		Amount:        20000.0,
		ExpenseDate:   now,
		PaymentStatus: model.PaymentStatusPaid,
	})
	if err != nil {
		t.Fatalf("failed to create expense1: %v", err)
	}
	if expense1.ID == "" || expense1.Amount != 20000.0 {
		t.Fatalf("unexpected expense fields: %+v", expense1)
	}

	// Check Budget spent amount updated to $20,000
	updatedBudget, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to fetch budget: %v", err)
	}
	if updatedBudget.SpentAmount != 20000.0 {
		t.Fatalf("expected spent amount 20000.0, got %f", updatedBudget.SpentAmount)
	}

	// 4. Create Second Expense ($35,000) -> Exceeding allocated $50,000
	expense2, err := svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:      budget.ID,
		CostCenterID:  cc.ID,
		VendorID:      "vendor_gcp",
		CIID:          "ci_bigquery",
		InvoiceNumber: "INV-2026-002",
		Description:   "GCP Analytics Pipeline",
		CostPool:      model.CostPoolCloudInfrastructure,
		ExpenseType:   model.ExpenseTypeOpex,
		Amount:        35000.0,
		ExpenseDate:   now,
		PaymentStatus: model.PaymentStatusApproved,
	})
	if err != nil {
		t.Fatalf("failed to create expense2: %v", err)
	}

	// Check Budget status transitioned to exceeded ($55,000 spent > $50,000 allocated)
	updatedBudget2, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to fetch budget: %v", err)
	}
	if updatedBudget2.SpentAmount != 55000.0 {
		t.Fatalf("expected spent amount 55000.0, got %f", updatedBudget2.SpentAmount)
	}
	if updatedBudget2.Status != model.BudgetStatusExceeded {
		t.Fatalf("expected budget status exceeded, got %s", updatedBudget2.Status)
	}

	// 5. Update Expense amount ($35,000 -> $25,000)
	reducedAmt := 25000.0
	updatedExp2, err := svc.UpdateITExpense(ctx, expense2.ID, model.UpdateITExpenseRequest{
		Amount: &reducedAmt,
	})
	if err != nil {
		t.Fatalf("failed to update expense: %v", err)
	}
	if updatedExp2.Amount != 25000.0 {
		t.Fatalf("expected amount 25000.0, got %f", updatedExp2.Amount)
	}

	// Budget spent should now be 20000 + 25000 = 45000.0
	updatedBudget3, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to fetch budget: %v", err)
	}
	if updatedBudget3.SpentAmount != 45000.0 {
		t.Fatalf("expected spent amount 45000.0, got %f", updatedBudget3.SpentAmount)
	}

	// 6. List Expenses
	expenses, total, err := svc.ListITExpenses(ctx, model.ITExpenseFilterCriteria{
		WorkspaceID:  "ws_1",
		CostCenterID: cc.ID,
		CostPool:     model.CostPoolCloudInfrastructure,
	})
	if err != nil {
		t.Fatalf("failed to list expenses: %v", err)
	}
	if total != 2 || len(expenses) != 2 {
		t.Fatalf("expected 2 expenses, got total=%d len=%d", total, len(expenses))
	}

	// 7. Delete Expense1 ($20,000)
	if err := svc.DeleteITExpense(ctx, expense1.ID); err != nil {
		t.Fatalf("failed to delete expense1: %v", err)
	}
	updatedBudget4, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to fetch budget: %v", err)
	}
	if updatedBudget4.SpentAmount != 25000.0 {
		t.Fatalf("expected spent amount 25000.0 after deletion, got %f", updatedBudget4.SpentAmount)
	}
}

func TestFinOps_ChargebackLifecycle(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	// Cost Center
	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-OPS-401",
		Name:       "Operations Center",
		Department: model.DeptOperations,
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}

	// 1. Chargeback Rule Validation
	_, err = svc.CreateChargebackRule(ctx, "ws_1", model.CreateChargebackRuleRequest{
		CostCenterID:       cc.ID,
		TargetBusinessUnit: "Product Management",
		AllocationMethod:   model.AllocMethodFixedPercent,
	})
	if err == nil {
		t.Fatalf("expected error for missing rule_name")
	}

	// 2. Successful Chargeback Rule Creation
	rule, err := svc.CreateChargebackRule(ctx, "ws_1", model.CreateChargebackRuleRequest{
		CostCenterID:         cc.ID,
		RuleName:             "Shared Database Cluster Allocation",
		TargetBusinessUnit:   "Digital Commerce",
		AllocationMethod:     model.AllocMethodFixedPercent,
		AllocationPercentage: 35.0,
	})
	if err != nil {
		t.Fatalf("failed to create chargeback rule: %v", err)
	}
	if rule.ID == "" || rule.AllocationPercentage != 35.0 || !rule.IsActive {
		t.Fatalf("unexpected chargeback rule fields: %+v", rule)
	}

	// 3. Update Rule
	newPct := 40.0
	updatedRule, err := svc.UpdateChargebackRule(ctx, rule.ID, model.UpdateChargebackRuleRequest{
		AllocationPercentage: &newPct,
	})
	if err != nil {
		t.Fatalf("failed to update chargeback rule: %v", err)
	}
	if updatedRule.AllocationPercentage != 40.0 {
		t.Fatalf("expected percentage 40.0, got %f", updatedRule.AllocationPercentage)
	}

	// 4. List Rules
	rules, totalRules, err := svc.ListChargebackRules(ctx, model.ChargebackRuleFilterCriteria{
		WorkspaceID:        "ws_1",
		TargetBusinessUnit: "Digital Commerce",
	})
	if err != nil {
		t.Fatalf("failed to list chargeback rules: %v", err)
	}
	if totalRules != 1 || len(rules) != 1 {
		t.Fatalf("expected 1 rule, got total=%d len=%d", totalRules, len(rules))
	}

	// 5. Chargeback Statement Creation
	stmt, err := svc.CreateChargebackStatement(ctx, "ws_1", model.CreateChargebackStatementRequest{
		TargetBusinessUnit: "Digital Commerce",
		StatementPeriod:    "2026-Q1",
		TotalCharge:        18500.0,
		Currency:           "USD",
		Notes:              "Q1 allocated shared cloud services charge",
	})
	if err != nil {
		t.Fatalf("failed to create chargeback statement: %v", err)
	}
	if stmt.ID == "" || stmt.Status != model.StatementStatusDraft || stmt.TotalCharge != 18500.0 {
		t.Fatalf("unexpected statement fields: %+v", stmt)
	}

	// 6. Settle Statement
	settledStatus := model.StatementStatusSettled
	updatedStmt, err := svc.UpdateChargebackStatement(ctx, stmt.ID, model.UpdateChargebackStatementRequest{
		Status: &settledStatus,
	})
	if err != nil {
		t.Fatalf("failed to settle chargeback statement: %v", err)
	}
	if updatedStmt.Status != model.StatementStatusSettled || updatedStmt.SettledAt == nil {
		t.Fatalf("statement not settled properly: %+v", updatedStmt)
	}

	// 7. List Statements
	stmts, totalStmts, err := svc.ListChargebackStatements(ctx, model.ChargebackStatementFilterCriteria{
		WorkspaceID:        "ws_1",
		TargetBusinessUnit: "Digital Commerce",
	})
	if err != nil {
		t.Fatalf("failed to list statements: %v", err)
	}
	if totalStmts != 1 || len(stmts) != 1 {
		t.Fatalf("expected 1 statement, got total=%d len=%d", totalStmts, len(stmts))
	}
}

func TestFinOps_AnalyticsReport(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()
	currentYear := time.Now().UTC().Year()

	// 1. Setup Cost Center
	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-ENG-501",
		Name:       "Engineering Shared Services",
		Department: model.DeptEngineering,
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}

	// 2. Setup Budget ($200,000)
	budget, err := svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:          cc.ID,
		Name:                  "FY Global Cloud",
		FiscalYear:            currentYear,
		PeriodType:            model.BudgetPeriodAnnual,
		AllocatedAmount:       200000.0,
		AlertThresholdPercent: 80.0,
		Status:                model.BudgetStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	// 3. Log Expenses across pools and vendors
	now := time.Now().UTC()
	_, err = svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:      budget.ID,
		CostCenterID:  cc.ID,
		VendorID:      "vendor_aws",
		CIID:          "ci_prod_db",
		Description:   "Aurora Database Cluster",
		CostPool:      model.CostPoolCloudInfrastructure,
		ExpenseType:   model.ExpenseTypeOpex,
		Amount:        45000.0,
		ExpenseDate:   now,
		PaymentStatus: model.PaymentStatusPaid,
	})
	if err != nil {
		t.Fatalf("failed to log expense 1: %v", err)
	}

	_, err = svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
		BudgetID:      budget.ID,
		CostCenterID:  cc.ID,
		VendorID:      "vendor_datadog",
		CIID:          "ci_observability",
		Description:   "APM and log analytics licenses",
		CostPool:      model.CostPoolSoftwareLicenses,
		ExpenseType:   model.ExpenseTypeOpex,
		Amount:        15000.0,
		ExpenseDate:   now,
		PaymentStatus: model.PaymentStatusPaid,
	})
	if err != nil {
		t.Fatalf("failed to log expense 2: %v", err)
	}

	// 4. Chargeback Statement
	stmt, err := svc.CreateChargebackStatement(ctx, "ws_1", model.CreateChargebackStatementRequest{
		TargetBusinessUnit: "Platform Engineering",
		StatementPeriod:    "2026-Q1",
		TotalCharge:        20000.0,
	})
	if err != nil {
		t.Fatalf("failed to create statement: %v", err)
	}
	settled := model.StatementStatusSettled
	_, _ = svc.UpdateChargebackStatement(ctx, stmt.ID, model.UpdateChargebackStatementRequest{Status: &settled})

	// 5. Get FinOps Analytics Report
	report, err := svc.GetFinOpsAnalytics(ctx, "ws_1", currentYear)
	if err != nil {
		t.Fatalf("failed to get finops analytics: %v", err)
	}

	if report.TotalCostCenters != 1 {
		t.Fatalf("expected 1 cost center, got %d", report.TotalCostCenters)
	}
	if report.TotalBudgetAllocated != 200000.0 {
		t.Fatalf("expected 200000.0 allocated, got %f", report.TotalBudgetAllocated)
	}
	if report.TotalActualSpend != 60000.0 {
		t.Fatalf("expected 60000.0 actual spend, got %f", report.TotalActualSpend)
	}
	if report.BudgetVariance != 140000.0 {
		t.Fatalf("expected 140000.0 variance, got %f", report.BudgetVariance)
	}
	if report.BudgetUtilizationPercent != 30.0 {
		t.Fatalf("expected 30.0%% utilization, got %f", report.BudgetUtilizationPercent)
	}
	if report.SpendByCostPool[model.CostPoolCloudInfrastructure] != 45000.0 {
		t.Fatalf("expected 45000.0 in cloud, got %f", report.SpendByCostPool[model.CostPoolCloudInfrastructure])
	}
	if report.SpendByCostPool[model.CostPoolSoftwareLicenses] != 15000.0 {
		t.Fatalf("expected 15000.0 in licenses, got %f", report.SpendByCostPool[model.CostPoolSoftwareLicenses])
	}
	if report.SpendByDepartment[model.DeptEngineering] != 60000.0 {
		t.Fatalf("expected 60000.0 in engineering dept, got %f", report.SpendByDepartment[model.DeptEngineering])
	}
	if report.TotalChargebackSettled != 20000.0 {
		t.Fatalf("expected 20000.0 settled chargeback, got %f", report.TotalChargebackSettled)
	}
	if report.CostEfficiencyScore <= 0.0 || report.CostEfficiencyScore > 100.0 {
		t.Fatalf("expected score between 0 and 100, got %f", report.CostEfficiencyScore)
	}
}

func TestFinOps_ConcurrencyAndRace(t *testing.T) {
	svc, _, dbConn := setupTestFinOpsSvc(t)
	defer dbConn.Close()

	ctx := context.Background()

	cc, err := svc.CreateCostCenter(ctx, "ws_1", model.CreateCostCenterRequest{
		Code:       "CC-CONCUR-01",
		Name:       "Concurrent Cost Center",
		Department: model.DeptEngineering,
	})
	if err != nil {
		t.Fatalf("failed to create cost center: %v", err)
	}

	budget, err := svc.CreateITBudget(ctx, "ws_1", model.CreateITBudgetRequest{
		CostCenterID:    cc.ID,
		Name:            "Concurrent Budget",
		FiscalYear:      time.Now().UTC().Year(),
		PeriodType:      model.BudgetPeriodAnnual,
		AllocatedAmount: 1000000.0,
		Status:          model.BudgetStatusActive,
	})
	if err != nil {
		t.Fatalf("failed to create budget: %v", err)
	}

	var wg sync.WaitGroup
	numWorkers := 10

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			_, _ = svc.CreateITExpense(ctx, "ws_1", model.CreateITExpenseRequest{
				BudgetID:      budget.ID,
				CostCenterID:  cc.ID,
				VendorID:      "vendor_cloud",
				Description:   "Concurrent Worker Expense",
				CostPool:      model.CostPoolCloudInfrastructure,
				ExpenseType:   model.ExpenseTypeOpex,
				Amount:        100.0,
				ExpenseDate:   time.Now().UTC(),
				PaymentStatus: model.PaymentStatusPending,
			})
			_, _ = svc.GetFinOpsAnalytics(ctx, "ws_1", time.Now().UTC().Year())
		}(i)
	}

	wg.Wait()

	finalBudget, err := svc.GetITBudgetByID(ctx, budget.ID)
	if err != nil {
		t.Fatalf("failed to fetch final budget: %v", err)
	}
	expectedSpent := float64(numWorkers * 100)
	if finalBudget.SpentAmount != expectedSpent {
		t.Fatalf("expected final spent %f, got %f", expectedSpent, finalBudget.SpentAmount)
	}
}
