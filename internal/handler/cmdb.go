package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// CMDBHandler exposes HTTP endpoints for ITIL Configuration Management Database (CMDB) & Service Assets
type CMDBHandler struct {
	svc service.CMDBService
}

var GlobalCMDBHandler *CMDBHandler

// InitCMDBHandler initializes the singleton CMDB handler
func InitCMDBHandler(svc service.CMDBService) {
	GlobalCMDBHandler = &CMDBHandler{svc: svc}
}

// NewCMDBHandler creates a new CMDBHandler instance
func NewCMDBHandler(svc service.CMDBService) *CMDBHandler {
	return &CMDBHandler{svc: svc}
}

func getCMDBUser(r *http.Request) string {
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

func writeCMDBJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeCMDBError(w http.ResponseWriter, status int, msg string) {
	writeCMDBJSON(w, status, map[string]string{"error": msg})
}

func handleCMDBServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCINotFound):
		writeCMDBError(w, http.StatusNotFound, "configuration item not found")
	case errors.Is(err, service.ErrRelationshipNotFound):
		writeCMDBError(w, http.StatusNotFound, "ci relationship not found")
	case errors.Is(err, service.ErrLinkNotFound):
		writeCMDBError(w, http.StatusNotFound, "ci entity link not found")
	case errors.Is(err, service.ErrMaintenanceNotFound):
		writeCMDBError(w, http.StatusNotFound, "ci maintenance schedule not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeCMDBError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeCMDBError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrCINameRequired):
		writeCMDBError(w, http.StatusBadRequest, "configuration item name is required")
	case errors.Is(err, service.ErrInvalidCIType):
		writeCMDBError(w, http.StatusBadRequest, "invalid ci type: must be service, application, database, server, cluster, network, or cloud_resource")
	case errors.Is(err, service.ErrInvalidCICriticality):
		writeCMDBError(w, http.StatusBadRequest, "invalid criticality tier: must be tier_1_mission_critical, tier_2_business_critical, tier_3_supporting, or tier_4_non_critical")
	case errors.Is(err, service.ErrInvalidCIStatus):
		writeCMDBError(w, http.StatusBadRequest, "invalid ci status: must be planning, active, maintenance, degraded, decommissioned, or retired")
	case errors.Is(err, service.ErrInvalidCIEnvironment):
		writeCMDBError(w, http.StatusBadRequest, "invalid environment: must be production, staging, testing, or development")
	case errors.Is(err, service.ErrSelfRelationship):
		writeCMDBError(w, http.StatusBadRequest, "cannot create relationship to the same configuration item")
	case errors.Is(err, service.ErrRelationshipExists):
		writeCMDBError(w, http.StatusConflict, "relationship already exists between these configuration items")
	case errors.Is(err, service.ErrLinkAlreadyExists):
		writeCMDBError(w, http.StatusConflict, "entity is already linked to this configuration item")
	case errors.Is(err, service.ErrMaintTitleRequired):
		writeCMDBError(w, http.StatusBadRequest, "maintenance title is required")
	case errors.Is(err, service.ErrInvalidMaintDates):
		writeCMDBError(w, http.StatusBadRequest, "window_end_at must be after window_start_at")
	default:
		writeCMDBError(w, http.StatusBadRequest, err.Error())
	}
}

// CreateCI handles POST /v2/cmdb/items
func (h *CMDBHandler) CreateCI(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getCMDBUser(r)
	ci, err := h.svc.CreateCI(userID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusCreated, ci)
}

// GetCI handles GET /v2/cmdb/items/{id}
func (h *CMDBHandler) GetCI(w http.ResponseWriter, r *http.Request, id int64) {
	ci, err := h.svc.GetCI(id)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, ci)
}

// ListCIs handles GET /v2/cmdb/items
func (h *CMDBHandler) ListCIs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	var filter model.CIFilter

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
	if t := q.Get("ci_type"); t != "" {
		filter.CIType = &t
	}
	if st := q.Get("status"); st != "" {
		filter.Status = &st
	}
	if cr := q.Get("criticality"); cr != "" {
		filter.Criticality = &cr
	}
	if env := q.Get("environment"); env != "" {
		filter.Environment = &env
	}
	if owner := q.Get("owner_user_id"); owner != "" {
		filter.OwnerUserID = &owner
	}
	if search := q.Get("search"); search != "" {
		filter.Search = &search
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

	cis, err := h.svc.ListCIs(filter)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	if cis == nil {
		cis = []*model.ConfigurationItem{}
	}
	writeCMDBJSON(w, http.StatusOK, cis)
}

// UpdateCI handles PUT /v2/cmdb/items/{id}
func (h *CMDBHandler) UpdateCI(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getCMDBUser(r)
	ci, err := h.svc.UpdateCI(id, userID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, ci)
}

// DeleteCI handles DELETE /v2/cmdb/items/{id}
func (h *CMDBHandler) DeleteCI(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getCMDBUser(r)
	if err := h.svc.DeleteCI(id, userID); err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, map[string]string{"message": "configuration item deleted successfully"})
}

// UpdateCIStatus handles POST /v2/cmdb/items/{id}/status
func (h *CMDBHandler) UpdateCIStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateCIStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getCMDBUser(r)
	ci, err := h.svc.UpdateCIStatus(id, userID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, ci)
}
