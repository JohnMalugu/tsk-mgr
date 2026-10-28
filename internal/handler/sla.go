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

// SlaHandler handles HTTP endpoints for SLA policies, escalation rules, and compliance tracking
type SlaHandler struct {
	svc service.SlaService
}

var GlobalSlaHandler *SlaHandler

// InitSlaHandler initializes the global SLA handler singleton
func InitSlaHandler(svc service.SlaService) {
	GlobalSlaHandler = &SlaHandler{svc: svc}
}

// NewSlaHandler constructs a new SlaHandler instance
func NewSlaHandler(svc service.SlaService) *SlaHandler {
	return &SlaHandler{svc: svc}
}

func getSlaUser(r *http.Request) string {
	if uid, ok := r.Context().Value("userID").(string); ok && uid != "" {
		return uid
	}
	if uid := r.Header.Get("X-User-ID"); uid != "" {
		return uid
	}
	if uid := r.URL.Query().Get("user_id"); uid != "" {
		return uid
	}
	return "default_user"
}

func writeSlaJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeSlaError(w http.ResponseWriter, status int, msg string) {
	writeSlaJSON(w, status, map[string]string{"error": msg})
}

// HandleSlasRoot routes /v2/slas for policy creation (POST) and list policies (GET)
func (h *SlaHandler) HandleSlasRoot(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		h.HandleCreatePolicy(w, r)
	case http.MethodGet:
		h.HandleListPolicies(w, r)
	default:
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleCreatePolicy creates a new SLA policy
func (h *SlaHandler) HandleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	userID := getSlaUser(r)

	var req model.CreateSlaPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	policy, err := h.svc.CreatePolicy(userID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlaPolicyNameRequired),
			errors.Is(err, service.ErrSlaWorkspaceRequired),
			errors.Is(err, service.ErrInvalidTargetMinutes),
			errors.Is(err, service.ErrInvalidTriggerEvent),
			errors.Is(err, service.ErrInvalidSlaActionType),
			errors.Is(err, service.ErrInvalidEscalationRule):
			writeSlaError(w, http.StatusBadRequest, err.Error())
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to create SLA policy: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusCreated, policy)
}

// HandleListPolicies lists SLA policies with optional workspace and status filters
func (h *SlaHandler) HandleListPolicies(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.SlaFilter{}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = &wsID
		}
	}
	if actStr := q.Get("is_active"); actStr != "" {
		if act, err := strconv.ParseBool(actStr); err == nil {
			filter.IsActive = &act
		}
	}
	if pri := q.Get("priority"); pri != "" {
		filter.Priority = pri
	}

	policies, err := h.svc.ListPolicies(filter)
	if err != nil {
		writeSlaError(w, http.StatusInternalServerError, "Failed to list SLA policies: "+err.Error())
		return
	}

	if policies == nil {
		policies = []*model.SlaSummary{}
	}

	writeSlaJSON(w, http.StatusOK, policies)
}

// HandlePolicyByID routes /v2/slas/{id} for GET, PUT, and DELETE
func (h *SlaHandler) HandlePolicyByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/slas/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSlaError(w, http.StatusBadRequest, "Missing policy ID in path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid policy ID format")
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.HandleGetPolicy(w, r, id)
	case http.MethodPut, http.MethodPatch:
		h.HandleUpdatePolicy(w, r, id)
	case http.MethodDelete:
		h.HandleDeletePolicy(w, r, id)
	default:
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleGetPolicy retrieves an SLA policy by ID
func (h *SlaHandler) HandleGetPolicy(w http.ResponseWriter, r *http.Request, id int64) {
	policy, err := h.svc.GetPolicy(id)
	if err != nil {
		if errors.Is(err, service.ErrSlaPolicyNotFound) {
			writeSlaError(w, http.StatusNotFound, "SLA policy not found")
			return
		}
		writeSlaError(w, http.StatusInternalServerError, "Failed to retrieve SLA policy: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, policy)
}

// HandleUpdatePolicy updates an existing SLA policy
func (h *SlaHandler) HandleUpdatePolicy(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateSlaPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	policy, err := h.svc.UpdatePolicy(id, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrSlaPolicyNotFound):
			writeSlaError(w, http.StatusNotFound, "SLA policy not found")
		case errors.Is(err, service.ErrSlaPolicyNameRequired),
			errors.Is(err, service.ErrInvalidTargetMinutes):
			writeSlaError(w, http.StatusBadRequest, err.Error())
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to update SLA policy: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusOK, policy)
}

// HandleDeletePolicy removes an SLA policy by ID
func (h *SlaHandler) HandleDeletePolicy(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeletePolicy(id); err != nil {
		if errors.Is(err, service.ErrSlaPolicyNotFound) {
			writeSlaError(w, http.StatusNotFound, "SLA policy not found")
			return
		}
		writeSlaError(w, http.StatusInternalServerError, "Failed to delete SLA policy: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "SLA policy deleted successfully",
	})
}

// HandleTogglePolicy toggles active status of an SLA policy
func (h *SlaHandler) HandleTogglePolicy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v2/slas/")
	path = strings.TrimSuffix(path, "/toggle")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSlaError(w, http.StatusBadRequest, "Missing policy ID in path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid policy ID format")
		return
	}

	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	policy, err := h.svc.TogglePolicy(id, req.IsActive)
	if err != nil {
		if errors.Is(err, service.ErrSlaPolicyNotFound) {
			writeSlaError(w, http.StatusNotFound, "SLA policy not found")
			return
		}
		writeSlaError(w, http.StatusInternalServerError, "Failed to toggle SLA policy: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, policy)
}

// HandlePolicyEscalations handles GET and POST for /v2/slas/{id}/escalations
func (h *SlaHandler) HandlePolicyEscalations(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/slas/")
	path = strings.TrimSuffix(path, "/escalations")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSlaError(w, http.StatusBadRequest, "Missing policy ID in path")
		return
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid policy ID format")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rules, err := h.svc.GetEscalationRules(id)
		if err != nil {
			if errors.Is(err, service.ErrSlaPolicyNotFound) {
				writeSlaError(w, http.StatusNotFound, "SLA policy not found")
				return
			}
			writeSlaError(w, http.StatusInternalServerError, "Failed to get escalation rules: "+err.Error())
			return
		}
		if rules == nil {
			rules = []model.EscalationRule{}
		}
		writeSlaJSON(w, http.StatusOK, rules)

	case http.MethodPost, http.MethodPut:
		var req model.SetEscalationRulesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeSlaError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
			return
		}

		rules, err := h.svc.SetEscalationRules(id, req)
		if err != nil {
			switch {
			case errors.Is(err, service.ErrSlaPolicyNotFound):
				writeSlaError(w, http.StatusNotFound, "SLA policy not found")
			case errors.Is(err, service.ErrInvalidTriggerEvent),
				errors.Is(err, service.ErrInvalidSlaActionType),
				errors.Is(err, service.ErrInvalidEscalationRule):
				writeSlaError(w, http.StatusBadRequest, err.Error())
			default:
				writeSlaError(w, http.StatusInternalServerError, "Failed to set escalation rules: "+err.Error())
			}
			return
		}
		writeSlaJSON(w, http.StatusOK, rules)

	default:
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// HandleDeleteEscalationRule deletes a specific escalation rule via DELETE /v2/slas/escalations/{ruleId}
func (h *SlaHandler) HandleDeleteEscalationRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/v2/slas/escalations/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSlaError(w, http.StatusBadRequest, "Missing rule ID in path")
		return
	}

	ruleID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid rule ID format")
		return
	}

	if err := h.svc.DeleteEscalationRule(ruleID); err != nil {
		writeSlaError(w, http.StatusInternalServerError, "Failed to delete escalation rule: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Escalation rule deleted successfully",
	})
}

// HandleApplySla attaches or computes an SLA target for a task via POST /v2/slas/tasks/apply
func (h *SlaHandler) HandleApplySla(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.ApplySlaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	rec, err := h.svc.ApplySlaToTask(req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			writeSlaError(w, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrSlaPolicyNotFound):
			writeSlaError(w, http.StatusNotFound, "SLA policy not found for task")
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to apply SLA to task: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusCreated, rec)
}

// HandleTaskSla fetches SLA status for a task via GET /v2/slas/tasks/{taskId}
func (h *SlaHandler) HandleTaskSla(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/slas/tasks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		writeSlaError(w, http.StatusBadRequest, "Missing task ID in path")
		return
	}

	taskID, err := strconv.Atoi(parts[0])
	if err != nil {
		writeSlaError(w, http.StatusBadRequest, "Invalid task ID format")
		return
	}

	if len(parts) > 1 {
		switch parts[1] {
		case "pause":
			h.HandlePauseSla(w, r, taskID)
			return
		case "resume":
			h.HandleResumeSla(w, r, taskID)
			return
		case "evaluate":
			h.HandleEvaluateSla(w, r, taskID)
			return
		}
	}

	if r.Method != http.MethodGet {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	rec, err := h.svc.GetTaskSla(taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskSlaNotFound) {
			writeSlaError(w, http.StatusNotFound, "Task SLA record not found")
			return
		}
		writeSlaError(w, http.StatusInternalServerError, "Failed to get task SLA record: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, rec)
}

// HandlePauseSla pauses an active SLA countdown via POST /v2/slas/tasks/{taskId}/pause
func (h *SlaHandler) HandlePauseSla(w http.ResponseWriter, r *http.Request, taskID int) {
	if r.Method != http.MethodPost {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req model.PauseSlaRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	rec, err := h.svc.PauseSla(taskID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskSlaNotFound):
			writeSlaError(w, http.StatusNotFound, "Task SLA record not found")
		case errors.Is(err, service.ErrSlaAlreadyPaused):
			writeSlaError(w, http.StatusBadRequest, err.Error())
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to pause SLA: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusOK, rec)
}

// HandleResumeSla resumes a paused SLA countdown via POST /v2/slas/tasks/{taskId}/resume
func (h *SlaHandler) HandleResumeSla(w http.ResponseWriter, r *http.Request, taskID int) {
	if r.Method != http.MethodPost {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	rec, err := h.svc.ResumeSla(taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskSlaNotFound):
			writeSlaError(w, http.StatusNotFound, "Task SLA record not found")
		case errors.Is(err, service.ErrSlaNotPaused):
			writeSlaError(w, http.StatusBadRequest, err.Error())
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to resume SLA: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusOK, rec)
}

// HandleEvaluateSla triggers on-demand evaluation of SLA stage and escalations via POST /v2/slas/tasks/{taskId}/evaluate
func (h *SlaHandler) HandleEvaluateSla(w http.ResponseWriter, r *http.Request, taskID int) {
	if r.Method != http.MethodPost {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	rec, err := h.svc.EvaluateTaskSla(taskID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			writeSlaError(w, http.StatusNotFound, "Task not found")
		case errors.Is(err, service.ErrTaskSlaNotFound):
			writeSlaError(w, http.StatusNotFound, "Task SLA record not found")
		default:
			writeSlaError(w, http.StatusInternalServerError, "Failed to evaluate SLA: "+err.Error())
		}
		return
	}

	writeSlaJSON(w, http.StatusOK, rec)
}

// HandleListBreaches retrieves historical breach records via GET /v2/slas/breaches
func (h *SlaHandler) HandleListBreaches(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	q := r.URL.Query()
	filter := model.BreachFilter{}

	if wsStr := q.Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil {
			filter.WorkspaceID = &wsID
		}
	}
	if taskStr := q.Get("task_id"); taskStr != "" {
		if tID, err := strconv.Atoi(taskStr); err == nil {
			filter.TaskID = &tID
		}
	}
	if bt := q.Get("breach_type"); bt != "" {
		filter.BreachType = bt
	}
	if limStr := q.Get("limit"); limStr != "" {
		if lim, err := strconv.Atoi(limStr); err == nil {
			filter.Limit = lim
		}
	}
	if offStr := q.Get("offset"); offStr != "" {
		if off, err := strconv.Atoi(offStr); err == nil {
			filter.Offset = off
		}
	}

	breaches, err := h.svc.ListBreaches(filter)
	if err != nil {
		writeSlaError(w, http.StatusInternalServerError, "Failed to list SLA breaches: "+err.Error())
		return
	}

	if breaches == nil {
		breaches = []*model.SlaBreachLog{}
	}

	writeSlaJSON(w, http.StatusOK, breaches)
}

// HandleSlaMetrics aggregates SLA metrics for a workspace via GET /v2/slas/metrics
func (h *SlaHandler) HandleSlaMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeSlaError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var workspaceID int64 = 1
	if wsStr := r.URL.Query().Get("workspace_id"); wsStr != "" {
		if wsID, err := strconv.ParseInt(wsStr, 10, 64); err == nil && wsID > 0 {
			workspaceID = wsID
		}
	}

	metrics, err := h.svc.GetSlaMetrics(workspaceID)
	if err != nil {
		writeSlaError(w, http.StatusInternalServerError, "Failed to retrieve SLA metrics: "+err.Error())
		return
	}

	writeSlaJSON(w, http.StatusOK, metrics)
}
