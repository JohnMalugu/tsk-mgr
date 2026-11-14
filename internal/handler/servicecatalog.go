package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

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
