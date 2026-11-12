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

// OnCallHandler exposes HTTP endpoints for on-call rotations, escalation policies, and paging
type OnCallHandler struct {
	svc service.OnCallService
}

// GlobalOnCallHandler is the package singleton for OnCallHandler
var GlobalOnCallHandler *OnCallHandler

// InitOnCallHandler initializes the singleton on-call handler
func InitOnCallHandler(svc service.OnCallService) {
	GlobalOnCallHandler = &OnCallHandler{svc: svc}
}

// NewOnCallHandler creates a new OnCallHandler instance
func NewOnCallHandler(svc service.OnCallService) *OnCallHandler {
	return &OnCallHandler{svc: svc}
}

func getOnCallUser(r *http.Request) string {
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

func writeOnCallJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeOnCallError(w http.ResponseWriter, status int, msg string) {
	writeOnCallJSON(w, status, map[string]string{"error": msg})
}

func handleOnCallError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrScheduleNotFound),
		errors.Is(err, service.ErrEscalationPolicyNotFound),
		errors.Is(err, service.ErrOnCallOverrideNotFound),
		errors.Is(err, service.ErrOnCallPageNotFound):
		writeOnCallError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrScheduleNameRequired),
		errors.Is(err, service.ErrInvalidRotationType),
		errors.Is(err, service.ErrInvalidTimezone),
		errors.Is(err, service.ErrEscalationPolicyNameReq),
		errors.Is(err, service.ErrInvalidEscalationTarget),
		errors.Is(err, service.ErrInvalidOverrideTimeRange),
		errors.Is(err, service.ErrOverrideUserRequired),
		errors.Is(err, service.ErrPageTitleRequired):
		writeOnCallError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrNoOnCallResponders):
		writeOnCallError(w, http.StatusUnprocessableEntity, err.Error())
	default:
		writeOnCallError(w, http.StatusInternalServerError, err.Error())
	}
}

// HandleSchedules handles GET (list) and POST (create) for /v2/oncall/schedules
func (h *OnCallHandler) HandleSchedules(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		wsStr := r.URL.Query().Get("workspace_id")
		if wsStr == "" {
			writeOnCallError(w, http.StatusBadRequest, "workspace_id query parameter is required")
			return
		}
		wsID, err := strconv.ParseInt(wsStr, 10, 64)
		if err != nil || wsID <= 0 {
			writeOnCallError(w, http.StatusBadRequest, "invalid workspace_id")
			return
		}

		filter := model.OnCallFilter{
			WorkspaceID: wsID,
			Search:      r.URL.Query().Get("search"),
		}

		if projStr := r.URL.Query().Get("project_id"); projStr != "" {
			if pid, err := strconv.ParseInt(projStr, 10, 64); err == nil && pid > 0 {
				filter.ProjectID = &pid
			}
		}

		if actStr := r.URL.Query().Get("is_active"); actStr != "" {
			act := actStr == "true" || actStr == "1"
			filter.IsActive = &act
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

		schedules, total, err := h.svc.ListSchedules(filter)
		if err != nil {
			handleOnCallError(w, err)
			return
		}

		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{
			"schedules": schedules,
			"total":     total,
			"limit":     filter.Limit,
			"offset":    filter.Offset,
		})

	case http.MethodPost:
		userID := getOnCallUser(r)
		var req model.CreateScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.WorkspaceID <= 0 {
			writeOnCallError(w, http.StatusBadRequest, "workspace_id is required")
			return
		}

		schedule, err := h.svc.CreateSchedule(userID, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}

		writeOnCallJSON(w, http.StatusCreated, schedule)

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleScheduleByID handles GET, PUT, DELETE and sub-resources for /v2/oncall/schedules/{id}
func (h *OnCallHandler) HandleScheduleByID(w http.ResponseWriter, r *http.Request, id int64, subPath string) {
	switch {
	case subPath == "" || subPath == "/":
		switch r.Method {
		case http.MethodGet:
			schedule, err := h.svc.GetSchedule(id)
			if err != nil {
				handleOnCallError(w, err)
				return
			}
			writeOnCallJSON(w, http.StatusOK, schedule)

		case http.MethodPut:
			var req model.UpdateScheduleRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeOnCallError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			schedule, err := h.svc.UpdateSchedule(id, req)
			if err != nil {
				handleOnCallError(w, err)
				return
			}
			writeOnCallJSON(w, http.StatusOK, schedule)

		case http.MethodDelete:
			if err := h.svc.DeleteSchedule(id); err != nil {
				handleOnCallError(w, err)
				return
			}
			writeOnCallJSON(w, http.StatusOK, map[string]string{"message": "schedule deleted successfully"})

		default:
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	case subPath == "active" || subPath == "active/":
		if r.Method != http.MethodGet {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		responder, err := h.svc.GetActiveResponder(id)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, responder)

	case subPath == "shifts" || subPath == "shifts/":
		if r.Method != http.MethodGet {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var start, end time.Time
		if sStr := r.URL.Query().Get("start_time"); sStr != "" {
			if t, err := time.Parse(time.RFC3339, sStr); err == nil {
				start = t
			}
		}
		if eStr := r.URL.Query().Get("end_time"); eStr != "" {
			if t, err := time.Parse(time.RFC3339, eStr); err == nil {
				end = t
			}
		}
		shifts, err := h.svc.ListShifts(id, start, end)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"shifts": shifts})

	case subPath == "shifts/generate" || subPath == "shifts/generate/":
		if r.Method != http.MethodPost {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		horizon := 30
		if hStr := r.URL.Query().Get("horizon_days"); hStr != "" {
			if hd, err := strconv.Atoi(hStr); err == nil && hd > 0 {
				horizon = hd
			}
		}
		shifts, err := h.svc.GenerateShifts(id, horizon)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"shifts": shifts})

	case subPath == "participants" || subPath == "participants/":
		switch r.Method {
		case http.MethodGet:
			parts, err := h.svc.ListParticipants(id)
			if err != nil {
				handleOnCallError(w, err)
				return
			}
			writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"participants": parts})

		case http.MethodPut:
			var req struct {
				UserIDs []string `json:"user_ids"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeOnCallError(w, http.StatusBadRequest, "invalid request body")
				return
			}
			if err := h.svc.SetParticipants(id, req.UserIDs); err != nil {
				handleOnCallError(w, err)
				return
			}
			parts, _ := h.svc.ListParticipants(id)
			writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"participants": parts})

		default:
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
		}

	default:
		writeOnCallError(w, http.StatusNotFound, "resource not found")
	}
}

// ServeHTTP handles on-call HTTP requests
func (h *OnCallHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/oncall")
	path = strings.TrimPrefix(path, "/")

	if path == "schedules" || path == "schedules/" {
		h.HandleSchedules(w, r)
		return
	}

	if strings.HasPrefix(path, "schedules/") {
		remainder := strings.TrimPrefix(path, "schedules/")
		parts := strings.SplitN(remainder, "/", 2)
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid schedule id")
			return
		}
		subPath := ""
		if len(parts) > 1 {
			subPath = parts[1]
		}
		h.HandleScheduleByID(w, r, id, subPath)
		return
	}

	writeOnCallError(w, http.StatusNotFound, "endpoint not found")
}
