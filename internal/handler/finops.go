package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// FinOpsHandler exposes HTTP endpoints for IT Financial Management and FinOps
type FinOpsHandler struct {
	svc service.FinOpsService
}

// GlobalFinOpsHandler is the singleton handler instance
var GlobalFinOpsHandler *FinOpsHandler

// InitFinOpsHandler initializes the global finops handler singleton
func InitFinOpsHandler(svc service.FinOpsService) {
	GlobalFinOpsHandler = &FinOpsHandler{svc: svc}
}

// NewFinOpsHandler creates a new FinOpsHandler instance
func NewFinOpsHandler(svc service.FinOpsService) *FinOpsHandler {
	return &FinOpsHandler{svc: svc}
}

func getFinOpsUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "user_1"
}

func getFinOpsWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeFinOpsJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeFinOpsError(w http.ResponseWriter, status int, msg string) {
	writeFinOpsJSON(w, status, map[string]string{"error": msg})
}

func handleFinOpsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCostCenterNotFound):
		writeFinOpsError(w, http.StatusNotFound, "cost center not found")
	case errors.Is(err, service.ErrCostCenterCodeRequired):
		writeFinOpsError(w, http.StatusBadRequest, "cost center code is required")
	case errors.Is(err, service.ErrCostCenterCodeConflict):
		writeFinOpsError(w, http.StatusConflict, "cost center with this code already exists in workspace")
	case errors.Is(err, service.ErrCostCenterNameRequired):
		writeFinOpsError(w, http.StatusBadRequest, "cost center name is required")
	case errors.Is(err, service.ErrInvalidCostCenterStatus):
		writeFinOpsError(w, http.StatusBadRequest, "invalid cost center status")
	case errors.Is(err, service.ErrInvalidDepartment):
		writeFinOpsError(w, http.StatusBadRequest, "invalid department")
	case errors.Is(err, service.ErrBudgetNotFound):
		writeFinOpsError(w, http.StatusNotFound, "budget not found")
	case errors.Is(err, service.ErrBudgetNameRequired):
		writeFinOpsError(w, http.StatusBadRequest, "budget name is required")
	case errors.Is(err, service.ErrInvalidFiscalYear):
		writeFinOpsError(w, http.StatusBadRequest, "invalid fiscal year")
	case errors.Is(err, service.ErrInvalidBudgetPeriod):
		writeFinOpsError(w, http.StatusBadRequest, "invalid budget period type")
	case errors.Is(err, service.ErrInvalidBudgetStatus):
		writeFinOpsError(w, http.StatusBadRequest, "invalid budget status")
	case errors.Is(err, service.ErrInvalidAllocatedAmount):
		writeFinOpsError(w, http.StatusBadRequest, "invalid allocated amount")
	case errors.Is(err, service.ErrInvalidAlertThreshold):
		writeFinOpsError(w, http.StatusBadRequest, "invalid alert threshold percent")
	case errors.Is(err, service.ErrExpenseNotFound):
		writeFinOpsError(w, http.StatusNotFound, "expense not found")
	case errors.Is(err, service.ErrInvalidCostPool):
		writeFinOpsError(w, http.StatusBadRequest, "invalid cost pool")
	case errors.Is(err, service.ErrInvalidExpenseType):
		writeFinOpsError(w, http.StatusBadRequest, "invalid expense type")
	case errors.Is(err, service.ErrInvalidExpenseAmount):
		writeFinOpsError(w, http.StatusBadRequest, "invalid expense amount")
	case errors.Is(err, service.ErrInvalidPaymentStatus):
		writeFinOpsError(w, http.StatusBadRequest, "invalid payment status")
	case errors.Is(err, service.ErrExpenseDateRequired):
		writeFinOpsError(w, http.StatusBadRequest, "expense date is required")
	case errors.Is(err, service.ErrChargebackRuleNotFound):
		writeFinOpsError(w, http.StatusNotFound, "chargeback rule not found")
	case errors.Is(err, service.ErrInvalidAllocationMethod):
		writeFinOpsError(w, http.StatusBadRequest, "invalid allocation method")
	case errors.Is(err, service.ErrInvalidAllocationValue):
		writeFinOpsError(w, http.StatusBadRequest, "invalid allocation percentage or value")
	case errors.Is(err, service.ErrBusinessUnitRequired):
		writeFinOpsError(w, http.StatusBadRequest, "target business unit is required")
	case errors.Is(err, service.ErrStatementNotFound):
		writeFinOpsError(w, http.StatusNotFound, "chargeback statement not found")
	case errors.Is(err, service.ErrInvalidStatementStatus):
		writeFinOpsError(w, http.StatusBadRequest, "invalid statement status")
	default:
		writeFinOpsError(w, http.StatusInternalServerError, err.Error())
	}
}

// Cost Center HTTP Handlers

func (h *FinOpsHandler) ListCostCenters(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := model.CostCenterFilterCriteria{
		WorkspaceID: wsID,
		Department:  q.Get("department"),
		Status:      q.Get("status"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	list, total, err := h.svc.ListCostCenters(r.Context(), filter)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]interface{}{
		"cost_centers": list,
		"total":        total,
		"limit":        limit,
		"offset":       offset,
	})
}

func (h *FinOpsHandler) CreateCostCenter(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)

	var req model.CreateCostCenterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cc, err := h.svc.CreateCostCenter(r.Context(), wsID, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusCreated, cc)
}

func (h *FinOpsHandler) GetCostCenterByID(w http.ResponseWriter, r *http.Request, id string) {
	cc, err := h.svc.GetCostCenterByID(r.Context(), id)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, cc)
}

func (h *FinOpsHandler) UpdateCostCenter(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCostCenterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	cc, err := h.svc.UpdateCostCenter(r.Context(), id, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, cc)
}

func (h *FinOpsHandler) DeleteCostCenter(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteCostCenter(r.Context(), id); err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// IT Budget HTTP Handlers

func (h *FinOpsHandler) ListITBudgets(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	var fiscalYear int
	if fy := q.Get("fiscal_year"); fy != "" {
		if val, err := strconv.Atoi(fy); err == nil {
			fiscalYear = val
		}
	}

	filter := model.ITBudgetFilterCriteria{
		WorkspaceID:  wsID,
		CostCenterID: q.Get("cost_center_id"),
		FiscalYear:   fiscalYear,
		PeriodType:   q.Get("period_type"),
		Status:       q.Get("status"),
		Limit:        limit,
		Offset:       offset,
	}

	list, total, err := h.svc.ListITBudgets(r.Context(), filter)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]interface{}{
		"budgets": list,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func (h *FinOpsHandler) CreateITBudget(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)

	var req model.CreateITBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	budget, err := h.svc.CreateITBudget(r.Context(), wsID, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusCreated, budget)
}

func (h *FinOpsHandler) GetITBudgetByID(w http.ResponseWriter, r *http.Request, id string) {
	budget, err := h.svc.GetITBudgetByID(r.Context(), id)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, budget)
}

func (h *FinOpsHandler) UpdateITBudget(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateITBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	budget, err := h.svc.UpdateITBudget(r.Context(), id, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, budget)
}

func (h *FinOpsHandler) DeleteITBudget(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteITBudget(r.Context(), id); err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// IT Expense HTTP Handlers

func (h *FinOpsHandler) ListITExpenses(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := model.ITExpenseFilterCriteria{
		WorkspaceID:   wsID,
		BudgetID:      q.Get("budget_id"),
		CostCenterID:  q.Get("cost_center_id"),
		VendorID:      q.Get("vendor_id"),
		CIID:          q.Get("ci_id"),
		CostPool:      q.Get("cost_pool"),
		ExpenseType:   q.Get("expense_type"),
		PaymentStatus: q.Get("payment_status"),
		Limit:         limit,
		Offset:        offset,
	}

	if fd := q.Get("from_date"); fd != "" {
		if t, err := time.Parse(time.RFC3339, fd); err == nil {
			filter.FromDate = &t
		} else if t, err := time.Parse("2006-01-02", fd); err == nil {
			filter.FromDate = &t
		}
	}
	if td := q.Get("to_date"); td != "" {
		if t, err := time.Parse(time.RFC3339, td); err == nil {
			filter.ToDate = &t
		} else if t, err := time.Parse("2006-01-02", td); err == nil {
			filter.ToDate = &t
		}
	}

	list, total, err := h.svc.ListITExpenses(r.Context(), filter)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]interface{}{
		"expenses": list,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

func (h *FinOpsHandler) CreateITExpense(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)

	var req model.CreateITExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense, err := h.svc.CreateITExpense(r.Context(), wsID, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusCreated, expense)
}

func (h *FinOpsHandler) GetITExpenseByID(w http.ResponseWriter, r *http.Request, id string) {
	expense, err := h.svc.GetITExpenseByID(r.Context(), id)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, expense)
}

func (h *FinOpsHandler) UpdateITExpense(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateITExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	expense, err := h.svc.UpdateITExpense(r.Context(), id, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, expense)
}

func (h *FinOpsHandler) DeleteITExpense(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteITExpense(r.Context(), id); err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Chargeback Rule HTTP Handlers

func (h *FinOpsHandler) ListChargebackRules(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := model.ChargebackRuleFilterCriteria{
		WorkspaceID:        wsID,
		CostCenterID:       q.Get("cost_center_id"),
		TargetBusinessUnit: q.Get("target_business_unit"),
		Limit:              limit,
		Offset:             offset,
	}

	list, total, err := h.svc.ListChargebackRules(r.Context(), filter)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]interface{}{
		"rules":  list,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *FinOpsHandler) CreateChargebackRule(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)

	var req model.CreateChargebackRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.CreateChargebackRule(r.Context(), wsID, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusCreated, rule)
}

func (h *FinOpsHandler) GetChargebackRuleByID(w http.ResponseWriter, r *http.Request, id string) {
	rule, err := h.svc.GetChargebackRuleByID(r.Context(), id)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, rule)
}

func (h *FinOpsHandler) UpdateChargebackRule(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateChargebackRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.UpdateChargebackRule(r.Context(), id, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, rule)
}

func (h *FinOpsHandler) DeleteChargebackRule(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteChargebackRule(r.Context(), id); err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// Chargeback Statement HTTP Handlers

func (h *FinOpsHandler) ListChargebackStatements(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	limit := 50
	offset := 0
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	filter := model.ChargebackStatementFilterCriteria{
		WorkspaceID:        wsID,
		TargetBusinessUnit: q.Get("target_business_unit"),
		StatementPeriod:    q.Get("statement_period"),
		Status:             q.Get("status"),
		Limit:              limit,
		Offset:             offset,
	}

	list, total, err := h.svc.ListChargebackStatements(r.Context(), filter)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]interface{}{
		"statements": list,
		"total":      total,
		"limit":      limit,
		"offset":     offset,
	})
}

func (h *FinOpsHandler) CreateChargebackStatement(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)

	var req model.CreateChargebackStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	stmt, err := h.svc.CreateChargebackStatement(r.Context(), wsID, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusCreated, stmt)
}

func (h *FinOpsHandler) GetChargebackStatementByID(w http.ResponseWriter, r *http.Request, id string) {
	stmt, err := h.svc.GetChargebackStatementByID(r.Context(), id)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, stmt)
}

func (h *FinOpsHandler) UpdateChargebackStatement(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateChargebackStatementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeFinOpsError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	stmt, err := h.svc.UpdateChargebackStatement(r.Context(), id, req)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, stmt)
}

func (h *FinOpsHandler) DeleteChargebackStatement(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteChargebackStatement(r.Context(), id); err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// FinOps Analytics HTTP Handler

func (h *FinOpsHandler) GetFinOpsAnalytics(w http.ResponseWriter, r *http.Request) {
	wsID := getFinOpsWorkspaceID(r)
	q := r.URL.Query()

	var fiscalYear int
	if fy := q.Get("fiscal_year"); fy != "" {
		if val, err := strconv.Atoi(fy); err == nil {
			fiscalYear = val
		}
	}

	report, err := h.svc.GetFinOpsAnalytics(r.Context(), wsID, fiscalYear)
	if err != nil {
		handleFinOpsError(w, err)
		return
	}

	writeFinOpsJSON(w, http.StatusOK, report)
}

// ServeHTTP dispatches requests for /v2/finops
func (h *FinOpsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/finops")
	path = strings.TrimPrefix(path, "/")

	if path == "analytics" {
		if r.Method == http.MethodGet {
			h.GetFinOpsAnalytics(w, r)
		} else {
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "chargeback/rules") {
		sub := strings.TrimPrefix(path, "chargeback/rules")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListChargebackRules(w, r)
			case http.MethodPost:
				h.CreateChargebackRule(w, r)
			default:
				writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		ruleID := sub
		switch r.Method {
		case http.MethodGet:
			h.GetChargebackRuleByID(w, r, ruleID)
		case http.MethodPut:
			h.UpdateChargebackRule(w, r, ruleID)
		case http.MethodDelete:
			h.DeleteChargebackRule(w, r, ruleID)
		default:
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "chargeback/statements") {
		sub := strings.TrimPrefix(path, "chargeback/statements")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListChargebackStatements(w, r)
			case http.MethodPost:
				h.CreateChargebackStatement(w, r)
			default:
				writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		stmtID := sub
		switch r.Method {
		case http.MethodGet:
			h.GetChargebackStatementByID(w, r, stmtID)
		case http.MethodPut:
			h.UpdateChargebackStatement(w, r, stmtID)
		case http.MethodDelete:
			h.DeleteChargebackStatement(w, r, stmtID)
		default:
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "budgets") {
		sub := strings.TrimPrefix(path, "budgets")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListITBudgets(w, r)
			case http.MethodPost:
				h.CreateITBudget(w, r)
			default:
				writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		budgetID := sub
		switch r.Method {
		case http.MethodGet:
			h.GetITBudgetByID(w, r, budgetID)
		case http.MethodPut:
			h.UpdateITBudget(w, r, budgetID)
		case http.MethodDelete:
			h.DeleteITBudget(w, r, budgetID)
		default:
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "expenses") {
		sub := strings.TrimPrefix(path, "expenses")
		sub = strings.TrimPrefix(sub, "/")
		if sub == "" {
			switch r.Method {
			case http.MethodGet:
				h.ListITExpenses(w, r)
			case http.MethodPost:
				h.CreateITExpense(w, r)
			default:
				writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		expenseID := sub
		switch r.Method {
		case http.MethodGet:
			h.GetITExpenseByID(w, r, expenseID)
		case http.MethodPut:
			h.UpdateITExpense(w, r, expenseID)
		case http.MethodDelete:
			h.DeleteITExpense(w, r, expenseID)
		default:
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeFinOpsError(w, http.StatusNotFound, "route not found")
}

// ServeCostCentersHTTP dispatches requests for /v2/cost-centers
func (h *FinOpsHandler) ServeCostCentersHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/cost-centers")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListCostCenters(w, r)
		case http.MethodPost:
			h.CreateCostCenter(w, r)
		default:
			writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	// /v2/cost-centers/{id}
	ccID := path
	switch r.Method {
	case http.MethodGet:
		h.GetCostCenterByID(w, r, ccID)
	case http.MethodPut:
		h.UpdateCostCenter(w, r, ccID)
	case http.MethodDelete:
		h.DeleteCostCenter(w, r, ccID)
	default:
		writeFinOpsError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
