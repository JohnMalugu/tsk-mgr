package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// ContinuityHandler exposes HTTP endpoints for Service Continuity and Disaster Recovery
type ContinuityHandler struct {
	svc service.ContinuityService
}

// GlobalContinuityHandler is the singleton handler instance
var GlobalContinuityHandler *ContinuityHandler

// InitContinuityHandler initializes the global continuity handler singleton
func InitContinuityHandler(svc service.ContinuityService) {
	GlobalContinuityHandler = &ContinuityHandler{svc: svc}
}

// NewContinuityHandler creates a new ContinuityHandler instance
func NewContinuityHandler(svc service.ContinuityService) *ContinuityHandler {
	return &ContinuityHandler{svc: svc}
}

func getContinuityUser(r *http.Request) string {
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

func getContinuityWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeContinuityJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeContinuityError(w http.ResponseWriter, status int, msg string) {
	writeContinuityJSON(w, status, map[string]string{"error": msg})
}

func handleContinuityError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrDRPlanNotFound):
		writeContinuityError(w, http.StatusNotFound, "disaster recovery plan not found")
	case errors.Is(err, service.ErrDRPlanCodeRequired):
		writeContinuityError(w, http.StatusBadRequest, "plan_code is required")
	case errors.Is(err, service.ErrDRPlanCodeConflict):
		writeContinuityError(w, http.StatusConflict, "disaster recovery plan with this code already exists in workspace")
	case errors.Is(err, service.ErrDRPlanNameRequired):
		writeContinuityError(w, http.StatusBadRequest, "plan name is required")
	case errors.Is(err, service.ErrInvalidRecoveryTier):
		writeContinuityError(w, http.StatusBadRequest, "invalid recovery tier")
	case errors.Is(err, service.ErrInvalidDRPlanStatus):
		writeContinuityError(w, http.StatusBadRequest, "invalid disaster recovery plan status")
	case errors.Is(err, service.ErrRecoveryObjectiveNotFound):
		writeContinuityError(w, http.StatusNotFound, "recovery objective not found")
	case errors.Is(err, service.ErrServiceNameRequired):
		writeContinuityError(w, http.StatusBadRequest, "service name is required")
	case errors.Is(err, service.ErrInvalidRTOValue):
		writeContinuityError(w, http.StatusBadRequest, "rto_value_minutes must be non-negative")
	case errors.Is(err, service.ErrInvalidRPOValue):
		writeContinuityError(w, http.StatusBadRequest, "rpo_value_minutes must be non-negative")
	case errors.Is(err, service.ErrInvalidAvailability):
		writeContinuityError(w, http.StatusBadRequest, "target availability must be between 0.0 and 100.0 percent")
	case errors.Is(err, service.ErrInvalidDataLossTolerance):
		writeContinuityError(w, http.StatusBadRequest, "invalid data loss tolerance")
	case errors.Is(err, service.ErrBIANotFound):
		writeContinuityError(w, http.StatusNotFound, "business impact analysis not found")
	case errors.Is(err, service.ErrInvalidMTDValue):
		writeContinuityError(w, http.StatusBadRequest, "mtd_minutes must be greater than zero")
	case errors.Is(err, service.ErrInvalidOperationalImpact):
		writeContinuityError(w, http.StatusBadRequest, "invalid operational impact")
	case errors.Is(err, service.ErrInvalidRegulatoryImpact):
		writeContinuityError(w, http.StatusBadRequest, "invalid regulatory impact")
	case errors.Is(err, service.ErrInvalidCriticalWindow):
		writeContinuityError(w, http.StatusBadRequest, "invalid critical recovery window")
	case errors.Is(err, service.ErrRunbookNotFound):
		writeContinuityError(w, http.StatusNotFound, "recovery runbook not found")
	case errors.Is(err, service.ErrRunbookStepNotFound):
		writeContinuityError(w, http.StatusNotFound, "runbook step not found")
	case errors.Is(err, service.ErrRunbookTitleRequired):
		writeContinuityError(w, http.StatusBadRequest, "runbook title is required")
	case errors.Is(err, service.ErrStepNameRequired):
		writeContinuityError(w, http.StatusBadRequest, "step name is required")
	case errors.Is(err, service.ErrInvalidExecutionMode):
		writeContinuityError(w, http.StatusBadRequest, "invalid runbook execution mode")
	case errors.Is(err, service.ErrInvalidRunbookActionType):
		writeContinuityError(w, http.StatusBadRequest, "invalid runbook step action type")
	case errors.Is(err, service.ErrInvalidExecutionRole):
		writeContinuityError(w, http.StatusBadRequest, "invalid runbook step execution role")
	case errors.Is(err, service.ErrDrillNotFound):
		writeContinuityError(w, http.StatusNotFound, "continuity drill not found")
	case errors.Is(err, service.ErrDrillFindingNotFound):
		writeContinuityError(w, http.StatusNotFound, "continuity drill finding not found")
	case errors.Is(err, service.ErrInvalidDrillType):
		writeContinuityError(w, http.StatusBadRequest, "invalid continuity drill type")
	case errors.Is(err, service.ErrInvalidDrillStatus):
		writeContinuityError(w, http.StatusBadRequest, "invalid continuity drill status")
	case errors.Is(err, service.ErrDrillScenarioRequired):
		writeContinuityError(w, http.StatusBadRequest, "drill scenario is required")
	case errors.Is(err, service.ErrFindingDescRequired):
		writeContinuityError(w, http.StatusBadRequest, "drill finding description is required")
	case errors.Is(err, service.ErrInvalidFindingSeverity):
		writeContinuityError(w, http.StatusBadRequest, "invalid drill finding severity")
	case errors.Is(err, service.ErrInvalidFindingCategory):
		writeContinuityError(w, http.StatusBadRequest, "invalid drill finding category")
	case errors.Is(err, service.ErrInvalidDrillFindingStatus):
		writeContinuityError(w, http.StatusBadRequest, "invalid drill finding status")
	default:
		writeContinuityError(w, http.StatusInternalServerError, err.Error())
	}
}

// CreateDRPlan handles POST /v2/dr-plans
func (h *ContinuityHandler) CreateDRPlan(w http.ResponseWriter, r *http.Request) {
	wsID := getContinuityWorkspaceID(r)
	var req model.CreateDRPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	plan, err := h.svc.CreateDRPlan(r.Context(), wsID, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusCreated, plan)
}

// GetDRPlanByID handles GET /v2/dr-plans/{id}
func (h *ContinuityHandler) GetDRPlanByID(w http.ResponseWriter, r *http.Request, id string) {
	plan, err := h.svc.GetDRPlanByID(r.Context(), id)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, plan)
}

// UpdateDRPlan handles PUT /v2/dr-plans/{id}
func (h *ContinuityHandler) UpdateDRPlan(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateDRPlanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	plan, err := h.svc.UpdateDRPlan(r.Context(), id, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, plan)
}

// DeleteDRPlan handles DELETE /v2/dr-plans/{id}
func (h *ContinuityHandler) DeleteDRPlan(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteDRPlan(r.Context(), id); err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]string{"message": "disaster recovery plan deleted"})
}

// ListDRPlans handles GET /v2/dr-plans
func (h *ContinuityHandler) ListDRPlans(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.DRPlanFilterCriteria{
		WorkspaceID:  getContinuityWorkspaceID(r),
		RecoveryTier: q.Get("recovery_tier"),
		Status:       q.Get("status"),
		Search:       q.Get("search"),
		Limit:        limit,
		Offset:       offset,
	}

	plans, total, err := h.svc.ListDRPlans(r.Context(), filter)
	if err != nil {
		handleContinuityError(w, err)
		return
	}

	writeContinuityJSON(w, http.StatusOK, map[string]interface{}{
		"plans":  plans,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// ApproveDRPlan handles POST /v2/dr-plans/{id}/approve
func (h *ContinuityHandler) ApproveDRPlan(w http.ResponseWriter, r *http.Request, id string) {
	approverID := getContinuityUser(r)
	plan, err := h.svc.ApproveDRPlan(r.Context(), id, approverID)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, plan)
}

// ActivateDRPlan handles POST /v2/dr-plans/{id}/activate
func (h *ContinuityHandler) ActivateDRPlan(w http.ResponseWriter, r *http.Request, id string) {
	plan, err := h.svc.ActivateDRPlan(r.Context(), id)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, plan)
}

// CreateRecoveryObjective handles POST /v2/continuity/objectives
func (h *ContinuityHandler) CreateRecoveryObjective(w http.ResponseWriter, r *http.Request) {
	wsID := getContinuityWorkspaceID(r)
	var req model.CreateRecoveryObjectiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	obj, err := h.svc.CreateRecoveryObjective(r.Context(), wsID, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusCreated, obj)
}

// GetRecoveryObjectiveByID handles GET /v2/continuity/objectives/{id}
func (h *ContinuityHandler) GetRecoveryObjectiveByID(w http.ResponseWriter, r *http.Request, id string) {
	obj, err := h.svc.GetRecoveryObjectiveByID(r.Context(), id)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, obj)
}

// UpdateRecoveryObjective handles PUT /v2/continuity/objectives/{id}
func (h *ContinuityHandler) UpdateRecoveryObjective(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateRecoveryObjectiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	obj, err := h.svc.UpdateRecoveryObjective(r.Context(), id, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, obj)
}

// DeleteRecoveryObjective handles DELETE /v2/continuity/objectives/{id}
func (h *ContinuityHandler) DeleteRecoveryObjective(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteRecoveryObjective(r.Context(), id); err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]string{"message": "recovery objective deleted"})
}

// ListRecoveryObjectives handles GET /v2/continuity/objectives
func (h *ContinuityHandler) ListRecoveryObjectives(w http.ResponseWriter, r *http.Request) {
	wsID := getContinuityWorkspaceID(r)
	objs, err := h.svc.ListRecoveryObjectives(r.Context(), wsID)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]interface{}{
		"objectives": objs,
		"total":      len(objs),
	})
}

// ListRecoveryObjectivesByPlan handles GET /v2/dr-plans/{id}/objectives
func (h *ContinuityHandler) ListRecoveryObjectivesByPlan(w http.ResponseWriter, r *http.Request, planID string) {
	objs, err := h.svc.ListRecoveryObjectivesByPlan(r.Context(), planID)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]interface{}{
		"objectives": objs,
		"total":      len(objs),
	})
}

// CreateBIA handles POST /v2/continuity/bias
func (h *ContinuityHandler) CreateBIA(w http.ResponseWriter, r *http.Request) {
	wsID := getContinuityWorkspaceID(r)
	var req model.CreateBIARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	bia, err := h.svc.CreateBIA(r.Context(), wsID, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusCreated, bia)
}

// GetBIAByID handles GET /v2/continuity/bias/{id}
func (h *ContinuityHandler) GetBIAByID(w http.ResponseWriter, r *http.Request, id string) {
	bia, err := h.svc.GetBIAByID(r.Context(), id)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, bia)
}

// UpdateBIA handles PUT /v2/continuity/bias/{id}
func (h *ContinuityHandler) UpdateBIA(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateBIARequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeContinuityError(w, http.StatusBadRequest, "invalid json payload")
		return
	}

	bia, err := h.svc.UpdateBIA(r.Context(), id, req)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, bia)
}

// DeleteBIA handles DELETE /v2/continuity/bias/{id}
func (h *ContinuityHandler) DeleteBIA(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteBIA(r.Context(), id); err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]string{"message": "business impact analysis deleted"})
}

// ListBIAs handles GET /v2/continuity/bias
func (h *ContinuityHandler) ListBIAs(w http.ResponseWriter, r *http.Request) {
	wsID := getContinuityWorkspaceID(r)
	bias, err := h.svc.ListBIAs(r.Context(), wsID)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]interface{}{
		"bias":  bias,
		"total": len(bias),
	})
}

// ListBIAsByPlan handles GET /v2/dr-plans/{id}/bias
func (h *ContinuityHandler) ListBIAsByPlan(w http.ResponseWriter, r *http.Request, planID string) {
	bias, err := h.svc.ListBIAsByPlan(r.Context(), planID)
	if err != nil {
		handleContinuityError(w, err)
		return
	}
	writeContinuityJSON(w, http.StatusOK, map[string]interface{}{
		"bias":  bias,
		"total": len(bias),
	})
}
