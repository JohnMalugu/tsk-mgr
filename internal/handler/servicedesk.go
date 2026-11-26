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

// ServiceDeskHandler exposes HTTP endpoints for ITIL Service Desk Management
type ServiceDeskHandler struct {
	svc service.ServiceDeskService
}

// GlobalServiceDeskHandler is the singleton handler instance
var GlobalServiceDeskHandler *ServiceDeskHandler

// InitServiceDeskHandler initializes the global ServiceDeskHandler singleton
func InitServiceDeskHandler(svc service.ServiceDeskService) {
	GlobalServiceDeskHandler = &ServiceDeskHandler{svc: svc}
}

// NewServiceDeskHandler creates a new ServiceDeskHandler instance
func NewServiceDeskHandler(svc service.ServiceDeskService) *ServiceDeskHandler {
	return &ServiceDeskHandler{svc: svc}
}

func getSDWorkspaceID(r *http.Request) string {
	if ws := r.URL.Query().Get("workspace_id"); ws != "" {
		return ws
	}
	if ws := r.Header.Get("X-Workspace-ID"); ws != "" {
		return ws
	}
	return "ws_1"
}

func getSDUser(r *http.Request) string {
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

func writeSDJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeSDError(w http.ResponseWriter, status int, msg string) {
	writeSDJSON(w, status, map[string]string{"error": msg})
}

func handleSDError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrSDQueueNotFound):
		writeSDError(w, http.StatusNotFound, "service desk queue not found")
	case errors.Is(err, service.ErrSDQueueCodeConflict):
		writeSDError(w, http.StatusConflict, "service desk queue with this code already exists in workspace")
	case errors.Is(err, service.ErrSDTicketNotFound):
		writeSDError(w, http.StatusNotFound, "service desk ticket not found")
	case errors.Is(err, service.ErrSDTicketNumberConflict):
		writeSDError(w, http.StatusConflict, "service desk ticket with this number already exists in workspace")
	case errors.Is(err, service.ErrSDInteractionNotFound):
		writeSDError(w, http.StatusNotFound, "ticket interaction not found")
	case errors.Is(err, service.ErrCSATSurveyNotFound):
		writeSDError(w, http.StatusNotFound, "CSAT survey not found")
	default:
		writeSDError(w, http.StatusBadRequest, err.Error())
	}
}

// ListQueues handles GET /v2/servicedesk/queues
func (h *ServiceDeskHandler) ListQueues(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.SDQueueFilter{
		WorkspaceID: getSDWorkspaceID(r),
		Tier:        q.Get("tier"),
		Limit:       limit,
		Offset:      offset,
	}
	if isActiveStr := q.Get("is_active"); isActiveStr != "" {
		active := isActiveStr == "true" || isActiveStr == "1"
		filter.IsActive = &active
	}

	queues, total, err := h.svc.ListQueues(r.Context(), filter)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, map[string]interface{}{
		"queues": queues,
		"total":  total,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// CreateQueue handles POST /v2/servicedesk/queues
func (h *ServiceDeskHandler) CreateQueue(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSDQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSDError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	queue, err := h.svc.CreateQueue(r.Context(), getSDWorkspaceID(r), req)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusCreated, queue)
}

// GetQueueByID handles GET /v2/servicedesk/queues/{id}
func (h *ServiceDeskHandler) GetQueueByID(w http.ResponseWriter, r *http.Request, id string) {
	queue, err := h.svc.GetQueueByID(r.Context(), id)
	if err != nil {
		handleSDError(w, err)
		return
	}
	writeSDJSON(w, http.StatusOK, queue)
}

// UpdateQueue handles PUT /v2/servicedesk/queues/{id}
func (h *ServiceDeskHandler) UpdateQueue(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateSDQueueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSDError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	queue, err := h.svc.UpdateQueue(r.Context(), id, req)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, queue)
}

// DeleteQueue handles DELETE /v2/servicedesk/queues/{id}
func (h *ServiceDeskHandler) DeleteQueue(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteQueue(r.Context(), id); err != nil {
		handleSDError(w, err)
		return
	}
	writeSDJSON(w, http.StatusOK, map[string]string{"message": "service desk queue deleted"})
}

// ListTickets handles GET /v2/servicedesk/tickets
func (h *ServiceDeskHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))

	filter := model.SDTicketFilter{
		WorkspaceID: getSDWorkspaceID(r),
		QueueID:     q.Get("queue_id"),
		RequesterID: q.Get("requester_id"),
		AssigneeID:  q.Get("assignee_id"),
		Channel:     q.Get("channel"),
		TicketType:  q.Get("ticket_type"),
		Priority:    q.Get("priority"),
		Status:      q.Get("status"),
		Search:      q.Get("search"),
		Limit:       limit,
		Offset:      offset,
	}

	tickets, total, err := h.svc.ListTickets(r.Context(), filter)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, map[string]interface{}{
		"tickets": tickets,
		"total":   total,
		"limit":   filter.Limit,
		"offset":  filter.Offset,
	})
}

// CreateTicket handles POST /v2/servicedesk/tickets
func (h *ServiceDeskHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSDTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSDError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.RequesterID == "" {
		req.RequesterID = getSDUser(r)
	}

	ticket, err := h.svc.CreateTicket(r.Context(), getSDWorkspaceID(r), req)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusCreated, ticket)
}

// GetTicketByID handles GET /v2/servicedesk/tickets/{id}
func (h *ServiceDeskHandler) GetTicketByID(w http.ResponseWriter, r *http.Request, id string) {
	ticket, err := h.svc.GetTicketByID(r.Context(), id)
	if err != nil {
		handleSDError(w, err)
		return
	}
	writeSDJSON(w, http.StatusOK, ticket)
}

// UpdateTicket handles PUT /v2/servicedesk/tickets/{id}
func (h *ServiceDeskHandler) UpdateTicket(w http.ResponseWriter, r *http.Request, id string) {
	var req model.UpdateSDTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeSDError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ticket, err := h.svc.UpdateTicket(r.Context(), id, req)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, ticket)
}

// ResolveTicket handles POST /v2/servicedesk/tickets/{id}/resolve
func (h *ServiceDeskHandler) ResolveTicket(w http.ResponseWriter, r *http.Request, id string) {
	var req model.ResolveSDTicketRequest
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	ticket, err := h.svc.ResolveTicket(r.Context(), id, req)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, ticket)
}

// CloseTicket handles POST /v2/servicedesk/tickets/{id}/close
func (h *ServiceDeskHandler) CloseTicket(w http.ResponseWriter, r *http.Request, id string) {
	ticket, err := h.svc.CloseTicket(r.Context(), id)
	if err != nil {
		handleSDError(w, err)
		return
	}

	writeSDJSON(w, http.StatusOK, ticket)
}

// DeleteTicket handles DELETE /v2/servicedesk/tickets/{id}
func (h *ServiceDeskHandler) DeleteTicket(w http.ResponseWriter, r *http.Request, id string) {
	if err := h.svc.DeleteTicket(r.Context(), id); err != nil {
		handleSDError(w, err)
		return
	}
	writeSDJSON(w, http.StatusOK, map[string]string{"message": "service desk ticket deleted"})
}

// Interactions & Surveys stubs
func (h *ServiceDeskHandler) ListInteractions(w http.ResponseWriter, r *http.Request) {}
func (h *ServiceDeskHandler) AddInteraction(w http.ResponseWriter, r *http.Request) {}
func (h *ServiceDeskHandler) ListSurveys(w http.ResponseWriter, r *http.Request) {}
func (h *ServiceDeskHandler) SubmitSurvey(w http.ResponseWriter, r *http.Request) {}
func (h *ServiceDeskHandler) GetServiceDeskAnalytics(w http.ResponseWriter, r *http.Request) {}

// ServeHTTP stub
func (h *ServiceDeskHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeSDError(w, http.StatusNotFound, "route not found")
}
