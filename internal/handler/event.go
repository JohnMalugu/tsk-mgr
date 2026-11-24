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

// Correlation Rules stubs
func (h *EventHandler) ListRules(w http.ResponseWriter, r *http.Request) {}
func (h *EventHandler) CreateRule(w http.ResponseWriter, r *http.Request) {}
func (h *EventHandler) GetRuleByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *EventHandler) UpdateRule(w http.ResponseWriter, r *http.Request, id string) {}
func (h *EventHandler) DeleteRule(w http.ResponseWriter, r *http.Request, id string) {}

// Suppression stubs
func (h *EventHandler) ListSuppressions(w http.ResponseWriter, r *http.Request) {}
func (h *EventHandler) CreateSuppression(w http.ResponseWriter, r *http.Request) {}
func (h *EventHandler) GetSuppressionByID(w http.ResponseWriter, r *http.Request, id string) {}
func (h *EventHandler) DeleteSuppression(w http.ResponseWriter, r *http.Request, id string) {}

// Actions stubs
func (h *EventHandler) ListActions(w http.ResponseWriter, r *http.Request) {}
func (h *EventHandler) TriggerAction(w http.ResponseWriter, r *http.Request) {}

// Analytics stub
func (h *EventHandler) GetEventAnalytics(w http.ResponseWriter, r *http.Request) {}

// ServeHTTP stub
func (h *EventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeEventError(w, http.StatusNotFound, "route not found")
}
