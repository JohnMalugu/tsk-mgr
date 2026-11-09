package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// ChangeHandler exposes HTTP endpoints for ITIL Change Management & CAB Approval workflows
type ChangeHandler struct {
	svc service.ChangeService
}

var GlobalChangeHandler *ChangeHandler

// InitChangeHandler initializes the singleton change handler
func InitChangeHandler(svc service.ChangeService) {
	GlobalChangeHandler = &ChangeHandler{svc: svc}
}

// NewChangeHandler creates a new ChangeHandler instance
func NewChangeHandler(svc service.ChangeService) *ChangeHandler {
	return &ChangeHandler{svc: svc}
}

func getChangeUser(r *http.Request) string {
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

func writeChangeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeChangeError(w http.ResponseWriter, status int, msg string) {
	writeChangeJSON(w, status, map[string]string{"error": msg})
}

func handleChangeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrChangeNotFound):
		writeChangeError(w, http.StatusNotFound, "change request not found")
	case errors.Is(err, service.ErrApprovalNotFound):
		writeChangeError(w, http.StatusNotFound, "change approval not found")
	case errors.Is(err, service.ErrPIRNotFound):
		writeChangeError(w, http.StatusNotFound, "change post-implementation review not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeChangeError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeChangeError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrChangeTitleRequired):
		writeChangeError(w, http.StatusBadRequest, "change title is required")
	case errors.Is(err, service.ErrInvalidChangeType):
		writeChangeError(w, http.StatusBadRequest, "invalid change type: must be standard, normal, or emergency")
	case errors.Is(err, service.ErrInvalidChangeRiskLevel):
		writeChangeError(w, http.StatusBadRequest, "invalid risk level: must be low, medium, high, or critical")
	case errors.Is(err, service.ErrInvalidChangeStatus):
		writeChangeError(w, http.StatusBadRequest, "invalid change status")
	case errors.Is(err, service.ErrInvalidChangeTransition):
		writeChangeError(w, http.StatusBadRequest, "invalid change status transition")
	case errors.Is(err, service.ErrChangeNotEditable):
		writeChangeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, service.ErrReviewerRequired):
		writeChangeError(w, http.StatusBadRequest, "reviewer_id is required")
	case errors.Is(err, service.ErrInvalidCABDecision):
		writeChangeError(w, http.StatusBadRequest, "invalid CAB decision: must be approved, rejected, or abstained")
	case errors.Is(err, service.ErrInvalidPIROutcome):
		writeChangeError(w, http.StatusBadRequest, "invalid PIR outcome")
	case errors.Is(err, service.ErrPIRNotesRequired):
		writeChangeError(w, http.StatusBadRequest, "review_notes are required")
	case errors.Is(err, service.ErrTaskAlreadyLinked):
		writeChangeError(w, http.StatusConflict, "task is already linked to this change request")
	case errors.Is(err, service.ErrInvalidChangePhase):
		writeChangeError(w, http.StatusBadRequest, "invalid phase: must be pre_implementation, implementation, or post_implementation")
	default:
		writeChangeError(w, http.StatusBadRequest, err.Error())
	}
}

// CreateChange handles POST /v2/changes
func (h *ChangeHandler) CreateChange(w http.ResponseWriter, r *http.Request) {
	var req model.CreateChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeChangeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getChangeUser(r)
	change, err := h.svc.CreateChange(userID, req)
	if err != nil {
		handleChangeServiceError(w, err)
		return
	}
	writeChangeJSON(w, http.StatusCreated, change)
}

// GetChange handles GET /v2/changes/{id}
func (h *ChangeHandler) GetChange(w http.ResponseWriter, r *http.Request, id int64) {
	change, err := h.svc.GetChange(id)
	if err != nil {
		handleChangeServiceError(w, err)
		return
	}
	writeChangeJSON(w, http.StatusOK, change)
}

// ListChanges handles GET /v2/changes
func (h *ChangeHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var filter model.ChangeFilter

	if wsIDStr := q.Get("workspace_id"); wsIDStr != "" {
		if wsID, err := strconv.ParseInt(wsIDStr, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if projIDStr := q.Get("project_id"); projIDStr != "" {
		if projID, err := strconv.ParseInt(projIDStr, 10, 64); err == nil {
			filter.ProjectID = &projID
		}
	}
	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}
	if ct := q.Get("change_type"); ct != "" {
		filter.ChangeType = &ct
	}
	if rl := q.Get("risk_level"); rl != "" {
		filter.RiskLevel = &rl
	}
	if rb := q.Get("requested_by"); rb != "" {
		filter.RequestedBy = &rb
	}
	if at := q.Get("assigned_to"); at != "" {
		filter.AssignedTo = &at
	}
	if search := q.Get("search"); search != "" {
		filter.Search = &search
	}
	if sd := q.Get("start_date"); sd != "" {
		filter.StartDate = &sd
	}
	if ed := q.Get("end_date"); ed != "" {
		filter.EndDate = &ed
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = l
		}
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = o
		}
	}

	changes, err := h.svc.ListChanges(filter)
	if err != nil {
		handleChangeServiceError(w, err)
		return
	}
	if changes == nil {
		changes = []*model.ChangeRequest{}
	}
	writeChangeJSON(w, http.StatusOK, changes)
}

// UpdateChange handles PUT /v2/changes/{id}
func (h *ChangeHandler) UpdateChange(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateChangeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeChangeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getChangeUser(r)
	change, err := h.svc.UpdateChange(id, userID, req)
	if err != nil {
		handleChangeServiceError(w, err)
		return
	}
	writeChangeJSON(w, http.StatusOK, change)
}

// DeleteChange handles DELETE /v2/changes/{id}
func (h *ChangeHandler) DeleteChange(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getChangeUser(r)
	if err := h.svc.DeleteChange(id, userID); err != nil {
		handleChangeServiceError(w, err)
		return
	}
	writeChangeJSON(w, http.StatusOK, map[string]string{"message": "change request deleted successfully"})
}

// ChangeStatus handles POST /v2/changes/{id}/status
func (h *ChangeHandler) ChangeStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.ChangeStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeChangeError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getChangeUser(r)
	change, err := h.svc.ChangeStatus(id, userID, req)
	if err != nil {
		handleChangeServiceError(w, err)
		return
	}
	writeChangeJSON(w, http.StatusOK, change)
}
