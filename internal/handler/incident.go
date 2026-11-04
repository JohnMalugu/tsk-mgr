package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// IncidentHandler exposes HTTP endpoints for operational incidents and postmortems
type IncidentHandler struct {
	svc service.IncidentService
}

var GlobalIncidentHandler *IncidentHandler

// InitIncidentHandler initializes the singleton incident handler
func InitIncidentHandler(svc service.IncidentService) {
	GlobalIncidentHandler = &IncidentHandler{svc: svc}
}

// NewIncidentHandler creates a new IncidentHandler instance
func NewIncidentHandler(svc service.IncidentService) *IncidentHandler {
	return &IncidentHandler{svc: svc}
}

func getIncidentUser(r *http.Request) string {
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

func writeIncidentJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeIncidentError(w http.ResponseWriter, status int, msg string) {
	writeIncidentJSON(w, status, map[string]string{"error": msg})
}

// HandleIncidents handles GET and POST for /v2/incidents
func (h *IncidentHandler) HandleIncidents(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeIncidentError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeIncidentError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		filter := model.IncidentFilter{
			WorkspaceID:  wsID,
			Severity:     r.URL.Query().Get("severity"),
			Status:       r.URL.Query().Get("status"),
			CommanderID:  r.URL.Query().Get("commander_id"),
			ReportedByID: r.URL.Query().Get("reported_by_id"),
			Search:       r.URL.Query().Get("search"),
		}

		if projStr := r.URL.Query().Get("project_id"); projStr != "" {
			if pid, err := strconv.Atoi(projStr); err == nil {
				filter.ProjectID = &pid
			}
		}

		if limStr := r.URL.Query().Get("limit"); limStr != "" {
			if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
				filter.Limit = lim
			}
		}
		if offStr := r.URL.Query().Get("offset"); offStr != "" {
			if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
				filter.Offset = off
			}
		}

		incidents, err := h.svc.ListIncidents(filter)
		if err != nil {
			if errors.Is(err, service.ErrWorkspaceRequired) {
				writeIncidentError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeIncidentError(w, http.StatusInternalServerError, "Failed to list incidents: "+err.Error())
			return
		}
		writeIncidentJSON(w, http.StatusOK, incidents)

	case http.MethodPost:
		var req model.CreateIncidentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeIncidentError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		reporterID := getIncidentUser(r)
		inc, err := h.svc.CreateIncident(reporterID, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrWorkspaceRequired),
				errors.Is(err, service.ErrIncidentTitleRequired),
				errors.Is(err, service.ErrInvalidIncidentSeverity):
				writeIncidentError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrProjectNotFound):
				writeIncidentError(w, http.StatusNotFound, err.Error())
			default:
				writeIncidentError(w, http.StatusInternalServerError, "Failed to create incident: "+err.Error())
			}
			return
		}
		writeIncidentJSON(w, http.StatusCreated, inc)

	default:
		writeIncidentError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleIncidentDetail handles GET, PUT, and DELETE for /v2/incidents/{id}
func (h *IncidentHandler) HandleIncidentDetail(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		inc, err := h.svc.GetIncident(id)
		if err != nil {
			if errors.Is(err, service.ErrIncidentNotFound) {
				writeIncidentError(w, http.StatusNotFound, err.Error())
				return
			}
			writeIncidentError(w, http.StatusInternalServerError, "Failed to fetch incident: "+err.Error())
			return
		}
		writeIncidentJSON(w, http.StatusOK, inc)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateIncidentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeIncidentError(w, http.StatusBadRequest, "Invalid request payload: "+err.Error())
			return
		}

		inc, err := h.svc.UpdateIncident(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrIncidentTitleRequired):
				writeIncidentError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrIncidentNotFound), errors.Is(err, service.ErrProjectNotFound):
				writeIncidentError(w, http.StatusNotFound, err.Error())
			default:
				writeIncidentError(w, http.StatusInternalServerError, "Failed to update incident: "+err.Error())
			}
			return
		}
		writeIncidentJSON(w, http.StatusOK, inc)

	case http.MethodDelete:
		err := h.svc.DeleteIncident(id)
		if err != nil {
			if errors.Is(err, service.ErrIncidentNotFound) {
				writeIncidentError(w, http.StatusNotFound, err.Error())
				return
			}
			writeIncidentError(w, http.StatusInternalServerError, "Failed to delete incident: "+err.Error())
			return
		}
		writeIncidentJSON(w, http.StatusOK, map[string]string{"message": "Incident deleted successfully"})

	default:
		writeIncidentError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}
