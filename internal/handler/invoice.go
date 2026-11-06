package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/JohnMalugu/tsk-mgr-api/internal/model"
	"github.com/JohnMalugu/tsk-mgr-api/internal/repository"
	"github.com/JohnMalugu/tsk-mgr-api/internal/service"
)

// InvoiceHandler exposes HTTP endpoints for client billing, invoices, line items, and payments
type InvoiceHandler struct {
	svc service.InvoiceService
}

var GlobalInvoiceHandler *InvoiceHandler

// InitInvoiceHandler initializes the global invoice handler
func InitInvoiceHandler(svc service.InvoiceService) {
	GlobalInvoiceHandler = &InvoiceHandler{svc: svc}
}

// NewInvoiceHandler creates a new InvoiceHandler instance
func NewInvoiceHandler(svc service.InvoiceService) *InvoiceHandler {
	return &InvoiceHandler{svc: svc}
}

func getInvoiceUser(r *http.Request) string {
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

func writeInvoiceJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeInvoiceError(w http.ResponseWriter, status int, msg string) {
	writeInvoiceJSON(w, status, map[string]string{"error": msg})
}

func handleInvoiceServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrClientNotFound), errors.Is(err, repository.ErrClientNotFound):
		writeInvoiceError(w, http.StatusNotFound, "client not found")
	case errors.Is(err, service.ErrInvoiceNotFound), errors.Is(err, repository.ErrInvoiceNotFound):
		writeInvoiceError(w, http.StatusNotFound, "invoice not found")
	case errors.Is(err, service.ErrInvoiceItemNotFound), errors.Is(err, repository.ErrInvoiceItemNotFound):
		writeInvoiceError(w, http.StatusNotFound, "invoice item not found")
	case errors.Is(err, service.ErrWorkspaceNotFound):
		writeInvoiceError(w, http.StatusNotFound, "workspace not found")
	case errors.Is(err, service.ErrWorkspaceRequired):
		writeInvoiceError(w, http.StatusBadRequest, "workspace_id is required")
	case errors.Is(err, service.ErrProjectNotFound):
		writeInvoiceError(w, http.StatusNotFound, "project not found")
	case errors.Is(err, service.ErrTimesheetNotFound):
		writeInvoiceError(w, http.StatusNotFound, "timesheet not found")
	case errors.Is(err, service.ErrTimesheetNotApproved):
		writeInvoiceError(w, http.StatusBadRequest, "timesheet must be in approved status to generate an invoice")
	case errors.Is(err, service.ErrInvoiceNotEditable):
		writeInvoiceError(w, http.StatusBadRequest, "invoice cannot be edited in current status")
	case errors.Is(err, service.ErrInvalidInvoiceDates):
		writeInvoiceError(w, http.StatusBadRequest, "invalid invoice dates: due date must be on or after issue date")
	case errors.Is(err, service.ErrInvalidPaymentAmount):
		writeInvoiceError(w, http.StatusBadRequest, "payment amount must be greater than zero and cannot exceed balance due")
	case errors.Is(err, service.ErrInvalidItemQuantity):
		writeInvoiceError(w, http.StatusBadRequest, "item quantity must be greater than zero")
	case errors.Is(err, service.ErrInvalidItemUnitPrice):
		writeInvoiceError(w, http.StatusBadRequest, "unit price must be non-negative")
	case errors.Is(err, service.ErrItemDescriptionRequired):
		writeInvoiceError(w, http.StatusBadRequest, "item description is required")
	case errors.Is(err, service.ErrInvalidInvoiceStatus):
		writeInvoiceError(w, http.StatusBadRequest, "invalid invoice status")
	case errors.Is(err, model.ErrClientNameRequired):
		writeInvoiceError(w, http.StatusBadRequest, "client name is required")
	case errors.Is(err, model.ErrInvalidClientEmail):
		writeInvoiceError(w, http.StatusBadRequest, "invalid client email format")
	case errors.Is(err, service.ErrUnauthorizedInvoice):
		writeInvoiceError(w, http.StatusForbidden, "unauthorized to perform action on invoice")
	default:
		writeInvoiceError(w, http.StatusInternalServerError, err.Error())
	}
}

// CreateClient handles POST /v2/invoices/clients
func (h *InvoiceHandler) CreateClient(w http.ResponseWriter, r *http.Request) {
	var req model.CreateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	client, err := h.svc.CreateClient(userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusCreated, client)
}

// GetClient handles GET /v2/invoices/clients/{id}
func (h *InvoiceHandler) GetClient(w http.ResponseWriter, r *http.Request, id int64) {
	client, err := h.svc.GetClient(id)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, client)
}

// ListClients handles GET /v2/invoices/clients
func (h *InvoiceHandler) ListClients(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		writeInvoiceError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeInvoiceError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	clients, err := h.svc.ListClients(wsID)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, clients)
}

// UpdateClient handles PUT/PATCH /v2/invoices/clients/{id}
func (h *InvoiceHandler) UpdateClient(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateClientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	client, err := h.svc.UpdateClient(id, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, client)
}

// DeleteClient handles DELETE /v2/invoices/clients/{id}
func (h *InvoiceHandler) DeleteClient(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getInvoiceUser(r)
	if err := h.svc.DeleteClient(id, userID); err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, map[string]string{"message": "client deleted successfully"})
}

// CreateInvoice handles POST /v2/invoices
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	inv, err := h.svc.CreateInvoice(userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusCreated, inv)
}

// GetInvoice handles GET /v2/invoices/{id}
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request, id int64) {
	inv, err := h.svc.GetInvoice(id)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, inv)
}

// ListInvoices handles GET /v2/invoices
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		writeInvoiceError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeInvoiceError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	filter := model.InvoiceFilter{
		WorkspaceID: wsID,
	}

	if clientIDStr := r.URL.Query().Get("client_id"); clientIDStr != "" {
		if cID, err := strconv.ParseInt(clientIDStr, 10, 64); err == nil && cID > 0 {
			filter.ClientID = &cID
		}
	}
	if projIDStr := r.URL.Query().Get("project_id"); projIDStr != "" {
		if pID, err := strconv.ParseInt(projIDStr, 10, 64); err == nil && pID > 0 {
			filter.ProjectID = &pID
		}
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if startDate := r.URL.Query().Get("start_date"); startDate != "" {
		filter.StartDate = &startDate
	}
	if endDate := r.URL.Query().Get("end_date"); endDate != "" {
		filter.EndDate = &endDate
	}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	invoices, err := h.svc.ListInvoices(filter)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, invoices)
}

// UpdateInvoice handles PUT/PATCH /v2/invoices/{id}
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateInvoiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	inv, err := h.svc.UpdateInvoice(id, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, inv)
}

// DeleteInvoice handles DELETE /v2/invoices/{id}
func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request, id int64) {
	userID := getInvoiceUser(r)
	if err := h.svc.DeleteInvoice(id, userID); err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, map[string]string{"message": "invoice deleted successfully"})
}

// ChangeStatus handles POST/PUT /v2/invoices/{id}/status
func (h *InvoiceHandler) ChangeStatus(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.ChangeInvoiceStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	inv, err := h.svc.ChangeStatus(id, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, inv)
}

// AddItem handles POST /v2/invoices/{id}/items
func (h *InvoiceHandler) AddItem(w http.ResponseWriter, r *http.Request, invoiceID int64) {
	var req model.InvoiceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	item, err := h.svc.AddItem(invoiceID, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusCreated, item)
}

// UpdateItem handles PUT/PATCH /v2/invoices/{id}/items/{itemId}
func (h *InvoiceHandler) UpdateItem(w http.ResponseWriter, r *http.Request, invoiceID, itemID int64) {
	var req model.InvoiceItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	item, err := h.svc.UpdateItem(invoiceID, itemID, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, item)
}

// DeleteItem handles DELETE /v2/invoices/{id}/items/{itemId}
func (h *InvoiceHandler) DeleteItem(w http.ResponseWriter, r *http.Request, invoiceID, itemID int64) {
	userID := getInvoiceUser(r)
	if err := h.svc.DeleteItem(invoiceID, itemID, userID); err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, map[string]string{"message": "item deleted successfully"})
}

// ListItems handles GET /v2/invoices/{id}/items
func (h *InvoiceHandler) ListItems(w http.ResponseWriter, r *http.Request, invoiceID int64) {
	items, err := h.svc.ListItems(invoiceID)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, items)
}

// GenerateFromTimesheet handles POST /v2/invoices/from-timesheet
func (h *InvoiceHandler) GenerateFromTimesheet(w http.ResponseWriter, r *http.Request) {
	var req model.GenerateInvoiceFromTimesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	inv, err := h.svc.GenerateFromTimesheet(userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusCreated, inv)
}

// RecordPayment handles POST /v2/invoices/{id}/payments
func (h *InvoiceHandler) RecordPayment(w http.ResponseWriter, r *http.Request, invoiceID int64) {
	var req model.RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeInvoiceError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := getInvoiceUser(r)

	payment, err := h.svc.RecordPayment(invoiceID, userID, req)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusCreated, payment)
}

// ListPayments handles GET /v2/invoices/{id}/payments
func (h *InvoiceHandler) ListPayments(w http.ResponseWriter, r *http.Request, invoiceID int64) {
	payments, err := h.svc.ListPayments(invoiceID)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, payments)
}

// GetInvoiceMetrics handles GET /v2/invoices/metrics
func (h *InvoiceHandler) GetInvoiceMetrics(w http.ResponseWriter, r *http.Request) {
	wsIDStr := r.URL.Query().Get("workspace_id")
	if wsIDStr == "" {
		writeInvoiceError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsIDStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeInvoiceError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	metrics, err := h.svc.GetInvoiceMetrics(wsID)
	if err != nil {
		handleInvoiceServiceError(w, err)
		return
	}
	writeInvoiceJSON(w, http.StatusOK, metrics)
}

// ServeHTTP dispatches incoming requests to the appropriate InvoiceHandler endpoint
func (h *InvoiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/invoices")
	path = strings.TrimPrefix(path, "/")

	if path == "" {
		switch r.Method {
		case http.MethodGet:
			h.ListInvoices(w, r)
		case http.MethodPost:
			h.CreateInvoice(w, r)
		default:
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if path == "metrics" || path == "metrics/" {
		if r.Method != http.MethodGet {
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetInvoiceMetrics(w, r)
		return
	}

	if path == "from-timesheet" || path == "from-timesheet/" {
		if r.Method != http.MethodPost {
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GenerateFromTimesheet(w, r)
		return
	}

	if path == "clients" || path == "clients/" {
		switch r.Method {
		case http.MethodGet:
			h.ListClients(w, r)
		case http.MethodPost:
			h.CreateClient(w, r)
		default:
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if strings.HasPrefix(path, "clients/") {
		clientPart := strings.TrimPrefix(path, "clients/")
		clientParts := strings.Split(clientPart, "/")
		clientID, err := strconv.ParseInt(clientParts[0], 10, 64)
		if err != nil || clientID <= 0 {
			writeInvoiceError(w, http.StatusBadRequest, "invalid client id")
			return
		}

		if len(clientParts) == 1 || (len(clientParts) == 2 && clientParts[1] == "") {
			switch r.Method {
			case http.MethodGet:
				h.GetClient(w, r, clientID)
			case http.MethodPut, http.MethodPatch:
				h.UpdateClient(w, r, clientID)
			case http.MethodDelete:
				h.DeleteClient(w, r, clientID)
			default:
				writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		writeInvoiceError(w, http.StatusNotFound, "not found")
		return
	}

	parts := strings.Split(path, "/")
	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id <= 0 {
		writeInvoiceError(w, http.StatusBadRequest, "invalid invoice id")
		return
	}

	if len(parts) == 1 || (len(parts) == 2 && parts[1] == "") {
		switch r.Method {
		case http.MethodGet:
			h.GetInvoice(w, r, id)
		case http.MethodPut, http.MethodPatch:
			h.UpdateInvoice(w, r, id)
		case http.MethodDelete:
			h.DeleteInvoice(w, r, id)
		default:
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	if len(parts) == 2 {
		switch parts[1] {
		case "status":
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
				return
			}
			h.ChangeStatus(w, r, id)
		case "items":
			switch r.Method {
			case http.MethodGet:
				h.ListItems(w, r, id)
			case http.MethodPost:
				h.AddItem(w, r, id)
			default:
				writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		case "payments":
			switch r.Method {
			case http.MethodGet:
				h.ListPayments(w, r, id)
			case http.MethodPost:
				h.RecordPayment(w, r, id)
			default:
				writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
		default:
			writeInvoiceError(w, http.StatusNotFound, "not found")
		}
		return
	}

	if len(parts) == 3 && parts[1] == "items" {
		itemID, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil || itemID <= 0 {
			writeInvoiceError(w, http.StatusBadRequest, "invalid invoice item id")
			return
		}
		switch r.Method {
		case http.MethodPut, http.MethodPatch:
			h.UpdateItem(w, r, id, itemID)
		case http.MethodDelete:
			h.DeleteItem(w, r, id, itemID)
		default:
			writeInvoiceError(w, http.StatusMethodNotAllowed, "method not allowed")
		}
		return
	}

	writeInvoiceError(w, http.StatusNotFound, "not found")
}
