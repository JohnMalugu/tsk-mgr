package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// AutomationHandler handles HTTP endpoints for task automation rules and execution logs
type AutomationHandler struct {
	svc service.AutomationService
}

var GlobalAutomationHandler *AutomationHandler

// InitAutomationHandler initializes the global automation handler singleton
func InitAutomationHandler(svc service.AutomationService) {
	GlobalAutomationHandler = &AutomationHandler{svc: svc}
}

// NewAutomationHandler constructs a new AutomationHandler instance
func NewAutomationHandler(svc service.AutomationService) *AutomationHandler {
	return &AutomationHandler{svc: svc}
}

func getAutomationUser(r *http.Request) string {
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

// HandleCreateRule registers a new automation rule
func (h *AutomationHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	userID := getAutomationUser(r)

	var req model.CreateAutomationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
		return
	}

	rule, err := h.svc.CreateRule(userID, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrWorkspaceRequired),
			errors.Is(err, service.ErrRuleNameRequired),
			errors.Is(err, service.ErrInvalidTriggerType),
			errors.Is(err, service.ErrInvalidActionType),
			errors.Is(err, service.ErrInvalidConditionField),
			errors.Is(err, service.ErrInvalidConditionOperator),
			errors.Is(err, service.ErrEmptyActions):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(rule)
}

// HandleGetRule retrieves an automation rule by ID
func (h *AutomationHandler) HandleGetRule(w http.ResponseWriter, r *http.Request, id int64) {
	rule, err := h.svc.GetRule(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrAutomationRuleNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rule)
}

// HandleListRules retrieves automation rules for a workspace
func (h *AutomationHandler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "workspace_id query parameter is required"})
		return
	}

	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid workspace_id"})
		return
	}

	var activeOnly *bool
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		act := activeStr == "true" || activeStr == "1"
		activeOnly = &act
	}

	triggerType := r.URL.Query().Get("trigger_type")

	rules, err := h.svc.ListRules(wsID, activeOnly, triggerType)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if rules == nil {
		rules = []*model.AutomationRule{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rules)
}

// HandleUpdateRule updates an existing automation rule
func (h *AutomationHandler) HandleUpdateRule(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateAutomationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
		return
	}

	rule, err := h.svc.UpdateRule(id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrAutomationRuleNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrRuleNameRequired),
			errors.Is(err, service.ErrInvalidTriggerType),
			errors.Is(err, service.ErrInvalidActionType),
			errors.Is(err, service.ErrInvalidConditionField),
			errors.Is(err, service.ErrInvalidConditionOperator),
			errors.Is(err, service.ErrEmptyActions):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(rule)
}

// HandleDeleteRule removes an automation rule
func (h *AutomationHandler) HandleDeleteRule(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteRule(id); err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrAutomationRuleNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Automation rule deleted successfully"})
}

// HandleToggleRule toggles the active state of an automation rule
func (h *AutomationHandler) HandleToggleRule(w http.ResponseWriter, r *http.Request, id int64) {
	isActive, err := h.svc.ToggleRule(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrAutomationRuleNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"id":        id,
		"is_active": isActive,
	})
}

// HandleDuplicateRule duplicates an automation rule with an optional new name
func (h *AutomationHandler) HandleDuplicateRule(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getAutomationUser(r)

	var req model.DuplicateRuleRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	cloned, err := h.svc.DuplicateRule(userID, id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case errors.Is(err, service.ErrAutomationRuleNotFound):
			w.WriteHeader(http.StatusNotFound)
		case errors.Is(err, service.ErrRuleNameRequired):
			w.WriteHeader(http.StatusBadRequest)
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(cloned)
}

// HandleGetMetrics retrieves automation analytics and execution aggregates for a workspace
func (h *AutomationHandler) HandleGetMetrics(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "workspace_id query parameter is required"})
		return
	}

	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid workspace_id"})
		return
	}

	metrics, err := h.svc.GetMetrics(wsID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(metrics)
}

// HandleDryRunRule tests conditions and provides action previews without executing mutations
func (h *AutomationHandler) HandleDryRunRule(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.AutomationDryRunRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	result, err := h.svc.DryRunRule(id, req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrAutomationRuleNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(result)
}

// HandleListLogs queries execution logs across a workspace with optional status and rule filters
func (h *AutomationHandler) HandleListLogs(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "workspace_id query parameter is required"})
		return
	}

	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid workspace_id"})
		return
	}

	filter := model.AutomationLogFilter{
		WorkspaceID: wsID,
		Status:      r.URL.Query().Get("status"),
	}

	if ruleIDStr := r.URL.Query().Get("rule_id"); ruleIDStr != "" {
		if rid, err := strconv.ParseInt(ruleIDStr, 10, 64); err == nil && rid > 0 {
			filter.RuleID = &rid
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if lim, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = lim
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if off, err := strconv.Atoi(offsetStr); err == nil {
			filter.Offset = off
		}
	}

	logs, err := h.svc.ListLogs(filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*model.AutomationExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(logs)
}

// HandleGetRuleLogs queries execution logs specifically for one automation rule
func (h *AutomationHandler) HandleGetRuleLogs(w http.ResponseWriter, r *http.Request, id int64) {
	rule, err := h.svc.GetRule(id)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrAutomationRuleNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	filter := model.AutomationLogFilter{
		WorkspaceID: rule.WorkspaceID,
		RuleID:      &id,
		Limit:       50,
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if lim, err := strconv.Atoi(limitStr); err == nil {
			filter.Limit = lim
		}
	}

	logs, err := h.svc.ListLogs(filter)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*model.AutomationExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(logs)
}

// HandleDispatchEvent triggers manual event evaluation against workspace automation rules
func (h *AutomationHandler) HandleDispatchEvent(w http.ResponseWriter, r *http.Request) {
	var event model.TriggerEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body: " + err.Error()})
		return
	}

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.UserID == "" {
		event.UserID = getAutomationUser(r)
	}

	logs, err := h.svc.DispatchEvent(event)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if errors.Is(err, service.ErrWorkspaceRequired) {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	if logs == nil {
		logs = []*model.AutomationExecutionLog{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(logs)
}
