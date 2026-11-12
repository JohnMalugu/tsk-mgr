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

// HandleOverrides handles GET and POST for /v2/oncall/schedules/{id}/overrides
func (h *OnCallHandler) HandleOverrides(w http.ResponseWriter, r *http.Request, scheduleID int64) {
	switch r.Method {
	case http.MethodGet:
		overrides, err := h.svc.ListOverrides(scheduleID)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"overrides": overrides})

	case http.MethodPost:
		userID := getOnCallUser(r)
		var req model.CreateOverrideRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		override, err := h.svc.CreateOverride(userID, scheduleID, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusCreated, override)

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleOverrideByID handles GET and DELETE for /v2/oncall/overrides/{id}
func (h *OnCallHandler) HandleOverrideByID(w http.ResponseWriter, r *http.Request, id int64) {
	switch r.Method {
	case http.MethodGet:
		override, err := h.svc.GetOverride(id)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, override)

	case http.MethodDelete:
		if err := h.svc.DeleteOverride(id); err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]string{"message": "override deleted successfully"})

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandlePolicies handles GET and POST for /v2/oncall/policies
func (h *OnCallHandler) HandlePolicies(w http.ResponseWriter, r *http.Request) {
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

		policies, err := h.svc.ListEscalationPolicies(wsID)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{"policies": policies})

	case http.MethodPost:
		var req model.CreateEscalationPolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.WorkspaceID <= 0 {
			writeOnCallError(w, http.StatusBadRequest, "workspace_id is required")
			return
		}

		policy, err := h.svc.CreateEscalationPolicy(req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusCreated, policy)

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandleDefaultPolicy handles GET for /v2/oncall/policies/default
func (h *OnCallHandler) HandleDefaultPolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
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

	policy, err := h.svc.GetDefaultEscalationPolicy(wsID)
	if err != nil {
		handleOnCallError(w, err)
		return
	}
	writeOnCallJSON(w, http.StatusOK, policy)
}

// HandlePolicyByID handles GET, PUT, DELETE for /v2/oncall/policies/{id}
func (h *OnCallHandler) HandlePolicyByID(w http.ResponseWriter, r *http.Request, id int64, subPath string) {
	switch r.Method {
	case http.MethodGet:
		policy, err := h.svc.GetEscalationPolicy(id)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, policy)

	case http.MethodPut:
		var req model.CreateEscalationPolicyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		policy, err := h.svc.UpdateEscalationPolicy(id, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, policy)

	case http.MethodDelete:
		if err := h.svc.DeleteEscalationPolicy(id); err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, map[string]string{"message": "escalation policy deleted successfully"})

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandlePages handles GET (list) and POST (trigger) for /v2/oncall/pages
func (h *OnCallHandler) HandlePages(w http.ResponseWriter, r *http.Request) {
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

		status := r.URL.Query().Get("status")
		limit := 0
		offset := 0
		if limStr := r.URL.Query().Get("limit"); limStr != "" {
			if lim, err := strconv.Atoi(limStr); err == nil && lim > 0 {
				limit = lim
			}
		}
		if offStr := r.URL.Query().Get("offset"); offStr != "" {
			if off, err := strconv.Atoi(offStr); err == nil && off >= 0 {
				offset = off
			}
		}

		pages, total, err := h.svc.ListPages(wsID, status, limit, offset)
		if err != nil {
			handleOnCallError(w, err)
			return
		}

		writeOnCallJSON(w, http.StatusOK, map[string]interface{}{
			"pages":  pages,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		})

	case http.MethodPost:
		userID := getOnCallUser(r)
		var req model.TriggerPageRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		if req.WorkspaceID <= 0 {
			writeOnCallError(w, http.StatusBadRequest, "workspace_id is required")
			return
		}

		page, err := h.svc.TriggerPage(userID, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusCreated, page)

	default:
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// HandlePageByID handles GET, acknowledge, resolve, and escalate for /v2/oncall/pages/{id}
func (h *OnCallHandler) HandlePageByID(w http.ResponseWriter, r *http.Request, id int64, action string) {
	switch action {
	case "":
		if r.Method != http.MethodGet {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		page, err := h.svc.GetPage(id)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, page)

	case "acknowledge":
		if r.Method != http.MethodPost {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		userID := getOnCallUser(r)
		var req model.AcknowledgePageRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		page, err := h.svc.AcknowledgePage(id, userID, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, page)

	case "resolve":
		if r.Method != http.MethodPost {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		var req model.ResolvePageRequest
		_ = json.NewDecoder(r.Body).Decode(&req)
		page, err := h.svc.ResolvePage(id, req)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, page)

	case "escalate":
		if r.Method != http.MethodPost {
			writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		page, err := h.svc.EscalatePage(id)
		if err != nil {
			handleOnCallError(w, err)
			return
		}
		writeOnCallJSON(w, http.StatusOK, page)

	default:
		writeOnCallError(w, http.StatusNotFound, "action not found")
	}
}

// HandleMetrics handles GET for /v2/oncall/metrics
func (h *OnCallHandler) HandleMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeOnCallError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

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

	metrics, err := h.svc.GetOnCallMetrics(wsID)
	if err != nil {
		handleOnCallError(w, err)
		return
	}
	writeOnCallJSON(w, http.StatusOK, metrics)
}

// ServeHTTP handles on-call HTTP requests
func (h *OnCallHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/oncall")
	path = strings.TrimPrefix(path, "/")

	// 1. Schedules: schedules, schedules/...
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
		if subPath == "overrides" || subPath == "overrides/" {
			h.HandleOverrides(w, r, id)
			return
		}
		h.HandleScheduleByID(w, r, id, subPath)
		return
	}

	// 2. Overrides by ID: overrides/{id}
	if strings.HasPrefix(path, "overrides/") {
		remainder := strings.TrimPrefix(path, "overrides/")
		id, err := strconv.ParseInt(strings.TrimSuffix(remainder, "/"), 10, 64)
		if err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid override id")
			return
		}
		h.HandleOverrideByID(w, r, id)
		return
	}

	// 3. Escalation Policies: policies, policies/default, policies/{id}
	if path == "policies/default" || path == "policies/default/" {
		h.HandleDefaultPolicy(w, r)
		return
	}
	if path == "policies" || path == "policies/" {
		h.HandlePolicies(w, r)
		return
	}
	if strings.HasPrefix(path, "policies/") {
		remainder := strings.TrimPrefix(path, "policies/")
		parts := strings.SplitN(remainder, "/", 2)
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid policy id")
			return
		}
		subPath := ""
		if len(parts) > 1 {
			subPath = parts[1]
		}
		h.HandlePolicyByID(w, r, id, subPath)
		return
	}

	// 4. Alert Paging: pages, pages/{id}, pages/{id}/acknowledge, pages/{id}/resolve, pages/{id}/escalate
	if path == "pages" || path == "pages/" {
		h.HandlePages(w, r)
		return
	}
	if strings.HasPrefix(path, "pages/") {
		remainder := strings.TrimPrefix(path, "pages/")
		parts := strings.SplitN(remainder, "/", 2)
		id, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			writeOnCallError(w, http.StatusBadRequest, "invalid page id")
			return
		}
		action := ""
		if len(parts) > 1 {
			action = strings.TrimSuffix(parts[1], "/")
		}
		h.HandlePageByID(w, r, id, action)
		return
	}

	// 5. Metrics: metrics
	if path == "metrics" || path == "metrics/" {
		h.HandleMetrics(w, r)
		return
	}

	writeOnCallError(w, http.StatusNotFound, "endpoint not found")
}
