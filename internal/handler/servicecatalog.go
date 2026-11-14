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

// ServiceCatalogHandler exposes HTTP endpoints for the IT Service Catalog & Request Fulfillment System
type ServiceCatalogHandler struct {
	svc service.ServiceCatalogService
}

var GlobalServiceCatalogHandler *ServiceCatalogHandler

// InitServiceCatalogHandler initializes the singleton ServiceCatalog handler
func InitServiceCatalogHandler(svc service.ServiceCatalogService) {
	GlobalServiceCatalogHandler = &ServiceCatalogHandler{svc: svc}
}

// NewServiceCatalogHandler creates a new ServiceCatalogHandler instance
func NewServiceCatalogHandler(svc service.ServiceCatalogService) *ServiceCatalogHandler {
	return &ServiceCatalogHandler{svc: svc}
}

func getServiceCatalogUser(r *http.Request) string {
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

func writeServiceCatalogJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeServiceCatalogError(w http.ResponseWriter, status int, msg string) {
	writeServiceCatalogJSON(w, status, map[string]string{"error": msg})
}

func handleServiceCatalogError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrCatalogCategoryNotFound):
		writeServiceCatalogError(w, http.StatusNotFound, "service catalog category not found")
	case errors.Is(err, service.ErrCatalogCategoryNameReq):
		writeServiceCatalogError(w, http.StatusBadRequest, "service catalog category name is required")
	case errors.Is(err, service.ErrCatalogCategorySlugReq):
		writeServiceCatalogError(w, http.StatusBadRequest, "service catalog category slug is required")
	case errors.Is(err, service.ErrDuplicateCategorySlug):
		writeServiceCatalogError(w, http.StatusConflict, "service catalog category slug already in use")
	case errors.Is(err, service.ErrCatalogItemNotFound):
		writeServiceCatalogError(w, http.StatusNotFound, "service catalog item not found")
	case errors.Is(err, service.ErrCatalogItemNameReq):
		writeServiceCatalogError(w, http.StatusBadRequest, "service catalog item name is required")
	case errors.Is(err, service.ErrCatalogItemSlugReq):
		writeServiceCatalogError(w, http.StatusBadRequest, "service catalog item slug is required")
	case errors.Is(err, service.ErrDuplicateItemSlug):
		writeServiceCatalogError(w, http.StatusConflict, "service catalog item slug already in use")
	case errors.Is(err, service.ErrInvalidApprovalRole):
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid approval role: must be manager, admin, or security_officer")
	case errors.Is(err, service.ErrServiceRequestNotFound):
		writeServiceCatalogError(w, http.StatusNotFound, "service request not found")
	case errors.Is(err, service.ErrServiceRequestTitleReq):
		writeServiceCatalogError(w, http.StatusBadRequest, "service request title is required")
	case errors.Is(err, service.ErrInvalidRequestStatus):
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid service request status")
	case errors.Is(err, service.ErrInvalidRequestPriority):
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid service request priority")
	case errors.Is(err, service.ErrRequestApprovalNotFound):
		writeServiceCatalogError(w, http.StatusNotFound, "service request approval not found")
	case errors.Is(err, service.ErrInvalidApprovalStatus):
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid approval status: must be pending, approved, or rejected")
	case errors.Is(err, service.ErrRequestAlreadyDecided):
		writeServiceCatalogError(w, http.StatusConflict, "service request approval has already been decided")
	case errors.Is(err, service.ErrRequestCommentEmpty):
		writeServiceCatalogError(w, http.StatusBadRequest, "service request comment cannot be empty")
	default:
		writeServiceCatalogError(w, http.StatusBadRequest, err.Error())
	}
}

// CreateCategory handles POST /v2/servicecatalog/categories
func (h *ServiceCatalogHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCatalogCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getServiceCatalogUser(r)

	cat, err := h.svc.CreateCategory(userID, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusCreated, cat)
}

// GetCategory handles GET /v2/servicecatalog/categories/{id}
func (h *ServiceCatalogHandler) GetCategory(w http.ResponseWriter, r *http.Request, id int64) {
	cat, err := h.svc.GetCategory(id)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, cat)
}

// ListCategories handles GET /v2/servicecatalog/categories
func (h *ServiceCatalogHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.CatalogCategoryFilter{
		Search: q.Get("search"),
	}
	if ws := q.Get("workspace_id"); ws != "" {
		if wsID, err := strconv.ParseInt(ws, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if act := q.Get("is_active"); act != "" {
		isActive := (act == "true" || act == "1")
		filter.IsActive = &isActive
	}

	categories, err := h.svc.ListCategories(filter)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]interface{}{
		"categories": categories,
	})
}

// UpdateCategory handles PUT/PATCH /v2/servicecatalog/categories/{id}
func (h *ServiceCatalogHandler) UpdateCategory(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateCatalogCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	cat, err := h.svc.UpdateCategory(id, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, cat)
}

// DeleteCategory handles DELETE /v2/servicecatalog/categories/{id}
func (h *ServiceCatalogHandler) DeleteCategory(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteCategory(id); err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]string{"message": "category deleted successfully"})
}

// CreateItem handles POST /v2/servicecatalog/items
func (h *ServiceCatalogHandler) CreateItem(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCatalogItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getServiceCatalogUser(r)

	item, err := h.svc.CreateItem(userID, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusCreated, item)
}

// GetItem handles GET /v2/servicecatalog/items/{id}
func (h *ServiceCatalogHandler) GetItem(w http.ResponseWriter, r *http.Request, id int64) {
	item, err := h.svc.GetItem(id)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, item)
}

// ListItems handles GET /v2/servicecatalog/items
func (h *ServiceCatalogHandler) ListItems(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.CatalogItemFilter{
		Search: q.Get("search"),
	}
	if ws := q.Get("workspace_id"); ws != "" {
		if wsID, err := strconv.ParseInt(ws, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if cat := q.Get("category_id"); cat != "" {
		if catID, err := strconv.ParseInt(cat, 10, 64); err == nil {
			filter.CategoryID = &catID
		}
	}
	if reqAppr := q.Get("requires_approval"); reqAppr != "" {
		ra := (reqAppr == "true" || reqAppr == "1")
		filter.RequiresApproval = &ra
	}
	if act := q.Get("is_active"); act != "" {
		isActive := (act == "true" || act == "1")
		filter.IsActive = &isActive
	}
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			filter.Limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	items, total, err := h.svc.ListItems(filter)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// UpdateItem handles PUT/PATCH /v2/servicecatalog/items/{id}
func (h *ServiceCatalogHandler) UpdateItem(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateCatalogItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	item, err := h.svc.UpdateItem(id, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, item)
}

// DeleteItem handles DELETE /v2/servicecatalog/items/{id}
func (h *ServiceCatalogHandler) DeleteItem(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.svc.DeleteItem(id); err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]string{"message": "catalog item deleted successfully"})
}

// SubmitRequest handles POST /v2/servicecatalog/requests
func (h *ServiceCatalogHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	var req model.SubmitServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getServiceCatalogUser(r)

	serviceReq, err := h.svc.SubmitRequest(userID, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusCreated, serviceReq)
}

// GetRequest handles GET /v2/servicecatalog/requests/{id}
func (h *ServiceCatalogHandler) GetRequest(w http.ResponseWriter, r *http.Request, id int64) {
	serviceReq, err := h.svc.GetRequest(id)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, serviceReq)
}

// ListRequests handles GET /v2/servicecatalog/requests
func (h *ServiceCatalogHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	filter := model.ServiceRequestFilter{
		RequesterID:  q.Get("requester_id"),
		AssignedToID: q.Get("assigned_to_id"),
		Status:       q.Get("status"),
		Priority:     q.Get("priority"),
		Search:       q.Get("search"),
	}
	if ws := q.Get("workspace_id"); ws != "" {
		if wsID, err := strconv.ParseInt(ws, 10, 64); err == nil {
			filter.WorkspaceID = wsID
		}
	}
	if item := q.Get("catalog_item_id"); item != "" {
		if itemID, err := strconv.ParseInt(item, 10, 64); err == nil {
			filter.CatalogItemID = &itemID
		}
	}
	if l := q.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			filter.Limit = val
		}
	}
	if o := q.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			filter.Offset = val
		}
	}

	requests, total, err := h.svc.ListRequests(filter)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]interface{}{
		"requests": requests,
		"total":    total,
	})
}

// UpdateRequest handles PUT/PATCH /v2/servicecatalog/requests/{id}
func (h *ServiceCatalogHandler) UpdateRequest(w http.ResponseWriter, r *http.Request, id int64) {
	var req model.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}

	serviceReq, err := h.svc.UpdateRequest(id, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, serviceReq)
}

// ReviewApproval handles POST /v2/servicecatalog/approvals/{id}
func (h *ServiceCatalogHandler) ReviewApproval(w http.ResponseWriter, r *http.Request, approvalID int64) {
	var req model.ReviewApprovalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getServiceCatalogUser(r)

	approval, err := h.svc.ReviewApproval(userID, approvalID, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, approval)
}

// AddComment handles POST /v2/servicecatalog/requests/{id}/comments
func (h *ServiceCatalogHandler) AddComment(w http.ResponseWriter, r *http.Request, requestID int64) {
	var req model.AddRequestCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid JSON payload")
		return
	}
	userID := getServiceCatalogUser(r)

	comment, err := h.svc.AddComment(userID, requestID, req)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusCreated, comment)
}

// ListComments handles GET /v2/servicecatalog/requests/{id}/comments
func (h *ServiceCatalogHandler) ListComments(w http.ResponseWriter, r *http.Request, requestID int64) {
	comments, err := h.svc.ListComments(requestID)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, map[string]interface{}{
		"comments": comments,
	})
}

// GetMetrics handles GET /v2/servicecatalog/metrics
func (h *ServiceCatalogHandler) GetMetrics(w http.ResponseWriter, r *http.Request) {
	wsStr := r.URL.Query().Get("workspace_id")
	if wsStr == "" {
		writeServiceCatalogError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}
	wsID, err := strconv.ParseInt(wsStr, 10, 64)
	if err != nil || wsID <= 0 {
		writeServiceCatalogError(w, http.StatusBadRequest, "invalid workspace_id")
		return
	}

	metrics, err := h.svc.GetMetrics(wsID)
	if err != nil {
		handleServiceCatalogError(w, err)
		return
	}
	writeServiceCatalogJSON(w, http.StatusOK, metrics)
}

// ServeHTTP routes requests under /v2/servicecatalog
func (h *ServiceCatalogHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v2/servicecatalog")
	path = strings.TrimPrefix(path, "/")

	if path == "metrics" {
		if r.Method != http.MethodGet {
			writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}
		h.GetMetrics(w, r)
		return
	}

	parts := strings.Split(path, "/")

	// Routes under /v2/servicecatalog/categories
	if parts[0] == "categories" {
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListCategories(w, r)
			case http.MethodPost:
				h.CreateCategory(w, r)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 {
			id, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || id <= 0 {
				writeServiceCatalogError(w, http.StatusBadRequest, "invalid category id")
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.GetCategory(w, r, id)
			case http.MethodPut, http.MethodPatch:
				h.UpdateCategory(w, r, id)
			case http.MethodDelete:
				h.DeleteCategory(w, r, id)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	// Routes under /v2/servicecatalog/items
	if parts[0] == "items" {
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListItems(w, r)
			case http.MethodPost:
				h.CreateItem(w, r)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 {
			id, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || id <= 0 {
				writeServiceCatalogError(w, http.StatusBadRequest, "invalid item id")
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.GetItem(w, r, id)
			case http.MethodPut, http.MethodPatch:
				h.UpdateItem(w, r, id)
			case http.MethodDelete:
				h.DeleteItem(w, r, id)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	// Routes under /v2/servicecatalog/requests
	if parts[0] == "requests" {
		if len(parts) == 1 {
			switch r.Method {
			case http.MethodGet:
				h.ListRequests(w, r)
			case http.MethodPost:
				h.SubmitRequest(w, r)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 2 {
			id, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || id <= 0 {
				writeServiceCatalogError(w, http.StatusBadRequest, "invalid request id")
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.GetRequest(w, r, id)
			case http.MethodPut, http.MethodPatch:
				h.UpdateRequest(w, r, id)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
		if len(parts) == 3 && parts[2] == "comments" {
			id, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || id <= 0 {
				writeServiceCatalogError(w, http.StatusBadRequest, "invalid request id")
				return
			}
			switch r.Method {
			case http.MethodGet:
				h.ListComments(w, r, id)
			case http.MethodPost:
				h.AddComment(w, r, id)
			default:
				writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
			}
			return
		}
	}

	// Routes under /v2/servicecatalog/approvals/{id}
	if parts[0] == "approvals" && len(parts) == 2 {
		id, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || id <= 0 {
			writeServiceCatalogError(w, http.StatusBadRequest, "invalid approval id")
			return
		}
		if r.Method == http.MethodPost {
			h.ReviewApproval(w, r, id)
			return
		}
		writeServiceCatalogError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeServiceCatalogError(w, http.StatusNotFound, "not found")
}
