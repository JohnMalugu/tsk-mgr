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

// StatusPageHandler exposes HTTP endpoints for status pages, components, incidents, maintenance, and subscribers
type StatusPageHandler struct {
	svc service.StatusPageService
}

var GlobalStatusPageHandler *StatusPageHandler

// InitStatusPageHandler initializes the singleton StatusPage handler
func InitStatusPageHandler(svc service.StatusPageService) {
	GlobalStatusPageHandler = &StatusPageHandler{svc: svc}
}

// NewStatusPageHandler creates a new StatusPageHandler instance
func NewStatusPageHandler(svc service.StatusPageService) *StatusPageHandler {
	return &StatusPageHandler{svc: svc}
}

func getStatusPageUser(r *http.Request) string {
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

func writeStatusPageJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeStatusPageError(w http.ResponseWriter, status int, msg string) {
	writeStatusPageJSON(w, status, map[string]string{"error": msg})
}

func handleStatusPageServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrStatusPageNotFound):
		writeStatusPageError(w, http.StatusNotFound, "status page not found")
	case errors.Is(err, service.ErrStatusPageNameRequired):
		writeStatusPageError(w, http.StatusBadRequest, "status page name is required")
	case errors.Is(err, service.ErrStatusPageSlugRequired):
		writeStatusPageError(w, http.StatusBadRequest, "status page slug is required")
	case errors.Is(err, service.ErrInvalidStatusPageSlug):
		writeStatusPageError(w, http.StatusBadRequest, "invalid status page slug: alphanumeric and hyphens only")
	case errors.Is(err, service.ErrDuplicateStatusPageSlug):
		writeStatusPageError(w, http.StatusConflict, "status page slug already in use")
	case errors.Is(err, service.ErrComponentNotFound):
		writeStatusPageError(w, http.StatusNotFound, "status component not found")
	case errors.Is(err, service.ErrComponentNameRequired):
		writeStatusPageError(w, http.StatusBadRequest, "status component name is required")
	case errors.Is(err, service.ErrInvalidComponentStatus):
		writeStatusPageError(w, http.StatusBadRequest, "invalid component status")
	case errors.Is(err, service.ErrStatusIncidentNotFound):
		writeStatusPageError(w, http.StatusNotFound, "status incident not found")
	case errors.Is(err, service.ErrStatusIncidentTitleReq):
		writeStatusPageError(w, http.StatusBadRequest, "status incident title is required")
	case errors.Is(err, service.ErrInvalidStatusIncidentSt):
		writeStatusPageError(w, http.StatusBadRequest, "invalid status incident status")
	case errors.Is(err, service.ErrInvalidIncidentImpact):
		writeStatusPageError(w, http.StatusBadRequest, "invalid incident impact level")
	case errors.Is(err, service.ErrStatusMaintenanceNotFound):
		writeStatusPageError(w, http.StatusNotFound, "maintenance window not found")
	case errors.Is(err, service.ErrMaintenanceTitleReq):
		writeStatusPageError(w, http.StatusBadRequest, "maintenance window title is required")
	case errors.Is(err, service.ErrInvalidMaintenanceTimes):
		writeStatusPageError(w, http.StatusBadRequest, "maintenance start time must be before end time")
	case errors.Is(err, service.ErrSubscriberNotFound):
		writeStatusPageError(w, http.StatusNotFound, "status subscriber not found")
	case errors.Is(err, service.ErrSubscriberTargetRequired):
		writeStatusPageError(w, http.StatusBadRequest, "subscriber target endpoint or email is required")
	case errors.Is(err, service.ErrInvalidSubscriberType):
		writeStatusPageError(w, http.StatusBadRequest, "invalid subscriber type: must be email, webhook, or slack")
	default:
		writeStatusPageError(w, http.StatusBadRequest, err.Error())
	}
}

// CreatePage handles POST /v2/statuspages
func (h *StatusPageHandler) CreatePage(w http.ResponseWriter, r *http.Request) {
	var req model.CreateStatusPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getStatusPageUser(r)

	page, err := h.svc.CreatePage(userID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, page)
}

// GetPage handles GET /v2/statuspages/{id}
func (h *StatusPageHandler) GetPage(w http.ResponseWriter, r *http.Request, id int64) {
	page, err := h.svc.GetPage(id)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, page)
}

// ListPages handles GET /v2/statuspages
func (h *StatusPageHandler) ListPages(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.StatusPageFilter{
		Visibility: q.Get("visibility"),
		Search:     q.Get("search"),
	}
	if ws := q.Get("workspace_id"); ws != "" {
		if wsID, err := strconv.ParseInt(ws, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if p := q.Get("project_id"); p != "" {
		if pID, err := strconv.ParseInt(p, 10, 64); err == nil {
			filter.ProjectID = &pID
		}
	}
	if l := q.Get("limit"); l != "" {
		if limit, err := strconv.Atoi(l); err == nil {
			filter.Limit = limit
		}
	}
	if o := q.Get("offset"); o != "" {
		if offset, err := strconv.Atoi(o); err == nil {
			filter.Offset = offset
		}
	}

	pages, total, err := h.svc.ListPages(filter)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}

	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"pages": pages,
		"total": total,
	})
}

// UpdatePage handles PUT/PATCH /v2/statuspages/{id}
func (h *StatusPageHandler) UpdatePage(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateStatusPageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	page, err := h.svc.UpdatePage(id, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, page)
}

// DeletePage handles DELETE /v2/statuspages/{id}
func (h *StatusPageHandler) DeletePage(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeletePage(id); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "status page deleted successfully"})
}

// CreateComponentGroup handles POST /v2/statuspages/{id}/groups
func (h *StatusPageHandler) CreateComponentGroup(w http.ResponseWriter, r *http.Request, pageID int64) {
	var req model.CreateComponentGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	group, err := h.svc.CreateComponentGroup(pageID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, group)
}

// ListComponentGroups handles GET /v2/statuspages/{id}/groups
func (h *StatusPageHandler) ListComponentGroups(w http.ResponseWriter, r *http.Request, pageID int64) {
	groups, err := h.svc.ListComponentGroups(pageID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groups,
	})
}

// DeleteComponentGroup handles DELETE /v2/statuspages/{id}/groups/{groupId}
func (h *StatusPageHandler) DeleteComponentGroup(w http.ResponseWriter, r *http.Request, groupID int64) {
	if err := h.svc.DeleteComponentGroup(groupID); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "component group deleted successfully"})
}

// CreateComponent handles POST /v2/statuspages/{id}/components
func (h *StatusPageHandler) CreateComponent(w http.ResponseWriter, r *http.Request, pageID int64) {
	var req model.CreateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	comp, err := h.svc.CreateComponent(pageID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, comp)
}

// GetComponent handles GET /v2/statuspages/{id}/components/{compId}
func (h *StatusPageHandler) GetComponent(w http.ResponseWriter, r *http.Request, compID int64) {
	comp, err := h.svc.GetComponent(compID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, comp)
}

// ListComponents handles GET /v2/statuspages/{id}/components
func (h *StatusPageHandler) ListComponents(w http.ResponseWriter, r *http.Request, pageID int64) {
	components, err := h.svc.ListComponents(pageID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"components": components,
	})
}

// UpdateComponent handles PUT/PATCH /v2/statuspages/{id}/components/{compId}
func (h *StatusPageHandler) UpdateComponent(w http.ResponseWriter, r *http.Request, compID int64) {
	var req model.UpdateComponentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	comp, err := h.svc.UpdateComponent(compID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, comp)
}

// DeleteComponent handles DELETE /v2/statuspages/{id}/components/{compId}
func (h *StatusPageHandler) DeleteComponent(w http.ResponseWriter, r *http.Request, compID int64) {
	if err := h.svc.DeleteComponent(compID); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "component deleted successfully"})
}

// SetComponentStatus handles POST/PATCH /v2/statuspages/{id}/components/{compId}/status
func (h *StatusPageHandler) SetComponentStatus(w http.ResponseWriter, r *http.Request, compID int64) {
	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	if err := h.svc.SetComponentStatus(compID, req.Status); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{
		"message": "component status updated successfully",
		"status":  req.Status,
	})
}

// CreateIncident handles POST /v2/statuspages/{id}/incidents
func (h *StatusPageHandler) CreateIncident(w http.ResponseWriter, r *http.Request, pageID int64) {
	var req model.CreateStatusIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getStatusPageUser(r)

	inc, err := h.svc.CreateIncident(userID, pageID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, inc)
}

// GetIncident handles GET /v2/statuspages/{id}/incidents/{incId}
func (h *StatusPageHandler) GetIncident(w http.ResponseWriter, r *http.Request, incID int64) {
	inc, err := h.svc.GetIncident(incID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, inc)
}

// ListIncidents handles GET /v2/statuspages/{id}/incidents
func (h *StatusPageHandler) ListIncidents(w http.ResponseWriter, r *http.Request, pageID int64) {
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

	incidents, total, err := h.svc.ListIncidents(pageID, limit, offset)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}

	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"incidents": incidents,
		"total":     total,
	})
}

// UpdateIncident handles PUT/PATCH /v2/statuspages/{id}/incidents/{incId}
func (h *StatusPageHandler) UpdateIncident(w http.ResponseWriter, r *http.Request, incID int64) {
	var req model.UpdateStatusIncidentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	inc, err := h.svc.UpdateIncident(incID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, inc)
}

// DeleteIncident handles DELETE /v2/statuspages/{id}/incidents/{incId}
func (h *StatusPageHandler) DeleteIncident(w http.ResponseWriter, r *http.Request, incID int64) {
	if err := h.svc.DeleteIncident(incID); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "status incident deleted successfully"})
}

// AddIncidentUpdate handles POST /v2/statuspages/{id}/incidents/{incId}/updates
func (h *StatusPageHandler) AddIncidentUpdate(w http.ResponseWriter, r *http.Request, incID int64) {
	var req model.AddStatusIncidentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getStatusPageUser(r)

	update, err := h.svc.AddIncidentUpdate(userID, incID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, update)
}

// CreateMaintenance handles POST /v2/statuspages/{id}/maintenance
func (h *StatusPageHandler) CreateMaintenance(w http.ResponseWriter, r *http.Request, pageID int64) {
	var req model.CreateMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getStatusPageUser(r)

	maint, err := h.svc.CreateMaintenance(userID, pageID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, maint)
}

// GetMaintenance handles GET /v2/statuspages/{id}/maintenance/{maintId}
func (h *StatusPageHandler) GetMaintenance(w http.ResponseWriter, r *http.Request, maintID int64) {
	maint, err := h.svc.GetMaintenance(maintID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, maint)
}

// ListMaintenances handles GET /v2/statuspages/{id}/maintenance
func (h *StatusPageHandler) ListMaintenances(w http.ResponseWriter, r *http.Request, pageID int64) {
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

	maintenances, total, err := h.svc.ListMaintenances(pageID, limit, offset)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}

	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"maintenances": maintenances,
		"total":        total,
	})
}

// UpdateMaintenance handles PUT/PATCH /v2/statuspages/{id}/maintenance/{maintId}
func (h *StatusPageHandler) UpdateMaintenance(w http.ResponseWriter, r *http.Request, maintID int64) {
	var req model.UpdateMaintenanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	maint, err := h.svc.UpdateMaintenance(maintID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, maint)
}

// DeleteMaintenance handles DELETE /v2/statuspages/{id}/maintenance/{maintId}
func (h *StatusPageHandler) DeleteMaintenance(w http.ResponseWriter, r *http.Request, maintID int64) {
	if err := h.svc.DeleteMaintenance(maintID); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "maintenance window deleted successfully"})
}

// CreateSubscriber handles POST /v2/statuspages/{id}/subscribers
func (h *StatusPageHandler) CreateSubscriber(w http.ResponseWriter, r *http.Request, pageID int64) {
	var req model.CreateSubscriberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeStatusPageError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	sub, err := h.svc.CreateSubscriber(pageID, req)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusCreated, sub)
}

// ListSubscribers handles GET /v2/statuspages/{id}/subscribers
func (h *StatusPageHandler) ListSubscribers(w http.ResponseWriter, r *http.Request, pageID int64) {
	subs, err := h.svc.ListSubscribers(pageID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"subscribers": subs,
	})
}

// DeleteSubscriber handles DELETE /v2/statuspages/{id}/subscribers/{subId}
func (h *StatusPageHandler) DeleteSubscriber(w http.ResponseWriter, r *http.Request, subID int64) {
	if err := h.svc.DeleteSubscriber(subID); err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, map[string]string{"message": "subscriber removed successfully"})
}

// GetPageMetrics handles GET /v2/statuspages/{id}/metrics
func (h *StatusPageHandler) GetPageMetrics(w http.ResponseWriter, r *http.Request, pageID int64) {
	metrics, err := h.svc.GetPageMetrics(pageID)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}
	writeStatusPageJSON(w, http.StatusOK, metrics)
}

// GetPublicOverview handles GET /v2/statuspages/public/{slug}
func (h *StatusPageHandler) GetPublicOverview(w http.ResponseWriter, r *http.Request, slug string) {
	page, groups, incidents, err := h.svc.GetPublicStatusOverview(slug)
	if err != nil {
		handleStatusPageServiceError(w, err)
		return
	}

	metrics, _ := h.svc.GetPageMetrics(page.ID)

	writeStatusPageJSON(w, http.StatusOK, map[string]interface{}{
		"page":      page,
		"groups":    groups,
		"incidents": incidents,
		"metrics":   metrics,
	})
}

// ServeHTTP routes requests under /v2/statuspages
func (h *StatusPageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/statuspages")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListPages(w, r)
		case http.MethodPost:
			h.CreatePage(w, r)
		default:
			writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "public/") {
		slug := strings.TrimPrefix(path, "public/")
		slug = strings.TrimSpace(slug)
		if r.Method != http.MethodGet {
			writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetPublicOverview(w, r, slug)
		return
	}

	parts := strings.Split(path, "/")
	if parts[0] == "by-slug" && len(parts) >= 2 {
		if r.Method != http.MethodGet {
			writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetPublicOverview(w, r, parts[1])
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeStatusPageError(w, http.StatusBadRequest, "invalid status page id")
		return
	}

	if len(parts) == 1 {
		switch r.Method {
		case http.MethodGet:
			h.GetPage(w, r, id)
		case http.MethodPut, http.MethodPatch:
			h.UpdatePage(w, r, id)
		case http.MethodDelete:
			h.DeletePage(w, r, id)
		default:
			writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "metrics":
			if r.Method != http.MethodGet {
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.GetPageMetrics(w, r, id)
		case "groups":
			switch r.Method {
			case http.MethodGet:
				h.ListComponentGroups(w, r, id)
			case http.MethodPost:
				h.CreateComponentGroup(w, r, id)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "components":
			switch r.Method {
			case http.MethodGet:
				h.ListComponents(w, r, id)
			case http.MethodPost:
				h.CreateComponent(w, r, id)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "incidents":
			switch r.Method {
			case http.MethodGet:
				h.ListIncidents(w, r, id)
			case http.MethodPost:
				h.CreateIncident(w, r, id)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "maintenance":
			switch r.Method {
			case http.MethodGet:
				h.ListMaintenances(w, r, id)
			case http.MethodPost:
				h.CreateMaintenance(w, r, id)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "subscribers":
			switch r.Method {
			case http.MethodGet:
				h.ListSubscribers(w, r, id)
			case http.MethodPost:
				h.CreateSubscriber(w, r, id)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		default:
			writeStatusPageError(w, http.StatusNotFound, "not found")
		}
		return
	}

	if len(parts) == 3 {
		subID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || subID <= 0 {
			writeStatusPageError(w, http.StatusBadRequest, "invalid resource id")
			return
		}

		switch parts[1] {
		case "groups":
			if r.Method != http.MethodDelete {
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.DeleteComponentGroup(w, r, subID)
		case "components":
			switch r.Method {
			case http.MethodGet:
				h.GetComponent(w, r, subID)
			case http.MethodPut, http.MethodPatch:
				h.UpdateComponent(w, r, subID)
			case http.MethodDelete:
				h.DeleteComponent(w, r, subID)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "incidents":
			switch r.Method {
			case http.MethodGet:
				h.GetIncident(w, r, subID)
			case http.MethodPut, http.MethodPatch:
				h.UpdateIncident(w, r, subID)
			case http.MethodDelete:
				h.DeleteIncident(w, r, subID)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "maintenance":
			switch r.Method {
			case http.MethodGet:
				h.GetMaintenance(w, r, subID)
			case http.MethodPut, http.MethodPatch:
				h.UpdateMaintenance(w, r, subID)
			case http.MethodDelete:
				h.DeleteMaintenance(w, r, subID)
			default:
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "subscribers":
			if r.Method != http.MethodDelete {
				writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.DeleteSubscriber(w, r, subID)
		default:
			writeStatusPageError(w, http.StatusNotFound, "not found")
		}
		return
	}

	if len(parts) == 4 {
		subID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || subID <= 0 {
			writeStatusPageError(w, http.StatusBadRequest, "invalid resource id")
			return
		}

		switch parts[1] {
		case "components":
			if parts[3] == "status" {
				if r.Method != http.MethodPost && r.Method != http.MethodPatch {
					writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
					return
				}
				h.SetComponentStatus(w, r, subID)
				return
			}
		case "incidents":
			if parts[3] == "updates" {
				if r.Method != http.MethodPost {
					writeStatusPageError(w, http.StatusMethodNotAllowed, "method not allowed")
					return
				}
				h.AddIncidentUpdate(w, r, subID)
				return
			}
		}
	}

	writeStatusPageError(w, http.StatusNotFound, "not found")
}
