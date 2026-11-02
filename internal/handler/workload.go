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

// WorkloadHandler manages HTTP endpoints for Resource Profiles, Time-offs, Allocations, Heatmap, and Workload Analytics
type WorkloadHandler struct {
	svc service.WorkloadService
}

var GlobalWorkloadHandler *WorkloadHandler

// InitWorkloadHandler initializes the global workload handler singleton
func InitWorkloadHandler(svc service.WorkloadService) {
	GlobalWorkloadHandler = &WorkloadHandler{svc: svc}
}

// NewWorkloadHandler creates a new WorkloadHandler instance
func NewWorkloadHandler(svc service.WorkloadService) *WorkloadHandler {
	return &WorkloadHandler{svc: svc}
}

func getWorkloadUser(r *http.Request) string {
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

func writeWorkloadJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeWorkloadError(w http.ResponseWriter, status int, msg string) {
	writeWorkloadJSON(w, status, map[string]string{"error": msg})
}

// HandleProfiles handles GET and POST for /v2/workload/profiles
func (h *WorkloadHandler) HandleProfiles(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		var isActive *bool
		if actStr := r.URL.Query().Get("is_active"); actStr != "" {
			b := actStr == "true" || actStr == "1"
			isActive = &b
		}

		profiles, err := h.svc.ListProfiles(wsID, isActive)
		if err != nil {
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to list profiles: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, profiles)

	case http.MethodPost:
		var req model.CreateResourceProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}
		if req.UserID == "" {
			req.UserID = getWorkloadUser(r)
		}

		profile, err := h.svc.CreateProfile(req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrUserIDRequired), errors.Is(err, service.ErrInvalidCapacity):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			case errors.Is(err, service.ErrProfileExists):
				writeWorkloadError(w, http.StatusConflict, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to create profile: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusCreated, profile)

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleProfileDetail handles GET, PUT, and DELETE for /v2/workload/profiles/{id}
func (h *WorkloadHandler) HandleProfileDetail(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		profile, err := h.svc.GetProfile(id)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Resource profile not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to get profile: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, profile)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateResourceProfileRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		profile, err := h.svc.UpdateProfile(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrProfileNotFound):
				writeWorkloadError(w, http.StatusNotFound, "Resource profile not found")
			case errors.Is(err, service.ErrInvalidCapacity):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to update profile: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusOK, profile)

	case http.MethodDelete:
		err := h.svc.DeleteProfile(id)
		if err != nil {
			if errors.Is(err, service.ErrProfileNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Resource profile not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to delete profile: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, map[string]string{"message": "Resource profile deleted successfully"})

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleTimeOffs handles GET and POST for /v2/workload/time-off
func (h *WorkloadHandler) HandleTimeOffs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		userID := r.URL.Query().Get("user_id")

		var from, to *time.Time
		if fromStr := r.URL.Query().Get("from"); fromStr != "" {
			if t, err := time.Parse("2006-01-02", fromStr); err == nil {
				from = &t
			} else if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
				from = &t
			}
		}
		if toStr := r.URL.Query().Get("to"); toStr != "" {
			if t, err := time.Parse("2006-01-02", toStr); err == nil {
				to = &t
			} else if t, err := time.Parse(time.RFC3339, toStr); err == nil {
				to = &t
			}
		}

		timeOffs, err := h.svc.ListTimeOffs(wsID, userID, from, to)
		if err != nil {
			if errors.Is(err, service.ErrInvalidDateRange) {
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to list time-offs: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, timeOffs)

	case http.MethodPost:
		var req model.CreateTimeOffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}
		if req.UserID == "" {
			req.UserID = getWorkloadUser(r)
		}

		timeOff, err := h.svc.CreateTimeOff(req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrUserIDRequired),
				errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidTimeOffType),
				errors.Is(err, service.ErrInvalidCapacity):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to create time-off: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusCreated, timeOff)

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleTimeOffDetail handles GET, PUT, and DELETE for /v2/workload/time-off/{id}
func (h *WorkloadHandler) HandleTimeOffDetail(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		to, err := h.svc.GetTimeOff(id)
		if err != nil {
			if errors.Is(err, service.ErrTimeOffNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Resource time-off not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to get time-off: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, to)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateTimeOffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		to, err := h.svc.UpdateTimeOff(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrTimeOffNotFound):
				writeWorkloadError(w, http.StatusNotFound, "Resource time-off not found")
			case errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidTimeOffType), errors.Is(err, service.ErrInvalidCapacity):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to update time-off: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusOK, to)

	case http.MethodDelete:
		err := h.svc.DeleteTimeOff(id)
		if err != nil {
			if errors.Is(err, service.ErrTimeOffNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Resource time-off not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to delete time-off: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, map[string]string{"message": "Resource time-off deleted successfully"})

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// Dispatch helper for /v2/workload subroutes
func (h *WorkloadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/workload")
	path = strings.TrimPrefix(path, "/")

	if path == "profiles" || path == "profiles/" {
		h.HandleProfiles(w, r)
		return
	}
	if strings.HasPrefix(path, "profiles/") {
		idStr := strings.TrimPrefix(path, "profiles/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid profile id")
			return
		}
		h.HandleProfileDetail(w, r, id)
		return
	}

	if path == "time-off" || path == "time-off/" {
		h.HandleTimeOffs(w, r)
		return
	}
	if strings.HasPrefix(path, "time-off/") {
		idStr := strings.TrimPrefix(path, "time-off/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid time-off id")
			return
		}
		h.HandleTimeOffDetail(w, r, id)
		return
	}

	if path == "allocations" || path == "allocations/" {
		h.HandleAllocations(w, r)
		return
	}
	if strings.HasPrefix(path, "allocations/") {
		idStr := strings.TrimPrefix(path, "allocations/")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil || id <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid allocation id")
			return
		}
		h.HandleAllocationDetail(w, r, id)
		return
	}

	if strings.HasPrefix(path, "users/") {
		userID := strings.TrimPrefix(path, "users/")
		h.HandleUserWorkload(w, r, userID)
		return
	}

	if path == "heatmap" || path == "heatmap/" {
		h.HandleHeatmap(w, r)
		return
	}

	if path == "overallocated" || path == "overallocated/" {
		h.HandleOverallocated(w, r)
		return
	}

	if path == "metrics" || path == "metrics/" {
		h.HandleMetrics(w, r)
		return
	}

	writeWorkloadError(w, http.StatusNotFound, "endpoint not found")
}

// HandleAllocations handles GET and POST for /v2/workload/allocations
func (h *WorkloadHandler) HandleAllocations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		var projID int
		if projStr := r.URL.Query().Get("project_id"); projStr != "" {
			if id, err := strconv.Atoi(projStr); err == nil && id > 0 {
				projID = id
			}
		}

		userID := r.URL.Query().Get("user_id")
		status := r.URL.Query().Get("status")

		allocs, err := h.svc.ListAllocations(wsID, projID, userID, status)
		if err != nil {
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to list allocations: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, allocs)

	case http.MethodPost:
		var req model.CreateResourceAllocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}
		if req.UserID == "" {
			req.UserID = getWorkloadUser(r)
		}

		alloc, err := h.svc.CreateAllocation(req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrUserIDRequired),
				errors.Is(err, service.ErrProjectNotFound), errors.Is(err, service.ErrInvalidDateRange),
				errors.Is(err, service.ErrInvalidCapacity):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to create allocation: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusCreated, alloc)

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleAllocationDetail handles GET, PUT, and DELETE for /v2/workload/allocations/{id}
func (h *WorkloadHandler) HandleAllocationDetail(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		alloc, err := h.svc.GetAllocation(id)
		if err != nil {
			if errors.Is(err, service.ErrAllocationNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Project resource allocation not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to get allocation: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, alloc)

	case http.MethodPut, http.MethodPatch:
		var req model.UpdateResourceAllocationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWorkloadError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		alloc, err := h.svc.UpdateAllocation(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrAllocationNotFound):
				writeWorkloadError(w, http.StatusNotFound, "Project resource allocation not found")
			case errors.Is(err, service.ErrInvalidCapacity), errors.Is(err, service.ErrInvalidDateRange), errors.Is(err, service.ErrInvalidAllocationStatus):
				writeWorkloadError(w, http.StatusBadRequest, err.Error())
			default:
				writeWorkloadError(w, http.StatusInternalServerError, "Failed to update allocation: "+err.Error())
			}
			return
		}
		writeWorkloadJSON(w, http.StatusOK, alloc)

	case http.MethodDelete:
		err := h.svc.DeleteAllocation(id)
		if err != nil {
			if errors.Is(err, service.ErrAllocationNotFound) {
				writeWorkloadError(w, http.StatusNotFound, "Project resource allocation not found")
				return
			}
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to delete allocation: "+err.Error())
			return
		}
		writeWorkloadJSON(w, http.StatusOK, map[string]string{"message": "Project resource allocation deleted successfully"})

	default:
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func parseDate(val string) time.Time {
	if val == "" {
		return time.Time{}
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return t
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return t
	}
	return time.Time{}
}

// HandleUserWorkload handles GET /v2/workload/users/{userId}
func (h *WorkloadHandler) HandleUserWorkload(w http.ResponseWriter, r *http.Request, userID string) {
	if r.Method != http.MethodGet {
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	from := parseDate(r.URL.Query().Get("from"))
	to := parseDate(r.URL.Query().Get("to"))

	summary, err := h.svc.GetUserWorkload(wsID, userID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrUserIDRequired), errors.Is(err, service.ErrInvalidDateRange):
			writeWorkloadError(w, http.StatusBadRequest, err.Error())
		default:
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to calculate user workload: "+err.Error())
		}
		return
	}
	writeWorkloadJSON(w, http.StatusOK, summary)
}

// HandleHeatmap handles GET /v2/workload/heatmap
func (h *WorkloadHandler) HandleHeatmap(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	from := parseDate(r.URL.Query().Get("from"))
	to := parseDate(r.URL.Query().Get("to"))

	heatmap, err := h.svc.GetTeamHeatmap(wsID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrInvalidDateRange):
			writeWorkloadError(w, http.StatusBadRequest, err.Error())
		default:
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to generate workload heatmap: "+err.Error())
		}
		return
	}
	writeWorkloadJSON(w, http.StatusOK, heatmap)
}

// HandleOverallocated handles GET /v2/workload/overallocated
func (h *WorkloadHandler) HandleOverallocated(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	from := parseDate(r.URL.Query().Get("from"))
	to := parseDate(r.URL.Query().Get("to"))

	alerts, err := h.svc.GetOverallocatedAlerts(wsID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired), errors.Is(err, service.ErrInvalidDateRange):
			writeWorkloadError(w, http.StatusBadRequest, err.Error())
		default:
			writeWorkloadError(w, http.StatusInternalServerError, "Failed to retrieve overallocated alerts: "+err.Error())
		}
		return
	}
	writeWorkloadJSON(w, http.StatusOK, alerts)
}

// HandleMetrics handles GET /v2/workload/metrics
func (h *WorkloadHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeWorkloadError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeWorkloadError(w, http.StatusBadRequest, "workspace_id query parameter is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeWorkloadError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	metrics, err := h.svc.GetWorkspaceMetrics(wsID)
	if err != nil {
		if errors.Is(err, service.ErrWorkspaceRequired) {
			writeWorkloadError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeWorkloadError(w, http.StatusInternalServerError, "Failed to get workspace workload metrics: "+err.Error())
		return
	}
	writeWorkloadJSON(w, http.StatusOK, metrics)
}
