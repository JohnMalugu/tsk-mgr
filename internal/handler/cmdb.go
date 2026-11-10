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

// CreateRelationship handles POST /v2/cmdb/items/{id}/relationships
func (h *CMDBHandler) CreateRelationship(w http.ResponseWriter, r *http.Request, ciID int64) {
	var req model.CreateCIRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	rel, err := h.svc.CreateRelationship(ciID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusCreated, rel)
}

// ListRelationships handles GET /v2/cmdb/items/{id}/relationships
func (h *CMDBHandler) ListRelationships(w http.ResponseWriter, r *http.Request, ciID int64) {
	rels, err := h.svc.ListRelationships(ciID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	if rels == nil {
		rels = []model.CIRelationship{}
	}
	writeCMDBJSON(w, http.StatusOK, rels)
}

// DeleteRelationship handles DELETE /v2/cmdb/relationships/{id}
func (h *CMDBHandler) DeleteRelationship(w http.ResponseWriter, r *http.Request, relID int64) {
	if err := h.svc.DeleteRelationship(relID); err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, map[string]string{"message": "relationship deleted successfully"})
}

// GetDependencyGraph handles GET /v2/cmdb/items/{id}/graph
func (h *CMDBHandler) GetDependencyGraph(w http.ResponseWriter, r *http.Request, ciID int64) {
	maxDepth := 5
	if dStr := r.URL.Query().Get("max_depth"); dStr != "" {
		if d, err := strconv.Atoi(dStr); err == nil && d > 0 {
			maxDepth = d
		}
	}
	graph, err := h.svc.GetDependencyGraph(ciID, maxDepth)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, graph)
}

// AnalyzeBlastRadius handles GET /v2/cmdb/items/{id}/blast-radius
func (h *CMDBHandler) AnalyzeBlastRadius(w http.ResponseWriter, r *http.Request, ciID int64) {
	analysis, err := h.svc.AnalyzeBlastRadius(ciID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, analysis)
}

// LinkEntity handles POST /v2/cmdb/items/{id}/links
func (h *CMDBHandler) LinkEntity(w http.ResponseWriter, r *http.Request, ciID int64) {
	var req model.CreateCIEntityLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	link, err := h.svc.LinkEntity(ciID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusCreated, link)
}

// ListEntityLinks handles GET /v2/cmdb/items/{id}/links
func (h *CMDBHandler) ListEntityLinks(w http.ResponseWriter, r *http.Request, ciID int64) {
	links, err := h.svc.ListEntityLinks(ciID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	if links == nil {
		links = []model.CIEntityLink{}
	}
	writeCMDBJSON(w, http.StatusOK, links)
}

// ListCIsForEntity handles GET /v2/cmdb/links?entity_type=X&entity_id=Y
func (h *CMDBHandler) ListCIsForEntity(w http.ResponseWriter, r *http.Request) {
	entityType := r.URL.Query().Get("entity_type")
	entityIDStr := r.URL.Query().Get("entity_id")
	if entityType == "" || entityIDStr == "" {
		writeCMDBError(w, http.StatusBadRequest, "entity_type and entity_id query parameters are required")
		return
	}
	entityID, err := strconv.ParseInt(entityIDStr, 10, 64)
	if err != nil || entityID <= 0 {
		writeCMDBError(w, http.StatusBadRequest, "invalid entity_id")
		return
	}
	links, err := h.svc.ListCIsForEntity(entityType, entityID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	if links == nil {
		links = []model.CIEntityLink{}
	}
	writeCMDBJSON(w, http.StatusOK, links)
}

// UnlinkEntity handles DELETE /v2/cmdb/links/{id}
func (h *CMDBHandler) UnlinkEntity(w http.ResponseWriter, r *http.Request, linkID int64) {
	if err := h.svc.UnlinkEntity(linkID); err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, map[string]string{"message": "entity link deleted successfully"})
}

// CreateMaintenance handles POST /v2/cmdb/items/{id}/maintenance
func (h *CMDBHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request, ciID int64) {
	var req model.CreateCIMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getCMDBUser(r)
	maint, err := h.svc.CreateMaintenance(ciID, userID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusCreated, maint)
}

// ListMaintenances handles GET /v2/cmdb/items/{id}/maintenance
func (h *CMDBHandler) ListMaintenances(w http.ResponseWriter, r *http.Request, ciID int64) {
	maints, err := h.svc.ListMaintenances(ciID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	if maints == nil {
		maints = []model.CIMaintenanceSchedule{}
	}
	writeCMDBJSON(w, http.StatusOK, maints)
}

// UpdateMaintenanceStatus handles POST /v2/cmdb/maintenance/{id}/status
func (h *CMDBHandler) UpdateMaintenanceStatus(w http.ResponseWriter, r *http.Request, maintID int64) {
	var req model.UpdateCIMaintenanceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeCMDBError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	maint, err := h.svc.UpdateMaintenanceStatus(maintID, req)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, maint)
}

// DeleteMaintenance handles DELETE /v2/cmdb/maintenance/{id}
func (h *CMDBHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request, maintID int64) {
	if err := h.svc.DeleteMaintenance(maintID); err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, map[string]string{"message": "maintenance schedule deleted successfully"})
}

// GetMetrics handles GET /v2/cmdb/metrics?workspace_id=X
func (h *CMDBHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		writeCMDBError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeCMDBError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}
	metrics, err := h.svc.GetCMDBMetrics(wsID)
	if err != nil {
		handleCMDBServiceError(w, err)
		return
	}
	writeCMDBJSON(w, http.StatusOK, metrics)
}

// ServeHTTP dispatches HTTP requests under /v2/cmdb
func (h *CMDBHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/cmdb")
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	if path == "metrics" {
		if r.Method != http.MethodGet {
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetMetrics(w, r)
		return
	}

	if path == "links" {
		if r.Method != http.MethodGet {
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.ListCIsForEntity(w, r)
		return
	}

	if strings.HasPrefix(path, "links/") {
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			linkID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || linkID <= 0 {
				writeCMDBError(w, http.StatusBadRequest, "invalid link id")
				return
			}
			if r.Method == http.MethodDelete {
				h.UnlinkEntity(w, r, linkID)
				return
			}
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	if strings.HasPrefix(path, "relationships/") {
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			relID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || relID <= 0 {
				writeCMDBError(w, http.StatusBadRequest, "invalid relationship id")
				return
			}
			if r.Method == http.MethodDelete {
				h.DeleteRelationship(w, r, relID)
				return
			}
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	if strings.HasPrefix(path, "maintenance/") {
		parts := strings.Split(path, "/")
		if len(parts) == 2 {
			maintID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || maintID <= 0 {
				writeCMDBError(w, http.StatusBadRequest, "invalid maintenance id")
				return
			}
			if r.Method == http.MethodDelete {
				h.DeleteMaintenance(w, r, maintID)
				return
			}
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		if len(parts) == 3 && parts[2] == "status" {
			maintID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || maintID <= 0 {
				writeCMDBError(w, http.StatusBadRequest, "invalid maintenance id")
				return
			}
			if r.Method == http.MethodPost || r.Method == http.MethodPatch {
				h.UpdateMaintenanceStatus(w, r, maintID)
				return
			}
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
	}

	// Routes under /v2/cmdb/items or /v2/cmdb
	itemPath := path
	if strings.HasPrefix(itemPath, "items") {
		itemPath = strings.TrimPrefix(itemPath, "items")
		itemPath = strings.TrimPrefix(itemPath, "/")
	}

	if itemPath == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListCIs(w, r)
		case http.MethodPost:
			h.CreateCI(w, r)
		default:
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	parts := strings.Split(itemPath, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeCMDBError(w, http.StatusBadRequest, "invalid configuration item id")
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetCI(w, r, id)
		case http.MethodPut, http.MethodPatch:
			h.UpdateCI(w, r, id)
		case http.MethodDelete:
			h.DeleteCI(w, r, id)
		default:
			writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "status":
			if r.Method != http.MethodPost && r.Method != http.MethodPatch {
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.UpdateCIStatus(w, r, id)
		case "relationships":
			switch r.Method {
			case http.MethodGet:
				h.ListRelationships(w, r, id)
			case http.MethodPost:
				h.CreateRelationship(w, r, id)
			default:
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "graph":
			if r.Method != http.MethodGet {
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.GetDependencyGraph(w, r, id)
		case "blast-radius":
			if r.Method != http.MethodGet {
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.AnalyzeBlastRadius(w, r, id)
		case "links":
			switch r.Method {
			case http.MethodGet:
				h.ListEntityLinks(w, r, id)
			case http.MethodPost:
				h.LinkEntity(w, r, id)
			default:
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "maintenance":
			switch r.Method {
			case http.MethodGet:
				h.ListMaintenances(w, r, id)
			case http.MethodPost:
				h.CreateMaintenance(w, r, id)
			default:
				writeCMDBError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		default:
			writeCMDBError(w, http.StatusNotFound, "not found")
		}
		return
	}

	writeCMDBError(w, http.StatusNotFound, "not found")
}
