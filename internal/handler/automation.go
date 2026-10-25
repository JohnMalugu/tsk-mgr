package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
