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

var (
	_ = strconv.Atoi
	_ = strings.TrimPrefix
)

// CSIHandler exposes HTTP endpoints for Continual Service Improvement
type CSIHandler struct {
	svc service.CSIService
}

// GlobalCSIHandler is the singleton handler instance
var GlobalCSIHandler *CSIHandler

// InitCSIHandler initializes the global CSI handler singleton
func InitCSIHandler(svc service.CSIService) {
	GlobalCSIHandler = &CSIHandler{svc: svc}
}

// NewCSIHandler creates a new CSIHandler instance
func NewCSIHandler(svc service.CSIService) *CSIHandler {
	return &CSIHandler{svc: svc}
}

func getCSIUser(r *http.Request) string {
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

func getCSIWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeCSIJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeCSIError(w http.ResponseWriter, status int, msg string) {
	writeCSIJSON(w, status, map[string]string{"error": msg})
}

func handleCSIError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCSIInitiativeNotFound):
		writeCSIError(w, http.StatusNotFound, "continual improvement initiative not found")
	case errors.Is(err, service.ErrCSIInitiativeTitleRequired):
		writeCSIError(w, http.StatusBadRequest, "initiative title is required")
	case errors.Is(err, service.ErrCSIInitiativeCodeConflict):
		writeCSIError(w, http.StatusConflict, "initiative with this code already exists in workspace")
	case errors.Is(err, model.ErrInvalidCSIAlignment):
		writeCSIError(w, http.StatusBadRequest, "invalid strategic alignment category")
	case errors.Is(err, model.ErrInvalidCSIStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative status")
	case errors.Is(err, model.ErrInvalidCSIPriority):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative priority")
	case errors.Is(err, model.ErrInvalidCSISource):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI initiative source")
	case errors.Is(err, service.ErrCSIBenchmarkNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI benchmark metric not found")
	case errors.Is(err, service.ErrCSIBenchmarkMetricNameRequired):
		writeCSIError(w, http.StatusBadRequest, "benchmark metric name is required")
	case errors.Is(err, model.ErrInvalidCSIMetricUnit):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI benchmark metric unit")
	case errors.Is(err, service.ErrCSIBenefitNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI benefit realization record not found")
	case errors.Is(err, model.ErrInvalidCSIBenefitStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI benefit realization status")
	case errors.Is(err, model.ErrInvalidEffortHours):
		writeCSIError(w, http.StatusBadRequest, "effort hours must be non-negative")
	case errors.Is(err, service.ErrCSIActionNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI action item not found")
	case errors.Is(err, model.ErrCSIActionDescriptionRequired):
		writeCSIError(w, http.StatusBadRequest, "action description is required")
	case errors.Is(err, model.ErrInvalidCSIActionStatus):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI action status")
	case errors.Is(err, service.ErrCSIReviewNotFound):
		writeCSIError(w, http.StatusNotFound, "CSI review record not found")
	case errors.Is(err, model.ErrCSIReviewNotesRequired):
		writeCSIError(w, http.StatusBadRequest, "review notes or findings are required")
	case errors.Is(err, model.ErrInvalidCSIReviewOutcome):
		writeCSIError(w, http.StatusBadRequest, "invalid CSI review outcome")
	default:
		writeCSIError(w, http.StatusInternalServerError, err.Error())
	}
}

// Initiative Endpoints

// ListInitiatives handles GET /v2/csi/initiatives
func (h *CSIHandler) ListInitiatives(w http.ResponseWriter, r *http.Request) {
	wsID := getCSIWorkspaceID(r)
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

	filter := model.CSIFilterCriteria{
		WorkspaceID:        wsID,
		StrategicAlignment: q.Get("strategic_alignment"),
		Priority:           q.Get("priority"),
		Status:             q.Get("status"),
		Source:             q.Get("source"),
		CIID:               q.Get("ci_id"),
		Limit:              limit,
		Offset:             offset,
	}

	initiatives, total, err := h.svc.ListInitiatives(r.Context(), filter)
	if err != nil {
		handleCSIError(w, err)
		return
	}

	writeCSIJSON(w, http.StatusOK, map[string]interface{}{
		"initiatives": initiatives,
		"total":       total,
		"limit":       limit,
		"offset":      offset,
	})
}

// CreateInitiative handles POST /v2/csi/initiatives
func (h *CSIHandler) CreateInitiative(w http.ResponseWriter, r *http.Request) {
	wsID := getCSIWorkspaceID(r)

	var req model.CreateCSIInitiativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	initiative, err := h.svc.CreateInitiative(r.Context(), wsID, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}

	writeCSIJSON(w, http.StatusCreated, initiative)
}

// GetInitiativeByID handles GET /v2/csi/initiatives/:id
func (h *CSIHandler) GetInitiativeByID(w http.ResponseWriter, r *http.Request, id string) {
	initiative, err := h.svc.GetInitiativeByID(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// UpdateInitiative handles PUT /v2/csi/initiatives/:id
func (h *CSIHandler) UpdateInitiative(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCSIInitiativeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCSIError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	initiative, err := h.svc.UpdateInitiative(r.Context(), id, req)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// CompleteInitiative handles POST /v2/csi/initiatives/:id/complete
func (h *CSIHandler) CompleteInitiative(w http.ResponseWriter, r *http.Request, id string) {
	initiative, err := h.svc.CompleteInitiative(r.Context(), id)
	if err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, initiative)
}

// DeleteInitiative handles DELETE /v2/csi/initiatives/:id
func (h *CSIHandler) DeleteInitiative(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteInitiative(r.Context(), id); err != nil {
		handleCSIError(w, err)
		return
	}
	writeCSIJSON(w, http.StatusOK, map[string]string{"message": "continual improvement initiative deleted successfully"})
}
