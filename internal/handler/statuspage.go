package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
