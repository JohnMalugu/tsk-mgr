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

var (
	_ = strconv.Atoi
	_ = strings.TrimPrefix
)

// EventHandler exposes HTTP endpoints for ITIL Monitoring & Event Management
type EventHandler struct {
	svc service.EventService
}

// GlobalEventHandler is the singleton handler instance
var GlobalEventHandler *EventHandler

// InitEventHandler initializes the global EventHandler singleton
func InitEventHandler(svc service.EventService) {
	GlobalEventHandler = &EventHandler{svc: svc}
}

// NewEventHandler creates a new EventHandler instance
func NewEventHandler(svc service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func getEventWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func writeEventJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeEventError(w http.ResponseWriter, status int, msg string) {
	writeEventJSON(w, status, map[string]string{"error": msg})
}

func handleEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEventRecordNotFound):
		writeEventError(w, http.StatusNotFound, "event record not found")
	case errors.Is(err, service.ErrCorrelatedAlertNotFound):
		writeEventError(w, http.StatusNotFound, "correlated alert not found")
	case errors.Is(err, service.ErrCorrelatedAlertCodeConflict):
		writeEventError(w, http.StatusConflict, "correlated alert with this code already exists in workspace")
	case errors.Is(err, service.ErrCorrelationRuleNotFound):
		writeEventError(w, http.StatusNotFound, "correlation rule not found")
	case errors.Is(err, service.ErrEventSuppressionNotFound):
		writeEventError(w, http.StatusNotFound, "event suppression not found")
	case errors.Is(err, service.ErrEventActionNotFound):
		writeEventError(w, http.StatusNotFound, "event action not found")
	default:
		writeEventError(w, http.StatusBadRequest, err.Error())
	}
}

// IngestEvent handles POST /v2/events/raw
func (h *EventHandler) IngestEvent(w http.ResponseWriter, r *http.Request) {
	var req model.IngestEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getEventWorkspaceID(r)
	record, alert, err := h.svc.IngestEvent(r.Context(), workspaceID, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusCreated, map[string]interface{}{
		"event": record,
		"alert": alert,
	})
}

// ListEvents handles GET /v2/events/raw
func (h *EventHandler) ListEvents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.EventRecordFilter{
		WorkspaceID: getEventWorkspaceID(r),
		Source:      q.Get("source"),
		Severity:    q.Get("severity"),
		CIID:        q.Get("ci_id"),
		Resource:    q.Get("resource"),
		Limit:       limit,
		Offset:      offset,
	}

	events, total, err := h.svc.ListEvents(r.Context(), filter)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetEventByID handles GET /v2/events/raw/{id}
func (h *EventHandler) GetEventByID(w http.ResponseWriter, r *http.Request, id string) {
	event, err := h.svc.GetEventByID(r.Context(), id)
	if err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, event)
}

// ListAlerts handles GET /v2/events/alerts
func (h *EventHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.CorrelatedAlertFilter{
		WorkspaceID: getEventWorkspaceID(r),
		Status:      q.Get("status"),
		Severity:    q.Get("severity"),
		Priority:    q.Get("priority"),
		CIID:        q.Get("ci_id"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	alerts, total, err := h.svc.ListAlerts(r.Context(), filter)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, map[string]interface{}{
		"alerts": alerts,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetAlertByID handles GET /v2/events/alerts/{id}
func (h *EventHandler) GetAlertByID(w http.ResponseWriter, r *http.Request, id string) {
	alert, err := h.svc.GetAlertByID(r.Context(), id)
	if err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, alert)
}

// UpdateAlertStatus handles PUT /v2/events/alerts/{id}/status
func (h *EventHandler) UpdateAlertStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateAlertStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	alert, err := h.svc.UpdateAlertStatus(r.Context(), id, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, alert)
}

// ListRules handles GET /v2/events/rules
func (h *EventHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.CorrelationRuleFilter{
		WorkspaceID: getEventWorkspaceID(r),
		Limit:       limit,
		Offset:      offset,
	}
	if isActiveStr := q.Get("is_active"); isActiveStr != "" {
		active := isActiveStr == "true" || isActiveStr == "1"
		filter.IsActive = &active
	}

	rules, total, err := h.svc.ListRules(r.Context(), filter)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, map[string]interface{}{
		"rules":  rules,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateRule handles POST /v2/events/rules
func (h *EventHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCorrelationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getEventWorkspaceID(r)
	rule, err := h.svc.CreateRule(r.Context(), workspaceID, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusCreated, rule)
}

// GetRuleByID handles GET /v2/events/rules/{id}
func (h *EventHandler) GetRuleByID(w http.ResponseWriter, r *http.Request, id string) {
	rule, err := h.svc.GetRuleByID(r.Context(), id)
	if err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, rule)
}

// UpdateRule handles PUT /v2/events/rules/{id}
func (h *EventHandler) UpdateRule(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateCorrelationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rule, err := h.svc.UpdateRule(r.Context(), id, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, rule)
}

// DeleteRule handles DELETE /v2/events/rules/{id}
func (h *EventHandler) DeleteRule(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteRule(r.Context(), id); err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, map[string]string{"message": "correlation rule deleted"})
}

// ListSuppressions handles GET /v2/events/suppressions
func (h *EventHandler) ListSuppressions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.EventSuppressionFilter{
		WorkspaceID: getEventWorkspaceID(r),
		CIID:        q.Get("ci_id"),
		Limit:       limit,
		Offset:      offset,
	}
	if isActiveStr := q.Get("is_active"); isActiveStr != "" {
		active := isActiveStr == "true" || isActiveStr == "1"
		filter.IsActive = &active
	}

	suppressions, total, err := h.svc.ListSuppressions(r.Context(), filter)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, map[string]interface{}{
		"suppressions": suppressions,
		"total":        total,
		"limit":        filter.Limit,
		"offset":       filter.Offset,
	})
}

// CreateSuppression handles POST /v2/events/suppressions
func (h *EventHandler) CreateSuppression(w http.ResponseWriter, r *http.Request) {
	var req model.CreateEventSuppressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getEventWorkspaceID(r)
	suppression, err := h.svc.CreateSuppression(r.Context(), workspaceID, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusCreated, suppression)
}

// GetSuppressionByID handles GET /v2/events/suppressions/{id}
func (h *EventHandler) GetSuppressionByID(w http.ResponseWriter, r *http.Request, id string) {
	suppression, err := h.svc.GetSuppressionByID(r.Context(), id)
	if err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, suppression)
}

// DeleteSuppression handles DELETE /v2/events/suppressions/{id}
func (h *EventHandler) DeleteSuppression(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteSuppression(r.Context(), id); err != nil {
		handleEventError(w, err)
		return
	}
	writeEventJSON(w, http.StatusOK, map[string]string{"message": "event suppression deleted"})
}

// ListActions handles GET /v2/events/actions
func (h *EventHandler) ListActions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.EventActionFilter{
		WorkspaceID: getEventWorkspaceID(r),
		AlertID:     q.Get("alert_id"),
		ActionType:  q.Get("action_type"),
		Status:      q.Get("status"),
		Limit:       limit,
		Offset:      offset,
	}

	actions, total, err := h.svc.ListActions(r.Context(), filter)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusOK, map[string]interface{}{
		"actions": actions,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// TriggerAction handles POST /v2/events/actions
func (h *EventHandler) TriggerAction(w http.ResponseWriter, r *http.Request) {
	var req model.TriggerEventActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeEventError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	workspaceID := getEventWorkspaceID(r)
	action, err := h.svc.TriggerAction(r.Context(), workspaceID, req)
	if err != nil {
		handleEventError(w, err)
		return
	}

	writeEventJSON(w, http.StatusCreated, action)
}

// Analytics stub
func (h *EventHandler) GetEventAnalytics(w http.ResponseWriter, r *http.Request) {}

// ServeHTTP stub
func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeEventError(w, http.StatusNotFound, "route not found")
}
