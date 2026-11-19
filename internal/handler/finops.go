package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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
